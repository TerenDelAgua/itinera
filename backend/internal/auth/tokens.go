package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

// NewSecureToken returns a fresh opaque token suitable for the access or
// refresh cookie. The token is 32 random bytes encoded as hex (64 chars).
//
// We use 32 bytes (256 bits). The hex encoding is
// intentional: copy/paste into the address bar is tractable on desktop and
// on mobile, and the entropy is sufficient that the hex-as-string reading
// difficulty doesn't translate into a brute-force advantage.
func NewSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("read crypto rand: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the hex-encoded SHA-256 of the supplied token. This is
// what gets stored in `sessions.access_token_hash` and
// `sessions.refresh_token_hash`.
//
// The function is intentionally hand-rolled rather than imported from an
// external library to keep the dependency surface tight; the algorithm is stable
func HashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// CompareTokenHash does a constant-time comparison between a raw token and a
// pre-computed SHA-256 hex. Callers use this when they want to assert that
// the hash they persisted matches the hash they'd compute from a known
// candidate (e.g. a code, not the cookie) without timing leakage.
//
// Most call sites don't need this — the unique constraint on the column
// already prevents brute-force lookup at the DB layer — but the email-code
// reset path benefits from it because the supplied code is short.
func CompareTokenHash(rawToken, storedHash string) bool {
	expected := HashToken(rawToken)
	return subtle.ConstantTimeCompare([]byte(expected), []byte(storedHash)) == 1
}
