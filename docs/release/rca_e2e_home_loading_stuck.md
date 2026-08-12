# RCA: Production E2E Failure — Golden Path Test 1 (Home Loading stuck)

**Date:** 2026-08-08  
**Verifier:** Live re-verification (Evidence First)  
**Status:** **STOP** — full suite not green; release **NOT READY FOR QA** until re-run after approved fix  
**Scope:** Test / production runtime behavior only — no application code changed, no commit/push

---

## 1. Failure Summary

| Field | Value |
| --- | --- |
| Suite | Playwright E2E → production |
| Target | `PLAYWRIGHT_TEST_BASE_URL=https://odyssey-beta-nine.vercel.app` |
| Result | **15 passed / 1 failed / 0 skipped / 0 retries** |
| Duration | ~3.2 minutes |
| Failed test | `golden-path.spec.ts` → **Test 1: Login -> Home -> Mission list appears** |
| Assertion | `helpers/home.ts` → `expect(locator('text="Loading..."')).toBeHidden({ timeout: 15000 })` |
| Observed UI | Full-screen / page body stuck on `Loading...` for entire assertion window |

---

## 2. What Passed in the Same Run (Evidence)

These are from the **same** production E2E execution (not prior reports):

| Area | Result |
| --- | --- |
| Auth Domain login/logout | PASS |
| Golden Path Test 2 (quest flow + journal) | PASS |
| Golden Path Test 3 (logout/re-login persistence) | PASS |
| Home Domain dashboard + quest list | PASS |
| Journal Domain | PASS |
| Multi-user demo1 / demo2 / demo3 | PASS |
| Session persist after reload | PASS |
| State after logout/login | PASS |
| Mission Domain complete challenge | PASS |
| Deep-link `/#/missions/103` | PASS |
| Browser back/forward | PASS |
| Unauthenticated redirect to `/#/login` | PASS |

**Implication:** Functional product paths mostly work; the single failure is concentrated on **home load readiness timing** for the first heavy `/api/home` after earlier logins.

---

## 3. API Smoke (Independent of Playwright) — Actual Execution

Target: `https://odyssey-beta-nine.vercel.app`

| Check | Result | Evidence |
| --- | --- | --- |
| Frontend shell `GET /` | **PASS** 200 | HTML, `#root`, assets present |
| `GET /api/status` | **PASS** 200 | `app=odyssey`, `schema_version=11`, missions=12 |
| `POST /api/login` demo1 | **PASS** 200 | `uid=demo-uid-1`, `role=SEEKER` |
| `POST /api/login` demo2 | **PASS** 200 | `uid=demo-uid-2`, `role=GUIDE` |
| `POST /api/login` demo3 | **PASS** 200 | `uid=demo-uid-3`, `role=BUILDER` |
| `POST /api/login` invalid | **PASS** 401 | wrong credential rejected |
| `GET /api/me` no auth | **PASS** 401 | |
| `GET /api/me` demo1 session | **PASS** 200 | explorer **Leo**, SEEKER, level 23 |
| `GET /api/me` demo2 session | **PASS** 200 | explorer **Maya**, GUIDE |
| `GET /api/me` demo3 session | **PASS** 200 | explorer **Sam**, BUILDER |
| `GET /api/missions` | **PASS** 200 | includes id 103 `Riddle of the Stones` |
| `GET /api/missions/103` | **PASS** 200 | title matches; exercises present |
| `GET /api/home` (post-login) | **PASS** 200 | **~8084–8732 ms** per call (3 rapid probes all ~8s) |

Session tokens were obtained and used for authenticated calls; tokens are **not** copied into this report.

### Mission 103 observation (non-blocking for routing)

- List view: status `ACTIVE`, `completed_count=2`
- Detail view: status `DONE`, both exercises `DONE`

No production data was mutated for this verification.

---

## 4. Failure Evidence (Playwright Trace)

From `test-results/golden-path-.../0-trace.network` for the failing test:

| Request | Started (UTC) | Duration | Status |
| --- | --- | --- | --- |
| `POST /api/login` | 2026-08-08T13:28:35.123Z | ~1413 ms | **200** |
| `GET /api/me` | 2026-08-08T13:28:36.739Z | ~1007 ms | **200** |
| `GET /api/home` | 2026-08-08T13:28:36.749Z | **time = -1** | **status = -1** (no completed response before test ended) |

Screenshot: dark page with only `Loading...` visible (matches `HomePage` early return when `loading && !home`).

### Code path

```ts
// web/src/features/home/HomePage.tsx
const loadHome = async () => {
  setLoading(true)
  // ...
  const data = await apiClient.get<HomeResponse>('/api/home')
  // setLoading(false) only after fetch resolves
}
if (loading && !home) {
  return <p className="p-4 text-sm text-muted-foreground">Loading...</p>
}
```

There is **no client-side fetch timeout**. If `/api/home` is still in flight, UI stays on `Loading...`.

---

## 5. Root Cause Analysis

### Primary cause (most probable)

**`GET /api/home` is a heavy serverless endpoint with multi-second latency on production.**  
Live probes after the suite consistently returned **HTTP 200 in ~8.0–8.7 seconds**. Under cold start / concurrent load, completion can exceed the E2E helper’s **15s** wait for `Loading...` to disappear.

In the failing run, Playwright recorded `/api/home` as **incomplete** (`time=-1`, `status=-1`) for the entire assertion window — i.e. no response body by the time the test failed (~15s after home mount).

### Contributing factors

1. **Test timeout budget is tight relative to observed production latency**
   - Helper: `toBeHidden({ timeout: 15000 })` in `helpers/home.ts`
   - Observed warm latency: ~8s
   - Cold / first-hit latency can reasonably exceed 15s on Vercel serverless + Supabase aggregation

2. **Order effect**
   - Auth tests only exercise login/logout (no `/api/home`)
   - Golden Path Test 1 is early in suite and first consumer of `/api/home` in that worker sequence
   - Subsequent tests that also call home (**Home Domain**, Golden Path 2/3, multi-user, etc.) **passed** once the function was warmer

3. **Not a HashRouter / auth / seed regression**
   - Login succeeded (200 + redirect to `/#/`)
   - `/api/me` succeeded
   - Same run: multi-user, deep-link, back/forward, unauthenticated redirect all PASS
   - Seed accounts and Local Auth confirmed outside Playwright

4. **Not a destructive DB issue**
   - `resetQuestState()` is a no-op for non-localhost baseURL / without service key / with `SKIP_DB_RESET=true`
   - No truncate/reseed/delete performed

### Ruled out (for this failure)

| Hypothesis | Why ruled out |
| --- | --- |
| Wrong baseURL / localhost | Env `PLAYWRIGHT_TEST_BASE_URL=https://odyssey-beta-nine.vercel.app`; network URLs all production |
| webServer local Vite | `webServer` only when baseURL contains `localhost` |
| Invalid credentials / Local Auth broken | Login 200; multi-user PASS |
| HashRouter path bug | URL assertions and regression suite PASS |
| Permanent `/api/home` outage | Live probes 200 after suite; later E2E home tests PASS |
| Rate limit 429 on login | Login for this test returned 200 in ~1.4s |

---

## 6. Severity & Release Impact

| Item | Assessment |
| --- | --- |
| Product broken for all users? | **Unproven.** Home eventually loads under warm conditions; cold latency is high. |
| Exit criteria “0 failed”? | **FAIL** — 1 failed test in full production E2E |
| Decision now | **NOT READY FOR QA** |

---

## 7. Proposed Fixes (Awaiting Approval)

**Do not bypass the assertion.** Prefer resilience + honest timeouts.

### Option A — Test infrastructure only (Recommended first)

Adjust `web/tests/e2e/helpers/home.ts` for production cold starts:

1. Increase `Loading...` wait to a **reasonable production timeout** (e.g. **45s**), not infinite.
2. Optionally wait for network: `page.waitForResponse(r => r.url().includes('/api/home') && r.ok(), { timeout: 45000 })` before asserting UI.
3. Keep retries at **0** (exit criteria).
4. Keep production baseURL; no localhost.

Rationale: matches stated policy (“serverless cold start… use reasonable timeout, not bypass”) and does not change production architecture.

### Option B — Application performance (separate track; needs product approval)

Investigate why `/api/home` consistently takes **~8s** warm:

- Home service aggregation cost
- Supabase round-trips
- Serverless cold start
- Optional: client fetch timeout + error UI so `Loading...` cannot hang forever

This is **not** required to flip a flaky 15s helper, but is a real production quality concern.

### Option C — Do not do

- Disable rate limiter
- Reset/reseed production to force PASS
- Change architecture only for tests
- Enable Playwright retries just to greenwash
- Commit/push without approval

---

## 8. Recommended Next Step

1. **Approve Option A** (test helper timeout / waitForResponse).
2. Re-run full production E2E with:

   ```text
   PLAYWRIGHT_TEST_BASE_URL=https://odyssey-beta-nine.vercel.app
   SKIP_DB_RESET=true
   npm run test:e2e -- --reporter=list
   ```

   from `web/` (use local `@playwright/test`, not bare `npx playwright` which can pull a mismatched CLI).

3. Require **16 passed / 0 failed / 0 skipped / 0 retries**.
4. Only then write final **READY FOR QA** evidence report from that run.

---

## 9. Safety Checklist for This Session

| Check | Status |
| --- | --- |
| No truncate/delete/reseed production | Held |
| No rate-limiter disable | Held |
| No app code change | Held |
| No commit/push | Held |
| Secrets not printed in reports | Held (session tokens redacted) |
| Target was production, not localhost | Held |

---

## 10. Decision

```text
STOP → RCA complete → proposed fix Option A (test infrastructure timeout for /api/home)
→ WAIT FOR APPROVAL before applying fix or re-running full suite as release gate
```

**Current release label:** **NOT READY FOR QA**
