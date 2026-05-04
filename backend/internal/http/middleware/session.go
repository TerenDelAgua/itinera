package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type ContextKeySessionId struct{}

// SessionMiddleware ensures every request has a session identifier.
// If the client doesn't have a session cookie, we issue one.
func SessionMiddleware(isProduction bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			cookieName := "session_id"
			headerName := "X-Session-Id"
			var sessionId string

			// 1. Try to get from cookie
			cookie, err := r.Cookie(cookieName)
			if err == nil && cookie.Value != "" {
				sessionId = cookie.Value
			}

			// 2. Fallback to header
			if sessionId == "" {
				sessionId = r.Header.Get(headerName)
			}

			// 3. Generate if still empty
			isNew := false
			if sessionId == "" {
				sessionId = uuid.New().String()
				isNew = true
			}

			// Always set the header in response for the frontend to track
			w.Header().Set(headerName, sessionId)

			if isNew {
				isSecureContext := isProduction || r.Header.Get("X-Forwarded-Proto") == "https"
				sameSite := http.SameSiteLaxMode
				if isSecureContext {
					sameSite = http.SameSiteNoneMode
				}

				http.SetCookie(w, &http.Cookie{
					Name:     cookieName,
					Value:    sessionId,
					Path:     "/",
					MaxAge:   60 * 60 * 24 * 365, // 1 year
					HttpOnly: true,
					Secure:   isSecureContext,
					SameSite: sameSite,
				})
			}

			ctx := context.WithValue(r.Context(), ContextKeySessionId{}, sessionId)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
