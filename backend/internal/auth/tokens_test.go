package auth

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewSecureToken_EntropyAndFormat verifies the format invariants so a
// future refactor can't silently degrade the security model.
func TestNewSecureToken_EntropyAndFormat(t *testing.T) {
	t.Parallel()
	tok, err := NewSecureToken()
	require.NoError(t, err)
	assert.Len(t, tok, 64, "32 raw bytes hex-encoded produces 64 chars")
	for _, r := range tok {
		assert.True(t,
			(r >= '0' && r <= '9') || (r >= 'a' && r <= 'f'),
			"token must be lowercase hex, found %q (%v)", tok, r)
	}
}

// TestNewSecureToken_UniqueAcrossIterations is a smoke test against
// catastrophic RNG failure.
func TestNewSecureToken_UniqueAcrossIterations(t *testing.T) {
	t.Parallel()
	const n = 64
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		tok, err := NewSecureToken()
		require.NoError(t, err)
		_, dup := seen[tok]
		assert.False(t, dup, "duplicate token at i=%d (catastrophic RNG failure): %s", i, tok)
		seen[tok] = struct{}{}
	}
}

// TestHashToken_Determinism verifies SHA-256 of identical inputs yields
// identical outputs (the whole reason the unique index on the column works).
func TestHashToken_Determinism(t *testing.T) {
	t.Parallel()
	a := HashToken("hello")
	b := HashToken("hello")
	assert.Equal(t, a, b)
	c := HashToken("Hello")
	assert.NotEqual(t, a, c, "case matters in the hash input")
}

// TestHashToken_KnownVector uses a fixed input and a published SHA-256
// result so accidental algorithm changes break this test loudly. The
// reference value below was computed with `printf hello | sha256sum`.
func TestHashToken_KnownVector(t *testing.T) {
	t.Parallel()
	// sha256("hello") = 2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
	assert.Equal(t,
		"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
		HashToken("hello"),
	)
}

// TestCompareTokenHash_ConstantTimeSpotCheck covers both halves of the
// expected behaviour without timing the calls (Go's stdlib already does the
// right thing; we're just pinning the contract for callers).
func TestCompareTokenHash_ConstantTimeSpotCheck(t *testing.T) {
	t.Parallel()
	raw := "123456"
	stored := HashToken(raw)

	assert.True(t, CompareTokenHash(raw, stored), "matches on equal inputs")
	assert.False(t, CompareTokenHash("654321", stored), "differs on different inputs")
	assert.False(t, CompareTokenHash(raw, "not-a-hash"), "differs on wrong format")
}
