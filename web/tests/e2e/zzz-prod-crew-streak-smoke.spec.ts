import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY || '';
const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;

/**
 * Slice 2.8 prod smoke — crew-level streak.
 * READ-ONLY: no writes, no historical mutation. Asserts the Home API exposes
 * crew_streak alongside the (unchanged) personal streak_days, and that the
 * Home UI renders the shared "Runtutan kru" progress line.
 */
test.describe('PROD smoke: crew-level streak (Slice 2.8)', () => {
  test('home exposes crew_streak + personal streak and UI renders crew line', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only crew streak smoke');
    test.setTimeout(240_000);

    await login(page, 'demo1');

    const home = await page.evaluate(async () => {
      const r = await fetch('/api/home', { credentials: 'same-origin' });
      return { status: r.status, body: await r.json() };
    }, undefined, { timeout: T });

    expect(home.status).toBe(200);

    const dt = home.body?.daily_mission as Record<string, unknown> | undefined;
    expect(dt, 'home daily_mission section must exist').toBeTruthy();
    expect(typeof dt?.streak_days).toBe('number');
    expect(typeof dt?.crew_streak, 'crew_streak must be present and numeric').toBe('number');
    expect((dt?.crew_streak as number) >= 0).toBe(true);

    // Sections parity (same value surfaced in sections.daily_mission).
    const sec = home.body?.sections?.daily_mission as Record<string, unknown> | undefined;
    expect(sec?.crew_streak).toBe(dt?.crew_streak);

    // UI renders the shared crew streak line on Home.
    await page.goto('/#/');
    await expect(page.locator('h1')).toContainText('Halo', { timeout: T });
    await expect(page.getByTestId('home-crew-streak')).toContainText(/Runtutan kru: \d+ hari bersama/, { timeout: 20_000 });

    console.log(
      '[PROD SMOKE]',
      JSON.stringify({
        status: home.status,
        streak_days: dt?.streak_days,
        crew_streak: dt?.crew_streak,
        sections_crew_streak: sec?.crew_streak,
      }),
    );
  });
});
