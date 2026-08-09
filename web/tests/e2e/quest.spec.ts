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
    await startQuest(page, 'Bab 6: Gerbang Roda Gigi');
    await completeChallenge(page, 'Peran 1 (Builder): Gambar sebuah lingkaran roda gigi di kertas.');
    await verifyXpIncreased(page);
  });
});
