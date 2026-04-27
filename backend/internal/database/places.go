package database

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend/internal/models"
)

func (db *DB) GetPlace(ctx context.Context, placeId uuid.UUID) (*models.Place, error) {
	var p models.Place
	var startDate, endDate *time.Time
	var createdAt time.Time
	query := `SELECT id, trip_id, name, notes, start_date, end_date, lat, lon, created_at 
		FROM places WHERE id = $1`
	err := db.Pool.QueryRow(ctx, query, placeId).
		Scan(&p.ID, &p.TripId, &p.Name, &p.Notes, &startDate, &endDate, &p.Lat, &p.Lon, &createdAt)
	if err != nil {
		return nil, err
	}
	if startDate != nil {
		s := startDate.Format("2006-01-02")
		p.StartDate = &s
	}
	if endDate != nil {
		e := endDate.Format("2006-01-02")
		p.EndDate = &e
	}
	p.CreatedAt = createdAt.Format(time.RFC3339)
	return &p, nil
}

func (db *DB) ListPlacesByTrip(ctx context.Context, tripID uuid.UUID) ([]models.Place, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, trip_id, name, notes, start_date, end_date, lat, lon, created_at 
		FROM places WHERE trip_id = $1 ORDER BY start_date ASC NULLS LAST, created_at ASC`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var places []models.Place
	for rows.Next() {
		var p models.Place
		var startDate, endDate *time.Time
		var createdAt time.Time
		if err := rows.Scan(&p.ID, &p.TripId, &p.Name, &p.Notes, &startDate, &endDate, &p.Lat, &p.Lon, &createdAt); err != nil {
			return nil, err
		}
		if startDate != nil {
			s := startDate.Format("2006-01-02")
			p.StartDate = &s
		}
		if endDate != nil {
			e := endDate.Format("2006-01-02")
			p.EndDate = &e
		}
		p.CreatedAt = createdAt.Format(time.RFC3339)
		places = append(places, p)
	}
	return places, nil
}

func (db *DB) CreatePlace(ctx context.Context, tripID uuid.UUID, p models.Place) (*models.Place, error) {
	query := `INSERT INTO places (trip_id, name, notes, start_date, end_date, lat, lon) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, trip_id, name, notes, start_date, end_date, lat, lon, created_at`
	var res models.Place
	var startDate, endDate *time.Time
	var createdAt time.Time
	err := db.Pool.QueryRow(ctx, query, tripID, p.Name, p.Notes, p.StartDate, p.EndDate, p.Lat, p.Lon).
		Scan(&res.ID, &res.TripId, &res.Name, &res.Notes, &startDate, &endDate, &res.Lat, &res.Lon, &createdAt)
	if startDate != nil {
		s := startDate.Format("2006-01-02")
		res.StartDate = &s
	}
	if endDate != nil {
		e := endDate.Format("2006-01-02")
		res.EndDate = &e
	}
	res.CreatedAt = createdAt.Format(time.RFC3339)
	return &res, err
}

func (db *DB) UpdatePlace(ctx context.Context, placeID uuid.UUID, updates map[string]any) (*models.Place, error) {
	allowed := map[string]bool{"name": true, "notes": true, "start_date": true, "end_date": true}
	var sets []string
	var args []any
	i := 1
	for k, v := range updates {
		if allowed[k] {
			sets = append(sets, k+" = $"+string(rune(i+48))) // $1, $2...
			args = append(args, v)
			i++
		}
	}
	if len(sets) == 0 {
		return nil, nil
	}

	query := "UPDATE places SET " + strings.Join(sets, ", ") + " WHERE id = $" + string(rune(i+48)) + " RETURNING id, trip_id, name, notes, start_date, end_date, lat, lon, created_at"
	args = append(args, placeID)

	var res models.Place
	var startDate, endDate *time.Time
	var createdAt time.Time
	err := db.Pool.QueryRow(ctx, query, args...).Scan(&res.ID, &res.TripId, &res.Name, &res.Notes, &startDate, &endDate, &res.Lat, &res.Lon, &createdAt)
	if startDate != nil {
		s := startDate.Format("2006-01-02")
		res.StartDate = &s
	}
	if endDate != nil {
		e := endDate.Format("2006-01-02")
		res.EndDate = &e
	}
	res.CreatedAt = createdAt.Format(time.RFC3339)
	return &res, err
}

func (db *DB) DeletePlace(ctx context.Context, placeID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, "DELETE FROM places WHERE id = $1", placeID)
	return err
}
