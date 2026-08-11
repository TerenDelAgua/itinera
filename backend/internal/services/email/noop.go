package email

import (
	"backend/internal/models"
	"context"
	"log"
)

// NoopSender is the development implementation: it renders the email
// (so any template bug shows up in the dev console) and logs the
// 6-digit reset code at Info level so a developer running the local
// backend can copy-paste it into /forgot without configuring Resend.
//
// It NEVER sends the email. Returning nil keeps handlers happy
// during local loops where you re-register a user every minute.
type NoopSender struct{}

// Compile-time guard that NoopSender satisfies Sender.
var _ Sender = NoopSender{}

func (NoopSender) SendWelcome(_ context.Context, user models.User, locale string) error {
	tpl := renderWelcome(user.Email, locale)
	log.Printf("[email:welcome] to=%s locale=%s subject=%s body=%s",
		user.Email, locale, tpl.subject, oneLine(tpl.text))
	return nil
}

func (NoopSender) SendPasswordReset(_ context.Context, user models.User, code, locale string) error {
	tpl := renderReset(code, locale)
	log.Printf("[email:reset] to=%s locale=%s code=%s subject=%s body=%s",
		user.Email, locale, code, tpl.subject, oneLine(tpl.text))
	return nil
}