package handlers

import (
	"backend/internal/http/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterApiRoutes(r chi.Router, h *Handlers) {
	r.Use(middleware.SessionMiddleware(h.Config.IsProduction()))
	r.Use(middleware.AuthMiddleware(h.Config.JWTSecret))

	r.Get("/health", h.healthCheck)
	r.Get("/stats/public", h.GetPublicStats)

	// Analytics (admin-only, behind IsInternalSession — 401 without token).
	// No trip context required: these endpoints aggregate across sessions.
	r.Get("/api/analytics/sessions", h.AnalyticsSessions)
	r.Get("/api/analytics/overview", h.AnalyticsOverview)
	r.Get("/api/analytics/activation", h.AnalyticsActivation)
	r.Get("/api/analytics/engagement", h.AnalyticsEngagement)
	r.Get("/api/analytics/retention", h.AnalyticsRetention)
	r.Get("/api/analytics/funnel", h.AnalyticsFunnel)
	r.Get("/api/analytics/power-users", h.AnalyticsPowerUsers)

	// Public Share endpoints
	r.Get("/share/{token}", h.GetSharedTrip)
	r.Post("/share/fork", h.ForkFromShare) //require auth

	r.Group(func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		r.Get("/trips", h.ListTrips)
		r.Post("/trips", h.CreateTrip)

		r.Post("/auth/upgrade", h.UpgradeSession)

		// Events: fire-and-forget analytics (no PII, rate-limited)
		r.Post("/events", h.TrackEvent)

		r.With(middleware.ResolveTripContext(h.TripSvc)).Group(func(r chi.Router) {
			r.Get("/trips/{id}", h.GetTrip)
			r.Put("/trips/{id}", h.UpdateTrip)
			r.Delete("/trips/{id}", h.DeleteTrip)

			r.Post("/trips/{id}/share", h.EnableShare)
			r.Delete("/trips/{id}/share", h.DisableShare)

			r.Get("/trips/{id}/places", h.ListPlaces)
			r.Post("/trips/{id}/places", h.CreatePlace)
			r.Get("/trips/{id}/places/{placeId}", h.GetPlace)
			r.Put("/trips/{id}/places/{placeId}", h.UpdatePlace)
			r.Delete("/trips/{id}/places/{placeId}", h.DeletePlace)

			r.Get("/trips/{id}/expenses", h.ListExpenses)
			r.Get("/trips/{id}/expenses/categories", h.GetCategories)
			r.Post("/trips/{id}/expenses", h.CreateExpense)
			r.Put("/trips/{id}/expenses/{expenseId}", h.UpdateExpense)
			r.Get("/trips/{id}/expenses/summary", h.GetExpenseSummary)
			r.Get("/trips/{id}/rates", h.GetExchangeRate)
			r.Delete("/trips/{id}/expenses/{expenseId}", h.DeleteExpense)

			r.Get("/trips/{id}/places/{placeId}/expenses", h.ListPlaceExpenses)
			r.Post("/trips/{id}/places/{placeId}/expenses", h.CreateExpense)
			r.Get("/trips/{id}/places/{placeId}/expenses/summary", h.GetPlaceExpenseSummary)

			//Activities
			r.Get("/trips/{trip_id}/activities", h.ListActivities)
			r.Post("/trips/{trip_id}/activities", h.CreateActivity)
			r.Put("/trips/{trip_id}/activities/{id}", h.UpdateActivity)
			r.Delete("/trips/{trip_id}/activities/{id}", h.DeleteActivity)
		})
	})
}
