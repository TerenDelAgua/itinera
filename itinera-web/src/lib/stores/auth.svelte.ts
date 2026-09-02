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
    // Spec 018 §7: forwarded to /auth/v2/register so the backend can
    // pin the accepted legal-doc versions to the user account. The
    // canonical type lives in `$lib/types/auth.ts`; we re-declare the
    // optional version fields here because the store used to take a
    // narrower shape and the tests in `register.test.ts` still build
    // inputs without them. Adding them as optional keeps both
    // callers working.
    terms_version?: string;
    privacy_version?: string;
}

interface AuthMeResponse {
    user: User | null;
}

/**
 * Defensive narrowing: the backend should return `{ user: ... }`, but
 * if a future deploy rolls back to a flat shape (or a misconfigured
 * proxy strips the wrapper) we still resolve a sensible value rather
 * than writing `undefined` into `this.user`. Returning `null` makes
 * `isLoggedIn` correctly false so the UserMenu shows "Sign in".
 */
function extractUser(raw: unknown): User | null {
    if (!raw || typeof raw !== 'object') return null;
    const candidate = (raw as { user?: unknown }).user ?? (raw as User);
    if (!candidate || typeof candidate !== 'object') return null;
    const c = candidate as Partial<User>;
    if (typeof c.id !== 'string' || typeof c.email !== 'string') return null;
    return c as User;
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

    isLoggedIn = $derived(!!this.user);

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
                this.user = extractUser(data);
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
            const data = await apiFetch<AuthMeResponse>('/auth/v2/register', {
                method: 'POST',
                body: {
                    email: input.email,
                    password: input.password,
                    locale: input.locale,
                    terms_version: input.terms_version,
                    privacy_version: input.privacy_version
                }
            });
            const user = extractUser(data);
            if (!user) {
                throw new ApiError('INVALID_RESPONSE', 500, 'Register response missing user');
            }
            this.user = user;
            return user;
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
            const data = await apiFetch<AuthMeResponse>('/auth/v2/login', {
                method: 'POST',
                body: { email: input.email, password: input.password }
            });
            const user = extractUser(data);
            if (!user) {
                throw new ApiError('INVALID_RESPONSE', 500, 'Login response missing user');
            }
            this.user = user;
            return user;
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

    /**
     * Claim guest trips after a successful register/login (Spec 017 §5.6).
     *
     * When a user creates trips as guest (cookie `session_id`), then
     * registers or logs in, the backend migrates those trips to the
     * user. This call returns the number of trips claimed so the UI
     * can show a toast ("We've added N trips to your account").
     *
     * Best-effort: a failed claim is NOT blocking. The user is logged
     * in either way; we just couldn't migrate some trips right now.

     * Returns 0 when there's no active user (caller shouldn't have
     * invoked this; we short-circuit instead of failing).
     */
    async claimGuest(): Promise<number> {
        if (!this.user) return 0;
        try {
            const data = await apiFetch<{ claimed_trips_count: number; message: string }>(
                '/auth/v2/claim-guest',
                { method: 'POST' }
            );
            return data.claimed_trips_count;
        } catch (e) {
            // Surface as lastError so a UI banner can offer the retry
            // path, but don't throw — the caller may want to keep
            // showing the logged-in state regardless.
            if (e instanceof ApiError) {
                this.lastError = e;
            }
            return 0;
        }
    }

    /** Manually clear the last error, e.g. after the user dismisses a banner. */
    clearError(): void {
        this.lastError = null;
    }
}

export const auth = new AuthStore();
export type { LoginInput, RegisterInput };
