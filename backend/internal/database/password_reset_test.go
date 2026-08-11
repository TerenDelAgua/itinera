package database_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"backend/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestPasswordReset_CreateAndFind covers the basic issuance + lookup.
func TestPasswordReset_CreateAndFind(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	auth := database.NewAuthRepository(pool)
	reset := database.NewPasswordResetRepository(pool)
	ctx := context.Background()

	u, err := auth.CreateUser(ctx, uniqueEmail(t, "prs"), "Pa55word!", "en")
	require.NoError(t, err)
	t.Cleanup(func() { cleanupPasswordResetsForUser(pool, u.ID) })
	t.Cleanup(func() { cleanupUserForTest(pool, u.ID) })

	hash := "sha256-of-6-digits-000123"
	expiry := time.Now().Add(1 * time.Hour)
	require.NoError(t, reset.Create(ctx, u.ID, hash, expiry, nil))

	found, err := reset.FindActiveByHash(ctx, hash)
	require.NoError(t, err)
	assert.Equal(t, u.ID, found.UserID)
	assert.Equal(t, 0, found.Attempts, "freshly created tokens must start at 0 attempts")
	assert.Nil(t, found.UsedAt)
	assert.Nil(t, found.LockedAt)
}

// TestPasswordReset_EmptyHash_NoRows asserts the early-exit.
func TestPasswordReset_EmptyHash_NoRows(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	reset := database.NewPasswordResetRepository(pool)
	_, err := reset.FindActiveByHash(context.Background(), "")
	assert.True(t, errors.Is(err, pgx.ErrNoRows))
}

// TestPasswordReset_RecordFailedAttemptLocksAtMax walks the 5-attempt lockout.
func TestPasswordReset_RecordFailedAttemptLocksAtMax(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	auth := database.NewAuthRepository(pool)
	reset := database.NewPasswordResetRepository(pool)
	ctx := context.Background()

	u, err := auth.CreateUser(ctx, uniqueEmail(t, "lock"), "Pa55word!", "en")
	require.NoError(t, err)
	t.Cleanup(func() { cleanupPasswordResetsForUser(pool, u.ID) })
	t.Cleanup(func() { cleanupUserForTest(pool, u.ID) })

	require.NoError(t, reset.Create(ctx, u.ID, "hash-1", time.Now().Add(time.Hour), nil))
	created, err := reset.FindActiveByHash(ctx, "hash-1")
	require.NoError(t, err)

	var locked bool
	for i := 0; i < database.MaxAttempts; i++ {
		attempts, l, err := reset.RecordFailedAttempt(ctx, created.ID)
		require.NoError(t, err)
		assert.Equal(t, i+1, attempts)
		if i < database.MaxAttempts-1 {
			assert.False(t, l, "should not lock before attempts == MaxAttempts")
		} else {
			locked = l
			assert.True(t, l, "lock must be set on the MaxAttempts-th failure")
		}
	}
	require.True(t, locked)

	// After lockout, FindActiveByHash must return ErrNoRows (matches
	// the contract for handlers: hide state from the attacker).
	_, err = reset.FindActiveByHash(ctx, "hash-1")
	assert.True(t, errors.Is(err, pgx.ErrNoRows), "locked rows must be invisible to FindActiveByHash")
}

// TestPasswordReset_MarkUsed_Idempotent asserts the row is closed once.
func TestPasswordReset_MarkUsed_Idempotent(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	auth := database.NewAuthRepository(pool)
	reset := database.NewPasswordResetRepository(pool)
	ctx := context.Background()

	u, err := auth.CreateUser(ctx, uniqueEmail(t, "use"), "Pa55word!", "en")
	require.NoError(t, err)
	t.Cleanup(func() { cleanupPasswordResetsForUser(pool, u.ID) })
	t.Cleanup(func() { cleanupUserForTest(pool, u.ID) })

	require.NoError(t, reset.Create(ctx, u.ID, "hash-mu", time.Now().Add(time.Hour), nil))
	tok, err := reset.FindActiveByHash(ctx, "hash-mu")
	require.NoError(t, err)

	require.NoError(t, reset.MarkUsed(ctx, tok.ID))

	// Second call must NOT succeed (was idempotent before, now distinguishes).
	err = reset.MarkUsed(ctx, tok.ID)
	require.Error(t, err, "second MarkUsed must error to make replay detectable")
}

// TestPasswordReset_MarkPreviousAsUsed enforces the "one active code per user"
// invariant required by the spec.
func TestPasswordReset_MarkPreviousAsUsed(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	auth := database.NewAuthRepository(pool)
	reset := database.NewPasswordResetRepository(pool)
	ctx := context.Background()

	u, err := auth.CreateUser(ctx, uniqueEmail(t, "prev"), "Pa55word!", "en")
	require.NoError(t, err)
	t.Cleanup(func() { cleanupPasswordResetsForUser(pool, u.ID) })
	t.Cleanup(func() { cleanupUserForTest(pool, u.ID) })

	require.NoError(t, reset.Create(ctx, u.ID, "first-hash", time.Now().Add(time.Hour), nil))
	require.NoError(t, reset.MarkPreviousAsUsed(ctx, u.ID))

	// First is no longer active.
	_, err = reset.FindActiveByHash(ctx, "first-hash")
	assert.True(t, errors.Is(err, pgx.ErrNoRows))

	// A new one issued AFTER the mark works fine.
	require.NoError(t, reset.Create(ctx, u.ID, "second-hash", time.Now().Add(time.Hour), nil))
	fresh, err := reset.FindActiveByHash(ctx, "second-hash")
	require.NoError(t, err)
	assert.Nil(t, fresh.UsedAt)
}

// cleanupPasswordResetsForUser deletes every reset token for the user (test-only).
func cleanupPasswordResetsForUser(pool *pgxpool.Pool, userID uuid.UUID) {
	_, _ = pool.Exec(context.Background(), "DELETE FROM password_reset_tokens WHERE user_id = $1", userID)
}
