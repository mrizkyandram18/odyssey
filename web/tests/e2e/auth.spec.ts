import { test, expect } from '@playwright/test';
import { login, logout } from './helpers/auth';

test.describe('Auth Domain', () => {
  const TEST_USER = 'family-test@example.com';

  test('can login successfully', async ({ page }) => {
    await login(page, TEST_USER);
  });

  test('can logout successfully', async ({ page }) => {
    await login(page, TEST_USER);
    await logout(page);
  });
});
