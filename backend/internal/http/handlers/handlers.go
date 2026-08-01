package handlers

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/http/middleware"
	"backend/internal/services"
	"net/http"

	"github.com/google/uuid"
)

// Handlers wires the HTTP layer to the storage and service packages.
//
// All repository fields are typed against narrow interfaces (TripStore,
// PlaceStore, …) declared in the database package. The concrete
// *Repository types satisfy them structurally, so swapping a real DB for a
// test double or an alternative backend is a one-line change in main.go.
type Handlers struct {
	TripsRepo     database.TripStore
	PlacesRepo    database.PlaceStore
	ExpensesRepo  database.ExpenseStore
	AuthRepo      database.AuthStore
	ActivityRepo  database.ActivityStore
	EventsRepo    database.EventStore
	RateLimitRepo database.RateLimitStore
	AnalyticsRepo database.AnalyticsStore
	ExpenseSvc    *services.ExpenseService
	TripSvc       *services.TripService
	Config        *config.Config
}

func NewHandlers(
	tripsRepo database.TripStore,
	placesRepo database.PlaceStore,
	expensesRepo database.ExpenseStore,
	authRepo database.AuthStore,
	activityRepo database.ActivityStore,
	eventsRepo database.EventStore,
	rateLimitRepo database.RateLimitStore,
	analyticsRepo database.AnalyticsStore,
	expenseSvc *services.ExpenseService,
	tripSvc *services.TripService,
	cfg *config.Config,
) *Handlers {
	return &Handlers{
		TripsRepo:     tripsRepo,
		PlacesRepo:    placesRepo,
		ExpensesRepo:  expensesRepo,
		AuthRepo:      authRepo,
		ActivityRepo:  activityRepo,
		EventsRepo:    eventsRepo,
		RateLimitRepo: rateLimitRepo,
		AnalyticsRepo: analyticsRepo,
		ExpenseSvc:    expenseSvc,
		TripSvc:       tripSvc,
		Config:        cfg,
	}
}

// helper for all handlers
func extractIdentity(r *http.Request) (*uuid.UUID, *string) {
	var userID *uuid.UUID
	if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID); ok {
		userID = &uid
	}
	var sessionID *string
	if userID == nil {
		if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok &&
			sid != "" {
			sessionID = &sid
		}
	}

	return userID, sessionID
}

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message":"Service is running"}`))
}
