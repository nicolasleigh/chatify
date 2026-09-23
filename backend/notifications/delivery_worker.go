package notifications

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/nicolasleigh/chat-app/store"
)

const (
	defaultDeliveryMaxAttempts = 5
	initialDeliveryRetryDelay  = time.Second
	maxDeliveryRetryDelay      = 5 * time.Minute
	maxNotificationBodyRunes   = 120
)

// DeliveryWorkerRepository is the database surface for external Push calls.
// Claiming and marking are separate from the provider call so a slow or
// unavailable push service never holds a PostgreSQL row lock open.
type DeliveryWorkerRepository interface {
	ClaimNotificationDeliveries(ctx context.Context, limit int32) ([]store.ClaimNotificationDeliveriesRow, error)
	GetNotificationDelivery(ctx context.Context, id int64) (store.GetNotificationDeliveryRow, error)
	GetNotificationMessage(ctx context.Context, id int64) (store.GetNotificationMessageRow, error)
	MarkNotificationDeliverySent(ctx context.Context, id int64) error
	MarkNotificationDeliveryFailed(ctx context.Context, arg store.MarkNotificationDeliveryFailedParams) error
	DisablePushSubscriptionByEndpoint(ctx context.Context, endpoint string) error
}

// DeliveryWorker retries provider failures from the durable delivery table.
// The worker uses at-least-once delivery; the database status and subscription
// ID provide the audit/idempotency boundary.
type DeliveryWorker struct {
	repository  DeliveryWorkerRepository
	provider    Provider
	poll        time.Duration
	batchSize   int32
	maxAttempts int32
	logger      *slog.Logger
}

type DeliveryWorkerConfig struct {
	PollInterval time.Duration
	BatchSize    int32
	MaxAttempts  int32
	Logger       *slog.Logger
}

func NewDeliveryWorker(repository DeliveryWorkerRepository, provider Provider, cfg DeliveryWorkerConfig) (*DeliveryWorker, error) {
	if repository == nil {
		return nil, errors.New("notification delivery repository is required")
	}
	if provider == nil {
		return nil, errors.New("notification provider is required")
	}
	if cfg.PollInterval <= 0 {
		return nil, errors.New("notification delivery poll interval must be greater than zero")
	}
	if cfg.BatchSize < 1 {
		return nil, errors.New("notification delivery batch size must be greater than zero")
	}
	if cfg.MaxAttempts < 1 {
		cfg.MaxAttempts = defaultDeliveryMaxAttempts
	}
	if cfg.Logger == nil {
		cfg.Logger = slog.Default()
	}

	return &DeliveryWorker{
		repository:  repository,
		provider:    provider,
		poll:        cfg.PollInterval,
		batchSize:   cfg.BatchSize,
		maxAttempts: cfg.MaxAttempts,
		logger:      cfg.Logger,
	}, nil
}

// Run drains immediately, then polls until ctx is canceled. Database errors
// are logged and retried on the next tick because a notification outage must
// not stop future deliveries permanently.
func (w *DeliveryWorker) Run(ctx context.Context) {
	if ctx == nil {
		ctx = context.Background()
	}

	w.drain(ctx)
	ticker := time.NewTicker(w.poll)
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

func (w *DeliveryWorker) drain(ctx context.Context) {
	if ctx.Err() != nil {
		return
	}
	deliveries, err := w.repository.ClaimNotificationDeliveries(ctx, w.batchSize)
	if err != nil {
		if ctx.Err() == nil {
			w.logger.Error("claim notification deliveries", "error", err)
		}
		return
	}

	for _, delivery := range deliveries {
		if ctx.Err() != nil {
			return
		}
		w.deliver(ctx, delivery)
	}
}

func (w *DeliveryWorker) deliver(ctx context.Context, claimed store.ClaimNotificationDeliveriesRow) {
	delivery, err := w.repository.GetNotificationDelivery(ctx, claimed.ID)
	if err != nil {
		w.retry(ctx, claimed, fmt.Errorf("load notification delivery: %w", err))
		return
	}
	if delivery.Endpoint == nil || delivery.P256dh == nil || delivery.Auth == nil || delivery.Enabled == nil || !*delivery.Enabled {
		w.skip(ctx, claimed, errors.New("push subscription is missing or disabled"))
		return
	}

	message, err := w.repository.GetNotificationMessage(ctx, delivery.MessageID)
	if err != nil {
		w.skip(ctx, claimed, fmt.Errorf("notification message is unavailable: %w", err))
		return
	}

	err = w.provider.Send(ctx, Subscription{
		ID:       delivery.SubscriptionID,
		Endpoint: *delivery.Endpoint,
		P256dh:   *delivery.P256dh,
		Auth:     *delivery.Auth,
	}, buildPayload(delivery.MessageID, message))
	if err == nil {
		if markErr := w.repository.MarkNotificationDeliverySent(ctx, claimed.ID); markErr != nil && ctx.Err() == nil {
			w.logger.Error("mark notification delivery sent", "delivery_id", claimed.ID, "error", markErr)
		}
		return
	}
	if ctx.Err() != nil {
		return
	}

	if IsPermanentError(err) {
		if disableErr := w.repository.DisablePushSubscriptionByEndpoint(ctx, *delivery.Endpoint); disableErr != nil {
			w.logger.Error("disable invalid push subscription", "subscription_id", delivery.SubscriptionID, "error", disableErr)
		}
		w.skip(ctx, claimed, err)
		return
	}
	w.retry(ctx, claimed, err)
}

func (w *DeliveryWorker) retry(ctx context.Context, delivery store.ClaimNotificationDeliveriesRow, deliveryErr error) {
	if ctx.Err() != nil {
		return
	}
	status := "pending"
	retryAfter := deliveryRetryDelaySeconds(delivery.Attempts)
	if delivery.Attempts >= w.maxAttempts {
		status = "failed"
		retryAfter = 0
	}
	w.markFailure(ctx, delivery, status, retryAfter, deliveryErr)
}

func (w *DeliveryWorker) skip(ctx context.Context, delivery store.ClaimNotificationDeliveriesRow, deliveryErr error) {
	if ctx.Err() != nil {
		return
	}
	w.markFailure(ctx, delivery, "skipped", 0, deliveryErr)
}

func (w *DeliveryWorker) markFailure(ctx context.Context, delivery store.ClaimNotificationDeliveriesRow, status string, retryAfter int32, deliveryErr error) {
	lastError := deliveryErr.Error()
	if err := w.repository.MarkNotificationDeliveryFailed(ctx, store.MarkNotificationDeliveryFailedParams{
		ID:               delivery.ID,
		Status:           status,
		RetryAfterSecond: retryAfter,
		LastError:        &lastError,
	}); err != nil {
		w.logger.Error("mark notification delivery failed", "delivery_id", delivery.ID, "error", err)
		return
	}
	w.logger.Warn("notification delivery did not send",
		"delivery_id", delivery.ID,
		"subscription_id", delivery.SubscriptionID,
		"attempts", delivery.Attempts,
		"status", status,
		"retry_after_seconds", retryAfter,
		"error", deliveryErr,
	)
}

func buildPayload(messageID int64, message store.GetNotificationMessageRow) Payload {
	title := "New message"
	if message.ConversationName != nil && strings.TrimSpace(*message.ConversationName) != "" {
		title = strings.TrimSpace(*message.ConversationName)
	}

	body := "You have a new message"
	if message.Content != nil && strings.TrimSpace(*message.Content) != "" {
		body = strings.TrimSpace(*message.Content)
		if message.SenderUsername != "" {
			body = message.SenderUsername + ": " + body
		}
		body = truncateRunes(body, maxNotificationBodyRunes)
	}

	return Payload{
		Title:     title,
		Body:      body,
		MessageID: messageID,
	}
}

func truncateRunes(value string, limit int) string {
	if utf8.RuneCountInString(value) <= limit {
		return value
	}
	runes := []rune(value)
	return string(runes[:limit-1]) + "…"
}

func deliveryRetryDelaySeconds(attempts int32) int32 {
	delay := initialDeliveryRetryDelay
	for attempt := int32(1); attempt < attempts; attempt++ {
		delay *= 2
		if delay >= maxDeliveryRetryDelay {
			return int32(maxDeliveryRetryDelay / time.Second)
		}
	}
	return int32(delay / time.Second)
}
