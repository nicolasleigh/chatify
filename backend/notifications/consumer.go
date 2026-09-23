package notifications

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/nicolasleigh/chat-app/messaging"
	"github.com/nicolasleigh/chat-app/store"
)

const webPushChannel = "web_push"

// Presence determines whether a recipient currently has a live chat
// connection. It is intentionally small so the in-memory API implementation
// and a future shared Redis implementation can both be used here.
type Presence interface {
	IsOnline(userID int64) bool
}

// DeliveryRepository contains only the queries needed to turn a message event
// into durable notification work. Inserting rows is idempotent at the
// database unique-key boundary, so the consumer may safely handle a redelivered
// RabbitMQ event.
type DeliveryRepository interface {
	GetConversationNotificationRecipients(ctx context.Context, arg store.GetConversationNotificationRecipientsParams) ([]int64, error)
	GetEnabledPushSubscriptions(ctx context.Context, userID int64) ([]store.PushSubscription, error)
	InsertNotificationDelivery(ctx context.Context, arg store.InsertNotificationDeliveryParams) error
}

// Consumer plans offline notifications but does not perform external HTTP
// calls. Separating planning from delivery keeps RabbitMQ acknowledgement
// quick and gives the delivery worker its own retry policy.
type Consumer struct {
	repository DeliveryRepository
	presence   Presence
}

func NewConsumer(repository DeliveryRepository, presence Presence) (*Consumer, error) {
	if repository == nil {
		return nil, errors.New("notification delivery repository is required")
	}
	if presence == nil {
		return nil, errors.New("notification presence store is required")
	}
	return &Consumer{repository: repository, presence: presence}, nil
}

type messageCreatedPayload struct {
	MessageID      int64  `json:"message_id"`
	ConversationID int64  `json:"conversation_id"`
	SenderID       int64  `json:"sender_id"`
	Type           string `json:"type"`
	Content        string `json:"content"`
}

// Handle validates the transport envelope, derives recipients from the
// membership table, and persists one delivery intent per offline subscription.
// It returns an error before acknowledgement whenever the database cannot
// safely record the intent.
func (c *Consumer) Handle(ctx context.Context, envelope messaging.Envelope) error {
	if err := envelope.Validate(); err != nil {
		return messaging.NewPermanentError(fmt.Errorf("validate notification event: %w", err))
	}
	if envelope.Type != messaging.MessageCreated {
		return messaging.NewPermanentError(fmt.Errorf("unsupported notification event type %q", envelope.Type))
	}
	eventID, err := strconv.ParseInt(envelope.ID, 10, 64)
	if err != nil || eventID <= 0 {
		return messaging.NewPermanentError(errors.New("notification event ID must be a positive database ID"))
	}

	var payload messageCreatedPayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		return messaging.NewPermanentError(fmt.Errorf("decode message.created payload: %w", err))
	}
	if payload.MessageID <= 0 || payload.ConversationID <= 0 || payload.SenderID <= 0 {
		return messaging.NewPermanentError(errors.New("message.created payload has invalid IDs"))
	}

	recipients, err := c.repository.GetConversationNotificationRecipients(ctx, store.GetConversationNotificationRecipientsParams{
		ConversationID: payload.ConversationID,
		SenderID:       payload.SenderID,
	})
	if err != nil {
		return fmt.Errorf("find notification recipients: %w", err)
	}

	for _, recipientID := range recipients {
		if recipientID <= 0 || c.presence.IsOnline(recipientID) {
			continue
		}

		subscriptions, err := c.repository.GetEnabledPushSubscriptions(ctx, recipientID)
		if err != nil {
			return fmt.Errorf("find push subscriptions for user %d: %w", recipientID, err)
		}
		for _, subscription := range subscriptions {
			if !subscription.Enabled || subscription.ID <= 0 {
				continue
			}
			if err := c.repository.InsertNotificationDelivery(ctx, store.InsertNotificationDeliveryParams{
				EventID:        eventID,
				MessageID:      payload.MessageID,
				RecipientID:    recipientID,
				SubscriptionID: subscription.ID,
				Channel:        webPushChannel,
			}); err != nil {
				return fmt.Errorf("create notification delivery for user %d: %w", recipientID, err)
			}
		}
	}

	return nil
}
