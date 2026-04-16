package database

import (
	"context"

	"github.com/google/uuid"

	"backend/internal/models"
)

func (db *DB) CreateTrip(ctx context.Context, userId *uuid.UUID, sessionId *string, tripData models.Trip) (*models.Trip, error) {

	var newTrip models.Trip

	query := `
	INSERT INTO trips (
		user_id, session_id, name, start_date, end_date, base_currency
	) VALUES ($1, $2, $3, $4, $5, $6)
	RETURNING 
		id, user_id, session_id, name, start_date, end_date, base_currency, created_at
	`

	err := db.Pool.QueryRow(ctx, query,
		userId,
		sessionId,
		tripData.Name,
		tripData.StartDate,
		tripData.EndDate,
		tripData.BaseCurrency,
	).Scan(
		&newTrip.ID,
		&newTrip.UserId,
		&newTrip.SessionId,
		&newTrip.Name,
		&newTrip.StartDate,
		&newTrip.EndDate,
		&newTrip.BaseCurrency,
		&newTrip.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &newTrip, nil
}
