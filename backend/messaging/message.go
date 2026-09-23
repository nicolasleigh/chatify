// Package messaging contains the broker-independent messaging contract.
//
// Business code should depend on this package instead of importing a concrete
// RabbitMQ client. The current adapter is RabbitMQ, but keeping the envelope
// and publisher interface small makes a future NATS JetStream adapter possible
// without leaking broker-specific types into handlers or repositories.
package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

const (
	// MessageCreated is emitted after a message and its outbox record have been
	// committed in PostgreSQL. Consumers may use it for notifications, search
	// indexing, analytics, or other asynchronous work.
	MessageCreated = "message.created"

	// CurrentEnvelopeVersion lets consumers evolve their decoding logic while
	// remaining compatible with older messages that are still in a queue.
	CurrentEnvelopeVersion int = 1
)

// Envelope is the transport-neutral shape shared by all asynchronous events.
//
// ID must remain stable when an outbox row is retried. Consumers can use it as
// an idempotency key, which is required because the delivery model is at least
// once rather than exactly once.
type Envelope struct {
	ID            string            `json:"id"`
	Type          string            `json:"type"`
	Version       int               `json:"version"`
	OccurredAt    time.Time         `json:"occurred_at"`
	TraceID       string            `json:"trace_id,omitempty"`
	CorrelationID string            `json:"correlation_id,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
	Payload       json.RawMessage   `json:"payload"`
}

// NewEnvelope serializes a typed payload into the common event envelope.
// Serialization happens before an outbox transaction is started so malformed
// payloads fail synchronously and can never be persisted as unusable events.
func NewEnvelope(id, eventType string, payload any) (Envelope, error) {
	serializedPayload, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, err
	}

	envelope := Envelope{
		ID:         id,
		Type:       eventType,
		Version:    CurrentEnvelopeVersion,
		OccurredAt: time.Now().UTC(),
		Payload:    serializedPayload,
	}
	if err := envelope.Validate(); err != nil {
		return Envelope{}, err
	}

	return envelope, nil
}

// Validate rejects envelopes that cannot be safely routed or deduplicated.
// Adapters call this before publishing, while outbox code may validate it when
// decoding a persisted payload before attempting a broker publish.
func (e Envelope) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return errors.New("message ID is required")
	}
	if strings.TrimSpace(e.Type) == "" {
		return errors.New("message type is required")
	}
	if e.Version < 1 {
		return errors.New("message version must be positive")
	}
	if e.OccurredAt.IsZero() {
		return errors.New("message timestamp is required")
	}
	if len(e.Payload) == 0 || !json.Valid(e.Payload) {
		return errors.New("message payload must be valid JSON")
	}

	return nil
}

// Publisher is the only broker capability needed by the outbox worker.
// Implementations must not report success until the broker has accepted the
// message (RabbitMQ uses publisher confirms for this guarantee).
type Publisher interface {
	Publish(ctx context.Context, envelope Envelope) error
	Close() error
}

// PermanentError marks an event that cannot become valid by retrying, such as
// malformed JSON or an unsupported schema version. Queue consumers should ACK
// these messages after recording the error instead of requeueing forever.
type PermanentError struct {
	Err error
}

func (e *PermanentError) Error() string {
	if e == nil || e.Err == nil {
		return "permanent messaging error"
	}
	return e.Err.Error()
}

func (e *PermanentError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewPermanentError(err error) error {
	return &PermanentError{Err: err}
}

func IsPermanentError(err error) bool {
	var permanentErr *PermanentError
	return errors.As(err, &permanentErr)
}
