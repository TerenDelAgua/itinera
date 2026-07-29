package database

import (
	"context"
	"crypto/rand"
	"encoding/base64"
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
		user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, is_public_demo
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	RETURNING 
		id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, is_public_demo, forked_from, created_at
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
		tripData.IsPublicDemo,
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
		&newTrip.IsPublicDemo,
		&newTrip.ForkedFrom,
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
			id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, 
			is_public_demo, forked_from, created_at,
			COALESCE((SELECT SUM(amount) FROM expenses WHERE trip_id = trips.id), 0) as total_spent,
			COALESCE((SELECT COUNT(*) FROM places WHERE trip_id = trips.id), 0) as place_count
		FROM trips
		WHERE (user_id = $1 AND $1 IS NOT NULL) OR (session_id = $2 AND $2 IS NOT NULL) OR (is_public_demo = true)
		ORDER BY 
			is_public_demo ASC, 
			CASE WHEN is_public_demo = false THEN created_at END DESC NULLS LAST,
			CASE WHEN is_public_demo = true AND (name ILIKE '%Japón%' OR name ILIKE '%Japan%') THEN 0 ELSE 1 END ASC,
			created_at DESC
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
			&trip.IsPublicDemo,
			&trip.ForkedFrom,
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
			id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, 
			is_public_demo, forked_from, created_at,
			COALESCE((SELECT SUM(amount) FROM expenses WHERE trip_id = trips.id), 0) as total_spent,
			COALESCE((SELECT COUNT(*) FROM places WHERE trip_id = trips.id), 0) as place_count
		FROM trips
		WHERE id = $1 AND ((user_id = $2 AND $2 IS NOT NULL) OR (session_id = $3 AND $3 IS NOT NULL) OR (is_public_demo = true))
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
		&trip.IsPublicDemo,
		&trip.ForkedFrom,
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

	query := `SELECT id, name, start_date, end_date, base_currency, default_expense_currency, forked_from FROM trips WHERE id = $1`

	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&trip.ID, &trip.Name, &startDate, &endDate, &trip.BaseCurrency, &trip.DefaultExpenseCurrency, &trip.ForkedFrom,
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

func (r *TripRepository) GetFork(ctx context.Context, forkedFrom string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
	var trip models.Trip
	query := `
		SELECT id FROM trips 
		WHERE forked_from = $1 AND ((user_id = $2 AND $2 IS NOT NULL) OR (session_id = $3 AND $3 IS NOT NULL))
		LIMIT 1
	`
	err := r.Pool.QueryRow(ctx, query, forkedFrom, userID, sessionID).Scan(&trip.ID)
	if err != nil {
		return nil, err
	}
	return &trip, nil
}

func (r *TripRepository) GetTripMeta(ctx context.Context, tripID string, userID *uuid.UUID, sessionID *string) (bool, bool, error) {
	var isOwner, isDemo bool
	query := `
		SELECT 
			((user_id = $2 AND $2 IS NOT NULL) OR (session_id = $3 AND $3 IS NOT NULL)), 
			is_public_demo 
		FROM trips WHERE id = $1
	`
	err := r.Pool.QueryRow(ctx, query, tripID, userID, sessionID).Scan(&isOwner, &isDemo)
	if err != nil {
		return false, false, err
	}
	return isOwner, isDemo, nil
}
func (r *TripRepository) ListPublicDemos(ctx context.Context, limit int) ([]models.Trip, error) {
	query := `
		SELECT 
			id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, 
			is_public_demo, forked_from, created_at,
			COALESCE((SELECT SUM(amount) FROM expenses WHERE trip_id = trips.id), 0) as total_spent,
			COALESCE((SELECT COUNT(*) FROM places WHERE trip_id = trips.id), 0) as place_count
		FROM trips
		WHERE is_public_demo = true
		ORDER BY created_at DESC
		LIMIT $1
	`
	rows, err := r.Pool.Query(ctx, query, limit)
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
			&trip.IsPublicDemo,
			&trip.ForkedFrom,
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
		trips = []models.Trip{}
	}

	return trips, nil
}

func (r *TripRepository) GetPublicStats(ctx context.Context) (int, error) {
	var count int
	err := r.Pool.QueryRow(ctx, "SELECT total_trips_count FROM public_stats WHERE id = 1").Scan(&count)
	if err != nil {
		// Fallback to real count if stats table fails
		err = r.Pool.QueryRow(ctx, "SELECT COUNT(*) FROM trips").Scan(&count)
	}
	return count, err
}

func (r *TripRepository) IsPublicDemo(ctx context.Context, tripID uuid.UUID) (bool, error) {
	var isDemo bool
	err := r.Pool.QueryRow(ctx,
		`SELECT is_public_demo FROM trips WHERE id = $1`, tripID,
	).Scan(&isDemo)

	// If the trip does not exist, it is by definition not a public demo.
	// Any other error is a real database failure and must be propagated.
	if err == pgx.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return isDemo, nil
}

func (r *TripRepository) GetActiveShareTripID(ctx context.Context, token string) (uuid.UUID, error) {
	var id uuid.UUID
	err := r.Pool.QueryRow(ctx,
		`SELECT id FROM trips
	WHERE share_token = $1
		AND share_enabled = true
		AND (share_expires_at IS NULL OR share_expires_at > NOW())`,
		token).Scan(&id)

	return id, err
}

func (r *TripRepository) GetBaseForClone(ctx context.Context, tx pgx.Tx, origID uuid.UUID) (*models.Trip, error) {
	var t models.Trip
	var startDate, endDate time.Time
	err := tx.QueryRow(ctx,
		`SELECT name, description, start_date, end_date, base_currency, default_expense_currency
FROM trips WHERE id = $1`,
		origID).Scan(
		&t.Name,
		&t.Description,
		&startDate,
		&endDate,
		&t.BaseCurrency,
		&t.DefaultExpenseCurrency,
	)
	if err != nil {
		return nil, err
	}
	t.StartDate = startDate.Format("2006-01-02")
	t.EndDate = endDate.Format("2006-01-02")
	return &t, nil
}

func (r *TripRepository) InsertFork(
	ctx context.Context,
	tx pgx.Tx,
	newTripID uuid.UUID,
	userID *uuid.UUID,
	sessionID *string,
	origID uuid.UUID,
	base *models.Trip,
) error {
	startDate, _ := time.Parse("2006-01-02", base.StartDate)
	endDate, _ := time.Parse("2006-01-02", base.EndDate)
	_, err := tx.Exec(ctx,
		`INSERT INTO trips (id, user_id, session_id, name, description,
							start_date, end_date, base_currency,
							default_expense_currency, is_public_demo, forked_from)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`,
		newTripID, userID, sessionID, base.Name, base.Description, startDate,
		endDate, base.BaseCurrency, base.DefaultExpenseCurrency, base.IsPublicDemo, origID)

	return err
}

func (r *TripRepository) EnableShare(
	ctx context.Context,
	tripID string,
	userID *uuid.UUID,
	sessionID *string,
) (string, *time.Time, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return "", nil, err
	}
	defer tx.Rollback(ctx)

	var ownerUserID *uuid.UUID
	var ownerSessionID *string
	err = tx.QueryRow(ctx,
		`SELECT user_id, session_id FROM trips WHERE id = $1`, tripID,
	).Scan(&ownerUserID, &ownerSessionID)

	if err != nil {
		if err == pgx.ErrNoRows {
			return "", nil, errors.New("trip not found")
		}
		return "", nil, err
	}

	isOwner := (userID != nil && ownerUserID != nil && *userID == *ownerUserID) ||
		(sessionID != nil && ownerSessionID != nil && *sessionID == *ownerSessionID)
	if !isOwner {
		return "", nil, errors.New("unauthorized")
	}

	var existingToken *string
	var existingExpiry *time.Time
	err = tx.QueryRow(ctx,
		`SELECT share_token, share_expires_at 
			FROM trips WHERE id = $1 AND share_enabled = true`,
		tripID).Scan(&existingToken, &existingExpiry)

	if err == nil && existingToken != nil {
		if existingExpiry == nil || existingExpiry.After(time.Now()) {
			if err := tx.Commit(ctx); err == nil {
				return "", nil, err
			}

		}
		return *existingToken, existingExpiry, nil
	}

	newToken, err := generateShareToken()
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}

	var newExpiry *time.Time
	if ownerUserID != nil {
		newExpiry = nil // permanent
	} else {
		exp := time.Now().Add(7 * 25 * time.Hour)
		newExpiry = &exp
	}

	_, err = tx.Exec(ctx,
		`UPDATE trips
			SET share_token = $1, share_enabled = true,
				share_expires_at = $2, share_created_at = NOW()
			WHERE id = $3`,
		newToken, newExpiry, tripID)
	if err != nil {
		return "", nil, fmt.Errorf("failed to enable share: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return "", nil, err
	}
	return newToken, newExpiry, nil
}

func (r *TripRepository) DisableShare(
	ctx context.Context,
	tripID string,
	userID *uuid.UUID,
	sessionID *string,
) error {
	res, err := r.Pool.Exec(ctx,
		`UPDATE trips
			SET share_token = NULL, share_enabled = false,
				share_expires_at = NULL, share_created_at = NULL
			WHERE id = $1
				AND ((user_id = $2 AND IS NOT NULL) OR (session_id = $3 AND IS NOT NULL))`,
		tripID, userID, sessionID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return errors.New("trip not found or unauthorized")
	}
	return nil
}

func (r *TripRepository) GetByShareToken(ctx context.Context, token string) (*models.Trip, error) {
	var t models.Trip
	var startDate, endDate, createdAt time.Time
	var shareExpiresAt, shareCreatedAt *time.Time

	err := r.Pool.QueryRow(ctx,
		`SELECT id, user_id, session_id, name, description,
					start_date, end_date, base_currency, default_expense_currency,
					is_public_demo, forked_from, created_at
					share_token, share_enabled, share_expires_at, share_created_at
			FROM trips
			WHERE share_token = $1
				AND share_enabled = true
				AND (share_expires_at IS NULL OR share_expires_at > NOW())
		`, token).Scan(
		&t.ID,
		&t.UserId,
		&t.SessionId,
		&t.Name,
		&t.Description,
		&startDate,
		&endDate,
		&t.BaseCurrency,
		&t.DefaultExpenseCurrency,
		&t.IsPublicDemo,
		&t.ForkedFrom,
		&createdAt,
		&t.ShareToken,
		&t.ShareEnabled,
		&shareExpiresAt,
		&shareCreatedAt,
	)
	if err != nil {
		return nil, err
	}

	t.StartDate = startDate.Format("2006-01-02")
	t.EndDate = endDate.Format("2006-01-02")
	t.CreatedAt = createdAt.Format(time.RFC3339)
	if shareExpiresAt != nil {
		s := shareExpiresAt.Format(time.RFC3339)
		t.ShareExpiresAt = &s
	}
	if shareCreatedAt != nil {
		s := shareCreatedAt.Format(time.RFC3339)
		t.ShareCreatedAt = &s
	}
	return &t, nil
}

// generateShareToken - 9 bytes random -> base64 URL without padding -> 12 chars exactly
const shareTokenLength = 12

func generateShareToken() (string, error) {
	raw := make([]byte, 9)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := base64.RawURLEncoding.EncodeToString(raw)
	if len(token) != shareTokenLength {
		return "", fmt.Errorf("unexpected token length: got %d, want %d", len(token), shareTokenLength)
	}
	return token, nil
}

func (r *TripRepository) CountActiveSharesByUser(
	ctx context.Context,
	userID *uuid.UUID,
	sessionID *string,
) (int, error) {
	var count int
	err := r.Pool.QueryRow(ctx, `
        SELECT COUNT(*) FROM trips
        WHERE share_enabled = true
          AND (share_expires_at IS NULL OR share_expires_at > NOW())
          AND ((user_id = $1 AND $1 IS NOT NULL) OR (session_id = $2 AND $2 IS NOT NULL))
    `, userID, sessionID).Scan(&count)
	return count, err
}
