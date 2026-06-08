package main

import (
	"context"
	"fmt"
	"log"

	"backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	cfg := config.Load()
	
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	queries := []string{
		`UPDATE trips SET name = 'inspiration.roma.title', description = 'inspiration.roma.desc' WHERE id = '550e8400-e29b-41d4-a716-446655440001'`,
		`UPDATE trips SET name = 'inspiration.japan.title', description = 'inspiration.japan.desc' WHERE id = '550e8400-e29b-41d4-a716-446655440002'`,
		`UPDATE trips SET name = 'inspiration.asia.title', description = 'inspiration.asia.desc' WHERE id = '550e8400-e29b-41d4-a716-446655440003'`,
	}

	for _, q := range queries {
		_, err = pool.Exec(context.Background(), q)
		if err != nil {
			log.Fatalf("Query failed: %v", err)
		}
	}

	fmt.Println("Trips updated successfully on the database!")
}
