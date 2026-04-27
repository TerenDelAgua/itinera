-- Migration: 004_activities.sql
-- Purpose: Agenda system for Itinera. Trip-level and Place-level activities.

DROP TABLE IF EXISTS activities ;

CREATE TABLE activities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trip_id UUID NOT NULL REFERENCES trips(id) ON DELETE CASCADE,
    place_id UUID REFERENCES places(id) ON DELETE SET NULL, -- NULL = trip-level activity
    title VARCHAR(255) NOT NULL,
    date DATE NOT NULL,
    time TIME, -- Optional time (NULL = all-day or unspecified time)
    notes TEXT DEFAULT '',
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_activities_trip_id ON activities(trip_id);
CREATE INDEX idx_activities_trip_timeline ON activities(trip_id, date ASC, time ASC NULLS LAST);
CREATE INDEX idx_activities_place_id ON activities(place_id);

COMMENT ON TABLE activities IS 'Agente items. Flexible hierarchy: trip-level or place-specific';
COMMENT ON COLUMN activities.place_id IS 'NULL indicates a global/trip-level activity. ON DELETE SET NULL preserves activities if a place is removed';
COMMENT ON COLUMN activities.time IS 'Local time only (MVP). NULL means no specifict time';