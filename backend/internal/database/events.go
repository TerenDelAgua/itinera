package database

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// EventRepository persists analytics events emitted by the client.
//
// Insert is fire-and-forget safe: callers should not propagate errors to the
// HTTP response, only log them internally. Keeping the failure mode here lets
// us swap the backing store (a future Kafka/ClickHouse pipeline) without
// touching the handler.
type EventRepository struct {
	Pool *pgxpool.Pool
}

func NewEventRepository(pool *pgxpool.Pool) *EventRepository {
	return &EventRepository{Pool: pool}
}

func (r *EventRepository) Insert(
	ctx context.Context,
	eventType, sessionID string,
	userID *string,
	tripID *string,
	metadataJSON []byte,
	createdAt time.Time,
) error {
	const query = `
		INSERT INTO events (type, session_id, user_id, trip_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.Pool.Exec(ctx, query, eventType, sessionID, userID, tripID, metadataJSON, createdAt)
	return err
}
