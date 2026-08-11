/**
 * Auth E2E — Spec 017 §11.2 #13: Frontend ApiError roundtrip.
 *
 * Forces `INVALID_CREDENTIALS` from a mocked backend, then asserts:
 *   - The login form shows the localised error message, not the raw
 *     English string from the server.
 *   - The post-delete hint banner is visible after the failure.
 *
 * Stability: the SvelteKit +layout.svelte fires `auth.bootstrap()`
 * on every page mount. Until that promise settles the form is in
 * a "loading" state and the UserMenu shows a skeleton. The first
 * Svelte rune update after that can wipe a freshly-typed value if
 * we fill() before bootstrap finishes — the input loses its
 * `bind:value` connection momentarily, then re-mounts, and the
 * typed text is gone (the test snapshot showed exactly this:
 * empty textboxes after fill()). We solve it by waiting on a
 * stable DOM signal — the header "Sign in" link — which only
 * appears AFTER bootstrap has set `auth.user = null`. That gives
 * us a deterministic "ready to fill" gate.
 *
 * The submission itself is wrapped in `page.waitForResponse` so we
 * fail fast if the click never reaches the network, instead of
 * waiting 20s on a banner that will never appear.
 */
import { test, expect } from '@playwright/test';

test.describe('Auth API error shape', () => {
	test('login surfaces the post-delete hint and a non-empty error banner', async ({ page }) => {
		test.slow();

		// /me returns 401 → guest state (the +layout probe).
		await page.route('**/api/v1/auth/v2/me', async (route) => {
			await route.fulfill({
				status: 401,
				json: {
					error: { code: 'INVALID_CREDENTIALS', message: 'Invalid email or password' }
				}
			});
		});

		// /login returns the same shape (the actual spec §9.3 contract).
		await page.route('**/api/v1/auth/v2/login', async (route) => {
			await route.fulfill({
				status: 401,
				json: {
					error: {
						code: 'INVALID_CREDENTIALS',
						message: 'Invalid email or password'  // English from the server
					}
				}
			});
		});

		await page.goto('/login', { waitUntil: 'load' });

		// Wait for the bootstrap to settle: the header "Sign in" link
		// only renders once auth.user is null (i.e. after the /me
		// probe has resolved). Without this gate, fill() races the
		// post-bootstrap re-render and the typed values vanish.
		const headerSignInLink = page.locator('header a[href="/login"]');
		await expect(headerSignInLink).toBeVisible({ timeout: 15_000 });

		// Now the form is stable. Sanity-check it's still empty.
		const emailInput = page.locator('#login-email');
		const passwordInput = page.locator('#login-password');
		await expect(emailInput).toHaveValue('');

		// Attach the response waiter BEFORE clicking.
		const loginResponsePromise = page.waitForResponse(
			(resp) => resp.url().endsWith('/api/v1/auth/v2/login'),
			{ timeout: 20_000 }
		);

		await emailInput.pressSequentially('foo@example.test', { delay: 20 });
		await passwordInput.pressSequentially('bad-password', { delay: 20 });
		// Sanity check: confirm Svelte bound the value BEFORE we click.
		// If this fails, the bind:value isn't connected and we'd just
		// be submitting an empty form.
		await expect(emailInput).toHaveValue('foo@example.test', { timeout: 5_000 });
		await page.locator('button[type="submit"]').first().click();

		const loginResponse = await loginResponsePromise;
		expect(loginResponse.status()).toBe(401);

		// Now assert on the banner. By the time the response promise
		// resolves, Svelte has flushed the `applyServerError` update.
		await expect(page.getByTestId('login-post-delete-hint')).toBeVisible({ timeout: 10_000 });

		const banner = page.locator('form div[role="alert"]').first();
		await expect(banner).toBeVisible({ timeout: 10_000 });
		await expect(banner).toHaveText(/.+/, { timeout: 10_000 });

		const bannerText = await banner.textContent();

		// The banner must NOT contain the raw English string from the
		// server ("Invalid email or password"), because the frontend
		// maps INVALID_CREDENTIALS to the i18n string.
		expect(bannerText?.toLowerCase()).not.toContain('invalid email or password');

		// The banner must contain the localised INVALID_CREDENTIALS message
		// (either Spanish "Email o contraseña incorrectos." or English
		// "Incorrect email or password." depending on which locale the
		// store has resolved to).
		const localised =
			/Email o contraseña incorrectos/i.test(bannerText ?? '') ||
			/Incorrect email or password/i.test(bannerText ?? '');
		expect(localised).toBe(true);
	});

	test('forgot surfaces the post-delete hint after a 202', async ({ page }) => {
		test.slow();

		// 202 from /forgot is identical for email-exists and email-missing
		// (anti-enumeration). The post-delete hint must appear.
		await page.route('**/api/v1/auth/v2/me', async (route) => {
			await route.fulfill({ status: 401, json: { error: { code: 'UNAUTHENTICATED' } } });
		});
		await page.route('**/api/v1/auth/v2/forgot', async (route) => {
			await route.fulfill({
				status: 202,
				json: { message: 'If that email exists, we sent a 6-digit code.' }
			});
		});

		await page.goto('/forgot', { waitUntil: 'load' });

		// Wait for bootstrap to settle (same gate as the login test).
		const headerSignInLink = page.locator('header a[href="/login"]');
		await expect(headerSignInLink).toBeVisible({ timeout: 15_000 });

		const emailInput = page.locator('input[type="email"]').first();
		await expect(emailInput).toHaveValue('');

		const forgotResponsePromise = page.waitForResponse(
			(resp) => resp.url().endsWith('/api/v1/auth/v2/forgot'),
			{ timeout: 20_000 }
		);

		await emailInput.pressSequentially('maybe-exists@example.test', { delay: 20 });
		await expect(emailInput).toHaveValue('maybe-exists@example.test', { timeout: 5_000 });
		await page.locator('button[type="submit"]').first().click();

		const forgotResponse = await forgotResponsePromise;
		expect(forgotResponse.status()).toBe(202);

		await expect(page.getByTestId('forgot-post-delete-hint')).toBeVisible({ timeout: 10_000 });
	});
});
