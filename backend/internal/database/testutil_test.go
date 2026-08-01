package database_test

import (
	"context"
	"log"
	"testing"
	"time"

	"backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var testPool *pgxpool.Pool

func getTestPool(t *testing.T) *pgxpool.Pool {
	if testPool != nil {
		return testPool
	}

	// Cargar .env desde la raíz del backend (2 niveles arriba)
	_ = godotenv.Load("../../.env")

	cfg := config.Load()
	dbURL := cfg.DatabaseURL
	if dbURL == "" {
		t.Skip("Skipping integration test: DATABASE_URL not set")
	}

	pgxConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		t.Skipf("Skipping integration test: invalid DATABASE_URL: %v", err)
	}

	pgxConfig.MaxConns = 5
	pgxConfig.MinConns = 1
	pgxConfig.MaxConnLifetime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), pgxConfig)
	if err != nil {
		t.Skipf("Skipping integration test: cannot connect to database: %v", err)
	}

	if err = pool.Ping(context.Background()); err != nil {
		t.Skipf("Skipping integration test: database ping failed: %v", err)
	}

	testPool = pool
	log.Println("✅ Test Database connection established")
	return testPool
}
