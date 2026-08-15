import { test, expect } from '@playwright/test';

test.describe('Team Booking Feature', () => {
  test.beforeEach(async ({ page }) => {
    // Navigate to the app and login
    await page.goto('http://localhost:5173');
    await page.fill('input[placeholder="Email"]', 'dinesh@test.com');
    await page.fill('input[placeholder="Password"]', 'password123');
    await page.click('button:has-text("Login")');
    await expect(page.locator('.welcome-pill')).toBeVisible();
  });

  test('should create a team booking successfully', async ({ page }) => {
    // 1. Go to explore and select a turf
    await page.click('.home-search-btn');
    await page.click('.turf-card:first-child');
    await page.click('button:has-text("BOOK NOW")');

    // 2. Select an available time slot
    await page.click('.slot:not(.booked):first-child');
    
    // 3. Click the TEAM SPLIT button
    await page.click('button:has-text("TEAM SPLIT")');

    // 4. Verify we are on the split share page
    await expect(page.locator('h2')).toHaveText('Team Session Created!');
    await expect(page.locator('text=The slot is held in PENDING state for 15 minutes.')).toBeVisible();
    
    // 5. Verify the shareable link was generated
    const shareLink = page.locator('div[style*="word-break: break-all"]');
    await expect(shareLink).toContainText('/split-pay/');
  });

  test('team booking should appear in My Bookings tab', async ({ page }) => {
    // Go directly to My Bookings tab
    await page.click('div.nav-bar div:has-text("Bookings")');
    
    // The team session should be listed as a PENDING slot
    // and should be sorted appropriately
    await expect(page.locator('text=PENDING').first()).toBeVisible();
  });
});
