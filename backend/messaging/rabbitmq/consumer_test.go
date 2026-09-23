package rabbitmq

import (
	"testing"

	"github.com/nicolasleigh/chat-app/messaging"
)

func TestNewConsumerAppliesSafeDefaults(t *testing.T) {
	consumer, err := NewConsumer(ConsumerConfig{
		URL:      "amqp://guest:guest@localhost:5672/",
		Exchange: "chatify.events",
		Queue:    "chatify.notifications",
	})
	if err != nil {
		t.Fatalf("NewConsumer() error = %v", err)
	}
	if consumer.config.RoutingKey != messaging.MessageCreated {
		t.Fatalf("routing key = %q, want %q", consumer.config.RoutingKey, messaging.MessageCreated)
	}
	if consumer.config.Prefetch != defaultConsumerPrefetch {
		t.Fatalf("prefetch = %d, want %d", consumer.config.Prefetch, defaultConsumerPrefetch)
	}
}

func TestDecodeEnvelopeRejectsInvalidJSONAsPermanent(t *testing.T) {
	_, err := decodeEnvelope([]byte("not-json"))
	if !messaging.IsPermanentError(err) {
		t.Fatal("decodeEnvelope() should classify invalid JSON as permanent")
	}
}
