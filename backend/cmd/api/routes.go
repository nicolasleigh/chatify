package main

import (
	"net/http"
)

func (app *application) NewRouter() http.Handler {
	mux := http.NewServeMux()
	hub := newHub()
	go hub.run()

	// Health
	mux.HandleFunc("GET /health", healthCheckHandler)
	// User (public: Clerk webhook sync)
	mux.HandleFunc("POST /user", app.createUserHandler)
	// Friend Request
	mux.HandleFunc("POST /request", app.requireUser(app.createRequest))
	mux.HandleFunc("DELETE /deny/{request_id}", app.requireUser(app.denyRequest))
	mux.HandleFunc("POST /request/accept/{request_id}", app.requireUser(app.acceptRequest))
	mux.HandleFunc("GET /friends", app.requireUser(app.getFriends))
	mux.HandleFunc("DELETE /friend/{conversation_id}", app.requireUser(app.deleteFriend))
	mux.HandleFunc("GET /requests", app.requireUser(app.getRequests))
	// Message
	mux.HandleFunc("POST /message", app.requireUser(app.createMessage))
	mux.HandleFunc("GET /messages/{conversation_id}", app.requireUser(app.getMessages))
	mux.HandleFunc("POST /message/mark_read", app.requireUser(app.markReadMessage))
	mux.HandleFunc("GET /message/{message_id}", app.requireUser(app.getConversationLastMessage))
	mux.HandleFunc("GET /message/unseen", app.requireUser(app.getAllUnseenMessageCount))
	// Conversation
	mux.HandleFunc("GET /conversation/{conversation_id}", app.requireUser(app.getConversation))
	mux.HandleFunc("GET /conversations", app.requireUser(app.getAllConversations))
	// Group
	mux.HandleFunc("POST /group/create", app.requireUser(app.createGroup))
	mux.HandleFunc("DELETE /group/leave/{conversation_id}", app.requireUser(app.leaveGroup))
	mux.HandleFunc("DELETE /group/delete/{conversation_id}", app.requireUser(app.deleteGroup))
	// WebSocket
	mux.HandleFunc("/ws/{conversation_id}", app.requireUser(func(w http.ResponseWriter, r *http.Request) {
		app.handleWebSocket(hub, w, r)
	}))

	return mux
}
