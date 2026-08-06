import { Page, expect } from '@playwright/test';

export async function verifyHomeLoaded(page: Page) {
  await expect(page.locator('text="Active Quests"').or(page.locator('text="Quests"'))).toBeVisible();
}

export async function verifyQuestExists(page: Page, questTitle: string) {
  await expect(page.locator(`text="${questTitle}"`)).toBeVisible();
}
