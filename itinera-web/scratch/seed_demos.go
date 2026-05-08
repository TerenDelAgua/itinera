package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	connStr := "postgres://user:password@localhost:5432/itinera?sslmode=disable"
	if env := os.Getenv("DATABASE_URL"); env != "" {
		connStr = env
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer pool.Close()

	// Check if demos already exist
	var count int
	err = pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM trips WHERE is_public_demo = true").Scan(&count)
	if err == nil && count > 0 {
		rows, _ := pool.Query(context.Background(), "SELECT name, is_public_demo FROM trips WHERE is_public_demo = true")
		log.Printf("Found %d public demos:", count)
		for rows.Next() {
			var name string
			var isPublic bool
			rows.Scan(&name, &isPublic)
			log.Printf("- %s (Public: %v)", name, isPublic)
		}
		return
	}

	fmt.Println("Seeding demo trips...")

	demos := []struct {
		Name        string
		Desc        string
		Days        int
		Currency    string
	}{
		{"Roma Express", "3 días · City break · EUR", 3, "EUR"},
		{"Japón Clásico", "15 días · Tour de país · JPY", 15, "JPY"},
		{"Asia Profunda", "60 días · Mochilero · USD", 60, "USD"},
	}

	for _, d := range demos {
		start := time.Now()
		end := start.AddDate(0, 0, d.Days)
		
		_, err := pool.Exec(context.Background(), `
			INSERT INTO trips (name, description, start_date, end_date, base_currency, is_public_demo, created_at)
			VALUES ($1, $2, $3, $4, $5, true, NOW())
		`, d.Name, d.Desc, start, end, d.Currency)
		
		if err != nil {
			log.Printf("Error seeding trip %s: %v\n", d.Name, err)
		} else {
			fmt.Printf("Seeded: %s\n", d.Name)
		}
	}

	// Update stats
	_, err = pool.Exec(context.Background(), "UPDATE public_stats SET total_trips_count = total_trips_count + 3 WHERE id = 1")
	if err != nil {
		log.Printf("Error updating stats: %v\n", err)
	}

	fmt.Println("Seeding complete.")
}
