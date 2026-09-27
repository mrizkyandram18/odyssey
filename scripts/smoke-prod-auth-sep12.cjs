require('dotenv').config();

// Real authenticated production smoke for 2026-09-12 tasks + payout announcement.
// Strategy: disposable test user -> real flows -> exact-ID manifest -> exact-ID
// cleanup -> post-cleanup verification. No broad deletes. No redeem/claims.
// BOUNDARY (documented, evidence-backed): task 624 is NEVER submitted with
// valid answers because AUTO approval writes an odyssey_coin_transactions row
// protected by an IMMUTABLE trigger (UPDATE/DELETE prohibited) which would also
// FK-block profile deletion. Instead 624 is exercised with INVALID choices to
// prove server-side config-driven validation (400, zero artifacts).

const PROD = 'https://odyssey-beta-nine.vercel.app';
const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const dbHeaders = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
};

const ts = Date.now();
const TEST_UID = `smoke-uid-${ts}`;
const TEST_USER = `odyssey-smoke-${ts}`;
const TEST_EXPLORER = `Smoke Sep12 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const ADMIN_DEVICE = `smoke-admin-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
let failures = 0;
const manifest = { user: null, session: false, submissions: {}, storage: [], tickets: [], ledger: [], claims: [], admin: {} };

function ok(msg) { console.log(`   🟢 ${msg}`); }
function bad(msg) { console.log(`   🔴 ${msg}`); failures++; }

async function dbGet(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers: dbHeaders });
  if (!r.ok) throw new Error(`DB GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}
async function dbDel(table, filter) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${table}?${filter}`, { method: 'DELETE', headers: dbHeaders });
  if (!r.ok) throw new Error(`DB DELETE ${table}?${filter}: ${r.status} ${await r.text()}`);
  return true;
}
async function prodApi(path, { method = 'GET', token = null, body = null, form = null } = {}) {
  const headers = {};
  if (token) { headers.Authorization = `Bearer ${token}`; headers.Cookie = `odyssey_session=${token}`; }
  let payload;
  if (form) { payload = { method, headers, body: form }; }
  else { headers['Content-Type'] = 'application/json'; payload = { method, headers }; if (body) payload.body = JSON.stringify(body); }
  const r = await fetch(PROD + path, payload);
  const text = await r.text();
  let json = null;
  try { json = JSON.parse(text); } catch { /* non-json */ }
  return { status: r.status, json, text };
}

async function run() {
  console.log('====================================================');
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-09-12 + announcement');
  console.log(` disposable user: ${TEST_USER} / ${TEST_UID}`);
  console.log('====================================================\n');

  // --- 0. Collision check ---
  console.log('[0] Collision check (must be empty)');
  const c1 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const c2 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  if (c1.length || c2.length) { bad('test identity collides with existing user — ABORT'); process.exit(1); }
  ok('no collision for disposable identity');

  // --- 1. Create disposable test user ---
  console.log('\n[1] Create disposable test user');
  let r = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
    method: 'POST', headers: { ...dbHeaders, Prefer: 'return=representation' },
    body: JSON.stringify([{ uid: TEST_UID, username: TEST_USER, password_hash: PASS_HASH, explorer_name: TEST_EXPLORER, role: 'MEMBER', family_id: 'demo-crew-1', coins: 0, xp: 0, level: 1, is_active: true }]),
  });
  if (!r.ok) { bad(`create user failed: ${r.status} ${await r.text()}`); process.exit(1); }
  manifest.user = TEST_UID;
  ok(`TEST USER user_id = ${TEST_UID}`);

  // --- 2. Member login ---
  console.log('\n[2] Authenticated member login');
  await sleep(6000);
  let login = await prodApi('/api/login', { method: 'POST', body: { uid: TEST_USER, login_method: 'PASSWORD', credential: 'admin123', device: { device_id: DEVICE } } });
  if (login.status !== 200 || !login.json?.session) { bad(`member login failed: ${login.status} ${login.text.slice(0, 200)}`); process.exit(1); }
  const token = login.json.session;
  manifest.session = true;
  ok(`login ok (token ${String(token).slice(0, 8)}…, len ${String(token).length})`);

  // --- 3. Announcement via member config path ---
  console.log('\n[3] Announcement from backend (/api/shop/config)');
  await sleep(3000);
  const cfg = await prodApi('/api/shop/config', { token });
  if (cfg.status !== 200) { bad(`shop config failed: ${cfg.status}`); }
  else {
    const a = cfg.json?.announcement;
    if (!a?.visible) bad(`announcement not visible: ${JSON.stringify(a)}`);
    else ok(`announcement visible from backend config`);
    if (a?.title !== 'Pengumuman Pencairan') bad(`title mismatch: ${JSON.stringify(a?.title)}`);
    else ok(`title = backend-configured value`);
    if (typeof a?.body !== 'string' || !a.body.includes('penyesuaian jadwal')) bad('body mismatch');
    else ok(`body = backend-configured value (len ${a.body.length})`);
    console.log(`       audience=${a?.audience} priority=${a?.priority} window=${a?.start_at}..${a?.end_at}`);
  }

  // --- 4. Task listing ---
  console.log('\n[4] Task listing (/api/tasks/today)');
  const today = await prodApi('/api/tasks/today', { token });
  if (today.status !== 200) bad(`tasks/today failed: ${today.status}`);
  else {
    const ids = (today.json?.tasks || []).map((t) => t.id);
    for (const id of [624, 625, 626]) {
      if (ids.includes(id)) ok(`task ${id} listed for member`);
      else bad(`task ${id} MISSING from member listing`);
    }
  }

  // --- 5. Open tasks 624/625/626 ---
  console.log('\n[5] Open task details');
  const t624 = await prodApi('/api/tasks/624', { token });
  const t625 = await prodApi('/api/tasks/625', { token });
  const t626 = await prodApi('/api/tasks/626', { token });
  if (t624.status !== 200) bad('open 624 failed');
  else {
    const n = t624.json?.config?.scenario?.events?.length;
    if (n === 6) ok('624 MINI_GAME: 6 scenarios from DB config');
    else bad(`624 scenario events = ${n}`);
  }
  if (t625.status !== 200) bad('open 625 failed');
  else ok(`625 DOCUMENT_UPLOAD config ext=${(t625.json?.config?.allowed_extensions || []).join(',')}`);
  if (t626.status !== 200) bad('open 626 failed');
  else {
    const c = t626.json?.config || {};
    if (c.minimum_characters === 80 && c.maximum_characters === 1000) ok('626 TEXT bounds 80/1000 from DB config');
    else bad(`626 bounds = ${c.minimum_characters}/${c.maximum_characters}`);
  }

  // --- 6. Task 624 INVALID submit -> 400, zero artifacts (ledger-immutability boundary) ---
  console.log('\n[6] Task 624 invalid submit (proves server-side config validation, zero writes)');
  const events = t624.json?.config?.scenario?.events || [];
  const badChoices = {};
  for (const e of events) badChoices[e.id] = '__invalid__';
  const bad624 = await prodApi('/api/tasks/624/submit', { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: badChoices } } });
  if (bad624.status !== 400) bad(`expected 400, got ${bad624.status}: ${bad624.text.slice(0, 160)}`);
  else ok(`624 invalid choices rejected 400 (${(bad624.json?.error || '').slice(0, 80)})`);
  const subCheck624 = await dbGet(`odyssey_task_submissions?task_id=eq.624&user_uid=eq.${TEST_UID}&select=id`);
  if (subCheck624.length) bad(`624 wrote ${subCheck624.length} submission rows!`);
  else ok('624: zero submission rows (nothing persisted)');

  // --- 7. Task 625: upload + PENDING submit ---
  console.log('\n[7] Task 625 document upload + submit (PENDING, no reward)');
  const fileText = `Rencana mingguan ${TEST_USER}: tujuan, langkah kecil, jadwal harian. Dokumen uji asap produksi, akan dihapus.`;
  const form = new FormData();
  form.append('file', new Blob([fileText], { type: 'text/plain' }), 'rencana-sederhana.txt');
  const up = await prodApi('/api/tasks/upload', { method: 'POST', token, form });
  if (up.status !== 200 || !up.json?.file_url) { bad(`upload failed: ${up.status} ${up.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up.json.storage_path);
    ok(`uploaded storage_path = ${up.json.storage_path}`);
    await sleep(3000);
    const s625 = await prodApi('/api/tasks/625/submit', { method: 'POST', token, body: { payload: { file_url: up.json.file_url, file_name: up.json.file_name, file_size: up.json.file_size, note: 'uji asap', submitted_at: new Date().toISOString() } } });
    if (s625.status !== 200 || s625.json?.success !== true) bad(`625 submit failed: ${s625.status} ${s625.text.slice(0, 160)}`);
    else ok(`625 submit accepted (manual review path)`);
    const rows = await dbGet(`odyssey_task_submissions?task_id=eq.625&user_uid=eq.${TEST_UID}&select=id,status`);
    if (rows.length !== 1 || rows[0].status !== 'PENDING') bad(`625 submission state wrong: ${JSON.stringify(rows)}`);
    else { manifest.submissions[625] = rows[0].id; ok(`625 submission_id = ${rows[0].id} (PENDING, coins 0 — no reward)`); }
  }

  // --- 8. Task 626: TEXT PENDING submit ---
  console.log('\n[8] Task 626 text submit (PENDING, no reward)');
  const essay = 'Rencanaku minggu ini berubah karena ada tugas mendadak dari atasan. Dampaknya jadwalku bergeser dua hari. Penyesuaiannya: saya susun ulang prioritas dan komunikasikan perubahan ke tim. Yang dihindari: panik dan menyalahkan keadaan. Pelajarannya: fleksibilitas adalah bagian dari rencana.';
  console.log(`       essay len = ${essay.length} (bounds 80/1000)`);
  await sleep(3000);
  const s626 = await prodApi('/api/tasks/626/submit', { method: 'POST', token, body: { payload: { text: essay, submitted_at: new Date().toISOString() } } });
  if (s626.status !== 200 || s626.json?.success !== true) bad(`626 submit failed: ${s626.status} ${s626.text.slice(0, 160)}`);
  else ok('626 submit accepted (manual review path)');
  const rows626 = await dbGet(`odyssey_task_submissions?task_id=eq.626&user_uid=eq.${TEST_UID}&select=id,status`);
  if (rows626.length !== 1 || rows626[0].status !== 'PENDING') bad(`626 submission state wrong: ${JSON.stringify(rows626)}`);
  else { manifest.submissions[626] = rows626[0].id; ok(`626 submission_id = ${rows626[0].id} (PENDING)`); }

  // --- 9. Enumerate all test-user artifacts (expect: 2 submissions, nothing else) ---
  console.log('\n[9] Artifact enumeration (ownership = disposable user only)');
  const allSubs = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id,task_id,status`);
  const ledger = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id`);
  const tickets = await dbGet(`reward_tickets?user_uid=eq.${TEST_UID}&select=user_uid,ticket_date,status`);
  const claims = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  manifest.tickets = tickets; manifest.ledger = ledger; manifest.claims = claims;
  console.log(`       submissions: ${JSON.stringify(allSubs)}`);
  console.log(`       ledger rows: ${ledger.length} | tickets: ${tickets.length} | claims: ${claims.length}`);
  if (allSubs.length !== 2) bad(`expected exactly 2 submissions, found ${allSubs.length}`);
  else ok('exactly 2 submission artifacts (625, 626)');
  if (ledger.length || tickets.length || claims.length) bad('unexpected ledger/ticket/claim rows');
  else ok('no ledger/ticket/claim artifacts (nothing financial created)');
  const prof = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid,coins,xp`);
  if (prof.length !== 1 || prof[0].coins !== 0) bad(`profile state wrong: ${JSON.stringify(prof)}`);
  else ok('profile coins=0 (no rewards granted)');

  // --- 10. Admin config path (read-only check via admin session, exact restore) ---
  console.log('\n[10] Admin announcement config path');
  const admins = await dbGet(`odyssey_user_profiles?username=eq.admin&select=uid,device_id,device_bound_at`);
  if (!admins.length) { bad('admin user not found — skipping admin path (boundary)'); }
  else {
    const adminUid = admins[0].uid;
    const origDevice = admins[0].device_id; const origBound = admins[0].device_bound_at;
    manifest.admin = { uid: adminUid, origDevice, origBound };
    console.log(`       admin uid=${adminUid} device=${origDevice}`);
    await fetch(`${SUPABASE_URL}/rest/v1/rpc/odyssey_admin_reset_device`, { method: 'POST', headers: dbHeaders, body: JSON.stringify({ p_target_uid: adminUid }) });
    await sleep(15000);
    const alogin = await prodApi('/api/login', { method: 'POST', body: { uid: 'admin', login_method: 'PASSWORD', credential: 'admin123', device: { device_id: ADMIN_DEVICE } } });
    if (alogin.status !== 200 || !alogin.json?.session) { bad(`admin login failed: ${alogin.status} — restoring device binding`); }
    else {
      ok('admin login ok');
      await sleep(3000);
      const acfg = await prodApi('/api/admin/config', { token: alogin.json.session });
      if (acfg.status !== 200) bad(`admin config failed: ${acfg.status}`);
      else {
        const a = acfg.json?.announcement;
        if (a?.title === 'Pengumuman Pencairan' && a?.visible === true) ok('admin config loads announcement (editable title, visible=true)');
        else bad(`admin announcement unexpected: ${JSON.stringify(a)}`);
      }
    }
    // Exact restore of admin device binding (pre-read values)
    const restore = {};
    if (origDevice === null) restore.device_id = null; else restore.device_id = origDevice;
    if (origBound === null) restore.device_bound_at = null; else restore.device_bound_at = origBound;
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${adminUid}`, { method: 'PATCH', headers: dbHeaders, body: JSON.stringify(restore) });
    const after = await dbGet(`odyssey_user_profiles?uid=eq.${adminUid}&select=device_id,device_bound_at`);
    if (String(after[0]?.device_id) === String(origDevice) && String(after[0]?.device_bound_at) === String(origBound)) ok(`admin device binding restored exactly (${origDevice})`);
    else bad(`admin restore mismatch: ${JSON.stringify(after[0])} vs orig ${origDevice}/${origBound}`);
  }

  // --- 11. Exact-ID cleanup ---
  console.log('\n[11] Exact-ID cleanup (only manifest artifacts)');
  for (const [task, sid] of Object.entries(manifest.submissions)) {
    const own = await dbGet(`odyssey_task_submissions?id=eq.${sid}&select=id,user_uid`);
    if (own.length !== 1 || own[0].user_uid !== TEST_UID) { bad(`ownership unproven for submission ${sid} — SKIPPING its delete`); continue; }
    await dbDel('odyssey_task_submissions', `id=eq.${sid}`);
    const gone = await dbGet(`odyssey_task_submissions?id=eq.${sid}&select=id`);
    if (gone.length) bad(`submission ${sid} still present`);
    else ok(`submission ${sid} (task ${task}) deleted + verified absent`);
  }
  for (const sp of manifest.storage) {
    const dr = await fetch(`${SUPABASE_URL}/storage/v1/object/task-proofs/${sp}`, { method: 'DELETE', headers: dbHeaders });
    if (!dr.ok) bad(`storage delete failed ${sp}: ${dr.status}`);
    else {
      const chk = await fetch(`${SUPABASE_URL}/storage/v1/object/task-proofs/${sp}`, { headers: dbHeaders });
      if (chk.status === 404 || chk.status === 400) ok(`storage object ${sp} deleted + verified absent (${chk.status})`);
      else bad(`storage object ${sp} still reachable (${chk.status})`);
    }
  }
  if (manifest.user) {
    await dbDel('odyssey_user_profiles', `uid=eq.${TEST_UID}`);
    const gone = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
    if (gone.length) bad('test profile still present');
    else ok(`profile ${TEST_UID} deleted + verified absent`);
  }

  // --- 12. Post-cleanup verification + unchanged production state ---
  console.log('\n[12] Post-cleanup verification');
  const r1 = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id`);
  const r2 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const r3 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  if (r1.length || r2.length || r3.length) bad(`residual artifacts: subs=${r1.length} uid=${r2.length} uname=${r3.length}`);
  else ok('zero remaining test artifacts (submissions/profiles)');
  const t912 = await dbGet('odyssey_tasks?active_date=eq.2026-09-12&is_active=eq.true&select=id&order=id.asc');
  if (JSON.stringify(t912.map((t) => t.id)) !== '[624,625,626]') bad(`09-12 tasks changed: ${JSON.stringify(t912)}`);
  else ok('tasks 624/625/626 intact');
  const ann = await dbGet('odyssey_system_config?key=in.(announcement_enabled,announcement_title)&select=key,value&order=key.asc');
  const annOk = ann.find((r) => r.key === 'announcement_enabled')?.value === 'true' && ann.find((r) => r.key === 'announcement_title')?.value === 'Pengumuman Pencairan';
  if (!annOk) bad(`announcement changed: ${JSON.stringify(ann)}`);
  else ok('announcement config unchanged');
  const pay = await dbGet('odyssey_system_config?key=in.(redemption_start_day,redemption_end_day,max_payout_coins)&select=key,value');
  console.log(`       payout keys: ${JSON.stringify(pay)} (read-only confirm)`);

  console.log('\n====================================================');
  console.log('ARTIFACT MANIFEST: ' + JSON.stringify(manifest));
  console.log(`SMOKE RESULT: ${failures === 0 ? 'ALL PASS 🟢' : 'FAILURES: ' + failures + ' 🔴'}`);
  console.log('====================================================');
  if (failures) process.exit(1);
}

run().catch((e) => { console.error(e); process.exit(1); });
