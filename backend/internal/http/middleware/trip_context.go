package middleware

import (
	"backend/internal/database"
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type contextKey string

const WorkingTripIDKey contextKey = "workingTripID"

func ResolveTripContext(store database.TripContextStore) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			tripID := chi.URLParam(r, "trip_id")
			if tripID == "" {
				tripID = chi.URLParam(r, "id")
			}

			if tripID == "" {
				next.ServeHTTP(w, r)
				return
			}

			var userID *uuid.UUID
			if uid, ok := ctx.Value(ContextKeyUserId{}).(uuid.UUID); ok {
				userID = &uid
			}
			var sessionID *string
			if sid, ok := ctx.Value(ContextKeySessionId{}).(string); ok && sid != "" {
				sessionID = &sid
			}

			if userID == nil && sessionID == nil {
				ctx = context.WithValue(ctx, WorkingTripIDKey, tripID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			forkedTrip, err := store.GetFork(ctx, tripID, userID, sessionID)

			if err == nil {
				ctx = context.WithValue(ctx, WorkingTripIDKey, forkedTrip.ID.String())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			if err != pgx.ErrNoRows {
				http.Error(w, "Database error checking fork", http.StatusInternalServerError)
				return
			}

			isOwner, isDemo, err := store.GetTripMeta(ctx, tripID, userID, sessionID)

			if err != nil {
				if err == pgx.ErrNoRows {
					http.Error(w, "Trip not found", http.StatusNotFound)
					return
				}
				http.Error(w, "Database error checking trip owner", http.StatusInternalServerError)
				return
			}

			if isOwner || !isDemo {
				ctx = context.WithValue(ctx, WorkingTripIDKey, tripID)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			isMutating := r.Method == "POST" || r.Method == "PUT" || r.Method == "DELETE"

			if isMutating {
				forkedTrip, err = store.ForkTrip(ctx, tripID, userID, sessionID)
				if err != nil {
					http.Error(w, "Failed to fork demo trip", http.StatusInternalServerError)
					return
				}

				ctx = context.WithValue(ctx, WorkingTripIDKey, forkedTrip.ID.String())
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			ctx = context.WithValue(ctx, WorkingTripIDKey, tripID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetWorkingTripID helper
func GetWorkingTripID(r *http.Request) string {
	if id, ok := r.Context().Value(WorkingTripIDKey).(string); ok && id != "" {
		return id
	}
	
	// Fallbacks just in case the middleware was bypassed somehow
	if id := chi.URLParam(r, "trip_id"); id != "" {
		return id
	}
	return chi.URLParam(r, "id")
}
