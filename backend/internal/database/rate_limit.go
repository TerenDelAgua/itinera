package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RateLimitRepository implements a fixed-window counter backed by PostgreSQL.
//
// Each row in the `rate_limits` table represents one active window. The
// single UPSERT in CheckAndIncrement performs the read, decide, write and
// return steps in one round-trip, so the limiter scales across replicas
// without distributed locks.
type RateLimitRepository struct {
	Pool *pgxpool.Pool
}

func NewRateLimitRepository(pool *pgxpool.Pool) *RateLimitRepository {
	return &RateLimitRepository{Pool: pool}
}

// CheckAndIncrement atomically advances the counter for the given key and
// returns whether the call is allowed.
//
// Behaviour:
//   - If the row is missing or its window_start is older than `window`,
//     the counter is reset to 1 and the call is allowed.
//   - Otherwise the counter is incremented by 1 and the call is allowed
//     when the new value is <= `max`.
//
// Errors are returned to the caller. The events handler logs and allows
// the event on error to preserve the fire-and-forget UX — but tests and
// other consumers can fail loudly if they prefer.
func (r *RateLimitRepository) CheckAndIncrement(
	ctx context.Context,
	key string,
	window time.Duration,
	max int,
) (bool, error) {
	if max <= 0 {
		return false, fmt.Errorf("rate limit max must be positive, got %d", max)
	}
	if window <= 0 {
		return false, fmt.Errorf("rate limit window must be positive, got %s", window)
	}

	const query = `
		INSERT INTO public.rate_limits (rate_key, count, window_start)
		VALUES ($1, 1, NOW())
		ON CONFLICT (rate_key) DO UPDATE
		SET count = CASE
				WHEN rate_limits.window_start < NOW() - $2::interval
					THEN 1
				ELSE rate_limits.count + 1
			END,
			window_start = CASE
				WHEN rate_limits.window_start < NOW() - $2::interval
					THEN NOW()
				ELSE rate_limits.window_start
			END
		RETURNING count
	`

	var count int
	if err := r.Pool.QueryRow(ctx, query, key, intervalString(window)).Scan(&count); err != nil {
		return false, fmt.Errorf("rate limit upsert: %w", err)
	}

	return count <= max, nil
}

// intervalString formats a Go duration into the PostgreSQL INTERVAL literal
// expected by the UPSERT (e.g. "1 minute", "30 seconds"). Truncating to
// seconds is fine for the windows we use (>= 1 second).
func intervalString(d time.Duration) string {
	seconds := int64(d / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("%d seconds", seconds)
}
