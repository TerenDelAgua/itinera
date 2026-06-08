package handlers

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/services"
	"net/http"
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
	ExpenseSvc    *services.ExpenseService
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
	expenseSvc *services.ExpenseService,
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
		ExpenseSvc:    expenseSvc,
		Config:        cfg,
	}
}

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message":"Service is running"}`))
}
