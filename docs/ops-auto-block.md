# Ops: Auto-block Inactive Users

## Schedule
- **00:05 Asia/Jakarta daily** → `05 17 * * *` UTC previous day
- GitHub Actions workflow: `.github/workflows/auto-block-inactive-users.yml`
- Concurrency group: `auto-block-inactive-users` (`cancel-in-progress: false`)
- Timeout: `5 minutes` job, `30s` curl `--max-time`

## Purpose
Cycle-aware automatic blocking of inactive `MEMBER`/`SEEKER` users based on approved task completion within the **current earning cycle** (`odyssey_target_period_bounds()`).

## Configuration
- Key: `auto_block_inactivity_days` (also `AUTO_BLOCK_INACTIVITY_DAYS` for compat)
- Default: `5`
- Allowed: `0..365`
- Semantics: `0 = disabled`, `1..365` = threshold calendar days
- Timezone: `Asia/Jakarta` (from `odyssey_system_config.timezone` or `ODYSSEY_TIMEZONE`)
- Invalid values fallback to `5` (safe default); `>365` fallback to `5`

## Inactivity Definition (Cycle-Aware)
- `inactive when (today_date - last_success_date) >= threshold` calendar days in `Asia/Jakarta`
- `last_success_date` = `MAX((COALESCE(reviewed_at,created_at) AT TIME ZONE v_tz)::date)` where `status='APPROVED'` **AND** date within `[period_start, period_end)` of current cycle
- Previous-cycle completions do **NOT** count (counter resets per cycle)
- Never-completed in current cycle → **NOT blocked**
- Already blocked (`is_active=false` or `blocked_at IS NOT NULL`) → skipped
- `ADMIN`/`GUIDE`/`BUILDER` → never blocked

## Idempotency / Retry Safety
- RPC `odyssey_auto_block_inactive_users()` is idempotent, atomic `UPDATE ... WHERE is_active=true ...`
- Second run same day: `blocked_count=0`, no duplicate side-effects, no ledger mutation
- Workflow uses `--retry 2` but repeated execution remains safe; no application duplicate state

## Authentication
- Workflow uses machine-to-machine token, not admin JWT
- Header: `X-Auto-Block-Token: ${ODYSSEY_AUTO_BLOCK_TOKEN}` (fallback `X-Internal-Token: ${ODYSSEY_INTERNAL_METRICS_TOKEN}`)
- Server validates via `subtle.ConstantTimeCompare` against `LoadConfig().AutoBlockToken` or `InternalMetricsToken`
- No hardcoded secrets in repo; tokens come from GitHub Actions secrets
- Secrets required:
  - `ODYSSEY_PRODUCTION_URL` — e.g. `https://odyssey.vercel.app` (no trailing slash needed)
  - `ODYSSEY_AUTO_BLOCK_TOKEN` — preferred; or `ODYSSEY_INTERNAL_METRICS_TOKEN` as fallback (reuse existing metrics token)
- Endpoint remains protected: unauthenticated or invalid token → falls through to `RequireAdmin` → `401/403`; no public unauthenticated access
- `ODYSSEY_AUTO_BLOCK_TOKEN` should be added to Vercel/Production env as well (`env(ODYSSEY_AUTO_BLOCK_TOKEN)`)

## Manual Recovery / Testing
- Admin can manually trigger: `POST /api/admin/members/auto-block` with admin JWT or internal token
- `workflow_dispatch` allows manual run from GitHub Actions UI
- Blocked users remain visible in `GET /api/admin/members`; unblock via `POST /api/admin/members/{uid}/unblock`

## Historical Data Safety
- Blocking only mutates `odyssey_user_profiles` columns `is_active`, `blocked_at`, `blocked_by`, `block_reason`, `updated_at`
- No `DELETE` on `odyssey_user_profiles`, `odyssey_task_submissions`, `odyssey_claims`, `odyssey_coin_transactions`
- No update to historical `coins_earned` or ledger

## Operations Checklist
1. Set `ODYSSEY_AUTO_BLOCK_TOKEN` (random 32+ chars) in GitHub Actions secrets and production env
2. Set `ODYSSEY_PRODUCTION_URL` in GitHub Actions secrets
3. Verify workflow runs: Actions → Auto-block inactive users → Run workflow → check `success`, `blocked_count`, `period_start/end`
4. Monitor: workflow fails on non-2xx, timeout 30s, retry 2, concurrency prevents overlap
