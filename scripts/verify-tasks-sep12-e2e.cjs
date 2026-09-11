require('dotenv').config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

async function get(path) {
  const res = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!res.ok) throw new Error(`GET ${path} failed: ${res.status} ${await res.text()}`);
  return res.json();
}

async function patch(path, body) {
  const res = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { method: 'PATCH', headers, body: JSON.stringify(body) });
  if (!res.ok) throw new Error(`PATCH ${path} failed: ${res.status} ${await res.text()}`);
  return res.json();
}

async function run() {
  console.log('====================================================');
  console.log('E2E VERIFICATION: 2026-09-12 tasks + payout announcement');
  console.log('(config-driven: DB is source of truth, no hardcoded runtime)');
  console.log('====================================================\n');

  let allOk = true;
  const fail = (msg) => { console.log(`   🔴 ${msg}`); allOk = false; };
  const pass = (msg) => console.log(`   🟢 ${msg}`);

  // --- 1. Verify 3 canonical tasks exist with expected admin-configured values ---
  console.log('[1] Canonical tasks for 2026-09-12');
  const tasks = await get('odyssey_tasks?active_date=eq.2026-09-12&is_active=eq.true&order=step_order.asc&select=id,title,description,task_type,evaluation_type,step_order,active_date,reward_coins,reward_xp,target_scope,is_active,config');
  if (tasks.length !== 3) fail(`Expected 3 active tasks, found ${tasks.length}`);
  else pass(`Found 3 active tasks`);

  const byStep = {};
  for (const t of tasks) byStep[t.step_order] = t;

  const s1 = byStep[1];
  if (!s1 || s1.title !== 'Pilih yang Paling Penting' || s1.task_type !== 'MINI_GAME' || s1.evaluation_type !== 'AUTO' || s1.reward_coins !== 50 || s1.reward_xp !== 100) {
    fail(`Step 1 mismatch: ${JSON.stringify(s1)}`);
  } else {
    pass(`Step 1 [ID ${s1.id}] MINI_GAME 50c/100xp AUTO`);
    const evts = s1.config?.scenario?.events;
    if (!Array.isArray(evts) || evts.length !== 6) fail(`Step 1 scenario must have 6 events, got ${evts?.length}`);
    else pass(`Step 1 scenario has 6 events from DB config`);
    if (s1.config?.scenario?.currency !== 'POINTS') fail(`Step 1 currency must be POINTS`);
    else pass(`Step 1 currency=POINTS initial_balance=${s1.config.scenario.initial_balance} target_score=${s1.config.target_score}`);
  }

  const s2 = byStep[2];
  if (!s2 || s2.title !== 'Bikin Rencana Sederhana' || s2.task_type !== 'DOCUMENT_UPLOAD' || s2.reward_coins !== 40 || s2.reward_xp !== 100) {
    fail(`Step 2 mismatch: ${JSON.stringify(s2)}`);
  } else {
    pass(`Step 2 [ID ${s2.id}] DOCUMENT_UPLOAD 40c/100xp eval=${s2.evaluation_type}`);
    if (!Array.isArray(s2.config?.allowed_extensions) || s2.config.max_file_size_mb !== 10) fail(`Step 2 document config mismatch`);
    else pass(`Step 2 document config from DB (extensions=${s2.config.allowed_extensions.join(',')} max=${s2.config.max_file_size_mb}MB)`);
  }

  const s3 = byStep[3];
  if (!s3 || s3.title !== 'Kalau Rencana Berubah' || s3.task_type !== 'TEXT_RESPONSE' || s3.evaluation_type !== 'ADMIN_REVIEW' || s3.reward_coins !== 30 || s3.reward_xp !== 100) {
    fail(`Step 3 mismatch: ${JSON.stringify(s3)}`);
  } else {
    pass(`Step 3 [ID ${s3.id}] TEXT_RESPONSE 30c/100xp ADMIN_REVIEW`);
    if (s3.config?.minimum_characters !== 80 || s3.config?.maximum_characters !== 1000) fail(`Step 3 text bounds must be 80/1000`);
    else pass(`Step 3 text bounds from DB config (min=80 max=1000)`);
  }

  // --- 2. Test A: admin editability of a task (non-destructive, restored) ---
  console.log('\n[2] Test A — Task admin editability (description suffix + restore)');
  const originalDesc = s3.description;
  const probeSuffix = ' [uji-konfigurasi]';
  await patch(`odyssey_tasks?id=eq.${s3.id}`, { description: originalDesc + probeSuffix });
  const afterEdit = await get(`odyssey_tasks?id=eq.${s3.id}&select=id,description`);
  if (afterEdit[0]?.description !== originalDesc + probeSuffix) fail('Edited description not reflected on re-read');
  else pass(`ID ${s3.id} description edit reflected (member runtime reads same row)`);
  await patch(`odyssey_tasks?id=eq.${s3.id}`, { description: originalDesc });
  const afterRestore = await get(`odyssey_tasks?id=eq.${s3.id}&select=id,description`);
  if (afterRestore[0]?.description !== originalDesc) fail('Description restore failed');
  else pass(`ID ${s3.id} description restored to original`);

  // --- 3. Announcement config seeded via existing odyssey_system_config ---
  console.log('\n[3] Payout announcement (existing system_config keys)');
  const annRows = await get('odyssey_system_config?key=like.announcement_*&select=key,value');
  const ann = Object.fromEntries(annRows.map(r => [r.key, r.value]));
  const expectedBody = 'Pencairan saat ini mengalami penyesuaian jadwal karena sedang dilakukan proses verifikasi dan pengecekan transaksi. Hal ini dilakukan agar setiap pencairan dapat diproses dengan aman dan akurat. Mohon menunggu dan tidak perlu mengajukan ulang pencairan yang sudah masuk. Terima kasih atas pengertiannya.';
  if (ann.announcement_enabled !== 'true') fail(`announcement_enabled must be 'true', got ${ann.announcement_enabled}`);
  else pass(`announcement_enabled=true`);
  if (ann.announcement_title !== 'Pengumuman Pencairan') fail('announcement_title mismatch');
  else pass(`announcement_title from DB config`);
  if (ann.announcement_body !== expectedBody) fail('announcement_body mismatch');
  else pass(`announcement_body from DB config`);
  if (ann.announcement_audience !== 'ALL') fail('announcement_audience mismatch');
  else pass(`announcement_audience=ALL schedule=${ann.announcement_start_at}..${ann.announcement_end_at} priority=${ann.announcement_priority}`);

  // --- 4. Test B: announcement admin editability (title suffix + restore) ---
  console.log('\n[4] Test B — Announcement admin editability (title suffix + restore)');
  const origTitle = ann.announcement_title;
  await patch('odyssey_system_config?key=eq.announcement_title', { value: origTitle + ' [uji]' });
  const edited = await get('odyssey_system_config?key=eq.announcement_title&select=key,value');
  if (edited[0]?.value !== origTitle + ' [uji]') fail('Edited announcement title not reflected on re-read');
  else pass(`key announcement_title edit reflected (member banner reads same key)`);
  await patch('odyssey_system_config?key=eq.announcement_title', { value: origTitle });
  const restored = await get('odyssey_system_config?key=eq.announcement_title&select=key,value');
  if (restored[0]?.value !== origTitle) fail('Announcement title restore failed');
  else pass(`key announcement_title restored to production value`);

  // --- 5. Legacy placeholders 535/536 stay deactivated; 608-620 untouched (checked separately) ---
  console.log('\n[5] Legacy placeholders 535/536 deactivated');
  const legacy = await get('odyssey_tasks?id=in.(535,536)&select=id,is_active');
  if (legacy.some(t => t.is_active)) fail(`Placeholders must be inactive: ${JSON.stringify(legacy)}`);
  else pass(`IDs 535,536 inactive`);

  console.log('\n====================================================');
  console.log(`E2E CHECK 2026-09-12: ${allOk ? 'ALL PASS 🟢' : 'FAILURES DETECTED 🔴'}`);
  console.log('====================================================');
  if (!allOk) process.exit(1);
}

run().catch(err => {
  console.error(err);
  process.exit(1);
});
