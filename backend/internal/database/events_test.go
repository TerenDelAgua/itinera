package database_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/database"
	"backend/internal/models"
)

func TestEventLogging_Integration(t *testing.T) {
	pool := getTestPool(t)
	repo := database.NewTripRepository(pool)
	ctx := context.Background()

	t.Run("ForkTrip should log demo_deep_forked event", func(t *testing.T) {
		// 1. Setup: Create a demo trip
		sessionID := "test-session-fork-log-" + uuid.New().String()
		demoTrip, err := repo.CreateTrip(ctx, nil, nil, models.Trip{
			Name:         "Original Demo Trip",
			BaseCurrency: "JPY",
			IsPublicDemo: true,
			StartDate:    "2026-05-07",
			EndDate:      "2026-05-14",
		})
		require.NoError(t, err)

		// Clean up the created trips and events at the end of the test
		defer func() {
			if demoTrip != nil {
				_, _ = pool.Exec(ctx, "DELETE FROM events WHERE trip_id = $1", demoTrip.ID)
				_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", demoTrip.ID)
			}
		}()
		
		// 2. Fork the trip
		forkedTrip, err := repo.ForkTrip(ctx, demoTrip.ID.String(), nil, &sessionID)
		require.NoError(t, err)
		require.NotNil(t, forkedTrip)

		defer func() {
			if forkedTrip != nil {
				_, _ = pool.Exec(ctx, "DELETE FROM events WHERE trip_id = $1", forkedTrip.ID)
				_, _ = pool.Exec(ctx, "DELETE FROM trips WHERE id = $1", forkedTrip.ID)
			}
		}()
 
		// 3. Verify event exists in the unified 'events' table
		var count int
		err = pool.QueryRow(ctx, 
			"SELECT COUNT(*) FROM events WHERE type = 'demo_deep_forked' AND trip_id = $1 AND session_id = $2",
			forkedTrip.ID, sessionID,
		).Scan(&count)
		
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Should have logged exactly one demo_deep_forked event")

		// 4. Verify metadata contains forked_from
		var metadata map[string]interface{}
		err = pool.QueryRow(ctx,
			"SELECT metadata FROM events WHERE type = 'demo_deep_forked' AND trip_id = $1",
			forkedTrip.ID,
		).Scan(&metadata)
		
		require.NoError(t, err)
		assert.Equal(t, demoTrip.ID.String(), metadata["forked_from"], "Metadata should point to original trip")
		assert.Equal(t, "ghost_mode_write", metadata["trigger"], "Metadata should have the correct trigger")
	})
}
