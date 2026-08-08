import { test } from '@playwright/test';
import { login } from './helpers/auth';
import { verifyHomeLoaded, verifyQuestExists } from './helpers/home';

test.describe('Home Domain', () => {
  const TEST_USER = 'demo1';

  test.beforeEach(async ({ page }) => {
    await login(page, TEST_USER);
  });

  test('displays home dashboard and quests', async ({ page }) => {
    await verifyHomeLoaded(page);
    await verifyQuestExists(page, 'Riddle of the Stones');
  });
});
