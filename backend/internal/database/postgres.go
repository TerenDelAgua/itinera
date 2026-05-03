package database

import (
	"backend/internal/config"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPostgress(cfg *config.Config) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse database config: %w", err)
	}

	pgxConfig.MaxConns = 10
	pgxConfig.MinConns = 2
	pgxConfig.MaxConnLifetime = 30 * time.Minute
	pgxConfig.MaxConnIdleTime = 5 * time.Minute

	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.NewWithConfig(context.Background(), pgxConfig)

		if err == nil {
			if err = pool.Ping(context.Background()); err == nil {
				log.Println("✅ Database connection established")
				return pool, nil
			}
		}
		log.Printf(" Database not ready (attempt %d/10), retrying in 2s...: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("Failed to connect to database after 10 attempts: %w", err)
}
