-- Purpose: Support multi-currency with offline caching

BEGIN;
ALTER TABLE expenses
    ADD COLUMN IF NOT EXISTS original_amount DECIMAL(12, 2) NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS original_currency VARCHAR(3) NOT NULL DEFAULT 'EUR',
    ADD COLUMN IF NOT EXISTS exchange_rate DECIMAL(10, 6) NOT NULL DEFAULT 1.0,
    ADD COLUMN IF NOT EXISTS conversion_date DATE NOT NULL DEFAULT CURRENT_DATE;

UPDATE expenses
SET original_amount = amount,
    original_currency = COALESCE((SELECT base_currency FROM trips WHERE id = trip_id LIMIT 1), 'EUR'),
    exchange_rate = 1.0,
    conversion_date = CURRENT_DATE;
CREATE TABLE IF NOT EXISTS exchange_rates_cache (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    base_currency VARCHAR(3) NOT NULL,
    target_currency VARCHAR(3) NOT NULL,
    rate DECIMAL(10,6) NOT NULL,
    fetched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(base_currency, target_currency)
);

CREATE INDEX IF NOT EXISTS idx_exchange_rates_lookup ON exchange_rates_cache(base_currency, target_currency);

COMMIT;