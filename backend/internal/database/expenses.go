package database

import (
	"backend/internal/models"
	"context"

	"github.com/google/uuid"
)

func (db *DB) GetCategories(ctx context.Context) ([]models.ExpenseCategory, error) {
	rows, err := db.Pool.Query(
		ctx,
		`SELECT
			id,
			slug,
			color_hex
		FROM expense_categories
		ORDER BY slug`,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.ExpenseCategory = []models.ExpenseCategory{}
	for rows.Next() {
		var c models.ExpenseCategory
		if err := rows.Scan(&c.Id, &c.Slug, &c.ColorHex); err != nil {
			return nil, err
		}
		categories = append(categories, c)
	}

	return categories, nil
}

func (db *DB) CreateExpense(ctx context.Context, tripID, placeID *uuid.UUID, exp models.Expense) (*models.Expense, error) {
	query := `INSERT INTO expenses (trip_id, place_id, amount, currency, category_id, notes, date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, trip_id, place_id, amount, currency, category_id, notes, date, created_at`

	var res models.Expense
	err := db.Pool.QueryRow(ctx, query, tripID, placeID, exp.Amount, exp.Currency, exp.CategoryId, exp.Notes, exp.Date).
		Scan(&res.Id, &res.TripId, &res.PlaceId, &res.Amount, &res.Currency, &res.CategoryId, &res.Notes, &res.Date, &res.CreatedAt)
	return &res, err
}
func (db *DB) GetExpensesByTrip(ctx context.Context, tripId uuid.UUID) ([]models.Expense, error) {
	rows, err := db.Pool.Query(
		ctx,
		`SELECT
			id,
			trip_id,
			category_id,
			amount,
			currency,
			notes,
			date,
			created_at
		FROM expenses
		WHERE trip_id = $1
		ORDER BY date DESC`,
		tripId,
	)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var expenses []models.Expense = []models.Expense{}
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(
			&e.Id, &e.TripId,
			&e.CategoryId, &e.Amount,
			&e.Currency, &e.Notes,
			&e.Date, &e.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil

}

func (db *DB) GetExpensesSummary(ctx context.Context, tripId uuid.UUID) ([]models.CategorySummary, error) {
	rows, err := db.Pool.Query(
		ctx,
		`SELECT
			category_id,
			SUM(amount) as total
		FROM expenses
		WHERE trip_id=$1
		GROUP BY category_id`,
		tripId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var summaries []models.CategorySummary = []models.CategorySummary{}
	for rows.Next() {
		var s models.CategorySummary
		if err := rows.Scan(&s.CategoryId, &s.Total); err != nil {
			return nil, err
		}
		summaries = append(summaries, s)
	}
	return summaries, nil
}

func (db *DB) ListGlobalExpenses(ctx context.Context, tripID uuid.UUID) ([]models.Expense, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, trip_id, place_id, amount, currency, category_id, notes, date, created_at 
		FROM expenses WHERE trip_id = $1 AND place_id IS NULL ORDER BY date DESC`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exps []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.Id, &e.TripId, &e.PlaceId, &e.Amount, &e.Currency, &e.CategoryId, &e.Notes, &e.Date, &e.CreatedAt); err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	return exps, nil
}

func (db *DB) ListPlaceExpenses(ctx context.Context, placeID uuid.UUID) ([]models.Expense, error) {
	rows, err := db.Pool.Query(ctx, `SELECT id, trip_id, place_id, amount, currency, category_id, notes, date, created_at 
		FROM expenses WHERE place_id = $1 ORDER BY date DESC`, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exps []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.Id, &e.TripId, &e.PlaceId, &e.Amount, &e.Currency, &e.CategoryId, &e.Notes, &e.Date, &e.CreatedAt); err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	return exps, nil
}
func (db *DB) UpdateExpense(ctx context.Context, id uuid.UUID, exp models.Expense) (*models.Expense, error) {
	query := `UPDATE expenses SET amount = $1, date = $2, notes = $3, category_id = $4 
		WHERE id = $5 
		RETURNING id, trip_id, place_id, amount, currency, category_id, notes, date, created_at`
	var res models.Expense
	err := db.Pool.QueryRow(ctx, query, exp.Amount, exp.Date, exp.Notes, exp.CategoryId, id).
		Scan(&res.Id, &res.TripId, &res.PlaceId, &res.Amount, &res.Currency, &res.CategoryId, &res.Notes, &res.Date, &res.CreatedAt)
	return &res, err
}
