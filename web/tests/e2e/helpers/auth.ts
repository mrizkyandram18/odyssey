import { Page, expect } from '@playwright/test';

export async function login(page: Page, username: string = 'demo1') {
  await page.addInitScript(() => {
    localStorage.setItem('odyssey_onboarded', 'true');
  });
  await page.goto('/#/login');
  await page.fill('input[autocomplete="username"]', username);
  await page.fill('input[type="password"]', 'odyssey123');
  await page.click('button:has-text("Mulai Petualangan")');

  try {
    await expect(page).toHaveURL(/.*#\/$/, { timeout: 4000 });
  } catch {
    // Retry once after a 2-second delay to handle Vercel API rate limiting / cold start
    await page.waitForTimeout(2500);
    const signInBtn = page.locator('button:has-text("Mulai Petualangan")');
    if (await signInBtn.isVisible() && await signInBtn.isEnabled()) {
      await signInBtn.click({ timeout: 2000 }).catch(() => {});
    }
    await expect(page).toHaveURL(/.*#\/$/, { timeout: 10000 });
  }
}

export async function logout(page: Page) {
  // Navigate via bottom nav or direct hash
  const profileLink = page.locator('nav a:has-text("Profil"), a[href="#/profile"]');
  if (await profileLink.isVisible()) {
    await profileLink.click();
  } else {
    await page.goto('/#/profile');
  }
  
  // Wait for settings button in header
  const settingsBtn = page.locator('header button:has-text("Pengaturan")');
  await settingsBtn.waitFor({ state: 'visible', timeout: 10000 });
  await settingsBtn.click();
  
  // Click logout button
  const logoutBtn = page.locator('button:has-text("Keluar (Sign Out)")');
  await logoutBtn.waitFor({ state: 'visible', timeout: 5000 });
  await logoutBtn.click();
  
  await expect(page).toHaveURL(/.*#\/login$/);
}
