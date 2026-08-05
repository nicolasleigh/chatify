package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
	"github.com/nicolasleigh/chat-app/store"
)

func (app *application) createMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		ConversationID int64   `json:"conversation_id"`
		Type           *string `json:"type"`
		Content        *string `json:"content"`
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

	if !app.requireConversationMember(w, r, body.ConversationID) {
		return
	}

	payload := store.CreateMessageParams{
		Content:  body.Content,
		ID:       body.ConversationID,
		SenderID: user.ID,
		Type:     body.Type,
	}

	_, err = app.query.CreateMessage(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, "Message created")
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}

func (app *application) getMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("conversation_id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	if !app.requireConversationMember(w, r, int64(id)) {
		return
	}

	messages, err := app.query.GetMessages(ctx, int64(id))
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, messages)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}

func (app *application) markReadMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	var payload store.MarkReadMessageParams

	err := readJSON(w, r, &payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	if !app.requireConversationMember(w, r, payload.ConversationID) {
		return
	}
	payload.MemberID = user.ID

	err = app.query.MarkReadMessage(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, nil)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}

func (app *application) getConversationLastMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idStr := r.PathValue("message_id")
	message_id, err := strconv.Atoi(idStr)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	conversationID, err := app.query.GetMessageConversationId(ctx, int64(message_id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			notFoundResponse(w, err)
			return
		}
		serverErrorResponse(w, err)
		return
	}

	if !app.requireConversationMember(w, r, conversationID) {
		return
	}

	message, err := app.query.GetConversationLastMessage(ctx, int64(message_id))
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, message)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}

func (app *application) getAllUnseenMessageCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	data, err := app.query.GetAllUnseenMessageCount(ctx, user.ClerkID)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	err = writeJSON(w, http.StatusOK, data)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}
