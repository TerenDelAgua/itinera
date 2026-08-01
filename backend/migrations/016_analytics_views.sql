-- 016_analytics_views.sql
-- Defines the canonical SQL views used by the analytics module.
-- Encapsulates "filter internal trips" so no endpoint has to remember it.
--
-- These views are the source of truth for the concept of "real user"
-- and "real trip". Any future query that wants to compute analytics over
-- real activity MUST go through these views (or replicate their filter).
--
-- Non-breaking: views are derived from existing tables. No data is moved.

-- Guard schema: verify preconditions hold. We log warnings, not errors,
-- so a partially-migrated DB still progresses.
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'trips'
    ) THEN
        RAISE WARNING 'trips table does not exist; skipping 016_analytics_views';
        RETURN;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_schema = 'public' AND table_name = 'trips' AND column_name = 'is_internal'
    ) THEN
        RAISE WARNING 'trips.is_internal column missing; migrations 015 must run before 016';
        RETURN;
    END IF;
END$$;

-- ─────────────────────────────────────────────────────────────────────────
-- 1. v_real_sessions
--    One row per session_id that has at least ONE non-internal trip.
--    Excludes any session_id that ever generated an internal trip, because
--    that session was "contaminated" by the dev (e.g. cookie activated
--    mid-navigation). Trade-off: under-counting beats over-counting.
-- ─────────────────────────────────────────────────────────────────────────
CREATE OR REPLACE VIEW v_real_sessions AS
SELECT DISTINCT session_id
FROM trips
WHERE is_internal = false
  AND session_id IS NOT NULL;

-- ─────────────────────────────────────────────────────────────────────────
-- 2. v_real_trips
--    Every trip from a real user, with aggregated metrics (spend, places,
--    expenses) pre-computed so the analytics handlers do not need to
--    repeat the subqueries.
-- ─────────────────────────────────────────────────────────────────────────
CREATE OR REPLACE VIEW v_real_trips AS
SELECT
    t.id,
    t.user_id,
    t.session_id,
    t.name,
    t.base_currency,
    t.created_at,
    t.forked_from,
    COALESCE((SELECT SUM(amount) FROM expenses WHERE trip_id = t.id), 0) AS total_spent,
    COALESCE((SELECT COUNT(*) FROM places WHERE trip_id = t.id), 0)     AS place_count,
    COALESCE((SELECT COUNT(*) FROM expenses WHERE trip_id = t.id), 0)   AS expense_count
FROM trips t
WHERE t.is_internal = false;

-- ─────────────────────────────────────────────────────────────────────────
-- 3. v_user_engagement
--    One row per session_id with aggregate usage metrics. The view that
--    answers "who actually uses the app deeply?" (is_power_user flag).
-- ─────────────────────────────────────────────────────────────────────────
CREATE OR REPLACE VIEW v_user_engagement AS
SELECT
    t.session_id,
    COUNT(DISTINCT t.id)                                    AS trip_count,
    COALESCE(SUM(e_sub.cnt), 0)                             AS total_expenses,
    COALESCE(SUM(p_sub.cnt), 0)                             AS total_places,
    MIN(t.created_at)                                       AS first_trip_at,
    MAX(t.created_at)                                       AS last_trip_at,
    -- "Power user" = at least 1 trip + at least 3 expenses + active on
    -- at least 2 distinct days. The thresholds are documented in spec 015 v2
    -- section 5.1 and can be tuned later without schema changes.
    (COUNT(DISTINCT t.id) >= 1
     AND COALESCE(SUM(e_sub.cnt), 0) >= 3
     AND COUNT(DISTINCT DATE(t.created_at)) >= 2)           AS is_power_user
FROM trips t
LEFT JOIN (
    SELECT trip_id, COUNT(*) AS cnt
    FROM expenses
    GROUP BY trip_id
) e_sub ON e_sub.trip_id = t.id
LEFT JOIN (
    SELECT trip_id, COUNT(*) AS cnt
    FROM places
    GROUP BY trip_id
) p_sub ON p_sub.trip_id = t.id
WHERE t.is_internal = false
  AND t.session_id IS NOT NULL
GROUP BY t.session_id;
