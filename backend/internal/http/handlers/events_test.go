package handlers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"backend/internal/http/handlers"
)

// noopEventStore is a stand-in for handlers.EventStore used in unit tests
// that do not need to assert on the persisted event (e.g. validation tests
// that fail before the Insert call is reached).
type noopEventStore struct{}

func (noopEventStore) Insert(_ context.Context, _ string, _ string, _ *string, _ *string, _ []byte, _ time.Time) error {
	return nil
}

// allowAllRateLimitStore is a stand-in for handlers.RateLimitStore that
// always returns "allowed". The TrackEvent handler treats errors as "allow"
// and a nil store also short-circuits the check, so a permissive fake is
// the safest test stand-in for validation-only tests.
type allowAllRateLimitStore struct{}

func (allowAllRateLimitStore) CheckAndIncrement(_ context.Context, _ string, _ time.Duration, _ int) (bool, error) {
	return true, nil
}

func TestTrackEvent_Validation(t *testing.T) {
	// The validation tests must fail BEFORE the EventStore.Insert call is
	// reached, so a no-op store is sufficient and keeps the test independent
	// of any database connection.
	h := &handlers.Handlers{
		EventsRepo:    noopEventStore{},
		RateLimitRepo: allowAllRateLimitStore{},
	}
	handler := http.HandlerFunc(h.TrackEvent)

	t.Run("Should reject invalid event type", func(t *testing.T) {
		payload := []byte(`{
			"type": "invalid_event_type",
			"metadata": {"foo": "bar"}
		}`)

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("Should reject malformed JSON", func(t *testing.T) {
		payload := []byte(`{bad json}`)

		req := httptest.NewRequest(http.MethodPost, "/events", bytes.NewBuffer(payload))
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}
