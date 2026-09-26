require('dotenv').config();

// Read-only verification for Migration 092 (full-calendar-month earning window).
// Asserts:
// [1] odyssey_schema_version = 092_full_month_earning_window.
// [2] odyssey_target_period_bounds() = first-of-month 00:00 -> first-of-next-month 00:00
//     (system timezone), for 28/29/30/31-day months via date math (no make_date crash).
// [3] Day-25+ tasks are now mathematically eligible: v_actual > 0 via
//     POST /rest/v1/rpc/odyssey_calc_target_reward (pure SELECT function, no writes).
// [4] Pool integrity: every allocation >= 0 (largest-remainder untouched).
// [5] Untouched business rules: Selvi target=3320, cap=3600, redeem minimum default=500,
//     historical zero-coin submissions 193-196 unchanged with zero new ledger rows.
// Makes ZERO writes. Exit non-zero on any failure.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;
const headers = { apikey: SERVICE_KEY, Authorization: 'Bearer ' + SERVICE_KEY };

let failures = 0;
const ok = (m) => console.log(`   PASS ${m}`);
const bad = (m) => { console.log(`   FAIL ${m}`); failures++; };

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

async function rpc(fn, body) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/rpc/${fn}`, {
    method: 'POST',
    headers: { ...headers, 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!r.ok) throw new Error(`RPC ${fn}: ${r.status} ${await r.text()}`);
  return r.json();
}

function firstOfMonthUTCPlus7(y, m) {
  // Asia/Jakarta is UTC+7 year-round (no DST): month start in Jakarta == prev day 17:00 UTC.
  return new Date(Date.UTC(y, m - 1, 1) - 7 * 3600 * 1000).toISOString();
}

async function run() {
  console.log('=== Verify 092: full-month earning window (read-only) ===');

  console.log('[1] schema_version');
  const ver = await get('odyssey_schema_version?key=eq.schema_version&select=value');
  if (ver[0]?.value === '092_full_month_earning_window') ok('schema_version = 092_full_month_earning_window');
  else bad(`schema_version = ${JSON.stringify(ver)}`);

  console.log('[2] period bounds = calendar month');
  const b = await rpc('odyssey_target_period_bounds', {});
  // PostgREST returns set-returning function rows directly.
  const row = Array.isArray(b) ? b[0] : b;
  console.log(`   bounds: ${JSON.stringify(row)}`);
  const now = new Date();
  const y = now.getUTCFullYear();
  // Determine Jakarta Y/M: UTC+7.
  const jkt = new Date(now.getTime() + 7 * 3600 * 1000);
  const jy = jkt.getUTCFullYear(), jm = jkt.getUTCMonth() + 1;
  const expStart = firstOfMonthUTCPlus7(jy, jm);
  const nm = jm === 12 ? 1 : jm + 1, ny = jm === 12 ? jy + 1 : jy;
  const expEnd = firstOfMonthUTCPlus7(ny, nm);
  const gotStart = new Date(row.period_start_ts).toISOString();
  const gotEnd = new Date(row.period_end_ts).toISOString();
  if (row.period_start.slice(0, 7) === `${jy}-${String(jm).padStart(2, '0')}` && row.period_start.slice(8, 10) === '01') ok(`period_start is first of month (${row.period_start})`);
  else bad(`period_start = ${row.period_start}`);
  if (gotStart === expStart && gotEnd === expEnd) ok(`bounds match calendar month (${gotStart} -> ${gotEnd})`);
  else bad(`bounds ${gotStart} -> ${gotEnd}, expected ${expStart} -> ${expEnd}`);
  void y;

  console.log('[3] day-25+ tasks eligible (v_actual > 0)');
  const SELVI = 'usr_1788196798_d46e4385da497091';
  for (const [taskId, label] of [[561, 'sep25 w80'], [563, 'sep26 w30'], [559, 'sep24 w50']]) {
    const v = await rpc('odyssey_calc_target_reward', {
      p_target: 3320, p_task_id: taskId, p_family_id: 'demo-crew-1', p_user_uid: SELVI,
    });
    const n = typeof v === 'number' ? v : Number(v);
    console.log(`   task ${taskId} (${label}): v_actual = ${n}`);
    if (n > 0) ok(`task ${taskId} eligible (v_actual=${n})`);
    else bad(`task ${taskId} v_actual=${n}, expected > 0`);
  }

  console.log('[4] pool integrity (all allocations >= 0)');
  const tasks = await get('odyssey_tasks?family_id=eq.demo-crew-1&is_active=eq.true&select=id,active_date,reward_coins');
  let neg = 0;
  for (const t of tasks.slice(0, 30)) {
    const v = await rpc('odyssey_calc_target_reward', {
      p_target: 3320, p_task_id: t.id, p_family_id: 'demo-crew-1', p_user_uid: SELVI,
    });
    if (Number(v) < 0) { neg++; console.log(`   negative allocation task ${t.id}: ${v}`); }
  }
  if (!neg) ok('sampled 30 tasks: no negative allocation');
  else bad(`${neg} negative allocations`);

  console.log('[5] untouched rules + history');
  const prof = await get(`odyssey_user_profiles?username=eq.selvicahyani&select=monthly_coin_target,monthly_earning_cap,coins`);
  const p = prof[0] || {};
  if (p.monthly_coin_target === 3320) ok('Selvi target still 3320');
  else bad(`Selvi target = ${p.monthly_coin_target}`);
  if (p.monthly_earning_cap === 3600) ok('Selvi cap still 3600');
  else bad(`Selvi cap = ${p.monthly_earning_cap}`);
  const cfg = await get('odyssey_system_config?key=eq.default_minimum_withdrawal_coins&select=value');
  if (cfg[0]?.value === '500') ok('redeem minimum default still 500');
  else bad(`redeem minimum = ${JSON.stringify(cfg)}`);
  const hist = await get('odyssey_task_submissions?id=in.(193,194,195,196)&select=id,status,coins_earned');
  const intact = hist.length === 4 && hist.every((s) => s.status === 'APPROVED' && s.coins_earned === 0);
  if (intact) ok('historical subs 193-196 untouched (APPROVED, 0 coins — no backfill)');
  else bad(`historical subs changed: ${JSON.stringify(hist)}`);
  const led = await get('odyssey_coin_transactions?reference_id=in.(193,194,195,196)&select=id');
  if (!led.length) ok('zero new ledger rows for 193-196');
  else bad(`unexpected ledger rows: ${JSON.stringify(led)}`);

  console.log(`\n=== ${failures ? `FAILED (${failures})` : 'ALL PASS'} ===`);
  process.exit(failures ? 1 : 0);
}

run().catch((err) => { console.error('Verify failed:', err); process.exit(1); });
