ALTER TABLE trips
    ADD COLUMN IF NOT EXISTS share_token VARCHAR(16) UNIQUE,
    ADD COLUMN IF NOT EXISTS share_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS share_expires_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS share_created_at TIMESTAMPTZ;

CREATE UNIQUE INDEX IF NOT EXISTS idx_trips_share_token
    ON trips(share_token) WHERE share_enabled = true;