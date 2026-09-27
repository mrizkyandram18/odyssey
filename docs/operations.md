# Operations

Run, configure, deploy, and fix Odyssey. Every command below was verified
against `go.mod`, `web/package.json`, `.env.example`,
`production.env.example`, `Dockerfile`, `docker-compose.yml`,
`vercel.json`, `.github/workflows/`, and `internal/api/dev/main.go`.

## Prerequisites

- Go 1.25 (`go.mod`: `go 1.25.0`, module `odyssey`)
- Node 20 (Dockerfile frontend stage: `node:20-alpine`)
- A Supabase Postgres instance (local `http://localhost:54321` default in
  `.env.example`, or hosted)
- No other services required (no broker, cache, or cron)

## Environment setup

```bash
cp .env.example .env
```

| Variable | Required | Default / notes |
|---|---|---|
| `SUPABASE_URL` | **yes** (boot fails without) | e.g. `http://localhost:54321` |
| `SUPABASE_SERVICE_KEY` | **yes** | service-role key, server-only |
| `SESSION_SIGNING_SECRET` | **yes** | HMAC secret; `ADMIN_SECRET` is its fallback — set both |
| `ADMIN_SECRET` | effectively yes | fallback for session secret |
| `PARENT_ID` | **yes, any non-empty** | unused leftover, still enforced (`dummy_parent` in CI) — do not remove without a config change |
| `ODYSSEY_TIMEZONE` | no | `Asia/Jakarta`; DB key `timezone` wins when set |
| `ODYSSEY_ALLOWED_ORIGINS` | no | empty = allow-all — **set in prod** |
| `ODYSSEY_MAX_BODY_BYTES` | no | `1048576` (1 MiB; uploads fixed 10 MiB) |
| `ODYSSEY_RATE_LIMIT_WINDOW_SEC` / `_MAX_HITS` / `_LOGIN_…_MAX` / `_ADMIN_…_MAX` | no | 60 s; 100 general / 5 login / 30 admin |
| `ODYSSEY_INTERNAL_METRICS_TOKEN` | no | gates `/metrics`, `/debug/profile*` |
| `TELEGRAM_BOT_TOKEN` / `TELEGRAM_CHAT_ID` | no | best-effort admin notify; approvals commit regardless |
| `VAPID_PUBLIC_KEY` / `VAPID_PRIVATE_KEY` / `VAPID_SUBJECT` | no | sender fails closed without the pair; frontend needs `VITE_VAPID_PUBLIC_KEY` |
| `PORT` | no | `8080` (`production.env.example`) |

Boot prints every missing required variable — fix env, not code. Frontend:
only `VITE_API_BASE_URL` (remote API override; default same-origin) and
`VITE_VAPID_PUBLIC_KEY` matter; dev Vite proxies `/api → localhost:8080`.

## Local development

```bash
# Backend (repo root; loads .env; serves API + web/dist on :8080)
go run ./internal/api/dev

# Frontend (new terminal)
cd web && npm install && npm run dev
```

Prototype accounts for local dev are reset by migration `049`
(`admin/admin123`, `user_testing/odyssey123` pattern) — use your local
seed data; never commit real credentials.

## Tests, lint, build

```bash
go test -count=1 ./...        # backend (52 *_test.go incl. adversarial concurrency)
go vet ./...                  # backend lint; gofmt must be clean
cd web && npm run test        # frontend (Vitest, 32 suites)
npx tsc --noEmit              # frontend types (web/; build runs tsc -b && vite build)
npx eslint .                  # frontend lint (web/)
```

```bash
go build ./...                # backend build
cd web && npm run build       # frontend build → web/dist (served by Go/Vercel)
```

What the tests assert (evidence, not aspiration): anti-double-claim
(`P0004`, incl. 100× concurrent adversarial), ledger immutability
(`P0012`), answer-key sanitizer (forbidden key list), cap semantics (`0`
= unlimited, `P0016` halt), ticket idempotency + consume-under-lock,
device-binding 403, family isolation, exact-count pagination.
`scripts/verify-*.cjs` hold per-date regression fingerprints.

## Database workflow

- Canonical history: `supabase/migrations/` (`001–088`, `016` missing).
  Business core `042–088`; daily content seeds `080–088`.
- Apply with the `scripts/apply-migration-*.cjs` pattern (exact-ID guarded
  inserts + `odyssey_schema_version` bump); **never hand-delete rows** —
  the immutable ledger blocks deletes and content scripts guard on zero
  submissions.
- Confirm with `odyssey_schema_version` + authed `GET /api/status`.
- New tables: migration **and** `allowedTables` entry in
  `pkg/db/supabase.go`, else the gateway rejects them.
- New tunables: `odyssey_system_config` keys (DB-as-SOT since `079`), not
  literals in Go/TS/SQL. Fresh full-replay order is `UNVERIFIED`.

Seed/smoke scripts: `scripts/seed-today-tasks.cjs` (REST insert for
today), `scripts/apply-seed-052.cjs`, `scripts/smoke-prod-sep*.cjs`
(disposable-user live flows with cleanup), `scripts/run-production-smoke.cjs`.

## Deployment & CI

- **Single container**: `Dockerfile` (Go 1.25 → Node 20 → Alpine final,
  `EXPOSE 8080`, `HEALTHCHECK /health`) + Postgres 16 via
  `docker-compose.yml` (`backend` with `env_file ./production.env`;
  `postgres` with `./scripts/migrations:/docker-entrypoint-initdb.d`).
- **Vercel**: `vercel.json` (install/build in `web/`, output `web/dist`;
  `/api/*`, `/metrics`, `/version`, `/health`, `/ready`, `/live`,
  `/debug/*` → Go function; everything else → `/index.html` SPA fallback;
  immutable `/assets` cache).
- **Blocking CI** (`.github/workflows/ci.yml`, push `main/develop`, PR
  `main`): lint job (gofmt check + `go vet` + `tsc --noEmit` + `eslint`) →
  unit job (`go test -count=1 ./...` + Vitest) → build job (both
  artifacts). `ci-heavy.yml` is manual-only (race, Playwright e2e,
  docker build).

## Health & telemetry

- `/health` — 200 or 503 `degraded` (5 s cache); gates on config +
  `odyssey_user_profiles` / `odyssey_tasks` / `odyssey_coin_transactions`
  probes. `/ready` same gates; `/live` always alive; `/version` build info.
- Token-gated `/metrics`, `/debug/profile*`; authed `/api/status`
  (version, schema version, uptime).
- Structured logs + request IDs + in-process metrics; 30 s graceful
  shutdown; server timeouts 10/30/60/120 s.

## Troubleshooting

- **Won't boot**: read the missing-variable list in the log; check
  `SUPABASE_URL`, `SUPABASE_SERVICE_KEY`, `SESSION_SIGNING_SECRET`,
  non-empty `PARENT_ID`.
- **401**: credential wrong (unknown users also 401 by design).
- **403 + device text** ("perangkat lain"): binding mismatch — admin
  resets via member edit (`reset_device`).
- **403 + inactive text**: unblock in Members admin tab.
- **Empty task list**: check `is_active`, `active_date` vs **system**
  timezone (not browser tz), `family_id` match, `USER`-scope UID match,
  join-date backlog guard, step-1 submission unlocking step-2.
- **Re-submit rejected (`P0004` already completed)**: expected exactly-once
  behavior — use admin verify/edit, never delete rows.
- **Earning locked**: compare calendar-month `TASK_REWARD` sum vs effective
  cap (per-user → system default); `0` = unlimited; boundaries follow
  system timezone. Note the banner still says period "1–24" in one string
  (`LinearPath`) while caps are calendar-month — trust the backend values.
- **Claim blocked**: check window days, balance, minimum withdrawal, payout
  frequency (per-user row beats system keys), one-pending-claim rule.
- **Missing Telegram message**: check `TELEGRAM_*`; approvals commit
  regardless (best-effort by design).
- **No push**: check VAPID keys, subscription row, browser permission;
  OS-level delivery is `UNVERIFIED` — confirm on hardware before promising.
  Note the sender has no production call sites (subscribe storage only).
- **Upload auth oddity**: `uploadTaskProof` reads stale key
  `odyssey_session_token` while sessions live under `odyssey_session` —
  confirm headers in DevTools if uploads 401.
- **Fresh-Compose replay**: `UNVERIFIED` — use apply-scripts, then verify
  schema version + `/api/status`.

## Manual payout flow (no gateway in code)

```text
Admin → Claims queue: confirm destination (masked number + provider)
  → Approve (ledger already debited at create; processed_at stamped)
  → Execute the transfer OUT-OF-BAND (bank/e-wallet app)
  → Keep the receipt reference per team policy (tooling UNVERIFIED)
Reject instead → CLAIM_REFUND row restores balance automatically.
```

There is no `midtrans|xendit|stripe|gateway` code — only docs mention those
names to say they don't exist. Telegram notify is best-effort.
