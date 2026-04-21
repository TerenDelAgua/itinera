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
		user_id, session_id, name, description, start_date, end_date, base_currency
	) VALUES ($1, $2, $3, $4, $5, $6, $7)
	RETURNING 
		id, user_id, session_id, name, description, start_date, end_date, base_currency, created_at
	`

	err := db.Pool.QueryRow(ctx, query,
		userId,
		sessionId,
		tripData.Name,
		tripData.Description,
		tripData.StartDate,
		tripData.EndDate,
		tripData.BaseCurrency,
	).Scan(
		&newTrip.ID,
		&newTrip.UserId,
		&newTrip.SessionId,
		&newTrip.Name,
		&newTrip.Description,
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
		SELECT id, user_id, session_id, name, description, start_date, end_date, base_currency, created_at
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
			&trip.Description,
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

func (db *DB) GetTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) (*models.Trip, error) {
	var trip models.Trip
	var startDate, endDate, createdAt time.Time

	query := `
		SELECT id, user_id, session_id, name, description, start_date, end_date, base_currency, created_at
		FROM trips
		WHERE id = $1 AND ((user_id = $2 AND $2 IS NOT NULL) OR (session_id = $3 AND $3 IS NOT NULL))
	`

	err := db.Pool.QueryRow(ctx, query, id, userId, sessionId).Scan(
		&trip.ID,
		&trip.UserId,
		&trip.SessionId,
		&trip.Name,
		&trip.Description,
		&startDate,
		&endDate,
		&trip.BaseCurrency,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	trip.StartDate = startDate.Format("2006-01-02")
	trip.EndDate = endDate.Format("2006-01-02")
	trip.CreatedAt = createdAt.Format(time.RFC3339)

	return &trip, nil
}

func (db *DB) UpdateTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string, tripData models.Trip) (*models.Trip, error) {
	var trip models.Trip
	var startDate, endDate, createdAt time.Time

	query := `
		UPDATE trips
		SET name = $1, description = $2, start_date = $3, end_date = $4, base_currency = $5
		WHERE id = $6 AND ((user_id = $7 AND $7 IS NOT NULL) OR (session_id = $8 AND $8 IS NOT NULL))
		RETURNING id, user_id, session_id, name, description, start_date, end_date, base_currency, created_at
	`

	err := db.Pool.QueryRow(ctx, query,
		tripData.Name,
		tripData.Description,
		tripData.StartDate,
		tripData.EndDate,
		tripData.BaseCurrency,
		id,
		userId,
		sessionId,
	).Scan(
		&trip.ID,
		&trip.UserId,
		&trip.SessionId,
		&trip.Name,
		&trip.Description,
		&startDate,
		&endDate,
		&trip.BaseCurrency,
		&createdAt,
	)
	if err != nil {
		return nil, err
	}

	trip.StartDate = startDate.Format("2006-01-02")
	trip.EndDate = endDate.Format("2006-01-02")
	trip.CreatedAt = createdAt.Format(time.RFC3339)

	return &trip, nil
}

func (db *DB) DeleteTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) error {
	query := `
		DELETE FROM trips
		WHERE id = $1 AND ((user_id = $2 AND $2 IS NOT NULL) OR (session_id = $3 AND $3 IS NOT NULL))
	`

	res, err := db.Pool.Exec(ctx, query, id, userId, sessionId)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return context.DeadlineExceeded // Or some other "not found" error
	}

	return nil
}
