import { Page, expect } from '@playwright/test';

export async function navigateToJournal(page: Page) {
  await page.click('a:has-text("Journal"), button:has-text("Journal")');
  await expect(page).toHaveURL(/.*journal/);
}

export async function verifyJournalEntryExists(page: Page, entryTitle: string) {
  await expect(page.locator(`text="${entryTitle}"`)).toBeVisible();
}
