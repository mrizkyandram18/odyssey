import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;
const tinyPng = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==';

interface MeSnapshot {
  xp: number;
  coins: number;
}

// Auth is session-cookie based on prod (Gatekeeper). Reads player xp/coins.
const fetchMe = async (page: import('@playwright/test').Page): Promise<MeSnapshot | null> =>
  await page.evaluate(async () => {
    const sessionRaw = localStorage.getItem('odyssey_session');
    const session = sessionRaw ? JSON.parse(sessionRaw) : null;
    const token = session?.token as string | undefined;
    const headers: Record<string, string> = {};
    if (token) {
      headers.Authorization = `Bearer ${token}`;
      headers['X-User-Session'] = token;
    }
    const r = await fetch('/api/me', { credentials: 'include', headers });
    if (!r.ok) return null;
    const body = (await r.json()) as { xp?: number; coins?: number };
    return { xp: body.xp ?? 0, coins: body.coins ?? 0 };
  });

/**
 * Slice 2.9 prod smoke — PHOTO submission kind.
 * Uses quest 102 (DONE) with the last challenge (39) for the post-completion
 * memory-capture path, matching the COMIC smoke.
 */
test.describe('PROD smoke: photo submission (Slice 2.9)', () => {
  test('rejects non-image photo; accepts photo; timeline renders caption', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!process.env.SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only photo smoke');
    test.setTimeout(240_000);

    await login(page, 'demo1');

    // Snapshot xp/coins before the smoke (no reward/XP/coin change expected
    // from a pending-review PHOTO submission).
    const meBefore = await fetchMe(page);
    if (!meBefore) {
      test.fail(true, '/api/me returned no player snapshot');
    }

    const postPhoto = async (photo: string, caption: string) =>
      await page.evaluate(
        async ({ photo: photoData, caption: cap }) => {
          const sessionRaw = localStorage.getItem('odyssey_session');
          const session = sessionRaw ? JSON.parse(sessionRaw) : null;
          const token = session?.token as string | undefined;

          const csrfResp = await fetch('/api/csrf', { credentials: 'include' });
          const csrfJson = (await csrfResp.json()) as { csrf_token?: string };
          const headers: Record<string, string> = {
            'Content-Type': 'application/json',
          };
          if (token) {
            headers.Authorization = `Bearer ${token}`;
            headers['X-User-Session'] = token;
          }
          if (csrfJson.csrf_token) {
            headers['X-CSRF-Token'] = csrfJson.csrf_token;
          }

          const r = await fetch('/api/creative', {
            method: 'POST',
            credentials: 'include',
            headers,
            body: JSON.stringify({
              quest_id: 102,
              challenge_id: 39,
              kind: 'PHOTO',
              content: JSON.stringify({ v: 1, photo: photoData, caption: cap }),
            }),
          });
          const text = await r.text();
          let json: Record<string, unknown> = {};
          try {
            json = JSON.parse(text);
          } catch {
            /* non-JSON */
          }
          return { status: r.status, body: json, text };
        },
        { photo, caption },
        { timeout: T },
      );

    // Invalid: non-image data URI -> 400
    const bad = await postPhoto('data:text/plain;base64,aGk=', 'not an image');
    expect(bad.status).toBe(400);
    expect(String(bad.body.error || bad.text)).toMatch(/photo/i);

    // Valid: tiny PNG photo with a unique caption
    const marker = `Slice29 photo ${Date.now()}`;
    const good = await postPhoto(`data:image/png;base64,${tinyPng}`, marker);
    expect(good.status, `submit photo failed: ${good.text}`).toBe(201);
    expect(good.body.kind === 'PHOTO' || (good.body as { Kind?: string }).Kind === 'PHOTO').toBeTruthy();

    // List by crew includes the new photo
    const list = await page.evaluate(async () => {
      const sessionRaw = localStorage.getItem('odyssey_session');
      const session = sessionRaw ? JSON.parse(sessionRaw) : null;
      const token = session?.token as string | undefined;
      const headers: Record<string, string> = {};
      if (token) {
        headers.Authorization = `Bearer ${token}`;
        headers['X-User-Session'] = token;
      }
      const r = await fetch('/api/creative', { credentials: 'include', headers });
      const body = await r.json();
      return { status: r.status, body };
    }, undefined, { timeout: T });
    expect(list.status).toBe(200);
    const rows = list.body as Array<{ kind?: string; Kind?: string; content?: string; Content?: string }>;
    expect(Array.isArray(rows)).toBe(true);
    const found = rows.some((s) => {
      const kind = s.kind || s.Kind;
      const content = s.content || s.Content || '';
      return kind === 'PHOTO' && content.includes(marker);
    });
    expect(found, 'crew creative list should include the new PHOTO').toBe(true);

    // Timeline UI renders the photo caption (not raw JSON blob only)
    await page.goto('/#/creative');
    await expect(page.getByRole('heading', { name: 'Family Journal' }).first()).toBeVisible({ timeout: T });
    await expect(page.getByText(marker, { exact: false }).first()).toBeVisible({ timeout: 20_000 });

    // PHOTO submission is pending review: xp/coins must be unchanged.
    const meAfter = await fetchMe(page);
    expect(meAfter).not.toBeNull();
    if (meBefore && meAfter) {
      expect(meAfter.xp, 'XP must not change from a pending PHOTO submission').toBe(meBefore.xp);
      expect(meAfter.coins, 'Coins must not change from a pending PHOTO submission').toBe(meBefore.coins);
    }

    // eslint-disable-next-line no-console
      console.log(
        '[PROD SMOKE]',
        JSON.stringify({
          invalid_status: bad.status,
          submit_status: good.status,
          marker,
          listed: found,
          xp_before: meBefore?.xp,
          xp_after: meAfter?.xp,
          coins_before: meBefore?.coins,
          coins_after: meAfter?.coins,
        }),
      );
  });
});
