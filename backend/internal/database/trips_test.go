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
	"backend/internal/services"
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
		newTrip, err := repo.CreateTrip(ctx, nil, &sessionID, false, tripData)
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

		guestTrip, err := repo.CreateTrip(ctx, nil, &guestSessionID, false, models.Trip{Name: "Guest Trip", BaseCurrency: "EUR", StartDate: "2024-01-01", EndDate: "2024-01-10"})
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

		userTrip, err := repo.CreateTrip(ctx, &userID, nil, false, models.Trip{Name: "User Trip", BaseCurrency: "EUR", StartDate: "2024-01-01", EndDate: "2024-01-10"})
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

	t.Run("TestForkTrip_Atomic", func(t *testing.T) {
		guestSessionID := "fork-test-session-" + uuid.New().String()

		// 1. Create a demo trip directly
		demoID := uuid.New()
		_, err := pool.Exec(ctx, `INSERT INTO trips (id, name, is_public_demo, session_id, created_at, base_currency, default_expense_currency, start_date, end_date)
			VALUES ($1, $2, $3, $4, $5, $6, $7, CURRENT_DATE, CURRENT_DATE)`,
			demoID, "Demo for Forking", true, guestSessionID, time.Now(), "EUR", "EUR")
		require.NoError(t, err)
		defer pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", demoID)

		// 2. Add Place to Demo
		placeID := uuid.New()
		_, err = pool.Exec(ctx, `INSERT INTO places (id, trip_id, name) VALUES ($1, $2, $3)`, placeID, demoID, "Demo Place")
		require.NoError(t, err)

		// 3. Add Activity to Place
		activityID := uuid.New()
		_, err = pool.Exec(ctx, `INSERT INTO activities (id, trip_id, place_id, title, date) VALUES ($1, $2, $3, $4, CURRENT_DATE)`, activityID, demoID, placeID, "Demo Activity")
		require.NoError(t, err)

		// 4. Add Expense to Trip
		expenseID := uuid.New()
		_, err = pool.Exec(ctx, `INSERT INTO expenses (id, trip_id, amount, currency, date, original_amount, original_currency, exchange_rate) VALUES ($1, $2, $3, $4, CURRENT_DATE, $5, $6, $7)`, expenseID, demoID, 100.0, "EUR", 100.0, "EUR", 1.0)
		require.NoError(t, err)

		// 5. Perform Fork via the orchestrating service.
		//    Repositories no longer expose ForkTrip directly: the fork is a
		//    multi-table operation coordinated by TripService.CloneTrip.
		activitiesRepo := database.NewActivityRepository(pool)
		eventsRepo := database.NewEventRepository(pool)
		tripSvc := services.NewTripService(
			pool, repo, placeRepo, activitiesRepo, expRepo, eventsRepo,
		)
		forkedTrip, err := tripSvc.ForkFromDemo(ctx, demoID, nil, &guestSessionID)
		require.NoError(t, err)
		require.NotNil(t, forkedTrip)
		defer repo.DeleteTrip(ctx, forkedTrip.ID.String(), nil, &guestSessionID)

		// 6. Assertions
		assert.NotEqual(t, demoID, forkedTrip.ID, "Forked trip should have a new ID")
		assert.Equal(t, "Demo for Forking", forkedTrip.Name, "Name should be copied")
		assert.False(t, forkedTrip.IsPublicDemo, "Forked trip should NOT be a public demo")
		assert.NotNil(t, forkedTrip.ForkedFrom, "ForkedFrom should be set")
		assert.Equal(t, demoID, *forkedTrip.ForkedFrom, "ForkedFrom should point to original demo")

		// 7. Verify Places cloned
		var placeCount int
		err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM places WHERE trip_id = $1`, forkedTrip.ID).Scan(&placeCount)
		require.NoError(t, err)
		assert.Equal(t, 1, placeCount, "Should have 1 cloned place")

		// 8. Verify Activities cloned and mapped correctly
		var actCount int
		var actPlaceID *uuid.UUID
		err = pool.QueryRow(ctx, `SELECT COUNT(*), (SELECT place_id FROM activities WHERE trip_id = $1 LIMIT 1) FROM activities WHERE trip_id = $1`, forkedTrip.ID).Scan(&actCount, &actPlaceID)
		require.NoError(t, err)
		assert.Equal(t, 1, actCount, "Should have 1 cloned activity")
		assert.NotNil(t, actPlaceID, "Cloned activity should have a mapped place_id")
		assert.NotEqual(t, placeID, *actPlaceID, "Cloned activity place_id should NOT be the original place_id")

		// 9. Verify Expenses cloned
		var expCount int
		err = pool.QueryRow(ctx, `SELECT COUNT(*) FROM expenses WHERE trip_id = $1`, forkedTrip.ID).Scan(&expCount)
		require.NoError(t, err)
		assert.Equal(t, 1, expCount, "Should have 1 cloned expense")
	})
}
