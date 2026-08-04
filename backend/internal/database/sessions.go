package database

import (
	"backend/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Rotation strategy:
//   - Every successful /refresh issues a new pair of (access, refresh) hashes
//     and updates the same row's refresh_family to a NEW value (Spec 017 §4.3
//     sliding session).
//   - Reusing a refresh hash triggers RevokeFamily for that family, logging
//     the user out across every device that shares it (Spec 017 §5.4).
type SessionRepository struct {
	Pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{Pool: pool}
}

// CreateSession inserts a new session row. The refresh_family is generated
// in the DB (gen_random_uuid) so concurrent session creation never collides.
// expires_at is the refresh expiry — the access expiry is independent and
// tracked client-side via the JWT-style claim in the cookie payload, NOT
// here (this table only stores hashed tokens).
func (r *SessionRepository) CreateSession(ctx context.Context, userID uuid.UUID, accessHash, refreshHash string, expiresAt time.Time, userAgent, ipAddress *string) (*models.Session, error) {
	var s models.Session
	var refreshRotatedAt, revokedAt, lastUsedAt pgtype.Timestamptz
	var ua, ip pgtype.Text

	if userAgent != nil {
		ua.String = *userAgent
		ua.Valid = true
	}
	if ipAddress != nil {
		ip.String = *ipAddress
		ip.Valid = true
	}

	query := `
		INSERT INTO sessions (
			user_id, access_token_hash, refresh_token_hash, refresh_family,
			user_agent, ip_address, expires_at
		)
		VALUES ($1, $2, $3, gen_random_uuid(), $4, $5, $6)
		RETURNING
			id, user_id, access_token_hash, refresh_token_hash, refresh_family,
			refresh_rotated_at, user_agent, host(ip_address), created_at,
			last_used_at, expires_at, revoked_at
	`

	err := r.Pool.QueryRow(ctx, query,
		userID, accessHash, refreshHash, ua, ip, expiresAt,
	).Scan(
		&s.ID,
		&s.UserID,
		&s.AccessTokenHash,
		&s.RefreshTokenHash,
		&s.RefreshFamily,
		&refreshRotatedAt,
		&ua,
		&ip,
		&s.CreatedAt,
		&lastUsedAt,
		&s.ExpiresAt,
		&revokedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("insert session: %w", err)
	}

	if refreshRotatedAt.Valid {
		v := refreshRotatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		s.RefreshRotatedAt = &v
	}
	if revokedAt.Valid {
		v := revokedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		s.RevokedAt = &v
	}
	if ua.Valid {
		v := ua.String
		s.UserAgent = &v
	}
	if ip.Valid {
		v := ip.String
		s.IPAddress = &v
	}
	return &s, nil
}

// FindSessionByAccessTokenHash is the lookup the auth middleware runs on
// every protected request. The previous_use_at field is bumped on success
// to give users a meaningful "last seen at" date in the security log.
func (r *SessionRepository) FindSessionByAccessTokenHash(ctx context.Context, accessHash string) (*models.Session, error) {
	if accessHash == "" {
		return nil, pgx.ErrNoRows
	}

	var s models.Session
	var refreshRotatedAt, revokedAt pgtype.Timestamptz
	var ua, ip pgtype.Text

	err := r.Pool.QueryRow(ctx, `
		UPDATE sessions
		SET last_used_at = now()
		WHERE access_token_hash = $1
		  AND revoked_at IS NULL
		  AND expires_at > now()
		RETURNING
			id, user_id, access_token_hash, refresh_token_hash, refresh_family,
			refresh_rotated_at, user_agent, host(ip_address), created_at,
			last_used_at, expires_at, revoked_at
	`, accessHash).Scan(
		&s.ID,
		&s.UserID,
		&s.AccessTokenHash,
		&s.RefreshTokenHash,
		&s.RefreshFamily,
		&refreshRotatedAt,
		&ua,
		&ip,
		&s.CreatedAt,
		&s.LastUsedAt,
		&s.ExpiresAt,
		&revokedAt,
	)
	if err != nil {
		return nil, err
	}

	if refreshRotatedAt.Valid {
		v := refreshRotatedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		s.RefreshRotatedAt = &v
	}
	if revokedAt.Valid {
		v := revokedAt.Time.UTC().Format("2006-01-02T15:04:05Z07:00")
		s.RevokedAt = &v
	}
	if ua.Valid {
		v := ua.String
		s.UserAgent = &v
	}
	if ip.Valid {
		v := ip.String
		s.IPAddress = &v
	}
	return &s, nil
}

// RotateSession swaps the access + refresh hashes on a session row and
// extends the expiry. Callers verify the previous refresh hash matches the
// supplied one BEFORE invoking this; otherwise the unique-key collision
// raises an error and the caller revokes the family.
func (r *SessionRepository) RotateSession(ctx context.Context, sessionID uuid.UUID, newAccessHash, newRefreshHash string, newExpiry time.Time) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE sessions
		SET access_token_hash = $1,
		    refresh_token_hash = $2,
		    refresh_family = gen_random_uuid(),
		    refresh_rotated_at = now(),
		    last_used_at = now(),
		    expires_at = $3
		WHERE id = $4 AND revoked_at IS NULL
	`, newAccessHash, newRefreshHash, newExpiry, sessionID)
	if err != nil {
		return fmt.Errorf("rotate session: %w", err)
	}
	return nil
}

// RevokeSession sets revoked_at to now on a single row. Used by /auth/logout.
func (r *SessionRepository) RevokeSession(ctx context.Context, sessionID uuid.UUID) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = now()
		WHERE id = $1 AND revoked_at IS NULL
	`, sessionID)
	return err
}

// RevokeFamily marks every session belonging to the family as revoked. The
// classic "I lost my phone, log out everywhere" button hits this directly.
func (r *SessionRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = now()
		WHERE refresh_family = $1 AND revoked_at IS NULL
	`, familyID)
	return err
}

// RevokeAllSessionsForUser is used by the GDPR delete-account endpoint
// and by /auth/logout-everywhere.
func (r *SessionRepository) RevokeAllSessionsForUser(ctx context.Context, userID uuid.UUID) error {
	_, err := r.Pool.Exec(ctx, `
		UPDATE sessions
		SET revoked_at = now()
		WHERE user_id = $1 AND revoked_at IS NULL
	`, userID)
	return err
}

// CountActiveSessionsForUser is what /auth/sessions (TBD) and the in-account
// device manager will read. Used to enforce any cap we may add later on
// concurrent sessions (currently uncapped, per Spec §4.4).
func (r *SessionRepository) CountActiveSessionsForUser(ctx context.Context, userID uuid.UUID) (int, error) {
	var n int
	err := r.Pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM sessions
		WHERE user_id = $1 AND revoked_at IS NULL AND expires_at > now()
	`, userID).Scan(&n)
	return n, err
}

// CleanupExpiredSessions removes revoked or expired sessions older than the
// retention window.
//
// Returns the number of rows deleted so callers can log the count without
// an extra round trip.
func (r *SessionRepository) CleanupExpiredSessions(ctx context.Context) (int64, error) {
	res, err := r.Pool.Exec(ctx, `
		DELETE FROM sessions
		WHERE revoked_at IS NOT NULL OR expires_at < now() - interval '7 days'
	`)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected(), nil
}
