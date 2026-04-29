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
	StartDate              string     `json:"start_date"` // ISO format
	EndDate                string     `json:"end_date"`   // ISO format
	BaseCurrency           string     `json:"base_currency"`
	DefaultExpenseCurrency string     `json:"default_expense_currency"`
	CreatedAt              string     `json:"created_at"`

	// Summary fields for Dashboard
	PlaceCount int     `json:"place_count"`
	TotalSpent float64 `json:"total_spent"`
}
