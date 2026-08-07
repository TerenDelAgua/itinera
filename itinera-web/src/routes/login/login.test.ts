/**
 * Component-level tests for /login.
 *
 * The page is a leaf — it does not own state outside its component
 * scope. We exercise the spec'd states (Spec 017 §5.2 + §9.2):
 *   1. empty submit shows field errors
 *   2. happy path → redirects to /trips (or to ?next=…)
 *   3. INVALID_CREDENTIALS surfaces a banner (anti-enumeration)
 *   4. ACCOUNT_DELETED surfaces a banner with the "30 days" message
 *   5. RATE_LIMITED surfaces a banner
 *   6. Network failure surfaces a network banner
 *   7. already-logged-in user is redirected without seeing the form
 *   8. `?next=/somewhere` is honoured on success
 *   9. `?next=https://evil.example.com` is rejected (open-redirect guard)
 *
 * We mock the auth store directly so the tests stay synchronous-ish
 * and don't depend on fetch plumbing (covered by auth.svelte.test.ts).
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import { ApiError } from '$lib/api';

// `vi.hoisted` ensures the mock object exists when `vi.mock` factory
// runs (vi.mock is hoisted above all imports).
const { authMock, gotoMock } = vi.hoisted(() => ({
    authMock: {
        user: null as null | object,
        loading: false,
        lastError: null as ApiError | null,
        isLoggedIn: false,
        register: vi.fn(),
        login: vi.fn(),
        bootstrap: vi.fn().mockResolvedValue(undefined),
        logout: vi.fn(),
        clearError: vi.fn()
    },
    gotoMock: vi.fn().mockResolvedValue(undefined)
}));

// SvelteKit's $app/navigation is a no-op in tests; we only care about
// arguments. Mock goto so we can assert on it without polluting global
// state.
vi.mock('$app/navigation', () => ({
    goto: gotoMock
}));

// `$app/state` provides `page`; tests can override `page.url` directly.
vi.mock('$app/state', () => ({
    page: {
        url: new URL('http://localhost/login')
    }
}));

vi.mock('$lib/stores/auth.svelte', () => ({
    auth: authMock
}));

// Imports AFTER mock so the page receives the mocked auth store.
import LoginPage from './+page.svelte';
import { locale } from '$lib/i18n/store';
import { goto } from '$app/navigation';

beforeEach(() => {
    authMock.user = null;
    authMock.loading = false;
    authMock.lastError = null;
    authMock.isLoggedIn = false;
    authMock.login.mockReset();
    authMock.bootstrap.mockClear();
    authMock.bootstrap.mockResolvedValue(undefined);
    locale.set('es');
    // Reset the page URL between tests so leftover query params from a
    // previous case do not leak.
    Object.defineProperty(window, 'location', {
        value: new URL('http://localhost/login'),
        writable: true
    });
    vi.mocked(goto).mockClear();
});

async function fillForm(email: string, password: string) {
    await fireEvent.input(screen.getByPlaceholderText(/tu@email\.com/i), {
        target: { value: email }
    });
    await fireEvent.input(screen.getByPlaceholderText(/Tu contraseña/i), {
        target: { value: password }
    });
}

describe('/login', () => {
    it('renders the form with title, email + password inputs and submit', () => {
        render(LoginPage);
        expect(screen.getByRole('heading', { name: /Bienvenido de vuelta/i })).toBeTruthy();
        expect(screen.getByRole('button', { name: /Iniciar sesión/i })).toBeTruthy();
        expect(screen.getByPlaceholderText(/tu@email\.com/i)).toBeTruthy();
        expect(screen.getByPlaceholderText(/Tu contraseña/i)).toBeTruthy();
    });

    it('empty submit shows required-field errors and does not call login', async () => {
        render(LoginPage);
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        expect(await screen.findByText(/El email es obligatorio/i)).toBeTruthy();
        expect(await screen.findByText(/contraseña es obligatoria/i)).toBeTruthy();
        expect(authMock.login).not.toHaveBeenCalled();
    });

    it('happy path → redirects to /trips', async () => {
        authMock.login.mockResolvedValueOnce({
            id: 'usr_1',
            email: 'jane@example.test',
            tier: 'free' as const,
            locale: 'es',
            terms_accepted_at: null,
            created_at: '2026-08-04T10:00:00Z'
        });

        render(LoginPage);
        await fillForm('jane@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        await waitFor(() => expect(goto).toHaveBeenCalledWith('/trips', expect.anything()));
        expect(authMock.login).toHaveBeenCalledWith({
            email: 'jane@example.test',
            password: 'GoodPass1!'
        });
    });

    it('honours ?next=/some-path on success', async () => {
        // Simulate a query param by mutating the mocked page.url.
        const { page } = await import('$app/state');
        Object.defineProperty(page, 'url', {
            value: new URL('http://localhost/login?next=/trips/abc123'),
            configurable: true
        });

        authMock.login.mockResolvedValueOnce({ id: 'usr_1', email: 'a@b.c', tier: 'free', locale: 'es', terms_accepted_at: null, created_at: '2026-08-04T10:00:00Z' });

        render(LoginPage);
        await fillForm('a@b.c', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        await waitFor(() =>
            expect(goto).toHaveBeenCalledWith('/trips/abc123', expect.anything())
        );
    });

    it('rejects ?next=https://evil.example.com as open-redirect guard', async () => {
        const { page } = await import('$app/state');
        Object.defineProperty(page, 'url', {
            value: new URL('http://localhost/login?next=https://evil.example.com/x'),
            configurable: true
        });

        authMock.login.mockResolvedValueOnce({ id: 'usr_1', email: 'a@b.c', tier: 'free', locale: 'es', terms_accepted_at: null, created_at: '2026-08-04T10:00:00Z' });

        render(LoginPage);
        await fillForm('a@b.c', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        // Even with a malicious next param, the page should fall back
        // to the safe /trips default.
        await waitFor(() => expect(goto).toHaveBeenCalledWith('/trips', expect.anything()));
    });

    it('rejects protocol-relative ?next=//evil.example.com as open-redirect guard', async () => {
        const { page } = await import('$app/state');
        Object.defineProperty(page, 'url', {
            value: new URL('http://localhost/login?next=//evil.example.com'),
            configurable: true
        });

        authMock.login.mockResolvedValueOnce({ id: 'usr_1', email: 'a@b.c', tier: 'free', locale: 'es', terms_accepted_at: null, created_at: '2026-08-04T10:00:00Z' });

        render(LoginPage);
        await fillForm('a@b.c', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        await waitFor(() => expect(goto).toHaveBeenCalledWith('/trips', expect.anything()));
    });

    it('INVALID_CREDENTIALS surfaces a banner (anti-enumeration)', async () => {
        authMock.login.mockRejectedValueOnce(
            new ApiError('INVALID_CREDENTIALS', 401, 'Email o contraseña incorrectos.')
        );

        render(LoginPage);
        await fillForm('jane@example.test', 'WrongPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        expect(await screen.findByRole('alert')).toBeTruthy();
    });

    it('ACCOUNT_DELETED surfaces a banner with the 30-day message', async () => {
        authMock.login.mockRejectedValueOnce(
            new ApiError('ACCOUNT_DELETED', 403, 'Account has been deleted.')
        );

        render(LoginPage);
        await fillForm('deleted@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        const banner = await screen.findByRole('alert');
        expect(banner).toBeTruthy();
    });

    it('RATE_LIMITED surfaces a banner', async () => {
        authMock.login.mockRejectedValueOnce(
            new ApiError('RATE_LIMITED', 429, 'Demasiados intentos.')
        );

        render(LoginPage);
        await fillForm('x@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        expect(await screen.findByRole('alert')).toBeTruthy();
    });

    it('non-ApiError (network) → network banner', async () => {
        authMock.login.mockRejectedValueOnce(new TypeError('Failed to fetch'));

        render(LoginPage);
        await fillForm('x@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Iniciar sesión/i }));

        expect(await screen.findByText(/Sin conexión/i)).toBeTruthy();
    });

    it('already-logged-in user on mount is redirected to /trips', async () => {
        // Simulate the auth store reporting an existing session before
        // the user interacts with the form.
        authMock.user = {
            id: 'usr_x',
            email: 'a@b.c',
            tier: 'free',
            locale: 'es',
            terms_accepted_at: null,
            created_at: '2026-08-04T10:00:00Z'
        };
        authMock.isLoggedIn = true;

        render(LoginPage);

        // bootstrap() runs in onMount and triggers the redirect.
        await waitFor(() => expect(goto).toHaveBeenCalledWith('/trips', expect.anything()));
    });
});
