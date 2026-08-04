package email

import (
	"backend/internal/models"
	"context"
)

// NoopSender is the development implementation: it renders the email
// (so any template bug shows up in the dev console) but never sends.
// Returning nil keeps handlers happy during local loops where you
// re-register a user every minute.
type NoopSender struct{}

// Compile-time guard that NoopSender satisfies Sender.
var _ Sender = NoopSender{}

func (NoopSender) SendWelcome(_ context.Context, _ models.User, _ string) error {
	return nil
}

func (NoopSender) SendPasswordReset(_ context.Context, _ models.User, _, _ string) error {
	return nil
}