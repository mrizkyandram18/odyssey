# Decisions

Architecture-significant decisions proven by the implementation. Format:
Context / Decision / Consequences / Evidence. Where the original rationale
is not in evidence: `Historical rationale: UNVERIFIED`. No small/detail
decisions recorded here.

## D1 — Local authentication with device binding and HMAC sessions

Context: no external IdP exists in the running system (no IdP dependency
in `go.mod`, no external adapter in `pkg/auth/`).
Decision: username lookup in `odyssey_user_profiles` + bcrypt verify
(`pkg/auth/local.go`, `pkg/db/profile.go`) + first-login `device_id`
binding with 403-on-mismatch until admin reset (RPC
`odyssey_bind_or_verify_device`, `054`) + HMAC-signed sessions (8 h user /
30 min setup) via Bearer + `X-User-Session` + cookie
(`pkg/auth/session.go`, `middleware.go`); roles normalized ADMIN vs MEMBER
(legacy `GUIDE`/`BUILDER` count as admin).
Consequences: zero external auth dependency and basic account-sharing
control; but `PARENT_ID` remains boot-required though unused, logout is
client-side only (no backend route), and clients must manage device ids.
Evidence: `pkg/auth/*`, `internal/api/login`, `internal/api/me`,
`web/src/features/login/LoginPage.tsx`, `web/src/shared/lib/device.ts`.
Historical rationale: UNVERIFIED.

## D2 — Supabase Postgres with RPC authority

Context: grading and money-like movements must not trust the browser.
Decision: PostgREST + `SECURITY DEFINER` RPCs (service-role only) own
grading, ledger appends, streaks, tickets, and cap resolution; Go validates
fast and pre-checks tenancy, RPCs re-validate everything
(`supabase/migrations/042–047`, `069–079`); a 15-table allowlist in
`pkg/db/supabase.go` gates all access; RLS is service-role-permissive with
authz in code.
Consequences: cheating-resistant commits and a small ops footprint; but
earning math is split Go↔SQL with subtly different fallbacks (Go `-1`
unconfigured vs old RPC fallback) that must be tested jointly, and the Go
submit path does a best-effort ticket-claim merge after approval.
Evidence: `pkg/db/supabase.go`, `internal/api/family_tasks/api.go`,
`supabase/migrations/042–048`, `061`, `069–079`. Historical rationale:
UNVERIFIED.

## D3 — Immutable coin ledger with exactly-once rewards

Context: coins convert to real-world rewards, so double-credit is the top
integrity risk (an `ON CONFLICT`-without-guard shape was hardened in
`044`/`047`).
Decision: `odyssey_coin_transactions` rejects `UPDATE`/`DELETE` by trigger
(`P0012`); balances carry `CHECK (coins >= 0)`; one rewarded approval per
user per task via `UNIQUE(task_id, user_uid)` plus the RPC approved-guard
(`P0004`); profile locked `FOR UPDATE` with earned sums re-read after the
lock; tickets idempotent by `(user_uid, ticket_date)` PK with
consume-under-lock.
Consequences: auditable, retry-safe earning (proven by 100× concurrent
adversarial tests); corrections require compensating rows, and resubmits
of approved work fail loudly by design.
Evidence: `042` (trigger, ledger), `043/044/047` (guards), `073` (current
reward paths), `078` (tickets), `pkg/adversarial/*_test.go`,
`internal/api/rewards/api_test.go`. Historical rationale: UNVERIFIED.

## D4 — Direct redemption with manual payout (no payment integration)

Context: no payment gateway or external reward system exists in code (grep
for gateway names hits only docs).
Decision: seeded coin catalog (`odyssey_reward_catalog`) + `odyssey_claims`
 (`PENDING → APPROVED | REJECTED`) gated by redemption window, balance,
 minimum withdrawal, per-user frequency (`THRESHOLD/WEEKLY/MONTHLY`), and
 one-pending-claim (lifetime max-payout cap removed in `095` — earning is
 capped upstream via `P0016`); debit at create (`CLAIM_REDEEM`), approve
stamps `processed_at` with no ledger row, reject writes `CLAIM_REFUND`;
transfer happens manually out-of-band; Telegram notify is best-effort
(`pkg/shared/telegram.go`, sole call site in `internal/api/shop/api.go`).
Consequences: end-to-end value flow without third-party integration, at
the cost of manual reconciliation risk and no in-app transfer trail
(tooling UNVERIFIED).
Evidence: `043` (catalog seed), `050`/`069` (config), `042` (process
semantics), `internal/api/shop`, `internal/api/payout_config`,
`pkg/payout/*`, `web/src/features/shop/RewardShopPage.tsx`. Historical
rationale: UNVERIFIED.

## D5 — Linear daily task path with family isolation

Context: a small closed group needs a short daily ritual, not a backlog
queue — and members must never see other families' tasks or answer keys.
Decision: today's family tasks served ordered `(step_order, id)` with step
N unlocking on step N−1 submission (`APPROVED` or `PENDING`; `REJECTED`
re-locks next); `USER`-scoped tasks filtered per UID; pre-join backlog
excluded; answer keys stripped by a recursive Go sanitizer before every
member response; family match re-checked in SQL (`P0003`).
Consequences: a phone-completable daily path with server-derived locks;
but lock state is computed per request (not stored), and every new
question-like config field must be added to the sanitizer's key list.
Evidence: `internal/api/family_tasks/api.go` (`HandleGetToday`,
`sanitizeValue`), `pkg/tasks/validator.go`
(`ResolveEvaluationTypeForConfig`), `web/src/features/stepper/LinearPath.tsx`.
Historical rationale: UNVERIFIED.

## D6 — Configuration as data with per-user → system → default precedence

Context: windows, rates, targets, caps, bonuses, and payout policy change
more often than code should.
Decision: runtime tunables live in `odyssey_system_config` (+ per-user
`odyssey_user_payout_config` and per-profile target/cap columns), resolved
per-user → system key → compiled default; since `079` the DB is the
source-of-truth with no hardcoded business config in Go/TS/SQL
(`P0026/P0027` fail closed when required keys are missing); `0` = unlimited
for the earning cap (bonus/ceiling skipped).
Consequences: ops can retune without deploys; but two readers (RPC + Go
mirror) must be kept in sync, and calendar-month caps vs 1–24 distribution
periods must not be conflated.
Evidence: `079_admin_configurable_progression.sql`, `071`
(calendar-month caps), `internal/api/family_tasks/api.go`
(`getEffectiveEarningCap`), `internal/api/admin_members/api.go`
(`resolveEarningCap`), `internal/api/payout_config/api.go`. Historical
rationale: UNVERIFIED.
