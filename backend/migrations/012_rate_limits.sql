-- ==========================================
-- TEREN Itinera - Schema Migration 012
-- Feature: DB-backed rate limiter
-- ==========================================
--
-- Replaces the in-memory eventRateLimiter in internal/http/handlers/events.go.
-- The previous implementation lived in a package-level map, which meant the
-- 60 events/min cap was enforced per-process: a multi-replica deploy
-- multiplied the effective limit by the replica count.
--
-- By storing the counter in PostgreSQL we make the cap global and atomic.
-- A future migration can introduce periodic cleanup of stale rows.

BEGIN;

CREATE TABLE IF NOT EXISTS public.rate_limits (
    -- Caller-supplied bucket key. For the events endpoint this is the
    -- session id; future limiters (login attempts, etc.) can use the same
    -- table with a namespaced key (e.g. "login:email@example.com").
    rate_key       TEXT        PRIMARY KEY,
    -- Number of operations consumed inside the current window.
    count          INTEGER     NOT NULL DEFAULT 0,
    -- Start of the current sliding window. We compare against NOW() in the
    -- UPSERT to decide whether to reset the count.
    window_start   TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- A row is "expired" once its window is older than the longest limit window
-- we enforce (currently 1 minute for events). The conditional UPSERT handles
-- the reset, so this index is mostly used by the optional cleanup query.
CREATE INDEX IF NOT EXISTS idx_rate_limits_window_start
    ON public.rate_limits (window_start);

-- Backend-only table: no RLS because rate limit counters are an internal
-- concern, never exposed to end-users.
ALTER TABLE public.rate_limits ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Backend can manage rate limits"
    ON public.rate_limits FOR ALL
    USING (true)
    WITH CHECK (true);

COMMIT;
