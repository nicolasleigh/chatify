package main

import (
	"net/http"
	"strconv"

	"github.com/nicolasleigh/chat-app/store"
)

func (app *application) getConversation(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("conversation_id")
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	conversation_id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	payload := store.GetConversationParams{
		ClerkID:        clerkID,
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
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var conversations [][]store.GetConversationRow

	conversationIds, err := app.query.GetConversationsByClerkId(ctx, clerkID)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	for _, conversation_id := range conversationIds {
		payload := store.GetConversationParams{
			ClerkID:        clerkID,
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
	idString := r.PathValue("clerk_id")

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

	payload.Name = &body.Name
	payload.Column3 = body.Member_id_arr
	payload.ClerkID = idString

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
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idString := r.PathValue("conversation_id")
	conversationID, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	payload := store.LeaveGroupParams{
		ClerkID: clerkID,
		ID:      int64(conversationID),
	}
	err = app.query.LeaveGroup(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	err = writeJSON(w, http.StatusOK, "success")
}

func (app *application) deleteGroup(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	idString := r.PathValue("conversation_id")
	conversationID, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	payload := store.DeleteGroupParams{
		ClerkID: clerkID,
		ID:      int64(conversationID),
	}
	err = app.query.DeleteGroup(ctx, payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	err = writeJSON(w, http.StatusOK, "success")
}
