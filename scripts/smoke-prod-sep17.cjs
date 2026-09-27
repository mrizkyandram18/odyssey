require('dotenv').config();

// Real authenticated production smoke for 2026-09-17 tasks (639/640/641).
// Strategy: disposable test user -> real flows -> exact-ID manifest ->
// exact-ID cleanup -> post-cleanup verification. No broad deletes.
// No redeem/claims. No purchase is made: task 640 evidence is fictional
// clearly-marked test data. No real-world problem is altered (639 photo
// is a disposable synthetic test image).
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
const TEST_EXPLORER = `Smoke Sep17 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const T1 = 639; // Benerin Sesuatu yang Sering Bikin Ribet (PHOTO_UPLOAD)
const T2 = 640; // Bandingkan Sebelum Beli (TEXT_RESPONSE, bounds 200/2000)
const T3 = 641; // Jelaskan Sampai Orang Lain Paham (VIDEO, max 60s)

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
let failures = 0;
const manifest = { user: null, session: false, submissions: {}, storage: [], ledger: [], claims: [] };

function ok(msg) { console.log(`   PASS ${msg}`); }
function bad(msg) { console.log(`   FAIL ${msg}`); failures++; }

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
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-09-17 (639/640/641)');
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

  // --- 3. Open tasks 639/640/641 (DB-rendered instructions) ---
  console.log('\n[3] Open task details (must be DB-rendered)');
  await sleep(3000);
  const t1 = await prodApi(`/api/tasks/${T1}`, { token });
  const t2 = await prodApi(`/api/tasks/${T2}`, { token });
  const t3 = await prodApi(`/api/tasks/${T3}`, { token });
  if (t1.status !== 200) bad(`open ${T1} failed: ${t1.status}`);
  else {
    const c = t1.json?.config || {};
    if (t1.json?.task_type === 'PHOTO_UPLOAD' && c.max_files === 1 && c.camera_only === false && /membuatnya ribet/.test(c.instruction || '')) ok(`${T1} PHOTO max_files=1 camera_only=false + 3-part note instruction (DB config)`);
    else bad(`${T1} config mismatch: ${JSON.stringify(c).slice(0, 200)}`);
    if (t1.json?.reward_coins === 40 && t1.json?.reward_xp === 100) ok(`${T1} reward 40c/100xp (DB)`);
    else bad(`${T1} reward mismatch`);
  }
  if (t2.status !== 200) bad(`open ${T2} failed: ${t2.status}`);
  else {
    const c = t2.json?.config || {};
    if (t2.json?.task_type === 'TEXT_RESPONSE' && c.minimum_characters === 200 && c.maximum_characters === 2000 && /Pilihan akhirmu/.test(c.prompt || '')) ok(`${T2} TEXT bounds 200/2000 + comparison prompt (DB config)`);
    else bad(`${T2} config mismatch: ${JSON.stringify(c).slice(0, 200)}`);
    if (t2.json?.reward_coins === 40 && t2.json?.reward_xp === 100) ok(`${T2} reward 40c/100xp (DB)`);
    else bad(`${T2} reward mismatch`);
  }
  if (t3.status !== 200) bad(`open ${T3} failed: ${t3.status}`);
  else {
    const c = t3.json?.config || {};
    if (t3.json?.task_type === 'VIDEO' && c.recording?.enabled === true && c.recording?.max_duration_seconds === 60) ok(`${T3} VIDEO recording enabled max 60s (DB config)`);
    else bad(`${T3} config mismatch: ${JSON.stringify(c).slice(0, 200)}`);
    if (t3.json?.reward_coins === 40 && t3.json?.reward_xp === 100) ok(`${T3} reward 40c/100xp (DB)`);
    else bad(`${T3} reward mismatch`);
  }

  // --- 4. Task 639: disposable test image + 3-part note (PENDING) ---
  console.log(`\n[4] Task ${T1} photo upload + submit (PENDING, no reward)`);
  const pngB64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';
  const pngBytes = Buffer.from(pngB64, 'base64');
  const form1 = new FormData();
  form1.append('file', new Blob([pngBytes], { type: 'image/png' }), 'SMOKE-TEST-ribet.png');
  const up1 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form1 });
  if (up1.status !== 200 || !up1.json?.file_url) { bad(`upload failed: ${up1.status} ${up1.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up1.json.storage_path);
    ok(`uploaded storage_path = ${up1.json.storage_path}`);
    const chk = await fetch(up1.json.file_url, { method: 'HEAD' });
    if (chk.ok) ok(`storage file reachable via file_url (HEAD ${chk.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk.status})`);
    await sleep(3000);
    const note = 'SMOKE-TEST (data uji, akan dihapus). (a) Yang bikin ribet: gambar uji ini mensimulasikan kabel contoh yang berbelit. (b) Kenapa masalah: contoh alasan uji — sulit menemukan ujung yang benar saat terburu-buru. (c) Ide sederhana: contoh ide uji — beri label warna di setiap ujung kabel contoh.';
    const s1 = await prodApi(`/api/tasks/${T1}/submit`, { method: 'POST', token, body: { payload: { file_url: up1.json.file_url, file_name: up1.json.file_name, file_size: up1.json.file_size, note, submitted_at: new Date().toISOString() } } });
    if (s1.status !== 200 || s1.json?.success !== true) bad(`${T1} submit failed: ${s1.status} ${s1.text.slice(0, 160)}`);
    else ok(`${T1} submit accepted (manual review path)`);
    const rows = await dbGet(`odyssey_task_submissions?task_id=eq.${T1}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows.length !== 1 || rows[0].status !== 'PENDING') bad(`${T1} submission state wrong: ${JSON.stringify(rows).slice(0, 300)}`);
    else {
      manifest.submissions[T1] = rows[0].id;
      ok(`submission_id = ${rows[0].id} (PENDING, file_url present: ${!!rows[0].payload?.file_url}, note present: ${!!rows[0].payload?.note})`);
    }
  }

  // --- 5. Task 640: server validation + fictional comparison submit (PENDING) ---
  console.log(`\n[5] Task ${T2} comparison evidence (validation + PENDING submit)`);
  await sleep(3000);
  const short = await prodApi(`/api/tasks/${T2}/submit`, { method: 'POST', token, body: { payload: { text: 'SMOKE-TEST pendek', submitted_at: new Date().toISOString() } } });
  if (short.status === 200 && short.json?.success === true) bad(`${T2} accepted too-short text — min-length validation BYPASSED`);
  else if (/minimal 200/.test(short.text)) ok(`${T2} server rejected short text with min-200 message (DB-configured validation)`);
  else bad(`${T2} short-text rejection unclear: ${short.status} ${short.text.slice(0, 160)}`);
  const cmpText = 'SMOKE-TEST DATA (fiktif, bukan pembelian asli, akan dihapus). 1) Nama barang: Charger HP contoh (CONTOH UJI). 2) Pilihan 1 — Charger Contoh A, Rp50.000 contoh, 20W + kabel contoh, kelebihan: murah contoh, kekurangan: lambat contoh. 3) Pilihan 2 — Charger Contoh B, Rp120.000 contoh, 65W + garansi contoh, kelebihan: cepat contoh, kekurangan: mahal contoh. 5) Biaya tambahan: ongkir contoh Rp10.000. 6) Pilihan akhir: Charger Contoh A (CONTOH). 7) Alasan: contoh alasan uji — sesuai kebutuhan contoh yang ringan dan bujet contoh yang terbatas.';
  console.log(`       evidence len = ${cmpText.length} (bounds 200/2000)`);
  await sleep(3000);
  const s2 = await prodApi(`/api/tasks/${T2}/submit`, { method: 'POST', token, body: { payload: { text: cmpText, submitted_at: new Date().toISOString() } } });
  if (s2.status !== 200 || s2.json?.success !== true) bad(`${T2} submit failed: ${s2.status} ${s2.text.slice(0, 160)}`);
  else ok(`${T2} submit accepted (manual review path)`);
  const rows2 = await dbGet(`odyssey_task_submissions?task_id=eq.${T2}&user_uid=eq.${TEST_UID}&select=id,status`);
  if (rows2.length !== 1 || rows2[0].status !== 'PENDING') bad(`${T2} submission state wrong: ${JSON.stringify(rows2)}`);
  else { manifest.submissions[T2] = rows2[0].id; ok(`submission_id = ${rows2[0].id} (PENDING)`); }

  // --- 6. Task 641: disposable test video upload + submit (PENDING) ---
  console.log(`\n[6] Task ${T3} video upload + submit (PENDING; short disposable clip)`);
  // Minimal synthetic MP4 (ftyp box only — upload infra accepts it; admin review
  // path stores file_url, duration is client-reported and admin-verified).
  const ftyp = Buffer.from([
    0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6f, 0x6d,
    0x00, 0x00, 0x00, 0x00, 0x69, 0x73, 0x6f, 0x6d, 0x69, 0x73, 0x6f, 0x32,
  ]);
  const form3 = new FormData();
  form3.append('file', new Blob([ftyp], { type: 'video/mp4' }), 'SMOKE-TEST-jelasin.mp4');
  const up3 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form3 });
  if (up3.status !== 200 || !up3.json?.file_url) { bad(`video upload failed: ${up3.status} ${up3.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up3.json.storage_path);
    ok(`uploaded storage_path = ${up3.json.storage_path}`);
    const chk3 = await fetch(up3.json.file_url, { method: 'HEAD' });
    if (chk3.ok) ok(`storage file reachable via file_url (HEAD ${chk3.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk3.status})`);
    await sleep(3000);
    const s3 = await prodApi(`/api/tasks/${T3}/submit`, {
      method: 'POST', token,
      body: { payload: { file_url: up3.json.file_url, file_name: up3.json.file_name, file_size: up3.json.file_size, duration_seconds: 12, mime_type: 'video/mp4', note: 'SMOKE-TEST: video uji disposable (bukan penjelasan asli), mohon abaikan', captured_at: new Date().toISOString() } },
    });
    if (s3.status !== 200 || s3.json?.success !== true) bad(`${T3} submit failed: ${s3.status} ${s3.text.slice(0, 160)}`);
    else ok(`${T3} submit accepted (manual review path)`);
    const rows3 = await dbGet(`odyssey_task_submissions?task_id=eq.${T3}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows3.length !== 1 || rows3[0].status !== 'PENDING') bad(`${T3} submission state wrong: ${JSON.stringify(rows3).slice(0, 300)}`);
    else { manifest.submissions[T3] = rows3[0].id; ok(`submission_id = ${rows3[0].id} (PENDING, file_url present: ${!!rows3[0].payload?.file_url})`); }
  }

  // --- 7. Enumerate all test-user artifacts ---
  console.log('\n[7] Artifact enumeration (ownership = disposable user only)');
  const allSubs = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id,task_id,status`);
  const ledger = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id`);
  const claims = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  manifest.ledger = ledger; manifest.claims = claims;
  console.log(`       submissions: ${JSON.stringify(allSubs)}`);
  console.log(`       ledger rows: ${ledger.length} | claims: ${claims.length}`);
  if (allSubs.length !== 3) bad(`expected exactly 3 submissions, found ${allSubs.length}`);
  else ok('exactly 3 submission artifacts (639, 640, 641 — all PENDING)');
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
  const t917 = await dbGet('odyssey_tasks?active_date=eq.2026-09-17&is_active=eq.true&select=id&order=id.asc');
  if (JSON.stringify(t917.map((t) => t.id)) !== '[639,640,641]') bad(`09-17 tasks changed: ${JSON.stringify(t917)}`);
  else ok('tasks 639/640/641 intact (exactly 3 active on 2026-09-17)');

  console.log('\n====================================================');
  console.log('ARTIFACT MANIFEST: ' + JSON.stringify(manifest));
  console.log(`SMOKE RESULT: ${failures === 0 ? 'ALL PASS' : 'FAILURES: ' + failures}`);
  console.log('====================================================');
  if (failures) process.exit(1);
}

run().catch((e) => { console.error(e); process.exit(1); });
