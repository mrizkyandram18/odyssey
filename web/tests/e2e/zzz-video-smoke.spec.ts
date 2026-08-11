import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';
import fs from 'fs';
import path from 'path';
import { fileURLToPath } from 'url';


const T = 90_000;

// Read the video fixture from disk and base64-encode it
const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const fixturePath = path.resolve(__dirname, '../fixtures/tiny-video.mp4');
const videoBuffer = fs.readFileSync(fixturePath);
const base64Video = videoBuffer.toString('base64');
const videoDataURL = `data:video/mp4;base64,${base64Video}`;
console.log(`[VIDEO SMOKE] fixture size: ${videoBuffer.length} bytes, base64: ${base64Video.length} chars`);

interface MeSnapshot {
  xp: number;
  coins: number;
}

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
 * Slice 2.10 VIDEO smoke test — mirrors the PHOTO smoke.
 * Uses quest 102 (DONE) with the last challenge (39).
 */
test.describe('VIDEO smoke (Slice 2.10)', () => {
  test('accepts valid video mp4; rejects non-video; timeline renders video+caption', async ({ page }) => {

    test.skip(!!process.env.SUPABASE_SERVICE_KEY, 'skip when Supabase service key set (DB reset mode)');
    test.setTimeout(240_000);

    await login(page, 'demo1');

    const meBefore = await fetchMe(page);
    if (!meBefore) {
      test.fail(true, '/api/me returned no player snapshot');
    }

    const postVideo = async (video: string, caption: string) =>
      await page.evaluate(
        async ({ video: videoData, caption: cap }) => {
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
              kind: 'VIDEO',
              content: JSON.stringify({ v: 1, video: videoData, caption: cap }),
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
        { video, caption },
        { timeout: T },
      );

    // Invalid: non-video data URI -> 400
    const bad = await postVideo('data:text/plain;base64,aGk=', 'not a video');
    expect(bad.status).toBe(400);
    expect(String(bad.body.error || bad.text)).toMatch(/video/i);

    // Invalid: non-base64 video URI -> 400
    const badB64 = await postVideo('data:video/mp4;base64,!!!notbase64!!!', 'bad b64');
    expect(badB64.status).toBe(400);

    // Valid: tiny mp4 video with a unique caption
    const marker = `Slice210-video ${Date.now()}`;
    const good = await postVideo(videoDataURL, marker);
    expect(good.status, `submit video failed: ${good.text}`).toBe(201);
    expect(good.body.kind === 'VIDEO' || (good.body as { Kind?: string }).Kind === 'VIDEO').toBeTruthy();

    // List by crew includes the new video
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
      return kind === 'VIDEO' && content.includes(marker);
    });
    expect(found, 'crew creative list should include the new VIDEO').toBe(true);

    // Timeline UI renders the video caption (and a <video> element)
    await page.goto('/#/creative');
    await expect(page.getByRole('heading', { name: 'Family Journal' }).first()).toBeVisible({ timeout: T });
    await expect(page.getByText(marker, { exact: false }).first()).toBeVisible({ timeout: 20_000 });
    // video element should exist on the timeline
    const hasVideoEl = await page.evaluate(() => {
      return document.querySelector('video') !== null;
    });
    expect(hasVideoEl, 'timeline should render a <video> element for the VIDEO submission').toBe(true);

    // VIDEO submission is pending review: xp/coins must be unchanged.
    const meAfter = await fetchMe(page);
    expect(meAfter).not.toBeNull();
    if (meBefore && meAfter) {
      expect(meAfter.xp, 'XP must not change from a pending VIDEO submission').toBe(meBefore.xp);
      expect(meAfter.coins, 'Coins must not change from a pending VIDEO submission').toBe(meBefore.coins);
    }


    console.log(
      '[VIDEO SMOKE]',
      JSON.stringify({
        invalid_nonvideo_status: bad.status,
        invalid_badbase64_status: badB64.status,
        submit_status: good.status,
        marker,
        listed: found,
        timeline_has_video_el: hasVideoEl,
        xp_before: meBefore?.xp,
        xp_after: meAfter?.xp,
        coins_before: meBefore?.coins,
        coins_after: meAfter?.coins,
      }),
    );
  });
});
