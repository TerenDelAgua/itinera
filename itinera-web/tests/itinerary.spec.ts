import { test, expect } from '@playwright/test';

test.describe('Itinerary Flow', () => {
  test('should create a new trip and navigate to its details', async ({ page }) => {
    // 1. Go to home
    await page.goto('/trips');
    
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

    // 5. Verify auto-navigation to details
    await expect(page).toHaveURL(/\/trips\/[0-9a-f-]+/, { timeout: 10000 });
    
    // 6. Verify the name in the header input
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
    await page.goto('/trips');
    
    // Wait for the page to load and cookies to be set
    await page.waitForLoadState('networkidle');
    
    const cookies = await context.cookies();
    const sessionCookie = cookies.find(c => c.name === 'session_id');
    
    expect(sessionCookie).toBeDefined();
    expect(sessionCookie?.value).toMatch(/[0-9a-f-]+/);
  });

  test('should only show local activities in place view', async ({ page }) => {
    await page.goto('/trips');
    await expect(page.locator('.animate-spin')).not.toBeVisible({ timeout: 15000 });
    
    // Create a trip
    const newTripBtn = page.getByRole('button', { name: /New trip/i });
    const firstTripBtn = page.getByText(/Create your first trip/i);
    if (await firstTripBtn.isVisible()) {
      await firstTripBtn.click();
    } else {
      await newTripBtn.click();
    }
    
    const tripName = `Filter Test ${Date.now()}`;
    await page.getByPlaceholder(/Trip name/i).fill(tripName);
    // 4. Submit and verify auto-navigation
    await page.locator('form button[type="submit"]').click();
    await expect(page).toHaveURL(/\/trips\/[0-9a-f-]+/, { timeout: 10000 });
    
    // Add a global activity via UpcomingActivityCard
    // The button has "+ Add" or similar text
    await page.getByTestId('add-activity-btn').first().click();
    
    const activityInput = page.getByTestId('activity-title-input');
    await expect(activityInput).toBeVisible();
    await activityInput.fill('Global Activity 1');
    
    // Wait for the POST to finish
    const postPromise = page.waitForResponse(resp => resp.url().includes('/activities') && resp.request().method() === 'POST');
    await activityInput.press('Enter');
    await postPromise;
    
    // Verify it's added
    await expect(page.getByText('Global Activity 1').first()).toBeVisible({ timeout: 5000 });

    // Create a place
    await page.getByTestId('add-place-button').click();
    const placeInput = page.getByTestId('place-name-input');
    await expect(placeInput).toBeVisible();
    const postPlacePromise = page.waitForResponse(resp => resp.url().includes('/places') && resp.request().method() === 'POST');
    await placeInput.fill('Test City');
    await placeInput.press('Enter');
    await postPlacePromise;
    
    // It should appear in the list
    const placeLink = page.getByText('Test City');
    await expect(placeLink).toBeVisible({ timeout: 5000 });

    // Navigate to place
    await placeLink.click();
    await expect(page).toHaveURL(/\/trips\/[0-9a-f-]+\/places\/[0-9a-f-]+/);
    
    // Verify Global Activity 1 is NOT visible in the UpcomingActivityCard
    await expect(page.getByText('Global Activity 1')).not.toBeVisible();
    
    // Open Activity Drawer
    await page.getByText(/View all/i).last().click();
    
    // Verify Global Activity 1 is NOT visible in the drawer either
    await expect(page.getByText('Global Activity 1')).not.toBeVisible();
  });
});
