package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port        string
	DatabaseURL string
	JWTSecret   string
	Environment string //Development or production
}

func Load() *Config {
	return &Config{
		Port:        getEnv("API_PORT", "8080"),
		DatabaseURL: getEnv("DATABASE_URL", "postgres://teren_dev:qwerty123@localhost:5432/itinera?sslmode=disable"),
		JWTSecret:   getEnv("JWT_SECRET", "dev-secret-change-me"),
		Environment: getEnv("ENVIRONMENT", "development"),
	}

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
