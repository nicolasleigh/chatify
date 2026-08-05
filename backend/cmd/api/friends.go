package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/nicolasleigh/chat-app/store"
)

func (app *application) createRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		Email string `json:"email"`
	}

	err := readJSON(w, r, &body)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	payload := store.CreateRequestParams{
		ClerkID: user.ClerkID,
		Email:   body.Email,
	}

	err = Validate.Struct(payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = app.query.CreateRequest(ctx, payload)
	if err != nil {
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

	req, err := app.query.GetRequest(ctx, int64(id))
	if err != nil {
		notFoundResponse(w, err)
		return
	}

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}
	// Only the request receiver may deny it.
	if req.ReceiverID != user.ID {
		forbiddenResponse(w, errors.New("you cannot deny this request"))
		return
	}

	friend, err := app.query.DeleteRequest(ctx, req.ID)
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

	req, err := app.query.GetRequest(ctx, int64(request_id))
	if err != nil {
		notFoundResponse(w, err)
		return
	}

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}
	// Only the request receiver may accept it. The sender/receiver ids are
	// derived server-side, never taken from the client.
	if req.ReceiverID != user.ID {
		forbiddenResponse(w, errors.New("you cannot accept this request"))
		return
	}

	payload := store.AcceptRequestParams{
		Column1: req.SenderID,
		Column2: req.ReceiverID,
	}

	err = app.query.AcceptRequest(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	_, err = app.query.DeleteRequest(ctx, req.ID)
	if err != nil {
		serverErrorResponse(w, err)
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

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	friends, err := app.query.GetFriends(ctx, user.ClerkID)
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
	idStr := r.PathValue("conversation_id")
	conversation_id, err := strconv.Atoi(idStr)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	if !app.requireConversationMember(w, r, int64(conversation_id)) {
		return
	}

	err = app.query.DeleteFriend(ctx, int64(conversation_id))
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	err = writeJSON(w, http.StatusOK, "success")
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
}

func (app *application) getRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	friendReq, err := app.query.GetRequests(ctx, user.ClerkID)
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
