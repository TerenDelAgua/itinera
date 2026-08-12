package database

import (
	"backend/internal/models"
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepository struct {
	Pool *pgxpool.Pool
}

func NewAuthRepository(pool *pgxpool.Pool) *AuthRepository {
	return &AuthRepository{Pool: pool}
}
func normalizeEmail(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}
func (r *AuthRepository) CreateUser(ctx context.Context, email, password, locale string) (*models.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	normalized := normalizeEmail(email)
	if normalized == "" {
		return nil, errors.New("email is required")
	}
	if !isSupportedLocale(locale) {
		locale = "en"
	}

	if _, err := r.Pool.Exec(ctx,
		`DELETE FROM users WHERE LOWER(email) = $1 AND deleted_at IS NOT NULL`,
		normalized,
	); err != nil {
		return nil, fmt.Errorf("purge soft-deleted user: %w", err)
	}

	var user models.User
	var createdAt, updatedAt, termsAcceptedAt, deletedAt pgtype.Timestamptz

	query := `
		INSERT INTO users (email, password_hash, locale)
		VALUES ($1, $2, $3)
		RETURNING
			id, email, password_hash, created_at, tier, locale,
			terms_accepted_at, updated_at, deleted_at
	`

	err = r.Pool.QueryRow(ctx, query, normalized, string(hash), locale).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&createdAt,
		&user.Tier,
		&user.Locale,
		&termsAcceptedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert user: %w", err)
	}

	user.CreatedAt = tsToISO(createdAt)
	user.UpdatedAt = tsToISO(updatedAt)
	if termsAcceptedAt.Valid {
		s := tsToISO(termsAcceptedAt)
		user.TermsAcceptedAt = &s
	}
	if deletedAt.Valid {
		s := tsToISO(deletedAt)
		user.DeletedAt = &s
	}
	return &user, nil
}

// GetUserByEmail returns the user record matching the (case-insensitive, trim-
// normalised) email. Returns pgx.ErrNoRows when no row exists so the caller
// can branch on errors.Is(err, pgx.ErrNoRows) without string-matching.
//
// Soft-deleted users (deleted_at IS NOT NULL) ARE returned: the login flow
// distinguishes them by inspecting user.DeletedAt and returns 403 ACCOUNT_DELETED.
func (r *AuthRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	normalized := normalizeEmail(email)
	if normalized == "" {
		return nil, errors.New("email is required")
	}

	var user models.User
	var createdAt, updatedAt, termsAcceptedAt, deletedAt pgtype.Timestamptz

	query := `
		SELECT
			id, email, password_hash, created_at, tier, locale,
			terms_accepted_at, updated_at, deleted_at
		FROM users
		WHERE LOWER(email) = $1
		ORDER BY (deleted_at IS NULL) DESC, created_at DESC
		LIMIT 1
	`

	err := r.Pool.QueryRow(ctx, query, normalized).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&createdAt,
		&user.Tier,
		&user.Locale,
		&termsAcceptedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = tsToISO(createdAt)
	user.UpdatedAt = tsToISO(updatedAt)
	if termsAcceptedAt.Valid {
		s := tsToISO(termsAcceptedAt)
		user.TermsAcceptedAt = &s
	}
	if deletedAt.Valid {
		s := tsToISO(deletedAt)
		user.DeletedAt = &s
	}
	return &user, nil
}

// GetUserByID is the canonical lookup used by /auth/me and the access-token
// middleware. Returns pgx.ErrNoRows when the row doesn't exist so the caller
// can treat it as "no authenticated session" without leaking that fact to
// the client.
func (r *AuthRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var user models.User
	var createdAt, updatedAt, termsAcceptedAt, deletedAt pgtype.Timestamptz

	query := `
		SELECT
			id, email, password_hash, created_at, tier, locale,
			terms_accepted_at, updated_at, deleted_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`

	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&createdAt,
		&user.Tier,
		&user.Locale,
		&termsAcceptedAt,
		&updatedAt,
		&deletedAt,
	)
	if err != nil {
		return nil, err
	}

	user.CreatedAt = tsToISO(createdAt)
	user.UpdatedAt = tsToISO(updatedAt)
	if termsAcceptedAt.Valid {
		s := tsToISO(termsAcceptedAt)
		user.TermsAcceptedAt = &s
	}
	if deletedAt.Valid {
		s := tsToISO(deletedAt)
		user.DeletedAt = &s
	}
	return &user, nil
}

// Returns the count of affected rows (0 if the guest had no trip). Errors
// are returned untouched so callers can wrap with context.
func (r *AuthRepository) ClaimGuestTrips(ctx context.Context, sessionID string, userID uuid.UUID) (int, error) {
	query := `
		UPDATE trips
		SET user_id = $1,
		    session_id = NULL
		WHERE session_id = $2 AND user_id IS NULL
	`
	res, err := r.Pool.Exec(ctx, query, userID, sessionID)
	if err != nil {
		return 0, err
	}
	return int(res.RowsAffected()), nil
}

// SoftDeleteUser marks the account as soft-deleted (deleted_at = now()) and
// returns the email that was set, so callers (e.g. the delete-account
// handler) can clear in-memory caches keyed on it. The email row stays so
// re-registration within the 30-day window (Spec 017 §5.9) keeps the email
// reserved and rate-limit-friendly.
func (r *AuthRepository) SoftDeleteUser(ctx context.Context, userID uuid.UUID) (string, error) {
	var email string
	err := r.Pool.QueryRow(ctx, `
		UPDATE users
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING email
	`, userID).Scan(&email)
	return email, err
}

// SoftDeleteUserCascade is the full GDPR delete: soft-delete
// the user, revoke every active session, and orphan the user's trips by
// clearing user_id + rewriting session_id to a sentinel so analytics
// aggregations still resolve.
//
// All three steps run in a single transaction: a partial failure must not
// leave a user deleted-but-still-signed-in or with trips still pointing at
// them. The trip cascade is intentional — `users.id` has `ON DELETE
// RESTRICT` on the FK (so we don't lose analytics rows referencing the
// user), but the handler pipeline treats `user_id = NULL` as "guest" and
// the sentinel session_id makes those rows easy to grep.
//
// Returns the email of the row we touched (empty + ErrNoRows when the user
// doesn't exist) so the handler can surface a 404 to the client.
func (r *AuthRepository) SoftDeleteUserCascade(ctx context.Context, userID uuid.UUID) (string, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Step 1 — soft-delete user, capture email.
	var email string
	err = tx.QueryRow(ctx, `
		UPDATE users
		SET deleted_at = now(), updated_at = now()
		WHERE id = $1
		RETURNING email
	`, userID).Scan(&email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", pgx.ErrNoRows
		}
		return "", fmt.Errorf("soft-delete user: %w", err)
	}

	// Step 2 — revoke every session belonging to the user.
	if _, err := tx.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID); err != nil {
		return "", fmt.Errorf("revoke sessions: %w", err)
	}

	// Step 3 — orphan trips owned by the user. The sentinel session_id
	// makes it easy to find these rows for cleanup later. The trip's
	// own id is preserved so external share links still resolve to a
	// tombstone instead of a 404.
	if _, err := tx.Exec(ctx, `
		UPDATE trips
		SET user_id = NULL,
		    session_id = 'deleted-' || id::text
		WHERE user_id = $1
	`, userID); err != nil {
		return "", fmt.Errorf("orphan trips: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("commit tx: %w", err)
	}
	return email, nil
}

// MarkUserAsHardDeleted anonymises a user after the 30-day retention window
// (Spec 017 §5.10). The row is kept (FK targets and analytics aggregations
// still resolve), but PII is overwritten:
//   - email becomes `deleted-<uuid>@itinera.invalid`
//   - password_hash is replaced by a fixed-length bcrypt of a random string
//     so it remains in the right shape but is unusable
//   - terms_accepted_at is cleared (no longer relevant)
//
// updated_at is bumped so subsequent cleanup jobs don't re-process the row.
func (r *AuthRepository) MarkUserAsHardDeleted(ctx context.Context, userID uuid.UUID, randomStringHash string) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE users
		SET email = concat('deleted-', $1::text, '@itinera.invalid'),
		    password_hash = $2,
		    terms_accepted_at = NULL,
		    updated_at = now()
		WHERE id = $1 AND deleted_at IS NOT NULL
	`, userID, randomStringHash)
	return err
}

// The caller is responsible for revoking all sessions immediately after.
func (r *AuthRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, newPassword string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}
	_, err = r.Pool.Exec(ctx, `
		UPDATE users
		SET password_hash = $1, updated_at = now()
		WHERE id = $2 AND deleted_at IS NULL
	`, string(hash), userID)
	return err
}

// isSupportedLocale mirrors the CHECK constraint in migration 017 §1
// (locale IN ('en','es','ja','id')). Centralised so handlers don't drift
// from the schema and migrations don't drift from handlers.
func isSupportedLocale(locale string) bool {
	switch locale {
	case "en", "es", "ja", "id":
		return true
	}
	return false
}

// tsToISO renders a pgtype.Timestamptz to a stable ISO-8601 string. We use
// Timestamptz rather than the (now removed) NullTime wrapper because the
// nullable semantics are encoded in pgtype.Timestamptz.Valid directly.
func tsToISO(ts pgtype.Timestamptz) string {
	if !ts.Valid {
		return ""
	}
	return ts.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
}

// HardDeleteExpired anonymises every user whose soft-delete timestamp is
// older than `after`. Each user gets:
//
//   - email replaced with `deleted-<id>@itinera.invalid` (stable, FK-safe)
//   - password_hash replaced with a UNIQUE random bcrypt per user
//     (so the cohort can't be inferred from a shared hash)
//   - terms_accepted_at set to NULL
//   - updated_at bumped
//
// The row is NOT physically deleted so that historical `events` and
// `expenses` rows referencing user_id keep their FK intact for analytics
// aggregation. After 30 days, the user is effectively erased from the
// product but the audit trail stays.
//
// HardDeleteExpired is batched at `limit` rows per call to bound memory
// in case of an unexpected deletion wave (Spec §5.10). Returns the number
// of users anonymised; errors from individual users are logged and
// skipped (the cycle never aborts the whole batch on a single failure).
func (r *AuthRepository) HardDeleteExpired(ctx context.Context, after time.Duration, limit int) (int, error) {
	if limit <= 0 {
		limit = 1000
	}
	days := fmt.Sprintf("%d days", int(after.Hours()/24))

	rows, err := r.Pool.Query(ctx, `
		SELECT id FROM users
		WHERE deleted_at IS NOT NULL
		  AND deleted_at < now() - $1::interval
		  AND email NOT LIKE 'deleted-%@itinera.invalid'
		LIMIT $2
	`, days, limit)
	if err != nil {
		return 0, fmt.Errorf("query soft-deleted users: %w", err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return 0, fmt.Errorf("scan user id: %w", err)
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate soft-deleted users: %w", err)
	}

	if len(ids) == 0 {
		return 0, nil
	}

	anonymised := 0
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return anonymised, err
		}

		// Random bcrypt per user (one cost-12 hash per call, ~100ms). Per
		// Spec §5.10 rationale: a shared hash across the cohort would be
		// a side-channel for an attacker with read access to the DB.
		randomBytes := make([]byte, 32)
		if _, err := rand.Read(randomBytes); err != nil {
			log.Printf("[hard-delete] rand failed for %s: %v", id, err)
			continue
		}
		newHash, err := bcrypt.GenerateFromPassword(randomBytes, bcrypt.DefaultCost)
		if err != nil {
			log.Printf("[hard-delete] bcrypt failed for %s: %v", id, err)
			continue
		}

		// Double-check the WHERE matches the soft-delete predicate. This
		// is the spec's race guard: a /account DELETE that ran between
		// the SELECT and this UPDATE would otherwise be reversed.
		//
		// The `email NOT LIKE 'deleted-%@itinera.invalid'` clause is the
		// idempotency guard: the SELECT above already filters on it, but
		// re-checking here keeps a stale-batch window from re-anonymising
		// the same user across cycles (which would otherwise pin a
		// constant ~100ms bcrypt cost on the same row for every cycle
		// until it leaves the SELECT filter — typically forever).
		_, err = r.Pool.Exec(ctx, `
			UPDATE users
			SET email = 'deleted-' || id::text || '@itinera.invalid',
			    password_hash = $1,
			    terms_accepted_at = NULL,
			    updated_at = now()
			WHERE id = $2
			  AND deleted_at IS NOT NULL
			  AND deleted_at < now() - $3::interval
			  AND email NOT LIKE 'deleted-%@itinera.invalid'
		`, string(newHash), id, days)
		if err != nil {
			log.Printf("[hard-delete] UPDATE failed for %s: %v", id, err)
			continue
		}
		anonymised++
	}

	if anonymised > 1000 {
		log.Printf("[hard-delete] WARN: high volume cycle: %d accounts anonymised (possible bug or deletion wave)", anonymised)
	} else if anonymised > 0 {
		log.Printf("[hard-delete] anonymised %d accounts (window=%s)", anonymised, after)
	}
	return anonymised, nil
}
