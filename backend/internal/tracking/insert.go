package tracking

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// InsertEvent is a fire-and-forget helper for internal server-side event tracking
// (e.g. from ResolveTripContext middleware). It is NOT the HTTP handler — that lives
// in handlers/events.go and is called directly by the frontend.
func InsertEvent(
	ctx context.Context,
	pool *pgxpool.Pool,
	eventType string,
	sessionID string,
	userID *string,
	tripID *string,
	metadata map[string]interface{},
) error {
	cleanMeta := SanitizeMetadata(metadata)

	if _, ok := cleanMeta["server.timestamp"]; !ok {
		cleanMeta["server.timestamp"] = time.Now().Unix()
	}

	metadataJSON, err := json.Marshal(cleanMeta)
	if err != nil {
		return err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO events (type, session_id, user_id, trip_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
	`, eventType, sessionID, userID, tripID, metadataJSON)

	if err != nil {
		log.Printf("⚠️ [tracking.InsertEvent] Failed to insert event '%s': %v", eventType, err)
	}

	return err
}