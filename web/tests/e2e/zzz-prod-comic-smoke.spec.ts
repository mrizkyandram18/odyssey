import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;

/**
 * Slice 2.7 prod smoke — COMIC multi-panel submission.
 * Uses quest 102 (DONE after prior completion smoke) for CREATE_MEMORY path.
 * PHOTO/VIDEO intentionally not exercised (no storage pipeline).
 */
test.describe('PROD smoke: comic submission (Slice 2.7)', () => {
  test('rejects invalid comic; accepts multi-panel comic; timeline renders panels', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!process.env.SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only comic smoke');
    test.setTimeout(240_000);

    await login(page, 'demo1');

    const postComic = async (content: string) =>
      await page.evaluate(
        async ({ content: bodyContent }) => {
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

          // Quest 102 DONE → post-quest memory against last challenge (39).
          const r = await fetch('/api/creative', {
            method: 'POST',
            credentials: 'include',
            headers,
            body: JSON.stringify({
              quest_id: 102,
              challenge_id: 39,
              kind: 'COMIC',
              content: bodyContent,
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
        { content },
        { timeout: T },
      );

    // Invalid: single panel → 400
    const bad = await postComic(JSON.stringify({ panels: [{ caption: 'only one' }] }));
    expect(bad.status).toBe(400);
    expect(String(bad.body.error || bad.text)).toMatch(/panel/i);

    // Valid multi-panel caption comic
    const marker = `Slice27 comic ${Date.now()}`;
    const goodPayload = JSON.stringify({
      v: 1,
      panels: [{ caption: marker }, { caption: 'Panel two of the strip.' }],
    });
    const good = await postComic(goodPayload);
    expect(good.status, `submit comic failed: ${good.text}`).toBe(201);
    expect(good.body.kind === 'COMIC' || (good.body as { Kind?: string }).Kind === 'COMIC').toBeTruthy();

    // List by crew includes the new comic
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
      return kind === 'COMIC' && content.includes(marker);
    });
    expect(found, 'crew creative list should include the new COMIC').toBe(true);

    // Timeline UI renders the comic caption (not raw JSON blob only)
    await page.goto('/#/creative');
    await expect(page.getByRole('heading', { name: 'Family Journal' }).first()).toBeVisible({ timeout: T });
    await expect(page.getByText(marker, { exact: false }).first()).toBeVisible({ timeout: 20_000 });

     
    console.log(
      '[PROD SMOKE]',
      JSON.stringify({
        invalid_status: bad.status,
        submit_status: good.status,
        marker,
        listed: found,
      }),
    );
  });
});
