package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"backend/internal/models"
)

type PlaceRepository struct {
	Pool *pgxpool.Pool
}

func NewPlaceRepository(pool *pgxpool.Pool) *PlaceRepository {
	return &PlaceRepository{Pool: pool}
}

func (r *PlaceRepository) GetPlace(ctx context.Context, placeId uuid.UUID) (*models.Place, error) {
	var p models.Place
	var startDate, endDate *time.Time
	var createdAt time.Time
	query := `SELECT id, trip_id, name, COALESCE(notes, ''), start_date, end_date, lat, lon, default_expense_currency, city, country_code, created_at 
		FROM places WHERE id = $1`
	err := r.Pool.QueryRow(ctx, query, placeId).
		Scan(&p.ID, &p.TripId, &p.Name, &p.Notes, &startDate, &endDate, &p.Lat, &p.Lon, &p.DefaultExpenseCurrency, &p.City, &p.CountryCode, &createdAt)
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

func (r *PlaceRepository) ListPlacesByTrip(ctx context.Context, tripID uuid.UUID) ([]models.Place, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, trip_id, name, COALESCE(notes, ''), start_date, end_date, lat, lon, default_expense_currency, city, country_code, created_at 
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
		if err := rows.Scan(&p.ID, &p.TripId, &p.Name, &p.Notes, &startDate, &endDate, &p.Lat, &p.Lon, &p.DefaultExpenseCurrency, &p.City, &p.CountryCode, &createdAt); err != nil {
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

func (r *PlaceRepository) CreatePlace(ctx context.Context, tripID uuid.UUID, p models.Place) (*models.Place, error) {
	query := `INSERT INTO places (trip_id, name, notes, start_date, end_date, lat, lon, default_expense_currency, city, country_code) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id, trip_id, name, COALESCE(notes, ''), start_date, end_date, lat, lon, default_expense_currency, city, country_code, created_at`
	var res models.Place
	var startDate, endDate *time.Time
	var createdAt time.Time
	err := r.Pool.QueryRow(ctx, query, tripID, p.Name, p.Notes, p.StartDate, p.EndDate, p.Lat, p.Lon, p.DefaultExpenseCurrency, p.City, p.CountryCode).
		Scan(&res.ID, &res.TripId, &res.Name, &res.Notes, &startDate, &endDate, &res.Lat, &res.Lon, &res.DefaultExpenseCurrency, &res.City, &res.CountryCode, &createdAt)
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

func (r *PlaceRepository) UpdatePlace(ctx context.Context, placeID uuid.UUID, updates map[string]any) (*models.Place, error) {
	allowed := map[string]bool{"name": true, "notes": true, "start_date": true, "end_date": true, "default_expense_currency": true, "city": true, "country_code": true}
	var sets []string
	var args []any
	i := 1
	for k, v := range updates {
		if allowed[k] {
			sets = append(sets, fmt.Sprintf("%s = $%d", k, i))
			args = append(args, v)
			i++
		}
	}
	if len(sets) == 0 {
		return nil, nil
	}

	query := fmt.Sprintf(
		"UPDATE places SET %s WHERE id = $%d RETURNING id, trip_id, name, COALESCE(notes, ''), start_date, end_date, lat, lon, default_expense_currency, city, country_code, created_at",
		strings.Join(sets, ", "),
		i,
	)
	args = append(args, placeID)

	var res models.Place
	var startDate, endDate *time.Time
	var createdAt time.Time
	err := r.Pool.QueryRow(ctx, query, args...).Scan(&res.ID, &res.TripId, &res.Name, &res.Notes, &startDate, &endDate, &res.Lat, &res.Lon, &res.DefaultExpenseCurrency, &res.City, &res.CountryCode, &createdAt)
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

func (r *PlaceRepository) DeletePlace(ctx context.Context, placeID uuid.UUID) error {
	_, err := r.Pool.Exec(ctx, "DELETE FROM places WHERE id = $1", placeID)
	return err
}
