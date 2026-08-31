# ODYSSEY FAMILY TASK PLATFORM — PHASE 2 CLEANUP & HARDENING REPORT

## Executive Summary
Odyssey has been transformed from an oversized, legacy cooperative RPG engine into a lean, hardened, single-purpose **Family Daily Task & Reward Platform**. All obsolete RPG systems (missions, journeys, realms, chapters, lore fragments, chests, relics, drop tables, comic canvas, social reactions, daily turns, seasons, cosmetics) have been eliminated.

---

## 1. What Was Deleted

### 1.1 Backend Packages & Handlers (`pkg/` and `internal/api/`)
- `pkg/game/*`: Completely deleted all 20+ legacy game packages (`mission`, `course`, `concepts`, `gift`, `collection`, `creative`, `dailyactivity`, `dailymission`, `familystreak`, `fragment`, `season`, `social`, `validation`, `world`, `reward`, `rewardsignal`, `balance`, `board`, `cosmetic`, `audit`, `events`).
- `pkg/content/*`: Completely deleted legacy content caching and definition stores.
- `pkg/db/*`: Deleted 38 obsolete database store files and mock tests querying renamed or dead RPG tables (`activity.go`, `balance.go`, `chest.go`, `chest_definition.go`, `concept_unlock.go`, `content.go`, `content_admin_store.go`, `content_store.go`, `content_store_impl.go`, `cosmetic.go`, `course_progress.go`, `creative.go`, `creative_submissions.go`, `daily_activity_engine.go`, `daily_missions.go`, `journey_progress.go`, `missions.go`, `player_relic.go`, `progression.go`, `reaction.go`, `relic.go`, `relic_definition.go`, `reward.go`, `reward_signals.go`, `user.go`, `wire.go`).
- `internal/api/*`: Deleted obsolete endpoints (`achievements/`, `admin/`, `board/`, `collections/`, `concepts/`, `cosmetics/`, `courses/`, `creative/`, `daily_activities/`, `daily_missions/`, `gifts/`, `home/`, `journey_progress/`, `missions/`, `reactions/`, `rewards/`, `seasons/`, `story_fragments/`).

### 1.2 Frontend Dead Features & Components (`web/src/`)
- `features/mission/*`: Deleted dead quest/exercise screens (`MissionsPage.tsx`, `MissionView.tsx`, `CreativeCanvas.tsx`).
- `features/creative/*`: Deleted dead creative studio (`CreativePage.tsx`, `ComicReaderPage.tsx`, `GalleryPage.tsx`, `StoryPage.tsx`, `FamilyTimeline.tsx`, `SharedTextBoard.tsx`, `SubmissionForm.tsx`).
- `features/gifts/*`: Deleted dead chest opening screens (`GiftPage.tsx`, `GiftOpeningPage.tsx`).
- `features/collections/*`: Deleted dead relic collection views (`CollectionInventoryPage.tsx`, `CollectionDetailPage.tsx`).
- `features/journal/*`: Deleted dead journal view (`JournalPage.tsx`).
- `features/home/DailyActivitySection.tsx`: Deleted obsolete RPG turn component.
- Shared Organisms: Deleted `WorldMap.tsx`, `CreativeCanvas.tsx`, `MissionDetail.tsx`, `FamilyDashboard.tsx`.
- Shared Molecules: Deleted `ReactionBar.tsx`, `ReactionList.tsx`, `ReactionPicker.tsx`, `CreativeCard.tsx`, `MissionCard.tsx`, `RelayRotation.tsx`, `SeasonBadge.tsx`, `StreakBadge.tsx`, `YourTurnBadge.tsx`, `DailyTurnBanner.tsx`, `CollectionDisplay.tsx`, `ConnectedReactionBar.tsx`.
- Shared Layouts: Deleted obsolete `MobileNav.tsx`, `Sidebar.tsx`, `PageTransition.tsx`.
- Shared Utils & Hooks: Deleted `comic.ts`, `journey.ts`, `media.ts`, `missionTurn.ts`, `relayRotation.ts`, `svg.ts`, `roleMastery.ts`, `useMission.ts`, `useReactions.ts`.

### 1.3 Obsolete Database Tables (Migration `044`)
- `odyssey_reactions_legacy`, `odyssey_player_story_fragments`, `odyssey_story_fragments`, `odyssey_lore_definitions`
- `odyssey_creative_submissions`, `odyssey_creative_items`, `odyssey_creative_prompt_definitions`
- `odyssey_drop_tables`, `odyssey_gift_definitions`, `odyssey_gifts`
- `odyssey_player_collections`, `odyssey_collection_definitions`, `odyssey_collections`
- `odyssey_exercises`, `odyssey_mission_definitions`, `odyssey_missions`
- `odyssey_course_progress`, `odyssey_course_definitions`
- `odyssey_journey_progress`, `odyssey_journey_definitions`
- `odyssey_learning_concepts`
- `odyssey_daily_activity_completions`, `odyssey_daily_activities`, `odyssey_daily_activity`, `odyssey_daily_missions`
- `odyssey_achievement_definitions`, `odyssey_achievements`
- `odyssey_reactions`, `odyssey_reward_signals`, `odyssey_season_definitions`, `odyssey_balance_configs`, `odyssey_cosmetic_unlocks`, `odyssey_reward_ledgers`

---

## 2. What Was Kept (And Why)

### 2.1 Core Active Entities & Tables
1. `odyssey_families`: Scopes all members, tasks, and redemptions to their private family domain.
2. `odyssey_local_users`: Bcrypt username and password authentication credentials.
3. `odyssey_user_profiles`: Member profile identity (name, family ID, role, coins, level, EXP, streak).
4. `odyssey_tasks`: Daily activities created by family admins (YouTube videos, quizzes, document uploads, photo proofs) with `family_id` scoping and `active_date`.
5. `odyssey_task_submissions`: Member submissions with status (`PENDING`, `APPROVED`, `REJECTED`), answers/proof URLs, and admin review notes.
6. `odyssey_coin_transactions`: Authoritative immutable financial ledger recording every coin movement (`TASK_REWARD`, `REWARD_REDEMPTION`, `REWARD_REFUND`).
7. `odyssey_reward_catalog`: Items available for coin redemption (Pulsa, GoPay, DANA, OVO, Cash).
8. `odyssey_claims`: Member redemption requests processed by family admins.
9. `odyssey_push_subscriptions`: Web Push notification endpoints for daily activity reminders.
10. `odyssey_audit_logs` & `odyssey_system_config` & `odyssey_schema_version`: System operations, configuration, and migration tracking.

### 2.2 Core Backend APIs
- `/api/login` & `/api/me`: Local authentication, session issuance, and profile avatar customization.
- `/api/csrf`: Anti-CSRF token endpoint.
- `/api/tasks/today`, `/api/tasks/upload`, `/api/tasks/:id/submit`: Member daily linear task stepper, proof uploads to Supabase storage bucket `task-proofs`, and atomic quiz/manual submission RPCs.
- `/api/shop/items`, `/api/shop/redeem`, `/api/shop/claims`: Reward catalog, coin redemption with atomic ledger deduction, and member claim history.
- `/api/admin/tasks`, `/api/admin/submissions`, `/api/admin/claims`: Admin task schedule CRUD, photo/document verification review queue, and redemption payout/refund processing.
- `/health`, `/ready`, `/live`, `/version`, `/metrics`: Operational monitoring.

### 2.3 Core Frontend Pages
- `/login`: Clean local username/password login.
- `/`: Linear step-by-step Daily Petualangan path (`LinearPath`, `VideoQuizModal`, `DocUploadModal`, `CameraCaptureModal`).
- `/shop`: Reward Shop catalog (`RewardShopPage`, `RedeemModal`, Claim History).
- `/profile`: Explorer profile identity (`ProfilePage`), level, EXP progress, streak count, coin balance with quick redeem link, avatar randomizer, push notification toggle, and sign out.
- `/admin`: Admin dashboard (`AdminPage`) with verification queue (with photo zoom), claims payout queue (with status update & refund), and task schedule manager.

---

## 3. Security & Integrity Fixes

### 3.1 Fixed Critical Infinite Coin Exploit
- **Root Cause:** RPC `odyssey_submit_auto_task` used `ON CONFLICT DO UPDATE` without checking if the task was already approved. Re-submitting the task repeatedly inserted new ledger rows and added coins indefinitely.
- **Fix:** Added strict database-level guard:
  ```sql
  IF EXISTS (
      SELECT 1 FROM odyssey_task_submissions 
      WHERE task_id = p_task_id AND user_uid = p_user_uid AND status = 'APPROVED'
  ) THEN
      RAISE EXCEPTION 'Tugas ini sudah diselesaikan dan reward sudah diterima' USING ERRCODE = 'P0004';
  END IF;
  ```
  Enforces the invariant: **1 User + 1 Task = Maximum 1 Approved Reward**.

### 3.2 Eliminated Cleartext Quiz Answer Leakage
- **Root Cause:** `GET /api/tasks/today` returned unstripped `task.config` JSON containing `correct_answer`, exposing all answer keys in DevTools.
- **Fix:** Implemented `sanitizeQuestions()` in `internal/api/family_tasks/api.go` which strips `correct_answer`, `expected_answer`, `is_correct`, and `answer_key` before JSON serialization.

### 3.3 Strict Multi-Tenant Family Isolation
- Added `family_id TEXT REFERENCES odyssey_families(id)` column to `odyssey_tasks` in migration `044`.
- Scoped all member and admin endpoints (`/tasks/today`, `/tasks/:id/submit`, `/admin/tasks`, `/admin/submissions/pending`, `/admin/submissions/:id/verify`, `/admin/claims`, `/admin/claims/:id/process`) strictly by the authenticated session's `claims.FamilyID`.
- An admin from Family A cannot view, edit, or approve tasks, submissions, or claims belonging to Family B.

---

## 4. Test Results

### 4.1 Backend Test Suite (`go test -v -count=1 ./...`)
```text
ok  	odyssey/internal/api/admin_tasks	(PASS)
ok  	odyssey/internal/api/families   	(PASS)
ok  	odyssey/internal/api/family_tasks	(PASS)
ok  	odyssey/internal/api/login      	(PASS)
ok  	odyssey/internal/api/me         	(PASS)
ok  	odyssey/internal/api/push       	(PASS)
ok  	odyssey/internal/api/shop       	(PASS)
ok  	odyssey/internal/api/status     	(PASS)
ok  	odyssey/pkg/auth                	(PASS)
ok  	odyssey/pkg/db                  	(PASS)
ok  	odyssey/pkg/observability       	(PASS)
ok  	odyssey/pkg/push                	(PASS)
ok  	odyssey/pkg/shared              	(PASS)
```
**100% PASS** — All unit tests and regression tests passed.

### 4.2 Frontend Build (`npm run build`)
```text
vite building client environment for production...
✓ 2646 modules transformed.
dist/index.html                   1.23 kB │ gzip:   0.57 kB
dist/assets/index-BYHX3Uec.css   33.59 kB │ gzip:   6.67 kB
dist/assets/index-DYtg13Ku.js   763.05 kB │ gzip: 256.88 kB
✓ built in 26.26s
```
**100% PASS** — TypeScript compilation (`tsc -b`) and Vite production bundle generated cleanly with zero errors.

---

## 5. Final Architecture

```text
                                 ┌─────────────────────────┐
                                 │     AUTH & SESSION      │
                                 │  (/api/login, /api/me)  │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │     FAMILY SCOPING      │
                                 │  (claims.FamilyID)      │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │       DAILY TASKS       │
                                 │  YouTube / Quiz / Proof │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │     TASK SUBMISSION     │
                                 │   Auto Quiz / Review    │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │       COIN LEDGER       │
                                 │    Immutable Ledger     │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │       REWARD SHOP       │
                                 │   Pulsa / E-Wallet /    │
                                 │          Cash           │
                                 └────────────┬────────────┘
                                              │
                                              ▼
                                 ┌─────────────────────────┐
                                 │    ADMIN REDEMPTION     │
                                 │  Approved / Refunded    │
                                 └─────────────────────────┘
```
