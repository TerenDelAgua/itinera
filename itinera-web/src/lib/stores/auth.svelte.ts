/**
 * Auth store — single source of truth for the current user's session.
 *
 * Uses Svelte 5 runes (`$state`, `$derived`) so subscribers in components
 * react automatically without explicit subscribe() / set() plumbing.
 *
 * Lifecycle contract:
 *
 *   - On boot, components should call `auth.bootstrap()` (idempotent)
 *     to populate `user` from `/auth/v2/me` using the existing HttpOnly
 *     cookies. This avoids the flash of "logged out" content for users
 *     who already have a session.
 *   - Every public mutator (`register`, `login`, `logout`) sets `loading`
 *     to `true` on entry and clears it on exit so buttons can render a
 *     spinner. `lastError` is reset at the start and set on failure.
 *   - `lastError` is always the most recent `ApiError`; components that
 *     want to map it to a localised message should branch on `code`,
 *     never on `message`.
 *
 * Test isolation: every call site uses the named export `auth` so
 * tests can mutate it directly (`auth.user = null`) between cases.
 */
import { apiFetch, ApiError } from '$lib/api';
import type { User, ForgotInput, ForgotResponse, ResetInput } from '$lib/types/auth';

interface LoginInput {
    email: string;
    password: string;
}

interface RegisterInput {
    email: string;
    password: string;
    locale?: string;
}

interface AuthMeResponse {
    user: User | null;
}

class AuthStore {
    user = $state<User | null>(null);
    loading = $state(false);
    /**
     * `true` while the boot-time `/auth/v2/me` probe is in flight.
     * The UserMenu uses this to render a 32px skeleton instead of
     * "Sign in" / Avatar — eliminates the layout-shift flash during
     * the 200-500ms between the SSR HTML and the boot probe
     * resolving.
     */
    isLoading = $state(false);
    lastError = $state<ApiError | null>(null);
    /**
     * In-flight bootstrap promise (single-flight). `null` when no
     * probe is running. Not part of the public API.
     */
    _bootstrapInflight: Promise<void> | null = null;

    isLoggedIn = $derived(this.user !== null);

    /**
     * Bootstrap the session from cookies. Safe to call multiple times;
     * subsequent calls return immediately without re-fetching.
     */
    async bootstrap(): Promise<void> {
        // Single-flight: if a bootstrap is already running, return the
        // existing promise instead of issuing a second `/me` request.
        if (this._bootstrapInflight) return this._bootstrapInflight;
        if (this.user) return;
        this.isLoading = true;
        this._bootstrapInflight = (async () => {
            try {
                const data = await apiFetch<AuthMeResponse>('/auth/v2/me');
                this.user = data.user;
            } catch (e) {
                if (e instanceof ApiError && e.status === 401) {
                    // Expected for guests: leave `user` null, no error.
                    this.user = null;
                    return;
                }
                // Network or server down: surface as lastError but keep the
                // guest state so the app remains usable.
                if (e instanceof ApiError) {
                    this.lastError = e;
                }
            } finally {
                this.isLoading = false;
                this._bootstrapInflight = null;
            }
        })();
        return this._bootstrapInflight;
    }

    async register(input: RegisterInput): Promise<User> {
        this.loading = true;
        this.lastError = null;
        try {
            const data = await apiFetch<{ user: User }>('/auth/v2/register', {
                method: 'POST',
                body: {
                    email: input.email,
                    password: input.password,
                    locale: input.locale
                }
            });
            this.user = data.user;
            return data.user;
        } catch (e) {
            if (e instanceof ApiError) {
                this.lastError = e;
            }
            throw e;
        } finally {
            this.loading = false;
        }
    }

    async login(input: LoginInput): Promise<User> {
        this.loading = true;
        this.lastError = null;
        try {
            const data = await apiFetch<{ user: User }>('/auth/v2/login', {
                method: 'POST',
                body: { email: input.email, password: input.password }
            });
            this.user = data.user;
            return data.user;
        } catch (e) {
            if (e instanceof ApiError) {
                this.lastError = e;
            }
            throw e;
        } finally {
            this.loading = false;
        }
    }

    /**
     * Idempotent server-side logout (Spec 017 §5.3). We always clear
     * the local user state even if the request fails, so the UI stays
     * consistent for users on flaky networks.
     */
    async logout(): Promise<void> {
        this.lastError = null;
        this.user = null;
        try {
            await apiFetch<void>('/auth/v2/logout', { method: 'POST' });
        } catch (e) {
            if (e instanceof ApiError && e.status >= 500) {
                console.warn('Logout failed on server; local state cleared but server session may still be active in other tabs.');
            }
        }
    }

    /**
     * Password reset: request a 6-digit code (Spec 017 §5.7).
     *
     * Anti-enumeration: the backend ALWAYS responds 202 with the
     * same generic message. From the user's perspective the call
     * succeeds whether the email exists or not. We propagate
     * validation errors (empty email, malformed body, RATE_LIMITED)
     * but otherwise just resolve.
     */
    async forgotPassword(input: ForgotInput): Promise<ForgotResponse> {
        this.lastError = null;
        try {
            return await apiFetch<ForgotResponse>('/auth/v2/forgot', {
                method: 'POST',
                body: { email: input.email, locale: input.locale }
            });
        } catch (e) {
            if (e instanceof ApiError) this.lastError = e;
            throw e;
        }
    }

    /**
     * Password reset: consume the 6-digit code and set a new password
     * (Spec 017 §5.8). On success the backend revokes every existing
     * session, so the user must sign back in. We don't auto-login:
     * redirecting to /login keeps the success state clean.
     */
    async resetPassword(input: ResetInput): Promise<void> {
        this.lastError = null;
        try {
            await apiFetch<void>('/auth/v2/reset', {
                method: 'POST',
                body: {
                    email: input.email,
                    code: input.code,
                    new_password: input.new_password,
                    locale: input.locale
                }
            });
            // The reset revokes all sessions, including this client's
            // cookie. Clear local user so subsequent guards don't
            // mis-route the user.
            this.user = null;
        } catch (e) {
            if (e instanceof ApiError) this.lastError = e;
            throw e;
        }
    }

    /**
     * GDPR account deletion (Spec 017 §5.9). Idempotent: deleting an
     * already-deleted account still resolves successfully because
     * the backend's `SoftDeleteUserCascade` returns `pgx.ErrNoRows`
     * for soft-deleted users, which the handler treats as 204.
     *
     * On success we wipe local state and the server clears the
     * access/refresh cookies. The user is redirected to / from the
     * caller (the delete-account page), not from here, so the
     * caller controls the success UX.
     */
    async deleteAccount(): Promise<void> {
        this.lastError = null;
        try {
            await apiFetch<void>('/auth/v2/account', { method: 'DELETE' });
            this.user = null;
        } catch (e) {
            if (e instanceof ApiError) this.lastError = e;
            throw e;
        }
    }

    /** Manually clear the last error, e.g. after the user dismisses a banner. */
    clearError(): void {
        this.lastError = null;
    }
}

export const auth = new AuthStore();
export type { LoginInput, RegisterInput };
