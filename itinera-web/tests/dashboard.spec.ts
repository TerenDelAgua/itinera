import { test, expect } from '@playwright/test';

test.describe('Dashboard ListTrips', () => {
	test('displays user trips and demo trips correctly separated', async ({ page }) => {
		// Mock the API response to return one user trip and one demo trip
		await page.route('**/api/v1/trips', async (route) => {
			await route.fulfill({
				json: [
					{
						id: 'user-trip-1',
						name: 'My Personal Trip',
						start_date: '2024-01-01',
						end_date: '2024-01-10',
						base_currency: 'EUR',
						is_public_demo: false,
						place_count: 1,
						total_spent: 100
					},
					{
						id: 'demo-trip-1',
						name: 'Japón Demo',
						start_date: '2024-02-01',
						end_date: '2024-02-15',
						base_currency: 'EUR',
						is_public_demo: true,
						place_count: 5,
						total_spent: 0
					}
				]
			});
		});

		await page.goto('/trips');

		// Ensure the page has loaded by checking for the main dashboard title
		await expect(page.getByRole('heading', { name: /My trips/i })).toBeVisible();

		// Check that the user trip is rendered
		await expect(page.getByRole('heading', { name: 'My Personal Trip' })).toBeVisible();
		
		// The user trip should NOT have the INSPIRATION badge
		const userTripCard = page.locator('article', { hasText: 'My Personal Trip' });
		await expect(userTripCard.getByText('INSPIRATION')).not.toBeVisible();

		// Check that the Inspiration section title is rendered
		await expect(page.getByText('Inspiration', { exact: true })).toBeVisible();

		// Check that the demo trip is rendered
		await expect(page.getByRole('heading', { name: 'Japón Demo' })).toBeVisible();

		// The demo trip SHOULD have the INSPIRATION badge
		const demoTripCard = page.locator('article', { hasText: 'Japón Demo' });
		await expect(demoTripCard.getByText('INSPIRATION')).toBeVisible();
	});
});
