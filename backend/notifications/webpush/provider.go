// Package webpush implements notifications.Provider using browser Web Push.
package webpush

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	webpushclient "github.com/SherClockHolmes/webpush-go"
	"github.com/nicolasleigh/chat-app/notifications"
)

const (
	defaultTTL         = 60
	defaultHTTPTimeout = 10 * time.Second
	maxResponseBody    = 4096
)

// Config contains VAPID credentials and the HTTP transport used to reach
// browser push services. The private key must only come from a secret manager
// or environment variable and must never be sent to the browser.
type Config struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	Subject         string
	TTL             int
	HTTPClient      *http.Client
}

// Client encrypts each payload for the browser subscription and authenticates
// the request with VAPID. A single client can safely be reused by all delivery
// worker goroutines because the underlying HTTP client is concurrency-safe.
type Client struct {
	vapidPublicKey  string
	vapidPrivateKey string
	subject         string
	ttl             int
	httpClient      *http.Client
}

func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.VAPIDPublicKey) == "" {
		return nil, errors.New("web push VAPID public key is required")
	}
	if strings.TrimSpace(cfg.VAPIDPrivateKey) == "" {
		return nil, errors.New("web push VAPID private key is required")
	}
	if strings.TrimSpace(cfg.Subject) == "" {
		return nil, errors.New("web push VAPID subject is required")
	}
	if cfg.TTL <= 0 {
		cfg.TTL = defaultTTL
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return &Client{
		vapidPublicKey:  cfg.VAPIDPublicKey,
		vapidPrivateKey: cfg.VAPIDPrivateKey,
		subject:         cfg.Subject,
		ttl:             cfg.TTL,
		httpClient:      cfg.HTTPClient,
	}, nil
}

func (c *Client) Send(ctx context.Context, subscription notifications.Subscription, payload notifications.Payload) error {
	if strings.TrimSpace(subscription.Endpoint) == "" ||
		strings.TrimSpace(subscription.P256dh) == "" ||
		strings.TrimSpace(subscription.Auth) == "" {
		return notifications.NewPermanentError(errors.New("incomplete web push subscription"))
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return notifications.NewPermanentError(fmt.Errorf("encode web push payload: %w", err))
	}
	if ctx == nil {
		ctx = context.Background()
	}

	response, err := webpushclient.SendNotificationWithContext(ctx, body, &webpushclient.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpushclient.Keys{
			Auth:   subscription.Auth,
			P256dh: subscription.P256dh,
		},
	}, &webpushclient.Options{
		HTTPClient:      c.httpClient,
		Subscriber:      c.subject,
		TTL:             c.ttl,
		VAPIDPublicKey:  c.vapidPublicKey,
		VAPIDPrivateKey: c.vapidPrivateKey,
	})
	if err != nil {
		return notifications.NewTransientError(err)
	}
	if response == nil {
		return notifications.NewTransientError(errors.New("web push provider returned no response"))
	}
	defer response.Body.Close()

	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}

	responseBody, _ := io.ReadAll(io.LimitReader(response.Body, maxResponseBody))
	err = fmt.Errorf("web push provider returned status %d: %s", response.StatusCode, strings.TrimSpace(string(responseBody)))
	if isPermanentStatus(response.StatusCode) {
		return notifications.NewPermanentError(err)
	}
	return notifications.NewTransientError(err)
}

func isPermanentStatus(statusCode int) bool {
	if statusCode == http.StatusRequestTimeout || statusCode == http.StatusTooEarly || statusCode == http.StatusTooManyRequests {
		return false
	}
	return statusCode >= http.StatusBadRequest && statusCode < http.StatusInternalServerError
}

var _ notifications.Provider = (*Client)(nil)
