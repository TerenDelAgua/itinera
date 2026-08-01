package models

import (
	"github.com/google/uuid"
)

type Trip struct {
	ID                     uuid.UUID  `json:"id"`
	UserId                 *uuid.UUID `json:"user_id,omitempty"`    // NULL for guests
	SessionId              *string    `json:"session_id,omitempty"` // For guests (maps to cookie)
	Name                   string     `json:"name" example:"Viaje a Japón"`
	Description            *string    `json:"description" example:"Explorando Tokyo y Kyoto"`
	StartDate              string     `json:"start_date" example:"2024-05-01"`
	EndDate                string     `json:"end_date" example:"2024-05-15"`
	BaseCurrency           string     `json:"base_currency" example:"EUR"`
	DefaultExpenseCurrency string     `json:"default_expense_currency" example:"JPY"`
	IsPublicDemo           bool       `json:"is_public_demo" example:"false"`
	ForkedFrom             *uuid.UUID `json:"forked_from,omitempty"`
	CreatedAt              string     `json:"created_at"`

	// Summary fields for Dashboard
	PlaceCount int     `json:"place_count"`
	TotalSpent float64 `json:"total_spent"`

	//Sharing itinerary
	ShareToken     *string `json:"share_token,omitempty"`
	ShareEnabled   bool    `json:"share_enabled"`
	ShareExpiresAt *string `json:"share_expires_at,omitempty"`
	ShareCreatedAt *string `json:"share_created_at,omitempty"`

	IsInternal bool `json:"is_internal" example:"false"`
}
