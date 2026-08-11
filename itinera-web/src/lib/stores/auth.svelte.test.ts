/**
 * Tests for the auth store. We stub `global.fetch` directly because
 * `apiFetch` is the only network dependency and `vitest-setup.ts` runs
 * in jsdom, so the real `fetch` is available globally.
 *
 * Each test case resets `auth.user` / `auth.lastError` to a known state
 * via `beforeEach`. We do NOT mock the auth store module itself — the
 * exported singleton is the surface under test.
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { auth } from './auth.svelte';
import { ApiError } from '$lib/api';

const sampleUser = {
    id: 'usr_123',
    email: 'jane@example.test',
    tier: 'free' as const,
    locale: 'es',
    terms_accepted_at: '2026-08-04T10:00:00Z',
    created_at: '2026-08-04T10:00:00Z'
};

function mockFetchOnce(status: number, body: unknown) {
    // 204/205/304 cannot carry a body in the Response constructor. Use
    // an empty body string and let the body slot stay null for those.
    const bodyText = status === 204 || status === 205 || status === 304 ? null : JSON.stringify(body);
    vi.spyOn(globalThis, 'fetch').mockResolvedValueOnce(
        new Response(bodyText, {
            status,
            headers: bodyText ? { 'Content-Type': 'application/json' } : {}
        })
    );
}

beforeEach(() => {
    auth.user = null;
    auth.lastError = null;
    auth.loading = false;
});

afterEach(() => {
    vi.restoreAllMocks();
});

describe('auth.register', () => {
    it('happy path: stores user, clears loading, no error', async () => {
        mockFetchOnce(200, {
            user: sampleUser,
            access_token: 'opaque_access',
            refresh_token: 'opaque_refresh',
            token_type: 'Bearer',
            expires_in: 86400
        });

        const user = await auth.register({
            email: 'jane@example.test',
            password: 'GoodPass1!',
            locale: 'es'
        });

        expect(user).toEqual(sampleUser);
        expect(auth.user).toEqual(sampleUser);
        expect(auth.loading).toBe(false);
        expect(auth.lastError).toBeNull();

        // Verify the request itself: POST /auth/v2/register with JSON body.
        const call = (fetch as unknown as ReturnType<typeof vi.fn>).mock.calls[0];
        // The apiFetch helper prepends VITE_API_URL (which already
        // includes `/api/v1`), so callers pass `/auth/v2/register`
        // and the final URL is `${API_URL}/auth/v2/register`.
        expect(call[0]).toMatch(/\/api\/v1\/auth\/v2\/register$/);
        expect(call[1].method).toBe('POST');
        const body = JSON.parse(call[1].body);
        expect(body).toEqual({
            email: 'jane@example.test',
            password: 'GoodPass1!',
            locale: 'es'
        });
        expect(call[1].credentials).toBe('include');
    });

    it('EMAIL_ALREADY_EXISTS: surfaces ApiError so caller can branch on code', async () => {
        mockFetchOnce(409, {
            error: {
                code: 'EMAIL_ALREADY_EXISTS',
                message: 'Email ya registrado'
            }
        });

        await expect(
            auth.register({ email: 'taken@example.test', password: 'GoodPass1!' })
        ).rejects.toThrow(ApiError);

        expect(auth.user).toBeNull();
        expect(auth.loading).toBe(false);
        const err = auth.lastError;
        expect(err).toBeInstanceOf(ApiError);
        expect(err?.code).toBe('EMAIL_ALREADY_EXISTS');
        expect(err?.status).toBe(409);
    });

    it('RATE_LIMITED: server message preserved', async () => {
        mockFetchOnce(429, {
            error: { code: 'RATE_LIMITED', message: 'Demasiados intentos' }
        });

        await expect(
            auth.register({ email: 'x@example.test', password: 'GoodPass1!' })
        ).rejects.toThrow(ApiError);

        expect(auth.lastError?.code).toBe('RATE_LIMITED');
        expect(auth.lastError?.message).toBe('Demasiados intentos');
    });

    it('non-ApiError failure: still rejects and clears loading', async () => {
        vi.spyOn(globalThis, 'fetch').mockRejectedValueOnce(new TypeError('Network down'));

        await expect(
            auth.register({ email: 'x@example.test', password: 'GoodPass1!' })
        ).rejects.toThrow('Network down');

        expect(auth.loading).toBe(false);
        // Non-ApiError: lastError stays null (the page surfaces its own network banner)
        expect(auth.lastError).toBeNull();
    });
});

describe('auth.bootstrap', () => {
    it('populates user when /me returns 200', async () => {
        mockFetchOnce(200, { user: sampleUser });
        await auth.bootstrap();
        expect(auth.user).toEqual(sampleUser);
        expect(auth.lastError).toBeNull();
    });

    it('treats 401 as guest state, no error', async () => {
        mockFetchOnce(401, { error: { code: 'UNAUTHENTICATED', message: 'No session' } });
        await auth.bootstrap();
        expect(auth.user).toBeNull();
        expect(auth.lastError).toBeNull();
    });

    it('surfaces 500 as lastError but keeps guest state', async () => {
        mockFetchOnce(500, { error: { code: 'INTERNAL_ERROR', message: 'Boom' } });
        await auth.bootstrap();
        expect(auth.user).toBeNull();
        expect(auth.lastError?.code).toBe('INTERNAL_ERROR');
    });

    it('is idempotent: a second call while loaded is a no-op', async () => {
        mockFetchOnce(200, { user: sampleUser });
        await auth.bootstrap();
        // No second mock: if bootstrap re-fetches, the test fails.
        await auth.bootstrap();
        expect(auth.user).toEqual(sampleUser);
        expect((fetch as unknown as ReturnType<typeof vi.fn>).mock.calls).toHaveLength(1);
    });
});

describe('auth.logout', () => {
    it('clears local state even when the server returns 500', async () => {
        auth.user = sampleUser;
        mockFetchOnce(500, { error: { code: 'INTERNAL_ERROR', message: 'Boom' } });

        await auth.logout();
        expect(auth.user).toBeNull();
    });

    it('clears local state on 204 success', async () => {
        auth.user = sampleUser;
        mockFetchOnce(204, {});

        await auth.logout();
        expect(auth.user).toBeNull();
    });
});

describe('auth.clearError', () => {
    it('resets lastError', () => {
        auth.lastError = new ApiError('INTERNAL_ERROR', 500, 'boom');
        auth.clearError();
        expect(auth.lastError).toBeNull();
    });
});

describe('auth.claimGuest', () => {
    it('returns 0 when there is no active user', async () => {
        auth.user = null;
        const fetchSpy = vi.spyOn(globalThis, 'fetch');
        const count = await auth.claimGuest();
        expect(count).toBe(0);
        // The fetch must NOT have been called when there's no user.
        expect(fetchSpy).not.toHaveBeenCalled();
    });

    it('returns the claimed trips count on success', async () => {
        auth.user = sampleUser;
        mockFetchOnce(200, {
            claimed_trips_count: 3,
            message: 'Hemos añadido 3 viajes a tu cuenta'
        });

        const count = await auth.claimGuest();
        expect(count).toBe(3);
        expect(auth.lastError).toBeNull();
    });

    it('does NOT throw on 401 (Spec 017 §11.2 #18) — lastError is set, return 0', async () => {
        auth.user = sampleUser;
        mockFetchOnce(401, {
            error: { code: 'UNAUTHENTICATED', message: 'No session' }
        });

        // The call must resolve, NOT reject, so the redirect in /login
        // isn't blocked by a claim failure.
        const count = await auth.claimGuest();
        expect(count).toBe(0);
        expect(auth.lastError).toBeInstanceOf(ApiError);
        expect(auth.lastError?.code).toBe('UNAUTHENTICATED');
        // Crucially: the user is still logged in. claimGuest is best-effort.
        expect(auth.user).toEqual(sampleUser);
    });

    it('does NOT throw on 500 — local state stays consistent', async () => {
        auth.user = sampleUser;
        mockFetchOnce(500, {
            error: { code: 'INTERNAL_ERROR', message: 'boom' }
        });

        const count = await auth.claimGuest();
        expect(count).toBe(0);
        expect(auth.user).toEqual(sampleUser);
    });
});
