package email

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"backend/internal/config"
	"backend/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------- Templates -----------------------------------------------------

// TestNormaliseLocale_AllSupported pins the four valid codes; unknown
// values fall through to "en".
func TestNormaliseLocale_AllSupported(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"en", "en"},
		{"ES", "es"},
		{" ja ", "ja"},
		{"\tid", "id"},
		{"", "en"},
		{"fr", "en"},
		{"ENGLISH", "en"},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, normaliseLocale(c.in), "normaliseLocale(%q)", c.in)
	}
}

// TestRenderWelcome_Localised asserts each of the four locales emits the
// canonical subject. Bodies intentionally aren't asserted in full here —
// the LogSender test below covers the wire payload end-to-end.
func TestRenderWelcome_Localised(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"en": "Welcome to Itinera",
		"es": "Bienvenido a Itinera",
		"ja": "Itineraへようこそ",
		"id": "Selamat datang di Itinera",
	}
	for locale, want := range cases {
		assert.Equal(t, want, renderWelcome("u@y.z", locale).subject, "locale=%s", locale)
	}
}

// TestRenderReset_Localised asserts the four reset subjects and that the
// 6-digit code is rendered verbatim. We don't assert positioning in the
// HTML body — that's a styling concern — but we DO assert the text body
// has the code so copy-paste survives a roundtrip to mobile.
func TestRenderReset_Localised(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"en": "Reset your Itinera password",
		"es": "Restablece tu contraseña de Itinera",
		"ja": "Itineraのパスワードを再設定",
		"id": "Reset kata sandi Itinera Anda",
	}
	for locale, want := range cases {
		tpl := renderReset("482915", locale)
		assert.Equal(t, want, tpl.subject, "subject for %s", locale)
		assert.Contains(t, tpl.text, "482915", "text body must include the code for locale %s", locale)
		assert.Contains(t, tpl.html, "482915", "html body must include the code for locale %s", locale)
	}
}

// TestRender_UnknownLocaleFallsBackToEnglish is the safety net: a
// half-translated email is worse than a fully-English one.
func TestRender_UnknownLocaleFallsBackToEnglish(t *testing.T) {
	t.Parallel()
	w := renderWelcome("u@y.z", "fr").subject
	r := renderReset("123456", "fr").subject
	assert.Equal(t, "Welcome to Itinera", w)
	assert.Equal(t, "Reset your Itinera password", r)
}

// ---------- NoopSender ----------------------------------------------------

// TestNoopSender_BothReturnNil documents the dev contract.
func TestNoopSender_BothReturnNil(t *testing.T) {
	t.Parallel()
	u := models.User{ID: uuid.New(), Email: "u@y.z"}
	var s Sender = NoopSender{}
	assert.NoError(t, s.SendWelcome(context.Background(), u, "en"))
	assert.NoError(t, s.SendPasswordReset(context.Background(), u, "123456", "es"))
}

// ---------- LogSender -----------------------------------------------------

// safeBuffer is a goroutine-safe wrapper around bytes.Buffer so the
// LogSender's calls (which run in arbitrary goroutines if a handler
// chooses to fire-and-forget) don't race against the test's reads.
type safeBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (s *safeBuffer) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.Write(p)
}

func (s *safeBuffer) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.buf.String()
}

func TestLogSender_WelcomeAppearsInLog(t *testing.T) {
	t.Parallel()
	buf := &safeBuffer{}
	s := &LogSender{Log: log.New(buf, "", 0)}

	u := models.User{ID: uuid.New(), Email: "u@y.z"}
	require.NoError(t, s.SendWelcome(context.Background(), u, "es"))

	out := buf.String()
	assert.Contains(t, out, "[email:welcome]")
	assert.Contains(t, out, "to=u@y.z")
	assert.Contains(t, out, "locale=es")
	assert.Contains(t, out, "Bienvenido a Itinera")
}

func TestLogSender_ResetIncludesCode(t *testing.T) {
	t.Parallel()
	buf := &safeBuffer{}
	s := &LogSender{Log: log.New(buf, "", 0)}

	u := models.User{ID: uuid.New(), Email: "x@y.z"}
	require.NoError(t, s.SendPasswordReset(context.Background(), u, "867530", "en"))

	out := buf.String()
	assert.Contains(t, out, "[email:reset]")
	assert.Contains(t, out, "code=867530")
	assert.Contains(t, out, "Reset your Itinera password")
}

// ---------- ResendSender --------------------------------------------------

// TestResendSender_HappyPath spins up an httptest.Server that pretends to
// be api.resend.com. We assert the exact JSON payload and headers.
func TestResendSender_HappyPath(t *testing.T) {
	t.Parallel()

	type captured struct {
		Auth string
		Body resendPayload
	}
	var got captured
	var mu sync.Mutex

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var p resendPayload
		_ = json.Unmarshal(body, &p)
		mu.Lock()
		got = captured{Auth: r.Header.Get("Authorization"), Body: p}
		mu.Unlock()
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"id":"abc"}`))
	}))
	defer srv.Close()

	// Redirect the production URL to our test server. We do this by
	// constructing the sender and stubbing the request URL — but the
	// sender hard-codes https://api.resend.com/emails. So instead we
	// test the inner `send` via a small mirror: we copy the sender's
	// behaviour by building a sender whose HTTPClient hits our test
	// server, and call send indirectly by sending through a helper.
	//
	// To avoid changing resend.go just for tests, we test the
	// rendered Payload via SendWelcome using a Sender that wraps a
	// custom HTTP client. That's what ResendSender does in production.
	s := NewResendSender("test-key", "Itinera <hello@goitinera.app>")

	// Swap the URL by routing via TestTransport. TestTransport routes
	// requests based on URL prefix — easier than exposing an unexported
	// field just for tests.
	s.HTTPClient = &http.Client{Transport: &rewriteTransport{
		baseURL:  "https://api.resend.com/emails",
		testURL:  srv.URL + "/emails",
		delegate: http.DefaultTransport,
	}}

	u := models.User{ID: uuid.New(), Email: "alice@example.test"}
	require.NoError(t, s.SendWelcome(context.Background(), u, "en"))

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, "Bearer test-key", got.Auth)
	assert.Equal(t, "Itinera <hello@goitinera.app>", got.Body.From)
	assert.Equal(t, []string{"alice@example.test"}, got.Body.To)
	assert.Equal(t, "Welcome to Itinera", got.Body.Subject)
	assert.Contains(t, got.Body.HTML, "alice@example.test")
}

// rewriteTransport is an http.RoundTripper that rewrites the URL of any
// request starting with baseURL to start with testURL instead. It's used
// solely to redirect ResendSender's hard-coded URL to an httptest.Server.
type rewriteTransport struct {
	baseURL  string
	testURL  string
	delegate http.RoundTripper
}

func (t *rewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.String(), t.baseURL) {
		rewritten := t.testURL + req.URL.String()[len(t.baseURL):]
		clone := req.Clone(req.Context())
		newURL, err := url.Parse(rewritten)
		if err != nil {
			return nil, err
		}
		clone.URL = newURL
		return t.delegate.RoundTrip(clone)
	}
	return t.delegate.RoundTrip(req)
}

// TestResendSender_5xxReturnsError asserts the unhappy path doesn't
// silently swallow the failure.
func TestResendSender_5xxReturnsError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":"bad"}`, http.StatusInternalServerError)
	}))
	defer srv.Close()

	s := NewResendSender("k", "Itinera <hi@y.z>")
	s.HTTPClient = &http.Client{Transport: &rewriteTransport{
		baseURL:  "https://api.resend.com/emails",
		testURL:  srv.URL + "/emails",
		delegate: http.DefaultTransport,
	}}

	u := models.User{ID: uuid.New(), Email: "x@y.z"}
	err := s.SendWelcome(context.Background(), u, "en")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
	assert.Contains(t, err.Error(), "bad")
}

// TestResendSender_EmptyAPIKey is a safety net: the factory never lets
// this happen, but a future refactor could.
func TestResendSender_EmptyAPIKey(t *testing.T) {
	t.Parallel()
	s := &ResendSender{APIKey: "", From: "Itinera <hi@y.z>"}
	u := models.User{ID: uuid.New(), Email: "x@y.z"}
	assert.Error(t, s.SendWelcome(context.Background(), u, "en"))
}

// ---------- Factory -------------------------------------------------------

// TestNewSender_RespectsEnvironment walks every branch of NewSender so a
// future addition (e.g. a SES implementation) can't silently flip the
// production behaviour.
func TestNewSender_RespectsEnvironment(t *testing.T) {
	t.Parallel()

	prod := &config.Config{Environment: "production", ResendAPIKey: "k", EmailFrom: "x@y.z"}
	if _, ok := NewSender(prod).(*ResendSender); !ok {
		t.Errorf("production should yield ResendSender, got %T", NewSender(prod))
	}

	dev := &config.Config{Environment: "development"}
	if _, ok := NewSender(dev).(NoopSender); !ok {
		t.Errorf("development should yield NoopSender, got %T", NewSender(dev))
	}

	staging := &config.Config{Environment: "staging"}
	if _, ok := NewSender(staging).(*LogSender); !ok {
		t.Errorf("staging should yield LogSender, got %T", NewSender(staging))
	}

	prodNoKey := &config.Config{Environment: "production", ResendAPIKey: ""}
	if _, ok := NewSender(prodNoKey).(*LogSender); !ok {
		t.Errorf("production without ResendAPIKey should fall back to LogSender, got %T", NewSender(prodNoKey))
	}
}