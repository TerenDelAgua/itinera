package database

import (
	"backend/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewPostgress(cfg *config.Config) (*DB, error) {
	config, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse database config: %w", err)
	}

	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 30 * time.Minute
	config.MaxConnIdleTime = 10 * time.Minute

	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.NewWithConfig(context.Background(), config)

		if err == nil {
			if err = pool.Ping(context.Background()); err == nil {
				log.Println("✅ Database connection establshed")
				return &DB{Pool: pool}, nil
			}
		}
		log.Printf(" Database not ready (attempt %d/10), retrying in 2s...: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("Failed to connect to database after 10 attempts: %w", err)
}

func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("🔌 Database connection closed")
	}
}

func (db *DB) HealthCheck(ctx context.Context) error {
	return db.Pool.Ping(ctx)
}

func (db *DB) QueryRow(ctx context.Context, query string, args ...interface{}) interface {
	Scan(dest ...interface{}) error
} {
	return db.Pool.QueryRow(ctx, query, args)
}

func (db *DB) Query(ctx context.Context, query string, args ...interface{}) interface {
	Close()
	Err() error
	Next() bool
	Scan(dest ...interface{}) error
} {
	rows, _ := db.Pool.Query(ctx, query, args...)
	return rows
}
