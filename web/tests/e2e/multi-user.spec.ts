import { test, expect } from '@playwright/test';
import { login, logout } from './helpers/auth';
import { verifyHomeLoaded } from './helpers/home';

test.describe('Multi-user Auth Verification', () => {
  const USERS = [
    { username: 'demo1', explorer: 'Leo', role: 'SEEKER' },
    { username: 'demo2', explorer: 'Maya', role: 'GUIDE' },
    { username: 'demo3', explorer: 'Sam', role: 'BUILDER' },
  ];

  for (const { username } of USERS) {
    test(`can login as ${username} and see dashboard`, async ({ page }) => {
      await login(page, username);
      await verifyHomeLoaded(page);
      await expect(page.locator('text=/Halo, /')).toBeVisible();
      await logout(page);
    });
  }
});
