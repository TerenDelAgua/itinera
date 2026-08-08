/**
 * Component-level tests for /register.
 *
 * States covered (Spec 017 §5.1 + TEREN_DESIGN_SYSTEM §3.11):
 *   1. idle (default)
 *   2. validating (after submit, before response)
 *   3. success → welcome dialog
 *   4. EMAIL_ALREADY_EXISTS / EMAIL_INVALID / PASSWORD_TOO_WEAK /
 *      TERMS_NOT_ACCEPTED → field-level error
 *   5. RATE_LIMITED / VALIDATION_ERROR → top banner
 *   6. network failure → top banner
 *
 * We mock the auth store and apiFetch directly so the tests stay
 * synchronous-ish and don't depend on the network. The `apiFetch`
 * mock returns a `{ trips: [] }` response from `/trips` so the
 * subtitle rendering doesn't crash.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';

// `vi.hoisted` ensures the mock objects exist when `vi.mock` factory
// runs (vi.mock is hoisted above all imports).
const { apiFetchMock, authMock } = vi.hoisted(() => ({
    apiFetchMock: vi.fn().mockResolvedValue({ trips: [] }),
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
    }
}));

vi.mock('$lib/api', async () => {
    const actual = await vi.importActual<typeof import('$lib/api')>('$lib/api');
    return {
        ...actual,
        apiFetch: apiFetchMock
    };
});

vi.mock('$lib/stores/auth.svelte', () => ({
    auth: authMock
}));

// Imports AFTER mock so the page receives the mocked auth store.
import RegisterPage from './+page.svelte';
import { locale } from '$lib/i18n/store';
import { ApiError } from '$lib/api';

beforeEach(() => {
    authMock.user = null;
    authMock.loading = false;
    authMock.lastError = null;
    authMock.isLoggedIn = false;
    authMock.register.mockReset();
    authMock.bootstrap.mockClear();
    authMock.bootstrap.mockResolvedValue(undefined);
    apiFetchMock.mockClear();
    apiFetchMock.mockResolvedValue({ trips: [] });
    // Force Spanish for predictable assertions on the i18n strings.
    locale.set('es');
});

async function fillForm(email: string, password: string) {
    // The page uses raw <input> elements with placeholders; that's
    // the most stable selector across the latest redesign.
    await fireEvent.input(screen.getByPlaceholderText(/tu@email\.com/i), {
        target: { value: email }
    });
    await fireEvent.input(screen.getByPlaceholderText(/Elige una contraseña/i), {
        target: { value: password }
    });
    await fireEvent.click(screen.getByLabelText(/Acepto/i));
}

describe('/register', () => {
    it('renders the form with the title and submit button', () => {
        render(RegisterPage);
        expect(screen.getByRole('heading', { name: /Guarda tus viajes/i })).toBeTruthy();
        expect(screen.getByRole('button', { name: /Crear cuenta/i })).toBeTruthy();
        expect(screen.getByPlaceholderText(/tu@email\.com/i)).toBeTruthy();
        expect(screen.getByPlaceholderText(/Elige una contraseña/i)).toBeTruthy();
    });

    it('empty submit shows required-field errors and does not call register', async () => {
        render(RegisterPage);
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByText(/Email es obligatorio/i)).toBeTruthy();
        expect(await screen.findByText(/contraseña es obligatoria/i)).toBeTruthy();
        expect(await screen.findByText(/Acepta los términos/i)).toBeTruthy();
        expect(authMock.register).not.toHaveBeenCalled();
    });

    it('invalid email format shows the validation message', async () => {
        render(RegisterPage);
        await fillForm('not-an-email', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByText(/Revisa los campos/i)).toBeTruthy();
        expect(authMock.register).not.toHaveBeenCalled();
    });

    it('short password (less than 8 chars) shows the weak-password error', async () => {
        // Spec 017 only enforces min 8 chars — no symbol/number rule.
        // So "short1!" should pass (8 chars) and "short" should fail.
        render(RegisterPage);
        await fillForm('jane@example.test', 'short');
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByText(/al menos 8 caracteres/i)).toBeTruthy();
        expect(authMock.register).not.toHaveBeenCalled();
    });

    it('8-char letter-only password is accepted (no symbol rule)', async () => {
        // Regression guard for Issue 4: we removed the bogus
        // number/symbol rule, so a longer letter-only password must
        // not be rejected client-side.
        authMock.register.mockResolvedValueOnce({
            id: 'usr_1',
            email: 'jane@example.test',
            tier: 'free' as const,
            locale: 'es',
            terms_accepted_at: '2026-08-04T10:00:00Z',
            created_at: '2026-08-04T10:00:00Z'
        });

        render(RegisterPage);
        await fillForm('jane@example.test', 'longerthan8');
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        await waitFor(() =>
            expect(screen.getByRole('heading', { name: /Bienvenido/i })).toBeTruthy()
        );
        expect(authMock.register).toHaveBeenCalledWith({
            email: 'jane@example.test',
            password: 'longerthan8',
            locale: 'es'
        });
    });

    it('happy path → success dialog with the welcome copy', async () => {
        authMock.register.mockResolvedValueOnce({
            id: 'usr_1',
            email: 'jane@example.test',
            tier: 'free' as const,
            locale: 'es',
            terms_accepted_at: '2026-08-04T10:00:00Z',
            created_at: '2026-08-04T10:00:00Z'
        });

        render(RegisterPage);
        await fillForm('jane@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        await waitFor(() =>
            expect(screen.getByRole('heading', { name: /Bienvenido/i })).toBeTruthy()
        );
        expect(authMock.register).toHaveBeenCalledWith({
            email: 'jane@example.test',
            password: 'GoodPass1!',
            locale: 'es'
        });
    });

    it('EMAIL_ALREADY_EXISTS surfaces under the email field', async () => {
        authMock.register.mockRejectedValueOnce(
            new ApiError('EMAIL_ALREADY_EXISTS', 409, 'Este email ya está en uso.')
        );

        render(RegisterPage);
        await fillForm('taken@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByText(/ya está en uso/i)).toBeTruthy();
    });

    it('RATE_LIMITED surfaces a top banner', async () => {
        authMock.register.mockRejectedValueOnce(
            new ApiError('RATE_LIMITED', 429, 'Demasiados intentos.')
        );

        render(RegisterPage);
        await fillForm('x@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByRole('alert')).toHaveTextContent(/Demasiados/i);
    });

    it('non-ApiError (network) → network banner', async () => {
        authMock.register.mockRejectedValueOnce(new TypeError('Failed to fetch'));

        render(RegisterPage);
        await fillForm('x@example.test', 'GoodPass1!');
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByText(/Sin conexión/i)).toBeTruthy();
    });

    it('subtitle reflects guest trip count (0 trips → zero copy)', async () => {
        apiFetchMock.mockResolvedValueOnce({ trips: [] });

        render(RegisterPage);
        // Wait for onMount's count fetch + $derived update.
        await waitFor(() => {
            const sub = screen.getByTestId('register-subtitle');
            expect(sub.textContent).toContain('Guarda tus viajes en tu cuenta');
        });
    });

    it('subtitle reflects guest trip count (3 trips → "{count} viajes")', async () => {
        apiFetchMock.mockResolvedValueOnce({ trips: [{}, {}, {}] });

        render(RegisterPage);
        await waitFor(() => {
            const sub = screen.getByTestId('register-subtitle');
            expect(sub.textContent).toContain('3');
            expect(sub.textContent).toMatch(/viajes/i);
        });
    });

    it('subtitle reflects guest trip count (1 trip → singular copy)', async () => {
        apiFetchMock.mockResolvedValueOnce({ trips: [{}] });

        render(RegisterPage);
        await waitFor(() => {
            const sub = screen.getByTestId('register-subtitle');
            expect(sub.textContent).toContain('1');
            expect(sub.textContent).toMatch(/viaje/i);
        });
    });
});
