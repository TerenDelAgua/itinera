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

	t.Run("TestListTrips_Guest", func(t *testing.T) {
		guestSessionID := "guest-session-" + uuid.New().String()
		
		japanDemoID := uuid.New()
		_, err := pool.Exec(ctx, `INSERT INTO trips (id, name, is_public_demo, session_id, created_at, base_currency, default_expense_currency, start_date, end_date) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_DATE, CURRENT_DATE)`,
			japanDemoID, "Viaje a Japón Clásico", true, "DEMO", time.Now().Add(-1*time.Hour), "EUR", "EUR")
		require.NoError(t, err)
		defer pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", japanDemoID)

		otherDemoID := uuid.New()
		_, err = pool.Exec(ctx, `INSERT INTO trips (id, name, is_public_demo, session_id, created_at, base_currency, default_expense_currency, start_date, end_date) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_DATE, CURRENT_DATE)`,
			otherDemoID, "Roma Express", true, "DEMO", time.Now().Add(-2*time.Hour), "EUR", "EUR")
		require.NoError(t, err)
		defer pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", otherDemoID)

		guestTrip, err := repo.CreateTrip(ctx, nil, &guestSessionID, models.Trip{Name: "Guest Trip", BaseCurrency: "EUR", StartDate: "2024-01-01", EndDate: "2024-01-10"})
		require.NoError(t, err)
		defer repo.DeleteTrip(ctx, guestTrip.ID.String(), nil, &guestSessionID)

		trips, err := repo.ListTrips(ctx, nil, &guestSessionID)
		require.NoError(t, err)

		var guestTripIdx, japanIdx, otherIdx = -1, -1, -1
		for i, tr := range trips {
			if tr.ID == guestTrip.ID { guestTripIdx = i }
			if tr.ID == japanDemoID { japanIdx = i }
			if tr.ID == otherDemoID { otherIdx = i }
		}
		
		require.True(t, guestTripIdx != -1, "Guest trip not found")
		require.True(t, japanIdx != -1, "Japan demo not found")
		require.True(t, otherIdx != -1, "Other demo not found")
		
		assert.Less(t, guestTripIdx, japanIdx, "Guest trip should be before Japan demo")
		assert.Less(t, guestTripIdx, otherIdx, "Guest trip should be before Other demo")
		assert.Less(t, japanIdx, otherIdx, "Japan demo should be before Other demo")
	})

	t.Run("TestListTrips_User", func(t *testing.T) {
		userID := uuid.New()
		
		_, err := pool.Exec(ctx, `INSERT INTO users (id, email, password_hash) VALUES ($1, $2, 'hash')`,
			userID, "test_user_"+userID.String()[:8]+"@example.com")
		require.NoError(t, err)
		defer pool.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)

		japanDemoID := uuid.New()
		_, err = pool.Exec(ctx, `INSERT INTO trips (id, name, is_public_demo, session_id, created_at, base_currency, default_expense_currency, start_date, end_date) 
			VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_DATE, CURRENT_DATE)`,
			japanDemoID, "Viaje a Japón Clásico 2", true, "DEMO", time.Now().Add(-1*time.Hour), "EUR", "EUR")
		require.NoError(t, err)
		defer pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", japanDemoID)

		userTrip, err := repo.CreateTrip(ctx, &userID, nil, models.Trip{Name: "User Trip", BaseCurrency: "EUR", StartDate: "2024-01-01", EndDate: "2024-01-10"})
		require.NoError(t, err)
		defer repo.DeleteTrip(ctx, userTrip.ID.String(), &userID, nil)

		trips, err := repo.ListTrips(ctx, &userID, nil)
		require.NoError(t, err)

		var userTripIdx, japanIdx = -1, -1
		for i, tr := range trips {
			if tr.ID == userTrip.ID { userTripIdx = i }
			if tr.ID == japanDemoID { japanIdx = i }
		}
		
		require.True(t, userTripIdx != -1, "User trip not found")
		require.True(t, japanIdx != -1, "Japan demo not found")
		
		assert.Less(t, userTripIdx, japanIdx, "User trip should be before Japan demo")
	})
}
