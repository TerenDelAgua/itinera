package models

import (
	"github.com/google/uuid"
)

// TokenHash is the SHA-256 of the 6-digit code (Spec 017 §4.2). It is never
// serialised to clients; only attempts / locked_at / expires_at are exposed
// for the rate-limit UI.
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt string    `json:"expires_at"`
	UsedAt    *string   `json:"used_at,omitempty"`
	Attempts  int       `json:"attempts"`
	LockedAt  *string   `json:"locked_at,omitempty"`
	CreatedAt string    `json:"created_at"`
	IPAddress *string   `json:"ip_address,omitempty"`
}
