package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/tracking"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ── Rate Limiter ─────────────────────────────────────────────────────────────

type eventRateLimiter struct {
	mu     sync.Mutex
	counts map[string]int
	window map[string]time.Time
}

var evtLimiter = &eventRateLimiter{
	counts: make(map[string]int),
	window: make(map[string]time.Time),
}

func (rl *eventRateLimiter) allow(sessionID string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	start, exists := rl.window[sessionID]

	if !exists || now.Sub(start) > time.Minute {
		rl.window[sessionID] = now
		rl.counts[sessionID] = 1
		return true
	}

	if rl.counts[sessionID] >= 60 {
		return false
	}

	rl.counts[sessionID]++
	return true
}

// ── Request Shape ─────────────────────────────────────────────────────────────

type EventRequest struct {
	Type     string                 `json:"type"`
	TripID   *string                `json:"trip_id,omitempty"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// ── Handler ───────────────────────────────────────────────────────────────────

// TrackEvent godoc
// @Summary      Track an analytics event
// @Description  Fire-and-forget endpoint for privacy-first analytics (no PII, no IP)
// @Tags         events
// @Accept       json
// @Param        event  body  EventRequest  true  "Event payload"
// @Success      204
// @Failure      400  {string}  string "Bad request"
// @Router       /events [post]
func TrackEvent(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Resolve session from middleware context (set by SessionMiddleware)
		sessionID := ""
		if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
			sessionID = sid
		}

		// Rate limit: max 60 events/min per session
		if !evtLimiter.allow(sessionID) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		var req EventRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !tracking.IsValid(req.Type) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Sanitize and add server-side metadata
		req.Metadata = tracking.SanitizeMetadata(req.Metadata)
		req.Metadata["server.timestamp"] = time.Now().Unix()

		metadataJSON, err := json.Marshal(req.Metadata)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Resolve user_id from JWT context if present (optional)
		var userID interface{}
		if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(interface{ String() string }); ok {
			userID = uid.String()
		}

		_, err = pool.Exec(r.Context(), `
			INSERT INTO events (type, session_id, user_id, trip_id, metadata, created_at)
			VALUES ($1, $2, $3, $4, $5, NOW())
		`, req.Type, sessionID, userID, req.TripID, metadataJSON)

		if err != nil {
			// Fire-and-forget: log internally, never break UX
			log.Printf("⚠️ [TrackEvent] Failed to insert event '%s': %v", req.Type, err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}