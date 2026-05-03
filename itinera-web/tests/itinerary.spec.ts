import { test, expect } from '@playwright/test';

test.describe('Itinerary Flow', () => {
  test('should create a new trip and navigate to its details', async ({ page }) => {
    // 1. Go to home
    await page.goto('/');
    
    // Wait for loading to finish
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 15000 });

    // Check title (from <title> in app.html or similar)
    // We'll just check for a header instead
    await expect(page.getByText(/My trips/i)).toBeVisible();

    // 2. Open Create Trip Form
    const newTripBtn = page.getByRole('button', { name: /New trip/i });
    const firstTripBtn = page.getByText(/Create your first trip/i);
    
    if (await firstTripBtn.isVisible()) {
      await firstTripBtn.click();
    } else {
      await newTripBtn.click();
    }

    // 3. Fill the form
    const tripName = `E2E Trip ${Date.now()}`;
    const nameInput = page.getByPlaceholder(/Trip name/i);
    await expect(nameInput).toBeVisible({ timeout: 5000 });
    await nameInput.fill(tripName);
    
    // 4. Submit
    const submitBtn = page.locator('form button[type="submit"]');
    await submitBtn.click();

    // 5. Verify it appears in the list
    const tripCard = page.getByText(tripName);
    await expect(tripCard).toBeVisible({ timeout: 10000 });

    // 6. Navigate to details
    await tripCard.click();
    
    // 7. Verify we are on the trip page
    await expect(page).toHaveURL(/\/trips\/[0-9a-f-]+/);
    // The name is in an input field in the header
    const titleInput = page.getByPlaceholder(/Trip Name/i);
    await expect(titleInput).toHaveValue(tripName, { timeout: 10000 });

    // 8. Cleanup (Optional but good practice)
    await page.goBack();
    // Find the delete button for this specific trip
    // The TripCard has a delete button (ConfirmModal logic)
    // We'll skip cleanup for now to ensure the first part works, 
    // but in a real CI we'd want a clean DB.
  });

  test('should persist session via cookie', async ({ page, context }) => {
    await page.goto('/');
    
    // Wait for the page to load and cookies to be set
    await page.waitForLoadState('networkidle');
    
    const cookies = await context.cookies();
    const sessionCookie = cookies.find(c => c.name === 'session_id');
    
    expect(sessionCookie).toBeDefined();
    expect(sessionCookie?.value).toMatch(/[0-9a-f-]+/);
  });
});
