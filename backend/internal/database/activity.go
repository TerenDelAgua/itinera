package database

import (
	"backend/internal/models"
	"context"

	"github.com/google/uuid"
)

func (db *DB) GetByTripId(ctx context.Context, tripId uuid.UUID) ([]models.Activity, error) {
	query := `
		SELECT id, trip_id, place_id, title, date, time, notes, created_at
		FROM activities
		WHERE trip_id = $1
		ORDER BY date ASC, time ASC NULLS LAST
		`
	rows, err := db.Pool.Query(ctx, query, tripId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.Activity

	for rows.Next() {
		var act models.Activity
		err := rows.Scan(
			&act.Id,
			&act.TripId,
			&act.PlaceId,
			&act.Title,
			&act.Date,
			&act.Time,
			&act.Notes,
			&act.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, act)
	}

	return activities, nil
}

func (db *DB) GetByPlaceId(ctx context.Context, placeId uuid.UUID) ([]models.Activity, error) {
	query := `
		SELECT id, trip_id, place_id, title, date, time, notes, created_at
		FROM activities
		WHERE place_id = $1
		ORDER BY date ASC, time ASC NULLS LAST
	`

	rows, err := db.Pool.Query(ctx, query, placeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities []models.Activity
	for rows.Next() {
		var act models.Activity
		err := rows.Scan(
			&act.Id,
			&act.TripId,
			&act.PlaceId,
			&act.Title,
			&act.Date,
			&act.Time,
			&act.Notes,
			&act.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		activities = append(activities, act)
	}

	return activities, nil
}

func (db *DB) CreateActivity(ctx context.Context, act *models.Activity) error {
	query := `
		INSERT INTO activities (id, trip_id, place_id, title, date, time, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`
	_, err := db.Pool.Exec(ctx, query,
		act.Id,
		act.TripId,
		act.PlaceId,
		act.Title,
		act.Date,
		act.Time,
		act.Notes,
	)

	return err
}

func (db *DB) UpdateActivity(ctx context.Context, act *models.Activity) error {
	query := `
		UPDATE activities
		SET
			title = $1,
			date = $2,
			time = $3,
			notes = $4,
			place_id = $5
		WHERE id = $6
	`

	_, err := db.Pool.Exec(ctx, query,
		act.Title,
		act.Date,
		act.Time,
		act.Notes,
		act.PlaceId,
		act.Id,
	)

	return err
}

func (db *DB) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	query := " DELETE FROM activities WHERE id = $1"

	_, err := db.Pool.Exec(ctx, query, id)

	return err
}
