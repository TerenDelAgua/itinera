package handlers

import (
	"backend/internal/http/middleware"
	"net/http"

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

		// Post-cutover login. Uses opaque access +
		// refresh cookies. Mounted alongside the JWT path during the
		// dual-stack window. Behaviour is identical to /auth/login but
		// the response omits the JWT and the cookies are `itinera_access`
		// + `itinera_refresh` instead of `auth_token`. Handler name is
		// LoginOpaque (see opaque_auth_handlers.go) — the route keeps the
		// /v2/ suffix for spec compatibility.
		r.Post("/auth/v2/login", h.LoginOpaque)

		// Public: /auth/v2/register, /auth/v2/forgot, /auth/v2/reset
		// All anti-enumeration by design:
		// the body shape is identical whether the email exists or not.
		r.Post("/auth/v2/register", h.RegisterOpaque)
		r.Post("/auth/v2/forgot", h.ForgotOpaque)
		r.Post("/auth/v2/reset", h.ResetOpaque)

		// Authenticated endpoints behind
		// AuthMiddlewareV2. Note: /auth/v2/logout and /auth/v2/me require
		// the access-token cookie, NOT the JWT path. We mount
		// AuthMiddlewareV2 alongside AuthMiddleware so during the
		// dual-stack window either cookie succeeds — the handlers then
		// short-circuit if the right middleware resolved them.
		r.With(middleware.AuthMiddlewareV2(h.SessionRepo)).Group(func(r chi.Router) {
			r.Post("/auth/v2/logout", h.LogoutOpaque)
			r.Get("/auth/v2/me", h.MeOpaque)
			r.Delete("/auth/v2/account", h.DeleteAccountOpaque)
		})

		// Refresh needs NO access cookie — by definition, the access
		// cookie expired (or expired soon); the refresh cookie is the
		// only credential the client still has. So it's a sibling of
		// the auth group rather than a child: public access, no JWT
		// middleware, no opaque middleware.
		r.Post("/auth/v2/refresh", h.RefreshOpaque)

		r.Get("/trips", h.ListTrips)
		r.Post("/trips", h.CreateTrip)

		r.Post("/auth/claim-guest", h.ClaimGuest)
		r.Method(http.MethodPost, "/auth/upgrade", upgradeAlias())

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

// upgradeAlias returns an http.Handler that responds 301 to /auth/claim-guest,
// preserving the request method and any cookies. It exists only as a backwards-
// compatible shim during the cutover window and MUST be removed once the v3
// rollout completes (Spec 017 §15 fase 4 paso 23).
func upgradeAlias() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.URL.Scheme + "://" + r.URL.Host + "/api/v1/auth/claim-guest"
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}
