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
			cookieName := "session_id"
			cookie, err := r.Cookie(cookieName)
			if err != nil || cookie.Value == "" {
				sessionId := uuid.New().String()

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

				ctx := context.WithValue(r.Context(), ContextKeySessionId{}, sessionId)
				next.ServeHTTP(w, r.WithContext(ctx))
			} else {
				ctx := context.WithValue(r.Context(), ContextKeySessionId{}, cookie.Value)
				next.ServeHTTP(w, r.WithContext(ctx))
			}
		})
	}
}
