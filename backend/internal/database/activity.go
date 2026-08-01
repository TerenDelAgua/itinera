package database

import (
	"backend/internal/models"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ActivityRepository struct {
	Pool *pgxpool.Pool
}

func NewActivityRepository(pool *pgxpool.Pool) *ActivityRepository {
	return &ActivityRepository{Pool: pool}
}

func (r *ActivityRepository) GetByTripId(ctx context.Context, tripId uuid.UUID) ([]models.Activity, error) {
	query := `
		SELECT id, trip_id, place_id, title, date, time, notes, created_at
		FROM activities
		WHERE trip_id = $1
		ORDER BY date ASC, time ASC NULLS LAST
		`
	rows, err := r.Pool.Query(ctx, query, tripId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities = []models.Activity{}

	for rows.Next() {
		var act models.Activity
		var date time.Time
		var activityTime *time.Time

		err := rows.Scan(
			&act.Id,
			&act.TripId,
			&act.PlaceId,
			&act.Title,
			&date,
			&activityTime,
			&act.Notes,
			&act.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		act.Date = date.Format("2006-01-02")
		if activityTime != nil {
			t := activityTime.Format("15:04")
			act.Time = &t
		}

		activities = append(activities, act)
	}

	return activities, nil
}

func (r *ActivityRepository) GetByPlaceId(ctx context.Context, placeId uuid.UUID) ([]models.Activity, error) {
	query := `
		SELECT id, trip_id, place_id, title, date, time, notes, created_at
		FROM activities
		WHERE place_id = $1
		ORDER BY date ASC, time ASC NULLS LAST
	`

	rows, err := r.Pool.Query(ctx, query, placeId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var activities = []models.Activity{}
	for rows.Next() {
		var act models.Activity
		var date time.Time
		var activityTime *time.Time

		err := rows.Scan(
			&act.Id,
			&act.TripId,
			&act.PlaceId,
			&act.Title,
			&date,
			&activityTime,
			&act.Notes,
			&act.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		act.Date = date.Format("2006-01-02")
		if activityTime != nil {
			t := activityTime.Format("15:04")
			act.Time = &t
		}

		activities = append(activities, act)
	}

	return activities, nil
}

func (r *ActivityRepository) GetById(ctx context.Context, id uuid.UUID) (*models.Activity, error) {
	query := `
		SELECT id, trip_id, place_id, title, date, time, notes, created_at
		FROM activities
		WHERE id = $1
	`
	var act models.Activity
	var date time.Time
	var activityTime *time.Time

	err := r.Pool.QueryRow(ctx, query, id).Scan(
		&act.Id,
		&act.TripId,
		&act.PlaceId,
		&act.Title,
		&date,
		&activityTime,
		&act.Notes,
		&act.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	act.Date = date.Format("2006-01-02")
	if activityTime != nil {
		t := activityTime.Format("15:04")
		act.Time = &t
	}

	return &act, nil
}

func (r *ActivityRepository) CreateActivity(ctx context.Context, act *models.Activity) error {
	// We deliberately avoid RETURNING here. pgx v5's Pool.Exec is intended
	// for non-row-returning commands; using it with RETURNING can leave the
	// returned rows undrained and the INSERT effectively aborted depending
	// on the column types involved (we hit this with uuid + time). The
	// created_at timestamp is set in the handler instead.
	query := `
		INSERT INTO activities (id, trip_id, place_id, title, date, time, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.Pool.Exec(ctx, query,
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

func (r *ActivityRepository) UpdateActivity(ctx context.Context, act *models.Activity) error {
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

	_, err := r.Pool.Exec(ctx, query,
		act.Title,
		act.Date,
		act.Time,
		act.Notes,
		act.PlaceId,
		act.Id,
	)

	return err
}

func (r *ActivityRepository) DeleteActivity(ctx context.Context, id uuid.UUID) error {
	query := " DELETE FROM activities WHERE id = $1"

	_, err := r.Pool.Exec(ctx, query, id)

	return err
}

func (r *ActivityRepository) CloneByTripID(
	ctx context.Context,
	tx pgx.Tx,
	origTripID, newTripID uuid.UUID,
	placeMap map[uuid.UUID]uuid.UUID,
) error {
	rows, err := tx.Query(ctx,
		`SELECT id, place_id, title, date, time, notes
		FROM activities WHERE trip_id = $1`,
		origTripID)

	if err != nil {
		return fmt.Errorf("failed to fetch activities: %w", err)
	}

	type actData struct {
		ID      uuid.UUID
		PlaceID *uuid.UUID
		Title   string
		Date    time.Time
		Time    *time.Time
		Notes   *string
	}
	var srcActs []actData
	for rows.Next() {
		var a actData
		if err := rows.Scan(
			&a.ID,
			&a.PlaceID,
			&a.Title,
			&a.Date,
			&a.Time,
			&a.Notes,
		); err != nil {
			rows.Close()
			return err
		}
		srcActs = append(srcActs, a)
	}
	rows.Close()

	if len(srcActs) == 0 {
		return nil
	}

	var actArgs []any
	var actVals []string
	for i, a := range srcActs {
		newActID := uuid.New()
		var mappedPlaceID *uuid.UUID
		if a.PlaceID != nil {
			if newPID, ok := placeMap[*a.PlaceID]; ok {
				mappedPlaceID = &newPID
			}
		}
		offset := i * 7
		actVals = append(actVals, fmt.Sprintf(
			"($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7,
		))
		actArgs = append(actArgs, newActID, newTripID, mappedPlaceID,
			a.Title, a.Date, a.Time, a.Notes,
		)
	}

	queryInsertActs := fmt.Sprintf(
		`INSERT INTO activities (id, trip_id, place_id, title, date, time, notes)
		VALUES %s`,
		strings.Join(actVals, ","))

	if _, err := tx.Exec(ctx, queryInsertActs, actArgs...); err != nil {
		return fmt.Errorf("failed to insert cloned activities: %w", err)
	}
	return nil
}
