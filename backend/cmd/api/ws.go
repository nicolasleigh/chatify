package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/nicolasleigh/chat-app/store"
)

// Client represents a connected websocket client
type Client struct {
	conn           *websocket.Conn
	send           chan []byte
	userID         int64
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Origin is validated in handleWebSocket before upgrading
	},
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
			log.Printf("User %d joined conversation %d. Total participants: %d",
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

					log.Printf("User %d left conversation %d. Remaining participants: %d",
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
			log.Printf("Warning: User %d tried to send message to conversation %d while in conversation %d",
				c.userID, msg.ConversationID, c.conversationID)
			continue
		}

		// The sender is the authenticated client, never a client-supplied value.
		msg.SenderID = c.userID

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

	for {
		select {
		case message, ok := <-c.send:
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				return
			}
		}
	}
}

func (app *application) handleWebSocket(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conversationID, err := strconv.ParseInt(r.PathValue("conversation_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid conversation ID", http.StatusBadRequest)
		return
	}

	// Validate the Origin header before upgrading (the shared upgrader's
	// CheckOrigin cannot hold per-app state without a race).
	if origin := r.Header.Get("Origin"); origin != "" && !app.isTrustedOrigin(origin) {
		http.Error(w, "Forbidden origin", http.StatusForbidden)
		return
	}

	// The current user was resolved by requireUser.
	if !app.requireConversationMember(w, r, conversationID) {
		return
	}

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Error upgrading connection: %v", err)
		return
	}

	client := &Client{
		conn:           conn,
		send:           make(chan []byte, 256),
		userID:         user.ID,
		conversationID: conversationID,
	}

	hub.register <- client

	// Start goroutines for pumping messages
	go client.writePump()
	go client.readPump(hub, app)
}
