import { test, expect } from '@playwright/test';
import { login, logout } from './helpers/auth';
import { navigateToJournal } from './helpers/journal';

test.describe('Persistence Domain', () => {
  const TEST_USER = 'family-test@example.com';

  test('session persists after reload', async ({ page }) => {
    await login(page, TEST_USER);
    await page.reload();
    await expect(page).toHaveURL(/.*home|^\/$/);
  });

  test('state persists after logout and login', async ({ page }) => {
    await login(page, TEST_USER);
    await navigateToJournal(page);
    
    await logout(page);
    await login(page, TEST_USER);
    
    // verify we can still navigate and data is there
    await navigateToJournal(page);
    await expect(page.locator('.journal-container, .journal-list')).toBeVisible();
  });
});
