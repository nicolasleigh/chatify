package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	DB *pgxpool.Pool
}

// https://donchev.is/post/working-with-postgresql-in-go-using-pgx/
func NewPG(ctx context.Context, connString string) (*postgres, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database configuration: %w", err)
	}

	db, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	return &postgres{DB: db}, nil

}

func (pg *postgres) Ping(ctx context.Context) error {
	return pg.DB.Ping(ctx)
}

func (pg *postgres) Close() {
	pg.DB.Close()
}
