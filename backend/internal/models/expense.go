package models

import (
	"time"

	"github.com/google/uuid"
)

type Expense struct {
	Id      uuid.UUID  `json:"id"`
	TripId  uuid.UUID  `json:"trip_id"`
	PlaceId *uuid.UUID `json:"place_id"`

	Amount           float64   `json:"amount"`
	OriginalAmount   float64   `json:"original_amount"`
	OriginalCurrency string    `json:"original_currency"`
	ExchangeRate     float64   `json:"exchange_rate"`
	ConversionDate   time.Time `json:"conversion_date"`

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

type PlaceSummary struct {
	PlaceId   uuid.UUID `json:"place_id"`
	PlaceName string    `json:"place_name"`
	Total     float64   `json:"total"`
}

type TripExpenseSummary struct {
	GlobalTotal float64           `json:"global_total"`
	PlacesTotal float64           `json:"places_total"`
	GrandTotal  float64           `json:"grand_total"`
	ByCategory  []CategorySummary `json:"by_category"`
	ByPlace     []PlaceSummary    `json:"by_place"`
}
