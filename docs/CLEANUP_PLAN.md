# ODYSSEY SYSTEM CLEANUP & HARDENING PLAN

## Target Product Scope
A private family-only daily task platform where members:
1. Receive assigned daily activities (YouTube learning, Quizzes, Document uploads, Photo proofs).
2. Complete activities and submit validations.
3. Earn coins and EXP upon valid completion.
4. Accumulate coin balance in an immutable ledger.
5. Redeem coins in the Reward Shop for real-world payouts (Pulsa, E-Wallets, Cash), approved and disbursed by family admins.

All legacy cooperative RPG engine components (missions, chapters, realms, chests, relics, drop tables, comics, story fragments, social reactions, daily turns, seasons, cosmetics) are fully deprecated and removed.

---

## 1. Critical Security & Financial Fixes

### 1.1 Infinite Coin Exploit in `odyssey_submit_auto_task`
- **Vulnerability:** RPC `odyssey_submit_auto_task` executes `ON CONFLICT (task_id, user_uid) DO UPDATE` without checking if the task was already approved. Re-submitting triggers `odyssey_coin_transactions` INSERT and adds coins to `odyssey_user_profiles` indefinitely.
- **Fix:** 
  1. Add strict guard at top of `odyssey_submit_auto_task`:
     ```sql
     IF EXISTS (
         SELECT 1 FROM odyssey_task_submissions 
         WHERE task_id = p_task_id AND user_uid = p_user_uid AND status = 'APPROVED'
     ) THEN
         RAISE EXCEPTION 'Tugas ini sudah diselesaikan dan reward sudah diklaim' USING ERRCODE = 'P0004';
     END IF;
     ```
  2. Add atomic row locking and regression tests for normal retry, double-click, and parallel submissions.

### 1.2 Cleartext Quiz Answer Leakage in `/api/tasks/today`
- **Vulnerability:** `HandleGetToday` returns unstripped `task.config` JSON containing `correct_answer`, exposing answer keys to browser DevTools.
- **Fix:** In `internal/api/family_tasks/api.go`, sanitize the `questions` array before serialization by stripping `correct_answer`, `expected_answer`, `answer_key`, and `is_correct`.
- **Regression Test:** Assert that `correct_answer` does NOT appear anywhere in the `/api/tasks/today` HTTP response.

### 1.3 Strict Family Multi-Tenant Isolation
- **Vulnerability:** Tasks, submissions, and claims do not verify family boundaries. An admin from Family A can view/approve resources from Family B.
- **Fix:**
  1. Add `family_id TEXT REFERENCES odyssey_families(id)` column to `odyssey_tasks`.
  2. Backfill existing tasks with family of `created_by` or default family.
  3. Scope all queries by authenticated user's `claims.FamilyID`:
     - `HandleGetToday`: filter `family_id = claims.FamilyID`.
     - `HandleSubmit`: verify task belongs to `claims.FamilyID`.
     - `HandleListTasks`: filter `family_id = claims.FamilyID`.
     - `HandleCreateTask`: inject `family_id = claims.FamilyID`.
     - `HandleListPendingSubmissions`: filter submissions by family.
     - `HandleVerifySubmission`: verify admin and submission share same `family_id`.
     - `HandleAdminListClaims` & `HandleAdminProcessClaim`: filter/verify claim user belongs to admin's `family_id`.

---

## 2. Files and Packages to Delete

### 2.1 Backend Packages (`pkg/` and `internal/`)
| Directory / Package | Reason |
| :--- | :--- |
| `pkg/game/*` (all subpackages: mission, course, concepts, gift, collection, creative, etc.) | Dead RPG gamification engine |
| `pkg/content/*` | Legacy content repository and caching gateway |
| `internal/api/admin/` | Legacy game admin handlers |
| `internal/api/missions/` | Legacy quest/exercise HTTP endpoints |
| `internal/api/daily_missions/` | Legacy daily turns HTTP endpoints |
| `internal/api/daily_activities/` | Legacy activity engine HTTP endpoints |
| `internal/api/courses/` | Legacy course/chapter HTTP endpoints |
| `internal/api/concepts/` | Legacy concept unlocks HTTP endpoints |
| `internal/api/story_fragments/` | Legacy story fragments HTTP endpoints |
| `internal/api/achievements/` | Legacy achievements HTTP endpoints |
| `internal/api/gifts/` | Legacy gift chests HTTP endpoints |
| `internal/api/collections/` | Legacy relic collection HTTP endpoints |
| `internal/api/creative/` | Legacy SVG/comic HTTP endpoints |
| `internal/api/board/` | Legacy text board HTTP endpoints |
| `internal/api/reactions/` | Legacy reactions HTTP endpoints |
| `internal/api/seasons/` | Legacy seasonal operations HTTP endpoints |
| `internal/api/home/` | Replaced by direct `/api/tasks/today` |
| `internal/api/cosmetics/` | Dead cosmetic unlocks API |
| `internal/api/rewards/` | Dead legacy reward signals API |
| `internal/api/journey_progress/` | Legacy journey progress API |

### 2.2 Backend Legacy Database Repositories (`pkg/db/`)
- `activity.go`, `balance.go`, `chest.go`, `chest_definition.go`, `concept_unlock.go`, `content.go`, `content_admin_store.go`, `content_store.go`, `content_store_impl.go`, `cosmetic.go`, `course_progress.go`, `creative.go`, `creative_submissions.go`, `daily_activity_engine.go`, `daily_missions.go`, `journey_progress.go`, `missions.go`, `player_relic.go`, `progression.go`, `reaction.go`, `relic.go`, `relic_definition.go`, `reward.go`, `reward_signals.go`.

### 2.3 Frontend Dead Features & Components (`web/src/`)
| Directory / File | Reason |
| :--- | :--- |
| `web/src/features/mission/*` | Dead quest/canvas pages |
| `web/src/features/creative/*` | Dead comic/canvas/board pages |
| `web/src/features/gifts/*` | Dead chest opening pages |
| `web/src/features/collections/*` | Dead relic inventory pages |
| `web/src/features/journal/*` | Dead journal pages |
| `web/src/features/home/DailyActivitySection.tsx` | Unused in linear task home |
| `web/src/shared/components/organisms/*` (WorldMap, CreativeCanvas, MissionDetail, FamilyDashboard) | Dead RPG organisms |
| `web/src/shared/components/molecules/*` (ReactionBar, ReactionList, ReactionPicker, CreativeCard, MissionCard, RelayRotation, SeasonBadge, StreakBadge, YourTurnBadge, DailyTurnBanner, CollectionDisplay, ConnectedReactionBar) | Dead RPG molecules |
| `web/src/shared/components/layout/MobileNav.tsx`, `Sidebar.tsx`, `PageTransition.tsx` | Replaced by `BottomNav.tsx` |
| `web/src/shared/utils/*` (comic.ts, journey.ts, media.ts, missionTurn.ts, relayRotation.ts, svg.ts, roleMastery.ts) | Dead RPG utility libraries |
| `web/src/shared/hooks/*` (useMission.ts, useReactions.ts) | Dead RPG hooks |

---

## 3. Database Migration `044_cleanup_legacy_family_platform.sql`

A new additive cleanup migration will be executed to:
1. **Alter `odyssey_tasks`:** Add `family_id TEXT REFERENCES odyssey_families(id)`. Backfill with existing family ID.
2. **Alter `odyssey_claims`:** Add index on `(user_uid, status)`.
3. **Replace RPC `odyssey_submit_auto_task`:** Enforce anti-double-claim guard and family check.
4. **Replace RPC `odyssey_verify_submission`:** Enforce family verification for admin.
5. **Drop Verified Dead Legacy Tables:**
   - `odyssey_reactions_legacy`, `odyssey_story_fragments`, `odyssey_player_story_fragments`, `odyssey_lore_definitions`
   - `odyssey_creative_items`, `odyssey_creative_prompt_definitions`, `odyssey_creative_submissions`
   - `odyssey_drop_tables`, `odyssey_gift_definitions`, `odyssey_gifts`
   - `odyssey_collection_definitions`, `odyssey_collections`, `odyssey_player_collections`
   - `odyssey_exercises`, `odyssey_mission_definitions`, `odyssey_missions`
   - `odyssey_course_definitions`, `odyssey_course_progress`
   - `odyssey_journey_definitions`, `odyssey_journey_progress`
   - `odyssey_learning_concepts`
   - `odyssey_daily_activities`, `odyssey_daily_activity`, `odyssey_daily_activity_completions`, `odyssey_daily_missions`
   - `odyssey_achievements`, `odyssey_achievement_definitions`
   - `odyssey_reactions`, `odyssey_reward_signals`, `odyssey_season_definitions`, `odyssey_balance_configs`, `odyssey_cosmetic_unlocks`, `odyssey_reward_ledgers`.

---

## 4. Retained Core Architecture

### Core Database Tables:
1. `odyssey_families`
2. `odyssey_local_users`
3. `odyssey_user_profiles`
4. `odyssey_tasks`
5. `odyssey_task_submissions`
6. `odyssey_coin_transactions` (Immutable Ledger)
7. `odyssey_reward_catalog`
8. `odyssey_claims`
9. `odyssey_audit_logs`
10. `odyssey_system_config`
11. `odyssey_schema_version`
12. `odyssey_push_subscriptions`

### Core Backend API Routes:
- `POST /api/login` (Authentication)
- `GET /api/me` (Profile)
- `GET /api/csrf` (CSRF Token)
- `GET /api/tasks/today` (Member daily linear tasks)
- `POST /api/tasks/upload` (Storage upload for proof files)
- `POST /api/tasks/:id/submit` (Submit auto-quiz or manual verification)
- `GET /api/shop/items` (Reward Catalog)
- `POST /api/shop/redeem` (Member submit coin redemption claim)
- `GET /api/shop/claims` (Member redemption claim history)
- `GET /api/admin/tasks` (Admin task list)
- `POST /api/admin/tasks` (Admin task creation)
- `PATCH, DELETE /api/admin/tasks/:id` (Admin task update/delete)
- `GET /api/admin/submissions/pending` (Admin review queue)
- `POST /api/admin/submissions/:id/verify` (Admin approve/reject task submission)
- `GET /api/admin/claims` (Admin redemption claim queue)
- `POST /api/admin/claims/:id/process` (Admin approve payout or reject with refund)
- `GET /health`, `/ready`, `/live`, `/version`, `/metrics` (Observability)

### Core Frontend Pages:
- `/login` (`LoginPage`)
- `/` (`HomePage` with `LinearPath`, `StepNode`, `VideoQuizModal`, `DocUploadModal`, `CameraCaptureModal`)
- `/shop` (`RewardShopPage` with `RedeemModal`)
- `/profile` (`ProfilePage` with avatar customization, streak, XP, level, coin balance, and account settings)
- `/admin` (`AdminPage` with submissions verification queue, payout claims processing, and task schedule CRUD)

---

## 5. Verification Commands
```bash
# 1. Backend tests
go test -v -race -count=1 ./...

# 2. Frontend tests & linting
npm test --prefix web
npm run build --prefix web
```
