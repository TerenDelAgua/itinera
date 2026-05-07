package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/models"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)



// ListActivities godoc
// @Summary      List trip activities
// @Description  Get all activities (agenda) for a specific trip
// @Tags         activities
// @Produce      json
// @Param        trip_id  path      string  true  "Trip ID (UUID)"
// @Success      200      {array}   models.Activity
// @Failure      400      {string}  string "Invalid ID"
// @Failure      500      {string}  string "Internal Server Error"
// @Router       /trips/{trip_id}/activities [get]
func (h *Handlers) ListActivities(w http.ResponseWriter, r *http.Request) {
	tripId := middleware.GetWorkingTripID(r)
	id, err := uuid.Parse(tripId)
	if err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	activities, err := h.ActivityRepo.GetByTripId(r.Context(), id)
	if err != nil {
		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
}

// CreateActivity godoc
// @Summary      Create an activity
// @Description  Add a new activity to the trip agenda
// @Tags         activities
// @Accept       json
// @Produce      json
// @Param        trip_id   path      string  true  "Trip ID (UUID)"
// @Param        activity  body      object  true  "Activity data (title, date, time, notes, place_id)"
// @Success      201       {object}  models.Activity
// @Failure      400       {string}  string "Invalid input"
// @Failure      500       {string}  string "Internal Server Error"
// @Router       /trips/{trip_id}/activities [post]
func (h *Handlers) CreateActivity(w http.ResponseWriter, r *http.Request) {
	tripIdStr := middleware.GetWorkingTripID(r)
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

	trip, err := h.TripsRepo.GetTripById(r.Context(), tripId)
	if err != nil {
		http.Error(w, "Trip not found", http.StatusNotFound)
		return
	}

	tripStart := trip.StartDate
	if len(tripStart) > 10 {
		tripStart = tripStart[:10]
	}
	tripEnd := trip.EndDate
	if len(tripEnd) > 10 {
		tripEnd = tripEnd[:10]
	}

	if input.Date < tripStart || input.Date > tripEnd {
		http.Error(w, "Activity date must be within trip dates", http.StatusBadRequest)
		return
	}

	var placeIdPtr *uuid.UUID
	if input.PlaceId != nil && *input.PlaceId != "" {
		pId, err := uuid.Parse(*input.PlaceId)
		if err != nil {
			http.Error(w, "invalid place ID", http.StatusBadRequest)
			return
		}

		place, err := h.PlacesRepo.GetPlace(r.Context(), pId)
		if err != nil || place.TripId != tripId {
			http.Error(w, "Place does not belong to this trip", http.StatusForbidden)
			return
		}
		placeIdPtr = &pId
	}

	var notesPtr *string
	if input.Notes != "" {
		notesPtr = &input.Notes
	}

	newActivity := models.Activity{
		Id:      uuid.New(),
		TripId:  tripId,
		PlaceId: placeIdPtr,
		Title:   input.Title,
		Date:    input.Date,
		Time:    input.Time,
		Notes:   notesPtr,
	}

	if err := h.ActivityRepo.CreateActivity(r.Context(), &newActivity); err != nil {
		http.Error(w, "Failed to create activity", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newActivity)
}

func (h *Handlers) UpdateActivity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var input struct {
		Title   *string `json:"title"`
		Time    *string `json:"time"`
		Notes   *string `json:"notes"`
		PlaceID *string `json:"place_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	current, err := h.ActivityRepo.GetById(r.Context(), id)
	if err != nil {
		http.Error(w, "Activity not found", http.StatusNotFound)
		return
	}

	if input.Title != nil {
		current.Title = *input.Title
	}
	if input.Time != nil {
		current.Time = input.Time
	}
	if input.Notes != nil {
		current.Notes = input.Notes
	}
	if input.PlaceID != nil {
		uid, _ := uuid.Parse(*input.PlaceID)
		current.PlaceId = &uid
	}

	if err := h.ActivityRepo.UpdateActivity(r.Context(), current); err != nil {
		http.Error(w, "Update failed", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteActivity godoc
// @Summary      Delete an activity
// @Description  Delete an activity by ID
// @Tags         activities
// @Param        trip_id  path      string  true  "Trip ID (UUID)"
// @Param        id       path      string  true  "Activity ID (UUID)"
// @Success      204      {string}  string "No Content"
// @Failure      400      {string}  string "Invalid ID"
// @Failure      500      {string}  string "Internal Server Error"
// @Router       /trips/{trip_id}/activities/{id} [delete]
func (h *Handlers) DeleteActivity(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid Id", http.StatusBadRequest)
		return
	}

	if err := h.ActivityRepo.DeleteActivity(r.Context(), id); err != nil {
		http.Error(w, "Failed to delete", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)

}
