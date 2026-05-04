package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/models"
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CreateTrip godoc
// @Summary      Create a trip
// @Description  Create a new trip associated with the current user or session
// @Tags         trips
// @Accept       json
// @Produce      json
// @Param        trip  body      models.Trip  true  "Trip data"
// @Success      201   {object}  models.Trip
// @Failure      400   {string}  string "Invalid JSON"
// @Failure      500   {string}  string "Internal Server Error"
// @Router       /trips [post]
func (h *Handlers) CreateTrip(w http.ResponseWriter, r *http.Request) {
	var input models.Trip
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// 1. Try authenticated user first
	var userID *uuid.UUID
	if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID); ok {
		userID = &uid
	}

	// 2. Fallback to guest session
	var sessionID *string
	if userID == nil {
		if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
			sessionID = &sid
		}
	}

	if userID == nil && sessionID == nil {
		http.Error(w, "Identity missing", http.StatusInternalServerError)
		return
	}

	newTrip, err := h.TripsRepo.CreateTrip(r.Context(), userID, sessionID, input)
	if err != nil {
		log.Printf("ERROR creating trip: %v", err)
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTrip)
}

// ListTrips godoc
// @Summary      List trips
// @Description  Get a list of trips associated with the current user or session
// @Tags         trips
// @Accept       json
// @Produce      json
// @Success      200  {array}   models.Trip
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /trips [get]
func (h *Handlers) ListTrips(w http.ResponseWriter, r *http.Request) {
	// 1. Try authenticated user first
	var userID *uuid.UUID
	if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID); ok {
		userID = &uid
	}

	// 2. Fallback to guest session
	var sessionID *string
	if userID == nil {
		if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
			sessionID = &sid
		}
	}

	if userID == nil && sessionID == nil {
		http.Error(w, "Identity missing", http.StatusInternalServerError)
		return
	}

	trips, err := h.TripsRepo.ListTrips(r.Context(), userID, sessionID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trips)
}

// GetTrip godoc
// @Summary      Get trip details
// @Description  Get detailed information about a specific trip by ID
// @Tags         trips
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Trip ID (UUID)"
// @Success      200  {object}  models.Trip
// @Failure      404  {string}  string "Trip not found"
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /trips/{id} [get]
func (h *Handlers) GetTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	 log.Printf("🔍 [GetTrip] ID recibido: '%s'", id)
    log.Printf("🔍 [GetTrip] URL completa: %s", r.URL.String())
    

	// 1. Try authenticated user first
	var userID *uuid.UUID
	if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID); ok {
		userID = &uid
	}

	// 2. Fallback to guest session
	var sessionID *string
	if userID == nil {
		if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
			sessionID = &sid
		}
	}

	if userID == nil && sessionID == nil {
		http.Error(w, "Identity missing", http.StatusInternalServerError)
		return
	}

	sidVal := "nil"
	if sessionID != nil {
		sidVal = *sessionID
	}
	log.Printf("[DEBUG] GetTrip - ID: %s, UserID: %v, SessionID: %s", id, userID, sidVal)

	// Validate UUID format before querying DB to avoid syntax errors
	if _, err := uuid.Parse(id); err != nil {
		http.Error(w, "Invalid trip ID format", http.StatusBadRequest)
		return
	}

	trip, err := h.TripsRepo.GetTrip(r.Context(), id, userID, sessionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "Trip not found or access denied", http.StatusNotFound)
		} else {
			http.Error(w, "Internal database error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trip)
}

// UpdateTrip godoc
// @Summary      Update a trip
// @Description  Partially update trip information
// @Tags         trips
// @Accept       json
// @Produce      json
// @Param        id    path      string          true  "Trip ID (UUID)"
// @Param        data  body      map[string]any  true  "Partial trip data"
// @Success      200   {object}  models.Trip
// @Failure      400   {string}  string "Invalid input"
// @Failure      403   {string}  string "Forbidden"
// @Failure      500   {string}  string "Internal Server Error"
// @Router       /trips/{id} [put]
func (h *Handlers) UpdateTrip(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(tripID); err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	// Parse partial update payload
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if len(updates) == 0 {
		http.Error(w, "No fields to update", http.StatusBadRequest)
		return
	}

	// Extract identity (User or Guest)
	var userID *uuid.UUID
	if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID); ok {
		userID = &uid
	}
	var sessionID *string
	if userID == nil {
		if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
			sessionID = &sid
		}
	}

	updatedTrip, err := h.TripsRepo.UpdateTrip(r.Context(), tripID, userID, sessionID, updates)
	if err != nil {
		if err.Error() == "trip not found or unauthorized" {
			http.Error(w, "Forbidden", http.StatusForbidden)
		} else {
			http.Error(w, "Database error", http.StatusInternalServerError)
		}
		return
	}

	// Recalculate expenses if base_currency changed
	if _, ok := updates["base_currency"]; ok {
		tID, _ := uuid.Parse(tripID)
		// We do it synchronously to ensure the next fetch gets updated data
		if err := h.ExpenseSvc.RecalculateTripExpenses(r.Context(), tID); err != nil {
			log.Printf("ERROR recalculating expenses for trip %s: %v", tripID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedTrip)
}

// DeleteTrip godoc
// @Summary      Delete a trip
// @Description  Delete a trip and all its associated data
// @Tags         trips
// @Param        id   path      string  true  "Trip ID (UUID)"
// @Success      204  {string}  string "No Content"
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /trips/{id} [delete]
func (h *Handlers) DeleteTrip(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// 1. Try authenticated user first
	var userID *uuid.UUID
	if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID); ok {
		userID = &uid
	}

	// 2. Fallback to guest session
	var sessionID *string
	if userID == nil {
		if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
			sessionID = &sid
		}
	}

	if userID == nil && sessionID == nil {
		http.Error(w, "Identity missing", http.StatusInternalServerError)
		return
	}

	err := h.TripsRepo.DeleteTrip(r.Context(), id, userID, sessionID)
	if err != nil {
		http.Error(w, "Failed to delete trip", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
