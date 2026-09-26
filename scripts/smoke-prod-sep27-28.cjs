require('dotenv').config();

// Real authenticated production smoke for 2026-09-27 + 2026-09-28 tasks
// from Migration 093. Task IDs are resolved at runtime by exact
// title+date (no hardcoded ID assumptions).
// Strategy: disposable test user -> real flows -> exact-ID manifest ->
// exact-ID cleanup -> post-cleanup verification. No broad deletes.
// No redeem/claims. No purchase is made: photos are disposable synthetic
// 1px test images (clearly-marked notes, will be deleted); videos are
// synthetic ftyp-only clips (no real content). ADMIN_REVIEW tasks reach
// PENDING only (zero ledger). The AUTO mini-game is approved by the
// existing engine — its exact ledger/submission IDs are recorded in the
// manifest and deleted in cleanup.

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
const TEST_EXPLORER = `Smoke Sep27-28 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const TITLES_27 = [
  'Sehari Kembali Pakai Seragam',
  'Kalau Sekarang Masih Pakai Seragam',
  'Balik Jadi Anak Sekolah Sehari',
];
const TITLES_28 = [
  'Seragam vs Dirimu Sekarang',
  'Kalau Bisa Balik ke Hari Pertama Sekolah',
  'Dari Seragam Sekolah ke Seragam Kerja',
];

const GAME_CHOICES = { ev_1: 'a', ev_2: 'a', ev_3: 'a', ev_4: 'a', ev_5: 'a' };

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
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-09-27/28 (093)');
  console.log(` disposable user: ${TEST_USER} / ${TEST_UID}`);
  console.log('====================================================\n');

  // --- 0. Collision check ---
  console.log('[0] Collision check (must be empty)');
  const c1 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const c2 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  if (c1.length || c2.length) { bad('test identity collides with existing user — ABORT'); process.exit(1); }
  ok('no collision for disposable identity');

  // --- 0b. Resolve task IDs by exact title+date ---
  console.log('\n[0b] Resolve task IDs (exact title + date, 3 active per date)');
  const r27 = await dbGet('odyssey_tasks?active_date=eq.2026-09-27&is_active=eq.true&select=id,title,step_order&order=step_order.asc');
  const r28 = await dbGet('odyssey_tasks?active_date=eq.2026-09-28&is_active=eq.true&select=id,title,step_order&order=step_order.asc');
  if (r27.length !== 3 || JSON.stringify(r27.map((t) => t.title)) !== JSON.stringify(TITLES_27)) {
    bad(`09-27 task set wrong: ${JSON.stringify(r27.map((t) => t.title))}`); process.exit(1);
  }
  if (r28.length !== 3 || JSON.stringify(r28.map((t) => t.title)) !== JSON.stringify(TITLES_28)) {
    bad(`09-28 task set wrong: ${JSON.stringify(r28.map((t) => t.title))}`); process.exit(1);
  }
  const [T1, T2, T3] = r27.map((t) => t.id);
  const [T4, T5, T6] = r28.map((t) => t.id);
  console.log(`   T1 PHOTO=${T1} T2 VIDEO=${T2} T3 MINI_GAME=${T3} T4 PHOTO=${T4} T5 TEXT=${T5} T6 VIDEO=${T6}`);
  ok('resolved 6 task IDs by exact title+date');

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

  // --- 3. Open all 6 tasks (DB-rendered config) ---
  console.log('\n[3] Open task details (must be DB-rendered)');
  await sleep(3000);
  const t = {};
  for (const id of [T1, T2, T3, T4, T5, T6]) {
    t[id] = await prodApi(`/api/tasks/${id}`, { token });
    if (t[id].status !== 200) bad(`open ${id} failed: ${t[id].status}`);
  }
  const c1cfg = t[T1].json?.config || {};
  if (t[T1].json?.task_type === 'PHOTO_UPLOAD' && t[T1].json?.evaluation_type === 'ADMIN_REVIEW' && c1cfg.max_files === 1 && c1cfg.camera_only === false && /Kesan/.test(c1cfg.instruction || '')) ok('T1 PHOTO max_files=1 camera_only=false ADMIN_REVIEW + 2-part note (DB config)');
  else bad(`T1 config mismatch: ${JSON.stringify(c1cfg).slice(0, 200)}`);
  if (t[T1].json?.reward_coins === 40 && t[T1].json?.reward_xp === 100) ok('T1 reward 40c/100xp (DB)');
  else bad('T1 reward mismatch');
  const c2cfg = t[T2].json?.config || {};
  if (t[T2].json?.task_type === 'VIDEO' && t[T2].json?.evaluation_type === 'ADMIN_REVIEW' && c2cfg.recording?.enabled === true && c2cfg.recording?.max_duration_seconds === 60 && /Perbedaan/.test(c2cfg.recording?.instruction || '')) ok('T2 VIDEO recording enabled max 60s ADMIN_REVIEW + story instruction (DB config)');
  else bad(`T2 config mismatch: ${JSON.stringify(c2cfg).slice(0, 200)}`);
  if (t[T2].json?.reward_coins === 40 && t[T2].json?.reward_xp === 100) ok('T2 reward 40c/100xp (DB)');
  else bad('T2 reward mismatch');
  const c3cfg = t[T3].json?.config || {};
  if (t[T3].json?.task_type === 'MINI_GAME' && c3cfg.game === 'DECISION_PRIORITY' && c3cfg.scenario?.currency === 'POINTS' && Array.isArray(c3cfg.scenario?.events) && c3cfg.scenario.events.length === 5) ok('T3 MINI_GAME DECISION_PRIORITY POINTS 5 events (DB config)');
  else bad(`T3 config mismatch: ${JSON.stringify(c3cfg).slice(0, 200)}`);
  if (t[T3].json?.reward_coins === 50 && t[T3].json?.reward_xp === 100) ok('T3 reward 50c/100xp (DB)');
  else bad('T3 reward mismatch');
  // Answer-key redaction through member API for the scenario engine.
  const memberCfgStr = JSON.stringify(c3cfg);
  if (!/correct_answer|answer_key|expected_answer|is_correct/.test(memberCfgStr)) ok('T3 no answer-key fields leak through member API');
  else bad('T3 answer-key leakage through member API');
  const c4cfg = t[T4].json?.config || {};
  if (t[T4].json?.task_type === 'PHOTO_UPLOAD' && t[T4].json?.evaluation_type === 'ADMIN_REVIEW' && c4cfg.max_files === 1 && c4cfg.camera_only === false && /Perubahan/.test(c4cfg.instruction || '')) ok('T4 PHOTO max_files=1 camera_only=false ADMIN_REVIEW + 2-part note (DB config)');
  else bad(`T4 config mismatch: ${JSON.stringify(c4cfg).slice(0, 200)}`);
  if (t[T4].json?.reward_coins === 40 && t[T4].json?.reward_xp === 100) ok('T4 reward 40c/100xp (DB)');
  else bad('T4 reward mismatch');
  const c5cfg = t[T5].json?.config || {};
  if (t[T5].json?.task_type === 'TEXT_RESPONSE' && t[T5].json?.evaluation_type === 'ADMIN_REVIEW' && c5cfg.minimum_characters === 150 && c5cfg.maximum_characters === 2000 && /Nasihat/.test(c5cfg.prompt || '')) ok('T5 TEXT bounds 150/2000 ADMIN_REVIEW + advice prompt (DB config)');
  else bad(`T5 config mismatch: ${JSON.stringify(c5cfg).slice(0, 200)}`);
  if (t[T5].json?.reward_coins === 40 && t[T5].json?.reward_xp === 100) ok('T5 reward 40c/100xp (DB)');
  else bad('T5 reward mismatch');
  const c6cfg = t[T6].json?.config || {};
  if (t[T6].json?.task_type === 'VIDEO' && t[T6].json?.evaluation_type === 'ADMIN_REVIEW' && c6cfg.recording?.enabled === true && c6cfg.recording?.max_duration_seconds === 60 && /Pelajaran/.test(c6cfg.recording?.instruction || '')) ok('T6 VIDEO recording enabled max 60s ADMIN_REVIEW + journey instruction (DB config)');
  else bad(`T6 config mismatch: ${JSON.stringify(c6cfg).slice(0, 200)}`);
  if (t[T6].json?.reward_coins === 40 && t[T6].json?.reward_xp === 100) ok('T6 reward 40c/100xp (DB)');
  else bad('T6 reward mismatch');

  // --- 4. Task T1: uniform photo + note (PENDING) ---
  console.log('\n[4] Task T1 uniform photo (upload + PENDING submit)');
  const pngB64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';
  const pngBytes = Buffer.from(pngB64, 'base64');
  const form1 = new FormData();
  form1.append('file', new Blob([pngBytes], { type: 'image/png' }), 'SMOKE-TEST-seragam.png');
  const up1 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form1 });
  if (up1.status !== 200 || !up1.json?.file_url) { bad(`T1 upload failed: ${up1.status} ${up1.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up1.json.storage_path);
    ok(`uploaded storage_path = ${up1.json.storage_path}`);
    const chk = await fetch(up1.json.file_url, { method: 'HEAD' });
    if (chk.ok) ok(`storage file reachable via file_url (HEAD ${chk.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk.status})`);
    await sleep(3000);
    const note1 = 'SMOKE-TEST (data uji, akan dihapus). (a) Kenangan: contoh kenangan uji — yang paling diingat contoh adalah upacara bendera contoh dan jajan di kantin contoh. (b) Kesan: contoh kesan uji — seragam contoh terasa sempit dan mengingatkan masa kecil yang seru contoh.';
    const s1 = await prodApi(`/api/tasks/${T1}/submit`, { method: 'POST', token, body: { payload: { file_url: up1.json.file_url, file_name: up1.json.file_name, file_size: up1.json.file_size, note: note1, submitted_at: new Date().toISOString() } } });
    if (s1.status !== 200 || s1.json?.success !== true) bad(`T1 submit failed: ${s1.status} ${s1.text.slice(0, 160)}`);
    else ok('T1 submit accepted (manual review path)');
    const rows1 = await dbGet(`odyssey_task_submissions?task_id=eq.${T1}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows1.length !== 1 || rows1[0].status !== 'PENDING') bad(`T1 submission state wrong: ${JSON.stringify(rows1).slice(0, 300)}`);
    else { manifest.submissions[T1] = rows1[0].id; ok(`submission_id = ${rows1[0].id} (PENDING, file_url: ${!!rows1[0].payload?.file_url}, note: ${!!rows1[0].payload?.note})`); }
  }

  // --- 5. Task T2: uniform video (upload + PENDING submit) ---
  console.log('\n[5] Task T2 uniform video (upload + PENDING submit)');
  const ftyp = Buffer.from([
    0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6f, 0x6d,
    0x00, 0x00, 0x00, 0x00, 0x69, 0x73, 0x6f, 0x6d, 0x69, 0x73, 0x6f, 0x32,
  ]);
  const form2 = new FormData();
  form2.append('file', new Blob([ftyp], { type: 'video/mp4' }), 'SMOKE-TEST-seragam-video.mp4');
  const up2 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form2 });
  if (up2.status !== 200 || !up2.json?.file_url) { bad(`T2 upload failed: ${up2.status} ${up2.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up2.json.storage_path);
    ok(`uploaded storage_path = ${up2.json.storage_path}`);
    const chk2 = await fetch(up2.json.file_url, { method: 'HEAD' });
    if (chk2.ok) ok(`storage file reachable via file_url (HEAD ${chk2.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk2.status})`);
    await sleep(3000);
    const s2 = await prodApi(`/api/tasks/${T2}/submit`, {
      method: 'POST', token,
      body: { payload: { file_url: up2.json.file_url, file_name: up2.json.file_name, file_size: up2.json.file_size, duration_seconds: 45, mime_type: 'video/mp4', note: 'SMOKE-TEST: video uji disposable (bukan konten asli), mohon abaikan', captured_at: new Date().toISOString() } },
    });
    if (s2.status !== 200 || s2.json?.success !== true) bad(`T2 submit failed: ${s2.status} ${s2.text.slice(0, 160)}`);
    else ok('T2 submit accepted (manual review path, duration 45s within 60s max)');
    const rows2 = await dbGet(`odyssey_task_submissions?task_id=eq.${T2}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows2.length !== 1 || rows2[0].status !== 'PENDING') bad(`T2 submission state wrong: ${JSON.stringify(rows2).slice(0, 300)}`);
    else { manifest.submissions[T2] = rows2[0].id; ok(`submission_id = ${rows2[0].id} (PENDING, file_url present: ${!!rows2[0].payload?.file_url})`); }
  }

  // --- 6. Task T3: invalid rejected, spoof ignored, valid approved ---
  console.log('\n[6] Task T3 school-day scenario (invalid rejected, spoof ignored, valid APPROVED)');
  await sleep(3000);
  const bad3 = await prodApi(`/api/tasks/${T3}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: { ev_1: 'zzz' }, final_balance: 999999, score: 100 } } });
  if (bad3.status === 200 && bad3.json?.success === true) bad('T3 accepted spoofed/invalid choices — server recompute BYPASSED');
  else ok(`T3 rejected invalid/spoofed choices (status ${bad3.status})`);
  await sleep(3000);
  const s3 = await prodApi(`/api/tasks/${T3}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: GAME_CHOICES, final_balance: 0, score: 0 } } });
  if (s3.status !== 200 || s3.json?.success !== true) bad(`T3 valid submit failed: ${s3.status} ${s3.text.slice(0, 200)}`);
  else ok(`T3 valid choices APPROVED despite spoofed score=0/balance=0 (server recompute authoritative, coins_earned=${s3.json?.coins_earned})`);
  const rows3 = await dbGet(`odyssey_task_submissions?task_id=eq.${T3}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows3.length !== 1 || rows3[0].status !== 'APPROVED') bad(`T3 submission state wrong: ${JSON.stringify(rows3).slice(0, 300)}`);
  else {
    manifest.submissions[T3] = rows3[0].id;
    const fb = rows3[0].payload?.final_balance;
    ok(`submission_id = ${rows3[0].id} (APPROVED, server final_balance=${fb}, expected 100)`);
    if (fb !== 100) bad(`T3 server balance wrong: ${fb} (expected 100)`);
  }
  const led3 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows3[0]?.id}`);
  console.log(`       ledger rows for T3 submission: ${JSON.stringify(led3)}`);
  manifest.ledger.push(...led3.map((l) => l.id));

  // --- 7. Task T4: then-vs-now photo + note (PENDING) ---
  console.log('\n[7] Task T4 then-vs-now photo (upload + PENDING submit)');
  const form4 = new FormData();
  form4.append('file', new Blob([pngBytes], { type: 'image/png' }), 'SMOKE-TEST-seragam-vs-sekarang.png');
  const up4 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form4 });
  if (up4.status !== 200 || !up4.json?.file_url) { bad(`T4 upload failed: ${up4.status} ${up4.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up4.json.storage_path);
    ok(`uploaded storage_path = ${up4.json.storage_path}`);
    const chk4 = await fetch(up4.json.file_url, { method: 'HEAD' });
    if (chk4.ok) ok(`storage file reachable via file_url (HEAD ${chk4.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk4.status})`);
    await sleep(3000);
    const note4 = 'SMOKE-TEST (data uji, akan dihapus). (a) Pose: contoh pose uji — pose percaya diri contoh menunjukkan perbedaan anak sekolah contoh dengan pekerja contoh. (b) Perubahan: contoh perubahan uji — dulu pemalu contoh, sekarang berani bicara contoh.';
    const s4 = await prodApi(`/api/tasks/${T4}/submit`, { method: 'POST', token, body: { payload: { file_url: up4.json.file_url, file_name: up4.json.file_name, file_size: up4.json.file_size, note: note4, submitted_at: new Date().toISOString() } } });
    if (s4.status !== 200 || s4.json?.success !== true) bad(`T4 submit failed: ${s4.status} ${s4.text.slice(0, 160)}`);
    else ok('T4 submit accepted (manual review path)');
    const rows4 = await dbGet(`odyssey_task_submissions?task_id=eq.${T4}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows4.length !== 1 || rows4[0].status !== 'PENDING') bad(`T4 submission state wrong: ${JSON.stringify(rows4).slice(0, 300)}`);
    else { manifest.submissions[T4] = rows4[0].id; ok(`submission_id = ${rows4[0].id} (PENDING, file_url: ${!!rows4[0].payload?.file_url}, note: ${!!rows4[0].payload?.note})`); }
  }

  // --- 8. Task T5: min-length validation + valid submit (PENDING) ---
  console.log('\n[8] Task T5 first-day-of-school text (validation + PENDING submit)');
  await sleep(3000);
  const short5 = await prodApi(`/api/tasks/${T5}/submit`, { method: 'POST', token, body: { payload: { text: 'SMOKE-TEST pendek', submitted_at: new Date().toISOString() } } });
  if (short5.status === 200 && short5.json?.success === true) bad('T5 accepted too-short text — min-length validation BYPASSED');
  else if (/minimal 150/.test(short5.text)) ok('T5 server rejected short text with min-150 message (DB-configured validation)');
  else bad(`T5 short-text rejection unclear: ${short5.status} ${short5.text.slice(0, 160)}`);
  const text5 = 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Berbeda: contoh hal uji pertama — akan lebih rajin belajar contoh dan tidak menunda tugas contoh; contoh hal uji kedua — akan lebih berani bertanya contoh dan ikut kegiatan contoh. 2) Nasihat: contoh nasihat uji — jangan takut salah contoh, karena setiap kesalahan contoh adalah pelajaran contoh yang berharga contoh.';
  console.log(`       evidence len = ${text5.length} (bounds 150/2000)`);
  await sleep(3000);
  const s5 = await prodApi(`/api/tasks/${T5}/submit`, { method: 'POST', token, body: { payload: { text: text5, submitted_at: new Date().toISOString() } } });
  if (s5.status !== 200 || s5.json?.success !== true) bad(`T5 submit failed: ${s5.status} ${s5.text.slice(0, 160)}`);
  else ok('T5 submit accepted (manual review path)');
  const rows5 = await dbGet(`odyssey_task_submissions?task_id=eq.${T5}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows5.length !== 1 || rows5[0].status !== 'PENDING') bad(`T5 submission state wrong: ${JSON.stringify(rows5).slice(0, 300)}`);
  else { manifest.submissions[T5] = rows5[0].id; ok(`submission_id = ${rows5[0].id} (PENDING, text present: ${!!rows5[0].payload?.text})`); }

  // --- 9. Task T6: journey video upload + submit (PENDING) ---
  console.log('\n[9] Task T6 school-to-work video (upload + PENDING submit)');
  const form6 = new FormData();
  form6.append('file', new Blob([ftyp], { type: 'video/mp4' }), 'SMOKE-TEST-seragam-ke-kerja.mp4');
  const up6 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form6 });
  if (up6.status !== 200 || !up6.json?.file_url) { bad(`T6 upload failed: ${up6.status} ${up6.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up6.json.storage_path);
    ok(`uploaded storage_path = ${up6.json.storage_path}`);
    const chk6 = await fetch(up6.json.file_url, { method: 'HEAD' });
    if (chk6.ok) ok(`storage file reachable via file_url (HEAD ${chk6.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk6.status})`);
    await sleep(3000);
    const s6 = await prodApi(`/api/tasks/${T6}/submit`, {
      method: 'POST', token,
      body: { payload: { file_url: up6.json.file_url, file_name: up6.json.file_name, file_size: up6.json.file_size, duration_seconds: 50, mime_type: 'video/mp4', note: 'SMOKE-TEST: video uji disposable (bukan cerita asli), mohon abaikan', captured_at: new Date().toISOString() } },
    });
    if (s6.status !== 200 || s6.json?.success !== true) bad(`T6 submit failed: ${s6.status} ${s6.text.slice(0, 160)}`);
    else ok('T6 submit accepted (manual review path, duration 50s within 60s max)');
    const rows6 = await dbGet(`odyssey_task_submissions?task_id=eq.${T6}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows6.length !== 1 || rows6[0].status !== 'PENDING') bad(`T6 submission state wrong: ${JSON.stringify(rows6).slice(0, 300)}`);
    else { manifest.submissions[T6] = rows6[0].id; ok(`submission_id = ${rows6[0].id} (PENDING, file_url present: ${!!rows6[0].payload?.file_url})`); }
  }

  // --- 10. Enumerate all test-user artifacts ---
  console.log('\n[10] Artifact enumeration (ownership = disposable user only)');
  const allSubs = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id,task_id,status`);
  const ledgerAll = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id`);
  const claims = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  manifest.claims = claims;
  console.log(`       submissions: ${JSON.stringify(allSubs)}`);
  console.log(`       ledger rows: ${JSON.stringify(ledgerAll)} | claims: ${claims.length}`);
  if (allSubs.length !== 6) bad(`expected exactly 6 submissions, found ${allSubs.length}`);
  else ok('exactly 6 submission artifacts (T1-T6)');
  const pendCount = allSubs.filter((s) => s.status === 'PENDING').length;
  const apprCount = allSubs.filter((s) => s.status === 'APPROVED').length;
  if (pendCount === 5 && apprCount === 1) ok('5 PENDING (ADMIN_REVIEW) + 1 APPROVED (AUTO mini-game) as designed');
  else bad(`status split wrong: PENDING=${pendCount} APPROVED=${apprCount}`);
  if (claims.length) bad('unexpected claim rows');
  else ok('no claim artifacts (nothing redeemed)');
  const prof = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid,coins,xp`);
  console.log(`       profile: ${JSON.stringify(prof)}`);

  // --- 11. Exact-ID cleanup ---
  console.log('\n[11] Exact-ID cleanup (only manifest artifacts)');
  for (const lid of manifest.ledger) {
    const own = await dbGet(`odyssey_coin_transactions?id=eq.${lid}&select=id,user_uid`);
    if (own.length !== 1 || own[0].user_uid !== TEST_UID) { bad(`ownership unproven for ledger ${lid} — SKIPPING its delete`); continue; }
    await dbDel('odyssey_coin_transactions', `id=eq.${lid}`);
    const gone = await dbGet(`odyssey_coin_transactions?id=eq.${lid}&select=id`);
    if (gone.length) bad(`ledger ${lid} still present`);
    else ok(`ledger ${lid} deleted + verified absent`);
  }
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

  // --- 12. Post-cleanup verification + unchanged production state ---
  console.log('\n[12] Post-cleanup verification');
  const r1 = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id`);
  const r2 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const r3 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  const r4 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id`);
  const r5 = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  if (r1.length || r2.length || r3.length || r4.length || r5.length) bad(`residual artifacts: subs=${r1.length} uid=${r2.length} uname=${r3.length} ledger=${r4.length} claims=${r5.length}`);
  else ok('zero remaining test artifacts (submissions/profile/ledger/claims)');
  const t927 = await dbGet('odyssey_tasks?active_date=eq.2026-09-27&is_active=eq.true&select=id,title&order=step_order.asc');
  const t928 = await dbGet('odyssey_tasks?active_date=eq.2026-09-28&is_active=eq.true&select=id,title&order=step_order.asc');
  if (t927.length === 3 && JSON.stringify(t927.map((x) => x.title)) === JSON.stringify(TITLES_27)) ok(`09-27 intact (exactly 3 active: ${t927.map((x) => x.id).join(',')})`);
  else bad(`09-27 tasks changed: ${JSON.stringify(t927)}`);
  if (t928.length === 3 && JSON.stringify(t928.map((x) => x.title)) === JSON.stringify(TITLES_28)) ok(`09-28 intact (exactly 3 active: ${t928.map((x) => x.id).join(',')})`);
  else bad(`09-28 tasks changed: ${JSON.stringify(t928)}`);

  console.log('\n====================================================');
  console.log('ARTIFACT MANIFEST: ' + JSON.stringify(manifest));
  console.log(`SMOKE RESULT: ${failures === 0 ? 'ALL PASS' : 'FAILURES: ' + failures}`);
  console.log('====================================================');
  if (failures) process.exit(1);
}

run().catch((err) => {
  console.error('Smoke failed:', err);
  process.exit(1);
});
