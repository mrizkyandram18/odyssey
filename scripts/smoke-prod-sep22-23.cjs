require('dotenv').config();

// Real authenticated production smoke for 2026-09-22 + 2026-09-23 tasks
// from Migration 091. Task IDs are resolved at runtime by exact
// title+date (no hardcoded ID assumptions).
// Strategy: disposable test user -> real flows -> exact-ID manifest ->
// exact-ID cleanup -> post-cleanup verification. No broad deletes.
// No redeem/claims. No purchase is made: task 2 evidence is a fictional
// clearly-marked test guide; task 3 photo is a disposable synthetic test
// image of a harmless household observation; task 6 video is a synthetic
// ftyp-only clip (no real content). ADMIN_REVIEW tasks reach PENDING only
// (zero ledger). AUTO tasks are approved by the existing engine — their
// exact ledger/submission IDs are recorded in the manifest and deleted
// in cleanup.

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
const TEST_EXPLORER = `Smoke Sep22-23 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const TITLES_22 = [
  'Bongkar Cara Kerjanya',
  'Bikin Petunjuk yang Nggak Bikin Bingung',
  'Apa yang Bisa Kamu Pelajari dari Ini?',
];
const TITLES_23 = [
  'Uangnya Lari ke Mana?',
  'Cek Sebelum Percaya',
  'Cerita dari Pengalamanmu',
];

const QUIZ_CORRECT = { q1: 'A', q2: 'A', q3: 'B', q4: 'A' };
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
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-09-22/23 (091)');
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
  const r22 = await dbGet('odyssey_tasks?active_date=eq.2026-09-22&is_active=eq.true&select=id,title,step_order&order=step_order.asc');
  const r23 = await dbGet('odyssey_tasks?active_date=eq.2026-09-23&is_active=eq.true&select=id,title,step_order&order=step_order.asc');
  if (r22.length !== 3 || JSON.stringify(r22.map((t) => t.title)) !== JSON.stringify(TITLES_22)) {
    bad(`09-22 task set wrong: ${JSON.stringify(r22.map((t) => t.title))}`); process.exit(1);
  }
  if (r23.length !== 3 || JSON.stringify(r23.map((t) => t.title)) !== JSON.stringify(TITLES_23)) {
    bad(`09-23 task set wrong: ${JSON.stringify(r23.map((t) => t.title))}`); process.exit(1);
  }
  const [T1, T2, T3] = r22.map((t) => t.id);
  const [T4, T5, T6] = r23.map((t) => t.id);
  console.log(`   T1 MINI_GAME=${T1} T2 DOC=${T2} T3 PHOTO=${T3} T4 QUIZ=${T4} T5 TEXT=${T5} T6 VIDEO=${T6}`);
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
  if (t[T1].json?.task_type === 'MINI_GAME' && c1cfg.game === 'DECISION_PRIORITY' && c1cfg.scenario?.currency === 'POINTS' && Array.isArray(c1cfg.scenario?.events) && c1cfg.scenario.events.length === 5) ok('T1 MINI_GAME DECISION_PRIORITY POINTS 5 events (DB config)');
  else bad(`T1 config mismatch: ${JSON.stringify(c1cfg).slice(0, 200)}`);
  if (t[T1].json?.reward_coins === 50 && t[T1].json?.reward_xp === 100) ok('T1 reward 50c/100xp (DB)');
  else bad('T1 reward mismatch');
  const c2cfg = t[T2].json?.config || {};
  if (t[T2].json?.task_type === 'DOCUMENT_UPLOAD' && c2cfg.max_file_size_mb === 4 && Array.isArray(c2cfg.allowed_extensions) && c2cfg.allowed_extensions.includes('pdf') && /BERNOMOR/.test(c2cfg.instruction || '')) ok('T2 DOCUMENT ext list + max 4MB + numbered-steps instruction (DB config)');
  else bad(`T2 config mismatch: ${JSON.stringify(c2cfg).slice(0, 200)}`);
  if (t[T2].json?.reward_coins === 40 && t[T2].json?.reward_xp === 100) ok('T2 reward 40c/100xp (DB)');
  else bad('T2 reward mismatch');
  const c3cfg = t[T3].json?.config || {};
  if (t[T3].json?.task_type === 'PHOTO_UPLOAD' && c3cfg.max_files === 1 && c3cfg.camera_only === false && /Manfaat/.test(c3cfg.instruction || '')) ok('T3 PHOTO max_files=1 camera_only=false + 3-part note (DB config)');
  else bad(`T3 config mismatch: ${JSON.stringify(c3cfg).slice(0, 200)}`);
  if (t[T3].json?.reward_coins === 40 && t[T3].json?.reward_xp === 100) ok('T3 reward 40c/100xp (DB)');
  else bad('T3 reward mismatch');
  const c4cfg = t[T4].json?.config || {};
  if (t[T4].json?.task_type === 'QUIZ' && Array.isArray(c4cfg.questions) && c4cfg.questions.length === 4 && c4cfg.questions.every((q) => !('correct_answer' in q) && q.explanation && q.options?.length >= 2)) ok('T4 QUIZ 4 questions, keys redacted from member API, explanations present (DB config renders)');
  else bad(`T4 config mismatch: ${JSON.stringify(c4cfg).slice(0, 200)}`);
  const db4 = await dbGet(`odyssey_tasks?id=eq.${T4}&select=config`);
  if (db4.length === 1 && Array.isArray(db4[0].config?.questions) && db4[0].config.questions.length === 4 && db4[0].config.questions.every((q) => !!q.correct_answer)) ok('T4 correct_answer keys present in DB config (server-side grading source)');
  else bad('T4 DB answer keys missing');
  if (t[T4].json?.reward_coins === 50 && t[T4].json?.reward_xp === 100) ok('T4 reward 50c/100xp (DB)');
  else bad('T4 reward mismatch');
  const c5cfg = t[T5].json?.config || {};
  if (t[T5].json?.task_type === 'TEXT_RESPONSE' && c5cfg.minimum_characters === 150 && c5cfg.maximum_characters === 2000 && /Cara memverifikasi/.test(c5cfg.prompt || '')) ok('T5 TEXT bounds 150/2000 + verify-before-trust prompt (DB config)');
  else bad(`T5 config mismatch: ${JSON.stringify(c5cfg).slice(0, 200)}`);
  if (t[T5].json?.reward_coins === 40 && t[T5].json?.reward_xp === 100) ok('T5 reward 40c/100xp (DB)');
  else bad('T5 reward mismatch');
  const c6cfg = t[T6].json?.config || {};
  if (t[T6].json?.task_type === 'VIDEO' && c6cfg.recording?.enabled === true && c6cfg.recording?.max_duration_seconds === 60 && /Perubahan/.test(c6cfg.recording?.instruction || '')) ok('T6 VIDEO recording enabled max 60s + 3-part story instruction (DB config)');
  else bad(`T6 config mismatch: ${JSON.stringify(c6cfg).slice(0, 200)}`);
  if (t[T6].json?.reward_coins === 40 && t[T6].json?.reward_xp === 100) ok('T6 reward 40c/100xp (DB)');
  else bad('T6 reward mismatch');

  // --- 4. Task T1: invalid rejected, spoof ignored, valid approved ---
  console.log('\n[4] Task T1 how-it-works scenario (invalid rejected, spoof ignored, valid APPROVED)');
  await sleep(3000);
  const bad1 = await prodApi(`/api/tasks/${T1}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: { ev_1: 'zzz' }, final_balance: 999999, score: 100 } } });
  if (bad1.status === 200 && bad1.json?.success === true) bad('T1 accepted spoofed/invalid choices — server recompute BYPASSED');
  else ok(`T1 rejected invalid/spoofed choices (status ${bad1.status})`);
  await sleep(3000);
  const s1 = await prodApi(`/api/tasks/${T1}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: GAME_CHOICES, final_balance: 0, score: 0 } } });
  if (s1.status !== 200 || s1.json?.success !== true) bad(`T1 valid submit failed: ${s1.status} ${s1.text.slice(0, 200)}`);
  else ok(`T1 valid choices APPROVED despite spoofed score=0/balance=0 (server recompute authoritative, coins_earned=${s1.json?.coins_earned})`);
  const rows1 = await dbGet(`odyssey_task_submissions?task_id=eq.${T1}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows1.length !== 1 || rows1[0].status !== 'APPROVED') bad(`T1 submission state wrong: ${JSON.stringify(rows1).slice(0, 300)}`);
  else {
    manifest.submissions[T1] = rows1[0].id;
    const fb = rows1[0].payload?.final_balance;
    ok(`submission_id = ${rows1[0].id} (APPROVED, server final_balance=${fb}, expected 100)`);
    if (fb !== 100) bad(`T1 server balance wrong: ${fb} (expected 100)`);
  }
  const led1 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows1[0]?.id}`);
  console.log(`       ledger rows for T1 submission: ${JSON.stringify(led1)}`);
  manifest.ledger.push(...led1.map((l) => l.id));

  // --- 5. Task T2: disposable guide document upload + PENDING submit ---
  console.log('\n[5] Task T2 clear-guide document (upload + PENDING submit)');
  const guideText = 'SMOKE-TEST DATA (panduan uji fiktif, akan dihapus).\nJudul: Contoh Panduan Uji Merapikan Meja.\nTujuan: contoh tujuan uji agar meja contoh rapi dalam 10 menit contoh.\nBahan: contoh bahan uji — kardus contoh, lap contoh, kantong contoh.\nLangkah:\n1. Contoh langkah uji pertama — keluarkan semua barang contoh dari meja contoh.\n2. Contoh langkah uji kedua — kelompokkan barang contoh per jenis contoh.\n3. Contoh langkah uji ketiga — lap meja contoh sampai bersih contoh.\n4. Contoh langkah uji keempat — kembalikan barang contoh ke tempatnya contoh.\nTips: contoh tips uji — lakukan 5 menit tiap hari contoh agar tidak menumpuk contoh.';
  console.log(`       evidence len = ${guideText.length} chars (txt, within 4MB)`);
  const form2 = new FormData();
  form2.append('file', new Blob([guideText], { type: 'text/plain' }), 'SMOKE-TEST-petunjuk.txt');
  const up2 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form2 });
  if (up2.status !== 200 || !up2.json?.file_url) { bad(`T2 upload failed: ${up2.status} ${up2.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up2.json.storage_path);
    ok(`uploaded storage_path = ${up2.json.storage_path}`);
    const chk2 = await fetch(up2.json.file_url, { method: 'HEAD' });
    if (chk2.ok) ok(`storage file reachable via file_url (HEAD ${chk2.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk2.status})`);
    await sleep(3000);
    const s2 = await prodApi(`/api/tasks/${T2}/submit`, { method: 'POST', token, body: { payload: { file_url: up2.json.file_url, file_name: up2.json.file_name, file_size: up2.json.file_size, note: 'SMOKE-TEST: panduan uji disposable (contoh fiktif), mohon abaikan', submitted_at: new Date().toISOString() } } });
    if (s2.status !== 200 || s2.json?.success !== true) bad(`T2 submit failed: ${s2.status} ${s2.text.slice(0, 160)}`);
    else ok('T2 submit accepted (manual review path)');
    const rows2 = await dbGet(`odyssey_task_submissions?task_id=eq.${T2}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows2.length !== 1 || rows2[0].status !== 'PENDING') bad(`T2 submission state wrong: ${JSON.stringify(rows2).slice(0, 300)}`);
    else { manifest.submissions[T2] = rows2[0].id; ok(`submission_id = ${rows2[0].id} (PENDING, file_url: ${!!rows2[0].payload?.file_url})`); }
  }

  // --- 6. Task T3: harmless observation photo + 3-part note (PENDING) ---
  console.log('\n[6] Task T3 observational-learning photo (upload + PENDING submit)');
  const pngB64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';
  const pngBytes = Buffer.from(pngB64, 'base64');
  const form3 = new FormData();
  form3.append('file', new Blob([pngBytes], { type: 'image/png' }), 'SMOKE-TEST-belajar.png');
  const up3 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form3 });
  if (up3.status !== 200 || !up3.json?.file_url) { bad(`T3 upload failed: ${up3.status} ${up3.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up3.json.storage_path);
    ok(`uploaded storage_path = ${up3.json.storage_path}`);
    const chk = await fetch(up3.json.file_url, { method: 'HEAD' });
    if (chk.ok) ok(`storage file reachable via file_url (HEAD ${chk.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk.status})`);
    await sleep(3000);
    const note3 = 'SMOKE-TEST (data uji, akan dihapus). (a) Temuan: contoh temuan uji — tanaman contoh tumbuh di celah tembok contoh. (b) Pelajaran: contoh pelajaran uji — biji contoh bisa tumbuh dengan sedikit tanah dan air hujan contoh. (c) Manfaat: contoh manfaat uji — mengingatkan untuk merawat tanaman contoh dengan rajin menyiram contoh.';
    const s3 = await prodApi(`/api/tasks/${T3}/submit`, { method: 'POST', token, body: { payload: { file_url: up3.json.file_url, file_name: up3.json.file_name, file_size: up3.json.file_size, note: note3, submitted_at: new Date().toISOString() } } });
    if (s3.status !== 200 || s3.json?.success !== true) bad(`T3 submit failed: ${s3.status} ${s3.text.slice(0, 160)}`);
    else ok('T3 submit accepted (manual review path)');
    const rows3 = await dbGet(`odyssey_task_submissions?task_id=eq.${T3}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows3.length !== 1 || rows3[0].status !== 'PENDING') bad(`T3 submission state wrong: ${JSON.stringify(rows3).slice(0, 300)}`);
    else { manifest.submissions[T3] = rows3[0].id; ok(`submission_id = ${rows3[0].id} (PENDING, file_url: ${!!rows3[0].payload?.file_url}, note: ${!!rows3[0].payload?.note})`); }
  }

  // --- 7. Task T4: wrong rejected, correct approved (AUTO + ledger) ---
  console.log('\n[7] Task T4 money-tracking quiz (invalid rejected, valid APPROVED)');
  await sleep(3000);
  const wrong4 = await prodApi(`/api/tasks/${T4}/submit`, { method: 'POST', token, body: { answers: { q1: 'B', q2: 'C', q3: 'A', q4: 'B' } } });
  if (wrong4.status === 200 && wrong4.json?.success === true) bad('T4 accepted all-wrong answers — AUTO grading BYPASSED');
  else ok(`T4 rejected all-wrong answers (status ${wrong4.status})`);
  await sleep(3000);
  const s4 = await prodApi(`/api/tasks/${T4}/submit`, { method: 'POST', token, body: { answers: QUIZ_CORRECT } });
  if (s4.status !== 200 || s4.json?.success !== true) bad(`T4 correct submit failed: ${s4.status} ${s4.text.slice(0, 200)}`);
  else ok(`T4 correct answers APPROVED (coins_earned=${s4.json?.coins_earned}, xp=${s4.json?.xp_earned})`);
  const rows4 = await dbGet(`odyssey_task_submissions?task_id=eq.${T4}&user_uid=eq.${TEST_UID}&select=id,status,coins_earned,xp_earned`);
  if (rows4.length !== 1 || rows4[0].status !== 'APPROVED') bad(`T4 submission state wrong: ${JSON.stringify(rows4)}`);
  else { manifest.submissions[T4] = rows4[0].id; ok(`submission_id = ${rows4[0].id} (APPROVED)`); }
  const led4 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows4[0]?.id}`);
  console.log(`       ledger rows for T4 submission: ${JSON.stringify(led4)}`);
  manifest.ledger.push(...led4.map((l) => l.id));

  // --- 8. Task T5: min-length validation + valid submit (PENDING) ---
  console.log('\n[8] Task T5 verify-before-trust text (validation + PENDING submit)');
  await sleep(3000);
  const short5 = await prodApi(`/api/tasks/${T5}/submit`, { method: 'POST', token, body: { payload: { text: 'SMOKE-TEST pendek', submitted_at: new Date().toISOString() } } });
  if (short5.status === 200 && short5.json?.success === true) bad('T5 accepted too-short text — min-length validation BYPASSED');
  else if (/minimal 150/.test(short5.text)) ok('T5 server rejected short text with min-150 message (DB-configured validation)');
  else bad(`T5 short-text rejection unclear: ${short5.status} ${short5.text.slice(0, 160)}`);
  const text5 = 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Informasi: contoh info uji — iklan contoh diskon besar contoh. 2) Diperiksa: contoh hal uji — pengirim contoh dan syarat contoh. 3) Curiga: contoh tanda uji — disuruh buru-buru contoh dan diminta data pribadi contoh. 4) Verifikasi: contoh cara uji — cek sumber resmi contoh dan tanya orang dewasa contoh. 5) Alasan: contoh alasan uji — agar tidak tertipu contoh dan uang contoh aman contoh.';
  console.log(`       evidence len = ${text5.length} (bounds 150/2000)`);
  await sleep(3000);
  const s5 = await prodApi(`/api/tasks/${T5}/submit`, { method: 'POST', token, body: { payload: { text: text5, submitted_at: new Date().toISOString() } } });
  if (s5.status !== 200 || s5.json?.success !== true) bad(`T5 submit failed: ${s5.status} ${s5.text.slice(0, 160)}`);
  else ok('T5 submit accepted (manual review path)');
  const rows5 = await dbGet(`odyssey_task_submissions?task_id=eq.${T5}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows5.length !== 1 || rows5[0].status !== 'PENDING') bad(`T5 submission state wrong: ${JSON.stringify(rows5).slice(0, 300)}`);
  else { manifest.submissions[T5] = rows5[0].id; ok(`submission_id = ${rows5[0].id} (PENDING, text present: ${!!rows5[0].payload?.text})`); }

  // --- 9. Task T6: disposable test video upload + submit (PENDING) ---
  console.log('\n[9] Task T6 experience-story video (upload + PENDING submit)');
  const ftyp = Buffer.from([
    0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6f, 0x6d,
    0x00, 0x00, 0x00, 0x00, 0x69, 0x73, 0x6f, 0x6d, 0x69, 0x73, 0x6f, 0x32,
  ]);
  const form6 = new FormData();
  form6.append('file', new Blob([ftyp], { type: 'video/mp4' }), 'SMOKE-TEST-cerita.mp4');
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
      body: { payload: { file_url: up6.json.file_url, file_name: up6.json.file_name, file_size: up6.json.file_size, duration_seconds: 45, mime_type: 'video/mp4', note: 'SMOKE-TEST: video uji disposable (bukan cerita asli), mohon abaikan', captured_at: new Date().toISOString() } },
    });
    if (s6.status !== 200 || s6.json?.success !== true) bad(`T6 submit failed: ${s6.status} ${s6.text.slice(0, 160)}`);
    else ok('T6 submit accepted (manual review path)');
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
  if (pendCount === 4 && apprCount === 2) ok('4 PENDING (ADMIN_REVIEW) + 2 APPROVED (AUTO) as designed');
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
  const t922 = await dbGet('odyssey_tasks?active_date=eq.2026-09-22&is_active=eq.true&select=id,title&order=step_order.asc');
  const t923 = await dbGet('odyssey_tasks?active_date=eq.2026-09-23&is_active=eq.true&select=id,title&order=step_order.asc');
  if (t922.length === 3 && JSON.stringify(t922.map((x) => x.title)) === JSON.stringify(TITLES_22)) ok(`09-22 intact (exactly 3 active: ${t922.map((x) => x.id).join(',')})`);
  else bad(`09-22 tasks changed: ${JSON.stringify(t922)}`);
  if (t923.length === 3 && JSON.stringify(t923.map((x) => x.title)) === JSON.stringify(TITLES_23)) ok(`09-23 intact (exactly 3 active: ${t923.map((x) => x.id).join(',')})`);
  else bad(`09-23 tasks changed: ${JSON.stringify(t923)}`);

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
