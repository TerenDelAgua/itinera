package config

import (
	"testing"
)

func TestConfig_Validate(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		expectError bool
	}{
		{
			name: "development environment skips validation",
			config: Config{
				Environment: "development",
				JWTSecret:   "short",
			},
			expectError: false,
		},
		{
			name: "production with valid secret",
			config: Config{
				Environment:   "production",
				JWTSecret:     "this-is-a-very-long-and-secure-jwt-secret-value-32-chars",
				InternalToken: "this-is-a-very-long-and-secure-internal-token-32-chars",
			},
			expectError: false,
		},
		{
			name: "production with empty secret",
			config: Config{
				Environment:   "production",
				JWTSecret:     "",
				InternalToken: "this-is-a-very-long-and-secure-internal-token-32-chars",
			},
			expectError: true,
		},
		{
			name: "production with default secret",
			config: Config{
				Environment:   "production",
				JWTSecret:     "dev-secret-change-me",
				InternalToken: "this-is-a-very-long-and-secure-internal-token-32-chars",
			},
			expectError: true,
		},
		{
			name: "production with short secret",
			config: Config{
				Environment:   "production",
				JWTSecret:     "too-short",
				InternalToken: "this-is-a-very-long-and-secure-internal-token-32-chars",
			},
			expectError: true,
		},
		{
			name: "production with spaces",
			config: Config{
				Environment:   "production",
				JWTSecret:     "  this-is-a-very-long-and-secure-jwt-secret-value-32-chars  ",
				InternalToken: "this-is-a-very-long-and-secure-internal-token-32-chars",
			},
			expectError: true,
		},
		{
			name: "production with missing internal token",
			config: Config{
				Environment: "production",
				JWTSecret:   "this-is-a-very-long-and-secure-jwt-secret-value-32-chars",
			},
			expectError: true,
		},
		{
			name: "production with short internal token",
			config: Config{
				Environment:   "production",
				JWTSecret:     "this-is-a-very-long-and-secure-jwt-secret-value-32-chars",
				InternalToken: "too-short",
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.expectError {
				t.Errorf("Config.Validate() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
