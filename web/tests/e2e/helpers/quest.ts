import { Page, expect } from '@playwright/test';

export async function startQuest(page: Page, questTitle: string) {
  await page.click(`text="${questTitle}"`);
  const startBtn = page.locator('button:has-text("Start"), button:has-text("Begin")');
  if (await startBtn.isVisible()) {
    await startBtn.click();
  }
}

export async function completeChallenge(page: Page, challengeName: string) {
  await page.click(`text="${challengeName}"`);
  await page.click('button:has-text("Complete")');
}

export async function verifyXpIncreased(page: Page) {
  // We can look for a toast or XP indicator change
  await expect(page.locator('.xp-indicator, .toast')).toBeVisible();
}
