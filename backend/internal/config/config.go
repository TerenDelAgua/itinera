package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment string
}

func Load() *Config {
	return &Config{
		Port:        getEnv("API_PORT", "8080"),
		DatabaseURL: buildDatabaseURL(),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

// buildDatabaseURL constructs the connection string with the following priority:
//  1. Individual DB_* vars (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE)
//     → uses DSN key-value format, safe for any password (no URL encoding needed)
//  2. DATABASE_URL env var (full URL — may need URL-encoding if password has special chars)
//  3. Local dev defaults
func buildDatabaseURL() string {
	if host := os.Getenv("DB_HOST"); host != "" {
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := os.Getenv("DB_PASSWORD") // no default — may be empty in some setups
		dbname := getEnv("DB_NAME", "postgres")
		sslmode := getEnv("DB_SSLMODE", "require")

		// DSN key-value format: special characters in password are safe here
		return fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode,
		)
	}

	// Fallback: DATABASE_URL (password must be URL-encoded if it contains special chars)
	return getEnv("DATABASE_URL", "postgres://teren_dev:qwerty123@localhost:5432/itinera?sslmode=disable")
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

func (c *Config) LogSummary() {
	fmt.Printf("🔧 Config loaded: port=%s, env=%s, db_url=***\n", c.Port, c.Environment)
}
