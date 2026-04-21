package database

import (
	"context"
	"time"

	"github.com/google/uuid"

	"backend/internal/models"
)

func (db *DB) CreateTrip(ctx context.Context, userId *uuid.UUID, sessionId *string, tripData models.Trip) (*models.Trip, error) {

	var newTrip models.Trip
	var startDate, endDate, createdAt time.Time

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
		&startDate,
		&endDate,
		&newTrip.BaseCurrency,
		&createdAt,
	)
	if err != nil {
		return nil, err

	}

	newTrip.StartDate = startDate.Format("2006-01-02")
	newTrip.EndDate = endDate.Format("2006-01-02")
	newTrip.CreatedAt = createdAt.Format(time.RFC3339)

	return &newTrip, nil
}

func (db *DB) ListTrips(ctx context.Context, userId *uuid.UUID, sessionId *string) ([]models.Trip, error) {
	query := `
		SELECT id, user_id, session_id, name, start_date, end_date, base_currency, created_at
		FROM trips
		WHERE (user_id = $1 AND $1 IS NOT NULL) OR (session_id = $2 AND $2 IS NOT NULL)
		ORDER BY created_at DESC
	`
	rows, err := db.Pool.Query(ctx, query, userId, sessionId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trips []models.Trip
	for rows.Next() {
		var trip models.Trip
		var startDate, endDate, createdAt time.Time
		if err := rows.Scan(
			&trip.ID,
			&trip.UserId,
			&trip.SessionId,
			&trip.Name,
			&startDate,
			&endDate,
			&trip.BaseCurrency,
			&createdAt,
		); err != nil {
			return nil, err
		}
		trip.StartDate = startDate.Format("2006-01-02")
		trip.EndDate = endDate.Format("2006-01-02")
		trip.CreatedAt = createdAt.Format(time.RFC3339)
		trips = append(trips, trip)
	}
	
	if trips == nil {
		trips = []models.Trip{} // Return empty array instead of null
	}
	
	return trips, nil
}