-- Migration: 010_update_public_stats.sql

-- Función para actualizar public_stats cuando se registra un trip_created
CREATE OR REPLACE FUNCTION update_public_stats_travelers()
RETURNS TRIGGER AS $$
BEGIN
    IF NEW.type IN ('trip_created', 'demo_deep_forked', 'demo_fork_reused') THEN
        UPDATE public.public_stats 
        SET 
            total_trips_count = (
                SELECT COUNT(DISTINCT session_id) 
                FROM public.events 
                WHERE type IN ('trip_created', 'demo_deep_forked', 'demo_fork_reused')
                AND created_at >= '2026-05-9 00:00:00+00'
            ),
            last_updated = NOW()
        WHERE id = 1;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Crear el trigger en la tabla events
DROP TRIGGER IF EXISTS trigger_update_public_stats ON public.events;

CREATE TRIGGER trigger_update_public_stats
AFTER INSERT ON public.events
FOR EACH ROW
EXECUTE FUNCTION update_public_stats_travelers();

-- Actualizar los datos actuales con la nueva lógica (si es 0, dejamos 0, de lo contrario usamos la cuenta)
UPDATE public.public_stats 
SET 
    total_trips_count = COALESCE(
        (SELECT NULLIF(COUNT(DISTINCT session_id), 0) 
         FROM public.events 
         WHERE type IN ('trip_created', 'demo_deep_forked', 'demo_fork_reused')
         AND created_at >= '2026-05-9 00:00:00+00'
        ), 
        total_trips_count
    ),
    last_updated = NOW()
WHERE id = 1;
