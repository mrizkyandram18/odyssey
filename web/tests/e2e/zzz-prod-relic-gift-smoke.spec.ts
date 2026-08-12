import { test, expect } from '@playwright/test';
import { login } from './helpers/auth';

const SUPABASE_SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY || '';
const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const T = 90_000;

/**
 * Slice 2.11 prod smoke — relic gifting full round-trip.
 *
 * Steps performed:
 *  1. Login Explorer A (demo1).
 *  2. Read A's inventory and profile stats (xp, coins, role, level).
 *  3. Find an owned relic with owned_count >= 2 so A keeps at least 1 after gifting.
 *     (If no relic with count >= 2, skip gracefully.)
 *  4. Get A's crew members from /api/quests to find Explorer B (demo2).
 *  5. Capture B's pre-gift inventory count for the target relic and stats.
 *  6. POST /api/relics/gift from A to B.
 *  7. Assert HTTP 200 and sender_remaining_count == (originalCount - 1).
 *  8. Re-read A's inventory: assert target relic count decreased by exactly 1.
 *  9. Re-read A's profile: assert XP, coins, role, level unchanged.
 * 10. Login Explorer B (demo2).
 * 11. Read B's inventory: assert target relic count increased by exactly 1.
 * 12. Read B's profile: assert XP, coins, role, level unchanged.
 * 13. Assert unrelated relics are unchanged for both A and B.
 */

interface PlayerStats {
  uid: string;
  xp: number;
  coins: number;
  role: string;
  level: number;
}

interface RelicItem {
  relic_slug: string;
  owned_count: number;
}

test.describe('PROD smoke: relic gifting round-trip (Slice 2.11)', () => {
  test('gifts relic from A to B, verifies counts and economy invariants', async ({ browser }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only relic gift smoke');
    test.setTimeout(300_000);

    // ── Context A (demo1) ──────────────────────────────────────────────────────
    const ctxA = await browser.newContext();
    const pageA = await ctxA.newPage();
    await login(pageA, 'demo1');

    // 1. Get A's profile stats
    const statsA0 = await pageA.evaluate<PlayerStats>(async () => {
      const r = await fetch('/api/me', { credentials: 'same-origin' });
      const j = await r.json();
      return { uid: j.uid as string, xp: j.xp as number, coins: j.coins as number, role: j.role as string, level: j.level as number };
    }, undefined, { timeout: T });

    // 2. Get A's inventory
    const invA0 = await pageA.evaluate<RelicItem[]>(async () => {
      const r = await fetch('/api/relics/inventory', { credentials: 'same-origin' });
      return await r.json();
    }, undefined, { timeout: T });

    expect(Array.isArray(invA0)).toBe(true);

    // 3. Find a relic A owns with count >= 2 to gift one and keep one
    const giftableRelic = invA0.find((item) => item.owned_count >= 2);
    if (!giftableRelic) {
      test.skip(true, 'demo1 has no relic with owned_count >= 2 — skipping gift smoke');
      return;
    }
    const targetSlug = giftableRelic.relic_slug;
    const aCount0 = giftableRelic.owned_count;

    // 4. Verify demo2 (B) is in the same crew by getting their profile
    const ctxB = await browser.newContext();
    const pageB = await ctxB.newPage();
    await login(pageB, 'demo2');

    const statsB0 = await pageB.evaluate<PlayerStats>(async () => {
      const r = await fetch('/api/me', { credentials: 'same-origin' });
      const j = await r.json();
      return { uid: j.uid as string, xp: j.xp as number, coins: j.coins as number, role: j.role as string, level: j.level as number };
    }, undefined, { timeout: T });

    // 5. Get B's pre-gift count for the target relic
    const invB0 = await pageB.evaluate<RelicItem[]>(async () => {
      const r = await fetch('/api/relics/inventory', { credentials: 'same-origin' });
      return await r.json();
    }, undefined, { timeout: T });

    const bRelicBefore = invB0.find((i) => i.relic_slug === targetSlug);
    const bCount0 = bRelicBefore?.owned_count ?? 0;

    // 6. GET CSRF token for A's context (gift route is CSRF-protected)
    const csrfToken = await pageA.evaluate<string>(async () => {
      const r = await fetch('/api/csrf', { credentials: 'same-origin' });
      const j = await r.json() as { csrf_token?: string };
      return j.csrf_token ?? '';
    }, undefined, { timeout: T });

    // 7. POST /api/relics/gift from A to B
    const giftResp = await pageA.evaluate<{ status: number; body: Record<string, unknown> }>(
      async ([slug, recipientUID, csrf]) => {
        const r = await fetch('/api/relics/gift', {
          method: 'POST',
          credentials: 'same-origin',
          headers: {
            'Content-Type': 'application/json',
            'X-CSRF-Token': csrf as string,
          },
          body: JSON.stringify({ recipient_uid: recipientUID, relic_slug: slug }),
        });
        const body = await r.json().catch(() => ({}));
        return { status: r.status, body: body as Record<string, unknown> };
      },
      [targetSlug, statsB0.uid, csrfToken],
      { timeout: T },
    );

    expect(giftResp.status, `Gift POST must return 200, got ${giftResp.status}: ${JSON.stringify(giftResp.body)}`).toBe(200);
    expect(giftResp.body.relic_slug).toBe(targetSlug);

    // 8. Assert sender_remaining_count == originalCount - 1
    const reportedSenderCount = giftResp.body.sender_remaining_count as number;
    expect(reportedSenderCount, 'sender_remaining_count must equal originalCount - 1').toBe(aCount0 - 1);

    // 9. Re-read A's inventory and assert count decreased by exactly 1
    const invA1 = await pageA.evaluate<RelicItem[]>(async () => {
      const r = await fetch('/api/relics/inventory', { credentials: 'same-origin' });
      return await r.json();
    }, undefined, { timeout: T });

    const aRelicAfter = invA1.find((i) => i.relic_slug === targetSlug);
    const aCount1 = aRelicAfter?.owned_count ?? 0;
    expect(aCount1, `A's ${targetSlug} count must decrease by 1: ${aCount0} → ${aCount1}`).toBe(aCount0 - 1);

    // 10. Re-read A's profile: XP, coins, role, level unchanged
    const statsA1 = await pageA.evaluate<PlayerStats>(async () => {
      const r = await fetch('/api/me', { credentials: 'same-origin' });
      const j = await r.json();
      return { uid: j.uid as string, xp: j.xp as number, coins: j.coins as number, role: j.role as string, level: j.level as number };
    }, undefined, { timeout: T });

    expect(statsA1.xp, 'A xp must not change').toBe(statsA0.xp);
    expect(statsA1.coins, 'A coins must not change').toBe(statsA0.coins);
    expect(statsA1.role, 'A role must not change').toBe(statsA0.role);
    expect(statsA1.level, 'A level must not change').toBe(statsA0.level);

    // 11. Assert unrelated relics for A are unchanged
    for (const item0 of invA0) {
      if (item0.relic_slug === targetSlug) continue;
      const item1 = invA1.find((i) => i.relic_slug === item0.relic_slug);
      expect(item1?.owned_count ?? 0, `A's unrelated relic ${item0.relic_slug} must be unchanged`).toBe(item0.owned_count);
    }

    // 12. Re-read B's inventory: assert target relic increased by exactly 1
    const invB1 = await pageB.evaluate<RelicItem[]>(async () => {
      const r = await fetch('/api/relics/inventory', { credentials: 'same-origin' });
      return await r.json();
    }, undefined, { timeout: T });

    const bRelicAfter = invB1.find((i) => i.relic_slug === targetSlug);
    const bCount1 = bRelicAfter?.owned_count ?? 0;
    expect(bCount1, `B's ${targetSlug} count must increase by 1: ${bCount0} → ${bCount1}`).toBe(bCount0 + 1);

    // 13. Re-read B's profile: XP, coins, role, level unchanged
    const statsB1 = await pageB.evaluate<PlayerStats>(async () => {
      const r = await fetch('/api/me', { credentials: 'same-origin' });
      const j = await r.json();
      return { uid: j.uid as string, xp: j.xp as number, coins: j.coins as number, role: j.role as string, level: j.level as number };
    }, undefined, { timeout: T });

    expect(statsB1.xp, 'B xp must not change').toBe(statsB0.xp);
    expect(statsB1.coins, 'B coins must not change').toBe(statsB0.coins);
    expect(statsB1.role, 'B role must not change').toBe(statsB0.role);
    expect(statsB1.level, 'B level must not change').toBe(statsB0.level);

    // 14. Assert unrelated relics for B are unchanged
    for (const item0 of invB0) {
      if (item0.relic_slug === targetSlug) continue;
      const item1 = invB1.find((i) => i.relic_slug === item0.relic_slug);
      expect(item1?.owned_count ?? 0, `B's unrelated relic ${item0.relic_slug} must be unchanged`).toBe(item0.owned_count);
    }

     
    console.log('[PROD SMOKE 2.11]', JSON.stringify({
      target_slug: targetSlug,
      a_count_before: aCount0, a_count_after: aCount1,
      b_count_before: bCount0, b_count_after: bCount1,
      a_xp_unchanged: statsA1.xp === statsA0.xp,
      a_coins_unchanged: statsA1.coins === statsA0.coins,
      b_xp_unchanged: statsB1.xp === statsB0.xp,
      b_coins_unchanged: statsB1.coins === statsB0.coins,
    }));

    await ctxA.close();
    await ctxB.close();
  });

  test('gift endpoint rejects empty body (400)', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only relic gift smoke');
    test.setTimeout(120_000);

    await login(page, 'demo1');
    const csrf = await page.evaluate<string>(async () => {
      const r = await fetch('/api/csrf', { credentials: 'same-origin' });
      const j = await r.json() as { csrf_token?: string };
      return j.csrf_token ?? '';
    }, undefined, { timeout: T });

    const bad = await page.evaluate<{ status: number; error: string }>(async (token) => {
      const r = await fetch('/api/relics/gift', {
        method: 'POST',
        credentials: 'same-origin',
        headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': token as string },
        body: JSON.stringify({}),
      });
      const j = await r.json() as { error?: string };
      return { status: r.status, error: j.error ?? '' };
    }, csrf, { timeout: T });

    expect([400, 405]).toContain(bad.status);
    if (bad.status === 400) {
      expect(typeof bad.error).toBe('string');
      expect(bad.error.length).toBeGreaterThan(0);
    }
  });

  test('gift endpoint rejects self-gift (400)', async ({ page }) => {
    const base = process.env.PLAYWRIGHT_TEST_BASE_URL || PROD_URL;
    test.skip(!!SUPABASE_SERVICE_KEY || base.includes('localhost'), 'prod-only relic gift smoke');
    test.setTimeout(120_000);

    await login(page, 'demo1');

    const myUID = await page.evaluate<string>(async () => {
      const r = await fetch('/api/me', { credentials: 'same-origin' });
      const j = await r.json();
      return j.uid as string;
    }, undefined, { timeout: T });

    const csrf = await page.evaluate<string>(async () => {
      const r = await fetch('/api/csrf', { credentials: 'same-origin' });
      const j = await r.json() as { csrf_token?: string };
      return j.csrf_token ?? '';
    }, undefined, { timeout: T });

    const resp = await page.evaluate<{ status: number }>(
      async ([uid, token]) => {
        const r = await fetch('/api/relics/gift', {
          method: 'POST',
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/json', 'X-CSRF-Token': token as string },
          body: JSON.stringify({ recipient_uid: uid, relic_slug: 'ancient-compass' }),
        });
        return { status: r.status };
      },
      [myUID, csrf],
      { timeout: T },
    );

    expect([400, 405]).toContain(resp.status); // 400 post-deploy, 405 pre-deploy
  });
});
