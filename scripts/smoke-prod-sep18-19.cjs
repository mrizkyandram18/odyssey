require('dotenv').config();

// Real authenticated production smoke for 2026-09-18 (642/643/644) +
// 2026-09-19 (645/646/647) tasks from Migration 088.
// Strategy: disposable test user -> real flows -> exact-ID manifest ->
// exact-ID cleanup -> post-cleanup verification. No broad deletes.
// No redeem/claims. No purchase is made: task 642 evidence is fictional
// clearly-marked test data; 643 photo is a disposable synthetic test
// image of a harmless household observation; 647 video is a synthetic
// ftyp-only clip (no real content). ADMIN_REVIEW tasks (642/643/647)
// reach PENDING only (zero ledger). AUTO tasks (644/645/646) are
// approved by the existing engine — their exact ledger/submission IDs
// are recorded in the manifest and deleted in cleanup.

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
const TEST_EXPLORER = `Smoke Sep18-19 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const T642 = 642; // Kenapa Barang Ini Bisa Begitu? (TEXT_RESPONSE 150/2000, ADMIN_REVIEW)
const T643 = 643; // Uji Coba Sederhana (PHOTO_UPLOAD max 1, ADMIN_REVIEW)
const T644 = 644; // Mitos atau Fakta? (QUIZ 5Q, AUTO)
const T645 = 645; // Hitung Mana yang Lebih Hemat? (QUIZ 4Q, AUTO)
const T646 = 646; // Cari Jalan Keluar dari Masalah (MINI_GAME scenario, AUTO)
const T647 = 647; // Ajarkan Satu Skill (VIDEO 60s, ADMIN_REVIEW)

const C644_CORRECT = { q1: 'A', q2: 'A', q3: 'A', q4: 'B', q5: 'A' };
const C645_CORRECT = { q1: 'B', q2: 'B', q3: 'A', q4: 'B' };
const C646_CHOICES = { ev_1: 'a', ev_2: 'a', ev_3: 'a', ev_4: 'a' };

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
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-09-18/19 (642-647)');
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

  // --- 3. Open all 6 tasks (DB-rendered config) ---
  console.log('\n[3] Open task details (must be DB-rendered)');
  await sleep(3000);
  const t = {};
  for (const id of [T642, T643, T644, T645, T646, T647]) {
    t[id] = await prodApi(`/api/tasks/${id}`, { token });
    if (t[id].status !== 200) bad(`open ${id} failed: ${t[id].status}`);
  }
  const c642 = t[T642].json?.config || {};
  if (t[T642].json?.task_type === 'TEXT_RESPONSE' && c642.minimum_characters === 150 && c642.maximum_characters === 2000 && /Dari mana kamu mendapatkan/.test(c642.prompt || '')) ok('642 TEXT bounds 150/2000 + research prompt (DB config)');
  else bad(`642 config mismatch: ${JSON.stringify(c642).slice(0, 200)}`);
  if (t[T642].json?.reward_coins === 40 && t[T642].json?.reward_xp === 100) ok('642 reward 40c/100xp (DB)');
  else bad('642 reward mismatch');
  const c643 = t[T643].json?.config || {};
  if (t[T643].json?.task_type === 'PHOTO_UPLOAD' && c643.max_files === 1 && c643.camera_only === false && /Prediksi/.test(c643.instruction || '') && /DILARANG/.test(c643.instruction || '')) ok('643 PHOTO max_files=1 camera_only=false + 4-part note + safety rules (DB config)');
  else bad(`643 config mismatch: ${JSON.stringify(c643).slice(0, 200)}`);
  if (t[T643].json?.reward_coins === 40 && t[T643].json?.reward_xp === 100) ok('643 reward 40c/100xp (DB)');
  else bad('643 reward mismatch');
  const c644 = t[T644].json?.config || {};
  // NOTE: member API deliberately strips correct_answer (anti-cheat
  // sanitizeQuestions in family_tasks/api.go). Keys are verified via
  // direct DB read below; grading is proven by wrong-rejected/correct-
  // approved submits in step [6].
  if (t[T644].json?.task_type === 'QUIZ' && Array.isArray(c644.questions) && c644.questions.length === 5 && c644.questions.every((q) => !('correct_answer' in q) && q.explanation && q.options?.length >= 2)) ok('644 QUIZ 5 questions, keys redacted from member API, explanations present (DB config renders)');
  else bad(`644 config mismatch: ${JSON.stringify(c644).slice(0, 200)}`);
  const db644 = await dbGet(`odyssey_tasks?id=eq.${T644}&select=config`);
  if (db644.length === 1 && Array.isArray(db644[0].config?.questions) && db644[0].config.questions.length === 5 && db644[0].config.questions.every((q) => !!q.correct_answer)) ok('644 correct_answer keys present in DB config (server-side grading source)');
  else bad('644 DB answer keys missing');
  if (t[T644].json?.reward_coins === 50 && t[T644].json?.reward_xp === 100) ok('644 reward 50c/100xp (DB)');
  else bad('644 reward mismatch');
  const c645 = t[T645].json?.config || {};
  if (t[T645].json?.task_type === 'QUIZ' && Array.isArray(c645.questions) && c645.questions.length === 4 && c645.questions.every((q) => !('correct_answer' in q) && q.explanation && q.options?.length >= 2)) ok('645 QUIZ 4 calc questions, keys redacted from member API, explanations present (DB config renders)');
  else bad(`645 config mismatch: ${JSON.stringify(c645).slice(0, 200)}`);
  const db645 = await dbGet(`odyssey_tasks?id=eq.${T645}&select=config`);
  if (db645.length === 1 && Array.isArray(db645[0].config?.questions) && db645[0].config.questions.length === 4 && db645[0].config.questions.every((q) => !!q.correct_answer)) ok('645 correct_answer keys present in DB config (server-side grading source)');
  else bad('645 DB answer keys missing');
  if (t[T645].json?.reward_coins === 50 && t[T645].json?.reward_xp === 100) ok('645 reward 50c/100xp (DB)');
  else bad('645 reward mismatch');
  const c646 = t[T646].json?.config || {};
  if (t[T646].json?.task_type === 'MINI_GAME' && c646.game === 'DECISION_PRIORITY' && c646.scenario?.currency === 'POINTS' && Array.isArray(c646.scenario?.events) && c646.scenario.events.length === 4) ok('646 MINI_GAME DECISION_PRIORITY POINTS 4 events (DB config)');
  else bad(`646 config mismatch: ${JSON.stringify(c646).slice(0, 200)}`);
  if (t[T646].json?.reward_coins === 50 && t[T646].json?.reward_xp === 100) ok('646 reward 50c/100xp (DB)');
  else bad('646 reward mismatch');
  const c647 = t[T647].json?.config || {};
  if (t[T647].json?.task_type === 'VIDEO' && c647.recording?.enabled === true && c647.recording?.max_duration_seconds === 60 && /Demonstrasi/.test(c647.recording?.instruction || '')) ok('647 VIDEO recording enabled max 60s + demo instruction (DB config)');
  else bad(`647 config mismatch: ${JSON.stringify(c647).slice(0, 200)}`);
  if (t[T647].json?.reward_coins === 40 && t[T647].json?.reward_xp === 100) ok('647 reward 40c/100xp (DB)');
  else bad('647 reward mismatch');

  // --- 4. Task 642: min-length validation + fictional research submit (PENDING) ---
  console.log('\n[4] Task 642 research evidence (validation + PENDING submit)');
  await sleep(3000);
  const short642 = await prodApi(`/api/tasks/${T642}/submit`, { method: 'POST', token, body: { payload: { text: 'SMOKE-TEST pendek', submitted_at: new Date().toISOString() } } });
  if (short642.status === 200 && short642.json?.success === true) bad('642 accepted too-short text — min-length validation BYPASSED');
  else if (/minimal 150/.test(short642.text)) ok('642 server rejected short text with min-150 message (DB-configured validation)');
  else bad(`642 short-text rejection unclear: ${short642.status} ${short642.text.slice(0, 160)}`);
  const researchText = 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Benda: charger HP contoh (CONTOH UJI). 2) Pertanyaan: kenapa charger contoh jadi hangat saat dipakai? 3) Penjelasan yang ditemukan: contoh penjelasan uji — saat mengisi daya ada energi yang berubah menjadi panas contoh karena hambatan di komponen contoh, sehingga wajar bila charger contoh terasa hangat contoh. 4) Sumber belajar: contoh sumber uji — label kemasan contoh dan bertanya ke teman yang paham contoh. 5) Hal baru dipelajari: contoh pelajaran uji — panas kecil itu normal contoh tetapi charger yang terlalu panas contoh sebaiknya dicabut dan diperiksa contoh.';
  console.log(`       evidence len = ${researchText.length} (bounds 150/2000)`);
  await sleep(3000);
  const s642 = await prodApi(`/api/tasks/${T642}/submit`, { method: 'POST', token, body: { payload: { text: researchText, submitted_at: new Date().toISOString() } } });
  if (s642.status !== 200 || s642.json?.success !== true) bad(`642 submit failed: ${s642.status} ${s642.text.slice(0, 160)}`);
  else ok('642 submit accepted (manual review path)');
  const rows642 = await dbGet(`odyssey_task_submissions?task_id=eq.${T642}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows642.length !== 1 || rows642[0].status !== 'PENDING') bad(`642 submission state wrong: ${JSON.stringify(rows642).slice(0, 300)}`);
  else { manifest.submissions[T642] = rows642[0].id; ok(`submission_id = ${rows642[0].id} (PENDING, text present: ${!!rows642[0].payload?.text})`); }

  // --- 5. Task 643: harmless observation photo + 4-part note (PENDING) ---
  console.log('\n[5] Task 643 safe experiment (upload + PENDING submit)');
  const pngB64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';
  const pngBytes = Buffer.from(pngB64, 'base64');
  const form643 = new FormData();
  form643.append('file', new Blob([pngBytes], { type: 'image/png' }), 'SMOKE-TEST-esbatu.png');
  const up643 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form643 });
  if (up643.status !== 200 || !up643.json?.file_url) { bad(`643 upload failed: ${up643.status} ${up643.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up643.json.storage_path);
    ok(`uploaded storage_path = ${up643.json.storage_path}`);
    const chk = await fetch(up643.json.file_url, { method: 'HEAD' });
    if (chk.ok) ok(`storage file reachable via file_url (HEAD ${chk.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk.status})`);
    await sleep(3000);
    const note643 = 'SMOKE-TEST (data uji, akan dihapus). (a) Prediksi: contoh prediksi uji — es batu contoh di piring mencair lebih cepat daripada di gelas contoh. (b) Percobaan: contoh langkah uji — taruh satu es contoh di piring dan satu di gelas pada waktu yang sama lalu amati contoh. (c) Hasil: contoh hasil uji — es di piring mencair duluan contoh. (d) Kesimpulan: contoh kesimpulan uji — prediksi benar contoh, permukaan lebih luas mempercepat mencair contoh.';
    const s643 = await prodApi(`/api/tasks/${T643}/submit`, { method: 'POST', token, body: { payload: { file_url: up643.json.file_url, file_name: up643.json.file_name, file_size: up643.json.file_size, note: note643, submitted_at: new Date().toISOString() } } });
    if (s643.status !== 200 || s643.json?.success !== true) bad(`643 submit failed: ${s643.status} ${s643.text.slice(0, 160)}`);
    else ok('643 submit accepted (manual review path)');
    const rows643 = await dbGet(`odyssey_task_submissions?task_id=eq.${T643}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows643.length !== 1 || rows643[0].status !== 'PENDING') bad(`643 submission state wrong: ${JSON.stringify(rows643).slice(0, 300)}`);
    else { manifest.submissions[T643] = rows643[0].id; ok(`submission_id = ${rows643[0].id} (PENDING, file_url: ${!!rows643[0].payload?.file_url}, note: ${!!rows643[0].payload?.note})`); }
  }

  // --- 6. Task 644: wrong rejected, correct approved (AUTO + ledger) ---
  console.log('\n[6] Task 644 myth quiz (invalid rejected, valid APPROVED)');
  await sleep(3000);
  const wrong644 = await prodApi(`/api/tasks/${T644}/submit`, { method: 'POST', token, body: { answers: { q1: 'B', q2: 'B', q3: 'B', q4: 'A', q5: 'B' } } });
  if (wrong644.status === 200 && wrong644.json?.success === true) bad('644 accepted all-wrong answers — AUTO grading BYPASSED');
  else ok(`644 rejected all-wrong answers (status ${wrong644.status})`);
  await sleep(3000);
  const s644 = await prodApi(`/api/tasks/${T644}/submit`, { method: 'POST', token, body: { answers: C644_CORRECT } });
  if (s644.status !== 200 || s644.json?.success !== true) bad(`644 correct submit failed: ${s644.status} ${s644.text.slice(0, 200)}`);
  else ok(`644 correct answers APPROVED (coins_earned=${s644.json?.coins_earned}, xp=${s644.json?.xp_earned})`);
  const rows644 = await dbGet(`odyssey_task_submissions?task_id=eq.${T644}&user_uid=eq.${TEST_UID}&select=id,status,coins_earned,xp_earned`);
  if (rows644.length !== 1 || rows644[0].status !== 'APPROVED') bad(`644 submission state wrong: ${JSON.stringify(rows644)}`);
  else { manifest.submissions[T644] = rows644[0].id; ok(`submission_id = ${rows644[0].id} (APPROVED)`); }
  const led644 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows644[0]?.id}`);
  console.log(`       ledger rows for 644 submission: ${JSON.stringify(led644)}`);
  manifest.ledger.push(...led644.map((l) => l.id));

  // --- 7. Task 645: wrong rejected, correct approved (AUTO + ledger) ---
  console.log('\n[7] Task 645 hemat quiz (invalid rejected, valid APPROVED)');
  await sleep(3000);
  const wrong645 = await prodApi(`/api/tasks/${T645}/submit`, { method: 'POST', token, body: { answers: { q1: 'A', q2: 'A', q3: 'B', q4: 'A' } } });
  if (wrong645.status === 200 && wrong645.json?.success === true) bad('645 accepted all-wrong answers — AUTO grading BYPASSED');
  else ok(`645 rejected all-wrong answers (status ${wrong645.status})`);
  await sleep(3000);
  const s645 = await prodApi(`/api/tasks/${T645}/submit`, { method: 'POST', token, body: { answers: C645_CORRECT } });
  if (s645.status !== 200 || s645.json?.success !== true) bad(`645 correct submit failed: ${s645.status} ${s645.text.slice(0, 200)}`);
  else ok(`645 correct answers APPROVED (coins_earned=${s645.json?.coins_earned}, xp=${s645.json?.xp_earned})`);
  const rows645 = await dbGet(`odyssey_task_submissions?task_id=eq.${T645}&user_uid=eq.${TEST_UID}&select=id,status`);
  if (rows645.length !== 1 || rows645[0].status !== 'APPROVED') bad(`645 submission state wrong: ${JSON.stringify(rows645)}`);
  else { manifest.submissions[T645] = rows645[0].id; ok(`submission_id = ${rows645[0].id} (APPROVED)`); }
  const led645 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows645[0]?.id}`);
  console.log(`       ledger rows for 645 submission: ${JSON.stringify(led645)}`);
  manifest.ledger.push(...led645.map((l) => l.id));

  // --- 8. Task 646: invalid choices rejected, valid approved (AUTO + ledger) ---
  console.log('\n[8] Task 646 decision scenario (invalid rejected, valid APPROVED)');
  await sleep(3000);
  const bad646 = await prodApi(`/api/tasks/${T646}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: { ev_1: 'zzz' }, final_balance: 999999, score: 100 } } });
  if (bad646.status === 200 && bad646.json?.success === true) bad('646 accepted spoofed/invalid choices — server recompute BYPASSED');
  else ok(`646 rejected invalid/spoofed choices (status ${bad646.status})`);
  await sleep(3000);
  const s646 = await prodApi(`/api/tasks/${T646}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: C646_CHOICES, final_balance: 0, score: 0 } } });
  if (s646.status !== 200 || s646.json?.success !== true) bad(`646 valid submit failed: ${s646.status} ${s646.text.slice(0, 200)}`);
  else ok(`646 valid choices APPROVED despite spoofed score=0/balance=0 (server recompute authoritative, coins_earned=${s646.json?.coins_earned})`);
  const rows646 = await dbGet(`odyssey_task_submissions?task_id=eq.${T646}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows646.length !== 1 || rows646[0].status !== 'APPROVED') bad(`646 submission state wrong: ${JSON.stringify(rows646).slice(0, 300)}`);
  else {
    manifest.submissions[T646] = rows646[0].id;
    const fb = rows646[0].payload?.final_balance;
    ok(`submission_id = ${rows646[0].id} (APPROVED, server final_balance=${fb}, expected 80)`);
    if (fb !== 80) bad(`646 server balance wrong: ${fb} (expected 80)`);
  }
  const led646 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows646[0]?.id}`);
  console.log(`       ledger rows for 646 submission: ${JSON.stringify(led646)}`);
  manifest.ledger.push(...led646.map((l) => l.id));

  // --- 9. Task 647: disposable test video upload + submit (PENDING) ---
  console.log('\n[9] Task 647 skill video (upload + PENDING submit)');
  const ftyp = Buffer.from([
    0x00, 0x00, 0x00, 0x18, 0x66, 0x74, 0x79, 0x70, 0x69, 0x73, 0x6f, 0x6d,
    0x00, 0x00, 0x00, 0x00, 0x69, 0x73, 0x6f, 0x6d, 0x69, 0x73, 0x6f, 0x32,
  ]);
  const form647 = new FormData();
  form647.append('file', new Blob([ftyp], { type: 'video/mp4' }), 'SMOKE-TEST-skill.mp4');
  const up647 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form647 });
  if (up647.status !== 200 || !up647.json?.file_url) { bad(`647 upload failed: ${up647.status} ${up647.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up647.json.storage_path);
    ok(`uploaded storage_path = ${up647.json.storage_path}`);
    const chk647 = await fetch(up647.json.file_url, { method: 'HEAD' });
    if (chk647.ok) ok(`storage file reachable via file_url (HEAD ${chk647.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk647.status})`);
    await sleep(3000);
    const s647 = await prodApi(`/api/tasks/${T647}/submit`, {
      method: 'POST', token,
      body: { payload: { file_url: up647.json.file_url, file_name: up647.json.file_name, file_size: up647.json.file_size, duration_seconds: 45, mime_type: 'video/mp4', note: 'SMOKE-TEST: video uji disposable (bukan skill asli), mohon abaikan', captured_at: new Date().toISOString() } },
    });
    if (s647.status !== 200 || s647.json?.success !== true) bad(`647 submit failed: ${s647.status} ${s647.text.slice(0, 160)}`);
    else ok('647 submit accepted (manual review path)');
    const rows647 = await dbGet(`odyssey_task_submissions?task_id=eq.${T647}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows647.length !== 1 || rows647[0].status !== 'PENDING') bad(`647 submission state wrong: ${JSON.stringify(rows647).slice(0, 300)}`);
    else { manifest.submissions[T647] = rows647[0].id; ok(`submission_id = ${rows647[0].id} (PENDING, file_url present: ${!!rows647[0].payload?.file_url})`); }
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
  else ok('exactly 6 submission artifacts (642-647)');
  const pendCount = allSubs.filter((s) => s.status === 'PENDING').length;
  const apprCount = allSubs.filter((s) => s.status === 'APPROVED').length;
  if (pendCount === 3 && apprCount === 3) ok('3 PENDING (ADMIN_REVIEW) + 3 APPROVED (AUTO) as designed');
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
  const t918 = await dbGet('odyssey_tasks?active_date=eq.2026-09-18&is_active=eq.true&select=id&order=id.asc');
  const t919 = await dbGet('odyssey_tasks?active_date=eq.2026-09-19&is_active=eq.true&select=id&order=id.asc');
  if (JSON.stringify(t918.map((x) => x.id)) !== '[642,643,644]') bad(`09-18 tasks changed: ${JSON.stringify(t918)}`);
  else ok('tasks 642/643/644 intact (exactly 3 active on 2026-09-18)');
  if (JSON.stringify(t919.map((x) => x.id)) !== '[645,646,647]') bad(`09-19 tasks changed: ${JSON.stringify(t919)}`);
  else ok('tasks 645/646/647 intact (exactly 3 active on 2026-09-19)');

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
