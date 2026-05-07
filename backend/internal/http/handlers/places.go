package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// GetPlace godoc
// @Summary      Get place details
// @Description  Get detailed information about a specific place (destination)
// @Tags         places
// @Produce      json
// @Param        id       path      string  true  "Trip ID (UUID)"
// @Param        placeId  path      string  true  "Place ID (UUID)"
// @Success      200      {object}  models.Place
// @Failure      404      {string}  string "Place not found"
// @Router       /trips/{id}/places/{placeId} [get]
func (h *Handlers) GetPlace(w http.ResponseWriter, r *http.Request) {
	placeID, _ := uuid.Parse(chi.URLParam(r, "placeId"))
	p, err := h.PlacesRepo.GetPlace(r.Context(), placeID)
	if err != nil {
		http.Error(w, "Place not found", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(p)
}

// ListPlaces godoc
// @Summary      List trip places
// @Description  Get all destinations (places) associated with a trip
// @Tags         places
// @Produce      json
// @Param        id   path      string  true  "Trip ID (UUID)"
// @Success      200  {array}   models.Place
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /trips/{id}/places [get]
func (h *Handlers) ListPlaces(w http.ResponseWriter, r *http.Request) {
	tripID, _ := uuid.Parse(middleware.GetWorkingTripID(r))
	places, err := h.PlacesRepo.ListPlacesByTrip(r.Context(), tripID)
	if err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(places)
}

// CreatePlace godoc
// @Summary      Create a place
// @Description  Add a new destination (place) to a trip
// @Tags         places
// @Accept       json
// @Produce      json
// @Param        id     path      string        true  "Trip ID (UUID)"
// @Param        place  body      models.Place  true  "Place data"
// @Success      201    {object}  models.Place
// @Failure      400    {string}  string "Invalid input"
// @Failure      500    {string}  string "Internal Server Error"
// @Router       /trips/{id}/places [post]
func (h *Handlers) CreatePlace(w http.ResponseWriter, r *http.Request) {
	tripID, _ := uuid.Parse(middleware.GetWorkingTripID(r))
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

// DeletePlace godoc
// @Summary      Delete a place
// @Description  Delete a destination and its associated activities
// @Tags         places
// @Param        id       path      string  true  "Trip ID (UUID)"
// @Param        placeId  path      string  true  "Place ID (UUID)"
// @Success      204      {string}  string "No Content"
// @Failure      500      {string}  string "Internal Server Error"
// @Router       /trips/{id}/places/{placeId} [delete]
func (h *Handlers) DeletePlace(w http.ResponseWriter, r *http.Request) {
	placeID, _ := uuid.Parse(chi.URLParam(r, "placeId"))
	if err := h.PlacesRepo.DeletePlace(r.Context(), placeID); err != nil {
		http.Error(w, "Delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
