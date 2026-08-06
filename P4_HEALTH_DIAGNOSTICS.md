# P4 Deliverable 4 — Health & Diagnostics

Milestone: P4 — Observability & Operations

## Health probes (pre-existing, unchanged)
`pkg/observability/health.go` — `HealthChecker` with 5s cache TTL.
`api/dev/main.go` registers checks: `configuration`, `database`
(`odyssey_system_config`), `cache`, `content_generation`, `audit_store`,
`admin_store`.

Exposed:
- `/health` — `HealthHandler` (aggregates all checks).
- `/ready` — `ReadyHandler` (cache + content + database).
- `/live` — `LiveHandler` (process liveness).

These were already wired; P4 keeps them and adds the diagnostics layer below.

## Public status endpoint (new — was 501)
`api/status/index.go` previously returned `501 Not Implemented`. Implemented as
a lightweight, unauthenticated endpoint returning app identity + uptime + live
content pointer:

```
GET /api/status  -> 200
{
  "app":"odyssey",
  "version":"dev",
  "schema_version":"9",
  "uptime_seconds":42,
  "boot_time":"2026-08-05T...",
  "content_generation":42,
  "cache_hit_ratio":0.66,
  "content_status": {...},
  "timestamp":"2026-08-05T..."
}
```
Provided via `status.Setup(status.FuncStatusProvider(...))` in `api/dev/main.go`,
composed from `observability.Version`/`SchemaVersion`/`Metrics.BootTime()` and
`content.ContentService.Status`/`CacheGeneration`/`CacheHitRatio`.

## Event pipeline diagnostics (new)
`pkg/game/events/events.go` `Dispatcher.Diagnostics()` returns per-event-type:
`published`, `handled`, `errors`, `handler_count`, `avg_handler_duration_ms`.
Wired to `Metrics` via the optional `Recorder` interface
(`Dispatcher.SetRecorder(metrics)`). The dispatcher now records publish-side,
per-handler duration, and per-handler errors (previously discarded) without
changing the synchronous dispatch contract.

## Admin runtime diagnostics (new/expanded)
`api/admin/index.go`:
- `GET /api/admin/status` — existing live content + cache stats (generation,
  hit ratio, CacheStats).
- `GET /api/admin/metrics` — existing cache counters.
- `GET /api/admin/validate` — **was stubbed "not configured"**; now runs
  `validation.Validator` over the live `ContentSet` and returns the result +
  audit-log entry + `validation_failures` metric.
- `POST /api/admin/reload` — existing full content reload.
- `GET /api/admin/balance` — **was stubbed**; now lists live overrides + load
  time (see balance runtime ops).
- `POST /api/admin/balance` — **was stubbed**; now reloads overrides from
  `odyssey_balance_configs`.

## Test coverage
- `api/status/index_test.go`: 200 payload fields, 405, no-provider fallback.
- `api/admin/balance_test.go`: `BalanceOverrides`, `ReloadBalance`,
  not-configured error, `Validate` (valid content → 0 validation failures;
  invalid content → failure counted).
