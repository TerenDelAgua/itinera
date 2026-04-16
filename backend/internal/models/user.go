package models

import "github.com/google/uuid"

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Exclude from JSON responses
	CreatedAt    string    `json:"created_at"`
}
