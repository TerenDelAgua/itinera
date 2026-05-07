
-- partial indexes for fast analytics queries (Fase 0)

-- 1. public counter (trip_created)
CREATE INDEX IF NOT EXISTS idx_events_type_created 
ON public.events(type, created_at) 
WHERE type = 'trip_created';

-- 2. Cohort analysis & Funnels
CREATE INDEX IF NOT EXISTS idx_events_session_created_type 
ON public.events(session_id, created_at, type);

-- 3. Tracking de Demos (Fork system)
CREATE INDEX IF NOT EXISTS idx_events_demo_type 
ON public.events(type, created_at) 
WHERE type IN ('demo_viewed', 'demo_deep_forked', 'demo_fork_reused');

-- 4. Landing tracking
CREATE INDEX IF NOT EXISTS idx_events_landing 
ON public.events(type, created_at) 
WHERE type = 'landing_viewed';