package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// EnableShare godoc
// @Summary      Enable trip share link
// @Description  Ghost users get 7-day expiry; registered users get permanent.
// @Tags         trips,share
// @Produce      json
// @Param        id   path      string  true  "Trip ID"
// @Success      200   {object}  map[string]any
// @Failure      403   {string}  string  "Forbidden"
// @Failure      404   {string}  string  "Trip not found"
// @Router       /trips/{id}/share [post]
func (h *Handlers) EnableShare(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(tripID); err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	userID, sessionID := extractIdentity(r)

	token, expiresAt, err := h.TripsRepo.EnableShare(r.Context(), tripID, userID, sessionID)
	if err != nil {
		switch err.Error() {
		case "trip not found":
			http.Error(w, "Trip not found", http.StatusNotFound)
		case "unauthorized":
			http.Error(w, "Forbidden", http.StatusForbidden)
		default:
			log.Printf("ERROR EnableShare: %v", err)
			http.Error(w, "Internal error", http.StatusInternalServerError)
		}
		return
	}

	if count, _ := h.TripsRepo.CountActiveSharesByUser(r.Context(), userID, sessionID); count > 10 {
		log.Printf("WARN user has %d active shares (>10 threshold)", count)
	}

	//TODO: Review hardcode url
	response := map[string]any{
		"token":      token,
		"url":        "https://goitinera.app/share/" + token,
		"expires_at": nil,
	}
	if expiresAt != nil {
		response["expires_at"] = expiresAt.Format(time.RFC3339)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// DisableShare godoc
// @Summary      Disable trip share link
// @Tags         trips,share
// @Param        id   path      string  true  "Trip ID"
// @Success      204   {string}  string  "No Content"
// @Failure      403   {string}  string  "Forbidden"
// @Router       /trips/{id}/share [delete]
func (h *Handlers) DisableShare(w http.ResponseWriter, r *http.Request) {
	tripID := chi.URLParam(r, "id")
	if _, err := uuid.Parse(tripID); err != nil {
		http.Error(w, "Invalid trip ID", http.StatusBadRequest)
		return
	}

	userID, sessionID := extractIdentity(r)

	if err := h.TripsRepo.DisableShare(r.Context(), tripID, userID, sessionID); err != nil {
		if err.Error() == "trip not found or unauthorized" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		log.Printf("ERROR DisableShare: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GetSharedTrip godoc
// @Summary      Get a shared trip (public, sanitized)
// @Description  Returns trip + places + activities + expense_summary. Rate limited 60/min per IP.
// @Tags         share
// @Produce      json
// @Param        token   path      string  true  "Share token"
// @Success      200     {object}  map[string]any
// @Failure      404     {string}  string  "Not found or expired"
// @Router       /share/{token} [get]
func (h *Handlers) GetSharedTrip(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	if len(token) != 12 {
		http.Error(w, "Invalid token", http.StatusBadRequest)
		return
	}

	// Rate limit per IP
	ipKey := "share_view:ip" + r.RemoteAddr
	if allowed, _ := h.RateLimitRepo.CheckAndIncrement(r.Context(), ipKey, time.Minute, 60); !allowed {
		http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
		return
	}

	trip, err := h.TripsRepo.GetByShareToken(r.Context(), token)
	if err != nil {
		//404 Generic.
		http.Error(w, "Share not found or expired", http.StatusNotFound)
		return
	}

	type publicPlace struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		Notes     *string `json:"notes,omitempty"`
		StartDate *string `json:"start_date,omitempty"`
		EndDate   *string `json:"end_date,omitempty"`
	}

	type publicActivity struct {
		ID      string
		PlaceID *string `json:"place_id,omitempty"`
		Title   string  `json:"title"`
		Date    string  `json:"date"`
		Time    *string `json:"time,omitempty"`
	}

	places, _ := h.PlacesRepo.ListPlacesByTrip(r.Context(), trip.ID)
	publicPlaces := make([]publicPlace, len(places))
	for i, p := range places {
		publicPlaces[i] = publicPlace{
			ID:        p.ID.String(),
			Name:      p.Name,
			Notes:     &p.Notes,
			StartDate: p.StartDate,
			EndDate:   p.EndDate,
		}
	}

	activities, _ := h.ActivityRepo.GetByTripId(r.Context(), trip.ID)
	publicActivities := make([]publicActivity, len(activities))
	for i, a := range activities {
		pidStr := ""
		if a.PlaceId != nil {
			pidStr = a.PlaceId.String()
		}
		var pidPtr *string
		if pidStr != "" {
			pidPtr = &pidStr
		}
		publicActivities[i] = publicActivity{
			ID:      a.Id.String(),
			PlaceID: pidPtr,
			Title:   a.Title,
			Date:    a.Date,
			Time:    a.Time,
		}
	}

	summary, _ := h.ExpensesRepo.GetExpensesSummary(r.Context(), trip.ID)

	response := map[string]any{
		"id":                       trip.ID,
		"name":                     trip.Name,
		"description":              trip.Description,
		"start_date":               trip.StartDate,
		"end_date":                 trip.EndDate,
		"base_currency":            trip.BaseCurrency,
		"default_expense_currency": trip.DefaultExpenseCurrency,
		"places":                   publicPlaces,
		"activities":               publicActivities,
		"expense_summary": map[string]any{
			"grand_total": summary.GrandTotal,
			"currency":    trip.BaseCurrency,
			"by_category": summary.ByCategory,
		},
	}
	if trip.ShareExpiresAt != nil {
		response["share_expires_at"] = *trip.ShareExpiresAt
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ForkFromShare godoc
// @Summary      Fork a trip from a share token
// @Tags         share
// @Accept       json
// @Produce      json
// @Param        body  body      object{token=string}  true  "Share token"
// @Success      201   {object}  models.Trip
// @Failure      404   {string}  string  "Not found or expired"
// @Failure      429   {string}  string  "Rate limit"
// @Router       /share/fork [post]
func (h *Handlers) ForkFromShare(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Token == "" {
		http.Error(w, "Invalid body: token required", http.StatusBadRequest)
		return
	}

	userId, sessionID := extractIdentity(r)

	// Rate limit per user/session
	var rateKey string
	if userId != nil {
		rateKey = "fork_from_share:user:" + userId.String()
	} else if sessionID != nil {
		rateKey = "fork_from_share:session:" + *sessionID
	}
	if rateKey != "" {
		if allowed, _ := h.RateLimitRepo.CheckAndIncrement(r.Context(), rateKey, time.Minute, 30); !allowed {
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}
	}

	newTrip, err := h.TripSvc.ForkFromShareToken(r.Context(), body.Token, userId, sessionID)
	if err != nil {
		if err == pgx.ErrNoRows {
			http.Error(w, "Share not found or expired", http.StatusNotFound)
			return
		}
		log.Printf("ERROR ForkFromShare: %v", err)
		http.Error(w, "Internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newTrip)
}
