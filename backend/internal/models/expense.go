package models

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	Id         uuid.UUID `json:"id"`
	TripId     uuid.UUID `json:"trip_id"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	CategoryId uuid.UUID `json:"category_id"`
	Notes      string    `json:"notes"`
	Date       time.Time `json:"date"`
	CreatedAt  time.Time `json:"created_at"`
}

type CategorySummary struct {
	CategoryId uuid.UUID `json:"category_id"`
	Total      float64   `json:"total"`
}
