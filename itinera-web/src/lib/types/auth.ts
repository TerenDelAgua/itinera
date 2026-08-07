/**
 * Auth-related types shared between stores, components, and tests.
 *
 * Mirrors the backend's `models.User` (internal/database/auth.go).
 * Keep field names in sync — when the backend renames a column, this
 * file is the one place that catches the divergence at typecheck.
 */
export interface User {
    id: string;
    email: string;
    tier: 'free' | 'pro';
    locale: string;
    terms_accepted_at: string | null;
    created_at: string;
}

/**
 * The shape of the registration response (Spec 017 §5.1). On success
 * the backend issues HttpOnly cookies and returns the user payload
 * alongside token metadata.
 */
export interface RegisterResponse {
    user: User;
    access_token: string;
    refresh_token: string;
    token_type: 'Bearer';
    expires_in: number;
}

/**
 * Inputs accepted by `/api/v1/auth/v2/register`. `locale` is optional
 * and falls back to the i18n store's current locale.
 */
export interface RegisterInput {
    email: string;
    password: string;
    locale?: string;
    terms_accepted?: boolean;
}
