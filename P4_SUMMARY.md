# P4 — Observability & Operations: Summary

Milestone: P4 — Observability & Operations  ·  Date: 2026-08-05  ·  Status: Complete

## What shipped

1. **Metrics inventory extended** (`pkg/observability/metrics.go`) — XP awarded,
   realm completions, chests created, rewards generated, duplicates prevented,
   lock conflicts, replay-ignored, validation failures, and an event-pipeline
   set (published/handled/errors/latency per type). All surfaced on `/metrics`
   behind the existing internal token.

2. **Event pipeline is observable** (`pkg/game/events/events.go`) —
   `Dispatcher.SetRecorder` + `Diagnostics()` now capture publish counts,
   per-handler durations, and previously-discarded handler errors. Wired to
   metrics in `api/dev/main.go`.

3. **Structured service-call logging** (`pkg/observability/logging.go`) —
   `Logger.LogServiceCall` emits `request_id`-correlated lines with
   `op`, `entity_id`, `duration_ms`, `outcome`, `retried`, `conflict`,
   `idempotency_skip`, `validation_failure`. Used by quest + daily-turn handlers.

4. **Game services instrumented** — `ProgressionService`, `ChestService`,
   `QuestAPIHandler`, `DailyTurnService`/`DailyTurnAPIHandler`,
   `ChapterService` gained optional nil-safe `SetMetrics`/`SetLogger`,
   preserving the existing DI setter pattern and constructor compatibility.

5. **Runtime operations completed** — balance reload & list wired into
   `/api/admin/balance` (was stubbed); `odyssey_balance_configs` added to the
   Supabase `allowedTables` whitelist (**bug fix** that blocked balance
   loading at boot); `/api/admin/validate` now runs the content validator
   (was stubbed) and records `validation_failures`.

6. **Public status endpoint** — `GET /api/status` returns app, version, schema,
   uptime, and live content generation/cache health (was `501 Not Implemented`).

## Bug fixed
`pkg/db/supabase.go`: `allowedTables` missing `odyssey_balance_configs` →
`supabaseBalanceStore` always failed `validateTable`, so balance overrides
could never load. Added the entry; balance now loads at boot and reloads at
runtime.

## Constraints preserved
- No architectural redesign; event pipeline stays synchronous/in-process.
- No new dependencies (stdlib only).
- Gatekeeper / Family Reward / Firestore untouched; no schema changes.
- No new tables; only the pre-provisioned `odyssey_balance_configs` (migration
  008) is now actually reachable.
- Optimistic locking + idempotency behavior unchanged; instrumentation only
  counts outcomes, doesn't alter them.

## Verification
- `gofmt -l .` clean; `go vet ./...` clean; `go build ./...` clean.
- `go test ./...` — **all packages pass**.
- `-race` not runnable here (no host `gcc`); concurrency tests written to be
  race-free and designed for CI `-race` runs. (See `P4_STRESS_VALIDATION.md`.)

## Files changed
Core: `pkg/observability/metrics.go`, `pkg/observability/middleware.go`,
`pkg/observability/logging.go`, `pkg/game/events/events.go`,
`pkg/game/progression/service.go`, `pkg/game/chest/service.go`,
`pkg/game/quest/handler.go`, `pkg/game/dailyturn/service.go`,
`pkg/game/dailyturn/handler.go`, `pkg/game/chapter/service.go`,
`pkg/game/balance/service.go`, `pkg/db/supabase.go`, `api/admin/index.go`,
`api/status/index.go`, `api/dev/main.go`.

Tests: the new `*_p4_test.go` / `*_test.go` files listed in `P4_TEST_SUMMARY.md`,
plus the corrected assertion in `pkg/game/progression/idempotency_test.go`.

## Deliverable artifacts
- `P4_OBSERVABILITY_GAP_REPORT.md`
- `P4_LOGGING_STANDARDIZATION.md`
- `P4_METRICS_INVENTORY.md`
- `P4_HEALTH_DIAGNOSTICS.md`
- `P4_RUNTIME_OPERATIONS.md`
- `P4_EVENT_PIPELINE.md`
- `P4_STRESS_VALIDATION.md`
- `P4_TEST_SUMMARY.md`
- `P4_SUMMARY.md` (this file)

## Out of scope (deferred)
- External metrics backend (Prometheus export / Grafana dashboards) — future phase.
- Balance **write** path (setting override values via API) — Phase 5 CMS.
- Offline-first / full e2e DB-backed stress — future phase.
- `-race` CI step is blocked only by the dev sandbox lacking `gcc`.
