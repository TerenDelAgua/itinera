/**
 * Legal pages E2E (Spec 018 §8 DoD).
 *
 * Smoke tests for /terms and /privacy — both must:
 *   - Return 200 with the rendered HTML (prerender works).
 *   - Show the meta header (title + version + locale toggle).
 *   - Have at least one TOC anchor on desktop widths.
 *   - Navigate between ES and EN via the toggle (cookie + URL).
 *   - Have a Footer present with /terms + /privacy links.
 *   - Register form links to /terms and /privacy with
 *     `rel="noopener noreferrer"`.
 */

import { test, expect } from '@playwright/test';

test.describe('Legal pages', () => {
	test('/terms renders and shows the meta header (ES canonical)', async ({ page }) => {
		test.slow();
		// Wait for the bootstrap gate before interacting (the
		// header "Sign in" link only renders once auth.user is null).
		await page.goto('/terms', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		// H1 comes from the rendered markdown's first line. Match
		// by visible h1, not by text (the title also appears in
		// <svelte:head><title>).
		await expect(page.locator('article h1').first()).toBeVisible();
		await expect(page.getByText(/Versión 1\.1/).first()).toBeVisible();
		await expect(page.getByText(/Última actualización/).first()).toBeVisible();
	});

	test('/terms/en renders the English variant', async ({ page }) => {
		test.slow();
		await page.goto('/terms/en', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		await expect(page.locator('article h1').first()).toBeVisible();
		await expect(page.getByText(/Version 1\.1/i).first()).toBeVisible();
		await expect(page.getByText(/Last updated/i).first()).toBeVisible();
	});

	test('/privacy renders and shows the meta header (ES canonical)', async ({ page }) => {
		test.slow();
		await page.goto('/privacy', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		await expect(page.locator('article h1').first()).toBeVisible();
		await expect(page.getByText(/Versión 1\.1/).first()).toBeVisible();
	});

	test('/privacy/en renders the English variant', async ({ page }) => {
		test.slow();
		await page.goto('/privacy/en', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		await expect(page.locator('article h1').first()).toBeVisible();
	});

	test('TOC anchors on desktop width', async ({ page }) => {
		test.slow();
		await page.setViewportSize({ width: 1280, height: 800 });
		await page.goto('/terms', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		// The aside is `hidden lg:block`, so at 1280px it must be
		// visible. Clicking the first TOC link must navigate to the
		// matching anchor in the rendered HTML.
		const toc = page.locator('aside nav ul li a').first();
		await expect(toc).toBeVisible({ timeout: 10_000 });
		const href = await toc.getAttribute('href');
		expect(href?.startsWith('#')).toBe(true);

		await toc.click();
		await page.waitForURL(/#/, { timeout: 5_000 });
	});

	test('Legal toggle navigates between ES and EN', async ({ page }) => {
		test.slow();
		await page.goto('/terms', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		// The ES button is `aria-pressed="true"` initially.
		const esBtn = page.getByRole('button', { name: 'ES' });
		const enBtn = page.getByRole('button', { name: 'EN' });
		await expect(esBtn).toHaveAttribute('aria-pressed', 'true');
		await expect(enBtn).toHaveAttribute('aria-pressed', 'false');

		// Click EN — should navigate to /terms/en.
		await enBtn.click();
		await page.waitForURL(/\/terms\/en$/, { timeout: 10_000 });
		await expect(page.locator('article h1').first()).toBeVisible();
	});

	test('Footer links to /terms and /privacy', async ({ page }) => {
		test.slow();
		await page.goto('/', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		// The app already has a different footer (the dark marketing
		// footer from the landing page). Our new legal footer is the
		// last one in the DOM. We scope to it to avoid the
		// strict-mode violation.
		const footer = page.locator('footer').last();
		await expect(footer).toBeVisible();
		await expect(footer.locator('a[href="/terms"]')).toBeVisible();
		await expect(footer.locator('a[href="/privacy"]')).toBeVisible();
	});

	test('Register form links open /terms and /privacy with secure rel', async ({ page }) => {
		test.slow();
		await page.goto('/register', { waitUntil: 'load' });
		await expect(page.locator('header a[href="/login"]')).toBeVisible({ timeout: 15_000 });

		// Scope to the form area so we don't accidentally pick up the
		// footer links (which have the same URLs but no `target`).
		const form = page.locator('form');
		const termsLink = form.locator('a[href="/terms"]');
		const privacyLink = form.locator('a[href="/privacy"]');

		// Spec 018 §2 #5: target="_blank" + rel="noopener noreferrer"
		// so the user doesn't lose the form state on click.
		await expect(termsLink).toHaveAttribute('target', '_blank');
		await expect(termsLink).toHaveAttribute('rel', /noopener/);
		await expect(termsLink).toHaveAttribute('rel', /noreferrer/);
		await expect(privacyLink).toHaveAttribute('target', '_blank');
		await expect(privacyLink).toHaveAttribute('rel', /noopener/);
		await expect(privacyLink).toHaveAttribute('rel', /noreferrer/);
	});
});