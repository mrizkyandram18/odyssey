import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

test.describe('Security & Authorization Boundary Journeys', () => {
  test('Journey 12a: Unauthenticated user is redirected to login from protected routes', async ({ page }) => {
    // Navigate directly to protected route
    await page.goto('/#/profile');
    await expect(page).toHaveURL(/.*#\/login$/);
    
    await page.goto('/#/shop');
    await expect(page).toHaveURL(/.*#\/login$/);
    
    await page.goto('/#/admin');
    await expect(page).toHaveURL(/.*#\/login$/);
  });

  test('Journey 12b: Member role cannot access Guide / Admin dashboard', async ({ page }) => {
    // Login as member (SEEKER role)
    await login(page, 'user_testing');
    await expect(page).toHaveURL(/.*#\/$/);
    
    // BottomNav should NOT have Admin link
    const adminLink = page.locator('nav a:has-text("Admin"), a[href="#/admin"]');
    await expect(adminLink).not.toBeVisible();
    
    // Attempt direct URL manipulation to /#/admin
    await page.goto('/#/admin');
    
    // Member should be immediately redirected away from /admin to /
    await expect(page).toHaveURL(/.*#\/$/);
  });
});
