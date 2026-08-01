package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

const insecureDevJWTSecret = "dev-secret-change-me"
const minProductionSecretLen = 32

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	Environment  string
	PublicOrigin string
}

func Load() *Config {
	return &Config{
		Port:         getEnv("API_PORT", "8080"),
		DatabaseURL:  buildDatabaseURL(),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-me"),
		Environment:  getEnv("ENVIRONMENT", "development"),
		PublicOrigin: normalizeOrigin(getEnv("PUBLIC_ORIGIN", "https://goitinera.app")),
	}
}

// normalizeOrigin strips trailing slashes so URLs built from origin + path don't end up with `//share/abc`.
func normalizeOrigin(origin string) string {
	return strings.TrimRight(origin, "/")
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

// BuildShareURL constructs the absolute public URL for a share token,
// e.g. "https://goitinera.app/share/<token>". Centralised here so that:
//   - the origin is read from a single config value (PUBLIC_ORIGIN env),
//   - the path "/share/" is not duplicated across handlers,
//   - future share-targets (e.g. "/share/<token>/fork") can reuse it.
func (c *Config) BuildShareURL(token string) string {
	return c.PublicOrigin + "/share/" + token
}

// Validate enforces invariants that must hold for the configured environment.
// It is intentionally strict in production so a missing or weak JWT_SECRET
// fails fast at boot instead of silently accepting attacker-forged tokens.
func (c *Config) Validate() error {
	if !c.IsProduction() {
		return nil
	}
	if c.JWTSecret == "" || c.JWTSecret == insecureDevJWTSecret {
		return errors.New("JWT_SECRET must be set to a non-default value in production")
	}
	if len(c.JWTSecret) < minProductionSecretLen {
		return fmt.Errorf("JWT_SECRET must be at least %d characters in production (got %d)", minProductionSecretLen, len(c.JWTSecret))
	}
	if strings.TrimSpace(c.JWTSecret) != c.JWTSecret {
		return errors.New("JWT_SECRET must not contain leading or trailing whitespace")
	}
	return nil
}
