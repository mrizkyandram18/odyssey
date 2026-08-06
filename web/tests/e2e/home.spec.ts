import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';
import { verifyHomeLoaded, verifyQuestExists } from './helpers/home';

test.describe('Home Domain', () => {
  const TEST_USER = 'family-test@example.com';

  test.beforeEach(async ({ page }) => {
    await login(page, TEST_USER);
  });

  test('displays home dashboard and quests', async ({ page }) => {
    await verifyHomeLoaded(page);
    await verifyQuestExists(page, 'Morning Light');
  });
});
