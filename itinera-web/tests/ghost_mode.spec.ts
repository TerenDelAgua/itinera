import { test, expect } from '@playwright/test';

/**
 * Ghost Mode Fork-On-Write - Backend-Centric v1.2.0
 *
 * Con el nuevo middleware ResolveTripContext, la URL NO cambia cuando el usuario edita un demo.
 * El backend resuelve el contexto de forma transparente:
 * - GET /trips/demo-id → Si existe un fork activo, devuelve los datos del fork.
 * - PUT/POST /trips/demo-id → Si es un demo y no es el propietario, crea el fork y opera sobre él.
 * El frontend siempre usa la misma URL.
 */
test.describe('Ghost Mode Fork-On-Write', () => {
	const demoTripId = 'demo-trip-123';
	const forkedTripId = 'forked-trip-456';
	const editedName = 'Japón Clásico (Demo) - Mi Viaje';

	test('URL stays the same after editing demo, but data reflects the fork', async ({ page }) => {
		let currentName = 'Japón Clásico (Demo)';

		// ── Mock: GET initial demo ──────────────────────────────────────────────
		await page.route(`**/api/v1/trips/${demoTripId}`, async (route, request) => {
			if (request.method() === 'GET') {
				await route.fulfill({
					json: {
						id: demoTripId,
						name: currentName,
						description: 'Demo trip description',
						start_date: '2024-05-01',
						end_date: '2024-05-15',
						base_currency: 'EUR',
						is_public_demo: true,
					},
				});
			} else if (request.method() === 'PUT') {
				// Backend forks internally and returns the fork's data.
				const body = JSON.parse(request.postData() || '{}');
				if (body.name) currentName = body.name;

				await route.fulfill({
					json: {
						id: forkedTripId,
						name: currentName,
						description: 'Demo trip description',
						start_date: '2024-05-01',
						end_date: '2024-05-15',
						base_currency: 'EUR',
						is_public_demo: false,
						forked_from: demoTripId,
					},
				});
			} else {
				await route.continue();
			}
		});

		// ── Mock: sub-resources for the demo URL ───────────────────────────────
		await page.route(`**/api/v1/trips/${demoTripId}/expenses/summary*`, async (route) => {
			await route.fulfill({ json: { grand_total: 0, global_total: 0, places_total: 0, by_category: [], by_place: [] } });
		});
		await page.route(`**/api/v1/trips/${demoTripId}/expenses/categories`, async (route) => {
			await route.fulfill({ json: [] });
		});
		await page.route(`**/api/v1/trips/${demoTripId}/places`, async (route) => {
			await route.fulfill({ json: [] });
		});
		await page.route(`**/api/v1/trips/${demoTripId}/activities`, async (route) => {
			await route.fulfill({ json: [] });
		});

		// ── 1. Navigate to the demo trip ───────────────────────────────────────
		await page.goto(`/trips/${demoTripId}`);
		await expect(page.locator('input[type="text"]').first()).toHaveValue('Japón Clásico (Demo)', { timeout: 5000 });

		// ── 2. Edit the title (Inline Editing) ─────────────────────────────────
		const titleInput = page.locator('input[type="text"]').first();
		await titleInput.click();
		await titleInput.fill(editedName);

		// ── 3. Trigger save ────────────────────────────────────────────────────
		await titleInput.blur();

		// ── 4. URL must NOT change (Backend-Centric Ghost Mode) ────────────────
		await page.waitForTimeout(600); // allow debounce (300ms) + save round-trip
		expect(page.url()).toContain(`/trips/${demoTripId}`);
		expect(page.url()).not.toContain(forkedTripId);

		// ── 5. The title input now shows the edited name ───────────────────────
		await expect(page.locator('input[type="text"]').first()).toHaveValue(editedName, { timeout: 5000 });
	});
});
