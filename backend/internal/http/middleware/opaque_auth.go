package middleware

import (
	"backend/internal/auth"
	"backend/internal/database"
	"context"
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
)

// opaque_auth.go implements the post-cutover authentication path (Spec 017
// §5). Tokens are opaque 64-char hex strings (see internal/auth.NewSecureToken),
// served via HttpOnly cookies, hashed with SHA-256 and looked up in the
// `sessions` table.

const (
	// CookieAccessToken is the HttpOnly cookie name for the access token,
	// post-cutover (Spec 017 §4.2). The legacy JWT cookie kept the name
	// `auth_token`; during the dual-stack cutover both names co-exist.
	CookieAccessToken = "itinera_access"

	// CookieRefreshToken is the HttpOnly cookie name for the refresh token.
	CookieRefreshToken = "itinera_refresh"
)

// AuthMiddlewareV2 resolves the access token from the cookie of the same
// name, hashes it with SHA-256 and looks the row up in `sessions`. When the
// row is found, the context is decorated with both ContextKeyUserId and
// ContextKeySessionId (the latter stored as the row's refresh family,
// so down-stream logic that wants "is this the user's session" can read it).
//
// Behaviour matrix:
//   - No cookie                → continue (guest path falls back to
//     SessionMiddleware).
//   - Cookie but row missing   → continue. The request proceeds as if there
//     were no cookie at all. We do NOT 401 here
//     because the same request may already be
//     in the JWT auth branch (dual-stack) and a
//     401 would surprise the caller.
//   - Cookie + row found       → decorate context, continue.
//
// The middleware is intentionally permissive: AuthZ (i.e. "this user is
// authorised to access this resource") is the per-route handler's job, not
// the middleware's. The middleware's contract is "I can tell you WHO is
// calling"; the handler decides if they're allowed.
func AuthMiddlewareV2(store database.SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(CookieAccessToken)
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			hash := auth.HashToken(cookie.Value)
			session, err := store.FindSessionByAccessTokenHash(r.Context(), hash)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					// Unknown or revoked cookie. Treat as anonymous. The
					// cookie itself isn't deleted here — that's the logout
					// handler's job — to keep this middleware idempotent
					// for retries behind a flaky proxy.
					next.ServeHTTP(w, r)
					return
				}
				// Real DB error: 500. We don't want to leak the underlying
				// error message to the client (Spec 017 §9.3: never expose
				// err.Error() in prod).
				http.Error(w, "Authentication lookup failed", http.StatusInternalServerError)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUserId{}, session.UserID)
			ctx = context.WithValue(ctx, ContextKeySessionId{}, session.RefreshFamily.String())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// SetAccessCookie and SetRefreshCookie are the writers the future login /
// refresh handlers will use. Centralised so the flags (Path, MaxAge,
// SameSite, Secure) stay in lockstep with what AuthMiddlewareV2 expects.
//
// `secure` MUST be true in production; the helper trusts the caller's env.
func SetAccessCookie(w http.ResponseWriter, rawToken string, maxAgeSec int, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieAccessToken,
		Value:    rawToken,
		Path:     "/",
		MaxAge:   maxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteFor(secure),
	})
}

func SetRefreshCookie(w http.ResponseWriter, rawToken string, maxAgeSec int, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieRefreshToken,
		Value:    rawToken,
		Path:     "/",
		MaxAge:   maxAgeSec,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteFor(secure),
	})
}

func ClearAuthCookies(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieAccessToken,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteFor(secure),
	})
	http.SetCookie(w, &http.Cookie{
		Name:     CookieRefreshToken,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: sameSiteFor(secure),
	})
}

// sameSiteFor keeps the SameSite attribute centralised. We always return
// Lax because Itinera is a first-party app and SameSite=None is only
// needed for embed-iframe use cases we don't have yet. Having a function
// (instead of a constant) leaves the door open for env-driven override later.
func sameSiteFor(_ bool) http.SameSite {
	return http.SameSiteLaxMode
}
