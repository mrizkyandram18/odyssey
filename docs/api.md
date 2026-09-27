# API

HTTP surface for someone implementing or debugging frontend features.
Source: route table `pkg/server/server.go` × `internal/api/*` handlers ×
client `web/src/shared/lib/api.ts`.

- **Base**: same-origin (`VITE_API_BASE_URL` overrides for remote API; dev
  Vite proxies `/api → http://localhost:8080`).
- **Auth** on all `/api/*` except login/CSRF/health: `Authorization: Bearer
  <session>` **and** `X-User-Session: <session>` (client sends both) plus
  `odyssey_session` cookie, `credentials: include`. `/api/admin/*` also
  need admin role and pass the admin rate limiter (5 login / 30 admin /
  100 general per 60 s, per instance).
- **Errors**: `{error: string}` with `401` credential, `403`
  device-bound/disabled/blocked/role, `400` validation, `404` unknown,
  `429` rate-limited. The client throws `Error(err.error)`; pages render
  inline with retry. Earning-cap failures match
  `earning_cap_reached|p0016|batas earning` (`web/src/shared/lib/earning.ts`).
- **CSRF**: `GET /api/csrf` issues a token consumed only by push calls.

## Authentication

`POST /api/login` (public). Purpose: username/password login with device
binding. Request: `{username, credential, device: {device_id,
login_method: "BOTH"}}` (`login_method` is a hardcoded label, not a mode).
Response: `{session, uid, family_id, role, expires}` + HttpOnly cookie.
Errors: `401` wrong credential (unknown users also 401 — no enumeration);
`403` device-bound-elsewhere ("…perangkat lain…", needs admin device
reset) or inactive account; `200 setup_needed` / `password_required`
variants when no credential is set. No backend logout — the client clears
`localStorage odyssey_session`.

`GET / PATCH /api/me` (user). Purpose: session bootstrap (`GET` returns own
profile incl. `must_change_password`) and avatar self-service (`PATCH
/api/me/avatar {avatar_style, avatar_seed}`).

`POST /api/me/change-password` (user). Purpose: password change/rotation.
Rules: new password min 6 chars; `confirm` must match when present;
`current_password` required unless `must_change_password=true`; new must
differ from current. Wrong current → `401`.

## Tasks (member)

`GET /api/tasks/today` (user). Purpose: today's playable list — the core
member read. Behavior: system-timezone date → family's active tasks for
that date ordered `(step_order, id)` → `USER`-scope + join-date backlog
filtering → answer keys stripped → linear locks evaluated (step N
unlocked only if N−1 is `APPROVED`/`PENDING`). Response:
`{tasks[], earning_locked, earning_cap, earned}` (`earning_locked` also set
per coin-bearing task). Empty list debugging: check `is_active`,
`active_date` vs **system** timezone, `family_id`, `USER`-scope UID,
join-date guard, step-1 submission state. PWA refetches on
focus/online/visible.

`GET /api/tasks/{id}` (user). Purpose: single task view (same sanitizing).

`POST /api/tasks/{id}/submit` (user). Purpose: answer or proof submission.
Request: `{answers?, payload?, submission_type?}` (modals send
`answers` for quizzes/games, `payload` for uploads/text/video).
Behavior: blocked-user 403 → cap-halted 403 (`EARNING_CAP_REACHED`) →
family/scope 403 → already-`APPROVED` 400 → branch: `AUTO` calls
`odyssey_submit_auto_task` (returns approval + coins/XP, merges
`ticket_granted`), `ADMIN_REVIEW` calls `odyssey_submit_manual_task`
(returns `PENDING`). Task statuses surfaced: `LOCKED / UNLOCKED / PENDING /
APPROVED / REJECTED`.

`POST /api/tasks/upload` (user, multipart). Purpose: proof-file intake
before a manual submit. Limits: 10 MiB per file/request, executable
extensions and `text/html|javascript|application/x-sh` rejected,
content-type sniffed. Stored at
`{family}/{uid}/{unix}_{rand}_{clean}` in the `task-proofs` bucket.
Response: `{file_url, file_name, file_size, storage_path}`.

## Tickets & cosmetics (member)

`GET /api/rewards` → `{tickets}` (count of `GRANTED`).

`POST /api/rewards/claim` → `{granted, tickets}`. Daily ticket; needs ≥1
**today's** `APPROVED` submission (system timezone); idempotent — claiming
twice returns `granted: false`. Failure when no approved work (`P0023`).

`POST /api/rewards/open` →
`{cosmetic_id, slot, asset, tier, is_new}`. Consumes oldest ticket under
lock (none → `P0024`), random active cosmetic (none → `P0025`), tier bump
max 3. Client cannot choose the cosmetic.

`GET /api/rewards/collection` → `{items[]}` (catalog + owned tier +
equipped state).

`POST /api/rewards/equip {cosmetic_id}` — requires ownership (403
otherwise) and `frame|effect` slot; writes profile
(`avatar_frame`/`equipped_explorer_effect`).

## Shop / redemption (member)

`GET /api/shop/config` → effective redemption policy: window days, open
status, conversion rate, minimum withdrawal, per-user frequency
(`THRESHOLD/WEEKLY/MONTHLY`), weekday/month windows, eligibility. Drives
the shop open/closed indicator and CTA enablement.

`GET /api/shop/items` → available catalog ordered by cost (currently
fetched but not rendered — the shop UI is claim-form based, not
catalog-card based).

`POST /api/shop/redeem {coins, target_type, target_value}` (reward_id
optional) → `{success, claim_id, new_balance}`. Validations in order:
active account → schedule window (non-`THRESHOLD` only) → `coins > 0` →
target present → `target_type ∈ EWALLET|BANK|CASH|GOPAY|DANA|OVO|SHOPEEPAY|
TRANSFER_BANK` → ≥ minimum withdrawal → balance covers → atomic
`odyssey_create_claim` (re-checks all + one-pending `P0006`). Triggers a
best-effort Telegram admin message; the claim commits regardless.

`GET /api/shop/claims` → own claims (`PENDING/APPROVED/REJECTED` with
target, coins, timestamps).

## Family & push

`GET /api/families` → own family. `PATCH /api/families`
`{banner_url?, theme?}` — only those two fields are writable.
`GET /api/families/members` → family roster.

`POST /api/push {endpoint, keys: {p256dh, auth}}` → 201 subscribe (needs
`VITE_VAPID_PUBLIC_KEY` client-side; endpoint must be https, length-capped;
upsert). `DELETE /api/push?endpoint=` → unsubscribe (empty endpoint clears
all user subs). 503 when push store is unconfigured.

## Admin

All under `/api/admin/*`, admin role required (normalized `ADMIN`).

**Members** — `GET /api/admin/members` (paginated, excludes soft-deleted;
each row enriched with month earned, earning status, inactivity, effective
payout) · `POST /api/admin/members` (username `^[a-zA-Z0-9._-]+$` ≥3,
password ≥6 bcrypt, role, targets/caps 0–10000 with `0` = unlimited,
payout fields; creates `level 1, must_change_password=true`) ·
`PATCH /api/admin/members/{uid}` (name/role/active/targets/password/reset_device —
device reset clears binding via `odyssey_admin_reset_device`) ·
`POST …/{uid}/reset-password` (returns `{temporary_password}`, 14-char) ·
`POST …/{uid}/block {reason?}` / `…/unblock` (refuse admin roles; delete is
soft-delete preserving history and freeing the username).

**Tasks** — `GET /api/admin/tasks?date=` (defaults today, family-scoped) ·
`POST /api/admin/tasks` (validates type/config/rewards; `step_order`
collision → 400) · `PATCH / DELETE /api/admin/tasks/{id}` (`task_type`
immutable once submissions exist; eval re-synced on config/type touch) ·
`POST …/{id}/duplicate` (copies as `"(Salinan)"` at `max step + 1`) ·
`PATCH /api/admin/tasks/reorder {taskIds[]}` (must cover all same-date
tasks exactly; two-phase renumber).

**Verification** — `GET /api/admin/submissions` (`?status=&page=&limit=`,
alias `…/pending` defaults `PENDING`, exact-count pagination) · `POST
/api/admin/submissions/{id}/verify {APPROVED|REJECTED, notes?,
penalty_coins?}` (penalty ≥ 0, never on approve; RPC commits ledger) ·
`PATCH …/{id} {payload, notes?}` (bounded edit).

**Claims** — `GET /api/admin/claims` (family-scoped + catalog titles) ·
`POST /api/admin/claims/{id}/process {status, notes?}` (approve stamps
`processed_at`, no ledger row; reject writes `CLAIM_REFUND`).

**Config** — `GET/POST /api/admin/config` (economy: windows, rate, targets,
caps, ceiling, `level_cap_bonus` JSON, timezone, auto-block days,
announcements) · `/api/admin/cosmetics[/{id}]` (catalog CRUD + active
toggle) · `/api/admin/payout-config[/user|/system]` (per-user frequency/
minimum/windows upsert, or the 5 global keys).

## Status / observability

Authed `GET /api/status` (version, schema version, uptime). Public
`GET /health` (gates: config + `odyssey_user_profiles`/`odyssey_tasks`/
`odyssey_coin_transactions` probes; 503 `degraded`, 5 s cache), `/ready`
(same gates), `/live` (always alive), `/version`. Token-gated `/metrics`
and `/debug/profile*` (`ODYSSEY_INTERNAL_METRICS_TOKEN`).
