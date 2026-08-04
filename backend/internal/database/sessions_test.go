package database_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSessions_CreateAndRevoke covers the happy-path of token issuance and
// single-row revocation. Cleanup is per-test via UUIDs.
func TestSessions_CreateAndRevoke(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	auth := database.NewAuthRepository(pool)
	sess := database.NewSessionRepository(pool)
	ctx := context.Background()

	u, err := auth.CreateUser(ctx, uniqueEmail(t, "sess"), "Pa55word!", "en")
	require.NoError(t, err)
	t.Cleanup(func() { cleanupSessionsForUser(pool, u.ID) })
	t.Cleanup(func() { cleanupUserForTest(pool, u.ID) })

	expires := time.Now().Add(24 * time.Hour)
	created, err := sess.CreateSession(ctx, u.ID, "access-hash-a", "refresh-hash-a", expires, nil, nil)
	require.NoError(t, err)
	require.NotNil(t, created)
	assert.Equal(t, u.ID, created.UserID)
	assert.NotEqual(t, uuid.Nil, created.RefreshFamily)
	assert.Equal(t, "access-hash-a", created.AccessTokenHash, "struct holds the hash for in-process callers")

	found, err := sess.FindSessionByAccessTokenHash(ctx, "access-hash-a")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	require.NoError(t, sess.RevokeSession(ctx, created.ID))

	_, err = sess.FindSessionByAccessTokenHash(ctx, "access-hash-a")
	require.True(t, errors.Is(err, pgx.ErrNoRows), "expected ErrNoRows after revoke, got %v", err)
}

// TestSessions_EmptyHashReturnsNoRows asserts the early-exit path.
func TestSessions_EmptyHashReturnsNoRows(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	sess := database.NewSessionRepository(pool)
	_, err := sess.FindSessionByAccessTokenHash(context.Background(), "")
	assert.True(t, errors.Is(err, pgx.ErrNoRows))
}

// TestSessions_RotateAndRevokeFamily walks through refresh-rotation and the
// family-wide revoke that the reuse-detection handler triggers.
func TestSessions_RotateAndRevokeFamily(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	auth := database.NewAuthRepository(pool)
	sess := database.NewSessionRepository(pool)
	ctx := context.Background()

	u, err := auth.CreateUser(ctx, uniqueEmail(t, "rot"), "Pa55word!", "en")
	require.NoError(t, err)
	t.Cleanup(func() { cleanupSessionsForUser(pool, u.ID) })
	t.Cleanup(func() { cleanupUserForTest(pool, u.ID) })

	expires := time.Now().Add(24 * time.Hour)
	created, err := sess.CreateSession(ctx, u.ID, "ah-1", "rh-1", expires, nil, nil)
	require.NoError(t, err)

	originalFamily := created.RefreshFamily
	require.NoError(t, sess.RotateSession(ctx, created.ID, "ah-2", "rh-2", expires.Add(time.Hour)))

	fetched, err := sess.FindSessionByAccessTokenHash(ctx, "ah-2")
	require.NoError(t, err)
	assert.Equal(t, originalFamily, fetched.RefreshFamily,
		"Spec 017 §4.3: rotation uses the SAME row, so the family is preserved")
	assert.Nil(t, fetched.RefreshRotatedAt, "RotateSession bumps refresh_rotated_at; Scan sets it")
	if fetched.RefreshRotatedAt != nil {
		assert.True(t, strings.HasPrefix(*fetched.RefreshRotatedAt, "20"))
	}

	// Revoke the entire family; subsequent lookups by either hash must fail.
	require.NoError(t, sess.RevokeFamily(ctx, originalFamily))

	_, err = sess.FindSessionByAccessTokenHash(ctx, "ah-2")
	require.True(t, errors.Is(err, pgx.ErrNoRows))
}

// TestSessions_CleanupExpired verifies the job hook removes old rows.
func TestSessions_CleanupExpired(t *testing.T) {
	pool := getTestPoolOrSkip(t)
	auth := database.NewAuthRepository(pool)
	sess := database.NewSessionRepository(pool)
	ctx := context.Background()

	u, err := auth.CreateUser(ctx, uniqueEmail(t, "exp"), "Pa55word!", "en")
	require.NoError(t, err)
	t.Cleanup(func() { cleanupSessionsForUser(pool, u.ID) })
	t.Cleanup(func() { cleanupUserForTest(pool, u.ID) })

	// Insert a row already revoked and expired enough to be cleaned up.
	_, err = pool.Exec(ctx, `
		INSERT INTO sessions (user_id, access_token_hash, refresh_token_hash, refresh_family, revoked_at, expires_at, last_used_at)
		VALUES ($1, 'old-access', 'old-refresh', gen_random_uuid(), now() - interval '8 days', now() - interval '10 days', now() - interval '11 days')
	`, u.ID)
	require.NoError(t, err)

	deleted, err := sess.CleanupExpiredSessions(ctx)
	require.NoError(t, err)
	require.True(t, deleted >= 1)
}

// cleanupSessionsForUser deletes every session for a user (test-only).
func cleanupSessionsForUser(pool *pgxpool.Pool, userID uuid.UUID) {
	_, _ = pool.Exec(context.Background(), "DELETE FROM sessions WHERE user_id = $1", userID)
}

// cleanupUserForTest deletes the user row (sessions cascade automatically).
func cleanupUserForTest(pool *pgxpool.Pool, userID uuid.UUID) {
	_, _ = pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", userID)
}
