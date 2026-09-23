package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nicolasleigh/chat-app/store"
)

// Client represents a connected websocket client
type Client struct {
	conn           *websocket.Conn
	send           chan []byte
	userID         string
	conversationID int64
}

// Hub maintains active clients grouped by conversation
type Hub struct {
	// Map of conversation ID to a map of clients in that conversation
	conversations map[int64]map[*Client]bool
	broadcast     chan []byte
	register      chan *Client
	unregister    chan *Client
	mu            sync.RWMutex
}

type Message struct {
	SenderID       int64   `json:"sender_id"`
	ConversationID int64   `json:"conversation_id"`
	Type           *string `json:"type"`
	Content        *string `json:"content"`
}

const (
	websocketWriteWait      = 10 * time.Second
	websocketPongWait       = 60 * time.Second
	websocketPingPeriod     = (websocketPongWait * 9) / 10
	websocketMaxMessageSize = 64 * 1024
)

func isTrustedOrigin(origin string, trustedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, trustedOrigin := range trustedOrigins {
		if origin == trustedOrigin {
			return true
		}
	}

	return false
}

func newHub() *Hub {
	return &Hub{
		conversations: make(map[int64]map[*Client]bool),
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		unregister:    make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			// Initialize conversation map if it doesn't exist
			if _, exists := h.conversations[client.conversationID]; !exists {
				h.conversations[client.conversationID] = make(map[*Client]bool)
			}
			// Add client to their conversation group
			h.conversations[client.conversationID][client] = true

			// Log connection for debugging
			log.Printf("User %s joined conversation %d. Total participants: %d",
				client.userID,
				client.conversationID,
				len(h.conversations[client.conversationID]))
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			// Remove client from their conversation group
			if clients, exists := h.conversations[client.conversationID]; exists {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.send)

					// Remove conversation if empty
					if len(clients) == 0 {
						delete(h.conversations, client.conversationID)
					}

					log.Printf("User %s left conversation %d. Remaining participants: %d",
						client.userID,
						client.conversationID,
						len(clients))
				}
			}
			h.mu.Unlock()

		case message := <-h.broadcast:
			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("Error unmarshaling message: %v", err)
				continue
			}

			h.mu.RLock()
			// Send message only to clients in the same conversation
			if clients, exists := h.conversations[msg.ConversationID]; exists {
				for client := range clients {
					select {
					case client.send <- message:
					default:
						close(client.send)
						delete(clients, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (c *Client) readPump(hub *Hub, app *application) {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(websocketMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(websocketPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(websocketPongWait))
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("Error unmarshaling message: %v", err)
			continue
		}

		// Validate that the message is for the correct conversation
		if msg.ConversationID != c.conversationID {
			log.Printf("Warning: User %s tried to send message to conversation %d while in conversation %d",
				c.userID, msg.ConversationID, c.conversationID)
			continue
		}

		// content: "hi"
		// conversation_id: 30
		// created_at: "2025-02-16T22:00:46+08:00"
		// email: "jier@e.com"
		// image_url: "https://cdn.pixabay.com/photo/2021/11/12/03/04/woman-6787784_1280.png"
		// message_id: 46
		// type: "text"
		// user_id: 1
		// username: "JJJJ"

		// content: "hi"
		// conversation_id: 30
		// sender_id: 15
		// type: "text"

		// Store message in database
		payload := store.CreateMessageParams{
			Content:  msg.Content,
			ID:       msg.ConversationID,
			SenderID: msg.SenderID,
			Type:     msg.Type,
		}

		messageId, err := app.query.CreateMessage(context.Background(), payload)
		if err != nil {
			log.Printf("Error storing message: %v", err)
			continue
		}

		returnMessage, err := app.query.GetMessageById(context.Background(), int64(messageId))
		if err != nil {
			log.Printf("Error get message: %v", err)
			continue
		}

		returnMsg, err := json.Marshal(returnMessage)
		if err != nil {
			log.Printf("Error marshaling message: %v", err)
		}

		hub.broadcast <- returnMsg
	}
}

func (c *Client) writePump() {
	defer func() {
		c.conn.Close()
	}()

	ticker := time.NewTicker(websocketPingPeriod)
	defer ticker.Stop()

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				_ = c.conn.SetWriteDeadline(time.Now().Add(websocketWriteWait))
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			_ = c.conn.SetWriteDeadline(time.Now().Add(websocketWriteWait))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(websocketWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (app *application) handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	if !isTrustedOrigin(r.Header.Get("Origin"), app.config.cors.trustedOrigins) {
		http.Error(w, "forbidden origin", http.StatusForbidden)
		return
	}

	conversationID, err := strconv.ParseInt(r.PathValue("conversation_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	clerkUser, err := getClerkUser(r.Context())
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	// userID := getUserIDFromRequest(r) // Implement this based on your auth system
	userID := clerkUser.ID

	// Validate that the user has access to this conversation
	if !app.hasAccessToConversation(userID, conversationID) {
		http.Error(w, "Unauthorized access to conversation", http.StatusForbidden)
		return
	}

	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return isTrustedOrigin(r.Header.Get("Origin"), app.config.cors.trustedOrigins)
		},
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         userID,
		conversationID: conversationID,
	}

	hub.register <- client

	// Start goroutines for pumping messages
	go client.writePump()
	go client.readPump(hub, app)
}

// Helper function to check if a user has access to a conversation
func (app *application) hasAccessToConversation(userID string, conversationID int64) bool {
	// Implement your access control logic here
	// Example: Check if the user is a member of the conversation in your database
	return true
}
