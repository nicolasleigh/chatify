package rabbitmq

import "testing"

func TestNewRejectsMissingURLBeforeDialing(t *testing.T) {
	_, err := New(Config{EventsExchange: "chatify.events"})
	if err == nil {
		t.Fatal("New() error = nil, want missing URL validation error")
	}
}

func TestNewRejectsMissingExchangeBeforeDialing(t *testing.T) {
	_, err := New(Config{URL: "amqp://guest:guest@localhost:5672/"})
	if err == nil {
		t.Fatal("New() error = nil, want missing exchange validation error")
	}
}

func TestNewDoesNotRequireBrokerAtConstruction(t *testing.T) {
	client, err := New(Config{
		URL:            "amqp://guest:guest@127.0.0.1:1/",
		EventsExchange: "chatify.events",
	})
	if err != nil {
		t.Fatalf("New() error = %v, want lazy construction without a broker", err)
	}
	if err := client.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}
}
