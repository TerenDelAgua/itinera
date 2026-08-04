package models

import "github.com/google/uuid"

// PasswordHash is always "-"'d out of JSON responses — the only consumer of
// this field is the login flow, which uses bcrypt's constant-time compare.
// DeletedAt is omitempty so non-soft-deleted users don't leak it.
type User struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"` // never sent to clients
	CreatedAt       string    `json:"created_at"`
	Tier            string    `json:"tier"`
	Locale          string    `json:"locale"`
	TermsAcceptedAt *string   `json:"terms_accepted_at,omitempty"`
	UpdatedAt       string    `json:"updated_at"`
	DeletedAt       *string   `json:"deleted_at,omitempty"` // internal-only; included so soft-delete can be distinguished
}
