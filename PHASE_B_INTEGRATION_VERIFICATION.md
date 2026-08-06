# Phase B Integration Verification Report

**Date:** 2026-08-05
**Scope:** P0 Event Wiring — QuestCompletedEvent, ChapterCompletedEvent, LevelReachedEvent, DailyTurnCompletedEvent, CreativeSubmissionEvent
**Directive:** Verification/audit only. No production code changes unless a critical blocker is discovered.

---

## 1. Event Flow Verification

### 1.1 Quest Completion Chain

**Specified:** QuestCompletedEvent -> Chest creation -> Achievement evaluation -> Realm progression

**Actual wiring (api/dev/main.go):**
- `QuestCompletedEvent` is published by `QuestAPIHandler.publishQuestCompleted()` (line 197-218 of `pkg/game/quest/handler.go`)
- Subscribed handlers:
  - `chest.NewQuestCompletedHandler(chestSvc, contentSvc)` — creates chest instance if quest has `RewardChest`
  - `achievement.NewQuestCompletedHandler(achieveSvc)` — evaluates quest-completed achievements
  - `chapter.NewQuestCompletedHandler(chapterSvc)` — checks chapter completion, emits `ChapterCompletedEvent`

**Realm progression** is performed directly in `QuestAPIHandler.CompleteChallenge()` via `h.advanceRealm()` (line 126-127), **before** the `QuestCompletedEvent` is published. It is NOT triggered by the event itself.

**Verdict:** Chain is wired correctly for chest creation, achievement evaluation, and chapter progression. Realm progression is decoupled from the event (done directly), which is acceptable for idempotency but means realm progress is not replay-safe via the event system.

### 1.2 Chapter Completion Chain

**Specified:** ChapterCompletedEvent -> Lore unlock -> Chapter achievement

**Actual wiring:**
- `ChapterCompletedEvent` is published by `ChapterService.MarkComplete()` (line 337-343 of `pkg/game/chapter/service.go`)
- Subscribed handlers:
  - `lore.NewChapterCompletedHandler(loreSvc)` — unlocks lore for the chapter
  - `achievement.NewChapterCompletedHandler(achieveSvc)` — evaluates chapter-completed achievements

**Verdict:** Correctly wired.

### 1.3 Level Progression Chain

**Specified:** AwardXP -> LevelReachedEvent -> Achievement evaluation

**Actual wiring:**
- `LevelReachedEvent` is published by `ProgressionService.AwardXP()` (line 129-135 of `pkg/game/progression/service.go`)
- Subscribed handler:
  - `achievement.NewLevelReachedHandler(achieveSvc)` — evaluates level-reached achievements

**CRITICAL:** `ProgressionService` is constructed with `NopPublisher` at `api/dev/main.go:144` via `progression.NewProgressionService(repo.Users, &progCfg)`. There is no `SetPublisher(dispatcher)` call and no `NewProgressionServiceWithPublisher` usage. **`LevelReachedEvent` is never published in production.**

**Verdict:** Chain is broken at the publisher level. `LevelReachedEvent` is never emitted.

### 1.4 Daily Turn Chain

**Specified:** ConsumeDailyTurn -> DailyTurnCompletedEvent -> Streak tracking

**Actual wiring:**
- `DailyTurnCompletedEvent` is published by `DailyTurnAPIHandler.Consume()` (line 69-72 of `pkg/game/dailyturn/handler.go`)
- Subscribed handler:
  - `achievement.NewDailyTurnCompletedHandler(achieveSvc)` — evaluates daily-streak achievements

**CRITICAL:** `DailyTurnAPIHandler` is constructed with `NopPublisher` at `api/dev/main.go:149` via `dailyturn.NewDailyTurnAPIHandler(dailyTurnSvc, progSvc)`. There is no `SetPublisher(dispatcher)` call and no `NewDailyTurnAPIHandlerWithPublisher` usage. **`DailyTurnCompletedEvent` is never published in production.**

**Verdict:** Chain is broken at the publisher level. `DailyTurnCompletedEvent` is never emitted.

### 1.5 Creative Submission Chain

**Specified:** Submit -> CreativeSubmissionEvent -> Achievement evaluation

**Actual wiring:**
- `CreativeSubmissionEvent` is published by `CreativeService.Submit()` (line 95-101 of `pkg/game/creative/service.go`)
- Subscribed handler:
  - `achievement.NewCreativeSubmissionHandler(achieveSvc)` — evaluates creative-submission achievements

**Verdict:** Correctly wired.

### 1.6 Chest Flow

**Specified:** Quest reward chest -> Chest instance -> Open chest -> Relic reward -> Relic achievement

**Actual wiring:**
- Quest completion creates a chest via `chest.QuestCompletedHandler.Handle()` (line 282-298 of `pkg/game/chest/service.go`)
- `OpenChest` is triggered via the `api/chests` HTTP handler (POST to `/api/chests/<id>/open`)
- `OpenChest` publishes `RelicCollectedEvent` for new relics (line 197-202 of `pkg/game/chest/service.go`)
- `RelicCollectedEvent` is subscribed by `achievement.NewRelicCollectedHandler(achieveSvc)` (main.go:129)

**`ChestCreatedEvent`** is published by `CreateChest()` (line 104-108 of `pkg/game/chest/service.go`) but has **no subscribers** in the dispatcher wiring. This event is emitted into a void.

**Verdict:** Chest creation, opening, relic reward, and relic achievement chains are wired correctly. `ChestCreatedEvent` has no consumers.

---

## 2. Duplicate Event Risks

| Risk Area | Severity | Finding |
|---|---|---|
| Duplicate chest creation | **HIGH** | `QuestCompletedHandler.Handle()` in chest/service.go creates a chest every time it receives a `QuestCompletedEvent`. No idempotency check exists. If the event is replayed, a duplicate chest is created. |
| Duplicate achievement progress | LOW | `AchievementService.evaluate()` checks `GetAchievementByCode` before creating. Already-unlocked achievements are skipped. |
| Duplicate realm progression | LOW | `advanceRealm()` is called only when `questCompleted` is true, which is false on retry (quest already DONE). |
| Duplicate level events | **MEDIUM** | `AwardXP()` publishes `LevelReachedEvent` on every call that causes a level-up. If `CompleteChallenge` is retried, XP is re-awarded and a duplicate `LevelReachedEvent` is published. |
| Duplicate chapter completion | LOW | `handleQuestCompleted()` checks `cp.Status == ChapterStatusComplete` and returns early. |
| Duplicate daily turn streak | LOW | `DailyTurnCompletedEvent` triggers achievement evaluation which checks existing unlocks. |
| Duplicate creative submission | **MEDIUM** | `CreativeService.Submit()` has no dedup. If the same API call is retried, a duplicate submission and event are created. |

---

## 3. Replay / Idempotency Risks

| Risk | Severity | Details |
|---|---|---|
| No event dedup in Dispatcher | **HIGH** | `Dispatcher.Publish()` has no event ID, no dedup tracking, and no replay protection. Every publish delivers to every handler. |
| Quest XP double-counting on retry | **HIGH** | `CompleteChallenge()` calls `AwardXP()` for challenge XP before checking if the quest was newly completed. On retry, XP is re-awarded even though the challenge was already done. |
| Chest creation on event replay | **HIGH** | `QuestCompletedHandler.Handle()` creates a new chest instance on every `QuestCompletedEvent` with no check for prior creation. |
| `LevelReachedEvent` never fires | **CRITICAL** | `ProgressionService` uses `NopPublisher`. Even if `AwardXP()` is called correctly, no `LevelReachedEvent` is emitted. |
| `DailyTurnCompletedEvent` never fires | **CRITICAL** | `DailyTurnAPIHandler` uses `NopPublisher`. Even if `Consume()` succeeds, no `DailyTurnCompletedEvent` is emitted. |
| Realm progression not event-driven | MEDIUM | `advanceRealm()` is a direct DB call, not triggered by an event. This is idempotent (only fires on `questCompleted=true`) but means realm progression cannot be reconstructed from event replay. |

---

## 4. Integration Risks

### 4.1 Critical Blockers

1. **`DailyTurnCompletedEvent` never published** — `DailyTurnAPIHandler` at `api/dev/main.go:149` is constructed with `NewDailyTurnAPIHandler` (NopPublisher). The `Consume()` method at `pkg/game/dailyturn/handler.go:69` calls `h.publisher.Publish()` but the publisher is a no-op. The streak tracking and daily-streak achievements are non-functional.

2. **`LevelReachedEvent` never published** — `ProgressionService` at `api/dev/main.go:144` is constructed with `NewProgressionService` (NopPublisher). The `AwardXP()` method at `pkg/game/progression/service.go:129` calls `s.publisher.Publish()` but the publisher is a no-op. Level-up achievements and any level-reached consumers are non-functional.

3. **`ChestCreatedEvent` has no subscribers** — Published by `ChestService.CreateChest()` but no handler is subscribed to `EventTypeChestCreated` in the dispatcher wiring. If any downstream system needs to react to chest creation, it will not receive events.

### 4.2 High-Severity Risks

4. **Quest XP double-counting on retry** — `QuestAPIHandler.CompleteChallenge()` awards challenge XP before checking `questCompleted`. If the HTTP handler retries the same challenge completion, XP is double-counted.

5. **Duplicate chest creation on event replay** — `chest.QuestCompletedHandler.Handle()` has no idempotency check. Replaying a `QuestCompletedEvent` creates a duplicate chest.

6. **No event-level dedup** — The `Dispatcher` has no mechanism to prevent duplicate event delivery. All idempotency must be handled at the handler level.

### 4.3 Medium-Severity Risks

7. **Daily turn rollback inconsistency** — If `AwardXP()` fails after `ConsumeDailyTurn()` succeeds, the handler attempts a best-effort rollback (`UpdateDailyTurn` with `completed: false`), but this could also fail, leaving an inconsistent state (turn consumed but no XP awarded, no event published).

8. **Creative submission has no dedup** — `CreativeService.Submit()` has no idempotency key. Retried API calls create duplicate submissions and emit duplicate events.

9. **`RelicCollectedEvent` has empty `CrewID`** — In `chest/service.go:198`, `CrewID` is set to `""` when publishing `RelicCollectedEvent`. This may cause issues for consumers that expect a non-empty crew ID.

### 4.4 Low-Severity Observations

10. **`ChapterCompletedEvent` has empty `SeasonSlug`** — In `chapter/service.go:341`, `SeasonSlug` is hardcoded to `""` in `MarkComplete()`. The `handleQuestCompleted()` method does not propagate the season from the quest definition.

11. **`RewardEngine` uses time-based seed** — `defaultRandomSource` seeds with `time.Now().UnixNano()`. This is not deterministic across server restarts, which may cause reward inconsistency in replay scenarios.

12. **`OpenChest` has a TOCTOU race** — The `ch.Opened` check and the `UpdateChest` call are not atomic. Concurrent opens of the same chest could both pass the check before either updates the DB.

---

## 5. Remaining Blockers

| Blocker | Type | Impact |
|---|---|---|
| DailyTurnCompletedEvent never emitted | Critical | Daily streak tracking and daily-streak achievements are non-functional |
| LevelReachedEvent never emitted | Critical | Level-up achievements are non-functional |
| ChestCreatedEvent has no subscribers | Medium | No downstream reaction to chest creation |
| Quest XP double-counting on retry | High | Player progression can be inflated by retries |
| Duplicate chest creation on replay | High | Players can receive duplicate chests from event replay |
| No event dedup in Dispatcher | High | All handlers are vulnerable to duplicate processing |

---

## 6. Recommended Fixes

### 6.1 Critical Fixes (must fix before P1)

1. **Wire `DailyTurnAPIHandler` to dispatcher** — In `api/dev/main.go`, change line 149 from:
   ```go
   dailyTurnHandler := dailyturn.NewDailyTurnAPIHandler(dailyTurnSvc, progSvc)
   ```
   to:
   ```go
   dailyTurnHandler := dailyturn.NewDailyTurnAPIHandlerWithPublisher(dailyTurnSvc, progSvc, dispatcher)
   ```

2. **Wire `ProgressionService` to dispatcher** — In `api/dev/main.go`, change line 144 from:
   ```go
   progSvc := progression.NewProgressionService(repo.Users, &progCfg)
   ```
   to:
   ```go
   progSvc := progression.NewProgressionServiceWithPublisher(repo.Users, &progCfg, dispatcher)
   ```

3. **Add `EventTypeChestCreated` subscriber** — Either subscribe a handler (e.g., for analytics/observability) or remove the `ChestCreatedEvent` publication if it is not needed.

### 6.2 High-Priority Fixes

4. **Add idempotency to `QuestCompletedHandler.Handle()` (chest)** — Check if a chest already exists for the player+quest+source before creating a new one.

5. **Move XP award after quest completion check** — In `QuestAPIHandler.CompleteChallenge()`, restructure so that challenge XP is only awarded when the quest transitions to DONE, or add a dedup mechanism.

6. **Add event ID and dedup to `Dispatcher`** — Add an `EventID` field to the `Event` interface and track processed event IDs in the `Dispatcher` to prevent duplicate delivery.

### 6.3 Medium-Priority Fixes

7. **Fix `RelicCollectedEvent.CrewID`** — In `chest/service.go:198`, populate `CrewID` from the player's profile instead of using an empty string.

8. **Fix `ChapterCompletedEvent.SeasonSlug`** — In `chapter/service.go:341`, propagate the season slug from the quest definition or content gateway.

9. **Add dedup to `CreativeService.Submit()`** — Add an idempotency key (e.g., based on `QuestID + ChallengeID + AuthorUID + Kind`) to prevent duplicate submissions on retry.

10. **Fix daily turn rollback** — Make the XP award and turn consumption atomic, or defer the turn creation until after XP is successfully awarded.

---

## 7. Go Test Results

### gofmt
```
$ gofmt -l pkg/ api/
(no output — all files properly formatted)
```
**Result: PASS**

### go vet
```
$ go vet ./...
(no output — no issues)
```
**Result: PASS**

### go test
```
$ go test ./...
all packages: OK
```
**Result: PASS** — All 28 test packages pass. No failures.

---

## 8. Overall Readiness Assessment

**NOT PRODUCTION-READY for P1 feature work.**

Two critical blockers prevent the event architecture from being production-ready:

1. **`DailyTurnCompletedEvent` is never emitted** — The daily turn handler uses `NopPublisher`. Daily streak tracking and daily-streak achievements are completely non-functional.

2. **`LevelReachedEvent` is never emitted** — The progression service uses `NopPublisher`. Level-up achievements are completely non-functional.

Additionally, the system lacks event-level dedup, has no idempotency on chest creation, and has a quest XP double-counting vulnerability on retry. These issues mean that P1 features built on top of this event architecture would inherit these defects.

**Recommended action:** Fix the two critical wiring issues (items 1 and 2 in Recommended Fixes) before any P1 implementation begins. The remaining issues should be addressed in a follow-up pass before declaring the event system fully production-ready.
