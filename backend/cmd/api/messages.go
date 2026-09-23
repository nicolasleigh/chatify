package main

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/nicolasleigh/chat-app/store"
)

func (app *application) createMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	localUser, err := app.query.GetUser(ctx, clerkID)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	var payload store.CreateMessageWithOutboxParams
	var body struct {
		SenderID       int64   `json:"sender_id"`
		ConversationID int64   `json:"conversation_id"`
		Type           *string `json:"type"`
		Content        *string `json:"content"`
	}
	err = readJSON(w, r, &body)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	if err := validateCreateMessageInput(body.ConversationID, body.Type, body.Content); err != nil {
		badRequestResponse(w, err)
		return
	}
	payload.Content = body.Content
	payload.ConversationID = body.ConversationID
	payload.SenderID = localUser.ID
	payload.Type = body.Type

	hasAccess, err := app.hasAccessToConversation(ctx, localUser.ID, body.ConversationID)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
	if !hasAccess {
		forbiddenResponse(w, errors.New("conversation access denied"))
		return
	}

	// The message, conversation preview, and outbox event are committed by one
	// database statement. RabbitMQ can therefore be unavailable here without
	// causing a successful user action to lose its integration event.
	_, err = app.query.CreateMessageWithOutbox(ctx, payload)
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

const maxMessageContentBytes = 64 * 1024

func validateCreateMessageInput(conversationID int64, messageType, content *string) error {
	if conversationID <= 0 {
		return errors.New("conversation ID must be positive")
	}
	if messageType == nil || strings.TrimSpace(*messageType) == "" {
		return errors.New("message type is required")
	}
	if len(*messageType) > 200 {
		return errors.New("message type is too long")
	}
	if content == nil || len(*content) == 0 {
		return errors.New("message content is required")
	}
	if len(*content) > maxMessageContentBytes {
		return errors.New("message content is too long")
	}

	return nil
}

func (app *application) getMessages(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	idString := r.PathValue("conversation_id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	limit, err := parseMessageHistoryLimit(r.URL.Query().Get("limit"))
	if err != nil {
		badRequestResponse(w, err)
		return
	}

	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	localUser, err := app.query.GetUser(ctx, clerkID)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	hasAccess, err := app.hasAccessToConversation(ctx, localUser.ID, int64(id))
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
	if !hasAccess {
		forbiddenResponse(w, errors.New("conversation access denied"))
		return
	}

	messages, err := app.query.GetMessages(ctx, store.GetMessagesParams{
		ConversationID: int64(id),
		Limit:          limit,
	})
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

const (
	defaultMessageHistoryLimit int32 = 50
	maxMessageHistoryLimit     int32 = 100
)

func parseMessageHistoryLimit(rawLimit string) (int32, error) {
	if rawLimit == "" {
		return defaultMessageHistoryLimit, nil
	}

	limit, err := strconv.ParseInt(rawLimit, 10, 32)
	if err != nil || limit < 1 || limit > int64(maxMessageHistoryLimit) {
		return 0, errors.New("message history limit must be between 1 and 100")
	}

	return int32(limit), nil
}

func validateMarkReadInput(conversationID int64, lastSeenMessageID *int64) error {
	if conversationID <= 0 {
		return errors.New("conversation ID must be positive")
	}
	if lastSeenMessageID == nil || *lastSeenMessageID <= 0 {
		return errors.New("last seen message ID must be positive")
	}

	return nil
}

func (app *application) markReadMessage(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var payload store.MarkReadMessageParams
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	err = readJSON(w, r, &payload)
	if err != nil {
		badRequestResponse(w, err)
		return
	}
	if err := validateMarkReadInput(payload.ConversationID, payload.LastSeenMessageID); err != nil {
		badRequestResponse(w, err)
		return
	}

	localUser, err := app.query.GetUser(ctx, clerkID)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}

	hasAccess, err := app.hasAccessToConversation(ctx, localUser.ID, payload.ConversationID)
	if err != nil {
		serverErrorResponse(w, err)
		return
	}
	if !hasAccess {
		forbiddenResponse(w, errors.New("conversation access denied"))
		return
	}
	payload.MemberID = localUser.ID

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

	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	message, err := app.query.GetConversationLastMessage(ctx, store.GetConversationLastMessageParams{
		ID:      int64(message_id),
		ClerkID: clerkID,
	})
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
	clerkID, err := authenticatedClerkID(ctx)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	data, err := app.query.GetAllUnseenMessageCount(ctx, clerkID)
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
