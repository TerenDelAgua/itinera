package handlers

import (
	"backend/internal/http/middleware"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/cors"
)

func ApiRouter(h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{AllowedOrigins: []string{"*"}, AllowCredentials: true}))
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

	return r
}
