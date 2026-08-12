# Production QA Evidence Report

**Date:** 2026-08-08  
**Branch:** `feat/phase-1-development`  
**Verifier:** Evidence First — complete seed audit + live production re-verification  
**Production URL:** `https://odyssey-beta-nine.vercel.app`

---

## 0. Gate Summary (This Session)

| Gate | Result |
| --- | --- |
| Complete seed audit (repo) | **PASS** |
| Production DB seed/content (read-only) | **PASS — COMPLETE** |
| Production API smoke | **PASS** |
| Local Auth (demo1/2/3 + invalid) | **PASS** |
| Full Playwright E2E (production) | **16 passed / 0 failed / 0 skipped / 0 retries** |
| Exit code | **0** |
| Free-tier / no paid infra introduced | **PASS** |
| Destructive production ops | **None** |
| Local unit / lint | **PASS** (Go after excluding untracked junk mains; web tsc/eslint/vitest) |
| Remote CI | See Git section (post-push) |

```text
Seed: COMPLETE — no reseed performed
E2E: PASS
```

---

## 1. Production Target

| Layer | Value |
| --- | --- |
| Frontend | `https://odyssey-beta-nine.vercel.app` |
| API | `https://odyssey-beta-nine.vercel.app/api` |
| Database | Supabase production (`odyssey_*` only) |
| Auth runtime | **Local Auth** (`LocalAuthProvider` + `odyssey_local_users`) — **not** Supabase Auth |
| Router | React `HashRouter` (`/#/...`) |
| Schema | `schema_version = "11"` (DB + live `/api/status`) |

Hash routes exercised by E2E: `/#/login`, `/#/`, `/#/journal`, `/#/missions/103`, `/#/profile`.

---

## 2. Phase 0 — Repository & Constraint Audit

| Area | Source | Finding | Evidence | Risk |
| --- | --- | --- | --- | --- |
| Project constraints | `CLAUDE.md` | Only `odyssey_*` tables; Gatekeeper/Family Reward immutable; no real money | Read CLAUDE.md | Note: CLAUDE.md still states Gatekeeper as sole auth path; **prototype runtime uses Local Auth** per deployment docs + code (`pkg/server/server.go` → `NewLocalAuthProvider`). User requires Local Auth preserved. No Gatekeeper/FR changes made. |
| Branch | git | `feat/phase-1-development` @ `9596595` | `git branch -vv` | Uncommitted release work present |
| Migrations canonical | `supabase/migrations` | Canonical migration tree for Supabase | 18 SQL files 001–015, 017–019 (no 016) | Low |
| Migration sync | `scripts/migrations` vs `supabase/migrations` | **Byte-identical** for every shared file | PowerShell `Compare-Object` all files SAME | Low — intentionally synchronized |
| Seed completeness assumption | Prior reports | `019` alone is **not** complete seed | Seeds span 010, 017, 018, 019 | Low after full audit |
| E2E | `web/tests/e2e` | 16 tests across 8 specs | Inventory + run | Low |
| CI | `.github/workflows/ci.yml` | lint → unit → race → integration(Playwright) → build → docker | Workflow file | Integration job uses dummy env / incomplete backend start — known CI gap for real E2E |
| Free tier | `vercel.json`, go.mod, web deps | Vercel + Supabase only; no paid Redis/AI/queues | Audit | Low |

**STOP check:** No requested change conflicts with preserving Local Auth, `odyssey_*` scope, or non-destructive production rules.

---

## 3. Complete Seed Inventory (Repository)

### Schema / structure migrations (no content seed)

| Migration | Role |
| --- | --- |
| 001 | Core tables: families, profiles, missions, exercises, journey_progress, creative_items, daily_missions, achievements, collections, gifts |
| 002 | Indexes |
| 003 | creative_submissions, system_config, schema_version |
| 004 | player_collections + chest/relic expansion columns |
| 005 | chest/drop/relic definition tables |
| 006 | journey/course/quest/prompt/achievement/season/concepts definition tables |
| 007 | course_progress, concept_unlocks + progression columns |
| 008 | balance_configs, audit_logs + live ops |
| 009 | request correlation + schema_version bump |
| 011 | quest prerequisites |
| 012 | reliability/idempotency |
| 013 | reactions, daily_activity |
| 014 | cooperative flow columns |
| 015 | reward_ledgers |

### Content seed migrations

| Migration | Tables seeded | Idempotent? |
| --- | --- | --- |
| **010** | journey/course/quest/prompt/achievement/concepts/chest/drop/relic/season definitions + balance_configs | Yes (`ON CONFLICT DO NOTHING` / key update for balance) |
| **017** | families, profiles, journey_progress, missions, exercises, creative_items, daily_missions, collections, gifts; schema_version→11 | Yes (`ON CONFLICT DO NOTHING`) |
| **018** | reactions (5 seed IDs) | Yes (`ON CONFLICT (id) DO NOTHING`) |
| **019** | `odyssey_local_users` DDL + 3 demo users | Yes (`CREATE IF NOT EXISTS`, `ON CONFLICT (username) DO NOTHING`) |

**Conclusion:** Complete prototype seed is **010 + 017 + 018 + 019**, not users alone.

---

## 4. Production Database Verification (Read-Only)

Method: Supabase REST with service role, `Prefer: count=exact`, selective `select` of non-secret columns. **No writes.** Secrets not printed.

### Schema version

```text
odyssey_schema_version: key=schema_version value=11
```

### Seed matrix (definitions + demo content)

| Dataset | Migration | Expected | Production | Notes | Status |
| --- | --- | --- | --- | --- | --- |
| `odyssey_schema_version` | 017 | 11 | 11 | Matches `/api/status` | **PASS** |
| `odyssey_local_users` | 019 | 3 | 3 | demo1→demo-uid-1, demo2→demo-uid-2, demo3→demo-uid-3 (hashes not logged) | **PASS** |
| `odyssey_families` | 017 | 1 | 1 | demo-crew-1 / The Starseekers | **PASS** |
| `odyssey_user_profiles` | 017 | 3 | 3 | Leo/Maya/Sam present; **Leo XP/level advanced by QA** (level 23 / xp 2210 vs seed 2/150) | **PASS** (present; evolved) |
| `odyssey_missions` | 017 | 3 | 3 | 101 Morning Light DONE; 102 Gather Herbs ACTIVE; 103 Riddle of the Stones present | **PASS** |
| `odyssey_exercises` | 017 | 6 | 6 | All 6 challenge rows for 101–103 exist; **103 exercises DONE via QA** (seed PENDING) | **PASS** (present; evolved) |
| `odyssey_journey_progress` | 017 | ≥1 | 2 | Seed: whispering-woods ACTIVE 50; **now COMPLETE 100 + clockwork-city ACTIVE 0** (gameplay) | **PASS** (superset) |
| `odyssey_creative_items` | 017 | 5 | 5 | Journal/gallery stories match seed samples | **PASS** |
| `odyssey_daily_missions` | 017 | 6 | 6 | Present | **PASS** |
| `odyssey_collections` (instance) | 017 | 2 | 2 | acorn-shard, whispering-leaf | **PASS** |
| `odyssey_gifts` (instance) | 017 | 3 | 4 | Seed 3 + **1 extra from gameplay** | **PASS** (superset) |
| `odyssey_reactions` | 018 | 5 | 5 | Seed UUIDs 000…001–005 | **PASS** |
| `odyssey_journey_definitions` | 010 | 3 | 3 | whispering-woods, clockwork-city, starlit-library | **PASS** |
| `odyssey_course_definitions` | 010 | 4 | 4 | the-awakening, the-deep-woods, gears-and-gold, first-stars | **PASS** |
| `odyssey_quest_definitions` | 010 | 12 | 12 | includes `riddle-of-the-stones` | **PASS** |
| `odyssey_creative_prompt_definitions` | 010 | 6 | 6 | | **PASS** |
| `odyssey_achievement_definitions` | 010 | 10 | 10 | codes match seed | **PASS** |
| `odyssey_concept_definitions` | 010 | 8 | 8 | | **PASS** |
| `odyssey_chest_definitions` | 010 | 5 | 5 | wooden…mystic | **PASS** |
| `odyssey_drop_tables` | 010 | 22 | 22 | | **PASS** |
| `odyssey_relic_definitions` | 010 | 15 | 15 | | **PASS** |
| `odyssey_season_definitions` | 010 | 1 | 1 | season-spring-2026 | **PASS** |
| `odyssey_balance_configs` | 010 | 9 | 9 | all 9 keys present | **PASS** |
| `odyssey_achievements` (instance) | — | 0 seeded | 2 | Gameplay awards; not a seed gap | N/A (runtime) |
| `odyssey_course_progress` | — | 0 seeded | 1 | Gameplay | N/A (runtime) |
| `odyssey_player_collections` | — | 0 | 0 | Schema only | OK |
| `odyssey_system_config` | — | 0 | 0 | Schema only | OK |
| `odyssey_creative_submissions` | — | 0 | 0 | Schema only | OK |
| `odyssey_concept_unlocks` | — | 0 | 0 | Schema only | OK |
| `odyssey_audit_logs` | — | 0 | 0 | Runtime | OK |
| `odyssey_daily_activity` | — | 0 | 0 | Runtime | OK |
| `odyssey_reward_ledgers` | — | 0 | 0 | Runtime | OK |

### Critical content spot-check

```text
Mission 103:
  id=103
  title=Riddle of the Stones
  template_slug=riddle-of-the-stones
  status (list)=ACTIVE  status (detail)=DONE   # QA progress / path drift; content present
  exercises: stone-shape DONE, solve-riddle DONE

Local users:
  demo1 → demo-uid-1
  demo2 → demo-uid-2
  demo3 → demo-uid-3

Profiles:
  demo-uid-1 Leo SEEKER
  demo-uid-2 Maya GUIDE
  demo-uid-3 Sam BUILDER
```

### Seed gap analysis

| Question | Answer |
| --- | --- |
| Is production only users-seeded? | **No** — full definitions + demo content present |
| Missing required seed? | **None found** |
| Reseed required? | **No — DO NOT reseed** |
| Risk if 010/017 re-run? | Idempotent inserts; balance_configs would UPDATE values; would **not** reverse QA progress (ON CONFLICT DO NOTHING on demo rows) |
| Action taken | **None** (non-destructive) |

---

## 5. Free-Tier / Cost Constraint

| Check | Result |
| --- | --- |
| Paid DB / Redis / queues / AI / monitoring added? | **No** |
| Hosting | Existing Vercel free-tier deployment |
| Database | Existing Supabase project |
| Auth | Local Auth in-app (no Supabase Auth product change) |
| New SaaS? | None |

---

## 6. Production API Evidence (Live — This Session)

Target: `https://odyssey-beta-nine.vercel.app`

| Endpoint | Method | Result | Evidence |
| --- | --- | --- | --- |
| `/` | GET | **PASS** 200 | HTML shell, `#root` |
| `/api/status` | GET | **PASS** 200 | app=odyssey, schema_version=11, missions=12, realms=3, achievements=10, concept=8, chapters=4, collections=15, gifts=5, prompts=6, seasons=1 |
| `/api/login` demo1 | POST | **PASS** 200 | uid=demo-uid-1, role=SEEKER, family_id=demo-crew-1 |
| `/api/login` demo2 | POST | **PASS** 200 | uid=demo-uid-2, role=GUIDE |
| `/api/login` demo3 | POST | **PASS** 200 | uid=demo-uid-3, role=BUILDER |
| `/api/login` invalid user | POST | **PASS** 401 | rejected |
| `/api/login` wrong password | POST | **PASS** 401 then 429 under load | rate limiter **enabled** (not disabled) |
| `/api/me` unauthenticated | GET | **PASS** 401 | |
| `/api/me` + Bearer demo1 | GET | **PASS** 200 | Leo SEEKER |
| `/api/missions` | GET | **PASS** 200 | ids 101, 102, 103 |
| `/api/missions/103` | GET | **PASS** 200 | title Riddle of the Stones |

Session tokens **redacted** from this report.

---

## 7. Local Auth Evidence

Login contract:

```json
{
  "uid": "<username>",
  "credential": "odyssey123",
  "device": { "device_id": "web-pwa", "login_method": "BOTH" }
}
```

| Account | Login | Profile UID | Role | Explorer |
| --- | --- | --- | --- | --- |
| demo1 | 200 | demo-uid-1 | SEEKER | Leo |
| demo2 | 200 | demo-uid-2 | GUIDE | Maya |
| demo3 | 200 | demo-uid-3 | BUILDER | Sam |
| invalid | 401 | — | — | — |

**Conclusion:** Local Auth is working on production. No migration to Supabase Auth.

---

## 8. Full E2E Evidence (Production Gate)

**Command:**

```powershell
cd web
$env:PLAYWRIGHT_TEST_BASE_URL="https://odyssey-beta-nine.vercel.app"
$env:SKIP_DB_RESET="true"
npm run test:e2e -- --reporter=list
```

| Metric | Value |
| --- | --- |
| Target | production only (no localhost webServer) |
| Total | **16** |
| Passed | **16** |
| Failed | **0** |
| Skipped | **0** |
| Retries | **0** |
| Workers | 1 |
| Browser | chromium |
| Duration | **3.0 minutes** |
| Exit code | **0** |

```text
Running 16 tests using 1 worker
  ok  1 Auth Domain › can login successfully
  ok  2 Auth Domain › can logout successfully
  ok  3 Golden Path › Test 1: Login -> Home -> Mission list appears
  ok  4 Golden Path › Test 2: Start Mission -> Complete Exercise -> XP increases -> Journal updates
  ok  5 Golden Path › Test 3: Logout -> Login -> Progress persists
  ok  6 Home Domain › displays home dashboard and missions
  ok  7 Journal Domain › can view journal entries
  ok  8 Multi-user › can login as demo1 and see dashboard
  ok  9 Multi-user › can login as demo2 and see dashboard
  ok 10 Multi-user › can login as demo3 and see dashboard
  ok 11 Persistence › session persists after reload
  ok 12 Persistence › state persists after logout and login
  ok 13 Mission Domain › can complete a challenge in a quest
  ok 14 Regression › deep-link to /#/missions/103 after login
  ok 15 Regression › browser back and forward navigation
  ok 16 Regression › unauthenticated deep-link redirects to login
  16 passed (3.0m)
```

### Production data safety

| Rule | Status |
| --- | --- |
| `resetQuestState()` | No-op unless localhost + service key + `SKIP_DB_RESET` not true |
| This run | `SKIP_DB_RESET=true`, production baseURL → **no DB patch** |
| No truncate/delete/reseed | Held |

---

## 9. Application / Test Changes Included in Release Commit

Intentional (reviewed) changes:

| Path | Purpose |
| --- | --- |
| `supabase/migrations/019_seed_local_users.sql` | Official migration for local users (synced) |
| `scripts/migrations/019_seed_local_users.sql` | Mirror of official migration |
| `web/playwright.config.ts` | Production-capable baseURL; no webServer for prod |
| `web/tests/e2e/**` | Production-hardened helpers, multi-user + regression specs |
| `pkg/db/supabase.go` | Prefer header handling for `return=representation` |
| `pkg/game/mission/service.go` | ACTIVE status when quest started but compute returns PENDING |
| `internal/api/dev/main.go` | dotenv path for local dev |
| `docs/release/*` | Evidence / RCA reports |

**Excluded from commit:** temporary audit scripts (`audit_*.js`, `curl_supabase.js`, `run_migration.js`, `verify_db.js`, `test_*.js`, `web/investigate*`, root junk `hash.go`/`verify_hash.go`), playwright-report/test-results junk, secrets.

---

## 10. Local CI-Equivalent Checks (Pre-Push)

| Check | Result |
| --- | --- |
| `go vet ./...` | PASS (without untracked junk mains) |
| `go test -count=1 ./...` | PASS (same) |
| `npx tsc --noEmit` | PASS |
| `npx eslint .` | PASS |
| `npm run test` (Vitest) | **16 passed** |
| Production E2E | **16 passed** |

Remote CI status is recorded after push/PR.

---

## 11. Security & Secret Handling

| Control | Status |
| --- | --- |
| Service keys / DB passwords / JWT secrets not in report | Yes |
| Session tokens redacted | Yes |
| Demo password `odyssey123` is seed QA credential only | Documented |
| Login rate limit not disabled | Yes |
| No secrets committed | Yes (`.env*` gitignored) |
| Hardcoded Supabase project URL removed from E2E helper | Yes (env-only) |

---

## 12. Remaining Non-Blocking Risks

| Risk | Severity | Notes |
| --- | --- | --- |
| `/api/home` cold/warm latency | Medium (perf) | Functional PASS; E2E uses 45s home wait |
| Login rate limit → 429 under rapid automation | Low | Expected; do not disable |
| Mission 103 list vs detail status drift | Low | Content intact; exercises complete from QA |
| Leo level/XP advanced vs seed baseline | Info | Expected QA progress — do not reseed |
| CLAUDE.md Gatekeeper wording vs Local Auth prototype | Docs | Runtime is Local Auth; Gatekeeper not modified |
| CI integration job uses dummy Supabase + incomplete backend start | Medium (CI fidelity) | Production E2E is the true release gate for this prototype |
| Shared demo accounts | Medium (ops) | Concurrent testers may race state |

---

## 13. Verified vs Not Verified

### Verified

- Production frontend HTTP 200
- Production API status/login/me/missions/103
- Local Auth for demo1/2/3 + invalid rejection
- Complete seed/content matrix via production DB counts + content samples
- schema_version 11
- Critical quest 103 content
- Production E2E 16/16
- Non-destructive safety (`resetQuestState` no-op; no reseed)
- Free-tier posture (no new paid infra)
- Local lint/unit checks

### Not verified in this session

- Gatekeeper / Firestore path (out of prototype Local Auth scope)
- Family Reward integration (Phase 5 deferred)
- Remote CI green **before** push (recorded after PR)
- Post-merge Vercel deployment (recorded after merge)
- Firefox/WebKit (Chromium gate only)
- Direct Postgres SQL console (REST used instead)
- Password hash bytes (intentionally not logged)

---

## 14. Exit Criteria Checklist

| Criterion | Status |
| --- | --- |
| Complete seed audit PASS | **PASS** |
| Production data PASS (not users-only) | **PASS** |
| API PASS | **PASS** |
| Local Auth PASS | **PASS** |
| Full E2E 16/0/0/0 | **PASS** |
| No destructive production ops | **PASS** |
| No secrets exposed | **PASS** |
| Free-tier only | **PASS** |

---

## 15. Final Decision (pre-merge)

```text
READY FOR QA
```

Subject to remote CI green + successful PR merge (tracked in Git section of the final handoff).
