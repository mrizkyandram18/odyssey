# P4 Deliverable 3 — Lightweight Metrics Inventory

Milestone: P4 — Observability & Operations
Scope: `pkg/observability/metrics.go`. In-memory counters + `/metrics` (HTTP). No external
metrics backend; Prometheus scrape can be added later by a sidecar (future phase).

## Inventory

### Pre-existing (kept, backward compatible)
`request_count`, `request_latency_ms`, `login_success`, `login_failure`,
`cache_hits`, `cache_misses`, `quest_completed`, `chest_opened`,
`creative_submitted`, `admin_operations`, `db_calls`, `db_avg_latency_ms`,
`db_slow_queries`, `boot_time`, `uptime_seconds`.

The HTTP-path business events (`quest_completed`, `chest_opened`,
`creative_submitted`) remain as a fast path for HTTP-triggered actions and keep
their existing tests green.

### New (service-accurate)
| Metric (JSON) | Signal | Where recorded |
|---|---|---|
| `xp_awarded` | total XP granted | `ProgressionService.AwardXP` (only on committed CAS) |
| `realm_completed` | realm reached completion threshold | `QuestAPIHandler.advanceRealm` |
| `chests_created` | chest instances created | `ChestService.CreateChest` |
| `rewards_generated` | total relic rewards produced | `ChestService.OpenChest` |
| `duplicates_prevented` | duplicate relics merged | `ChestService.OpenChest` |
| `lock_conflicts` | optimistic-lock CAS misses | `ProgressionService.AwardXP`, `QuestAPIHandler.advanceRealm`, `ChestService.OpenChest` |
| `replay_ignored` | idempotency guards / re-play | `DailyTurnService.ConsumeDailyTurn`, `ChestService` (already-opened + quest-chest dedup), `ChapterService.handleQuestCompleted` |
| `validation_failures` | invalid content sets | `AdminService.Validate` |
| `events_published` / `event_type_*` | events entering pipeline | `Dispatcher.Publish` (+ per-type via `EventRecorder`) |
| `events_handled` / `event_type_*` | handler invocations | `Dispatcher.Publish` |
| `events_handler_errors` | handler errors | `Dispatcher.Publish` |
| `events_handler_avg_latency_ms` | avg handler duration | `Dispatcher.Publish` |
| `event_types` | per-type {published,handled,errors} map | `Dispatcher.Diagnostics`/`Snapshot` |

## Access
- `/metrics` — `MetricsHandler`, gated by `ODYSSESSE_INTERNAL_METRICS_TOKEN`
  (`InternalTokenMiddleware`). Returns `MetricsSnapshot` JSON.
- `Metrics.BootTime()` — exposed for uptime on `/api/status`.

## Instrumentation pattern
Services receive `*observability.Metrics` via optional, nil-safe setters
(`SetMetrics`). This mirrors the existing `SetBalance` / `SetContentGateway`
setter pattern and preserves DI/backward compatibility — constructors are
unchanged.

## Test coverage
- `pkg/observability/metrics_p4_test.go`: every new counter + `RecordEventPublished`/
  `RecordEventHandler` (incl. error path + per-type map), `SnapshotJSON`, `BootTime`.
- `pkg/observability/stress_test.go::TestObservability_Wrap_Concurrent`: 50×40=2000
  concurrent requests through `Wrap`, asserts counters are consistent and
  goroutine-safe (run under `-race` in CI; local env lacks `gcc`).
