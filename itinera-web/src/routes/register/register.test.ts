/**
 * Component-level tests for /register.
 *
 * The page is a leaf — it does not own state outside its component
 * scope. We exercise the eight states from RegistrationForm_Design.md:
 *   1. empty submit shows field errors
 *   2. invalid email shows email error
 *   3. weak password shows password error
 *   4. terms unchecked shows terms error
 *   5. happy path renders the success dialog
 *   6. EMAIL_ALREADY_EXISTS surfaces the error under the email field
 *   7. RATE_LIMITED surfaces a banner
 *   8. Network failure surfaces the network banner
 *
 * We mock the auth store directly (`vi.mock('$lib/stores/auth.svelte')`)
 * so the tests stay synchronous-ish and don't depend on fetch plumbing
 * — those plumbing details are covered by `auth.svelte.test.ts`.
 */
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { render, screen, waitFor, fireEvent } from '@testing-library/svelte';
import { ApiError } from '$lib/api';

// We mock the auth store with mutable state so the component reacts.
// `vi.hoisted` ensures the mock object exists when `vi.mock` factory
// runs (vi.mock is hoisted above all imports).
const { authMock } = vi.hoisted(() => ({
    authMock: {
        user: null as null | object,
        loading: false,
        lastError: null as ApiError | null,
        isLoggedIn: false,
        register: vi.fn(),
        bootstrap: vi.fn().mockResolvedValue(undefined),
        logout: vi.fn(),
        clearError: vi.fn()
    }
}));

vi.mock('$lib/stores/auth.svelte', () => ({
    auth: authMock
}));

// Imports AFTER mock so the page receives the mocked auth store.
import RegisterPage from './+page.svelte';
import { locale } from '$lib/i18n/store';

beforeEach(() => {
    authMock.user = null;
    authMock.loading = false;
    authMock.lastError = null;
    authMock.isLoggedIn = false;
    authMock.register.mockReset();
    authMock.bootstrap.mockClear();
    authMock.bootstrap.mockResolvedValue(undefined);
    // Force Spanish for predictable assertions on the i18n strings.
    locale.set('es');
});

async function fillForm(email: string, password: string) {
    // We set values via input events so Svelte's bind:value reacts.
    await fireEvent.input(screen.getByLabelText(/Email/i), { target: { value: email } });
    await fireEvent.input(screen.getByLabelText(/Contraseña/i), { target: { value: password } });
    await fireEvent.click(screen.getByLabelText(/Acepto/i));
}

describe('/register', () => {
    it('renders the form with the title and submit button', () => {
        render(RegisterPage);
        expect(screen.getByRole('heading', { name: /Guarda tus viajes/i })).toBeInTheDocument();
        expect(screen.getByRole('button', { name: /Crear cuenta/i })).toBeInTheDocument();
        expect(screen.getByPlaceholderText(/tu@email\.com/i)).toBeInTheDocument();
    });

    it('empty submit shows required-field errors and does not call register', async () => {
        render(RegisterPage);
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        // Use `getAllByText` because the i18n string for "Email es
        // obligatorio" appears once in the inline error AND the field
        // already shows it as an aria-label association.
        expect(await screen.findByText(/Email es obligatorio/i)).toBeInTheDocument();
        expect(await screen.findByText(/Contraseña es obligatoria/i)).toBeInTheDocument();
        expect(await screen.findByText(/Acepta los términos/i)).toBeInTheDocument();
        expect(authMock.register).not.toHaveBeenCalled();
    });

    it('invalid email format shows the validation message', async () => {
        render(RegisterPage);
        await fireEvent.input(screen.getByPlaceholderText(/tu@email\.com/i), { target: { value: 'not-an-email' } });
        await fireEvent.input(screen.getByPlaceholderText(/8 caracteres/i), { target: { value: 'GoodPass1!' } });
        await fireEvent.click(screen.getByLabelText(/Acepto/i));
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByText(/Revisa los campos/i)).toBeInTheDocument();
        expect(authMock.register).not.toHaveBeenCalled();
    });

    it('weak password (no digit/symbol) shows the weak-password error', async () => {
        render(RegisterPage);
        await fireEvent.input(screen.getByPlaceholderText(/tu@email\.com/i), { target: { value: 'jane@example.test' } });
        await fireEvent.input(screen.getByPlaceholderText(/8 caracteres/i), { target: { value: 'onlyletters' } });
        await fireEvent.click(screen.getByLabelText(/Acepto/i));
        await fireEvent.click(screen.getByRole('button', { name: /Crear cuenta/i }));

        expect(await screen.findByText(/8 caracteres con un número/i)).toBeInTheDocument();
        expect(authMock.register).not.toHaveBeenCalled();
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
            expect(screen.getByRole('heading', { name: /Bienvenido/i })).toBeInTheDocument()
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

        expect(await screen.findByText(/ya está en uso/i)).toBeInTheDocument();
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

        expect(await screen.findByText(/Sin conexión/i)).toBeInTheDocument();
    });
});
