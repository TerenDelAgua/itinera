-- Purpose: Add default expense currency to trips and places
-- trips: required, defaults to EUR, migrated from base_currency
-- places: optional (nullable), NULL means inherit from trip

BEGIN;

ALTER TABLE trips 
    ADD COLUMN IF NOT EXISTS default_expense_currency VARCHAR(3) NOT NULL DEFAULT 'EUR';

-- Migrate existing trips to use their base_currency as the default expense currency
UPDATE trips 
SET default_expense_currency = base_currency;

ALTER TABLE places 
    ADD COLUMN IF NOT EXISTS default_expense_currency VARCHAR(3) NULL;

-- Documentation note: NULL in places.default_expense_currency means it inherits from trips.default_expense_currency

COMMIT;
