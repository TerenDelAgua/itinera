package handlers

import (
	"backend/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) GetPlace(w http.ResponseWriter, r *http.Request) {
	placeID, _ := uuid.Parse(chi.URLParam(r, "placeId"))
	p, err := h.PlacesRepo.GetPlace(r.Context(), placeID)
	if err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(p)
}

func (h *Handlers) ListPlaces(w http.ResponseWriter, r *http.Request) {
	tripID, _ := uuid.Parse(chi.URLParam(r, "id"))
	places, err := h.PlacesRepo.ListPlacesByTrip(r.Context(), tripID)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(places)
}

func (h *Handlers) CreatePlace(w http.ResponseWriter, r *http.Request) {
	tripID, _ := uuid.Parse(chi.URLParam(r, "id"))
	var input models.Place
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	p, err := h.PlacesRepo.CreatePlace(r.Context(), tripID, input)
	if err != nil {
		http.Error(w, "Create failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (h *Handlers) UpdatePlace(w http.ResponseWriter, r *http.Request) {
	placeID, _ := uuid.Parse(chi.URLParam(r, "placeId"))
	var updates map[string]any
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	p, err := h.PlacesRepo.UpdatePlace(r.Context(), placeID, updates)
	if err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(p)
}

func (h *Handlers) DeletePlace(w http.ResponseWriter, r *http.Request) {
	placeID, _ := uuid.Parse(chi.URLParam(r, "placeId"))
	if err := h.PlacesRepo.DeletePlace(r.Context(), placeID); err != nil {
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
