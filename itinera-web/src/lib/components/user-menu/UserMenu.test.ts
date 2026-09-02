/**
 * Five scenarios:
 *   1. Three render states (guest / loading / authenticated) — covers
 *      the layout-shift fix and the Sign-in link copy.
 *   2. Dropdown opens/closes on click + outside click + Escape.
 *   3. Dropdown keyboard navigation (ArrowDown / ArrowUp / Home / End).
 *   4. Sign-out flow: closes menu, calls auth.logout(), no confirmation.
 *   5. A11y invariants (aria-haspopup, aria-expanded, role="menu",
 *      role="menuitem", role="separator", aria-label on avatar).
 *
 * The auth store is mutated directly between tests because every
 * call site uses the named export `auth` — see auth.svelte.ts
 * "Test isolation" note.
 */
import { render, screen, fireEvent, cleanup } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, vi, afterEach } from 'vitest';
import UserMenu from './UserMenu.svelte';
import { auth } from '$lib/stores/auth.svelte';
import type { User } from '$lib/types/auth';

// Mock SvelteKit navigation: UserMenu calls `goto` and
// `invalidateAll` after `auth.logout()` so cached data (e.g. the
// trips list) refetches with the cleared cookies. jsdom doesn't
// initialise the SvelteKit client runtime, so we stub these to
// resolvable no-ops.
vi.mock('$app/navigation', () => ({
    goto: vi.fn().mockResolvedValue(undefined),
    invalidateAll: vi.fn().mockResolvedValue(undefined)
}));

// Mock i18n store so the component reads a deterministic translator
// that echoes its key (with optional `{var}` substitution) — keeps
// test assertions about visible labels language-agnostic. We mock
// `t` as a Svelte readable so the component's `$t(...)` auto-
// subscription works the same way as in production.
vi.mock('$lib/i18n/store', () => {
    const { writable, derived } = require('svelte/store');
    const locale = writable('en');
    // Mirror the production store shape: `t` is a Svelte store whose
    // value is a function. The component uses `$t(...)` to auto-subscribe.
    const t = derived(locale, () => (key: string, params?: Record<string, string>) => {
        if (key === 'auth.user_menu.avatar_label' && params?.email) {
            return `Account of ${params.email}`;
        }
        return key;
    });
    return { t, locale };
});

// Mock transitions to keep tests deterministic and avoid the
// Svelte transition microtask chain.
vi.mock('svelte/transition', () => ({
    fly: () => ({ duration: 0 })
}));

// jsdom doesn't ship `window.matchMedia`; UserMenu reads
// `prefers-reduced-motion` to disable the fly animation.
if (typeof window !== 'undefined' && !window.matchMedia) {
    window.matchMedia = vi.fn().mockImplementation((query: string) => ({
        matches: false,
        media: query,
        onchange: null,
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        addListener: vi.fn(),
        removeListener: vi.fn(),
        dispatchEvent: vi.fn()
    }));
}

// Stable user fixture — a real email with predictable initials ("JU").
const userFixture: User = {
    id: 'usr_1',
    email: 'juan.carlos@example.test',
    tier: 'free',
    locale: 'en',
    terms_accepted_at: '2026-08-08T10:00:00Z',
    created_at: '2026-08-08T10:00:00Z'
};

describe('UserMenu.svelte', () => {
    beforeEach(() => {
        // Reset between cases. Tests are sequential; without this,
        // `auth.user = userFixture` would persist across files.
        auth.user = null;
        auth.isLoading = false;
        auth.lastError = null;
        auth._bootstrapInflight = null;
        vi.restoreAllMocks();
    });

    afterEach(() => {
        cleanup();
    });

    // ─────────────────────── Test #19: render states
    it('renders the Sign-in link for guests', () => {
        auth.user = null;
        auth.isLoading = false;

        render(UserMenu);

        // Skeleton must NOT appear for guests.
        expect(screen.queryByTestId('user-menu-skeleton')).toBeNull();
        // Avatar must NOT appear for guests.
        expect(screen.queryByTestId('user-menu-avatar')).toBeNull();
        // Sign-in anchor IS the only render output.
        const link = screen.getByTestId('user-menu-signin');
        expect(link).toBeTruthy();
        expect(link.getAttribute('href')).toBe('/login');
        expect(link.textContent?.trim()).toBe('auth.sign_in');
    });

    it('renders the 32px skeleton while auth.isLoading is true', () => {
        auth.isLoading = true;

        render(UserMenu);

        // Only the skeleton; no Sign-in, no Avatar.
        const skeleton = screen.getByTestId('user-menu-skeleton');
        expect(skeleton).toBeTruthy();
        // a11y invariants from DS §5.6.2.
        expect(skeleton.getAttribute('aria-hidden')).toBe('true');
        expect(skeleton.getAttribute('role')).toBe('presentation');
        expect(screen.queryByTestId('user-menu-signin')).toBeNull();
        expect(screen.queryByTestId('user-menu-avatar')).toBeNull();
    });

    it('renders the avatar with initials for authenticated users', () => {
        auth.user = userFixture;

        render(UserMenu);

        const avatar = screen.getByTestId('user-menu-avatar');
        expect(avatar).toBeTruthy();
        // a11y: aria-haspopup + aria-expanded=false by default.
        expect(avatar.getAttribute('aria-haspopup')).toBe('menu');
        expect(avatar.getAttribute('aria-expanded')).toBe('false');
        // DS §5.6.2: aria-label must contain the email so SR users
        // know which account this menu opens.
        expect(avatar.getAttribute('aria-label')).toBe(
            'Account of juan.carlos@example.test'
        );
        // Initials: "juan.carlos" → strip dots → "JUANC" → first 2.
        expect(avatar.textContent?.trim()).toBe('JU');
    });

    // ─────────────────────── Test #20: open / close
    it('opens and closes the dropdown on avatar click', async () => {
        auth.user = userFixture;
        render(UserMenu);

        // Initially closed.
        expect(screen.queryByTestId('user-menu-dropdown')).toBeNull();

        // Open.
        await fireEvent.click(screen.getByTestId('user-menu-avatar'));
        const dropdown = screen.getByTestId('user-menu-dropdown');
        expect(dropdown).toBeTruthy();
        expect(screen.getByTestId('user-menu-avatar').getAttribute('aria-expanded'))
            .toBe('true');

        // Close by re-clicking the avatar.
        await fireEvent.click(screen.getByTestId('user-menu-avatar'));
        expect(screen.queryByTestId('user-menu-dropdown')).toBeNull();
        expect(screen.getByTestId('user-menu-avatar').getAttribute('aria-expanded'))
            .toBe('false');
    });

    it('closes the dropdown on Escape and restores focus to the avatar', async () => {
        auth.user = userFixture;
        render(UserMenu);

        const avatar = screen.getByTestId('user-menu-avatar');
        await fireEvent.click(avatar);
        expect(screen.queryByTestId('user-menu-dropdown')).toBeTruthy();

        await fireEvent.keyDown(document, { key: 'Escape' });

        expect(screen.queryByTestId('user-menu-dropdown')).toBeNull();
        expect(document.activeElement).toBe(avatar);
    });

    it('closes the dropdown on outside click', async () => {
        auth.user = userFixture;
        render(UserMenu);

        await fireEvent.click(screen.getByTestId('user-menu-avatar'));
        expect(screen.queryByTestId('user-menu-dropdown')).toBeTruthy();

        // A click outside both the menu and the avatar must close it.
        await fireEvent.click(document.body);

        expect(screen.queryByTestId('user-menu-dropdown')).toBeNull();
    });

    // ─────────────────────── Test #21: keyboard navigation
    it('navigates menuitems with ArrowDown / ArrowUp / Home / End', async () => {
        auth.user = userFixture;
        render(UserMenu);

        await fireEvent.click(screen.getByTestId('user-menu-avatar'));

        const account = screen.getByTestId('user-menu-account');
        const signOut = screen.getByTestId('user-menu-signout');

        // On open, focus moves to the FIRST menuitem (account).
        expect(document.activeElement).toBe(account);

        // ArrowDown → signOut.
        await fireEvent.keyDown(document, { key: 'ArrowDown' });
        expect(document.activeElement).toBe(signOut);

        // ArrowUp → back to account.
        await fireEvent.keyDown(document, { key: 'ArrowUp' });
        expect(document.activeElement).toBe(account);

        // End → last menuitem.
        await fireEvent.keyDown(document, { key: 'End' });
        expect(document.activeElement).toBe(signOut);

        // Home → first.
        await fireEvent.keyDown(document, { key: 'Home' });
        expect(document.activeElement).toBe(account);
    });

    // ─────────────────────── Test #21: sign-out flow
    it('Sign out closes the menu, calls auth.logout(), then redirects', async () => {
        auth.user = userFixture;
        const logoutSpy = vi.spyOn(auth, 'logout').mockResolvedValue();
        const { goto, invalidateAll } = await import('$app/navigation');

        render(UserMenu);

        await fireEvent.click(screen.getByTestId('user-menu-avatar'));
        await fireEvent.click(screen.getByTestId('user-menu-signout'));

        // Dropdown closed immediately (DS §5.6.2 "Optimistic state
        // closure": no spinner, no confirm modal).
        expect(screen.queryByTestId('user-menu-dropdown')).toBeNull();

        // The logout promise is awaited so observers can await.
        expect(logoutSpy).toHaveBeenCalledTimes(1);
        // Then we flush the load cache and navigate to "/" so any
        // page (e.g. /trips) re-runs its data fetch with the now-guest
        // cookies instead of showing the stale account-owned trips.
        expect(invalidateAll).toHaveBeenCalled();
        expect(goto).toHaveBeenCalledWith('/', expect.objectContaining({ invalidateAll: true }));
    });

    // ─────────────────────── Test #22: a11y invariants
    it('marks the dropdown and items with the correct roles', async () => {
        auth.user = userFixture;
        render(UserMenu);

        await fireEvent.click(screen.getByTestId('user-menu-avatar'));

        const dropdown = screen.getByTestId('user-menu-dropdown');
        expect(dropdown.getAttribute('role')).toBe('menu');

        const items = screen.getAllByRole('menuitem');
        expect(items).toHaveLength(2);

        // Separators exist (Spec §7.9).
        const separators = screen.getAllByRole('separator');
        expect(separators.length).toBeGreaterThanOrEqual(2);

        // Email label is present but NOT a menuitem (it's decorative
        // info, not an action).
        const email = screen.getByTestId('user-menu-email');
        expect(email.textContent?.trim()).toBe(userFixture.email);
        expect(email.getAttribute('role')).not.toBe('menuitem');
    });
});