

ALTER TABLE places RENAME COLUMN description TO Notes;
ALTER TABLE places DROP COLUMN IF EXISTS order_index;

ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS place_id UUID REFERENCES places(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS notes    TEXT DEFAULT '';

-- Índices optimizados
CREATE INDEX IF NOT EXISTS idx_places_trip_date ON places(trip_id, start_date ASC NULLS LAST);
CREATE INDEX IF NOT EXISTS idx_expenses_place_id ON expenses(place_id);
CREATE INDEX IF NOT EXISTS idx_expenses_trip_amount ON expenses(trip_id, amount) WHERE place_id IS NULL;