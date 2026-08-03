BEGIN;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'users'
    ) THEN
        RAISE WARNING 'users table does not exists; skipping 017_auth_mvp';
        RETURN;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'trips'
    ) THEN
        RAISE WARNING 'trips table does not exists; skipping 017_auth_mvp';
        RETURN;
    END IF;
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'rate_limits'
    ) THEN
        RAISE WARNING 'rate_limits table does not exists; skipping 017_auth_mvp';
        RETURN;
    END IF;
END $$;

-- =====================================================================
-- 1. Extender users con tier, terms_accepted_at, locale, soft delete
-- =====================================================================
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS tier TEXT NOT NULL DEFAULT 'free'
        CHECK (tier IN ('free', 'explorer', 'premium')),
    ADD COLUMN IF NOT EXISTS terms_accepted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS locale CHAR(2) NOT NULL DEFAULT 'en'
        CHECK (locale IN ('en', 'es', 'ja', 'id')),
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

-- Email must be normalize (lowercase, trim) on app.
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_key;
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_email_unique
    ON users (LOWER(email))
    WHERE deleted_at IS NULL;


-- =====================================================================
-- 2. Tabla sessions (tokens opacos, rotación, reuse detection)
-- =====================================================================
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- Hash SHA-256
    access_token_hash TEXT NOT NULL UNIQUE,
    -- Hash SHA-256
    refresh_token_hash TEXT NOT NULL UNIQUE,
    -- Refresh token family if refresh token rotate and the old is reused,
    -- revoke the whole family (logout from all devices)
    refresh_family UUID NOT NULL,
    refresh_rotated_at TIMESTAMPTZ,
    user_agent TEXT,
    ip_address INET,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_used_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL, --refresh expiry
    revoked_at TIMESTAMPTZ
);

-- lookup for user to list avtive session and logout all
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
-- Reuse detection: look the family when detect a reuse refresh
CREATE INDEX IF NOT EXISTS idx_sessions_refresh_family ON sessions (refresh_family);
-- Expire sessions cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at 
    ON sessions (expires_at)
    WHERE revoked_at IS NULL;

-- =====================================================================
-- 3. Tabla password_reset_tokens
-- =====================================================================
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    -- Hash SHA-256 of the 6 digits code send by email
    token_hash TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ,
    -- fail attempts counter
    attempts INT NOT NULL DEFAULT 0,
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ip_address INET
);

CREATE INDEX IF NOT EXISTS idx_password_reset_user_id 
    ON password_reset_tokens (user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_expires
    ON password_reset_tokens (expires_at)
    WHERE used_at IS NULL AND locked_at IS NULL;

-- =====================================================================
-- 4. Soft-delete column en trips (necesario para §5.6 claim-guest filter
--    y para §5.9 account delete cascade)
-- =====================================================================
ALTER TABLE trips
    ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

CREATE INDEX IF NOT EXISTS idx_trips_deleted_at
    ON trips(deleted_at) WHERE deleted_at IS NOT NULL;

-- =====================================================================
-- 5. Asegurar que trips.user_id está indexado (puede que ya lo esté)
-- =====================================================================
CREATE INDEX IF NOT EXISTS idx_trips_user_id
    ON trips(user_id) WHERE user_id IS NOT NULL;

COMMIT;