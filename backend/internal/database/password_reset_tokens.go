package database

import (
	"backend/internal/models"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The repo owns the integrity of the reset flow but NOT the rate limiter —
// callers (handlers) gate the per-email / per-IP attempts via the existing
// rate_limits table before invoking Create().
//
// Hardening:
//   - All write paths reject rows whose previous attempt counter passed the
//     5-try lockout threshold (Spec 017 §5.7) by consulting attempts/locked_at.
//   - hash comparisons happen in constant time (handled in the handler).
type PasswordResetRepository struct {
	Pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{Pool: pool}
}

// MaxAttempts is the per-code brute-force cap defined by Spec 017 §5.7.
// Once exceeded, the row's locked_at is set and subsequent RecordFailedAttempt
// calls return errors matching pgx.ErrNoRows so the handler can return a
// generic ERROR_INVALID_TOKEN without confirming whether the token exists.
const MaxAttempts = 5

// Create inserts a fresh reset-token row. The caller is expected to mark any
// pre-existing active tokens for the same user as used BEFORE calling this
// (MarkPreviousAsUsed), so that each /forgot-password invocation replaces the
// in-flight code rather than leaving the user with multiple live ones.
func (r *PasswordResetRepository) Create(ctx context.Context, userID uuid.UUID, hash string, expiresAt time.Time, ip *string) error {
	var ipText pgtype.Text
	if ip != nil {
		ipText.String = *ip
		ipText.Valid = true
	}

	_, err := r.Pool.Exec(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at, ip_address)
		VALUES ($1, $2, $3, $4)
	`, userID, hash, expiresAt, ipText)
	if err != nil {
		return fmt.Errorf("insert password_reset_token: %w", err)
	}
	return nil
}

// MarkPreviousAsUsed sets used_at = now() on every active row for the user
// that doesn't have a matching hash. Called right before Create() to enforce
// the "one active code per user" invariant from Spec 017 §4.2.
func (r *PasswordResetRepository) MarkPreviousAsUsed(ctx context.Context, userID uuid.UUID) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE user_id = $1 AND used_at IS NULL
	`, userID)
	return err
}

// FindActiveByHash returns the row matching the supplied hash IF and only
// IF: used_at IS NULL, locked_at IS NULL, expires_at > now(). The handler
// uses this to enforce the contract "active code matches what the user typed"
// while keeping the comparison constant-time at the Go layer.
func (r *PasswordResetRepository) FindActiveByHash(ctx context.Context, hash string) (*models.PasswordResetToken, error) {
	if hash == "" {
		return nil, pgx.ErrNoRows
	}

	var t models.PasswordResetToken
	var expiresAt, usedAt, lockedAt, createdAt pgtype.Timestamptz
	var ip pgtype.Text

	err := r.Pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, attempts,
		       locked_at, created_at, ip_address
		FROM password_reset_tokens
		WHERE token_hash = $1
		  AND used_at IS NULL
		  AND locked_at IS NULL
		  AND expires_at > now()
		LIMIT 1
	`, hash).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &expiresAt, &usedAt, &t.Attempts,
		&lockedAt, &createdAt, &ip,
	)
	if err != nil {
		return nil, err
	}

	t.ExpiresAt = formatTimestamptz(expiresAt)
	t.CreatedAt = formatTimestamptz(createdAt)
	if usedAt.Valid {
		v := formatTimestamptz(usedAt)
		t.UsedAt = &v
	}
	if lockedAt.Valid {
		v := formatTimestamptz(lockedAt)
		t.LockedAt = &v
	}
	if ip.Valid {
		v := ip.String
		t.IPAddress = &v
	}
	return &t, nil
}

// FindActiveByUser is a debugging / admin helper used by the security log
// viewer; not exposed via the API yet.
func (r *PasswordResetRepository) FindActiveByUser(ctx context.Context, userID uuid.UUID) (*models.PasswordResetToken, error) {
	var t models.PasswordResetToken
	var expiresAt, usedAt, lockedAt, createdAt pgtype.Timestamptz
	var ip pgtype.Text

	err := r.Pool.QueryRow(ctx, `
		SELECT id, user_id, token_hash, expires_at, used_at, attempts,
		       locked_at, created_at, ip_address
		FROM password_reset_tokens
		WHERE user_id = $1 AND used_at IS NULL AND locked_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1
	`, userID).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &expiresAt, &usedAt, &t.Attempts,
		&lockedAt, &createdAt, &ip,
	)
	if err != nil {
		return nil, err
	}

	t.ExpiresAt = formatTimestamptz(expiresAt)
	t.CreatedAt = formatTimestamptz(createdAt)
	if usedAt.Valid {
		v := formatTimestamptz(usedAt)
		t.UsedAt = &v
	}
	if lockedAt.Valid {
		v := formatTimestamptz(lockedAt)
		t.LockedAt = &v
	}
	if ip.Valid {
		v := ip.String
		t.IPAddress = &v
	}
	return &t, nil
}

// RecordFailedAttempt increments the row's attempts counter and locks it
// once the count crosses MaxAttempts. The returned (attempts, locked, err)
// tuple lets the handler either:
//   - return INVALID_TOKEN when locked == true (hides state from attacker), or
//   - return INVALID_TOKEN with a "X attempts left" hint when attempts < MaxAttempts.
//
// The flag is intentionally returned rather than a boolean wrapper so the
// caller can audit the count alongside the lock state without an extra query.
func (r *PasswordResetRepository) RecordFailedAttempt(ctx context.Context, tokenID uuid.UUID) (attempts int, locked bool, err error) {
	err = r.Pool.QueryRow(ctx, `
		UPDATE password_reset_tokens
		SET attempts = attempts + 1,
		    locked_at = CASE WHEN attempts + 1 >= $2 THEN now() ELSE locked_at END
		WHERE id = $1
		RETURNING attempts, locked_at IS NOT NULL
	`, tokenID, MaxAttempts).Scan(&attempts, &locked)
	if err != nil {
		return 0, false, err
	}
	return attempts, locked, nil
}

// MarkUsed closes the row when the user supplies the correct code. Idempotent
// (subsequent calls return ErrNoRows).
func (r *PasswordResetRepository) MarkUsed(ctx context.Context, tokenID uuid.UUID) error {
	res, err := r.Pool.Exec(ctx, `
		UPDATE password_reset_tokens
		SET used_at = now()
		WHERE id = $1 AND used_at IS NULL
	`, tokenID)
	if err != nil {
		return fmt.Errorf("mark reset token used: %w", err)
	}
	if res.RowsAffected() == 0 {
		return errors.New("password reset token already used")
	}
	return nil
}

// HardDeleteOldTokens is the cleanup hook for the password reset tokens
// table. Used after the
// token expires (or after it's used) to keep the index small.
func (r *PasswordResetRepository) HardDeleteOldTokens(ctx context.Context) (int64, error) {
	res, err := r.Pool.Exec(ctx, `
		DELETE FROM password_reset_tokens
		WHERE used_at IS NOT NULL
		   OR expires_at < now() - interval '7 days'
	`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}
