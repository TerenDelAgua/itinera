// Package email centralises the dispatch of transactional emails so handlers
// never import the Resend SDK directly. Three implementations share the
// same `Sender` interface so tests run deterministically and production
// sends via Resend's HTTP API
package email

import (
	"backend/internal/config"
	"backend/internal/models"
	"context"
)

// Sender is the single contract handlers depend on. Implementations:
//   - ResendSender: posts to api.resend.com (production)
//   - NoopSender:  silently succeeds (development hot-loop)
//   - LogSender:   renders the email body to a *log.Logger (CI + unit tests)
type Sender interface {
	SendWelcome(ctx context.Context, user models.User, locale string) error
	SendPasswordReset(ctx context.Context, user models.User, code, locale string) error
}

// NewSender chooses the implementation based on environment:
//
//	production + RESEND_API_KEY set → ResendSender
//	development                     → NoopSender
//	otherwise (staging, CI, test)   → LogSender
//
// The factory is the ONLY place that decides which implementation wins,
// so a single change here flips the entire app without editing handlers.
func NewSender(cfg *config.Config) Sender {
	if cfg.Environment == "production" && cfg.ResendAPIKey != "" {
		return NewResendSender(cfg.ResendAPIKey, cfg.EmailFrom)
	}
	if cfg.Environment == "development" {
		return NoopSender{}
	}
	return NewLogSender()
}
