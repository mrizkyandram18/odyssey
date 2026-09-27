require('dotenv').config();

// Real authenticated production smoke for 2026-09-16 tasks (636/637/638).
// Strategy: disposable test user -> real flows -> exact-ID manifest ->
// exact-ID cleanup -> post-cleanup verification. No broad deletes.
// No redeem/claims. No real email is ever sent: task 638's flow is
// "user sends from their own mail client, reports evidence"; the smoke
// only submits clearly-marked TEST evidence text (recipient = self
// placeholder, no third party contacted).
// BOUNDARY: all three tasks are ADMIN_REVIEW, so PENDING submissions
// create zero ledger rows. Nothing is approved (no rewards granted).

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
const TEST_EXPLORER = `Smoke Sep16 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const T1 = 636; // Bikin CV Sederhana (DOCUMENT_UPLOAD)
const T2 = 637; // Cari Lowongan yang Cocok (TEXT_RESPONSE)
const T3 = 638; // Kirim CV untuk Melamar (TEXT_RESPONSE)

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
let failures = 0;
const manifest = { user: null, session: false, submissions: {}, storage: [], ledger: [], claims: [] };

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
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-09-16 (636/637/638)');
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

  // --- 3. Open tasks 636/637/638 (DB-rendered instructions) ---
  console.log('\n[3] Open task details (must be DB-rendered)');
  await sleep(3000);
  const t1 = await prodApi(`/api/tasks/${T1}`, { token });
  const t2 = await prodApi(`/api/tasks/${T2}`, { token });
  const t3 = await prodApi(`/api/tasks/${T3}`, { token });
  if (t1.status !== 200) bad(`open ${T1} failed: ${t1.status}`);
  else {
    const c = t1.json?.config || {};
    if (t1.json?.task_type === 'DOCUMENT_UPLOAD' && Array.isArray(c.allowed_extensions) && c.max_file_size_mb === 4) ok(`${T1} DOCUMENT_UPLOAD ext=${c.allowed_extensions.join(',')} max=4MB (DB config)`);
    else bad(`${T1} config mismatch: ${JSON.stringify(c).slice(0, 160)}`);
    if ((t1.json?.reward_coins) === 40 && (t1.json?.reward_xp) === 100) ok(`${T1} reward 40c/100xp (DB)`);
    else bad(`${T1} reward mismatch`);
  }
  if (t2.status !== 200) bad(`open ${T2} failed: ${t2.status}`);
  else {
    const c = t2.json?.config || {};
    if (t2.json?.task_type === 'TEXT_RESPONSE' && c.minimum_characters === 120 && c.maximum_characters === 2000 && /Link lowongan/.test(c.prompt || '')) ok(`${T2} TEXT bounds 120/2000 + link prompt (DB config)`);
    else bad(`${T2} config mismatch: ${JSON.stringify(c).slice(0, 200)}`);
  }
  if (t3.status !== 200) bad(`open ${T3} failed: ${t3.status}`);
  else {
    const c = t3.json?.config || {};
    if (t3.json?.task_type === 'TEXT_RESPONSE' && c.minimum_characters === 120 && /Subject email/.test(c.prompt || '')) ok(`${T3} TEXT email-evidence prompt (DB config)`);
    else bad(`${T3} config mismatch: ${JSON.stringify(c).slice(0, 200)}`);
  }

  // --- 4. Task 636: upload disposable CV + PENDING submit ---
  console.log(`\n[4] Task ${T1} CV document upload + submit (PENDING, no reward)`);
  const cvText = `SMOKE-TEST CV (disposable, will be deleted)\nNama: ${TEST_EXPLORER}\nKontak: 0800-SMOKE-TEST / smoke@example.invalid\nRingkasan: Data uji asap produksi, bukan CV asli.\nPendidikan: SMA Uji Asap\nPengalaman: Organisasi Smoke Testing\nSkill: testing, cleanup`;
  const form = new FormData();
  form.append('file', new Blob([cvText], { type: 'text/plain' }), 'CV-SMOKE-TEST.txt');
  const up = await prodApi('/api/tasks/upload', { method: 'POST', token, form });
  if (up.status !== 200 || !up.json?.file_url) { bad(`upload failed: ${up.status} ${up.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up.json.storage_path);
    ok(`uploaded storage_path = ${up.json.storage_path}`);
    // Verify file actually exists in storage (admin-readable)
    const chk = await fetch(up.json.file_url, { method: 'HEAD' });
    if (chk.ok) ok(`storage file reachable via file_url (HEAD ${chk.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk.status})`);
    await sleep(3000);
    const s1 = await prodApi(`/api/tasks/${T1}/submit`, { method: 'POST', token, body: { payload: { file_url: up.json.file_url, file_name: up.json.file_name, file_size: up.json.file_size, note: 'SMOKE-TEST: CV uji asap, mohon abaikan', submitted_at: new Date().toISOString() } } });
    if (s1.status !== 200 || s1.json?.success !== true) bad(`${T1} submit failed: ${s1.status} ${s1.text.slice(0, 160)}`);
    else ok(`${T1} submit accepted (manual review path)`);
    const rows = await dbGet(`odyssey_task_submissions?task_id=eq.${T1}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows.length !== 1 || rows[0].status !== 'PENDING') bad(`${T1} submission state wrong: ${JSON.stringify(rows).slice(0, 300)}`);
    else {
      manifest.submissions[T1] = rows[0].id;
      ok(`submission_id = ${rows[0].id} (PENDING, admin-readable payload file_url present: ${!!rows[0].payload?.file_url})`);
    }
  }

  // --- 5. Task 637: job-search TEXT evidence (harmless, clearly marked smoke) ---
  console.log(`\n[5] Task ${T2} job-search evidence submit (PENDING)`);
  const jobText = 'SMOKE-TEST DATA (bukan lamaran asli, akan dihapus). 1) Posisi yang dilamar: Asisten Toko (CONTOH UJI). 2) Nama perusahaan: Toko Contoh Smoke-Invalid (FIKTIF). 3) Link lowongan: https://example.invalid/lowongan/smoke-test-123. 4) Sumber lowongan: example.invalid (URL uji). 5) Kenapa cocok: contoh alasan uji — sesuai minat pelayanan pelanggan, dekat lokasi contoh, dan jam kerja contoh yang fleksibel.';
  console.log(`       evidence len = ${jobText.length} (bounds 120/2000)`);
  await sleep(3000);
  const s2 = await prodApi(`/api/tasks/${T2}/submit`, { method: 'POST', token, body: { payload: { text: jobText, submitted_at: new Date().toISOString() } } });
  if (s2.status !== 200 || s2.json?.success !== true) bad(`${T2} submit failed: ${s2.status} ${s2.text.slice(0, 160)}`);
  else ok(`${T2} submit accepted (manual review path)`);
  const rows2 = await dbGet(`odyssey_task_submissions?task_id=eq.${T2}&user_uid=eq.${TEST_UID}&select=id,status`);
  if (rows2.length !== 1 || rows2[0].status !== 'PENDING') bad(`${T2} submission state wrong: ${JSON.stringify(rows2)}`);
  else { manifest.submissions[T2] = rows2[0].id; ok(`submission_id = ${rows2[0].id} (PENDING)`); }

  // --- 6. Task 638: email evidence (NO real email sent — evidence text only) ---
  console.log(`\n[6] Task ${T3} email-application evidence submit (PENDING; no email sent by app)`);
  const mailText = 'SMOKE-TEST DATA (tidak ada email yang dikirim, akan dihapus). 1) Email penerima: diri-sendiri-smoke@example.invalid (latihan ke alamat sendiri). 2) Subject: Lamaran — Asisten Toko — Smoke Test (CONTOH). 3) Isi: Yth. Bapak/Ibu, perkenalkan saya contoh uji asap. Saya bermaksud melamar posisi Asisten Toko contoh. Terima kasih atas perhatiannya. Hormat saya, Smoke. 4) File CV dilampirkan: CV-SMOKE-TEST.txt. 5) Tanggal/jam: uji asap ini. 6) Pernyataan: contoh cek ulang — penerima benar (diri sendiri), subject jelas, body sopan, CV terlampir.';
  console.log(`       evidence len = ${mailText.length} (bounds 120/2000)`);
  await sleep(3000);
  const s3 = await prodApi(`/api/tasks/${T3}/submit`, { method: 'POST', token, body: { payload: { text: mailText, submitted_at: new Date().toISOString() } } });
  if (s3.status !== 200 || s3.json?.success !== true) bad(`${T3} submit failed: ${s3.status} ${s3.text.slice(0, 160)}`);
  else ok(`${T3} submit accepted (manual review path; app sent nothing)`);
  const rows3 = await dbGet(`odyssey_task_submissions?task_id=eq.${T3}&user_uid=eq.${TEST_UID}&select=id,status`);
  if (rows3.length !== 1 || rows3[0].status !== 'PENDING') bad(`${T3} submission state wrong: ${JSON.stringify(rows3)}`);
  else { manifest.submissions[T3] = rows3[0].id; ok(`submission_id = ${rows3[0].id} (PENDING)`); }

  // --- 7. Enumerate all test-user artifacts ---
  console.log('\n[7] Artifact enumeration (ownership = disposable user only)');
  const allSubs = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id,task_id,status`);
  const ledger = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id`);
  const claims = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  manifest.ledger = ledger; manifest.claims = claims;
  console.log(`       submissions: ${JSON.stringify(allSubs)}`);
  console.log(`       ledger rows: ${ledger.length} | claims: ${claims.length}`);
  if (allSubs.length !== 3) bad(`expected exactly 3 submissions, found ${allSubs.length}`);
  else ok('exactly 3 submission artifacts (636, 637, 638 — all PENDING)');
  if (ledger.length || claims.length) bad('unexpected ledger/claim rows');
  else ok('no ledger/claim artifacts (nothing approved, nothing financial)');
  const prof = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid,coins,xp`);
  if (prof.length !== 1 || prof[0].coins !== 0 || prof[0].xp !== 0) bad(`profile state wrong: ${JSON.stringify(prof)}`);
  else ok('profile coins=0 xp=0 (no rewards granted)');

  // --- 8. Exact-ID cleanup ---
  console.log('\n[8] Exact-ID cleanup (only manifest artifacts)');
  for (const [task, sid] of Object.entries(manifest.submissions)) {
    const own = await dbGet(`odyssey_task_submissions?id=eq.${sid}&select=id,user_uid`);
    if (own.length !== 1 || own[0].user_uid !== TEST_UID) { bad(`ownership unproven for submission ${sid} — SKIPPING its delete`); continue; }
    await dbDel('odyssey_task_submissions', `id=eq.${sid}`);
    const gone = await dbGet(`odyssey_task_submissions?id=eq.${sid}&select=id`);
    if (gone.length) bad(`submission ${sid} still present`);
    else ok(`submission ${sid} (task ${task}) deleted + verified absent`);
  }
  for (const sp of manifest.storage) {
    // Remove-style endpoint (DELETE path form returns 400 on this project)
    const dr = await fetch(`${SUPABASE_URL}/storage/v1/object/task-proofs`, { method: 'DELETE', headers: dbHeaders, body: JSON.stringify({ prefixes: [sp] }) });
    if (!dr.ok) bad(`storage delete failed ${sp}: ${dr.status}`);
    else {
      const chk = await fetch(`${SUPABASE_URL}/storage/v1/object/list/task-proofs`, { method: 'POST', headers: dbHeaders, body: JSON.stringify({ prefix: sp, limit: 1 }) });
      const items = await chk.json();
      if (Array.isArray(items) && items.length === 0) ok(`storage object ${sp} deleted + verified absent (list empty)`);
      else bad(`storage object ${sp} still listed: ${JSON.stringify(items).slice(0, 160)}`);
    }
  }
  if (manifest.user) {
    await dbDel('odyssey_user_profiles', `uid=eq.${TEST_UID}`);
    const gone = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
    if (gone.length) bad('test profile still present');
    else ok(`profile ${TEST_UID} deleted + verified absent`);
  }

  // --- 9. Post-cleanup verification + unchanged production state ---
  console.log('\n[9] Post-cleanup verification');
  const r1 = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id`);
  const r2 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const r3 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  const r4 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id`);
  const r5 = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  if (r1.length || r2.length || r3.length || r4.length || r5.length) bad(`residual artifacts: subs=${r1.length} uid=${r2.length} uname=${r3.length} ledger=${r4.length} claims=${r5.length}`);
  else ok('zero remaining test artifacts (submissions/profile/ledger/claims)');
  const t916 = await dbGet('odyssey_tasks?active_date=eq.2026-09-16&is_active=eq.true&select=id&order=id.asc');
  if (JSON.stringify(t916.map((t) => t.id)) !== '[636,637,638]') bad(`09-16 tasks changed: ${JSON.stringify(t916)}`);
  else ok('tasks 636/637/638 intact (exactly 3 active on 2026-09-16)');

  console.log('\n====================================================');
  console.log('ARTIFACT MANIFEST: ' + JSON.stringify(manifest));
  console.log(`SMOKE RESULT: ${failures === 0 ? 'ALL PASS 🟢' : 'FAILURES: ' + failures + ' 🔴'}`);
  console.log('====================================================');
  if (failures) process.exit(1);
}

run().catch((e) => { console.error(e); process.exit(1); });
