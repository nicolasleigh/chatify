package main

import (
	"context"
	"log/slog"
	"os"
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
	// store store.Storage
}

type dbConfig struct {
	dsn string
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
			// dsn: env.GetString("DB_DSN", "postgres://admin:adminpassword@localhost:5432/chat?sslmode=disable"),
			dsn: dsnEnv,
		},
		cors: cors{
			trustedOrigins: []string{"http://localhost:3000", "https://chat.linze.pro"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := pg.NewPG(ctx, cfg.db.dsn)
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
		// store: store,
	}

	srv := app.NewServer()
	err = srv.ListenAndServe()
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
