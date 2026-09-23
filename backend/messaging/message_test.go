package messaging

import (
	"errors"
	"testing"
	"time"
)

func TestNewEnvelope(t *testing.T) {
	envelope, err := NewEnvelope("outbox-1", MessageCreated, map[string]any{"message_id": 42})
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}
	if envelope.ID != "outbox-1" {
		t.Fatalf("ID = %q, want %q", envelope.ID, "outbox-1")
	}
	if envelope.Version != CurrentEnvelopeVersion {
		t.Fatalf("Version = %d, want %d", envelope.Version, CurrentEnvelopeVersion)
	}
	if envelope.OccurredAt.Before(time.Now().Add(-time.Minute)) {
		t.Fatalf("OccurredAt = %s, want a recent timestamp", envelope.OccurredAt)
	}
	if err := envelope.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestPermanentErrorPreservesCause(t *testing.T) {
	rootErr := errors.New("invalid event")
	err := NewPermanentError(rootErr)
	if !IsPermanentError(err) {
		t.Fatal("IsPermanentError() = false, want true")
	}
	if !errors.Is(err, rootErr) {
		t.Fatal("PermanentError should unwrap its cause")
	}
}

func TestEnvelopeValidationRejectsMissingIdentity(t *testing.T) {
	envelope := Envelope{Type: MessageCreated, Version: 1, OccurredAt: time.Now(), Payload: []byte(`{}`)}
	if err := envelope.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want missing ID error")
	}
}
