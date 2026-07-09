package services

import (
	"backend/internal/database"
	"backend/internal/models"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type ExpenseService struct {
	TripRepo     *database.TripRepository
	ExpensesRepo *database.ExpenseRepository
	ExchangeRate *ExchangeRateService
}

func NewExpenseService(
	tripRepo *database.TripRepository,
	expensesRepo *database.ExpenseRepository,
	exchangeRate *ExchangeRateService,
) *ExpenseService {
	return &ExpenseService{
		TripRepo:     tripRepo,
		ExpensesRepo: expensesRepo,
		ExchangeRate: exchangeRate,
	}
}

type CreateExpenseInput struct {
	Amount           float64 `json:"amount"`
	OriginalCurrency string  `json:"original_currency"`
	CategoryID       string  `json:"category_id"`
	Date             string  `json:"date"`
	Notes            string  `json:"notes"`
	PlaceID          *string `json:"place_id"`
}

func (s *ExpenseService) CalculateAndCreateExpense(ctx context.Context, tripID uuid.UUID, input CreateExpenseInput) (*models.Expense, error) {
	// 1. Obtener configuración del viaje (moneda base)
	trip, err := s.TripRepo.GetTripById(ctx, tripID)
	if err != nil {
		return nil, fmt.Errorf("trip not found: %w", err)
	}

	targetCurrency := trip.BaseCurrency
	if targetCurrency == "" {
		targetCurrency = "EUR"
	}

	// 2. Lógica de Conversión Multi-divisa
	var finalAmount float64
	var rate float64 = 1.0

	if input.OriginalCurrency == targetCurrency {
		finalAmount = input.Amount
	} else {
		rate, err = s.ExchangeRate.GetRate(ctx, input.OriginalCurrency, targetCurrency)
		if err != nil {
			rate = 1.0
			fmt.Printf("Warning: failed to get exchange rate, using 1.0 fallback: %v\n", err)
		}
		finalAmount = input.Amount * rate
	}

	// 3. Formateo y Mapeo de datos
	// Intentar parsear como ISO8601/RFC3339 primero, luego fallback a YYYY-MM-DD
	parsedDate, err := time.Parse(time.RFC3339, input.Date)
	if err != nil {
		parsedDate, err = time.Parse("2006-01-02", input.Date)
		if err != nil {
			return nil, fmt.Errorf("invalid date format (expected RFC3339 or YYYY-MM-DD): %w", err)
		}
	}

	var placeUUID *uuid.UUID
	if input.PlaceID != nil && *input.PlaceID != "" {
		pUUID, err := uuid.Parse(*input.PlaceID)
		if err == nil {
			placeUUID = &pUUID
		}
	}

	newExpense := models.Expense{
		Id:               uuid.New(),
		TripId:           tripID,
		PlaceId:          placeUUID,
		Amount:           finalAmount,
		OriginalAmount:   input.Amount,
		OriginalCurrency: input.OriginalCurrency,
		ExchangeRate:     rate,
		ConversionDate:   time.Now(),
		Currency:         targetCurrency,
		Notes:            input.Notes,
		Date:             parsedDate,
	}

	catID, err := uuid.Parse(input.CategoryID)
	if err != nil {
		return nil, fmt.Errorf("invalid category id: %w", err)
	}
	newExpense.CategoryId = catID

	// 4. Persistencia
	exp, err := s.ExpensesRepo.CreateExpense(ctx, &tripID, placeUUID, newExpense)
	if err != nil {
		return nil, fmt.Errorf("failed to save expense: %w", err)
	}

	return exp, nil
}

func (s *ExpenseService) RecalculateTripExpenses(ctx context.Context, tripID uuid.UUID) error {
	trip, err := s.TripRepo.GetTripById(ctx, tripID)
	if err != nil {
		return err
	}

	expenses, err := s.ExpensesRepo.GetExpensesByTrip(ctx, tripID)
	if err != nil {
		return err
	}

	targetCurrency := trip.BaseCurrency
	if targetCurrency == "" {
		targetCurrency = "EUR"
	}

	for _, exp := range expenses {
		var finalAmount float64
		var rate float64 = 1.0

		if exp.OriginalCurrency == targetCurrency {
			finalAmount = exp.OriginalAmount
		} else {
			rate, err = s.ExchangeRate.GetRate(ctx, exp.OriginalCurrency, targetCurrency)
			if err != nil {
				rate = 1.0
			}
			finalAmount = exp.OriginalAmount * rate
		}

		// Update individual expense
		exp.Amount = finalAmount
		exp.ExchangeRate = rate
		exp.Currency = targetCurrency
		exp.ConversionDate = time.Now()

		_, err = s.ExpensesRepo.UpdateExpense(ctx, exp.Id, exp)
		if err != nil {
			fmt.Printf("Error updating expense %s during recalculation: %v\n", exp.Id, err)
		}
	}

	return nil
}
