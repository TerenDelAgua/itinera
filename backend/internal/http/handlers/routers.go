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

	r.Group(func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		r.Get("/trips", h.ListTrips)
		r.Post("/trips", h.CreateTrip)

		r.Post("/auth/upgrade", h.UpgradeSession)

		// Events: fire-and-forget analytics (no PII, rate-limited)
		r.Post("/events", TrackEvent(h.TripsRepo.Pool))

		r.With(middleware.ResolveTripContext(h.TripsRepo)).Group(func(r chi.Router) {
			r.Get("/trips/{id}", h.GetTrip)
			r.Put("/trips/{id}", h.UpdateTrip)
			r.Delete("/trips/{id}", h.DeleteTrip)

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
