package models

import (
	"time"

	"github.com/google/uuid"
)

type Activity struct {
	Id        uuid.UUID  `json:"id"`
	TripId    uuid.UUID  `json:"trip_id"`
	PlaceId   *uuid.UUID `json:"place_id,omitempty"` // NULL = trip-level activity
	Title     string     `json:"title"`
	Date      string     `json:"date"`            // Format: YYYY-MM-DD
	Time      *string    `json:"time,omitempty"`  // Format: HH:MM
	Notes     *string    `json:"notes,omitempty"` // Optional notes
	CreatedAt time.Time  `json:"created_at"`
}
