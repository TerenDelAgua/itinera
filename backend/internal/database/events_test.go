package database_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/database"
	"backend/internal/models"
	"backend/internal/services"
)

// TestForkFromDemo_LogsEvent verifies that the orchestrating service
// (TripService.CloneTrip, called via ForkFromDemo) writes the expected
// trip_forked event with the correct metadata. The "fork" itself is no
// longer a repository operation — the service composes it from
// GetBaseForClone, InsertFork, and the per-table CloneByTripID helpers.
func TestForkFromDemo_LogsEvent(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()

	tripsRepo := database.NewTripRepository(pool)
	placesRepo := database.NewPlaceRepository(pool)
	expensesRepo := database.NewExpenseRepository(pool)
	activitiesRepo := database.NewActivityRepository(pool)
	eventsRepo := database.NewEventRepository(pool)

	tripSvc := services.NewTripService(
		pool, tripsRepo, placesRepo, activitiesRepo, expensesRepo, eventsRepo,
	)

	t.Run("ForkFromDemo should log trip_forked event with ghost_mode_write trigger", func(t *testing.T) {
		// 1. Setup: create a demo trip owned by a guest session.
		sessionID := "test-session-fork-log-" + uuid.New().String()
		demoTrip, err := tripsRepo.CreateTrip(ctx, nil, &sessionID, false, models.Trip{
			Name:         "Original Demo Trip",
			BaseCurrency: "JPY",
			IsPublicDemo: true,
			StartDate:    "2026-05-07",
			EndDate:      "2026-05-14",
		})
		require.NoError(t, err)

		defer func() {
			_, _ = pool.Exec(ctx, "DELETE FROM events WHERE trip_id IN ($1, $2)", demoTrip.ID, demoTrip.ID)
			_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE id = $1 OR forked_from = $1", demoTrip.ID)
		}()

		// 2. Fork through the service (orchestrator). ForkFromDemo is the
		//    canonical entry point now; it validates IsPublicDemo + clones.
		var sessionPtr *string
		session := sessionID
		sessionPtr = &session

		forkedTrip, err := tripSvc.ForkFromDemo(ctx, demoTrip.ID, nil, sessionPtr)
		require.NoError(t, err)
		require.NotNil(t, forkedTrip)

		// 3. Verify event exists in the unified events table.
		var count int
		err = pool.QueryRow(ctx,
			"SELECT COUNT(*) FROM events WHERE type = 'trip_forked' AND trip_id = $1 AND session_id = $2",
			forkedTrip.ID, sessionID,
		).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have logged exactly one trip_forked event")

		// 4. Verify metadata contains forked_from + trigger.
		var metadata map[string]interface{}
		err = pool.QueryRow(ctx,
			"SELECT metadata FROM events WHERE type = 'trip_forked' AND trip_id = $1",
			forkedTrip.ID,
		).Scan(&metadata)
		require.NoError(t, err)
		assert.Equal(t, demoTrip.ID.String(), metadata["forked_from"], "Metadata should point to original trip")
		assert.Equal(t, "ghost_mode_write", metadata["trigger"], "Metadata should have the correct trigger")
	})
}
