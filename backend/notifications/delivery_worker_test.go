package notifications

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/nicolasleigh/chat-app/store"
)

type recordingNotificationProvider struct {
	err   error
	calls []Subscription
}

func (p *recordingNotificationProvider) Send(_ context.Context, subscription Subscription, _ Payload) error {
	p.calls = append(p.calls, subscription)
	return p.err
}

type recordingDeliveryWorkerRepository struct {
	delivery     store.GetNotificationDeliveryRow
	message      store.GetNotificationMessageRow
	sentCount    int
	usedIDs      []int64
	failures     []store.MarkNotificationDeliveryFailedParams
	disabledURLs []string
}

func (r *recordingDeliveryWorkerRepository) ClaimNotificationDeliveries(context.Context, int32) ([]store.ClaimNotificationDeliveriesRow, error) {
	return nil, nil
}

func (r *recordingDeliveryWorkerRepository) GetNotificationDelivery(context.Context, int64) (store.GetNotificationDeliveryRow, error) {
	return r.delivery, nil
}

func (r *recordingDeliveryWorkerRepository) GetNotificationMessage(context.Context, int64) (store.GetNotificationMessageRow, error) {
	return r.message, nil
}

func (r *recordingDeliveryWorkerRepository) MarkNotificationDeliverySent(context.Context, int64) error {
	r.sentCount++
	return nil
}

func (r *recordingDeliveryWorkerRepository) MarkPushSubscriptionUsed(_ context.Context, id int64) error {
	r.usedIDs = append(r.usedIDs, id)
	return nil
}

func (r *recordingDeliveryWorkerRepository) MarkNotificationDeliveryFailed(_ context.Context, arg store.MarkNotificationDeliveryFailedParams) error {
	r.failures = append(r.failures, arg)
	return nil
}

func (r *recordingDeliveryWorkerRepository) DisablePushSubscriptionByEndpoint(_ context.Context, endpoint string) error {
	r.disabledURLs = append(r.disabledURLs, endpoint)
	return nil
}

func newTestDeliveryWorker(t *testing.T, repository DeliveryWorkerRepository, provider Provider, maxAttempts int32) *DeliveryWorker {
	t.Helper()
	worker, err := NewDeliveryWorker(repository, provider, DeliveryWorkerConfig{
		PollInterval: time.Second,
		BatchSize:    1,
		MaxAttempts:  maxAttempts,
		Logger:       slog.New(slog.NewTextHandler(io.Discard, nil)),
	})
	if err != nil {
		t.Fatalf("NewDeliveryWorker() error = %v", err)
	}
	return worker
}

func validWorkerRepository() *recordingDeliveryWorkerRepository {
	endpoint := "https://push.example.test/subscription/abc"
	p256dh := "public-key"
	auth := "auth-secret"
	enabled := true
	content := "hello"
	conversationName := "Project Chat"
	return &recordingDeliveryWorkerRepository{
		delivery: store.GetNotificationDeliveryRow{
			ID:             1,
			MessageID:      42,
			SubscriptionID: 7,
			Endpoint:       &endpoint,
			P256dh:         &p256dh,
			Auth:           &auth,
			Enabled:        &enabled,
		},
		message: store.GetNotificationMessageRow{
			ConversationID:   99,
			Content:          &content,
			SenderUsername:   "Alice",
			ConversationName: &conversationName,
		},
	}
}

func TestDeliveryWorkerMarksSuccessfulPushAndSubscriptionUse(t *testing.T) {
	repository := validWorkerRepository()
	provider := &recordingNotificationProvider{}
	worker := newTestDeliveryWorker(t, repository, provider, 3)

	worker.deliver(context.Background(), store.ClaimNotificationDeliveriesRow{ID: 1, Attempts: 1})

	if len(provider.calls) != 1 {
		t.Fatalf("provider calls = %d, want 1", len(provider.calls))
	}
	if repository.sentCount != 1 {
		t.Fatalf("sent count = %d, want 1", repository.sentCount)
	}
	if len(repository.usedIDs) != 1 || repository.usedIDs[0] != 7 {
		t.Fatalf("used subscription IDs = %v, want [7]", repository.usedIDs)
	}
	if len(repository.failures) != 0 || len(repository.disabledURLs) != 0 {
		t.Fatalf("successful delivery recorded failures=%v disabled=%v", repository.failures, repository.disabledURLs)
	}
}

func TestDeliveryWorkerDisablesPermanentProviderFailure(t *testing.T) {
	repository := validWorkerRepository()
	provider := &recordingNotificationProvider{err: NewPermanentError(errors.New("subscription expired"))}
	worker := newTestDeliveryWorker(t, repository, provider, 3)

	worker.deliver(context.Background(), store.ClaimNotificationDeliveriesRow{ID: 1, Attempts: 1})

	if len(repository.disabledURLs) != 1 || repository.disabledURLs[0] != *repository.delivery.Endpoint {
		t.Fatalf("disabled endpoints = %v, want [%q]", repository.disabledURLs, *repository.delivery.Endpoint)
	}
	if len(repository.failures) != 1 {
		t.Fatalf("failures = %d, want 1", len(repository.failures))
	}
	if got := repository.failures[0]; got.Status != "skipped" || got.RetryAfterSecond != 0 {
		t.Fatalf("failure = %+v, want skipped with no retry", got)
	}
}

func TestDeliveryWorkerExhaustsTransientRetries(t *testing.T) {
	repository := validWorkerRepository()
	provider := &recordingNotificationProvider{err: NewTransientError(errors.New("push service unavailable"))}
	worker := newTestDeliveryWorker(t, repository, provider, 2)

	worker.deliver(context.Background(), store.ClaimNotificationDeliveriesRow{ID: 1, Attempts: 2})

	if len(repository.failures) != 1 {
		t.Fatalf("failures = %d, want 1", len(repository.failures))
	}
	if got := repository.failures[0]; got.Status != "failed" || got.RetryAfterSecond != 0 {
		t.Fatalf("failure = %+v, want failed with no retry", got)
	}
	if len(repository.disabledURLs) != 0 {
		t.Fatalf("disabled endpoints = %v, want none for transient failure", repository.disabledURLs)
	}
}

func TestDeliveryRetryDelayUsesCappedExponentialBackoff(t *testing.T) {
	tests := []struct {
		attempts int32
		want     int32
	}{
		{attempts: 1, want: 1},
		{attempts: 2, want: 2},
		{attempts: 3, want: 4},
		{attempts: 10, want: 300},
	}

	for _, test := range tests {
		if got := deliveryRetryDelaySeconds(test.attempts); got != test.want {
			t.Errorf("deliveryRetryDelaySeconds(%d) = %d, want %d", test.attempts, got, test.want)
		}
	}
}

func TestBuildPayloadUsesConversationAndTruncatesBody(t *testing.T) {
	conversationName := "Project Chat"
	content := strings.Repeat("message ", 40)
	payload := buildPayload(42, store.GetNotificationMessageRow{
		ConversationID:   7,
		Content:          &content,
		SenderUsername:   "Alice",
		ConversationName: &conversationName,
	})

	if payload.Title != conversationName {
		t.Fatalf("title = %q, want %q", payload.Title, conversationName)
	}
	if payload.MessageID != 42 {
		t.Fatalf("message ID = %d, want 42", payload.MessageID)
	}
	if payload.ConversationID != 7 {
		t.Fatalf("conversation ID = %d, want 7", payload.ConversationID)
	}
	if len([]rune(payload.Body)) > maxNotificationBodyRunes {
		t.Fatalf("body has %d runes, want at most %d", len([]rune(payload.Body)), maxNotificationBodyRunes)
	}
	if !containsRune(payload.Body, '…') {
		t.Fatalf("body = %q, want truncation marker", payload.Body)
	}
}

func containsRune(value string, target rune) bool {
	for _, current := range value {
		if current == target {
			return true
		}
	}
	return false
}
