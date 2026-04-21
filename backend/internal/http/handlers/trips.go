package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/models"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
)

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

	newTrip, err := h.DB.CreateTrip(r.Context(), userID, sessionID, input)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTrip)
}

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

	trips, err := h.DB.ListTrips(r.Context(), userID, sessionID)
	if err != nil {
		http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trips)
}

func (h *Handlers) GetTrip(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"Message":"endpoint not implemented yet"}`))
}
