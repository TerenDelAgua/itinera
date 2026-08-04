package email

import (
	"backend/internal/models"
	"context"
	"log"
	"strings"
)

// LogSender renders every email into the supplied logger. The default
// constructor points at log.Default(); the tests use a buffer logger so
// they can assert the exact rendered body without touching the network.
//
// This is the implementation that runs in CI and staging. It is also the
// fallback for production deployments that haven't configured a Resend
// API key yet — better a logged email than no email at all.
type LogSender struct {
	Log *log.Logger
}

// NewLogSender returns a sender that writes to log.Default().
func NewLogSender() *LogSender {
	return &LogSender{Log: log.Default()}
}

// Compile-time guard that *LogSender satisfies Sender.
var _ Sender = (*LogSender)(nil)

func (s *LogSender) logger() *log.Logger {
	if s.Log != nil {
		return s.Log
	}
	return log.Default()
}

func (s *LogSender) SendWelcome(_ context.Context, user models.User, locale string) error {
	tpl := renderWelcome(user.Email, locale)
	s.logger().Printf("[email:welcome] to=%s locale=%s subject=%s body=%s",
		user.Email, locale, tpl.subject, oneLine(tpl.text))
	return nil
}

func (s *LogSender) SendPasswordReset(_ context.Context, user models.User, code, locale string) error {
	tpl := renderReset(code, locale)
	s.logger().Printf("[email:reset] to=%s locale=%s code=%s subject=%s body=%s",
		user.Email, locale, code, tpl.subject, oneLine(tpl.text))
	return nil
}

// oneLine collapses the template into a single line so the log entry
// stays greppable. The HTML version intentionally keeps its newlines
// in the log because re-formatting HTML to one line adds little.
func oneLine(s string) string {
	return strings.ReplaceAll(strings.TrimSpace(s), "\n", " | ")
}