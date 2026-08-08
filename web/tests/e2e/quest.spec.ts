import { test } from '@playwright/test';
import { login } from './helpers/auth';
import { verifyHomeLoaded } from './helpers/home';
import { startQuest, completeChallenge, verifyXpIncreased, resetQuestState } from './helpers/quest';

test.describe('Quest Domain', () => {
  const TEST_USER = 'demo1';

  test.beforeEach(async ({ page }) => {
    await resetQuestState();
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);
  });

  test('can complete a challenge in a quest', async ({ page }) => {
    await startQuest(page, 'Riddle of the Stones');
    await completeChallenge(page, 'Find a stone or brick and describe its shape.');
    await verifyXpIncreased(page);
  });
});
