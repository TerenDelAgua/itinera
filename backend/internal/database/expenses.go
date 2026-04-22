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

func (db *DB) CreateExpense(ctx context.Context, tripId uuid.UUID, exp models.Expense) (*models.Expense, error) {
	query := `INSERT INTO expenses(
		trip_id,
		category_id,
		amount,
		currency,
		notes,
		date
	) VALUES (
		$1,
		$2,
		$3,
		$4,
		$5,
		$6
	)
	 RETURNING 
	 	id,
	 	trip_id,
	 	category_id,
	 	amount,
	 	currency,
	 	notes,
	 	date;
	`

	var res models.Expense
	err := db.Pool.QueryRow(
		ctx, query, tripId, exp.CategoryId,
		exp.Amount, exp.Currency,
		exp.Notes, exp.Date).
		Scan(&res.Id, &res.TripId,
			&res.CategoryId, &res.Amount,
			&res.Currency, &res.Notes,
			&res.Date)

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
