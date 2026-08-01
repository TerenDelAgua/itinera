package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"backend/internal/http/middleware"
)

// This file hosts the admin-only analytics endpoints called out in
// Spec 015 v2 (sections 5.2 and 5.4). Every endpoint:
//
//   - verifies the request carries the internal token via
//     middleware.IsInternalSession (returns 401 otherwise);
//   - delegates ALL database work to the AnalyticsRepository, never
//     touches the connection pool directly (matches the project-wide
//     "handler -> repo" pattern);
//   - encodes the response as JSON with a single Content-Type header.
//
// Allowed query parameters (when applicable) are clamped to a sane range
// to avoid expensive scans caused by a malicious or buggy client.

const (
	defaultRetentionDays = 7
	maxRetentionDays     = 365
	defaultPowerUsersLimit = 20
	maxPowerUsersLimit     = 100
)

// safePct returns 100 * num / den, or 0 when den is 0. Used by every
// percentage field in the analytics responses.
func safePct(num, den int) float64 {
	if den == 0 {
		return 0
	}
	return 100.0 * float64(num) / float64(den)
}

// analyticsGuard is the shared 401 short-circuit. Pulled out so each
// endpoint reads as a single, declarative predicate.
func (h *Handlers) analyticsGuard(w http.ResponseWriter, r *http.Request) bool {
	if !middleware.IsInternalSession(r) {
		http.Error(w, "Internal token required", http.StatusUnauthorized)
		return false
	}
	return true
}

// AnalyticsOverview godoc
// @Summary      Real-user overview
// @Description  High-level metrics: real users, trips, DAU/WAU/MAU, total expenses
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  models.AnalyticsOverview
// @Failure      401  {string}  string "Internal token required"
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /analytics/overview [get]
func (h *Handlers) AnalyticsOverview(w http.ResponseWriter, r *http.Request) {
	if !h.analyticsGuard(w, r) {
		return
	}

	out, err := h.AnalyticsRepo.GetAnalyticsOverview(r.Context())
	if err != nil {
		log.Printf("ERROR AnalyticsOverview: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// AnalyticsActivation godoc
// @Summary      Activation funnel
// @Description  % of landing sessions that created a trip within the cohort window
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  models.AnalyticsActivation
// @Failure      401  {string}  string "Internal token required"
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /analytics/activation [get]
func (h *Handlers) AnalyticsActivation(w http.ResponseWriter, r *http.Request) {
	if !h.analyticsGuard(w, r) {
		return
	}

	out, err := h.AnalyticsRepo.GetAnalyticsActivation(r.Context())
	if err != nil {
		log.Printf("ERROR AnalyticsActivation: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// AnalyticsEngagement godoc
// @Summary      User engagement depth
// @Description  Avg/median expenses per trip, power-user counts
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  models.AnalyticsEngagement
// @Failure      401  {string}  string "Internal token required"
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /analytics/engagement [get]
func (h *Handlers) AnalyticsEngagement(w http.ResponseWriter, r *http.Request) {
	if !h.analyticsGuard(w, r) {
		return
	}

	out, err := h.AnalyticsRepo.GetAnalyticsEngagement(r.Context())
	if err != nil {
		log.Printf("ERROR AnalyticsEngagement: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// AnalyticsRetention godoc
// @Summary      Retention cohort
// @Description  % of users that returned N days after their first trip
// @Tags         analytics
// @Produce      json
// @Param        days  query  int  false  "Days after first trip (default 7, max 365)"
// @Success      200   {object}  models.AnalyticsRetention
// @Failure      401   {string}  string "Internal token required"
// @Failure      500   {string}  string "Internal Server Error"
// @Router       /analytics/retention [get]
func (h *Handlers) AnalyticsRetention(w http.ResponseWriter, r *http.Request) {
	if !h.analyticsGuard(w, r) {
		return
	}

	days := defaultRetentionDays
	if d := r.URL.Query().Get("days"); d != "" {
		parsed, err := strconv.Atoi(d)
		if err != nil || parsed <= 0 {
			http.Error(w, "days must be a positive integer", http.StatusBadRequest)
			return
		}
		if parsed > maxRetentionDays {
			parsed = maxRetentionDays
		}
		days = parsed
	}

	out, err := h.AnalyticsRepo.GetAnalyticsRetention(r.Context(), days)
	if err != nil {
		log.Printf("ERROR AnalyticsRetention: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// AnalyticsFunnel godoc
// @Summary      Ghost Mode funnel
// @Description  demo_viewed -> demo_deep_forked -> demo_fork_reused counts
// @Tags         analytics
// @Produce      json
// @Success      200  {object}  models.AnalyticsFunnel
// @Failure      401  {string}  string "Internal token required"
// @Failure      500  {string}  string "Internal Server Error"
// @Router       /analytics/funnel [get]
func (h *Handlers) AnalyticsFunnel(w http.ResponseWriter, r *http.Request) {
	if !h.analyticsGuard(w, r) {
		return
	}

	out, err := h.AnalyticsRepo.GetAnalyticsFunnel(r.Context())
	if err != nil {
		log.Printf("ERROR AnalyticsFunnel: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// AnalyticsPowerUsers godoc
// @Summary      Top engaged users
// @Description  List of users ranked by trip/expense count
// @Tags         analytics
// @Produce      json
// @Param        limit  query  int  false  "Max rows (default 20, max 100)"
// @Success      200    {object}  models.AnalyticsPowerUsers
// @Failure      401    {string}  string "Internal token required"
// @Failure      500    {string}  string "Internal Server Error"
// @Router       /analytics/power-users [get]
func (h *Handlers) AnalyticsPowerUsers(w http.ResponseWriter, r *http.Request) {
	if !h.analyticsGuard(w, r) {
		return
	}

	limit := defaultPowerUsersLimit
	if l := r.URL.Query().Get("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed <= 0 {
			http.Error(w, "limit must be a positive integer", http.StatusBadRequest)
			return
		}
		if parsed > maxPowerUsersLimit {
			parsed = maxPowerUsersLimit
		}
		limit = parsed
	}

	out, err := h.AnalyticsRepo.GetAnalyticsPowerUsers(r.Context(), limit)
	if err != nil {
		log.Printf("ERROR AnalyticsPowerUsers: %v", err)
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, out)
}

// writeJSON encodes v as JSON with the standard Content-Type header.
// Centralised so every analytics endpoint has the same shape.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
