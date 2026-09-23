package main

import (
	"testing"
	"time"
)

func validTestConfig() config {
	return config{
		port: 8084,
		db: dbConfig{
			dsn:      "postgres://localhost/chatify",
			maxConns: 10,
			minConns: 2,
		},
		cors:                  cors{trustedOrigins: []string{"http://localhost:3000"}},
		internalWebhookSecret: "test-secret",
		rabbitmq: rabbitmqConfig{
			url:          "amqp://guest:guest@localhost:5672/",
			exchange:     "chatify.events",
			pollInterval: time.Second,
			batchSize:    50,
		},
		notification: notificationConfig{
			queue:        "chatify.notifications",
			pollInterval: time.Second,
			batchSize:    50,
			maxAttempts:  5,
		},
	}
}

func TestValidateConfig(t *testing.T) {
	t.Setenv("CLERK_KEY", "test-clerk-key")

	if err := validateConfig(validTestConfig()); err != nil {
		t.Fatalf("validateConfig() error = %v", err)
	}
}

func TestValidateConfigRequiresWebhookSecret(t *testing.T) {
	t.Setenv("CLERK_KEY", "test-clerk-key")
	cfg := validTestConfig()
	cfg.internalWebhookSecret = ""

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig() error = nil, want missing webhook secret error")
	}
}

func TestValidateConfigRequiresVAPIDSettingsWhenNotificationsEnabled(t *testing.T) {
	t.Setenv("CLERK_KEY", "test-clerk-key")
	cfg := validTestConfig()
	cfg.notification.enabled = true

	if err := validateConfig(cfg); err == nil {
		t.Fatal("validateConfig() error = nil, want missing VAPID settings error")
	}
}
