package database_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/http/handlers"
	"backend/internal/http/middleware"
	"backend/internal/models"
	"backend/internal/services"
	"backend/internal/services/email"
)

// TestShareForkFlow_EnableShareAfterMiddlewareFork is the regression test
// for the bug: a guest session edits a public demo (which the middleware
// silently forks) and then enables share. Before the fix, EnableShare used
// chi.URLParam("id") — the demo id from the URL — and the owner check
// 403'd because the demo belongs to a different session. After the fix,
// the handler reads WorkingTripIDKey from the middleware context (the
// fork's id, owned by the guest session) and EnableShare returns 200.
func TestShareForkFlow_EnableShareAfterMiddlewareFork(t *testing.T) {
	pool := getTestPool(t)
	pgPool := pool

	tripsRepo := database.NewTripRepository(pgPool)
	placesRepo := database.NewPlaceRepository(pgPool)
	expensesRepo := database.NewExpenseRepository(pgPool)
	authRepo := database.NewAuthRepository(pgPool)
	activityRepo := database.NewActivityRepository(pgPool)
	eventsRepo := database.NewEventRepository(pgPool)
	rateLimitRepo := database.NewRateLimitRepository(pgPool)
	analyticsRepo := database.NewAnalyticsRepository(pgPool)
	sessionRepo := database.NewSessionRepository(pgPool)
	emailSender := email.NoopSender{}

	exchangeRateSvc := services.NewExchangeRateService(pgPool)
	expenseSvc := services.NewExpenseService(tripsRepo, expensesRepo, exchangeRateSvc)
	tripSvc := services.NewTripService(pgPool, tripsRepo, placesRepo, activityRepo, expensesRepo, eventsRepo)

	cfg := config.Load()
	h := handlers.NewHandlers(
		tripsRepo, placesRepo, expensesRepo, authRepo,
		activityRepo, eventsRepo, rateLimitRepo, analyticsRepo, sessionRepo, expenseSvc, tripSvc, emailSender, cfg,
	)

	ctx := context.Background()

	// 1. Create a public demo owned by a different session.
	originalOwner := "demo-owner-" + uuid.New().String()
	demo, err := h.TripsRepo.CreateTrip(ctx, nil, &originalOwner, false, models.Trip{
		Name:                   "Demo for fork",
		BaseCurrency:           "JPY",
		DefaultExpenseCurrency: "JPY",
		IsPublicDemo:           true,
		StartDate:              "2026-08-01",
		EndDate:                "2026-08-10",
	})
	require.NoError(t, err)

	defer func() {
		_, _ = pgPool.Exec(ctx, "DELETE FROM trips WHERE id = $1 OR forked_from = $1", demo.ID)
	}()

	// 2. Build a router with the same middleware that production uses.
	sessionID := "share-fork-test-" + uuid.New().String()
	withSession := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			ctx = context.WithValue(ctx, middleware.ContextKeySessionId{}, sessionID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}

	r := chi.NewRouter()
	r.Use(withSession)
	r.Route("/api/v1", func(r chi.Router) {
		r.With(middleware.ResolveTripContext(h.TripSvc)).Group(func(r chi.Router) {
			r.Post("/trips/{id}/share", h.EnableShare)
		})
	})

	// 3. POST /trips/{demoId}/share from the guest session.
	//    Middleware forks the demo (since isMutating && isDemo) and sets
	//    WorkingTripIDKey to the fork's id.
	req := httptest.NewRequest(http.MethodPost, "/api/v1/trips/"+demo.ID.String()+"/share", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code,
		"EnableShare after middleware fork should succeed but got %d: %s",
		rec.Code, rec.Body.String())

	var respBody struct {
		Token     string `json:"token"`
		URL       string `json:"url"`
		ExpiresAt string `json:"expires_at"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &respBody))
	assert.NotEmpty(t, respBody.Token, "token must be returned")
	assert.Len(t, respBody.Token, 12, "token must be 12 chars (base64url of 9 bytes)")
	assert.Contains(t, respBody.URL, respBody.Token, "url should embed the token")

	// 4. Verify the share actually belongs to the FORK, not the demo.
	//    Read the demo's state — share must still be disabled.
	demoState, err := h.TripsRepo.GetTrip(ctx, demo.ID.String(), nil, &originalOwner)
	require.NoError(t, err)
	assert.False(t, demoState.ShareEnabled, "demo itself must NOT be shared — only the fork")
}

// TestShareForkFlow_PublicEndpoint_JSONShape is a regression test for the
// `each_key_duplicate` error on the frontend. The bug was that
// publicActivity.ID in handlers/share.go lacked a `json:"id"` tag, so Go
// serialized it as "ID" instead of "id". The frontend uses (activity.id)
// as the each-block key, so all rows had key=undefined → duplicate keys.
//
// This test fetches /share/{token} and asserts that:
//   - publicPlaces[i].id is present and is a non-empty string
//   - publicActivities[i].id is present and is a non-empty string
//
// If either is missing, the each_key crash reappears on the frontend.
func TestShareForkFlow_PublicEndpoint_JSONShape(t *testing.T) {
	pool := getTestPool(t)
	pgPool := pool

	tripsRepo := database.NewTripRepository(pgPool)
	placesRepo := database.NewPlaceRepository(pgPool)
	expensesRepo := database.NewExpenseRepository(pgPool)
	authRepo := database.NewAuthRepository(pgPool)
	activityRepo := database.NewActivityRepository(pgPool)
	eventsRepo := database.NewEventRepository(pgPool)
	rateLimitRepo := database.NewRateLimitRepository(pgPool)
	analyticsRepo := database.NewAnalyticsRepository(pgPool)
	sessionRepo := database.NewSessionRepository(pgPool)
	emailSender := email.NoopSender{}

	exchangeRateSvc := services.NewExchangeRateService(pgPool)
	expenseSvc := services.NewExpenseService(tripsRepo, expensesRepo, exchangeRateSvc)
	tripSvc := services.NewTripService(pgPool, tripsRepo, placesRepo, activityRepo, expensesRepo, eventsRepo)
	cfg := config.Load()
	h := handlers.NewHandlers(
		tripsRepo, placesRepo, expensesRepo, authRepo,
		activityRepo, eventsRepo, rateLimitRepo, analyticsRepo, sessionRepo, expenseSvc, tripSvc, emailSender, cfg,
	)

	ctx := context.Background()

	// Setup: create a trip, add a place + activity, enable share, save token.
	ownerSession := "json-shape-owner-" + uuid.New().String()
	trip, err := tripsRepo.CreateTrip(ctx, nil, &ownerSession, false, models.Trip{
		Name:                   "JSON Shape Test",
		BaseCurrency:           "EUR",
		DefaultExpenseCurrency: "EUR",
		StartDate:              "2026-08-01",
		EndDate:                "2026-08-10",
	})
	require.NoError(t, err)

	defer func() {
		_, _ = pgPool.Exec(ctx, "DELETE FROM events WHERE trip_id = $1", trip.ID)
		_, _ = pgPool.Exec(ctx, "DELETE FROM activities WHERE trip_id = $1", trip.ID)
		_, _ = pgPool.Exec(ctx, "DELETE FROM places WHERE trip_id = $1", trip.ID)
		_, _ = pgPool.Exec(ctx, "DELETE FROM trips WHERE id = $1", trip.ID)
	}()

	// Add a place (so the publicPlaces array is non-empty).
	place, err := placesRepo.CreatePlace(ctx, trip.ID, models.Place{
		Name: "Test Place",
	})
	require.NoError(t, err)

	// Add an activity (so the publicActivities array is non-empty).
	placeID := place.ID
	now := time.Now()
	_, err = pgPool.Exec(ctx, `
		INSERT INTO activities (id, trip_id, place_id, title, date)
		VALUES ($1, $2, $3, $4, $5)
	`, uuid.New(), trip.ID, placeID, "Test Activity", now)
	require.NoError(t, err)

	// Enable share.
	token, _, err := tripsRepo.EnableShare(ctx, trip.ID.String(), nil, &ownerSession)
	require.NoError(t, err)

	// Build a router and hit GET /share/{token}.
	r := chi.NewRouter()
	r.Get("/share/{token}", h.GetSharedTrip)

	req := httptest.NewRequest(http.MethodGet, "/share/"+token, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// Decode body and inspect JSON shape.
	var body map[string]interface{}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))

	// publicPlaces[i].id must exist (was correctly tagged, but check anyway).
	places, ok := body["places"].([]interface{})
	require.True(t, ok, "body.places must be a list")
	require.NotEmpty(t, places, "test trip has a place — places list must not be empty")
	for i, p := range places {
		pm, _ := p.(map[string]interface{})
		require.NotNil(t, pm, "place[%d] must be an object", i)
		id, ok := pm["id"].(string)
		require.True(t, ok, "places[%d].id must be a string (frontend uses it as each-key)", i)
		assert.NotEmpty(t, id)
	}

	// publicActivities[i].id must exist — this was the bug.
	activities, ok := body["activities"].([]interface{})
	require.True(t, ok, "body.activities must be a list")
	require.NotEmpty(t, activities, "test trip has an activity — activities list must not be empty")
	for i, a := range activities {
		am, _ := a.(map[string]interface{})
		require.NotNil(t, am, "activity[%d] must be an object", i)
		id, ok := am["id"].(string)
		require.True(t, ok, "activities[%d].id must be a string (frontend uses it as each-key)", i)
		assert.NotEmpty(t, id, "activities[%d].id must not be empty (would produce duplicate each-keys)", i)
	}
}
