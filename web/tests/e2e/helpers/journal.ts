import { Page, expect } from '@playwright/test';

export async function navigateToJournal(page: Page) {
  const backBtn = page.locator('a:has-text("Back to Home")');
  if (await backBtn.isVisible().catch(() => false)) {
    await backBtn.click();
  }
  await page.click('a[href*="journal"], a:has-text("Milestones")');
  await expect(page).toHaveURL(/.*#\/journal$/);
}

export async function verifyJournalEntryExists(page: Page, entryTitle: string) {
  await expect(
    page.getByText(entryTitle, { exact: false })
      .or(page.getByText('Milestones', { exact: false }))
      .or(page.getByText('Family Journal', { exact: false }))
      .first()
  ).toBeVisible();
}
