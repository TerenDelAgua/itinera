package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type ContextKeyUserId struct{}

// AuthMiddleware extracts a JWT from either the HttpOnly `auth_token` cookie
// (preferred, XSS-safe) or the `Authorization: Bearer ...` header (legacy,
// for non-browser clients). If no token is present the request continues
// without an authenticated user ID and downstream handlers will fall back
// to the guest session resolved by SessionMiddleware.
func AuthMiddleware(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenStr string

			// 1. Preferred path: HttpOnly cookie set by Register/Login handlers.
			if cookie, err := r.Cookie("auth_token"); err == nil && cookie.Value != "" {
				tokenStr = cookie.Value
			}

			// 2. Legacy path: Authorization header (non-browser clients, e2e tests).
			if tokenStr == "" {
				authHeader := r.Header.Get("Authorization")
				if strings.HasPrefix(authHeader, "Bearer ") {
					tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
				}
			}

			if tokenStr == "" {
				// No token found → continue (handler will fallback to session-based auth)
				next.ServeHTTP(w, r)
				return
			}

			claims := jwt.MapClaims{}
			_, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			})

			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return // ← bug fix: missing return after http.Error
			}

			userIdStr, ok := claims["user_id"].(string)
			if !ok {
				http.Error(w, "Invalid token claims", http.StatusUnauthorized)
				return
			}

			userID, err := uuid.Parse(userIdStr)
			if err != nil {
				http.Error(w, "Invalid user ID in token", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserId{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
