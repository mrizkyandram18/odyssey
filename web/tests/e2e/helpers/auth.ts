import { Page, expect } from '@playwright/test';

export async function login(page: Page, username: string = 'demo1') {
  await page.goto('/#/login');
  await page.fill('input[placeholder="Enter your username"]', username);
  await page.fill('input[type="password"]', 'odyssey123');
  await page.click('button:has-text("Begin Adventure")');

  try {
    await expect(page).toHaveURL(/.*#\/$/, { timeout: 4000 });
  } catch {
    // Retry once after a 2-second delay to handle Vercel API rate limiting / cold start
    await page.waitForTimeout(2500);
    const signInBtn = page.locator('button:has-text("Begin Adventure")');
    if (await signInBtn.isVisible() && await signInBtn.isEnabled()) {
      await signInBtn.click({ timeout: 2000 }).catch(() => {});
    }
    await expect(page).toHaveURL(/.*#\/$/, { timeout: 10000 });
  }
}

export async function logout(page: Page) {
  await page.goto('/#/profile');
  await page.click('button:has-text("Leave Journey (Sign Out)")');
  await expect(page).toHaveURL(/.*#\/login$/);
}
