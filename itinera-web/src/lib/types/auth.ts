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
 * Inputs accepted by `POST /auth/v2/register` (mounted under the
 * Vite-configured `VITE_API_URL`, which is `/api/v1`). `locale`
 * is optional
 * and falls back to the i18n store's current locale.
 */
export interface RegisterInput {
    email: string;
    password: string;
    locale?: string;
    terms_accepted?: boolean;
}

/**
 * Inputs accepted by `POST /auth/v2/forgot` (Spec 017 §5.7).
 * Anti-enumeration: the backend ALWAYS returns 202 with the same
 * generic message, regardless of whether the email exists or not.
 * The frontend must reflect this — never show "user not found".
 *
 * `locale` is forwarded so the reset email is localised.
 */
export interface ForgotInput {
    email: string;
    locale?: string;
}

/**
 * Inputs accepted by `POST /auth/v2/reset` (Spec 017 §5.8).
 *
 * `code` is the 6-digit numeric string the user received by email
 * (leading zeros preserved — the backend zero-pads the input
 * before SHA-256 hashing). `new_password` is validated against the
 * same 3-rule policy as register.
 */
export interface ResetInput {
    email: string;
    code: string;
    new_password: string;
    locale?: string;
}

/**
 * Response of `POST /auth/v2/forgot`. 202 Accepted with a generic
 * acknowledgement. The server's actual outcome (code sent /
 * user doesn't exist) is hidden from the caller.
 */
export interface ForgotResponse {
    message: string;
}
