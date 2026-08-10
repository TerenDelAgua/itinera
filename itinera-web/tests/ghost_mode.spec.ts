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
	// UUID-formatted fake ID. The trip detail page (/trips/[id]+page.svelte)
	// rejects non-UUID ids with a /trips redirect — see the regex at the
	// top of loadAllData(). We use a fake UUID here so the page accepts
	// the route and loadAllData() runs.
	const demoTripId = '00000000-0000-0000-0000-000000000123';
	const forkedTripId = '00000000-0000-0000-0000-000000000456';
	const editedName = 'Japón Clásico (Demo) - Mi Viaje';

	/**
	 * Instala TODOS los mocks que loadAllData() dispara al cargar la página de trip detail,
	 * más los que disparan los componentes hijos (TransportBadge, ExpensesSummaryCard, etc.).
	 *
	 * Sin este set completo, los endpoints sin mockear caen al backend real, que rechaza
	 * el ID fake (`demo-trip-123`) con HTTP 400, lo que rompe Promise.all y produce
	 * la página de error 500 de SvelteKit antes de que se renderice el input del título.
	 */
	async function installMocks(page: import('@playwright/test').Page) {
		let currentName = 'Japón Clásico (Demo)';

		// ── Trip detail (GET + PUT) ─────────────────────────────────────────────
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
						default_expense_currency: 'EUR',
						is_public_demo: true,
						created_at: '2024-01-01T00:00:00Z',
					},
				});
				return;
			}
			if (request.method() === 'PUT') {
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
						default_expense_currency: 'EUR',
						is_public_demo: false,
						forked_from: demoTripId,
						created_at: '2024-01-01T00:00:00Z',
					},
				});
				return;
			}
			await route.continue();
		});

		// ── Sub-resources for the demo URL ─────────────────────────────────────
		await page.route(`**/api/v1/trips/${demoTripId}/expenses/summary*`, async (route) => {
			await route.fulfill({
				json: {
					grand_total: 0,
					global_total: 0,
					places_total: 0,
					total: 0,
					currency: 'EUR',
					by_category: [],
					by_place: [],
				},
			});
		});

		await page.route(`**/api/v1/trips/${demoTripId}/expenses/categories`, async (route) => {
			await route.fulfill({ json: [] });
		});

		await page.route(`**/api/v1/trips/${demoTripId}/expenses*`, async (route) => {
			await route.fulfill({ json: [] });
		});

		await page.route(`**/api/v1/trips/${demoTripId}/places`, async (route) => {
			await route.fulfill({ json: [] });
		});

		await page.route(`**/api/v1/trips/${demoTripId}/places/*`, async (route) => {
			await route.fulfill({
				json: {
					id: 'mock-place-1',
					name: 'Mock Place',
					city: 'Tokyo',
					start_date: null,
					end_date: null,
					notes: '',
					default_expense_currency: null,
				},
			});
		});

		await page.route(`**/api/v1/trips/${demoTripId}/activities*`, async (route) => {
			await route.fulfill({ json: [] });
		});

		// ── Exchange rates (consumed by BackendExchangeService in costPredictor) ─
		// Sin este mock el predictor falla y la Promise.all de loadAllData se rompe.
		await page.route(`**/api/v1/trips/${demoTripId}/rates**`, async (route) => {
			await route.fulfill({ json: { rate: 1, from: 'EUR', to: 'EUR' } });
		});

		// ── Tracking endpoint (sendBeacon + fetch keepalive) ────────────────────
		// Acepta cualquier método (POST desde sendBeacon o fetch). Sin este mock, los
		// eventos de telemetría disparados por Events.demoViewed() y friends llegan
		// al backend real y, combinados con el ID fake, producen 4xx que contaminan
		// la consola del navegador.
		await page.route('**/api/v1/events', async (route) => {
			await route.fulfill({ json: { ok: true } });
		});

		// ── /auth/v2/me (Spec 017 §7.2: boot-time probe) ───────────────────────
		// +layout.svelte fires this probe once at app boot (Spec 017 §7.9:
		// "centralizar en +layout.svelte"). Without this mock the probe hits
		// the real backend, returns 401 (no session cookie in tests), and the
		// resulting unhandled error breaks the Promise.all chain in loadAllData.
		await page.route('**/api/v1/auth/v2/me', async (route) => {
			await route.fulfill({
				status: 401,
				json: {
					error: 'INVALID_CREDENTIALS',
					message: 'No session',
				},
			});
		});
	}

	test('URL stays the same after editing demo, but data reflects the fork', async ({ page }) => {
		await installMocks(page);

		// ── 1. Navigate to the demo trip ───────────────────────────────────────
		await page.goto(`/trips/${demoTripId}`);

		// Wait until the trip data has hydrated into the DetailHeader input.
		await expect(page.locator('input[type="text"]').first()).toHaveValue('Japón Clásico (Demo)', { timeout: 10000 });

		// ── 2. Edit the title (Inline Editing) ─────────────────────────────────
		const titleInput = page.locator('input[type="text"]').first();
		await titleInput.click();
		await titleInput.fill(editedName);

		// ── 3. Trigger save ────────────────────────────────────────────────────
		await titleInput.blur();

		// ── 4. URL must NOT change (Backend-Centric Ghost Mode) ────────────────
		await page.waitForTimeout(800); // allow debounce (300ms) + save round-trip
		expect(page.url()).toContain(`/trips/${demoTripId}`);
		expect(page.url()).not.toContain(forkedTripId);

		// ── 5. The title input now shows the edited name ───────────────────────
		await expect(page.locator('input[type="text"]').first()).toHaveValue(editedName, { timeout: 5000 });
	});
});
