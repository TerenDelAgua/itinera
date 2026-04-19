package handlers

import (
	"backend/internal/http/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterApiRoutes(r chi.Router, h *Handlers) {
	r.Use(middleware.SessionMiddleware)                  // ← Siempre activo
	r.Use(middleware.AuthMiddleware(h.Config.JWTSecret)) // ← Opcional, no bloquea

	r.Get("/health", h.healthCheck)

	r.Group(func(r chi.Router) {
		r.Post("/auth/register", h.Register)
		r.Post("/auth/login", h.Login)

		r.Get("/trips", h.ListTrips)
		r.Post("/trips", h.CreateTrip)

		r.Post("/auth/upgrade", h.UpgradeSession)

		r.Get("/trips/{id}", h.GetTrip)
		r.Put("/trips/{id}", h.UpdateTrip)
		r.Delete("/trips/{id}", h.DeleteTrip)
	})
}
