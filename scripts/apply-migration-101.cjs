require('dotenv').config();

// Applies Migration 101: Fresh tasks for 21-25 October 2026 (15 tasks).
// Payloads are parsed from supabase/migrations/101_tasks_2026_10_21_25.sql
// (single source of truth; no duplicated text here).
// 1. Pre-audit: expects 0 rows on each of the 5 dates. ABORTS if any row
//    exists with submissions/ledger, or any unexpected row at all.
// 2. Idempotent pre-clean: deletes ONLY the 15 exact new titles on their
//    exact dates (previous partial run of THIS migration).
// 3. Inserts the 15 tasks (payload mirrors 101_tasks_2026_10_21_25.sql).
// 4. Verifies exactly 3 active tasks per date, step 1-3, exact titles,
//    and reward sum 20+30+40 = 90 (<= 100) per date.
// 5. Economy fingerprint before/after must be identical.
// 6. Bumps odyssey_schema_version to 101_tasks_2026_10_21_25.
// Does NOT touch any other date, task (incl. 600/610 anchors, batches
// 666-677, 678-692), config, or economy value.

const fs = require('fs');
const path = require('path');

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const DATES = ['2026-10-21', '2026-10-22', '2026-10-23', '2026-10-24', '2026-10-25'];

// Parse tuples out of the canonical SQL file.
function parseSqlPayloads() {
  const sqlPath = path.join(__dirname, '..', 'supabase', 'migrations', '101_tasks_2026_10_21_25.sql');
  const sql = fs.readFileSync(sqlPath, 'utf8');
  const tupleRe = /\(\s*'demo-crew-1',\s*'([^']+)',\s*'((?:[^'])*)',\s*'([A-Z_]+)',\s*'([A-Z_]+)',\s*(\d),\s*'(\d{4}-\d{2}-\d{2})',\s*(\d+),\s*(\d+),\s*'ALL',\s*TRUE,\s*'\{[\s\S]*?\}'::jsonb\s*\)/g;
  const tasks = [];
  let m;
  while ((m = tupleRe.exec(sql)) !== null) {
    const [, title, description, taskType, evalType, step, date, coins, xp, , tail] = m;
    const cfgRaw = m[0].match(/'\{[\s\S]*?\}'::jsonb/)[0];
    const config = JSON.parse(cfgRaw.slice(1, -8));
    tasks.push({
      family_id: 'demo-crew-1',
      title,
      description,
      task_type: taskType,
      evaluation_type: evalType,
      step_order: Number(step),
      active_date: date,
      reward_coins: Number(coins),
      reward_xp: Number(xp),
      target_scope: 'ALL',
      is_active: true,
      config,
    });
  }
  return tasks;
}

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

async function run() {
  console.log('=== Applying Migration 101: Fresh tasks for 2026-10-21..25 ===');

  const tasks = parseSqlPayloads();
  if (tasks.length !== 15) throw new Error(`Expected 15 parsed payloads, got ${tasks.length}`);
  const NEW_TITLES = {};
  for (const t of tasks) (NEW_TITLES[t.active_date] ||= []).push(t.title);
  for (const d of DATES) {
    if (!NEW_TITLES[d] || NEW_TITLES[d].length !== 3) throw new Error(`Expected 3 titles on ${d}`);
    const sum = tasks.filter((t) => t.active_date === d).reduce((a, t) => a + t.reward_coins, 0);
    if (sum > 100) throw new Error(`Reward sum ${sum} exceeds 100 coins on ${d}`);
  }

  console.log('[0] Economy fingerprint (before)...');
  const cfgBefore = await get('odyssey_system_config?select=key,value');
  const keep = ['coin_conversion_rate', 'payout_target_rupiah', 'max_payout_coins', 'default_monthly_earning_cap', 'default_monthly_coin_target', 'default_minimum_withdrawal_coins', 'default_payout_frequency', 'timezone'];
  const fp = {};
  for (const row of cfgBefore) if (keep.includes(row.key)) fp[row.key] = row.value;
  console.log('   BEFORE: ' + JSON.stringify(fp));

  console.log('[1] Pre-apply audit of target dates...');
  for (const d of DATES) {
    const rows = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title,step_order,is_active`);
    console.log(`   rows on ${d}: ` + JSON.stringify(rows));
    if (rows.length) {
      for (const t of rows) {
        const s = await get(`odyssey_task_submissions?task_id=eq.${t.id}&select=id,status`);
        if (s.length) throw new Error(`ABORT: task ${t.id} carries ${s.length} submissions — deactivate instead of delete.`);
      }
      throw new Error(`ABORT: unexpected rows on ${d} without submissions — audit manually before proceeding.`);
    }
  }

  console.log('[2] Idempotent pre-clean (exact new titles, exact dates)...');
  const enc = (s) => encodeURIComponent(`"${s}"`);
  let cleaned = 0;
  for (const d of DATES) {
    const pre = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title&title=in.(${NEW_TITLES[d].map(enc).join(',')})`);
    if (pre.length) {
      const ids = pre.map((t) => t.id).join(',');
      const del = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
      if (!del.ok) throw new Error(`Pre-clean DELETE failed: ${del.status} ${await del.text()}`);
      cleaned += pre.length;
      console.log(`   removed ${pre.length} partial-run row(s) on ${d}.`);
    }
  }
  if (!cleaned) console.log('   no partial-run rows found.');

  console.log('[3] Inserting 15 tasks...');
  const postRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks`, {
    method: 'POST',
    headers,
    body: JSON.stringify(tasks),
  });
  if (!postRes.ok) throw new Error(`INSERT failed: ${postRes.status} ${await postRes.text()}`);
  const created = await postRes.json();
  console.log(`   inserted ${created.length} tasks:`);
  created.forEach((t) => {
    console.log(`   - [ID ${t.id}] ${t.active_date} #${t.step_order}: "${t.title}" (${t.task_type}, ${t.evaluation_type}, ${t.reward_coins}c/${t.reward_xp}xp)`);
  });

  console.log('[4] Verifying target dates...');
  for (const d of DATES) {
    const after = await get(
      `odyssey_tasks?active_date=eq.${d}&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,target_scope,family_id,is_active,config&order=step_order.asc`
    );
    const active = after.filter((t) => t.is_active);
    console.log(`   ${d}: total rows=${after.length}, active=${active.length}`);
    active.forEach((t) => console.log(`   - [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp scope=${t.target_scope} fam=${t.family_id}`));
    if (active.length !== 3) throw new Error(`Expected exactly 3 active tasks on ${d}, found ${active.length}`);
    if (JSON.stringify(active.map((t) => t.title)) !== JSON.stringify(NEW_TITLES[d])) {
      throw new Error('Title/step mismatch on ' + d + ': ' + JSON.stringify(active.map((t) => t.title)));
    }
    const sum = active.reduce((a, t) => a + t.reward_coins, 0);
    if (sum !== 90) throw new Error(`Reward sum on ${d} is ${sum}, expected 90 (limit 100)`);
  }

  console.log('[5] Economy fingerprint (after)...');
  const cfgAfter = await get('odyssey_system_config?select=key,value');
  const fpAfter = {};
  for (const row of cfgAfter) if (keep.includes(row.key)) fpAfter[row.key] = row.value;
  console.log('   AFTER:  ' + JSON.stringify(fpAfter));
  if (JSON.stringify(fp) !== JSON.stringify(fpAfter)) throw new Error('ECONOMY FINGERPRINT CHANGED — investigate before proceeding.');

  console.log('[6] Updating odyssey_schema_version...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_schema_version`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '101_tasks_2026_10_21_25', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 101 Applied Successfully ===');
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
