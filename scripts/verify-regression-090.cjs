require('dotenv').config();

// Regression verification for Migration 090 execution.
// Read-only EXCEPT: no writes at all. Verifies:
// - 627-629 (09-12), 630-632 (09-13), 633-635 (09-15), 636-638 (09-16),
//   639-641 (09-17), 642-644 (09-18), 645-647 (09-19): titles / active /
//   type / eval / rewards / steps / camera_only flags (610=true,
//   628=false, 632=false, 643=false, 648-653 new) / submission counts.
// - No orphan submissions/ledger referencing deleted IDs 550-553.
// - Economy read-only fingerprint: payout config, monthly cap,
//   system config (announcement), shop row count.

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

async function run() {
  console.log('====================================================');
  console.log('REGRESSION VERIFICATION — Migration 090 (627-647 + economy)');
  console.log('====================================================\n');

  console.log('[1] Task sets 627-647 (must be intact, 3 active per date)');
  const dates = ['2026-09-12', '2026-09-13', '2026-09-15', '2026-09-16', '2026-09-17', '2026-09-18', '2026-09-19'];
  for (const d of dates) {
    const rows = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,is_active,config&order=step_order.asc`);
    const active = rows.filter((t) => t.is_active);
    console.log(`   ${d}: total=${rows.length} active=${active.length}`);
    active.forEach((t) => console.log(`     - [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp`));
    if (active.length !== 3) bad(`${d}: expected 3 active, found ${active.length}`);
    else ok(`${d} has exactly 3 active tasks`);
  }

  console.log('\n[2] camera_only flags (610=true, 628=false, 632=false, 643=false)');
  const cam = await get('odyssey_tasks?id=in.(610,628,632,643)&select=id,title,config');
  const flag = (id, want) => {
    const row = cam.find((t) => t.id === id);
    const val = row?.config?.camera_only;
    // Absent camera_only behaves as false (gallery allowed) per LinearPath.
    const eff = val === true;
    if (eff === want) ok(`${id} "${row?.title}" camera_only effective=${eff} (raw=${JSON.stringify(val)})`);
    else bad(`${id} "${row?.title}" camera_only effective=${eff}, expected ${want}`);
  };
  flag(610, true); flag(628, false); flag(632, false); flag(643, false);

  console.log('\n[3] No orphan artifacts referencing deleted IDs 550-553');
  const orphSubs = await get('odyssey_task_submissions?task_id=in.(550,551,552,553)&select=id');
  if (orphSubs.length) bad(`orphan submissions: ${JSON.stringify(orphSubs)}`);
  else ok('zero submissions reference deleted tasks 550-553');
  const gone = await get('odyssey_tasks?id=in.(550,551,552,553)&select=id');
  if (gone.length) bad(`deleted tasks still present: ${JSON.stringify(gone)}`);
  else ok('tasks 550-553 absent (exact-ID delete verified)');

  console.log('\n[4] Submission/ledger totals on regression sets (report)');
  for (const id of [627, 628, 629, 630, 631, 632, 633, 634, 635, 636, 637, 638, 639, 640, 641, 642, 643, 644, 645, 646, 647]) {
    const s = await get(`odyssey_task_submissions?task_id=eq.${id}&select=id,status`);
    console.log(`   task ${id}: submissions=${s.length} [${s.map((x) => x.status).join(',')}]`);
  }

  console.log('\n[5] New 09-20/09-21 tasks present (report)');
  for (const d of ['2026-09-20', '2026-09-21']) {
    const rows = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,is_active&order=step_order.asc`);
    rows.forEach((t) => console.log(`   ${d} [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp active=${t.is_active}`));
    const active = rows.filter((t) => t.is_active);
    if (active.length !== 3) bad(`${d}: expected 3 active, found ${active.length}`);
    else ok(`${d} has exactly 3 active tasks`);
  }

  console.log('\n[6] Economy fingerprint (read-only, must be untouched by 090)');
  const payout = await get('odyssey_user_payout_config?select=*');
  console.log(`   odyssey_user_payout_config rows: ${payout.length} ${JSON.stringify(payout).slice(0, 400)}`);
  const sys = await get('odyssey_system_config?select=key,value&order=key.asc');
  console.log(`   odyssey_system_config keys: ${JSON.stringify(sys.map((x) => x.key))}`);
  for (const x of sys) console.log(`     - ${x.key} = ${JSON.stringify(x.value).slice(0, 200)}`);
  const shop = await get('odyssey_reward_catalog?select=id&limit=100');
  console.log(`   reward catalog items: ${shop.length}`);
  const ver = await get(`odyssey_schema_version?select=value&key=eq.schema_version`);
  console.log(`   schema_version: ${JSON.stringify(ver)}`);
  if (ver.length === 1 && ver[0].value === '090_tasks_replace_2026_09_20_21') ok('schema_version = 090 (only expected version bump)');
  else bad(`unexpected schema_version: ${JSON.stringify(ver)}`);

  console.log('\n====================================================');
  console.log(`REGRESSION RESULT: ${failures === 0 ? 'ALL PASS' : 'FAILURES: ' + failures}`);
  console.log('====================================================');
  if (failures) process.exit(1);
}

run().catch((err) => {
  console.error('Regression check failed:', err);
  process.exit(1);
});
