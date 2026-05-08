ALTER TABLE public.trips 
ADD COLUMN IF NOT EXISTS is_public_demo BOOLEAN NOT NULL DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_trips_public_demo ON public.trips(is_public_demo) 
WHERE is_public_demo = TRUE;

-- 3. Events (tracking interno, privacy-first)
-- Don't store PII. session_id is hash, not identifiable.
CREATE TABLE IF NOT EXISTS public.events (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    type VARCHAR(50) NOT NULL,           -- 'landing_view', 'trip_created', 'demo_viewed', etc.
    session_id VARCHAR(64),                -- Cookie-based, rotación diaria
    user_id UUID REFERENCES public.users(id) ON DELETE SET NULL,
    trip_id UUID REFERENCES public.trips(id) ON DELETE SET NULL,
    metadata JSONB DEFAULT '{}',           -- {country: 'JP', device: 'mobile', source: 'demo'}
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Índices para queries de contador y cohortes
CREATE INDEX IF NOT EXISTS idx_events_type_created ON public.events(type, created_at);
CREATE INDEX IF NOT EXISTS idx_events_session_created ON public.events(session_id, created_at);

-- 4. Tabla de stats públicas (materializada, evita COUNT(*) en landing)
CREATE TABLE IF NOT EXISTS public.public_stats (
    id INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1), -- Singleton table
    total_trips_count INTEGER NOT NULL DEFAULT 0,
    last_updated TIMESTAMPTZ DEFAULT NOW()
);

-- Insertar registro inicial
INSERT INTO public.public_stats (id, total_trips_count) 
VALUES (1, 0) 
ON CONFLICT (id) DO NOTHING;