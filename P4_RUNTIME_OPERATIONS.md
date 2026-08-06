# P4 Deliverable 5 — Runtime Operations

Milestone: P4 — Observability & Operations

## Content (already present)
`pkg/content/service.go` exposes a generation-aware TTL cache with
`CacheStats`, `CacheGeneration`, `CacheHitRatio`, `Reload`,
`ReloadResource`, `Status`, `Invalidate`. `pkg/observability` already syncs
cache stats into metrics every 5s (`syncCacheStats` in `api/dev/main.go`).

Admin endpoints:
- `POST /api/admin/reload` — full reload (used on publish/save/delete drafts).
- `GET /api/admin/status` — live cache + generation snapshot.

## Balance overrides (new — was stubbed)
`pkg/game/balance/service.go` already supported runtime reload
(`Service.Reload/Load`). It was **not exposed** via admin, and a bug blocked it
altogether:

> **Bug fixed**: `pkg/db/supabase.go` `allowedTables` did not contain
> `odyssey_balance_configs`, so `supabaseBalanceStore.GetOverride`/`ListOverrides`
> (which call `client.Get("odyssey_balance_configs", ...)`) failed
> `validateTable`. Balance overrides could never load at boot or reload at
> runtime. Added `odyssey_balance_configs: true`.

Wiring (`api/admin/index.go`):
- `AdminService.SetBalance(BalanceService)` + `SetMetrics`.
- `GET /api/admin/balance` → lists current overrides (`balance.Overrides()`),
  override count, and `loaded_at`.
- `POST /api/admin/balance` → calls `balance.Reload(ctx)` (re-reads
  `odyssey_balance_configs`), audit-logged under `OpReload`.
- `POST /api/admin/balance` writes a DB row? — **Out of scope**: the
  `balance.Store` is read-only (no write path). Setting individual override
  values is a Phase-5 CMS concern (values edited directly in the DB table).
  P4 delivers reload-at-runtime, which is the live-ops requirement.

`api/dev/main.go` now calls `adminSvc.SetBalance(balanceSvc).SetMetrics(metrics)`.

## Test coverage
`api/admin/balance_test.go` (service-level, no auth required):
- `TestAdmin_BalanceOverrides`
- `TestAdmin_BalanceReload`
- `TestAdmin_BalanceNotConfigured` (503 path)

## Audit trail
Admin mutations (create/save-draft/publish/delete/restore/reload-balance/
validate) are already written to `odyssey_audit_logs` (migration 008) with
`request_id` correlation (migration 009). P4 reuses this for balance reload
and validate.
