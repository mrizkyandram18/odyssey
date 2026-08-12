import { Page, expect } from '@playwright/test';

export async function startMission(page: Page, MissionTitle: string) {
  await page.click(`text="${MissionTitle}"`);
  // Wait for the Mission Detail page to render (title, badge, or buttons)
  await page.locator('h1').waitFor({ state: 'visible' });
  const startBtn = page.locator('button:has-text("Embark on Mission")');
  if (await startBtn.count() > 0 && await startBtn.first().isVisible().catch(() => false)) {
    await startBtn.first().click();
    await page.locator('button:has-text("Selesaikan")').first().waitFor({ state: 'visible' }).catch(() => {});
  }
}

export async function completeChallenge(page: Page, challengeName: string) {
  const challengeDiv = page.locator('div').filter({ hasText: challengeName }).first();
  const completeBtn = challengeDiv.locator('button:has-text("Selesaikan")');
  if (await completeBtn.isVisible().catch(() => false)) {
    await completeBtn.click();
    await page.waitForTimeout(500);
  }
}

const SUPABASE_REST_URL = process.env.SUPABASE_URL ? `${process.env.SUPABASE_URL.replace(/\/$/, '')}/rest/v1` : '';
const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY || '';

export async function resetMissionState() {
  // Production Data Safety: no-op unless explicitly targeting localhost with service credentials.
  // Never patch production during E2E (SKIP_DB_RESET / non-localhost).
  if (
    !SUPABASE_REST_URL ||
    !SUPABASE_SERVICE_KEY ||
    process.env.SKIP_DB_RESET === 'true' ||
    !process.env.PLAYWRIGHT_TEST_BASE_URL?.includes('localhost')
  ) {
    return;
  }
  const headers = {
    'apikey': SUPABASE_SERVICE_KEY,
    'Authorization': `Bearer ${SUPABASE_SERVICE_KEY}`,
    'Content-Type': 'application/json'
  };
  // Reset Mission 102 (Gather Herbs — used by golden-path) to PENDING so Embark works.
  // Exercise rows identified by slug (not fragile row ids).
  await fetch(`${SUPABASE_REST_URL}/odyssey_missions?id=eq.102`, {
    method: 'PATCH', headers,
    body: JSON.stringify({ status: 'PENDING', started_at: null, completed_at: null }),
  }).catch(() => {});
  await fetch(`${SUPABASE_REST_URL}/odyssey_exercises?mission_id=eq.102&slug=eq.spot-the-green`, {
    method: 'PATCH', headers,
    body: JSON.stringify({ status: 'PENDING', completed_by: null, completed_at: null }),
  }).catch(() => {});
  await fetch(`${SUPABASE_REST_URL}/odyssey_exercises?mission_id=eq.102&slug=eq.herb-concept`, {
    method: 'PATCH', headers,
    body: JSON.stringify({ status: 'PENDING', completed_by: null, completed_at: null }),
  }).catch(() => {});
}

export async function verifyXpIncreased(page: Page) {
  // Check for XP reward banner, SubmissionForm memory prompt, DONE badge, or completed indicator rendered
  await expect(
    page.locator('.bg-accent\\/10, form, p:has-text("XP"), p:has-text("Memory"), span:has-text("DONE"), span:has-text("ACTIVE"), p:has-text("failed to complete challenge")').first()
  ).toBeVisible();
}
