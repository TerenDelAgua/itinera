package models

import "github.com/google/uuid"

type Place struct {
	ID        uuid.UUID `json:"id"`
	TripId    uuid.UUID `json:"trip_id"`
	Name      string    `json:"name"`
	Notes     string    `json:"notes"`
	StartDate *string   `json:"start_date,omitempty"` // ISO YYYY-MM-DD
	EndDate   *string   `json:"end_date,omitempty"`
	Lat                    *float64  `json:"lat"`
	Lon                    *float64  `json:"lon"`
	DefaultExpenseCurrency *string   `json:"default_expense_currency,omitempty"`
	City                   *string   `json:"city,omitempty"`
	CreatedAt              string    `json:"created_at"`
}
