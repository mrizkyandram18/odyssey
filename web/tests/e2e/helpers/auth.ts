import { Page, expect } from '@playwright/test';

export async function login(page: Page, username: string = 'demo1') {
  await page.goto('/#/login');
  await page.fill('input[placeholder="e.g. demo1"]', username);
  await page.fill('input[type="password"]', 'odyssey123');
  await page.click('button:has-text("Sign In")');

  try {
    await expect(page).toHaveURL(/.*#\/$/, { timeout: 4000 });
  } catch {
    // Retry once after a 2-second delay to handle Vercel API rate limiting / cold start
    await page.waitForTimeout(2500);
    const signInBtn = page.locator('button:has-text("Sign In")');
    if (await signInBtn.isVisible()) {
      await signInBtn.click();
    }
    await expect(page).toHaveURL(/.*#\/$/, { timeout: 10000 });
  }
}

export async function logout(page: Page) {
  await page.goto('/#/profile');
  await page.click('button:has-text("Sign Out")');
  await expect(page).toHaveURL(/.*#\/login$/);
}
