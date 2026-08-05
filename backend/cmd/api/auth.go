package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/jackc/pgx/v5"
	"github.com/nicolasleigh/chat-app/store"
)

type contextKey string

const userContextKey contextKey = "chatifyUser"

// requireUser resolves the authenticated user from the Clerk JWT session claims
// (populated by clerkhttp.WithHeaderAuthorization) and stashes store.User in ctx.
// Handlers wrapped with it are guaranteed to have a valid, synced user.
func (app *application) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims, ok := clerk.SessionClaimsFromContext(r.Context())
		if !ok || claims == nil {
			unauthorizedResponse(w, errors.New("missing session claims"))
			return
		}
		user, err := app.query.GetUser(r.Context(), claims.Subject)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				unauthorizedResponse(w, errors.New("user not synced"))
				return
			}
			serverErrorResponse(w, err)
			return
		}
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next(w, r.WithContext(ctx))
	}
}

// currentUser returns the authenticated user stashed by requireUser.
func currentUser(r *http.Request) (store.User, bool) {
	user, ok := r.Context().Value(userContextKey).(store.User)
	return user, ok
}

// requireConversationMember verifies the current user is a member of the
// conversation. It writes the response and returns false when access is denied.
func (app *application) requireConversationMember(w http.ResponseWriter, r *http.Request, conversationID int64) bool {
	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return false
	}
	_, err := app.query.GetConversationMember(r.Context(), store.GetConversationMemberParams{
		ConversationID: conversationID,
		MemberID:       user.ID,
	})
	if err != nil {
		forbiddenResponse(w, errors.New("not a member of this conversation"))
		return false
	}
	return true
}

// isTrustedOrigin reports whether the given Origin header value is allowed.
func (app *application) isTrustedOrigin(origin string) bool {
	for _, trusted := range app.config.cors.trustedOrigins {
		if origin == trusted {
			return true
		}
	}
	return false
}
