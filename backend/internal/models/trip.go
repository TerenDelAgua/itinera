package models

import (
	"github.com/google/uuid"
)

type Trip struct {
	ID           uuid.UUID  `json:"id"`
	UserId       *uuid.UUID `json:"user_id,omitempty"`    // NULL for guests
	SessionId    *string    `json:"session_id,omitempty"` // For guests (maps to cookie)
	Name         string     `json:"name"`
	Description  *string    `json:"description"`
	StartDate    string     `json:"start_date"` // ISO format
	EndDate      string     `json:"end_date"`   // ISO format
	BaseCurrency string     `json:"base_currency"`
	CreatedAt    string     `json:"created_at"`
}
