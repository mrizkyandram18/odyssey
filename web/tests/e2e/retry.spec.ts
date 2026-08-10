import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

// Gate scenario 1 (retry): Home must not stay blank forever, and the user
// must be able to recover from a failed /api/home request.
test.describe('Home Retry Behaviour', () => {
  const TEST_USER = 'demo1';

  test('shows retry UI when /api/home fails, and recovers after Coba Lagi', async ({ page }) => {
    await login(page, TEST_USER);

    // Fail the first /api/home call.
    let failed = false;
    await page.route('**/api/home', async (route) => {
      if (!failed) {
        failed = true;
        await route.abort();
      } else {
        await route.continue();
      }
    });

    await page.goto('/#/');
    await page.reload();
    await expect(page.locator('.text-accent-danger')).toBeVisible({ timeout: 30_000 });

    // Retry succeeds and Home renders quests.
    await page.click('button:has-text("Coba Lagi")');
    await expect(page.locator('text="Buku Misi Aktif"').or(page.locator('text="Quests"'))).toBeVisible({ timeout: 45_000 });
  });
});