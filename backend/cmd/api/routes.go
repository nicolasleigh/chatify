package main

import (
	"net/http"

	clerkhttp "github.com/clerk/clerk-sdk-go/v2/http"
)

func (app *application) NewRouter() http.Handler {
	mux := http.NewServeMux()
	hub := newHub()
	requireAuth := clerkhttp.RequireHeaderAuthorization(clerkhttp.AuthorizationJWTExtractor(extractClerkJWT))
	go hub.run()

	// Health
	mux.HandleFunc("GET /health", healthCheckHandler)
	mux.HandleFunc("GET /live", healthCheckHandler)
	mux.HandleFunc("GET /ready", app.readinessCheckHandler)
	// User
	mux.HandleFunc("POST /user", app.createUserHandler)
	// Friend Request
	mux.Handle("POST /request", requireAuth(http.HandlerFunc(app.createRequest)))
	mux.Handle("DELETE /deny/{request_id}", requireAuth(http.HandlerFunc(app.denyRequest)))
	mux.Handle("POST /request/accept/{request_id}", requireAuth(http.HandlerFunc(app.acceptRequest)))
	mux.Handle("GET /friends/{clerk_id}", requireAuth(http.HandlerFunc(app.getFriends)))
	mux.HandleFunc("DELETE /friend/{conversation_id}", app.deleteFriend)
	mux.Handle("GET /requests/{clerk_id}", requireAuth(http.HandlerFunc(app.getRequests)))
	// Message
	mux.Handle("POST /message", requireAuth(http.HandlerFunc(app.createMessage)))
	mux.Handle("GET /messages/{conversation_id}", requireAuth(http.HandlerFunc(app.getMessages)))
	mux.Handle("POST /message/mark_read", requireAuth(http.HandlerFunc(app.markReadMessage)))
	mux.Handle("GET /message/{message_id}", requireAuth(http.HandlerFunc(app.getConversationLastMessage)))
	mux.Handle("GET /message/unseen/{clerk_id}", requireAuth(http.HandlerFunc(app.getAllUnseenMessageCount)))
	// Conversation
	mux.Handle("GET /conversation/{clerk_id}/{conversation_id}", requireAuth(http.HandlerFunc(app.getConversation)))
	mux.Handle("GET /conversations/{clerk_id}", requireAuth(http.HandlerFunc(app.getAllConversations)))
	// Group
	mux.HandleFunc("POST /group/create/{clerk_id}", app.createGroup)
	mux.HandleFunc("DELETE /group/leave/{clerk_id}/{conversation_id}", app.leaveGroup)
	mux.HandleFunc("DELETE /group/delete/{clerk_id}/{conversation_id}", app.deleteGroup)
	// WebSocket
	mux.Handle("/ws/{conversation_id}", requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.handleWebSocket(hub, w, r)
	})))

	return mux
}
