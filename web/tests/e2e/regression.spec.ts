import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';
import { verifyHomeLoaded } from './helpers/home';

test.describe('Regression: Routing & Navigation', () => {
  const TEST_USER = 'demo1';

  test('deep-link to /#/quests/103 after login', async ({ page }) => {
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);

    await page.goto('/#/quests/103');
    await expect(page).toHaveURL(/.*#\/quests\/103$/);
    await expect(page.locator('h1')).toContainText('Bab 3: Melodi dari Dedaunan');
  });

  test('browser back and forward navigation', async ({ page }) => {
    await login(page, TEST_USER);
    await verifyHomeLoaded(page);

    await page.goto('/#/quests/103');
    await expect(page).toHaveURL(/.*#\/quests\/103$/);

    await page.goBack();
    await expect(page).toHaveURL(/.*#\/$/);

    await page.goForward();
    await expect(page).toHaveURL(/.*#\/quests\/103$/);
  });

  test('unauthenticated deep-link redirects to login', async ({ page }) => {
    await page.goto('/#/journal');
    await expect(page).toHaveURL(/.*#\/login$/);
  });
});
