import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';
import { getRoleMastery } from '../../src/shared/utils/roleMastery';

const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY || '';
const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;

/**
 * Slice prod smoke — role mastery flavor text.
 * READ-ONLY: no writes, no XP/coin/stat mutation. Verifies the Profile hero
 * renders the mastery title + flavor derived from the user's role + level,
 * stays stable across refresh, and that merely visiting Profile changes
 * nothing on the explorer (coins/XP/role/level identical before/after).
 */
test.describe('PROD smoke: role mastery flavor text (read-only)', () => {
  test('profile renders mastery title+flavor per role/level; stable on refresh; no stat changes', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only role mastery smoke');
    test.setTimeout(240_000);

    await login(page, 'demo1');

    const me = async () =>
      await page.evaluate(async () => {
        const r = await fetch('/api/me', { credentials: 'same-origin' });
        const j = await r.json();
        return { role: j.role, level: j.level, xp: j.xp, coins: j.coins, name: j.explorer_name };
      }, undefined, { timeout: T });

    const before = await me();
    expect(typeof before.level).toBe('number');
    expect(typeof before.role).toBe('string');

    const mastery = getRoleMastery(before.role, before.level);
    expect(typeof mastery.title).toBe('string');
    expect(typeof mastery.flavor).toBe('string');

    await page.goto('/#/profile');
    await expect(page.getByText(before.name, { exact: false }).first()).toBeVisible({ timeout: T });
    await expect(page.getByText(mastery.title).first()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(mastery.flavor).first()).toBeVisible({ timeout: 20_000 });
    await expect(page.getByText(`Level ${before.level}`, { exact: false }).first()).toBeVisible({ timeout: 20_000 });

    // Refresh -> mastery stays consistent.
    await page.reload();
    await expect(page.getByText(mastery.title).first()).toBeVisible({ timeout: T });
    await expect(page.getByText(mastery.flavor).first()).toBeVisible({ timeout: T });

    // Read-only: profile visit must not mutate XP / coins / role / level.
    const after = await me();
    expect(after.xp).toBe(before.xp);
    expect(after.coins).toBe(before.coins);
    expect(after.role).toBe(before.role);
    expect(after.level).toBe(before.level);

    // eslint-disable-next-line no-console
    console.log('[PROD SMOKE]', JSON.stringify({ role: before.role, level: before.level, mastery, before, after }));
  });
});
