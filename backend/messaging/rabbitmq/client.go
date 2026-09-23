// Package rabbitmq adapts the broker-independent messaging contract to
// RabbitMQ. Keeping this code behind messaging.Publisher means handlers and
// the outbox worker do not need to know about AMQP channels or deliveries.
package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/nicolasleigh/chat-app/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultConfirmTimeout = 10 * time.Second
	exchangeTypeTopic     = "topic"
)

// Config contains the small set of topology and delivery settings owned by
// this adapter. Credentials stay inside the AMQP URL so they can be supplied
// by the runtime secret manager instead of being compiled into the service.
type Config struct {
	URL                   string
	EventsExchange        string
	PublishConfirmTimeout time.Duration
}

// Client is a long-lived RabbitMQ publisher. One channel is deliberately
// shared by the outbox worker because the worker publishes one claimed event
// at a time; the mutex also makes accidental concurrent callers safe.
type Client struct {
	mu                    sync.Mutex
	connection            *amqp.Connection
	channel               *amqp.Channel
	eventsExchange        string
	publishConfirmTimeout time.Duration
	closed                bool
}

// New connects to RabbitMQ and declares the durable exchange used for
// application events. Startup fails if either operation fails, because
// silently accepting events without a broker would leave the service looking
// healthy while its asynchronous integration is disabled.
func New(cfg Config) (*Client, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("rabbitmq URL is required")
	}
	if strings.TrimSpace(cfg.EventsExchange) == "" {
		return nil, errors.New("rabbitmq events exchange is required")
	}
	if cfg.PublishConfirmTimeout <= 0 {
		cfg.PublishConfirmTimeout = defaultConfirmTimeout
	}

	connection, err := amqp.DialConfig(cfg.URL, amqp.Config{
		Properties: amqp.NewConnectionProperties(),
		Locale:     "en_US",
	})
	if err != nil {
		return nil, fmt.Errorf("connect to rabbitmq: %w", err)
	}

	channel, err := connection.Channel()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("open rabbitmq channel: %w", err)
	}

	// A durable topic exchange survives broker restarts. Routing keys are event
	// types such as message.created, allowing future consumers to subscribe to
	// one event family or to a wildcard without changing producers.
	if err := channel.ExchangeDeclare(
		cfg.EventsExchange,
		exchangeTypeTopic,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("declare rabbitmq events exchange: %w", err)
	}

	// Deferred confirmations correlate each publish with the broker's ack/nack.
	// The buffered listener is drained continuously because the AMQP client
	// dispatches confirmations to it and can otherwise block a channel when
	// more than one publish is in flight.
	confirmations := channel.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := channel.Confirm(false); err != nil {
		_ = channel.Close()
		_ = connection.Close()
		return nil, fmt.Errorf("enable rabbitmq publisher confirms: %w", err)
	}
	go drainConfirmations(confirmations)

	return &Client{
		connection:            connection,
		channel:               channel,
		eventsExchange:        cfg.EventsExchange,
		publishConfirmTimeout: cfg.PublishConfirmTimeout,
	}, nil
}

func drainConfirmations(confirmations <-chan amqp.Confirmation) {
	for range confirmations {
		// DeferredConfirmation.WaitContext observes the same confirmation. This
		// goroutine only prevents the library's listener from back-pressuring
		// the channel; it intentionally does not make delivery decisions.
	}
}

// Publish serializes a validated envelope as JSON and waits for RabbitMQ's
// publisher confirmation. Returning only after an ack is important: the
// outbox worker marks the database row published immediately after this method
// succeeds, so an early return would create a permanent delivery gap.
func (c *Client) Publish(ctx context.Context, envelope messaging.Envelope) error {
	if err := envelope.Validate(); err != nil {
		return fmt.Errorf("validate event envelope: %w", err)
	}
	if ctx == nil {
		ctx = context.Background()
	}

	body, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode event envelope: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return errors.New("rabbitmq publisher is closed")
	}
	publishContext := ctx
	if _, hasDeadline := ctx.Deadline(); !hasDeadline && c.publishConfirmTimeout > 0 {
		var cancel context.CancelFunc
		publishContext, cancel = context.WithTimeout(ctx, c.publishConfirmTimeout)
		defer cancel()
	}

	confirmation, err := c.channel.PublishWithDeferredConfirmWithContext(
		publishContext,
		c.eventsExchange,
		envelope.Type,
		true,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			MessageId:    envelope.ID,
			Type:         envelope.Type,
			Timestamp:    envelope.OccurredAt,
			Headers: amqp.Table{
				"x-event-version": int32(envelope.Version),
			},
			Body: body,
		},
	)
	if err != nil {
		return fmt.Errorf("publish event to rabbitmq: %w", err)
	}
	if confirmation == nil {
		return errors.New("rabbitmq publisher confirmation was not created")
	}

	acked, err := confirmation.WaitContext(publishContext)
	if err != nil {
		return fmt.Errorf("wait for rabbitmq publisher confirmation: %w", err)
	}
	if !acked || !confirmation.Acked() {
		return errors.New("rabbitmq rejected event publish")
	}

	return nil
}

// Close stops new publishes and closes the AMQP resources. The operation is
// idempotent so shutdown paths can safely defer it and also call it after a
// failed startup cleanup.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.closed = true

	channelErr := c.channel.Close()
	connectionErr := c.connection.Close()
	if errors.Is(channelErr, amqp.ErrClosed) {
		channelErr = nil
	}
	if errors.Is(connectionErr, amqp.ErrClosed) {
		connectionErr = nil
	}
	return errors.Join(channelErr, connectionErr)
}

var _ messaging.Publisher = (*Client)(nil)
