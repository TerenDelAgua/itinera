package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
)

// IsInternalSession returns true if the request comes from the dev's
// own browser (identified by the ITINERA_INTERNAL_TOKEN). Fails closed:
// if the token is not configured, every request is treated as non-internal.
//
// The token can be supplied via:
//   - Header X-Internal-Session: <token>   (preferred: lets the dev use
//                                          browser extensions like ModHeader)
//   - Cookie itinera_internal=<token>      (fallback: easy to inject from
//                                          the browser DevTools console)
//
// Comparison is constant-time to prevent timing side-channels.

func IsInternalSession(r *http.Request) bool {
	expected := os.Getenv("ITINERA_INTERNAL_TOKEN")
	if expected == "" {
		return false
	}
	expectedBytes := []byte(expected)

	if header := r.Header.Get("X-Internal-Session"); header != "" {
		return subtle.ConstantTimeCompare([]byte(header), expectedBytes) == 1
	}
	if cookie, err := r.Cookie("itinera_internal"); err == nil && cookie.Value != "" {
		return subtle.ConstantTimeCompare([]byte(cookie.Value), expectedBytes) == 1
	}
	return false
}
