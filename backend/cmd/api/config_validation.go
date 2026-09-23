package main

import (
	"errors"
	"os"
	"strings"
)

func validateConfig(cfg config) error {
	if cfg.port < 1 || cfg.port > 65535 {
		return errors.New("PORT must be between 1 and 65535")
	}
	if strings.TrimSpace(cfg.db.dsn) == "" {
		return errors.New("database DSN is required")
	}
	if strings.TrimSpace(os.Getenv("CLERK_KEY")) == "" {
		return errors.New("CLERK_KEY is required")
	}
	if strings.TrimSpace(cfg.internalWebhookSecret) == "" {
		return errors.New("INTERNAL_WEBHOOK_SECRET is required")
	}
	if len(cfg.cors.trustedOrigins) == 0 {
		return errors.New("at least one trusted CORS origin is required")
	}
	if strings.TrimSpace(cfg.rabbitmq.url) == "" {
		return errors.New("RABBITMQ_URL is required")
	}
	if strings.TrimSpace(cfg.rabbitmq.exchange) == "" {
		return errors.New("RABBITMQ_EVENTS_EXCHANGE is required")
	}
	if cfg.rabbitmq.pollInterval <= 0 {
		return errors.New("OUTBOX_POLL_INTERVAL_SECONDS must be greater than zero")
	}
	if cfg.rabbitmq.batchSize < 1 {
		return errors.New("OUTBOX_BATCH_SIZE must be greater than zero")
	}
	if cfg.db.maxConns < 1 {
		return errors.New("DB_MAX_CONNS must be greater than zero")
	}
	if cfg.db.minConns < 0 || cfg.db.minConns > cfg.db.maxConns {
		return errors.New("DB_MIN_CONNS must be between zero and DB_MAX_CONNS")
	}

	return nil
}
