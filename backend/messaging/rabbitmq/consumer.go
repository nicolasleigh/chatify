package rabbitmq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nicolasleigh/chat-app/messaging"
	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	defaultNotificationRoutingKey = messaging.MessageCreated
	defaultConsumerRetryDelay     = time.Second
	defaultConsumerPrefetch       = 10
)

// EventHandler processes one decoded event. The handler should return a
// messaging.PermanentError for poison messages that should be acknowledged and
// discarded; all other errors are treated as transient and requeued.
type EventHandler func(context.Context, messaging.Envelope) error

type ConsumerConfig struct {
	URL         string
	Exchange    string
	Queue       string
	RoutingKey  string
	ConsumerTag string
	RetryDelay  time.Duration
	Prefetch    int
}

// Consumer owns a dedicated AMQP connection because consuming and publishing
// on separate connections prevents a slow handler from affecting Outbox
// publisher confirms.
type Consumer struct {
	config ConsumerConfig
}

func NewConsumer(cfg ConsumerConfig) (*Consumer, error) {
	if strings.TrimSpace(cfg.URL) == "" {
		return nil, errors.New("rabbitmq consumer URL is required")
	}
	if strings.TrimSpace(cfg.Exchange) == "" {
		return nil, errors.New("rabbitmq consumer exchange is required")
	}
	if strings.TrimSpace(cfg.Queue) == "" {
		return nil, errors.New("rabbitmq consumer queue is required")
	}
	if strings.TrimSpace(cfg.RoutingKey) == "" {
		cfg.RoutingKey = defaultNotificationRoutingKey
	}
	if strings.TrimSpace(cfg.ConsumerTag) == "" {
		cfg.ConsumerTag = "chatify-notifications"
	}
	if cfg.RetryDelay <= 0 {
		cfg.RetryDelay = defaultConsumerRetryDelay
	}
	if cfg.Prefetch < 1 {
		cfg.Prefetch = defaultConsumerPrefetch
	}
	return &Consumer{config: cfg}, nil
}

// Consume declares the durable notification queue and processes deliveries
// until the context is canceled or RabbitMQ closes the channel. It ACKs only
// after the handler has durably planned the notification, so a database outage
// causes a requeue rather than silently losing the event.
func (c *Consumer) Consume(ctx context.Context, handler EventHandler) error {
	if handler == nil {
		return errors.New("rabbitmq consumer handler is required")
	}
	if ctx == nil {
		ctx = context.Background()
	}

	connection, err := amqp.DialConfig(c.config.URL, amqp.Config{
		Properties: amqp.NewConnectionProperties(),
		Locale:     "en_US",
	})
	if err != nil {
		return fmt.Errorf("connect notification consumer to rabbitmq: %w", err)
	}
	defer connection.Close()

	channel, err := connection.Channel()
	if err != nil {
		return fmt.Errorf("open notification consumer channel: %w", err)
	}
	defer channel.Close()

	if err := channel.ExchangeDeclare(c.config.Exchange, exchangeTypeTopic, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare notification exchange: %w", err)
	}
	queue, err := channel.QueueDeclare(c.config.Queue, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare notification queue: %w", err)
	}
	if err := channel.QueueBind(queue.Name, c.config.RoutingKey, c.config.Exchange, false, nil); err != nil {
		return fmt.Errorf("bind notification queue: %w", err)
	}
	if err := channel.Qos(c.config.Prefetch, 0, false); err != nil {
		return fmt.Errorf("configure notification consumer prefetch: %w", err)
	}

	deliveries, err := channel.Consume(queue.Name, c.config.ConsumerTag, false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume notification queue: %w", err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case delivery, ok := <-deliveries:
			if !ok {
				return errors.New("rabbitmq notification delivery channel closed")
			}
			if err := c.handleDelivery(ctx, delivery, handler); err != nil {
				return err
			}
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, delivery amqp.Delivery, handler EventHandler) error {
	envelope, err := decodeEnvelope(delivery.Body)
	if err != nil {
		// The message cannot become valid by retrying. ACK it after decoding
		// fails to prevent a poison message from blocking all notifications.
		if ackErr := delivery.Ack(false); ackErr != nil {
			return fmt.Errorf("ack invalid notification event: %w", ackErr)
		}
		return nil
	}

	if err := handler(ctx, envelope); err != nil {
		if messaging.IsPermanentError(err) {
			if ackErr := delivery.Ack(false); ackErr != nil {
				return fmt.Errorf("ack permanent notification event: %w", ackErr)
			}
			return nil
		}
		if err := waitForRetry(ctx, c.config.RetryDelay); err != nil {
			return nil
		}
		if nackErr := delivery.Nack(false, true); nackErr != nil {
			return fmt.Errorf("requeue notification event: %w", nackErr)
		}
		return nil
	}

	if err := delivery.Ack(false); err != nil {
		return fmt.Errorf("ack notification event: %w", err)
	}
	return nil
}

func decodeEnvelope(body []byte) (messaging.Envelope, error) {
	var envelope messaging.Envelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return messaging.Envelope{}, messaging.NewPermanentError(fmt.Errorf("decode event envelope: %w", err))
	}
	if err := envelope.Validate(); err != nil {
		return messaging.Envelope{}, messaging.NewPermanentError(fmt.Errorf("validate event envelope: %w", err))
	}
	return envelope, nil
}

func waitForRetry(ctx context.Context, delay time.Duration) error {
	timer := time.NewTimer(delay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
