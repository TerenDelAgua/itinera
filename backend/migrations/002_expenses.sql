CREATE TABLE IF NOT EXISTS expense_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(50) UNIQUE NOT NULL,
    color_hex VARCHAR(7) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Seed inicial (seguro y idempotente)
INSERT INTO expense_categories (slug, color_hex) VALUES
('accommodation', '#3B82F6'),
('transport', '#10B981'),
('food', '#F59E0B'),
('leisure', '#8B5CF6'),
('shopping', '#EC4899'),
('others', '#6B7280')
ON CONFLICT (slug) DO NOTHING;

ALTER TABLE expenses 
ADD COLUMN IF NOT EXISTS category_id UUID REFERENCES expense_categories(id),
ADD COLUMN IF NOT EXISTS notes TEXT;
ALTER TABLE expenses DROP COLUMN IF EXISTS category;
CREATE INDEX IF NOT EXISTS idx_expenses_category ON expenses(category_id);
CREATE INDEX IF NOT EXISTS idx_expenses_trip_date ON expenses(trip_id, date DESC);