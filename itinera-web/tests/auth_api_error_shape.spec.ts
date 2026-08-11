/**
 * Auth E2E — Spec 017 §11.2 #13: Frontend ApiError roundtrip.
 *
 * Forces `INVALID_CREDENTIALS` from a mocked backend, then asserts:
 *   - The login form shows the localised Spanish error message, not the
 *     raw English string from the server.
 *   - The post-delete hint banner is visible after the failure.
 *
 * We mock the API instead of standing up a real backend — the spec
 * only requires that the FRONTEND correctly maps server error codes
 * to i18n strings and renders the right banners.
 *
 * Stability note: when the Playwright suite runs in parallel with
 * other specs (default workers = 8) the dev server gets saturated
 * and Svelte hydration can race with our fill()/click() calls. We
 * mitigate this with `test.slow()` (3× the default timeout) and
 * `waitUntil: 'domcontentloaded'` plus `toBeEditable()` to ensure
 * the input is hydrated before we type into it.
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

		await page.goto('/login', { waitUntil: 'domcontentloaded' });

		// `toBeEditable` waits until the input is hydrated and event
		// handlers are attached, which is what we need to type into
		// without losing the input event to a pre-hydration form.
		const emailInput = page.locator('#login-email');
		const passwordInput = page.locator('#login-password');
		await expect(emailInput).toBeEditable();
		await expect(passwordInput).toBeEditable();

		// Fill credentials and submit.
		await emailInput.fill('foo@example.test');
		await passwordInput.fill('bad-password');
		await page.locator('button[type="submit"]').first().click();

		// Wait for the post-delete hint (Spec §5.9 — set after any
		// failed login). This is the simplest, most stable signal that
		// the fetch round-trip completed and applyServerError ran.
		await expect(page.getByTestId('login-post-delete-hint')).toBeVisible({ timeout: 20_000 });

		// Now check the banner exists and has non-empty content.
		// (bannerError is set right before the hint, so by this point
		// it must also be present.)
		const banner = page.locator('form div[role="alert"]').first();
		await expect(banner).toBeVisible({ timeout: 20_000 });
		await expect(banner).toHaveText(/.+/, { timeout: 20_000 });

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

		await page.goto('/forgot', { waitUntil: 'domcontentloaded' });

		// Wait for the email input to be ready and editable.
		const emailInput = page.locator('input[type="email"]').first();
		await expect(emailInput).toBeEditable();

		await emailInput.fill('maybe-exists@example.test');
		await page.locator('button[type="submit"]').first().click();

		// The hint is rendered after the mock returns 202 + the form
		// advances to `codeSent = true`. Generous timeout because in
		// parallel CI the network round-trip can take a moment.
		await expect(page.getByTestId('forgot-post-delete-hint')).toBeVisible({ timeout: 30_000 });
	});
});
