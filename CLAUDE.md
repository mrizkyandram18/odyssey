# Odyssey

Private family daily-task → coin → redemption platform.

## Stack

- Backend: Go 1.25 stdlib monolith, no framework (`pkg/`, `internal/api/`)
- Frontend: React 19, Vite, TypeScript, Tailwind PWA (`web/`)
- Database: Supabase Postgres via PostgREST + RPC + Storage
- Auth: local username/password (bcrypt) + device binding + HMAC sessions
- Integrations: Telegram (best-effort admin notify), Web Push VAPID
- Deploy: single container (compose) or Vercel; blocking CI (`ci.yml`)

## Rules (enforced by code/schema — do not violate)

- `odyssey_coin_transactions` is append-only (trigger rejects UPDATE/DELETE).
- One rewarded approval per user per task (`UNIQUE(task_id, user_uid)` + RPC `P0004` guard).
- Answer keys never reach the browser (response sanitizer in `family_tasks`).
- New tables need a migration AND an `allowedTables` entry in `pkg/db/supabase.go`.
- Member reads/writes are family-scoped in Go filters and re-checked in SQL.
- Coins leave only via admin-approved claims inside window/policy; money transfer is manual.
- Earning cap `0` = unlimited (bonuses/ceilings must not apply to it).
- Boot requires `SUPABASE_URL`, `SUPABASE_SERVICE_KEY`, `PARENT_ID` (unused leftover, still enforced), `SESSION_SIGNING_SECRET`.
- Config precedence: per-user → `odyssey_system_config` key → compiled default.

## Structure

- `internal/api/<domain>/` - HTTP handlers: `login`, `me`, `family_tasks`, `shop`, `rewards`, `admin_tasks`, `admin_members`, `payout_config`, `families`, `push`, `status`, `dev` (binary entry)
- `api/index.go` - Vercel serverless entry
- `pkg/auth|db|tasks|payout|push|shared|observability|server` - cross-domain packages
- `web/src/{app,features,shared}` - PWA; API client `shared/lib/api.ts`, types `shared/types`
- `supabase/migrations/` - canonical schema/RPC history
- `scripts/` - ops, smoke, seed, verify scripts
- `docs/` - current documentation (6 files, see below)

## Build & Run

- **Backend:** `go run ./internal/api/dev`
- **Frontend:** `cd web && npm install && npm run dev`
- **Build:** `go build ./...`, `cd web && npm run build`

## Test & Lint

- **Backend:** `go test -count=1 ./...`; `go vet ./...`; gofmt must be clean
- **Frontend:** `cd web && npm run test` (Vitest); `npx tsc --noEmit`; `npx eslint .`

## Documentation

Source of truth (in order): source code → migrations/RPC → config examples → tests → deployment/CI. These docs describe; the code decides.

| File | Covers |
|---|---|
| `docs/README.md` | Overview, map, reading order |
| `docs/architecture.md` | Goals, constraints, building blocks, flows, deployment, risks, glossary |
| `docs/database.md` | Tables, RPCs, ledger, invariants |
| `docs/api.md` | HTTP surface by capability |
| `docs/operations.md` | Dev, config, deploy, CI, troubleshooting |
| `docs/decisions.md` | Architecture-significant decisions with evidence |

If code and docs disagree, implementation/DB constraints win — fix the docs.

## Agent Rules

- Verify claims against code/schema/tests; never invent behavior, rationale, or SLOs.
- Mark the unverifiable as `UNVERIFIED`/`UNDEFINED` with the evidence still needed.
- Keep changes minimal and scoped; docs-only tasks must not touch code, migrations, or config behavior.
- No commit/push/PR unless explicitly requested.
