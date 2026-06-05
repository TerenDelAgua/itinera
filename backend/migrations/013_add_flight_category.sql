-- Migration 013: Add flight to expense_categories table
INSERT INTO expense_categories (slug, color_hex)
VALUES ('flight', '#06B6D4')
ON CONFLICT (slug) DO NOTHING;
