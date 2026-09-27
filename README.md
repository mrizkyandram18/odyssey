# Odyssey

Private family daily-task → coin → redemption platform.

An admin publishes a sequenced list of tasks each day (videos, quizzes,
photo/document uploads, text answers, mini-games). Family members complete
them in a mobile-first React PWA. Submissions are graded automatically in
Postgres or queued for admin review. Approved work earns **coins + XP +
streaks + levels**. Coins are redeemed from a reward catalog (pulsa,
e-wallet, cash) through admin-approved claims — the actual money transfer
is manual, outside the app. A secondary loop grants daily reward tickets
redeemable for avatar cosmetics.

**No public signup. No payment gateway. No external identity provider.**
Members are created by an admin.

## Main user flow

```text
Login (username + password, device-bound)
  → Daily Tasks (today's sequenced list)
  → Submit (answers or proof upload)
  → Grade / Review (auto in Postgres, or admin review)
  → Earn Coins (+ XP, streak, level)
  → Redeem (claim from catalog inside payout policy)
  → Payout (admin approves; transfer done manually)
```

## Architecture

```text
React 19 PWA (web/)
   ↓ HTTPS JSON + HMAC session
Go 1.25 stdlib monolith, no framework (pkg/, internal/api/)
   ↓ PostgREST + RPC + Storage (service-role key)
Supabase Postgres (supabase/migrations/ — grading, ledger, streak, caps)
   ↓ best-effort only
Telegram bot (admin notify) + Web Push (VAPID)
```

| Layer | Technology | Entry |
|---|---|---|
| Frontend | React 19, Vite, TypeScript, Tailwind PWA | `web/src/main.tsx` |
| Backend | Go 1.25 stdlib (`net/http` mux) | `internal/api/dev/main.go` (local), `api/index.go` (Vercel) |
| Database | Supabase Postgres via PostgREST + RPC + Storage | `supabase/migrations/` |
| Auth | Local username/password (bcrypt) + device binding + HMAC sessions | `pkg/auth/` |

Business-rule ownership is split — this matters before you change anything:

- **Postgres RPCs** own grading, ledger appends, streaks, tickets, cap
  resolution (server-authoritative; the browser is never trusted).
- **Go** does fail-fast validation, family tenancy pre-checks, answer-key
  sanitizing, and upload handling.
- **React** never grades and never sees answer keys.

See `docs/architecture.md` §7 for the full ownership map.

## Repository structure

```text
internal/api/<domain>/  HTTP handlers: login, me, family_tasks, shop,
                         rewards, admin_tasks, admin_members,
                         payout_config, families, push, status, dev
pkg/                     Cross-domain packages: auth, db, tasks, payout,
                         push, shared, observability, server
api/index.go             Vercel serverless entry (lazy singleton handler)
web/src/                 PWA: app/ (routes), features/ (per-domain UI),
                         shared/ (API client web/src/shared/lib/api.ts,
                         types web/src/shared/types/index.ts)
supabase/migrations/     Canonical schema + RPC history (001–088)
scripts/                 Ops/seed/smoke/verify scripts
docs/                    Developer documentation (6 files, see below)
```

## Quick start

Prerequisites: Go 1.25, Node 20, a Supabase Postgres instance.

```bash
# 1. Environment
cp .env.example .env
# Required: SUPABASE_URL, SUPABASE_SERVICE_KEY, SESSION_SIGNING_SECRET,
# and any non-empty PARENT_ID (unused leftover, still boot-required).
# See docs/operations.md for the full variable list.

# 2. Backend (from repo root, default :8080)
go run ./internal/api/dev

# 3. Frontend (dev proxy /api → http://localhost:8080)
cd web && npm install && npm run dev
```

```bash
# Tests & checks
go test -count=1 ./...          # backend
go vet ./...                    # backend lint (gofmt must be clean)
cd web && npm run test          # frontend (Vitest)
npx tsc --noEmit                # frontend types
npx eslint .                    # frontend lint
```

Prototype accounts for local dev (reset by migration `049`): see your local
`.env` / seed data — do not commit real credentials.

## Documentation

Read in this order:

1. `README.md` (this file)
2. `CLAUDE.md` (stack, enforced rules, build/test commands)
3. `docs/README.md` (index, source-of-truth hierarchy, conventions)
4. `docs/architecture.md` (system explanation, flows, rule ownership)
5. `docs/database.md` (tables, RPCs, ledger, invariants)
6. `docs/api.md` (HTTP surface by domain, with behavior notes)
7. `docs/operations.md` (run, configure, deploy, troubleshoot)
8. `docs/decisions.md` (architecture-significant decisions with evidence)

Source-of-truth priority: source code → migrations/RPC → config examples →
tests → deployment/CI → docs. If code and docs disagree, code wins — fix
the docs.

## Deployment

- **Single container**: `Dockerfile` (Go 1.25 → Node 20 → Alpine, serves
  `web/dist`) composed with Postgres 16 via `docker-compose.yml`.
- **Vercel**: static frontend + Go function (`vercel.json` rewrites,
  SPA fallback).
- **CI** (`.github/workflows/ci.yml`, blocking): fmt + vet + tsc + eslint
  → `go test` + Vitest → both build artifacts.

Details: `docs/operations.md`.

## Important notes (read before modifying)

- `odyssey_coin_transactions` is **append-only** — a trigger rejects
  `UPDATE`/`DELETE`. Corrections are new compensating rows.
- One rewarded approval per user per task (`UNIQUE(task_id, user_uid)` +
  RPC `P0004` guard). Never bypass with upserts.
- Answer keys must never reach the browser — the sanitizer in
  `internal/api/family_tasks/api.go` must cover every new question-like field.
- New tables need **both** a migration **and** an `allowedTables` entry in
  `pkg/db/supabase.go` (15-entry allowlist).
- Member reads/writes are family-scoped in Go **and** re-checked in SQL —
  keep both.
- Coins leave only via admin-approved claims inside window/policy; the
  money transfer itself is manual.
- Earning cap `0` means **unlimited** — bonuses and ceilings must not apply.
- Config precedence: per-user row → `odyssey_system_config` key → compiled
  default. Keep the Go mirror and RPC defaults in sync when touching caps.
