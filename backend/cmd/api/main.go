package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nicolasleigh/chat-app/env"
	"github.com/nicolasleigh/chat-app/pg"
	"github.com/nicolasleigh/chat-app/store"
)

type config struct {
	port int
	db   dbConfig
	cors cors
}

type application struct {
	config config
	query  *store.Queries
	db     databasePinger
	// store store.Storage
}

type dbConfig struct {
	dsn      string
	maxConns int32
	minConns int32
}

type cors struct {
	trustedOrigins []string
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
		port: env.GetInt("PORT", 8084),
		db: dbConfig{
			dsn:      dsnEnv,
			maxConns: int32(env.GetInt("DB_MAX_CONNS", 10)),
			minConns: int32(env.GetInt("DB_MIN_CONNS", 2)),
		},
		cors: cors{
			trustedOrigins: []string{"http://localhost:3000", "https://chat.linze.pro"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pg.NewPG(ctx, cfg.db.dsn, pg.PoolConfig{
		MaxConns: cfg.db.maxConns,
		MinConns: cfg.db.minConns,
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
