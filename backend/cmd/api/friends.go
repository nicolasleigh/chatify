package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nicolasleigh/chat-app/store"
)

func (app *application) createRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	body := r.Body
	defer body.Close()

	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input struct {
		Email string `json:"email"`
	}

	err = readJSON(w, r, &input)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	payload := store.CreateRequestParams{
		ClerkID: clerkID,
		Email:   input.Email,
	}
	err = Validate.Struct(payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	_, err = app.query.CreateRequest(ctx, payload)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			badRequestResponse(w, errors.New("cannot send a request to this user"))
			return
		}
		if data, ok := err.(*pgconn.PgError); ok && data.Code == "23502" {
			msg := fmt.Sprintf("Email %s does not exist", payload.Email)
			err = errors.New(msg)
			notFoundResponse(w, err)
			return
		}
		if data, ok := err.(*pgconn.PgError); ok && data.Code == "23505" {
			err = errors.New("You already sent request to this email!")
			forbiddenResponse(w, err)
			return
		}
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, "Created!")
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) denyRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idString := r.PathValue("request_id")

	id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	friend, err := app.query.DeleteRequest(ctx, store.DeleteRequestParams{
		ID:      int64(id),
		ClerkID: clerkID,
	})
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, friend)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) acceptRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	idString := r.PathValue("request_id")

	request_id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	payload := store.AcceptRequestParams{ID: int64(request_id), ClerkID: clerkID}

	_, err = app.query.AcceptRequest(ctx, payload)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			notFoundResponse(w, errors.New("friend request not found"))
			return
		}
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, nil)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) getFriends(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	friends, err := app.query.GetFriends(ctx, clerkID)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, friends)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) deleteFriend(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("conversation_id")
	conversationID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = app.query.DeleteFriend(ctx, store.DeleteFriendParams{
		ConversationID: &conversationID,
		ClerkID:        clerkID,
	})
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (app *application) getRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	friendReq, err := app.query.GetRequests(ctx, clerkID)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, friendReq)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}
