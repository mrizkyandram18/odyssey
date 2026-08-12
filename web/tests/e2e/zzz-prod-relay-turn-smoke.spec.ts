import { test, expect } from '@playwright/test';
import { login, logout } from './helpers/auth';

const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY || '';
const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;
const UID_TO_USER: Record<string, string> = { 'demo-uid-1': 'demo1', 'demo-uid-2': 'demo2', 'demo-uid-3': 'demo3' };

test.describe('PROD smoke: relay quest Your Turn handoff badge (Slice 2.5)', () => {
  test('list + detail expose the assignee; badge shows only for the assigned explorer', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only relay handoff smoke');
    test.setTimeout(240_000);

    const fetchJson = async (url: string) =>
      await page.evaluate(async (u) => {
        const r = await fetch(u, { credentials: 'same-origin' });
        return { status: r.status, body: await r.json() };
      }, url, { timeout: T });

    await login(page, 'demo1');

    // Slice 2.5 backend: list + detail expose the relay handoff assignee.
    const home = await fetchJson('/api/home');
    expect(home.status).toBe(200);
    const relay = (home.body.quests ?? []).find((q: any) => q.template_slug === 'shadow-trail');
    expect(relay, 'shadow-trail relay quest must exist').toBeTruthy();
    expect(relay.quest_type).toBe('RELAY');
    expect(relay.status).toBe('ACTIVE');
    expect(relay.active_challenge_assigned_to, 'list view must expose active_challenge_assigned_to').toBeTruthy();

    const detail = await fetchJson('/api/quests/104');
    expect(detail.status).toBe(200);
    expect(detail.body.challenges, 'detail must include challenges').toBeTruthy();
    expect(detail.body.active_challenge_assigned_to, 'detail must expose active_challenge_assigned_to').toBeTruthy();
    expect(detail.body.active_challenge_assigned_to).toBe(relay.active_challenge_assigned_to);

    const assignee = relay.active_challenge_assigned_to as string;
    const assigneeUser = UID_TO_USER[assignee];
    expect(assigneeUser, `assignee ${assignee} must map to a demo login`).toBeTruthy();

    // The assigned explorer sees the badge; a non-assigned explorer does not.
    if (assignee !== 'demo-uid-1') {
      await page.goto('/#/');
      await expect(page.locator('h1')).toContainText('Halo', { timeout: T });
      await expect(page.locator('h3:has-text("Shadow Trail")')).toBeVisible({ timeout: 15_000 });
      await expect(page.locator('text=Your Turn').first()).not.toBeVisible({ timeout: 15_000 });
      await logout(page);
    }

    if (assigneeUser !== 'demo1') {
      await login(page, assigneeUser);
    }

    // Home quest card carries the badge.
    await page.goto('/#/');
    await expect(page.locator('h1')).toContainText('Halo', { timeout: T });
    const shadowCard = page.locator('h3:has-text("Shadow Trail")').locator('xpath=ancestor::div[contains(@class,"cursor-pointer")]');
    await expect(shadowCard).toBeVisible({ timeout: 15_000 });
    await expect(shadowCard).toContainText('Your Turn', { timeout: 15_000 });

    // Quest detail header carries the badge too.
    await page.goto('/#/quests/104');
    await expect(page.locator('h1')).toContainText('Shadow Trail', { timeout: T });
    await expect(page.locator('text=Your Turn').first()).toBeVisible({ timeout: 15_000 });

     
    console.log('[PROD SMOKE]', JSON.stringify({ relay_quest: relay.template_slug, assignee, assignee_user: assigneeUser, detail_assignee: detail.body.active_challenge_assigned_to }));
  });
});
