package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/database"
	"backend/internal/models"
)

func TestTripRepository_Integration(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	expRepo := database.NewExpenseRepository(pool)
	placeRepo := database.NewPlaceRepository(pool)

	ctx := context.Background()
	sessionID := "test-session-" + uuid.New().String()
	
	// Ensure cleanup
	defer func() {
		// En la vida real, podrías querer borrar las filas usando un DELETE directo.
		// trips_test.go no expone un Delete que limpie en cascada a menos que la BD tenga ON DELETE CASCADE.
		// Asumimos que DeleteTrip está implementado y elimina el trip.
		_ = repo.DeleteTrip(ctx, "", nil, &sessionID)
	}()

	t.Run("Create and Get Trip", func(t *testing.T) {
		tripData := models.Trip{
			Name:                   "Test Integration Trip",
			BaseCurrency:           "EUR",
			DefaultExpenseCurrency: "USD",
			StartDate:              "2023-01-01",
			EndDate:                "2023-01-10",
		}

		// Create Trip
		newTrip, err := repo.CreateTrip(ctx, nil, &sessionID, tripData)
		require.NoError(t, err)
		require.NotNil(t, newTrip)

		tripID := newTrip.ID

		// Get Trip
		fetchedTrip, err := repo.GetTrip(ctx, tripID.String(), nil, &sessionID)
		require.NoError(t, err)
		assert.Equal(t, tripID, fetchedTrip.ID)

		// Add Place
		newPlace, err := placeRepo.CreatePlace(ctx, tripID, models.Place{
			Name: "Test Place",
		})
		require.NoError(t, err)
		require.NotNil(t, newPlace)

		// Fetch a valid CategoryId for testing
		var categoryId uuid.UUID
		err = pool.QueryRow(ctx, "SELECT id FROM expense_categories LIMIT 1").Scan(&categoryId)
		require.NoError(t, err)

		// Add Expense 1 (Base currency)
		exp1, err := expRepo.CreateExpense(ctx, &tripID, nil, models.Expense{
			Amount:           50.0,
			OriginalCurrency: "EUR",
			Currency:         "EUR",
			Date:             time.Now(),
			CategoryId:       categoryId,
		})
		require.NoError(t, err)
		require.NotNil(t, exp1)

		// Add Expense 2 (Different currency)
		_, err = expRepo.CreateExpense(ctx, &tripID, nil, models.Expense{
			Amount:           25.0, // This is the converted amount that goes into SUM
			OriginalCurrency: "USD",
			OriginalAmount:   28.0,
			Currency:         "EUR", // Saved as base
			Date:             time.Now(),
			CategoryId:       categoryId,
		})
		require.NoError(t, err)

		// Get Trip again and verify Totals
		updatedTrip, err := repo.GetTrip(ctx, tripID.String(), nil, &sessionID)
		require.NoError(t, err)
		
		// Expected: 50.0 + 25.0 = 75.0
		assert.Equal(t, 75.0, updatedTrip.TotalSpent, "TotalSpent should be accurately aggregated by SQL")
		
		// Expected: 1 Place
		assert.Equal(t, 1, updatedTrip.PlaceCount, "PlaceCount should be exactly 1")

		// Clean up explicitly
		err = repo.DeleteTrip(ctx, tripID.String(), nil, &sessionID)
		assert.NoError(t, err)
	})
}
