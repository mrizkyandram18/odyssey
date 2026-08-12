import { test, expect } from '@playwright/test';
import { login, logout } from './helpers/auth';

const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;
const UID_TO_NAME: Record<string, string> = { 'demo-uid-1': 'Leo', 'demo-uid-2': 'Maya', 'demo-uid-3': 'Sam' };
const UID_TO_USER: Record<string, string> = { 'demo-uid-1': 'demo1', 'demo-uid-2': 'demo2', 'demo-uid-3': 'demo3' };

test.describe('PROD smoke: relay rotation UI (Slice 2.6)', () => {
  test('detail payload carries crew roster; rotation panel shows active/done/next with names', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!process.env.SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only relay rotation smoke');
    test.setTimeout(240_000);

    const fetchJson = async (url: string) =>
      await page.evaluate(async (u) => {
        const r = await fetch(u, { credentials: 'same-origin' });
        return { status: r.status, body: await r.json() };
      }, url, { timeout: T });

    await login(page, 'demo1');

    // Slice 2.6 backend: detail exposes the crew roster (names) plus the
    // per-leg assignment already used by the rotation panel.
    const detail = await fetchJson('/api/quests/104');
    expect(detail.status).toBe(200);
    expect(detail.body.quest_type).toBe('RELAY');
    const members = detail.body.members as Array<{ uid: string; explorer_name: string }> | undefined;
    expect(members, 'detail must expose crew roster for name resolution').toBeTruthy();
    const names = members!.map((m) => m.explorer_name);
    expect(names).toEqual(expect.arrayContaining(['Leo', 'Maya', 'Sam']));
    const assigned = detail.body.active_challenge_assigned_to as string | undefined;
    expect(assigned, 'relay quest must have an active assignee').toBeTruthy();
    const assignedLeg = (detail.body.challenges as Array<{ status: string; assigned_to: string | null }>)
      .find((c) => c.status === 'PENDING' && c.assigned_to === assigned);
    expect(assignedLeg, 'the active assignee must own a PENDING leg').toBeTruthy();

    // Rotation panel renders with resolved names on the detail page.
    await page.goto('/#/quests/104');
    await expect(page.locator('h1')).toContainText('Shadow Trail', { timeout: T });
    const panel = page.locator('section[aria-label="Relay rotation"]');
    await expect(panel).toBeVisible({ timeout: 15_000 });
    await expect(panel).toContainText('Crew Relay');
    await expect(panel).toContainText(UID_TO_NAME[assigned as string] + "'s turn", { timeout: 15_000 });
    await expect(panel).not.toContainText(assigned as string, { timeout: 15_000 });

    // The "Your turn" chip appears only for the assigned explorer.
    const assignedUser = UID_TO_USER[assigned as string];
    expect(assignedUser, `assignee ${assigned} must map to a demo login`).toBeTruthy();
    if (assignedUser !== 'demo1') {
      await logout(page);
      await login(page, assignedUser);
      await page.goto('/#/quests/104');
      await expect(page.locator('section[aria-label="Relay rotation"]')).toBeVisible({ timeout: T });
      await expect(page.locator('section[aria-label="Relay rotation"]')).toContainText('Your turn', { timeout: 15_000 });
    }

     
    console.log('[PROD SMOKE]', JSON.stringify({ relay_quest: detail.body.template_slug, assignee: assigned, members: names }));
  });
});
