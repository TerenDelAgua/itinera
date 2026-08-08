package email

import (
	"backend/internal/models"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// ResendSender posts transactional emails to api.resend.com. Production
// uses this; tests and dev use the noop/log implementations so the
// integration never has to deal with network flakiness.
type ResendSender struct {
	APIKey string
	From   string

	// HTTPClient is overridable so tests can wire httptest.Server. The
	// production path uses http.DefaultClient via NewResendSender.
	HTTPClient *http.Client
}

// NewResendSender returns a sender ready for production traffic.
func NewResendSender(apiKey, from string) *ResendSender {
	return &ResendSender{
		APIKey:     apiKey,
		From:       from,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
	}
}

// resendPayload is the JSON shape Resend expects on POST /v1/emails.
// The struct lives here (not exported) because no caller should construct
// it directly — they call SendWelcome / SendPasswordReset instead.
type resendPayload struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Text    string   `json:"text,omitempty"`
	HTML    string   `json:"html,omitempty"`
}

func (r *ResendSender) SendWelcome(ctx context.Context, user models.User, locale string) error {
	tpl := renderWelcome(user.Email, locale)
	return r.send(ctx, []string{user.Email}, tpl.subject, tpl.text, tpl.html)
}

func (r *ResendSender) SendPasswordReset(ctx context.Context, user models.User, code, locale string) error {
	tpl := renderReset(code, locale)
	return r.send(ctx, []string{user.Email}, tpl.subject, tpl.text, tpl.html)
}

// send POSTs the payload to Resend. Errors from the network or the API
// are returned verbatim (no echo to the user; that's the handler's job).
func (r *ResendSender) send(ctx context.Context, to []string, subject, text, html string) error {
	if r.APIKey == "" {
		return fmt.Errorf("resend sender: APIKey is empty")
	}
	body, err := json.Marshal(resendPayload{
		From:    r.From,
		To:      to,
		Subject: subject,
		Text:    text,
		HTML:    html,
	})
	if err != nil {
		return fmt.Errorf("marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.resend.com/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("new request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+r.APIKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := r.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("post to resend: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	// Read up to 4 KiB of the error body to give operators something
	// to grep in logs without dumping a multi-MB JSON error envelope.
	buf, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	return fmt.Errorf("resend returned %d: %s", resp.StatusCode, string(buf))
}