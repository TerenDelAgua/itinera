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
import type { User } from '$lib/types/auth';

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
    /** Last error from a register/login/logout call. Cleared on entry. */
    lastError = $state<ApiError | null>(null);

    isLoggedIn = $derived(this.user !== null);

    /**
     * Bootstrap the session from cookies. Safe to call multiple times;
     * subsequent calls return immediately without re-fetching.
     */
    async bootstrap(): Promise<void> {
        if (this.user || this.loading) return;
        try {
            const data = await apiFetch<AuthMeResponse>('/api/v1/auth/v2/me');
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
        }
    }

    async register(input: RegisterInput): Promise<User> {
        this.loading = true;
        this.lastError = null;
        try {
            const data = await apiFetch<{ user: User }>('/api/v1/auth/v2/register', {
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
            const data = await apiFetch<{ user: User }>('/api/v1/auth/v2/login', {
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
            await apiFetch<void>('/api/v1/auth/v2/logout', { method: 'POST' });
        } catch (e) {
            if (e instanceof ApiError && e.status >= 500) {
                console.warn('Logout failed on server; local state cleared but server session may still be active in other tabs.');
            }
        }
    }

    /** Manually clear the last error, e.g. after the user dismisses a banner. */
    clearError(): void {
        this.lastError = null;
    }
}

export const auth = new AuthStore();
export type { LoginInput, RegisterInput };
