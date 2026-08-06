import { Page, expect } from '@playwright/test';

export async function login(page: Page, email: string = 'family-test@example.com') {
  await page.goto('/login');
  // Adjust these locators to match the Gatekeeper / mock adapter flow
  await page.fill('input[type="email"]', email);
  await page.click('button[type="submit"]');
  await expect(page).toHaveURL(/.*home|^\/$/);
}

export async function logout(page: Page) {
  await page.click('button:has-text("Logout"), a:has-text("Logout")');
  await expect(page).toHaveURL(/.*login/);
}
