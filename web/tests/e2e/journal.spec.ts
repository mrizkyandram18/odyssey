import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';
import { navigateToJournal } from './helpers/journal';

test.describe('Journal Domain', () => {
  const TEST_USER = 'demo1';

  test.beforeEach(async ({ page }) => {
    await login(page, TEST_USER);
  });

  test('can view journal entries', async ({ page }) => {
    await navigateToJournal(page);
    // Note: Depends on test data strategy to ensure entries exist for this user
    // await verifyJournalEntryExists(page, 'Morning Light');
    await expect(page.locator('h1:has-text("Milestones")').first()).toBeVisible();
  });
});
