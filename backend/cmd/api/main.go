package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nicolasleigh/chat-app/env"
	"github.com/nicolasleigh/chat-app/messaging/rabbitmq"
	"github.com/nicolasleigh/chat-app/outbox"
	"github.com/nicolasleigh/chat-app/pg"
	"github.com/nicolasleigh/chat-app/store"
)

type config struct {
	port                  int
	db                    dbConfig
	cors                  cors
	internalWebhookSecret string
	rabbitmq              rabbitmqConfig
	notification          notificationConfig
}

type application struct {
	config config
	query  *store.Queries
	db     databasePinger
	// store store.Storage
}

type dbConfig struct {
	dsn               string
	maxConns          int32
	minConns          int32
	maxConnLifetime   time.Duration
	maxConnIdleTime   time.Duration
	healthCheckPeriod time.Duration
}

type cors struct {
	trustedOrigins []string
}

type rabbitmqConfig struct {
	url          string
	exchange     string
	pollInterval time.Duration
	batchSize    int32
}

// notificationConfig contains the future-facing delivery settings for
// offline browser notifications. The feature is opt-in until the subscription
// API and Web Push provider are installed, so existing deployments can roll
// out the data model and consumers independently.
type notificationConfig struct {
	enabled         bool
	queue           string
	pollInterval    time.Duration
	batchSize       int32
	maxAttempts     int32
	vapidPublicKey  string
	vapidPrivateKey string
	vapidSubject    string
}

var (
	Validate *validator.Validate
	NewLog   *slog.Logger
)

func main() {
	// Logger
	NewLog = slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(NewLog)

	Validate = validator.New(validator.WithRequiredStructEnabled())
	var dsnEnv string
	if os.Getenv("APP_ENV") == "production" {
		dsnEnv = os.Getenv("CLOUD_DB_DSN")
	} else {
		dsnEnv = os.Getenv("DB_DSN")
	}

	cfg := config{
		port:                  env.GetInt("PORT", 8084),
		internalWebhookSecret: env.GetString("INTERNAL_WEBHOOK_SECRET", ""),
		db: dbConfig{
			dsn:               dsnEnv,
			maxConns:          int32(env.GetInt("DB_MAX_CONNS", 10)),
			minConns:          int32(env.GetInt("DB_MIN_CONNS", 2)),
			maxConnLifetime:   time.Duration(env.GetInt("DB_MAX_CONN_LIFETIME_SECONDS", 1800)) * time.Second,
			maxConnIdleTime:   time.Duration(env.GetInt("DB_MAX_CONN_IDLE_TIME_SECONDS", 300)) * time.Second,
			healthCheckPeriod: time.Duration(env.GetInt("DB_HEALTH_CHECK_PERIOD_SECONDS", 60)) * time.Second,
		},
		cors: cors{
			trustedOrigins: parseTrustedOrigins(env.GetString(
				"CORS_TRUSTED_ORIGINS",
				"http://localhost:3000,https://chat.linze.pro",
			)),
		},
		rabbitmq: rabbitmqConfig{
			url:          env.GetString("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
			exchange:     env.GetString("RABBITMQ_EVENTS_EXCHANGE", "chatify.events"),
			pollInterval: time.Duration(env.GetInt("OUTBOX_POLL_INTERVAL_SECONDS", 1)) * time.Second,
			batchSize:    int32(env.GetInt("OUTBOX_BATCH_SIZE", 50)),
		},
		notification: notificationConfig{
			enabled:         env.GetBool("NOTIFICATION_ENABLED", false),
			queue:           env.GetString("NOTIFICATION_CONSUMER_QUEUE", "chatify.notifications"),
			pollInterval:    time.Duration(env.GetInt("NOTIFICATION_POLL_INTERVAL_SECONDS", 1)) * time.Second,
			batchSize:       int32(env.GetInt("NOTIFICATION_BATCH_SIZE", 50)),
			maxAttempts:     int32(env.GetInt("NOTIFICATION_RETRY_LIMIT", 5)),
			vapidPublicKey:  env.GetString("WEB_PUSH_VAPID_PUBLIC_KEY", ""),
			vapidPrivateKey: env.GetString("WEB_PUSH_VAPID_PRIVATE_KEY", ""),
			vapidSubject:    env.GetString("WEB_PUSH_VAPID_SUBJECT", ""),
		},
	}
	if err := validateConfig(cfg); err != nil {
		slog.Error("invalid application configuration", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pg.NewPG(ctx, cfg.db.dsn, pg.PoolConfig{
		MaxConns:          cfg.db.maxConns,
		MinConns:          cfg.db.minConns,
		MaxConnLifetime:   cfg.db.maxConnLifetime,
		MaxConnIdleTime:   cfg.db.maxConnIdleTime,
		HealthCheckPeriod: cfg.db.healthCheckPeriod,
	})
	if err != nil {
		slog.Error("database connection pool initialization failed", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	err = db.Ping(ctx)
	if err != nil {
		slog.Error("database ping failed", "error", err)
		os.Exit(1)
	}

	q := store.New(db.DB)

	slog.Info("database connection pool established!")

	publisher, err := rabbitmq.New(rabbitmq.Config{
		URL:            cfg.rabbitmq.url,
		EventsExchange: cfg.rabbitmq.exchange,
	})
	if err != nil {
		slog.Error("rabbitmq publisher initialization failed", "error", err)
		os.Exit(1)
	}

	worker, err := outbox.New(q, publisher, outbox.Config{
		PollInterval: cfg.rabbitmq.pollInterval,
		BatchSize:    cfg.rabbitmq.batchSize,
		Logger:       NewLog,
	})
	if err != nil {
		_ = publisher.Close()
		slog.Error("outbox worker initialization failed", "error", err)
		os.Exit(1)
	}
	workerCtx, stopWorker := context.WithCancel(context.Background())
	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		worker.Run(workerCtx)
	}()
	defer func() {
		// Stop claiming new rows first and wait for an in-flight publish to
		// finish or observe cancellation before closing its AMQP channel.
		stopWorker()
		<-workerDone
		if err := publisher.Close(); err != nil {
			slog.Error("rabbitmq publisher shutdown failed", "error", err)
		}
	}()

	app := &application{
		config: cfg,
		query:  q,
		db:     db.DB,
		// store: store,
	}

	srv := app.NewServer()
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- srv.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err = <-serverErr:
		if !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server stopped unexpectedly", "error", err)
		}
	case <-shutdownSignal.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("http server graceful shutdown failed", "error", err)
		}
	}
}

func parseTrustedOrigins(raw string) []string {
	parts := strings.Split(raw, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if origin := strings.TrimSpace(part); origin != "" {
			origins = append(origins, origin)
		}
	}
	return origins
}
