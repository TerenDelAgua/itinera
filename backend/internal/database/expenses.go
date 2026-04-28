package database

import (
	"backend/internal/models"
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ExpenseRepository struct {
	Pool *pgxpool.Pool
}

func NewExpenseRepository(pool *pgxpool.Pool) *ExpenseRepository {
	return &ExpenseRepository{Pool: pool}
}

func (r *ExpenseRepository) GetCategories(ctx context.Context) ([]models.ExpenseCategory, error) {
	rows, err := r.Pool.Query(
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

func (r *ExpenseRepository) CreateExpense(ctx context.Context, tripID, placeID *uuid.UUID, exp models.Expense) (*models.Expense, error) {
	query := `INSERT INTO expenses (trip_id, place_id, amount, original_amount, original_currency, exchange_rate, conversion_date, currency, category_id, notes, date) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
		RETURNING id, trip_id, place_id, amount, original_amount, original_currency, exchange_rate, conversion_date, currency, category_id, notes, date, created_at`

	var res models.Expense
	err := r.Pool.QueryRow(ctx, query, 
		tripID, placeID, 
		exp.Amount, exp.OriginalAmount, exp.OriginalCurrency, exp.ExchangeRate, exp.ConversionDate,
		exp.Currency, exp.CategoryId, exp.Notes, exp.Date,
	).
		Scan(
			&res.Id, &res.TripId, &res.PlaceId, 
			&res.Amount, &res.OriginalAmount, &res.OriginalCurrency, &res.ExchangeRate, &res.ConversionDate,
			&res.Currency, &res.CategoryId, &res.Notes, &res.Date, &res.CreatedAt,
		)
	return &res, err
}
func (r *ExpenseRepository) GetExpensesByTrip(ctx context.Context, tripId uuid.UUID) ([]models.Expense, error) {
	rows, err := r.Pool.Query(
		ctx,
		`SELECT
			id,
			trip_id,
			place_id,
			category_id,
			amount,
			original_amount,
			original_currency,
			exchange_rate,
			conversion_date,
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
			&e.Id, &e.TripId, &e.PlaceId,
			&e.CategoryId, &e.Amount,
			&e.OriginalAmount, &e.OriginalCurrency, &e.ExchangeRate, &e.ConversionDate,
			&e.Currency, &e.Notes,
			&e.Date, &e.CreatedAt); err != nil {
			return nil, err
		}
		expenses = append(expenses, e)
	}
	return expenses, nil

}

func (r *ExpenseRepository) GetExpensesSummary(ctx context.Context, tripId uuid.UUID) (*models.TripExpenseSummary, error) {
	summary := &models.TripExpenseSummary{
		ByCategory: []models.CategorySummary{},
		ByPlace:    []models.PlaceSummary{},
	}

	// 1. Totals
	err := r.Pool.QueryRow(ctx, `
		SELECT 
			COALESCE(SUM(amount), 0) as grand_total,
			COALESCE(SUM(amount) FILTER (WHERE place_id IS NULL), 0) as global_total,
			COALESCE(SUM(amount) FILTER (WHERE place_id IS NOT NULL), 0) as places_total
		FROM expenses WHERE trip_id = $1`, tripId).
		Scan(&summary.GrandTotal, &summary.GlobalTotal, &summary.PlacesTotal)
	if err != nil {
		return nil, err
	}

	// 2. By Category
	rowsCat, err := r.Pool.Query(ctx, `
		SELECT category_id, SUM(amount) FROM expenses 
		WHERE trip_id = $1 GROUP BY category_id`, tripId)
	if err == nil {
		defer rowsCat.Close()
		for rowsCat.Next() {
			var s models.CategorySummary
			if err := rowsCat.Scan(&s.CategoryId, &s.Total); err == nil {
				summary.ByCategory = append(summary.ByCategory, s)
			}
		}
	}

	// 3. By Place
	rowsPlace, err := r.Pool.Query(ctx, `
		SELECT e.place_id, p.name, SUM(e.amount) 
		FROM expenses e
		JOIN places p ON e.place_id = p.id
		WHERE e.trip_id = $1
		GROUP BY e.place_id, p.name`, tripId)
	if err == nil {
		defer rowsPlace.Close()
		for rowsPlace.Next() {
			var s models.PlaceSummary
			if err := rowsPlace.Scan(&s.PlaceId, &s.PlaceName, &s.Total); err == nil {
				summary.ByPlace = append(summary.ByPlace, s)
			}
		}
	}

	return summary, nil
}

func (r *ExpenseRepository) ListGlobalExpenses(ctx context.Context, tripID uuid.UUID) ([]models.Expense, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, trip_id, place_id, amount, original_amount, original_currency, exchange_rate, conversion_date, currency, category_id, notes, date, created_at 
		FROM expenses WHERE trip_id = $1 AND place_id IS NULL ORDER BY date DESC`, tripID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exps []models.Expense
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.Id, &e.TripId, &e.PlaceId, &e.Amount, &e.OriginalAmount, &e.OriginalCurrency, &e.ExchangeRate, &e.ConversionDate, &e.Currency, &e.CategoryId, &e.Notes, &e.Date, &e.CreatedAt); err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	return exps, nil
}

func (r *ExpenseRepository) ListPlaceExpenses(ctx context.Context, placeID uuid.UUID) ([]models.Expense, error) {
	rows, err := r.Pool.Query(ctx, `SELECT id, trip_id, place_id, amount, original_amount, original_currency, exchange_rate, conversion_date, currency, category_id, notes, date, created_at 
		FROM expenses WHERE place_id = $1 ORDER BY date DESC`, placeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exps []models.Expense = []models.Expense{}
	for rows.Next() {
		var e models.Expense
		if err := rows.Scan(&e.Id, &e.TripId, &e.PlaceId, &e.Amount, &e.OriginalAmount, &e.OriginalCurrency, &e.ExchangeRate, &e.ConversionDate, &e.Currency, &e.CategoryId, &e.Notes, &e.Date, &e.CreatedAt); err != nil {
			return nil, err
		}
		exps = append(exps, e)
	}
	return exps, nil
}
func (r *ExpenseRepository) UpdateExpense(ctx context.Context, id uuid.UUID, exp models.Expense) (*models.Expense, error) {
	query := `UPDATE expenses SET amount = $1, original_amount = $2, original_currency = $3, exchange_rate = $4, conversion_date = $5, date = $6, notes = $7, category_id = $8 
		WHERE id = $9 
		RETURNING id, trip_id, place_id, amount, original_amount, original_currency, exchange_rate, conversion_date, currency, category_id, notes, date, created_at`
	var res models.Expense
	err := r.Pool.QueryRow(ctx, query, 
		exp.Amount, exp.OriginalAmount, exp.OriginalCurrency, exp.ExchangeRate, exp.ConversionDate, 
		exp.Date, exp.Notes, exp.CategoryId, id).
		Scan(&res.Id, &res.TripId, &res.PlaceId, &res.Amount, &res.OriginalAmount, &res.OriginalCurrency, &res.ExchangeRate, &res.ConversionDate, &res.Currency, &res.CategoryId, &res.Notes, &res.Date, &res.CreatedAt)
	return &res, err
}
func (r *ExpenseRepository) GetPlaceExpensesSummary(ctx context.Context, placeID uuid.UUID) ([]models.CategorySummary, error) {
	rows, err := r.Pool.Query(ctx, `
		SELECT category_id, SUM(amount) FROM expenses 
		WHERE place_id = $1 GROUP BY category_id`, placeID)
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
