package pg

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type postgres struct {
	DB *pgxpool.Pool
}

type PoolConfig struct {
	MaxConns int32
	MinConns int32
}

// https://donchev.is/post/working-with-postgresql-in-go-using-pgx/
func NewPG(ctx context.Context, connString string, poolConfig PoolConfig) (*postgres, error) {
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database configuration: %w", err)
	}
	if poolConfig.MaxConns > 0 {
		config.MaxConns = poolConfig.MaxConns
	}
	if poolConfig.MinConns > 0 {
		config.MinConns = poolConfig.MinConns
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
