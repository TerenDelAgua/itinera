package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/http/middleware"

	"github.com/stretchr/testify/assert"
)

func TestSessionMiddleware(t *testing.T) {
	t.Run("No cookie - should generate a new session_id cookie and context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()

		var ctxSessionID string

		// Mock next handler to capture the injected context
		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if val, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
				ctxSessionID = val
			}
		})

		middlewareInstance := middleware.SessionMiddleware(false)
		middlewareInstance(nextHandler).ServeHTTP(rec, req)

		// 1. Verify that context was updated
		assert.NotEmpty(t, ctxSessionID, "Session ID should be injected into context")

		// 2. Verify that Set-Cookie header is present
		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 1)
		
		cookie := cookies[0]
		assert.Equal(t, "session_id", cookie.Name)
		assert.Equal(t, ctxSessionID, cookie.Value, "Cookie value should match context value")
		assert.True(t, cookie.HttpOnly, "Cookie should be HttpOnly")
		assert.Equal(t, "/", cookie.Path)
	})

	t.Run("With cookie - should reuse existing session_id", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		existingSessionID := "existing-test-session-123"
		req.AddCookie(&http.Cookie{
			Name:  "session_id",
			Value: existingSessionID,
		})

		rec := httptest.NewRecorder()
		var ctxSessionID string

		nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if val, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
				ctxSessionID = val
			}
		})

		middlewareInstance := middleware.SessionMiddleware(false)
		middlewareInstance(nextHandler).ServeHTTP(rec, req)

		// 1. Context should contain the EXISTING session ID
		assert.Equal(t, existingSessionID, ctxSessionID)

		// 2. Set-Cookie should NOT be called again
		cookies := rec.Result().Cookies()
		assert.Len(t, cookies, 0, "No new cookie should be set if one already exists")
	})
}
