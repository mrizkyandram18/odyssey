# Phase 1 MVP — Ops Verification Evidence

**Date:** 2026-08-10  
**Main tip:** `f19cceb`  
**Verdict:** **COMPLETE** (demo DB + code path verified; see notes)

## Scope verified

| Item | Result | Evidence |
| --- | --- | --- |
| Apply `022` MVP seed (6 WW quests) | **PASS** | Scoped REST apply on `demo-crew-1` only; schema `022_mvp_six_quests` |
| Login `demo1` | **PASS** | Local Auth `uid=demo-uid-1`, crew `demo-crew-1` |
| Exactly 6 Whispering Woods quests | **PASS** | API list count=6; titles match matrix below |
| SOLO start+complete | **PASS** | Quest 103 (and earlier 102 on production API) |
| RELAY 2 legs + assignment | **PASS** | Quest 104; leg2 `assigned_to` → demo-uid-2/3; completed by other user |
| CREATIVE + Story memory | **PASS** | Quest 105 `CREATE_MEMORY`; POST `/api/creative` **201** with CSRF |
| Daily turn consume once | **PASS** | POST `/api/daily_turns/consume` **200**, XP awarded |
| Progression XP delta | **PASS** | e.g. level 36→39, xp +330 during smoke (includes quests + daily) |

## Quest matrix (demo-crew-1)

| ID | Title | Type | Notes |
| ---: | --- | --- | --- |
| 101 | Morning Light | SOLO | PENDING at end of smoke (playable) |
| 102 | Gather Herbs | SOLO | Completed in production API smoke |
| 103 | Riddle of the Stones | SOLO | Completed local+live DB |
| 104 | Shadow Trail | RELAY | Assignment handoff verified |
| 105 | The Old Growth | CREATIVE | Story submission id=3 |
| 106 | Forest Riddle | SOLO | PENDING (playable) |

## Environment

| Layer | Target |
| --- | --- |
| Database | Supabase project `hmrkssfhcxlvjzyigufd` (demo/prod shared prototype) |
| Seed | Migration semantics of `supabase/migrations/022_seed_demo_quests.sql` applied via scoped REST (no non-demo crews touched) |
| API for full smoke after fixes | **Local** `go run ./internal/api/dev` on `:8080` against live Supabase |
| Production URL | `https://odyssey-beta-nine.vercel.app` — quest list/SOLO/RELAY verified; daily/creative fixes need redeploy of `f19cceb` |

## Bugs found & fixed during ops (small MVP blockers)

| Bug | Fix | PR |
| --- | --- | --- |
| Daily turn create POST `id:0` → 500 | `CreateDailyTurn` map without id | #15 |
| Creative CSRF 403 from UI client | `apiClient` fetches `/api/csrf` + `X-CSRF-Token` | #15 |
| CREATE_MEMORY rejected on DONE quest | Allow DONE quest for memory submit | #15 |
| `SubmitRequest` missing JSON tags | Add `quest_id` / `challenge_id` / … tags | #16 |
| Missing `crew_id` on submit | Set from quest before insert | #16 |
| Creative insert `id:0` payload | Map without identity id | #16 |

## Errors / residual notes

- Production Vercel instance **did not auto-redeploy** within ops window (`boot_time` stuck). Full daily/creative path verified on **local binary + live DB**. Redeploy production to pick up #15/#16.
- `/api/home` PowerShell client hit `HOME` env var name collision (client-side only); not an app bug.
- Phase 2 leftovers remain **untracked** and were **not** committed (`release_report_phase2b.md`, `021_seed_phase2a.sql`, extra e2e specs).
- Gatekeeper BOTH, heavy E2E CI, multi-realm: still deferred (MVP backlog).

## Release tag

- Tag: `v0.1.0` on `main` @ `f19cceb` after this verification.
- CI minimal (Lint / Unit / Build) green on merge PRs #14–#16.
