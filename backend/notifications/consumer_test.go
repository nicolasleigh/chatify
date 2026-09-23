package notifications

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/nicolasleigh/chat-app/messaging"
	"github.com/nicolasleigh/chat-app/store"
)

type testPresence struct {
	online map[int64]bool
}

func (p testPresence) IsOnline(userID int64) bool {
	return p.online[userID]
}

func TestNewConsumerRequiresDependencies(t *testing.T) {
	if _, err := NewConsumer(nil, testPresence{}); err == nil {
		t.Fatal("NewConsumer() error = nil, want repository validation error")
	}
}

func TestConsumerRejectsUnsupportedEventsBeforeDatabaseAccess(t *testing.T) {
	consumer, err := NewConsumer(testDeliveryRepository{}, testPresence{online: map[int64]bool{}})
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}

	event, err := messaging.NewEnvelope("1", "friend.created", map[string]any{})
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}
	if err := consumer.Handle(context.Background(), event); err == nil {
		t.Fatal("Handle() error = nil, want unsupported event error")
	}
}

type testDeliveryRepository struct{}

func (testDeliveryRepository) GetConversationNotificationRecipients(context.Context, store.GetConversationNotificationRecipientsParams) ([]int64, error) {
	return nil, nil
}

func (testDeliveryRepository) GetEnabledPushSubscriptions(context.Context, int64) ([]store.PushSubscription, error) {
	return nil, nil
}

func (testDeliveryRepository) InsertNotificationDelivery(context.Context, store.InsertNotificationDeliveryParams) error {
	return nil
}

type recordingDeliveryRepository struct {
	recipients     []int64
	subscriptions  map[int64][]store.PushSubscription
	inserted       []store.InsertNotificationDeliveryParams
	requestedUsers []int64
}

func (r *recordingDeliveryRepository) GetConversationNotificationRecipients(context.Context, store.GetConversationNotificationRecipientsParams) ([]int64, error) {
	return r.recipients, nil
}

func (r *recordingDeliveryRepository) GetEnabledPushSubscriptions(_ context.Context, userID int64) ([]store.PushSubscription, error) {
	r.requestedUsers = append(r.requestedUsers, userID)
	return r.subscriptions[userID], nil
}

func (r *recordingDeliveryRepository) InsertNotificationDelivery(_ context.Context, arg store.InsertNotificationDeliveryParams) error {
	r.inserted = append(r.inserted, arg)
	return nil
}

func TestConsumerPlansDeliveriesOnlyForOfflineEnabledSubscriptions(t *testing.T) {
	repository := &recordingDeliveryRepository{
		recipients: []int64{10, 20, 30},
		subscriptions: map[int64][]store.PushSubscription{
			10: {
				{ID: 101, UserID: 10, Enabled: true},
				{ID: 102, UserID: 10, Enabled: false},
			},
			20: {{ID: 201, UserID: 20, Enabled: true}},
		},
	}
	presence := testPresence{online: map[int64]bool{20: true}}
	consumer, err := NewConsumer(repository, presence)
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}

	payload, err := json.Marshal(map[string]any{
		"message_id":      501,
		"conversation_id": 601,
		"sender_id":       701,
		"type":            "text",
		"content":         "hello",
	})
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}
	event, err := messaging.NewEnvelope("801", messaging.MessageCreated, json.RawMessage(payload))
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	if err := consumer.Handle(context.Background(), event); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if len(repository.requestedUsers) != 2 || repository.requestedUsers[0] != 10 || repository.requestedUsers[1] != 30 {
		t.Fatalf("subscription lookup users = %v, want [10 30]", repository.requestedUsers)
	}
	if len(repository.inserted) != 1 {
		t.Fatalf("inserted deliveries = %d, want 1", len(repository.inserted))
	}
	if got := repository.inserted[0]; got.EventID != 801 || got.MessageID != 501 || got.RecipientID != 10 || got.SubscriptionID != 101 || got.Channel != webPushChannel {
		t.Fatalf("inserted delivery = %+v, want event/message/recipient/subscription/channel 801/501/10/101/%q", got, webPushChannel)
	}
}
