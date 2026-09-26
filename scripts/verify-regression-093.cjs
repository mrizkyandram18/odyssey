require('dotenv').config();

// Regression verification for Migration 093 execution.
// Read-only: no writes at all. Verifies:
// - 627-629 (09-12), 630-632 (09-13), 633-635 (09-15), 636-638 (09-16),
//   639-641 (09-17), 642-644 (09-18), 645-647 (09-19), 648-650 (09-20),
//   651-653 (09-21), 654-656 (09-22), 657-659 (09-23): 3 active per date,
//   titles intact.
// - 608-620 spot set intact.
// - camera_only flags: 610=true, 628=false, 632=false, 643=false,
//   650=false, plus the NEW 09-27/09-28 PHOTO tasks camera_only=false.
// - No orphan submissions referencing deleted IDs 564-567.
// - New 09-27/09-28 tasks (660-665) present with exact type/eval/reward.
// - Economy read-only fingerprint: payout config, system config,
//   shop row count, schema_version = 093.

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
  console.log('REGRESSION VERIFICATION — Migration 093 (608-665 + economy)');
  console.log('====================================================\n');

  console.log('[1] Task sets 627-659 (must be intact, 3 active per date)');
  const dates = ['2026-09-12', '2026-09-13', '2026-09-15', '2026-09-16', '2026-09-17', '2026-09-18', '2026-09-19', '2026-09-20', '2026-09-21', '2026-09-22', '2026-09-23'];
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

  console.log('\n[2] camera_only flags (610=true, 628/632/643/650=false, new 09-27/09-28 PHOTO=false)');
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
  const newPhotos = await get('odyssey_tasks?id=in.(660,663)&select=id,title,config');
  for (const row of newPhotos) {
    if (row.config?.camera_only === false && row.config?.max_files === 1) ok(`new PHOTO [ID ${row.id}] "${row.title}" camera_only=false max_files=1`);
    else bad(`new PHOTO [ID ${row.id}] config wrong: ${JSON.stringify(row.config).slice(0, 200)}`);
  }
  if (newPhotos.length !== 2) bad(`expected 2 new PHOTO rows, found ${newPhotos.length}`);

  console.log('\n[3] No orphan artifacts referencing deleted IDs 564-567');
  const orphSubs = await get('odyssey_task_submissions?task_id=in.(564,565,566,567)&select=id');
  if (orphSubs.length) bad(`orphan submissions: ${JSON.stringify(orphSubs)}`);
  else ok('zero submissions reference deleted tasks 564-567');
  const gone = await get('odyssey_tasks?id=in.(564,565,566,567)&select=id');
  if (gone.length) bad(`deleted tasks still present: ${JSON.stringify(gone)}`);
  else ok('tasks 564-567 absent (exact-ID delete verified)');

  console.log('\n[4] Prior-task real submissions untouched (654-659 report) + new tasks clean (660-665 must be 0)');
  for (const id of [654, 655, 656, 657, 658, 659]) {
    const s = await get(`odyssey_task_submissions?task_id=eq.${id}&select=id,status,user_uid`);
    console.log(`   task ${id}: submissions=${s.length} [${s.map((x) => x.status).join(',')}] owner=${s[0]?.user_uid || '-'}`);
    if (!s.length) bad(`task ${id} lost its real submissions — prior-task data touched`);
    else if (s.some((x) => String(x.user_uid).startsWith('smoke-uid-'))) bad(`task ${id} carries smoke residue`);
    else ok(`task ${id} real submissions intact, no smoke residue`);
  }
  for (const id of [660, 661, 662, 663, 664, 665]) {
    const s = await get(`odyssey_task_submissions?task_id=eq.${id}&select=id,status`);
    console.log(`   task ${id}: submissions=${s.length} [${s.map((x) => x.status).join(',')}]`);
    if (s.length) bad(`task ${id} has unexpected submissions post-cleanup`);
  }
  ok('new tasks 660-665 carry zero submissions after smoke cleanup');

  console.log('\n[5] New 09-27/09-28 tasks present with exact type/eval/reward/step');
  const exp = {
    660: ['Sehari Kembali Pakai Seragam', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 40, 100, 1],
    661: ['Kalau Sekarang Masih Pakai Seragam', 'VIDEO', 'ADMIN_REVIEW', 40, 100, 2],
    662: ['Balik Jadi Anak Sekolah Sehari', 'MINI_GAME', 'AUTO', 50, 100, 3],
    663: ['Seragam vs Dirimu Sekarang', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 40, 100, 1],
    664: ['Kalau Bisa Balik ke Hari Pertama Sekolah', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 40, 100, 2],
    665: ['Dari Seragam Sekolah ke Seragam Kerja', 'VIDEO', 'ADMIN_REVIEW', 40, 100, 3],
  };
  const news = await get('odyssey_tasks?id=in.(660,661,662,663,664,665)&select=id,title,task_type,evaluation_type,reward_coins,reward_xp,step_order,active_date,is_active,config&order=id.asc');
  for (const t of news) {
    const e = exp[t.id];
    const match = e && t.title === e[0] && t.task_type === e[1] && t.evaluation_type === e[2] && t.reward_coins === e[3] && t.reward_xp === e[4] && t.step_order === e[5] && t.is_active === true;
    console.log(`   [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp date=${t.active_date} active=${t.is_active}`);
    if (match) ok(`task ${t.id} fingerprint exact`);
    else bad(`task ${t.id} fingerprint MISMATCH`);
  }
  if (news.length !== 6) bad(`expected 6 new tasks, found ${news.length}`);
  // Config spot checks.
  const byId = Object.fromEntries(news.map((t) => [t.id, t]));
  if (byId[662]?.config?.scenario?.events?.length === 5) ok('662 MINI_GAME has 5 scenario events');
  else bad('662 MINI_GAME event count wrong');
  if (byId[661]?.config?.recording?.max_duration_seconds === 60 && byId[665]?.config?.recording?.max_duration_seconds === 60) ok('661+665 VIDEO max_duration_seconds=60');
  else bad('VIDEO max duration wrong');
  if (byId[664]?.config?.minimum_characters === 150 && byId[664]?.config?.maximum_characters === 2000) ok('664 TEXT min/max 150/2000');
  else bad('TEXT min/max wrong');

  console.log('\n[6] Economy fingerprint (read-only, must be untouched by 093)');
  const payout = await get('odyssey_user_payout_config?select=*');
  console.log(`   odyssey_user_payout_config rows: ${payout.length} ${JSON.stringify(payout).slice(0, 400)}`);
  const sys = await get('odyssey_system_config?select=key,value&order=key.asc');
  console.log(`   odyssey_system_config keys: ${JSON.stringify(sys.map((x) => x.key))}`);
  const econ = Object.fromEntries(sys.map((x) => [x.key, x.value]));
  if (econ.default_monthly_earning_cap === '3320') ok('default_monthly_earning_cap=3320 unchanged');
  else bad(`default_monthly_earning_cap changed: ${econ.default_monthly_earning_cap}`);
  if (econ.default_monthly_coin_target === '0') ok('default_monthly_coin_target=0 unchanged');
  else bad(`default_monthly_coin_target changed: ${econ.default_monthly_coin_target}`);
  if (econ.coin_conversion_rate === '100') ok('coin_conversion_rate=100 unchanged');
  else bad(`coin_conversion_rate changed: ${econ.coin_conversion_rate}`);
  const shop = await get('odyssey_reward_catalog?select=id&limit=100');
  console.log(`   reward catalog items: ${shop.length}`);
  const ver = await get(`odyssey_schema_version?select=value&key=eq.schema_version`);
  console.log(`   schema_version: ${JSON.stringify(ver)}`);
  if (ver.length === 1 && ver[0].value === '093_tasks_replace_2026_09_27_28') ok('schema_version = 093 (only expected version bump)');
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
