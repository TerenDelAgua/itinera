package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/database"
	"backend/internal/models"
)

// TestShare_FullLifecycle_EnabledToDisabled exercises the core share flow:
//  1. Create a trip without a share.
//  2. EnableShare → returns a token + non-nil expiry (for ghost sessions).
//  3. Re-fetch the trip via GetTrip → must include share_token/share_enabled.
//  4. DisableShare → token + enabled flag reset to NULL/false.
//  5. Re-fetch again → fields are now empty.
func TestShare_FullLifecycle_EnabledToDisabled(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	ctx := context.Background()

	// 1. Setup: create a guest-owned trip.
	sessionID := "test-share-session-" + uuid.New().String()
	trip, err := repo.CreateTrip(ctx, nil, &sessionID, false, models.Trip{
		Name:                  "Share Test Trip",
		BaseCurrency:          "EUR",
		DefaultExpenseCurrency: "EUR",
		IsPublicDemo:          false,
		StartDate:             "2026-08-01",
		EndDate:               "2026-08-10",
	})
	require.NoError(t, err)
	require.NotNil(t, trip)

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", trip.ID)
	}()

	// Sanity: initial GetTrip returns share_enabled=false and no token.
	initial, err := repo.GetTrip(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)
	require.NotNil(t, initial)
	assert.False(t, initial.ShareEnabled, "new trip should not be shared by default")
	assert.Nil(t, initial.ShareToken, "new trip should have no share token yet")

	// 2. EnableShare.
	token, expiresAt, err := repo.EnableShare(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)
	require.NotEmpty(t, token, "token should be 12 chars base64 URL")
	assert.Len(t, token, 12, "share token must be exactly 12 chars (9 bytes -> base64url)")
	if expiresAt != nil {
		// Ghost users get a 7-day expiry. Allow some slack.
		diff := time.Until(*expiresAt)
		assert.True(t, diff > 6*24*time.Hour && diff <= 7*24*time.Hour+time.Minute,
			"ghost share should expire in ~7 days, got %v", diff)
	}

	// 3. Re-fetch and verify the token came through.
	enabled, err := repo.GetTrip(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)
	assert.True(t, enabled.ShareEnabled, "GetTrip must report share_enabled=true after EnableShare")
	require.NotNil(t, enabled.ShareToken, "GetTrip must surface share_token (this was the bug fix)")
	assert.Equal(t, token, *enabled.ShareToken, "fetched token must match the one returned by EnableShare")

	// 4. DisableShare.
	err = repo.DisableShare(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)

	// 5. Re-fetch: token gone, share_enabled false.
	disabled, err := repo.GetTrip(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)
	assert.False(t, disabled.ShareEnabled)
	assert.Nil(t, disabled.ShareToken, "share_token should be NULL after DisableShare")
}

// TestShare_EnableShareIsIdempotent verifies that calling EnableShare twice
// returns the SAME token (no new token generated, no expiry extension). The
// second call should just return the existing one.
func TestShare_EnableShareIsIdempotent(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	ctx := context.Background()

	sessionID := "test-share-idempotent-" + uuid.New().String()
	trip, err := repo.CreateTrip(ctx, nil, &sessionID, false, models.Trip{
		Name:                  "Idempotent Share Test",
		BaseCurrency:          "EUR",
		DefaultExpenseCurrency: "EUR",
		StartDate:             "2026-08-01",
		EndDate:               "2026-08-10",
	})
	require.NoError(t, err)

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", trip.ID)
	}()

	token1, exp1, err := repo.EnableShare(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)

	// Tiny pause to ensure any "new token" generation would produce a different time.
	time.Sleep(50 * time.Millisecond)

	token2, exp2, err := repo.EnableShare(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)

	assert.Equal(t, token1, token2, "EnableShare must be idempotent — same token on second call")
	if exp1 != nil && exp2 != nil {
		assert.Equal(t, exp1.Unix(), exp2.Unix(), "EnableShare must not extend the expiry")
	}
}

// TestShare_GetByShareToken_PublicPath verifies that the public-facing
// GetByShareToken returns sanitized trip data and excludes the share_token
// from being accidentally exposed in non-public queries (the handler is the
// gatekeeper; this test just checks the repo returns what the handler will
// surface, so the public endpoint can scrub sensitive fields).
func TestShare_GetByShareToken_ValidToken(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	ctx := context.Background()

	sessionID := "test-share-token-" + uuid.New().String()
	trip, err := repo.CreateTrip(ctx, nil, &sessionID, false, models.Trip{
		Name:                  "Token Lookup Test",
		BaseCurrency:          "EUR",
		DefaultExpenseCurrency: "EUR",
		StartDate:             "2026-08-01",
		EndDate:               "2026-08-10",
	})
	require.NoError(t, err)

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", trip.ID)
	}()

	token, _, err := repo.EnableShare(ctx, trip.ID.String(), nil, &sessionID)
	require.NoError(t, err)

	// Resolve via token.
	fetched, err := repo.GetByShareToken(ctx, token)
	require.NoError(t, err)
	assert.Equal(t, trip.ID, fetched.ID, "GetByShareToken should resolve the trip id")
	assert.True(t, fetched.ShareEnabled, "fetched trip should still be share-enabled")
	require.NotNil(t, fetched.ShareToken)
	assert.Equal(t, token, *fetched.ShareToken)
}

// TestShare_GetByShareToken_Invalid verifies that querying with an
// unknown token returns pgx.ErrNoRows (the contract GetSharedTrip uses
// to return 404).
func TestShare_GetByShareToken_Invalid(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	ctx := context.Background()

	_, err := repo.GetByShareToken(ctx, "AAAAAAAAAAAA")
	require.Error(t, err, "invalid token must return an error")
}

// TestShare_DisableShare_Unauthorized verifies that DisableShare on a trip
// you don't own returns an error. This is the 403 path.
func TestShare_DisableShare_Unauthorized(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	ctx := context.Background()

	// Owner creates and enables.
	ownerSession := "test-share-owner-" + uuid.New().String()
	trip, err := repo.CreateTrip(ctx, nil, &ownerSession, false, models.Trip{
		Name:                  "Auth Test Trip",
		BaseCurrency:          "EUR",
		DefaultExpenseCurrency: "EUR",
		StartDate:             "2026-08-01",
		EndDate:               "2026-08-10",
	})
	require.NoError(t, err)

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", trip.ID)
	}()

	_, _, err = repo.EnableShare(ctx, trip.ID.String(), nil, &ownerSession)
	require.NoError(t, err)

	// Attacker tries to disable from a different session.
	attackerSession := "test-share-attacker-" + uuid.New().String()
	err = repo.DisableShare(ctx, trip.ID.String(), nil, &attackerSession)
	require.Error(t, err, "DisableShare from non-owner session must fail")

	// Owner should still see the share enabled.
	check, err := repo.GetTrip(ctx, trip.ID.String(), nil, &ownerSession)
	require.NoError(t, err)
	assert.True(t, check.ShareEnabled, "share should still be enabled after unauthorized disable attempt")
}

// TestShare_CountActiveSharesByUser verifies the soft-limit query used
// by handlers.EnableShare to log warnings when a user/session has more
// than 10 active shares.
func TestShare_CountActiveSharesByUser(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	ctx := context.Background()

	sessionID := "test-share-count-" + uuid.New().String()

	// Clean up any prior state for this session.
	_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE session_id = $1", sessionID)

	// Create 2 trips and enable both.
	for i := 0; i < 2; i++ {
		trip, err := repo.CreateTrip(ctx, nil, &sessionID, false, models.Trip{
			Name:                  "Count Test " + string(rune('A'+i)),
			BaseCurrency:          "EUR",
			DefaultExpenseCurrency: "EUR",
			StartDate:             "2026-08-01",
			EndDate:               "2026-08-10",
		})
		require.NoError(t, err)
		_, _, err = repo.EnableShare(ctx, trip.ID.String(), nil, &sessionID)
		require.NoError(t, err)
	}

	defer func() {
		_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE session_id = $1", sessionID)
	}()

	count, err := repo.CountActiveSharesByUser(ctx, nil, &sessionID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 2, "should have at least 2 active shares for this session")
}