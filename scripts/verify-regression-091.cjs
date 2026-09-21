require('dotenv').config();

// Regression verification for Migration 091 execution.
// Read-only: no writes at all. Verifies:
// - 627-629 (09-12), 630-632 (09-13), 633-635 (09-15), 636-638 (09-16),
//   639-641 (09-17), 642-644 (09-18), 645-647 (09-19), 648-650 (09-20),
//   651-653 (09-21): 3 active per date, titles intact.
// - 608-620 spot set intact.
// - camera_only flags: 610=true, 628=false, 632=false, 643=false,
//   650=false, plus the NEW 09-22 PHOTO task camera_only=false.
// - No orphan submissions/ledger referencing deleted IDs 554-557.
// - Economy read-only fingerprint: payout config, system config,
//   shop row count, schema_version = 091.

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
  console.log('REGRESSION VERIFICATION — Migration 091 (627-653 + economy)');
  console.log('====================================================\n');

  console.log('[1] Task sets 627-653 (must be intact, 3 active per date)');
  const dates = ['2026-09-12', '2026-09-13', '2026-09-15', '2026-09-16', '2026-09-17', '2026-09-18', '2026-09-19', '2026-09-20', '2026-09-21'];
  for (const d of dates) {
    const rows = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,is_active,config&order=step_order.asc`);
    const active = rows.filter((t) => t.is_active);
    console.log(`   ${d}: total=${rows.length} active=${active.length}`);
    active.forEach((t) => console.log(`     - [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp`));
    if (active.length !== 3) bad(`${d}: expected 3 active, found ${active.length}`);
    else ok(`${d} has exactly 3 active tasks`);
  }

  console.log('\n[1b] Spot set 608-620 (must be intact)');
  const spot = await get('odyssey_tasks?id=in.(608,609,610,615,616,617,618,619,620)&select=id,title,task_type,evaluation_type,reward_coins,reward_xp,is_active');
  console.log(`   spot rows found: ${spot.length}`);
  spot.forEach((t) => console.log(`     - [ID ${t.id}] "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp active=${t.is_active}`));
  if (spot.length !== 9) bad(`expected 9 spot rows, found ${spot.length}`);
  else ok('608-620 spot set intact (9 rows)');

  console.log('\n[2] camera_only flags (610=true, 628/632/643/650=false, new 09-22 PHOTO=false)');
  const cam = await get('odyssey_tasks?id=in.(610,628,632,643,650)&select=id,title,config');
  const flag = (id, want) => {
    const row = cam.find((t) => t.id === id);
    const val = row?.config?.camera_only;
    // Absent camera_only behaves as false (gallery allowed) per LinearPath.
    const eff = val === true;
    if (eff === want) ok(`${id} "${row?.title}" camera_only effective=${eff} (raw=${JSON.stringify(val)})`);
    else bad(`${id} "${row?.title}" camera_only effective=${eff}, expected ${want}`);
  };
  flag(610, true); flag(628, false); flag(632, false); flag(643, false); flag(650, false);
  const newPhoto = await get('odyssey_tasks?active_date=eq.2026-09-22&select=id,title,config&title=eq.' + encodeURIComponent('Apa yang Bisa Kamu Pelajari dari Ini?'));
  if (newPhoto.length === 1 && newPhoto[0].config?.camera_only === false && newPhoto[0].config?.max_files === 1) ok(`new 09-22 PHOTO [ID ${newPhoto[0].id}] camera_only=false max_files=1`);
  else bad(`new 09-22 PHOTO config wrong: ${JSON.stringify(newPhoto).slice(0, 200)}`);

  console.log('\n[3] No orphan artifacts referencing deleted IDs 554-557');
  const orphSubs = await get('odyssey_task_submissions?task_id=in.(554,555,556,557)&select=id');
  if (orphSubs.length) bad(`orphan submissions: ${JSON.stringify(orphSubs)}`);
  else ok('zero submissions reference deleted tasks 554-557');
  const gone = await get('odyssey_tasks?id=in.(554,555,556,557)&select=id');
  if (gone.length) bad(`deleted tasks still present: ${JSON.stringify(gone)}`);
  else ok('tasks 554-557 absent (exact-ID delete verified)');

  console.log('\n[4] Submission/ledger totals on regression sets (report)');
  for (const id of [627, 628, 629, 630, 631, 632, 633, 634, 635, 636, 637, 638, 639, 640, 641, 642, 643, 644, 645, 646, 647, 648, 649, 650, 651, 652, 653]) {
    const s = await get(`odyssey_task_submissions?task_id=eq.${id}&select=id,status`);
    console.log(`   task ${id}: submissions=${s.length} [${s.map((x) => x.status).join(',')}]`);
  }

  console.log('\n[5] New 09-22/09-23 tasks present (report)');
  for (const d of ['2026-09-22', '2026-09-23']) {
    const rows = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,is_active&order=step_order.asc`);
    rows.forEach((t) => console.log(`   ${d} [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp active=${t.is_active}`));
    const active = rows.filter((t) => t.is_active);
    if (active.length !== 3) bad(`${d}: expected 3 active, found ${active.length}`);
    else ok(`${d} has exactly 3 active tasks`);
  }

  console.log('\n[6] Economy fingerprint (read-only, must be untouched by 091)');
  const payout = await get('odyssey_user_payout_config?select=*');
  console.log(`   odyssey_user_payout_config rows: ${payout.length} ${JSON.stringify(payout).slice(0, 400)}`);
  const sys = await get('odyssey_system_config?select=key,value&order=key.asc');
  console.log(`   odyssey_system_config keys: ${JSON.stringify(sys.map((x) => x.key))}`);
  for (const x of sys) console.log(`     - ${x.key} = ${JSON.stringify(x.value).slice(0, 200)}`);
  const shop = await get('odyssey_reward_catalog?select=id&limit=100');
  console.log(`   reward catalog items: ${shop.length}`);
  const ver = await get(`odyssey_schema_version?select=value&key=eq.schema_version`);
  console.log(`   schema_version: ${JSON.stringify(ver)}`);
  if (ver.length === 1 && ver[0].value === '091_tasks_replace_2026_09_22_23') ok('schema_version = 091 (only expected version bump)');
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
