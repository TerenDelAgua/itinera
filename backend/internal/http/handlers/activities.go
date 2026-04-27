package handlers

import (
	"backend/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (h *Handlers) ListActivities(w http.ResponseWriter, r *http.Request) {
	tripId := chi.URLParam(r, "trip_id")
	id, err := uuid.Parse(tripId)
	if err != nil {
		http.Error(w, "Invalid trip Idd", http.StatusBadRequest)
		return
	}

	activities, err := h.DB.GetByTripId(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

func (h *Handlers) CreateActivity(w http.ResponseWriter, r *http.Request) {
	tripIdStr := chi.URLParam(r, "trip_id")
	tripId, err := uuid.Parse(tripIdStr)
	if err != nil {
		http.Error(w, "Invalid trip Id", http.StatusBadRequest)
		return
	}

	var input struct {
		Title   string  `json:"title"`
		Date    string  `json:"date"`
		Time    *string `json:"time"`
		Notes   string  `json:"notes"`
		PlaceId *string `json:"place_id"`
	}

	if err = json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if input.Title == "" || input.Date == "" {
		http.Error(w, "Title and Date are required", http.StatusBadRequest)
		return
	}

	trip, err := h.DB.GetTripById(r.Context(), tripId)
	if err != nil {
		http.Error(w, "Trip not found", http.StatusNotFound)
		return
	}

	if input.Date < trip.StartDate || input.Date > trip.EndDate {
		http.Error(w, "Activity date must be within trip dates", http.StatusBadRequest)
		return
	}

	var placeId uuid.UUID
	if input.PlaceId != nil {
		placeId, err = uuid.Parse(*input.PlaceId)
		if err != nil {
			http.Error(w, "invalid place ID", http.StatusBadRequest)
			return
		}

		place, err := h.DB.GetPlace(r.Context(), placeId)
		if err != nil || place.TripId != tripId {
			http.Error(w, "Place does not belong to this trip", http.StatusForbidden)
			return
		}
	}

	newActivity := models.Activity{
		Id:      uuid.New(),
		TripId:  tripId,
		PlaceId: &placeId,
		Title:   input.Title,
		Date:    input.Date,
		Time:    input.Time,
		Notes:   &input.Notes,
	}

	if err := h.DB.CreateActivity(r.Context(), &newActivity); err != nil {
		http.Error(w, "Failed to create activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newActivity)
}

func (h *Handlers) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid Id", http.StatusBadRequest)
		return
	}

	if err := h.DB.DeleteActivity(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
