import { test, expect } from '@playwright/test';
import { login, logout } from './helpers/auth';

test.describe('Auth and Session Journeys', () => {
  test('Journey 1 & 2: Member can login and logout successfully', async ({ page }) => {
    await login(page, 'demo1');
    await expect(page).toHaveURL(/.*#\/$/);
    await logout(page);
    await expect(page).toHaveURL(/.*#\/login$/);
  });

  test('Journey 7: Admin can login and access guide privileges', async ({ page }) => {
    await login(page, 'demo2');
    await expect(page).toHaveURL(/.*#\/$/);
    
    // Guide role should have Admin navigation item
    const adminLink = page.locator('nav a:has-text("Admin"), a[href="#/admin"]');
    await expect(adminLink).toBeVisible();
    await adminLink.click();
    await expect(page).toHaveURL(/.*#\/admin$/);
    await expect(page.locator('h1:has-text("Admin Panel Keluarga")')).toBeVisible();
  });
});
