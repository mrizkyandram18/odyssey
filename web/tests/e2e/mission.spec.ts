import { test } from '@playwright/test';
import { login } from './helpers/auth';
import { verifyHomeLoaded } from './helpers/home';
import { startMission, completeChallenge, verifyXpIncreased, resetMissionState } from './helpers/Mission';

test.describe('Mission Domain', () => {
  const TEST_USER = 'demo1';

  test.beforeEach(async ({ page }) => {
    await resetMissionState();
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);
  });

  test('can complete a challenge in a Mission', async ({ page }) => {
    await startMission(page, 'Bab 6: Gerbang Roda Gigi');
    await completeChallenge(page, 'Peran 1 (Builder): Gambar sebuah lingkaran roda gigi di kertas.');
    await verifyXpIncreased(page);
  });
});
