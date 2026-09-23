// Package outbox delivers database-backed events to the configured publisher.
// The database is the source of truth: RabbitMQ outages delay delivery but do
// not make an already-created chat message disappear.
package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/nicolasleigh/chat-app/messaging"
	"github.com/nicolasleigh/chat-app/store"
)

const (
	defaultMaxAttempts = 5
	initialRetryDelay  = time.Second
	maxRetryDelay      = 5 * time.Minute
)

// Repository is the minimum database surface required by the worker. Keeping
// it as an interface makes retry behavior testable without a live PostgreSQL
// or RabbitMQ instance, and prevents the worker from depending on unrelated
// application queries.
type Repository interface {
	ClaimOutboxEvents(ctx context.Context, limit int32) ([]store.ClaimOutboxEventsRow, error)
	MarkOutboxEventPublished(ctx context.Context, id int64) error
	MarkOutboxEventFailed(ctx context.Context, arg store.MarkOutboxEventFailedParams) error
}

// Config controls polling and retry behavior. MaxAttempts includes the first
// delivery attempt, so a value of five means one initial try plus four retries.
type Config struct {
	PollInterval time.Duration
	BatchSize    int32
	MaxAttempts  int32
	Logger       *slog.Logger
}

// Worker claims and publishes outbox rows until its context is canceled.
// Multiple workers may run safely because ClaimOutboxEvents uses PostgreSQL
// row locks with SKIP LOCKED and a lease for crashed workers.
type Worker struct {
	repository Repository
	publisher  messaging.Publisher
	config     Config
}

// New validates worker dependencies before the application starts its
// background goroutine. A zero logger is replaced with the process default so
// database or broker failures are never silently discarded.
func New(repository Repository, publisher messaging.Publisher, cfg Config) (*Worker, error) {
	if repository == nil {
		return nil, errors.New("outbox repository is required")
	}
	if publisher == nil {
		return nil, errors.New("outbox publisher is required")
	}
	if cfg.PollInterval <= 0 {
		return nil, errors.New("outbox poll interval must be greater than zero")
	}
	if cfg.BatchSize < 1 {
		return nil, errors.New("outbox batch size must be greater than zero")
	}
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = defaultMaxAttempts
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &Worker{
		repository: repository,
		publisher:  publisher,
		config:     cfg,
	}, nil
}

// Run performs one delivery pass immediately and then polls at the configured
// interval. A polling error is logged and retried on the next tick because a
// transient database failure must not permanently stop asynchronous delivery.
func (w *Worker) Run(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	w.drain(ctx)
	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			w.drain(ctx)
		}
	}
}

func (w *Worker) drain(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}

	events, err := w.repository.ClaimOutboxEvents(ctx, w.config.BatchSize)
	if err != nil {
		if ctx.Err() == nil {
			w.config.Logger.Error("claim outbox events", "error", err)
		}
		return
	}

	for _, event := range events {
		if ctx.Err() != nil {
			return
		}
		w.deliver(ctx, event)
	}
}

func (w *Worker) deliver(ctx context.Context, event store.ClaimOutboxEventsRow) {
	envelope := envelopeFromRow(event)
	if err := envelope.Validate(); err != nil {
		w.fail(ctx, event, fmt.Errorf("validate outbox event: %w", err))
		return
	}

	if err := w.publisher.Publish(ctx, envelope); err != nil {
		if ctx.Err() != nil {
			// On graceful shutdown leave the row leased. The SQL claim query
			// intentionally reclaims stale leases after five minutes.
			return
		}
		w.fail(ctx, event, fmt.Errorf("publish outbox event %d: %w", event.ID, err))
		return
	}

	if err := w.repository.MarkOutboxEventPublished(ctx, event.ID); err != nil && ctx.Err() == nil {
		w.config.Logger.Error("mark outbox event published", "event_id", event.ID, "error", err)
	}
}

func (w *Worker) fail(ctx context.Context, event store.ClaimOutboxEventsRow, deliveryErr error) {
	if ctx.Err() != nil {
		return
	}

	status := "pending"
	retryAfter := retryDelaySeconds(event.Attempts)
	if event.Attempts >= w.config.MaxAttempts {
		// A failed row is retained instead of being deleted or retried forever.
		// Operators can inspect last_error and decide whether to replay it after
		// the underlying issue is fixed.
		status = "failed"
		retryAfter = 0
	}

	lastError := deliveryErr.Error()
	if err := w.repository.MarkOutboxEventFailed(ctx, store.MarkOutboxEventFailedParams{
		ID:               event.ID,
		Status:           status,
		RetryAfterSecond: retryAfter,
		LastError:        &lastError,
	}); err != nil {
		w.config.Logger.Error("mark outbox event failed", "event_id", event.ID, "error", err)
		return
	}

	w.config.Logger.Warn("outbox event delivery failed",
		"event_id", event.ID,
		"attempts", event.Attempts,
		"status", status,
		"retry_after_seconds", retryAfter,
		"error", deliveryErr,
	)
}

func envelopeFromRow(event store.ClaimOutboxEventsRow) messaging.Envelope {
	return messaging.Envelope{
		// The database sequence ID is stable across every retry and is therefore
		// the idempotency key consumers should persist.
		ID:         strconv.FormatInt(event.ID, 10),
		Type:       event.EventType,
		Version:    int(event.EventVersion),
		OccurredAt: event.CreatedAt.Time.UTC(),
		Metadata: map[string]string{
			"aggregate_type": event.AggregateType,
			"aggregate_id":   strconv.FormatInt(event.AggregateID, 10),
		},
		Payload: json.RawMessage(event.Payload),
	}
}

func retryDelaySeconds(attempts int32) int32 {
	delay := initialRetryDelay
	for attempt := int32(1); attempt < attempts; attempt++ {
		delay *= 2
		if delay >= maxRetryDelay {
			return int32(maxRetryDelay / time.Second)
		}
	}
	return int32(delay / time.Second)
}
