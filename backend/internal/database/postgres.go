package database

import (
	"backend/internal/config"
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)
func sanitizeDatabaseURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil || u.User == nil {
		return rawURL
	}
	pass, hasPass := u.User.Password()
	if !hasPass {
		return rawURL
	}
	
	u.User = url.UserPassword(u.User.Username(), url.PathEscape(pass))
	return u.String()
}

func NewPostgress(cfg *config.Config) (*pgxpool.Pool, error) {
	pgxConfig, err := pgxpool.ParseConfig(sanitizeDatabaseURL(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("Failed to parse database config: %w", err)
	}

	pgxConfig.MaxConns = 25
	pgxConfig.MinConns = 5
	pgxConfig.MaxConnLifetime = 30 * time.Minute
	pgxConfig.MaxConnIdleTime = 10 * time.Minute

	var pool *pgxpool.Pool
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.NewWithConfig(context.Background(), pgxConfig)

		if err == nil {
			if err = pool.Ping(context.Background()); err == nil {
				log.Println("✅ Database connection establshed")
				return pool, nil
			}
		}
		log.Printf(" Database not ready (attempt %d/10), retrying in 2s...: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	return nil, fmt.Errorf("Failed to connect to database after 10 attempts: %w", err)
}
