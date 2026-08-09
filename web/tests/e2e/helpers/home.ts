import { Page, expect } from '@playwright/test';

/** Production /api/home is a heavy serverless aggregate (~8s warm; longer on cold start). */
const HOME_LOAD_TIMEOUT_MS = 45_000;

export async function verifyHomeLoaded(page: Page) {
  // Deterministic sync for production cold starts:
  // - If /api/home is still in flight, wait for it.
  // - If it already finished (Loading already gone), do not hang waiting for a past response.
  await Promise.race([
    page
      .waitForResponse(
        (res) =>
          res.url().includes('/api/home') &&
          res.request().method() === 'GET',
        { timeout: HOME_LOAD_TIMEOUT_MS },
      )
      .catch(() => null),
    page
      .locator('text="Memuat dunia..."')
      .waitFor({ state: 'hidden', timeout: HOME_LOAD_TIMEOUT_MS })
      .catch(() => null),
  ]);

  await expect(page.locator('text="Memuat dunia..."')).toBeHidden({ timeout: HOME_LOAD_TIMEOUT_MS });
  await expect(
    page.locator('text="Buku Misi Aktif"').or(page.locator('text="Quests"')),
  ).toBeVisible({ timeout: HOME_LOAD_TIMEOUT_MS });
}

export async function verifyQuestExists(page: Page, questTitle: string) {
  await expect(page.locator(`text="${questTitle}"`)).toBeVisible();
}
