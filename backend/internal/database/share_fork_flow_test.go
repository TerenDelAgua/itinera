package database_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"backend/internal/database"
	"backend/internal/http/handlers"
	"backend/internal/http/middleware"
	"backend/internal/models"
	"backend/internal/services"

	"bytes"
	"encoding/json"
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

	exchangeRateSvc := services.NewExchangeRateService(pgPool)
	expenseSvc := services.NewExpenseService(tripsRepo, expensesRepo, exchangeRateSvc)
	tripSvc := services.NewTripService(pgPool, tripsRepo, placesRepo, activityRepo, expensesRepo, eventsRepo)

	h := handlers.NewHandlers(
		tripsRepo, placesRepo, expensesRepo, authRepo,
		activityRepo, eventsRepo, rateLimitRepo, expenseSvc, tripSvc, nil,
	)

	ctx := context.Background()

	// 1. Create a public demo owned by a different session.
	originalOwner := "demo-owner-" + uuid.New().String()
	demo, err := h.TripsRepo.CreateTrip(ctx, nil, &originalOwner, models.Trip{
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
