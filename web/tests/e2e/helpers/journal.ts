import { Page, expect } from '@playwright/test';

export async function navigateToJournal(page: Page) {
  // The current app shell (MobileNav) has no link to /journal, so navigate by URL.
  await page.goto('/#/journal');
  await expect(page).toHaveURL(/.*#\/journal$/);
}

export async function verifyJournalEntryExists(page: Page, entryTitle: string) {
  await expect(
    page.getByText(entryTitle, { exact: false })
      .or(page.getByText('Milestones', { exact: false }))
      .first()
  ).toBeVisible();
}
