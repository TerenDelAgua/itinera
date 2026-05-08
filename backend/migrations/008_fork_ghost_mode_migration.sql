-- ==========================================
-- TEREN Itinera - Schema Migration 008
-- Feature: Fork-On-Write + Ghost Mode
-- ==========================================

BEGIN;

-- 1. Add forked_from column to trips
ALTER TABLE public.trips 
ADD COLUMN IF NOT EXISTS forked_from UUID REFERENCES public.trips(id) ON DELETE SET NULL;

-- 2. Performance indexes for search (session_id, public_demo)
CREATE INDEX IF NOT EXISTS idx_trips_forked_from ON public.trips(forked_from);
CREATE INDEX IF NOT EXISTS idx_trips_session_id ON public.trips(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_trips_public_demo ON public.trips(is_public_demo) WHERE is_public_demo = TRUE;

-- 3. Internal tracking table (Analytics)
CREATE TABLE IF NOT EXISTS public.event_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_name VARCHAR(50) NOT NULL,
    session_id VARCHAR(255),
    user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    trip_id UUID REFERENCES public.trips(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- 4. Performance indexes for analytics metrics
CREATE INDEX IF NOT EXISTS idx_event_logs_name_time ON public.event_logs(event_name, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_event_logs_session_id ON public.event_logs(session_id) WHERE session_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_event_logs_user_id ON public.event_logs(user_id) WHERE user_id IS NOT NULL;

-- 5. RLS for event_logs
ALTER TABLE public.event_logs ENABLE ROW LEVEL SECURITY;

CREATE POLICY "Backend can insert events"
    ON public.event_logs FOR INSERT
    WITH CHECK (true);

CREATE POLICY "Backend can read events"
    ON public.event_logs FOR SELECT
    USING (true);

COMMIT;