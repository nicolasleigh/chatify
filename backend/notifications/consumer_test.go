package notifications

import (
	"context"
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

func (testDeliveryRepository) GetEnabledPushSubscriptions(context.Context, int64) ([]store.GetEnabledPushSubscriptionsRow, error) {
	return nil, nil
}

func (testDeliveryRepository) InsertNotificationDelivery(context.Context, store.InsertNotificationDeliveryParams) error {
	return nil
}
