import { test } from '@playwright/test';
import { login } from './helpers/auth';
import { verifyHomeLoaded } from './helpers/home';
import { startQuest, completeChallenge, verifyXpIncreased } from './helpers/quest';

test.describe('Quest Domain', () => {
  const TEST_USER = 'family-test@example.com';

  test.beforeEach(async ({ page }) => {
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);
  });

  test('can complete a challenge in a quest', async ({ page }) => {
    await startQuest(page, 'Morning Light');
    await completeChallenge(page, 'Spot the Green');
    await verifyXpIncreased(page);
  });
});
