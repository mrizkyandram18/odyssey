# Stage C — Root Cause Analysis (RCA) Report

This document presents individual Root Cause Analyses for all 6 **FAIL** items identified during the Stage B Audit. Every RCA is supported by verified **E3 evidence** (file references, line numbers, and exact code constructs) and defines a safe fix strategy.

---

## RCA-01: `HomePage` Data Fetching Missing `useEffect`

```text
Issue:
HomePage does not trigger data fetching on initial mount, leaving the component permanently in a loading state.

Evidence: (E3)
- File: web/src/features/home/HomePage.tsx
- Lines 12 & 54: useState `loading` initializes to `true`. Line 54 returns `<p>Loading...</p>` when `loading && !home`.
- Line 41: `const loadHome = async () => { ... }` is defined.
- Inspection of lines 1-60 reveals NO `useEffect` hook to call `loadHome()` on component mount.

Root Cause:
`HomePage.tsx` was created with `loadHome()` defined as a handler function for manual refresh (e.g. after consuming daily turns or opening chests), but the initial mount lifecycle trigger (`useEffect(() => { loadHome() }, [])`) was omitted.

Impact:
Navigating to the home screen results in an infinite "Loading..." screen unless an external action triggers `loadHome()`.

Risk:
P1 (Critical Usability Blocker for Golden Path).

Fix Strategy:
Add a `useEffect` hook in `HomePage.tsx` that executes `loadHome()` once on component mount.

Regression Risk:
Low. Ensures standard React component lifecycle behavior.

Verification Plan:
1. Run `npm run build` to ensure TypeScript compilation passes.
2. Mount `HomePage` in browser/E2E test and verify `/api/home` is requested automatically on load.
```

---

## RCA-02: Service Worker Lacks Lifecycle & Update Handlers

```text
Issue:
The PWA Service Worker does not handle cache updates or notify the user when a new app version is available, causing clients to serve stale cached assets indefinitely.

Evidence: (E3)
- File: web/public/sw.js (lines 3 & 17-22)
  Uses hardcoded `odyssey-shell-v1` and `caches.match` without an `activate` event listener to clean up old cache versions or `skipWaiting()`.
- File: web/src/app/service-worker-registration.ts (lines 1-8)
  Calls `navigator.serviceWorker.register('/sw.js')` with no event listeners for `updatefound` or `controllerchange`.

Root Cause:
The PWA implementation was added as a baseline shell registration without incorporating Service Worker lifecycle management or cache invalidation strategies.

Impact:
When new code is deployed to production, existing users will remain stuck on older cached JavaScript/CSS assets until they manually clear site data.

Risk:
P1 (Critical Release Blocker).

Fix Strategy:
1. Update `sw.js` to add an `activate` event listener that purges caches whose names do not match `CACHE_NAME`, and call `self.skipWaiting()` during `install`.
2. Update `service-worker-registration.ts` to listen for `registration.onupdatefound` and notify the UI if a new worker is waiting.

Regression Risk:
Low. Standard PWA lifecycle pattern.

Verification Plan:
1. Build frontend (`npm run build`).
2. Verify SW registration in browser DevTools Application tab.
3. Update `CACHE_NAME` in `sw.js` and verify old caches are automatically deleted upon activation.
```

---

## RCA-03: `time.Sleep` Usage in Test Suite

```text
Issue:
The Go unit test suite contains wall-clock `time.Sleep` calls in time-sensitive logic (e.g. rate limiter TTL tests).

Evidence: (E3)
- File: pkg/shared/security_test.go (lines 135 & 144)
  `TestRateLimiter_ExpiredEntries` calls `time.Sleep(100 * time.Millisecond)` to wait for a 50ms TTL window to elapse.
- Additional 7 `time.Sleep` instances found in `pkg/content/service_test.go:177`, `pkg/observability/health_test.go:319`, `pkg/observability/profiler_test.go:142`, etc.

Root Cause:
Tests were written using wall-clock delays (`time.Sleep`) instead of a controllable clock abstraction or mock time provider.

Impact:
On heavily loaded CI runners or slow virtual machines, `time.Sleep(100ms)` may take unpredictable durations or fail due to thread scheduling context switches, causing flaky test runs.

Risk:
P1 (Test Reliability & CI Stability).

Fix Strategy:
Replace wall-clock `time.Sleep` in `pkg/shared/security_test.go` with a small, deterministic time-step or clock abstraction, or adjust buffer margins where clock mocking is impractical.

Regression Risk:
Low. Pure test-file modification.

Verification Plan:
Run `go test -v -count=50 ./pkg/shared/...` to verify 100% deterministic test execution under iteration.
```

---

## RCA-04: Basic Loading & Error UX in Profile and Journal Pages

```text
Issue:
`JournalPage` and `ProfilePage` use plain text strings for loading states and lack retry actions when data fetching fails.

Evidence: (E3)
- File: web/src/features/journal/JournalPage.tsx (line 71)
  Renders `{loading ? <p className="text-sm text-muted-foreground">Loading journal entries...</p> : ...}`
- File: web/src/features/profile/ProfilePage.tsx (lines 16-32)
  Renders `<p>Loading profile...</p>` and `<p>No profile data available.</p>` without retry controls or unauthenticated redirect logic.

Root Cause:
Initial implementation focused on wiring backend API hooks and used basic fallback UI components without skeleton placeholders or error-recovery handlers.

Impact:
Degraded user experience on high-latency networks; users cannot retry failed network requests without refreshing the entire page.

Risk:
P2 (UX Quality Gate).

Fix Strategy:
1. Replace plain text loading screens with structured skeleton components (`Skeleton` loader).
2. Add a "Retry" button on error states that re-executes the fetch function.
3. If `profile` is null and loading is false in `ProfilePage`, trigger session check or show a clear login redirect action.

Regression Risk:
Low. UI presentation enhancement.

Verification Plan:
1. Verify loading state displays skeleton layout in browser.
2. Simulate network failure (DevTools Offline mode) and verify Retry button re-fetches data upon click.
```

---

## RCA-05: Standard `log.Printf` Usage Bypassing Observability Logger

```text
Issue:
Chest catalog initialization and lookup in `pkg/game/chest/catalog.go` uses standard library `log.Printf` instead of the structured `observability.Logger`.

Evidence: (E3)
- File: pkg/game/chest/catalog.go
  - Line 68: `log.Printf("WARN: loading drop table %s: %v", def.Slug, err)`
  - Line 75: `log.Printf("WARN: no fallback exists for %s", slug)`
  - Line 99: `log.Printf("WARN: loading drop table %s: %v", def.Slug, err)`
  - Line 104: `log.Printf("WARN: no fallback exists for %s", def.Slug)`

Root Cause:
`ContentChestCatalog` was constructed prior to standardizing on `observability.Logger` and was not refactored to receive a logger dependency.

Impact:
Chest drop table warnings bypass JSON structured logging formats and request ID correlation headers in server logs.

Risk:
P2 (Observability Audit Gate).

Fix Strategy:
Inject `observability.Logger` into `ContentChestCatalog` (or `ChestService`) and replace `log.Printf` with `logger.Warn(...)`.

Regression Risk:
Low. Refactors logging call site only.

Verification Plan:
1. Run `go test ./pkg/game/chest/...`.
2. Inspect log outputs during test execution to verify structured JSON log output for catalog warnings.
```

---

## RCA-06: Redundant Network Fetch on Challenge Completion

```text
Issue:
`completeChallenge` in `useQuest` performs an unnecessary `fetchQuest()` HTTP request immediately after receiving the challenge completion response.

Evidence: (E3)
- File: web/src/shared/hooks/useQuest.ts (lines 43-54)
  Line 46: `const result = await questsApi.completeChallenge(questId, challengeId)`
  Line 47: `await fetchQuest()` (makes a GET `/api/quests/${questId}` request despite `result` already containing the updated `quest` object).

Root Cause:
`useQuest` was written defensively to ensure local state matched the server by re-fetching the entire quest object upon completion.

Impact:
Generates two back-to-back network requests for every challenge completion, increasing server load and UI latency.

Risk:
P2 (Performance Audit Gate).

Fix Strategy:
Update `useQuest.ts` so that if `result.quest` is present in the response from `completeChallenge`, update `data` directly with `result.quest` instead of triggering `await fetchQuest()`.

Regression Risk:
Low. Preserves state sync while eliminating the extra GET request.

Verification Plan:
1. Run `npm run test` in `web/`.
2. Perform challenge completion in browser Network tab and verify only 1 POST request occurs (`/api/quests/:id/challenges/:cid/complete`) without a subsequent GET `/api/quests/:id`.
```

---

## Summary of Proposed Fix Batches

Per our **Change Budget** (`One RCA → One Fix Batch → One Verification`), the fixes will be executed in isolated batches ordered by severity:

| Batch | Target RCA | Scope | Severity |
| :---: | :--- | :--- | :--- |
| **Batch 1** | **RCA-01** | Add `useEffect` in `HomePage.tsx` | **P1** |
| **Batch 2** | **RCA-02** | Add SW update listeners & cache cleanup | **P1** |
| **Batch 3** | **RCA-03** | Fix `time.Sleep` in `pkg/shared/security_test.go` | **P1** |
| **Batch 4** | **RCA-06** | Optimize state update in `useQuest.ts` | **P2** |
| **Batch 5** | **RCA-04** | Add skeletons & retry buttons in `JournalPage` / `ProfilePage` | **P2** |
| **Batch 6** | **RCA-05** | Inject structured logger in `catalog.go` | **P2** |
