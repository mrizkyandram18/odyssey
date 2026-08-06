# P4 Deliverable 1 — Observability Gap Report

Milestone: P4 — Observability & Operations
Date: 2026-08-05
Scope: Odyssey Go backend (`api/`, `pkg/`). Read-only audit of the existing
observability surface against the P4 requirements. No application code was
changed by this report; the follow-on parts implement the closing actions.

---

## 1. Existing infrastructure (baseline)

| Concern | Where | Status |
|---|---|---|
| Structured JSON logging | `pkg/observability/logging.go` (`Logger`) | Present, async channel-backed, sensitive-key redaction, `LogRequest` |
| Request correlation | `pkg/observability/requestid.go` + middleware | `X-Request-ID` set/preserved; `WithRequestID`/`RequestIDFromContext`; propagated into `odyssey_audit_logs.request_id` (migration 009) |
| Metrics | `pkg/observability/metrics.go` (`Metrics`) | Request counts/latency, login, cache, business events, admin ops, DB latency |
| Health endpoints | `pkg/observability/health.go` | `/health`, `/ready`, `/live` with 5s-cached checker (configuration, database, cache, content_generation, audit_store, admin_store) |
| Profiling | `pkg/observability/profiler.go` + `profiling_client.go` | `/debug/profile`, `/debug/profile/recommendations`, DB query profiling via `ProfilingClient` |
| Version | `pkg/observability/version.go` | `/version` with build info + content generation |
| Metrics exposure | `MetricsHandler` | `/metrics` behind `ODYSSEY_INTERNAL_METRICS_TOKEN` |
| Runtime content reload | `api/admin/index.go` + `pkg/content/service.go` | `POST /api/admin/reload`, `GET /api/admin/status`, `GET /api/admin/metrics` (live) |
| Balance overrides | `pkg/game/balance` + `pkg/db/balance.go` | Loaded at boot via `odyssey_balance_configs`; `Service.Reload` exists but is **not exposed** |
| Event pipeline | `pkg/game/events/events.go` (`Dispatcher`) | Sync publish to subscribed handlers; errors are **discarded**; no diagnostics |
| Content validation | `pkg/game/validation` (`Validator`) | Implemented + tested; **not wired** to any endpoint |

## 2. Gaps vs. P4 requirements

### 2.1 Structured logging standardization
- `Logger.LogRequest` covers request-level correlation, but **no service-level
  structured call log** exists for the business flows (quest completion, daily
  turn, chest opening, admin reload).
- **No `duration`, `retry`, `conflict`, `idempotency_skip`, or
  `validation_failure` fields** are emitted at the service layer.
- Log correlation is available (`RequestIDFromContext`) and already used by
  `pkg/game/audit`, so closing the gap is additive.

### 2.2 Metrics inventory
Present counters: request count/latency, login success/failure, cache hits/
misses, `quest_completed`, `chest_opened`, `creative_submitted`, admin ops,
DB latency/call/slow count.

Missing (required by P4):
- XP awarded (amount, not just quest count)
- Realm completions
- Chest creations
- Rewards generated
- Duplicates prevented (reward dedup)
- Optimistic-lock conflicts (version CAS misses)
- Replays ignored (idempotency guards, duplicate events)
- Validation failures
- Event pipeline: published / handled / handler errors / handler latency

### 2.3 Health / diagnostics
- `/health`, `/ready`, `/live`, `/metrics`, `/version`, `/debug/profile` exist.
- `GET /api/status` returns **501 Not Implemented** (`api/status/index.go`) —
  the one public status endpoint that is a stub.
- No diagnostics view of the **event pipeline** (subscription counts, per-type
  volumes, handler errors).

### 2.4 Runtime operations
- Content reload: **done** (`/api/admin/reload`, `ReloadResource` on CRUD).
- Balance reload: **not exposed**. `POST /api/admin/balance` and
  `GET /api/admin/balance` are stubs returning `"balance not configured"`.
- **Confirmed bug**: `pkg/db/supabase.go` `allowedTables` does **not** contain
  `odyssey_balance_configs`, so `supabaseBalanceStore.GetOverride`/`ListOverrides`
  fail `validateTable`. Balance overrides cannot load at boot or reload at
  runtime today.

### 2.5 Event pipeline diagnostics
- `Dispatcher.Publish` discards `Handler.Handle` errors; no durations, no
  per-type counters, no subscription introspection.
- Event-driven side effects (chapter completion, achievements, chest-on-quest,
  lore unlocks, level-up) are therefore **unobservable** in production.

### 2.6 Stress validation
- No concurrency/stress tests exist for the observability middleware,
  dispatcher, or content cache reload paths.

## 3. Required workflows → instrumentation map

| Workflow | Instrumentation point | Signals needed |
|---|---|---|
| Quest completion | `api/quests` → `QuestAPIHandler.CompleteChallenge` | duration, quest_completed, xp, level_up, lock conflict |
| XP award / level-up | `ProgressionService.AwardXP` | xp awarded, lock conflict (CAS miss), level_reached |
| Realm progression | `QuestAPIHandler.advanceRealm` | realm completed, lock conflict (CAS miss) |
| Chapter completion | `ChapterService.handleQuestCompleted` | replay ignored (already complete), chapter_completed |
| Achievement evaluation | dispatcher handlers (chapter/relic/daily turn/level/creative) | handler errors/durations |
| Daily turn | `DailyTurnService.ConsumeDailyTurn` + handler | replay ignored (same-slug), duration |
| Chest creation | `ChestService.CreateChest` | chests created |
| Chest opening | `ChestService.OpenChest` | rewards generated, duplicates prevented, lock conflict, duration |
| Reward generation | `RewardEngine.GenerateRewardsForChest` | rewards generated |
| Event publish/handle | `Dispatcher.Publish` | published / handled / errors / latency per type |
| Admin validate | `/api/admin/validate` (stub) | validation failures |
| Balance reload | `/api/admin/balance` (stub) | reload success/failure |

## 4. Actions (implemented in P4 parts 2–8)

1. **Metrics extension** — add XP/realm/chests/rewards/duplicates/locks/replay/
   validation/event counters to `Metrics` + snapshot (`pkg/observability/metrics.go`).
2. **Event recorder + diagnostics** — `Dispatcher.SetRecorder`, publish-time
   counters/durations, `Diagnostics()` view (`pkg/game/events/events.go`).
3. **Service logging** — `Logger.LogServiceCall` with request_id, op, entity,
   duration_ms, outcome, retried, conflict, idempotency_skip,
   validation_failure (`pkg/observability/logging.go`).
4. **Service instrumentation** — optional `SetMetrics`/`SetLogger` on
   `ProgressionService`, `ChestService`, `QuestAPIHandler`,
   `DailyTurnService`/`DailyTurnAPIHandler`, `ChapterService` (all additive,
   nil-safe, preserving the existing DI/setter pattern).
5. **Balance runtime ops** — add `odyssey_balance_configs` to `allowedTables`
   (bug fix), add `BalanceService` to admin, implement
   `GET /api/admin/balance` (list) and `POST /api/admin/balance` (reload).
6. **Validation diagnostics** — wire `Validator` to `/api/admin/validate` and
   record `validation_failures`.
7. **Public status** — implement `GET /api/status` (app, version, uptime,
   content generation, cache hit ratio) behind a `Setup(provider)`.
8. **Stress + tests** — concurrency stress for middleware + dispatcher; unit
   tests for every new counter, the recorder, `LogServiceCall`, balance reload,
   validate, and status.
9. **Main wiring** — `api/dev/main.go` connects metrics → dispatcher recorder,
   services, admin, and status.

## 5. Non-goals (unchanged from P4)
- No architectural redesign of the event pipeline (still synchronous, in-process).
- No new third-party dependencies; stdlib-only.
- No changes to Gatekeeper / Family Reward / Firestore.
- No schema changes required (metrics/logs are in-memory; audit `request_id`
  column already exists).
