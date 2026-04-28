package handlers

import (
	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/services"
	"net/http"
)

type Handlers struct {
	TripsRepo    *database.TripRepository
	PlacesRepo   *database.PlaceRepository
	ExpensesRepo *database.ExpenseRepository
	AuthRepo     *database.AuthRepository
	ActivityRepo *database.ActivityRepository
	ExpenseSvc   *services.ExpenseService
	Config       *config.Config
}

func NewHandlers(
	tripsRepo *database.TripRepository,
	placesRepo *database.PlaceRepository,
	expensesRepo *database.ExpenseRepository,
	authRepo *database.AuthRepository,
	activityRepo *database.ActivityRepository,
	expenseSvc *services.ExpenseService,
	cfg *config.Config,
) *Handlers {
	return &Handlers{
		TripsRepo:    tripsRepo,
		PlacesRepo:   placesRepo,
		ExpensesRepo: expensesRepo,
		AuthRepo:     authRepo,
		ActivityRepo: activityRepo,
		ExpenseSvc:   expenseSvc,
		Config:       cfg,
	}
}

func (h *Handlers) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"Message":"Service is running"}`))
}
