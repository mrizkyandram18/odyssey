import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY || '';
const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;

test.describe('PROD smoke: quest 102 completion flow (real demo path)', () => {
  test('completes herb-lore -> quest DONE -> COMPLETED badge -> reward -> persists -> replay idempotent', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only real completion smoke');
    test.setTimeout(180_000);

    const me = async () =>
      await page.evaluate(async () => {
        const r = await fetch('/api/me');
        const j = await r.json();
        return { xp: j.xp as number, coins: j.coins as number };
      }, undefined, { timeout: T });

    const postComplete = async () =>
      await page.evaluate(async () => {
        const r = await fetch('/api/quests/102/challenges/39/complete', {
          method: 'POST',
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/json' },
          body: '{}',
        });
        const body = await r.text();
        let json: any = {};
        try {
          json = JSON.parse(body);
        } catch {
          /* non-JSON */
        }
        return { status: r.status, quest_completed: json.quest_completed, xp: json.xp };
      }, undefined, { timeout: T });

    await login(page, 'demo1');
    await page.goto('/#/quests/102');
    await page.locator('h1').waitFor({ state: 'visible' });
    await expect(page.locator('h1')).toContainText('Gather Herbs');

    // Initial production state = ACTIVE quest with herb-lore actionable.
    await expect(page.locator('span:has-text("ACTIVE")')).toBeVisible();
    await expect(page.locator('button:has-text("Selesaikan")')).toHaveCount(1);

    const xp0 = await me();

    // Trigger the REAL completion via the UI button (mirrors real user action).
    const [resp] = await Promise.all([
      page.waitForResponse(
        (r) =>
          r.url().includes('/api/quests/102/challenges/') &&
          r.url().includes('/complete') &&
          r.request().method() === 'POST',
        { timeout: T },
      ),
      page.click('button:has-text("Selesaikan")'),
    ]);
    const bodyText = await resp.text();
    let json: any = {};
    try {
      json = JSON.parse(bodyText);
    } catch {
      /* non-JSON */
    }
    expect(resp.status()).toBe(200);
    expect(json.quest_completed).toBe(true);

    // Quest transitioned to DONE -> actual badge text COMPLETED.
    await expect(page.locator('span:has-text("COMPLETED")')).toBeVisible({ timeout: 20_000 });
    await expect(page.locator('span:has-text("ACTIVE")')).toBeHidden({ timeout: 10_000 });

    // Reward granted exactly once.
    const xp1 = await me();
    expect(xp1.xp).toBe(xp0.xp + (json.xp ?? 0));

    // Persistence across refresh.
    await page.reload();
    await expect(page.locator('span:has-text("COMPLETED")')).toBeVisible({ timeout: 20_000 });

    // Replay idempotent: no duplicate quest_completed / no XP delta.
    const replay = await postComplete();
    expect(replay.status).toBe(200);
    expect(replay.quest_completed).toBe(false);
    const xp2 = await me();
    expect(xp2.xp).toBe(xp1.xp);
    expect(xp2.coins).toBe(xp1.coins);

     
    console.log('[PROD SMOKE]', JSON.stringify({ initial: xp0, after: xp1, replay: xp2, completion: json }));
  });
});
