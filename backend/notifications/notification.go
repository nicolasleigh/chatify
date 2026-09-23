// Package notifications contains broker- and provider-independent notification
// contracts. RabbitMQ consumers create notification work, while providers
// deliver it through Web Push today or another channel in the future.
package notifications

import (
	"context"
	"errors"
	"fmt"
)

// Subscription contains only the browser credentials required by a provider.
// The database row ID is retained so delivery records can identify the exact
// browser that succeeded or became invalid.
type Subscription struct {
	ID       int64
	Endpoint string
	P256dh   string
	Auth     string
}

// Payload intentionally contains a short, privacy-conscious notification
// preview. Providers should not receive the complete message body by default.
type Payload struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	ConversationID int64  `json:"conversation_id"`
	MessageID      int64  `json:"message_id"`
}

// Provider sends one notification to one browser/device subscription.
// Implementations must classify permanent endpoint failures so the delivery
// worker can disable invalid subscriptions instead of retrying forever.
type Provider interface {
	Send(ctx context.Context, subscription Subscription, payload Payload) error
}

type ErrorKind string

const (
	// TransientError means the same subscription may succeed later, for example
	// because a push service returned 429 or a network request timed out.
	TransientError ErrorKind = "transient"

	// PermanentError means the endpoint is no longer usable, commonly HTTP 404
	// or 410. Retrying such an endpoint wastes worker capacity.
	PermanentError ErrorKind = "permanent"
)

// DeliveryError preserves retry classification while keeping the underlying
// provider error available for logs and metrics.
type DeliveryError struct {
	Kind ErrorKind
	Err  error
}

func (e *DeliveryError) Error() string {
	if e == nil || e.Err == nil {
		return "notification delivery failed"
	}
	return fmt.Sprintf("%s notification delivery error: %v", e.Kind, e.Err)
}

func (e *DeliveryError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewTransientError(err error) error {
	if err == nil {
		return &DeliveryError{Kind: TransientError, Err: errors.New("notification delivery failed")}
	}
	return &DeliveryError{Kind: TransientError, Err: err}
}

func NewPermanentError(err error) error {
	if err == nil {
		return &DeliveryError{Kind: PermanentError, Err: errors.New("notification endpoint is invalid")}
	}
	return &DeliveryError{Kind: PermanentError, Err: err}
}

func IsPermanentError(err error) bool {
	var deliveryErr *DeliveryError
	return errors.As(err, &deliveryErr) && deliveryErr.Kind == PermanentError
}
