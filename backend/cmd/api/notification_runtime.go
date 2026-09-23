package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/nicolasleigh/chat-app/messaging/rabbitmq"
)

const notificationConsumerRestartDelay = 5 * time.Second

// runNotificationConsumer keeps the notification queue attached to RabbitMQ
// without making the HTTP server's lifetime depend on the broker's health.
// The consumer owns its AMQP connection, so a disconnect causes Consume to
// return and this supervisor can establish a fresh connection.
func runNotificationConsumer(ctx context.Context, consumer *rabbitmq.Consumer, handler rabbitmq.EventHandler) {
	for {
		if ctx.Err() != nil {
			return
		}

		if err := consumer.Consume(ctx, handler); err != nil && ctx.Err() == nil {
			slog.Error("notification rabbitmq consumer stopped", "error", err)
		}
		if ctx.Err() != nil {
			return
		}

		timer := time.NewTimer(notificationConsumerRestartDelay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
		}
	}
}
