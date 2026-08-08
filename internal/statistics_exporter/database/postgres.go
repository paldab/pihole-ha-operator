// Package database contain database functions
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConnConfig struct {
	Host     string
	Port     uint16
	User     string
	Password string
	Database string
}

func CreatePostgresConnString(cfg PostgresConnConfig) string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Database,
	)
}

func NewPostgresPool(ctx context.Context, cfg PostgresConnConfig) (*pgxpool.Pool, error) {
	connString := CreatePostgresConnString(cfg)
	pool, err := pgxpool.New(ctx, connString)

	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
