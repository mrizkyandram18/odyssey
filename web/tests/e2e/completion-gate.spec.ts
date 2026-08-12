import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

const SUPABASE_REST_URL = process.env.SUPABASE_URL ? `${process.env.SUPABASE_URL.replace(/\/$/, '')}/rest/v1` : '';
const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY || '';

async function resetMission102ToCanonical() {
  if (!SUPABASE_REST_URL || !SUPABASE_SERVICE_KEY) return;
  const h = { apikey: SUPABASE_SERVICE_KEY, Authorization: `Bearer ${SUPABASE_SERVICE_KEY}`, 'Content-Type': 'application/json' };
  const patch = (path: string, body: object) => fetch(`${SUPABASE_REST_URL}/${path}`, { method: 'PATCH', headers: h, body: JSON.stringify(body) }).catch(() => {});
  const now = new Date().toISOString();
  await patch('odyssey_missions?id=eq.102', { status: 'ACTIVE', completed_at: null });
  await patch('odyssey_exercises?mission_id=eq.102&slug=eq.spot-the-green', { status: 'DONE', completed_by: 'demo-uid-3', completed_at: now });
  await patch('odyssey_exercises?mission_id=eq.102&slug=eq.herb-concept', { status: 'PENDING', completed_by: null, completed_at: null });
}

async function meXP(page: { evaluate: any }) {
  return await page.evaluate(async () => {
    const r = await fetch('/api/me');
    const j = await r.json();
    return j.xp as number;
  });
}

async function replayComplete(page: { evaluate: any }) {
  return await page.evaluate(async () => {
    const r = await fetch('/api/missions/102/exercises/39/complete', {
      method: 'POST', credentials: 'same-origin', headers: { 'Content-Type': 'application/json' }, body: '{}',
    });
    const body = await r.text();
    let json: any = {};
    try { json = JSON.parse(body); } catch { /* non-JSON bodies (e.g. redirects) */ }
    return { status: r.status, Mission_completed: json.Mission_completed, next_action: json.next_action, xp: json.xp };
  });
}

test.describe('Completion Gate (Mission 102 → DONE, exactly-once, journey unlock fix)', () => {
  test('completing herb-concept finalizes Mission DONE, then replay grants no duplicate reward', async ({ page }) => {
    if (!SUPABASE_REST_URL || !SUPABASE_SERVICE_KEY || !process.env.PLAYWRIGHT_TEST_BASE_URL?.includes('localhost')) {
      test.skip('only runs against localhost demo DB');
    }
    await resetMission102ToCanonical();
    await login(page, 'demo1');
    await page.goto('/#/missions/102');
    await page.locator('h1').waitFor({ state: 'visible' });
    await expect(page.locator('h1')).toContainText('Gather Herbs');

    const completeStatuses: number[] = [];
    page.on('response', (r) => {
      if (r.url().includes('/api/missions/102/exercises/') && r.url().includes('/complete') && r.request().method() === 'POST') {
        completeStatuses.push(r.status());
      }
    });

    const xp0 = await meXP(page);

    const buttons = page.locator('button:has-text("Selesaikan")');
    await expect(buttons).toHaveCount(1, { timeout: 10_000 });
    await buttons.first().click();
    await page.waitForTimeout(1500);

    // Final challenge completes the Mission -> no 500 (journey unlock via UPSERT fix).
    expect(completeStatuses).toContain(200);
    await expect(page.locator('span:has-text("COMPLETED")')).toBeVisible({ timeout: 15_000 });
    await expect(page.locator('.text-accent-danger')).toBeHidden();

    const xp1 = await meXP(page);
    expect(xp1).toBeGreaterThan(xp0);

    // Replay: idempotent — no duplicate XP, Mission already completed.
    const replay = await replayComplete(page);
    expect(replay.status).toBe(200);
    expect(replay.Mission_completed).toBe(false);
    const xp2 = await meXP(page);
    expect(xp2).toBe(xp1);
  });
});
