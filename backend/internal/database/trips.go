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

func (r *TripRepository) ForkTrip(ctx context.Context, originalTripID string, userID *uuid.UUID, sessionID *string) (*models.Trip, error) {
	tx, err := r.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	origID, err := uuid.Parse(originalTripID)
	if err != nil {
		return nil, errors.New("invalid original trip ID")
	}

	// 1. Fetch original trip (only if it's a public demo)
	var origTrip models.Trip
	var startDate, endDate time.Time
	queryOrig := `SELECT name, description, start_date, end_date, base_currency, default_expense_currency FROM trips WHERE id = $1 AND is_public_demo = true`
	err = tx.QueryRow(ctx, queryOrig, origID).Scan(&origTrip.Name, &origTrip.Description, &startDate, &endDate, &origTrip.BaseCurrency, &origTrip.DefaultExpenseCurrency)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch original demo trip: %w", err)
	}

	// 2. Insert new Trip
	newTripID := uuid.New()
	queryInsertTrip := `
		INSERT INTO trips (id, user_id, session_id, name, description, start_date, end_date, base_currency, default_expense_currency, is_public_demo, forked_from)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, $10)
	`
	_, err = tx.Exec(ctx, queryInsertTrip, newTripID, userID, sessionID, origTrip.Name, origTrip.Description, startDate, endDate, origTrip.BaseCurrency, origTrip.DefaultExpenseCurrency, origID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert forked trip: %w", err)
	}

	// 3. Clone Places
	queryPlaces := `SELECT id, name, notes, lat, lon, start_date, end_date, default_expense_currency FROM places WHERE trip_id = $1`
	rows, err := tx.Query(ctx, queryPlaces, origID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch places: %w", err)
	}
	
	type placeData struct {
		ID uuid.UUID
		Name string
		Notes *string
		Lat *float64
		Lon *float64
		StartDate *time.Time
		EndDate *time.Time
		DefaultExpenseCurrency *string
	}
	var places []placeData
	for rows.Next() {
		var p placeData
		if err := rows.Scan(&p.ID, &p.Name, &p.Notes, &p.Lat, &p.Lon, &p.StartDate, &p.EndDate, &p.DefaultExpenseCurrency); err != nil {
			rows.Close()
			return nil, err
		}
		places = append(places, p)
	}
	rows.Close()

	placeMap := make(map[uuid.UUID]uuid.UUID)
	if len(places) > 0 {
		var placeInsertArgs []any
		var placeValues []string
		
		for i, p := range places {
			newPlaceID := uuid.New()
			placeMap[p.ID] = newPlaceID
			
			offset := i * 9
			placeValues = append(placeValues, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9))
			placeInsertArgs = append(placeInsertArgs, newPlaceID, newTripID, p.Name, p.Notes, p.Lat, p.Lon, p.StartDate, p.EndDate, p.DefaultExpenseCurrency)
		}
		
		queryInsertPlaces := fmt.Sprintf(`INSERT INTO places (id, trip_id, name, notes, lat, lon, start_date, end_date, default_expense_currency) VALUES %s`, strings.Join(placeValues, ","))
		if _, err := tx.Exec(ctx, queryInsertPlaces, placeInsertArgs...); err != nil {
			return nil, fmt.Errorf("failed to insert places: %w", err)
		}
	}

	// 4. Clone Activities
	queryActivities := `SELECT id, place_id, title, date, time, notes FROM activities WHERE trip_id = $1`
	aRows, err := tx.Query(ctx, queryActivities, origID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch activities: %w", err)
	}
	type activityData struct {
		ID uuid.UUID
		PlaceID *uuid.UUID
		Title string
		Date time.Time
		Time *string
		Notes *string
	}
	var activities []activityData
	for aRows.Next() {
		var a activityData
		if err := aRows.Scan(&a.ID, &a.PlaceID, &a.Title, &a.Date, &a.Time, &a.Notes); err != nil {
			aRows.Close()
			return nil, err
		}
		activities = append(activities, a)
	}
	aRows.Close()

	if len(activities) > 0 {
		var actArgs []any
		var actVals []string
		for i, a := range activities {
			newActID := uuid.New()
			var mappedPlaceID *uuid.UUID
			if a.PlaceID != nil {
				if newPID, ok := placeMap[*a.PlaceID]; ok {
					mappedPlaceID = &newPID
				}
			}
			offset := i * 7
			actVals = append(actVals, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)", offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7))
			actArgs = append(actArgs, newActID, newTripID, mappedPlaceID, a.Title, a.Date, a.Time, a.Notes)
		}
		queryInsertActs := fmt.Sprintf(`INSERT INTO activities (id, trip_id, place_id, title, date, time, notes) VALUES %s`, strings.Join(actVals, ","))
		if _, err := tx.Exec(ctx, queryInsertActs, actArgs...); err != nil {
			return nil, fmt.Errorf("failed to insert activities: %w", err)
		}
	}

	// 5. Clone Expenses
	queryExpenses := `SELECT amount, currency, converted_amount, date, notes, category_id, place_id, original_amount, original_currency, exchange_rate, conversion_date FROM expenses WHERE trip_id = $1`
	eRows, err := tx.Query(ctx, queryExpenses, origID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch expenses: %w", err)
	}
	type expData struct {
		Amount float64
		Currency string
		ConvertedAmt *float64
		Date time.Time
		Notes *string
		CatID *uuid.UUID
		PlaceID *uuid.UUID
		OrigAmt float64
		OrigCur string
		ExchRate float64
		ConvDate time.Time
	}
	var expenses []expData
	for eRows.Next() {
		var e expData
		if err := eRows.Scan(&e.Amount, &e.Currency, &e.ConvertedAmt, &e.Date, &e.Notes, &e.CatID, &e.PlaceID, &e.OrigAmt, &e.OrigCur, &e.ExchRate, &e.ConvDate); err != nil {
			eRows.Close()
			return nil, err
		}
		expenses = append(expenses, e)
	}
	eRows.Close()

	if len(expenses) > 0 {
		var expArgs []any
		var expVals []string
		for i, e := range expenses {
			newExpID := uuid.New()
			var mappedPlaceID *uuid.UUID
			if e.PlaceID != nil {
				if newPID, ok := placeMap[*e.PlaceID]; ok {
					mappedPlaceID = &newPID
				}
			}
			offset := i * 14
			expVals = append(expVals, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d, $%d)", 
				offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7, offset+8, offset+9, offset+10, offset+11, offset+12, offset+13, offset+14))
			expArgs = append(expArgs, newExpID, newTripID, userID, e.Amount, e.Currency, e.ConvertedAmt, e.Date, e.Notes, e.CatID, mappedPlaceID, e.OrigAmt, e.OrigCur, e.ExchRate, e.ConvDate)
		}
		queryInsertExp := fmt.Sprintf(`INSERT INTO expenses (id, trip_id, user_id, amount, currency, converted_amount, date, notes, category_id, place_id, original_amount, original_currency, exchange_rate, conversion_date) VALUES %s`, strings.Join(expVals, ","))
		if _, err := tx.Exec(ctx, queryInsertExp, expArgs...); err != nil {
			return nil, fmt.Errorf("failed to insert expenses: %w", err)
		}
	}

	// 6. Log event (using unified events table)
	logQuery := `INSERT INTO events (type, session_id, user_id, trip_id, metadata, created_at) VALUES ($1, $2, $3, $4, $5, NOW())`
	metadata := fmt.Sprintf(`{"forked_from": "%s", "trigger": "ghost_mode_write"}`, originalTripID)
	if _, err := tx.Exec(ctx, logQuery, "demo_deep_forked", sessionID, userID, newTripID, metadata); err != nil {
		return nil, fmt.Errorf("failed to log fork event: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit fork transaction: %w", err)
	}

	// Fetch full new trip to return
	return r.GetTrip(ctx, newTripID.String(), userID, sessionID)
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
