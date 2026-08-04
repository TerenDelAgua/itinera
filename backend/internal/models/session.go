package models

import (
	"github.com/google/uuid"
)

// Tokens never leave the server in plaintext: only their SHA-256 hashes are
// stored, and the hashes are themselves hidden from JSON responses so a
// leaked log entry cannot be replayed against the API.
//
// RefreshFamily groups every successive rotation of the same refresh token:
// on reuse detection the server revokes every row sharing the family, which
// logs the user out across all devices (Spec 017 §5.4 reuse detection).
type Session struct {
	ID               uuid.UUID `json:"id"`
	UserID           uuid.UUID `json:"user_id"`
	AccessTokenHash  string    `json:"-"` // never serialised to clients
	RefreshTokenHash string    `json:"-"` // never serialised to clients
	RefreshFamily    uuid.UUID `json:"refresh_family"`
	RefreshRotatedAt *string   `json:"refresh_rotated_at,omitempty"`
	UserAgent        *string   `json:"user_agent,omitempty"`
	IPAddress        *string   `json:"ip_address,omitempty"` // rendered as text; INET in the schema
	CreatedAt        string    `json:"created_at"`
	LastUsedAt       string    `json:"last_used_at"`
	ExpiresAt        string    `json:"expires_at"`
	RevokedAt        *string   `json:"revoked_at,omitempty"`
}
