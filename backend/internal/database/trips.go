package database

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	
	"backend/internal/models"
)

type TripRepository struct {
	Pool *pgxpool.Pool
}

func NewTripRepository(pool *pgxpool.Pool) *TripRepository {
	return &TripRepository{Pool: pool}
}

func (r *TripRepository) CreateTrip(ctx context.Context, userId *uuid.UUID, sessionId *string, tripData models.Trip) (*models.Trip, error) {
	if tripData.BaseCurrency == "" {
		tripData.BaseCurrency = "EUR"
	}
	if tripData.DefaultExpenseCurrency == "" {
		tripData.DefaultExpenseCurrency = tripData.BaseCurrency
	}

	var newTrip models.Trip
	var startDate, endDate, createdAt time.Time

	query := `
	INSERT INTO trips (
		user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	RETURNING 
		id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, created_at
	`

	err := r.Pool.QueryRow(ctx, query,
		userId,
		sessionId,
		tripData.Name,
		tripData.Description,
		tripData.StartDate,
		tripData.EndDate,
		tripData.BaseCurrency,
		tripData.DefaultExpenseCurrency,
	).Scan(
		&newTrip.ID,
		&newTrip.UserId,
		&newTrip.SessionId,
		&newTrip.Name,
		&newTrip.Description,
		&startDate,
		&endDate,
		&newTrip.BaseCurrency,
		&newTrip.DefaultExpenseCurrency,
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

func (r *TripRepository) ListTrips(ctx context.Context, userId *uuid.UUID, sessionId *string) ([]models.Trip, error) {
	query := `
		SELECT 
			id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, created_at,
			COALESCE((SELECT SUM(amount) FROM expenses WHERE trip_id = trips.id), 0) as total_spent,
			COALESCE((SELECT COUNT(*) FROM places WHERE trip_id = trips.id), 0) as place_count
		FROM trips
		WHERE (user_id = $1 AND $1 IS NOT NULL) OR (session_id = $2 AND $2 IS NOT NULL)
		ORDER BY created_at DESC
	`
	rows, err := r.Pool.Query(ctx, query, userId, sessionId)
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
			&trip.DefaultExpenseCurrency,
			&createdAt,
			&trip.TotalSpent,
			&trip.PlaceCount,
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

func (r *TripRepository) GetTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) (*models.Trip, error) {
	var trip models.Trip
	var startDate, endDate, createdAt time.Time

	query := `
		SELECT 
			id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, created_at,
			COALESCE((SELECT SUM(amount) FROM expenses WHERE trip_id = trips.id), 0) as total_spent,
			COALESCE((SELECT COUNT(*) FROM places WHERE trip_id = trips.id), 0) as place_count
		FROM trips
		WHERE id = $1 AND ((user_id = $2 AND $2 IS NOT NULL) OR (session_id = $3 AND $3 IS NOT NULL))
	`

	err := r.Pool.QueryRow(ctx, query, id, userId, sessionId).Scan(
		&trip.ID,
		&trip.UserId,
		&trip.SessionId,
		&trip.Name,
		&trip.Description,
		&startDate,
		&endDate,
		&trip.BaseCurrency,
		&trip.DefaultExpenseCurrency,
		&createdAt,
		&trip.TotalSpent,
		&trip.PlaceCount,
	)
	if err != nil {
		return nil, err
	}

	trip.StartDate = startDate.Format("2006-01-02")
	trip.EndDate = endDate.Format("2006-01-02")
	trip.CreatedAt = createdAt.Format(time.RFC3339)

	return &trip, nil
}

func (r *TripRepository) GetTripById(ctx context.Context, id uuid.UUID) (*models.Trip, error) {
	var trip models.Trip
	var startDate, endDate time.Time

	query := `SELECT id, name, start_date, end_date, base_currency, default_expense_currency FROM trips WHERE id = $1`

	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.Name, &startDate, &endDate, &trip.BaseCurrency, &trip.DefaultExpenseCurrency,
	)
	if err != nil {
		return nil, err
	}

	trip.StartDate = startDate.Format("2006-01-02")
	trip.EndDate = endDate.Format("2006-01-02")
	return &trip, nil
}

func (r *TripRepository) UpdateTrip(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string, updates map[string]any) (*models.Trip, error) {
	allowedFields := map[string]bool{
		"name": true, "start_date": true, "end_date": true,
		"base_currency": true, "description": true, "default_expense_currency": true,
	}

	var setClauses []string
	var args []any
	idx := 1

	for key, val := range updates {
		if !allowedFields[key] {
			continue // Ignorar campos no permitidos por seguridad
		}
		setClauses = append(setClauses, fmt.Sprintf("%s = $%d", key, idx))
		args = append(args, val)
		idx++
	}

	if len(setClauses) == 0 {
		return nil, errors.New("no valid fields to update")
	}

	// Ownership check: solo el propietario (user o session) puede editar
	query := fmt.Sprintf(`
		UPDATE trips 
		SET %s 
		WHERE id = $%d AND (user_id = $%d OR session_id = $%d)
		RETURNING id, user_id, session_id, name, start_date, end_date, base_currency, default_expense_currency, description, created_at
	`, strings.Join(setClauses, ", "), idx, idx+1, idx+2)

	args = append(args, tripID)
	if userID != nil {
		args = append(args, userID, nil)
	} else {
		args = append(args, nil, sessionID)
	}

	var trip models.Trip
	var startDate, endDate, createdAt time.Time

	err := r.Pool.QueryRow(ctx, query, args...).Scan(
		&trip.ID, &trip.UserId, &trip.SessionId, &trip.Name,
		&startDate, &endDate, &trip.BaseCurrency, &trip.DefaultExpenseCurrency,
		&trip.Description, &createdAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("trip not found or unauthorized")
		}
		return nil, err
	}

	trip.StartDate = startDate.Format("2006-01-02")
	trip.EndDate = endDate.Format("2006-01-02")
	trip.CreatedAt = createdAt.Format(time.RFC3339)

	return &trip, nil
}

func (r *TripRepository) DeleteTrip(ctx context.Context, id string, userId *uuid.UUID, sessionId *string) error {
	query := `
		DELETE FROM trips
		WHERE id = $1 AND ((user_id = $2 AND $2 IS NOT NULL) OR (session_id = $3 AND $3 IS NOT NULL))
	`

	res, err := r.Pool.Exec(ctx, query, id, userId, sessionId)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return context.DeadlineExceeded // Or some other "not found" error
	}

	return nil
}
