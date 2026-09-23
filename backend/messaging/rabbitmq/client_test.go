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
