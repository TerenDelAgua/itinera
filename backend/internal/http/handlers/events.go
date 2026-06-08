package handlers

import (
	"backend/internal/http/middleware"
	"backend/internal/tracking"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const (
	eventRateLimitWindow = time.Minute
	eventRateLimitMax    = 60
)

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
func (h *Handlers) TrackEvent(w http.ResponseWriter, r *http.Request) {
	// Resolve session from middleware context (set by SessionMiddleware)
	sessionID := ""
	if sid, ok := r.Context().Value(middleware.ContextKeySessionId{}).(string); ok {
		sessionID = sid
	}

	// Rate limit: max 60 events/min per session, enforced globally via the
	// RateLimitStore. Errors here are logged and treated as "allow" so a
	// transient DB hiccup does not silently drop the event.
	if h.RateLimitRepo != nil {
		allowed, err := h.RateLimitRepo.CheckAndIncrement(r.Context(), "events:"+sessionID, eventRateLimitWindow, eventRateLimitMax)
		if err != nil {
			log.Printf("⚠️ [TrackEvent] rate limiter error, allowing event: %v", err)
		} else if !allowed {
			// Caller is over the cap. We respond 204 anyway (fire-and-forget
			// UX), but skip the INSERT to keep the events table small.
			w.WriteHeader(http.StatusNoContent)
			return
		}
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

	// Resolve user_id from JWT context if present (optional). The middleware
	// stores a uuid.UUID; we surface it as a pointer so the EventStore can
	// distinguish "no authenticated user" (NULL) from "user_id is empty".
	var userID *string
	if uid, ok := r.Context().Value(middleware.ContextKeyUserId{}).(uuid.UUID); ok {
		s := uid.String()
		userID = &s
	}

	if err := h.EventsRepo.Insert(r.Context(), req.Type, sessionID, userID, req.TripID, metadataJSON, time.Now().UTC()); err != nil {
		// Fire-and-forget: log internally, never break UX
		log.Printf("⚠️ [TrackEvent] Failed to insert event '%s': %v", req.Type, err)
	}

	w.WriteHeader(http.StatusNoContent)
}
