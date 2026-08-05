package main

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/nicolasleigh/chat-app/store"
)

func (app *application) getConversation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("conversation_id")
	conversation_id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	if !app.requireConversationMember(w, r, int64(conversation_id)) {
		return
	}

	payload := store.GetConversationParams{
		ClerkID:        user.ClerkID,
		ConversationID: int64(conversation_id),
	}

	data, err := app.query.GetConversation(ctx, payload)
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

func (app *application) getAllConversations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var conversations [][]store.GetConversationRow

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	conversationIds, err := app.query.GetConversationsByClerkId(ctx, user.ClerkID)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	for _, conversation_id := range conversationIds {
		payload := store.GetConversationParams{
			ClerkID:        user.ClerkID,
			ConversationID: int64(conversation_id),
		}
		data, err := app.query.GetConversation(ctx, payload)
		if err != nil {
			badRequestResponse(w, err)
			return
		}

		conversations = append(conversations, data)
	}

	err = writeJSON(w, http.StatusOK, conversations)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}

func (app *application) createGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var body struct {
		Name          string  `json:"name"`
		Member_id_arr []int64 `json:"member_id_arr"`
	}

	var payload store.CreateGroupParams

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

	payload.Name = &body.Name
	payload.Column3 = body.Member_id_arr
	payload.ClerkID = user.ClerkID

	err = app.query.CreateGroup(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusCreated, "Group created")
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}

func (app *application) leaveGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("conversation_id")
	conversation_id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	payload := store.LeaveGroupParams{
		ClerkID:        user.ClerkID,
		ConversationID: int64(conversation_id),
	}
	err = app.query.LeaveGroup(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, "success")
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}

func (app *application) deleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("conversation_id")
	conversation_id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	user, ok := currentUser(r)
	if !ok {
		unauthorizedResponse(w, errors.New("unauthorized"))
		return
	}

	payload := store.DeleteGroupParams{
		ClerkID: user.ClerkID,
		ID:      int64(conversation_id),
	}
	err = app.query.DeleteGroup(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	err = writeJSON(w, http.StatusOK, "success")
	if err != nil {
		badRequestResponse(w, err)
		return
	}
}
