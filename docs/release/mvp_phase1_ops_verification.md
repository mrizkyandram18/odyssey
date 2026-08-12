# Phase 1 MVP — Ops Verification Evidence

**Date:** 2026-08-10  
**Main tip / expected commit:** `206a8ac` (`v0.1.0`)  
**Verdict:** **COMPLETE** — demo DB + **production parity PASS**

## Production parity (post-redeploy)

| Item | Value |
| --- | --- |
| Expected commit | `206a8acadb80dfd05a42b60ffacb1ececaa67cc0` (`v0.1.0` / `origin/main`) |
| Includes #15 / #16 | **Yes** (ancestors of tag) |
| Prior production deploy | `dpl_Ck4aexJNsAWwKp1sf3WperoAXGVP` — **2026-08-10 13:00 +0700** (before #15/#16) |
| Redeploy | **PASS** — `vercel deploy --prod` from `main@206a8ac` |
| Deployed revision | `dpl_2RXRXs2mDoFZbkfKchWbuZaH6vCd` |
| Deployment URL | `https://odyssey-iu3ibsr6y-muhammad-rizky-andra-muchliss-projects.vercel.app` |
| Production alias | `https://odyssey-beta-nine.vercel.app` |
| Deploy created | **2026-08-10 20:16:14 +0700** |
| boot_time after redeploy | **2026-08-10T13:16:51Z** (changed vs pre-redeploy `13:14:58Z` / older `12:29:24Z`) |
| schema_version | `022_mvp_six_missions` |

### Why production lagged

- Vercel project **is not GitHub-auto-deploying** this repo (GitHub Deployments API count = 0).
- Prior production binary came from a **manual CLI deploy** at 13:00 +0700, before PRs #14–#16 landed.
- Fixes #15/#16 therefore never reached `odyssey-beta-nine` until explicit `vercel deploy --prod`.

### Production smoke (`https://odyssey-beta-nine.vercel.app`)

| Check | Result | Evidence |
| --- | --- | --- |
| Login demo1 | **PASS** | `uid=demo-uid-1` crew `demo-crew-1` |
| Mission load (6 WW) | **PASS** | count=6; types SOLO×4, RELAY, CREATIVE |
| SOLO complete | **PASS** | Mission 102 DONE type=SOLO |
| RELAY + assignment | **PASS** | Mission 104 DONE; leg2 `assigned_to=demo-uid-3` completed by demo3 |
| CREATIVE + Story | **PASS** | Mission 105 DONE type=CREATIVE; `next=CREATE_MEMORY`; POST `/api/creative` **201** |
| Daily turn | **PASS** | POST consume **200**, xp=10, date=2026-08-10 |
| Progression | **PASS** | L39/3890 → L43/4200 (**+310 XP**) |

## Earlier local+DB verification (pre-parity)

| Item | Result | Evidence |
| --- | --- | --- |
| Apply `022` MVP seed (6 WW missions) | **PASS** | Scoped REST on `demo-crew-1`; schema `022_mvp_six_missions` |
| Local binary smoke of #15/#16 | **PASS** | `go run ./internal/api/dev` against live Supabase |

## Mission matrix (demo-crew-1)

| ID | Title | Type |
| ---: | --- | --- |
| 101 | Morning Light | SOLO |
| 102 | Gather Herbs | SOLO |
| 103 | Riddle of the Stones | SOLO |
| 104 | Shadow Trail | RELAY |
| 105 | The Old Growth | CREATIVE |
| 106 | Forest Riddle | SOLO |

## Bugs fixed during MVP ops (shipped in #15/#16)

| Bug | Fix | PR |
| --- | --- | --- |
| Daily turn create POST `id:0` → 500 | `CreateDailyTurn` map without id | #15 |
| Creative CSRF 403 from UI client | `apiClient` fetches `/api/csrf` + `X-CSRF-Token` | #15 |
| CREATE_MEMORY rejected on DONE quest | Allow DONE quest for memory submit | #15 |
| `SubmitRequest` missing JSON tags | Add `mission_id` / `exercise_id` / … tags | #16 |
| Missing `family_id` on submit | Set from quest before insert | #16 |
| Creative insert `id:0` payload | Map without identity id | #16 |

## Residual backlog (not release blockers)

- Enable Vercel **Git integration** / auto-deploy on `main` so future merges promote without manual CLI.
- Phase 2 leftovers remain untracked (not committed): `release_report_phase2b.md`, `021_seed_phase2a.sql`, extra e2e specs.
- Gatekeeper BOTH, heavy E2E CI, multi-journey: deferred.

## Release tag

- Tag: **`v0.1.0`** → commit **`206a8ac`**
- Production alias serves deployment built from that tip (manual prod redeploy after tag)
- CI minimal (Lint / Unit / Build) green on PRs #14–#17
