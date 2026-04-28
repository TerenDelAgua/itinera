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
	parsedDate, err := time.Parse("2006-01-02", input.Date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
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
		CategoryId:       uuid.MustParse(input.CategoryID),
		Notes:            input.Notes,
		Date:             parsedDate,
	}

	// 4. Persistencia
	exp, err := s.ExpensesRepo.CreateExpense(ctx, &tripID, placeUUID, newExpense)
	if err != nil {
		return nil, fmt.Errorf("failed to save expense: %w", err)
	}

	return exp, nil
}
