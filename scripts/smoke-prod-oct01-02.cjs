require('dotenv').config();

// Real authenticated production smoke for 2026-10-01 + 2026-10-02 tasks
// from Migration 098. Task IDs are resolved at runtime by exact
// title+date (no hardcoded ID assumptions).
// Strategy: disposable test user -> real flows -> exact-ID manifest ->
// exact-ID cleanup -> post-cleanup verification. No broad deletes.
// No redeem/claims. T1 MINI_GAME (AUTO): invalid rejected, spoof ignored,
// valid APPROVED (+50c/+100xp, server final_balance 120). T2 PHOTO:
// real 1px png + 3-part note -> PENDING. T3 TEXT: 50ch rejected (P0008),
// >=200ch 3-part text -> PENDING. T4 VIDEO: real playable 5s MP4 fixture
// (ffmpeg-generated, h264+aac, 59KB) -> PENDING. T5 MINI_GAME (AUTO):
// invalid -> 400, valid -> APPROVED (+50c/+100xp, balance 120).
// T6 DOCUMENT: real .txt checklist file -> PENDING.
// NOTE: keep the disposable user on the default monthly_coin_target=0 so
// AUTO approvals yield v_actual=0 and create NO ledger rows (odyssey_coin_
// transactions is append-only, P0012 rejects DELETE; a coin-bearing probe
// would leave an indelible ledger row + undeletable profile).

const fs = require('fs');
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
const TEST_EXPLORER = `Smoke Oct01-02 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const TITLES_01 = [
  'Kalau Mau Mulai Lagi, Pilih yang Mana?',
  'Sebelum & Sesudah Awal Bulan',
  'Satu Hal yang Mau Kamu Benerin Bulan Ini',
];
const TITLES_02 = [
  'Tunjukkan Satu Trik yang Sering Kamu Pakai',
  'Sebelum Dipakai, Cek Dulu',
  'Bikin Checklist yang Bisa Langsung Dipakai',
];

const GAME_CHOICES_6 = { ev_1: 'a', ev_2: 'a', ev_3: 'a', ev_4: 'a', ev_5: 'a', ev_6: 'a' };
const EXPECTED_BALANCE = 120; // 6 events x delta 20

const MP4_FIXTURE = 'C:\\Users\\user\\AppData\\Local\\Temp\\opencode\\smoke-trick-5s.mp4';

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
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-10-01/02 (098)');
  console.log(` disposable user: ${TEST_USER} / ${TEST_UID}`);
  console.log('====================================================\n');

  console.log('[0] Collision check (must be empty)');
  const c1 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const c2 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  if (c1.length || c2.length) { bad('test identity collides with existing user — ABORT'); process.exit(1); }
  ok('no collision for disposable identity');

  console.log('\n[0b] Resolve task IDs (exact title + date, 3 active per date)');
  const r01 = await dbGet('odyssey_tasks?active_date=eq.2026-10-01&is_active=eq.true&select=id,title,step_order&order=step_order.asc');
  const r02 = await dbGet('odyssey_tasks?active_date=eq.2026-10-02&is_active=eq.true&select=id,title,step_order&order=step_order.asc');
  if (r01.length !== 3 || JSON.stringify(r01.map((t) => t.title)) !== JSON.stringify(TITLES_01)) {
    bad(`10-01 task set wrong: ${JSON.stringify(r01.map((t) => t.title))}`); process.exit(1);
  }
  if (r02.length !== 3 || JSON.stringify(r02.map((t) => t.title)) !== JSON.stringify(TITLES_02)) {
    bad(`10-02 task set wrong: ${JSON.stringify(r02.map((t) => t.title))}`); process.exit(1);
  }
  const [T1, T2, T3] = r01.map((t) => t.id);
  const [T4, T5, T6] = r02.map((t) => t.id);
  console.log(`   T1 MINI_GAME=${T1} T2 PHOTO=${T2} T3 TEXT=${T3} T4 VIDEO=${T4} T5 MINI_GAME=${T5} T6 DOCUMENT=${T6}`);
  ok('resolved 6 task IDs by exact title+date');

  console.log('\n[1] Create disposable test user');
  let r = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
    method: 'POST', headers: { ...dbHeaders, Prefer: 'return=representation' },
    body: JSON.stringify([{ uid: TEST_UID, username: TEST_USER, password_hash: PASS_HASH, explorer_name: TEST_EXPLORER, role: 'MEMBER', family_id: 'demo-crew-1', coins: 0, xp: 0, level: 1, is_active: true }]),
  });
  if (!r.ok) { bad(`create user failed: ${r.status} ${await r.text()}`); process.exit(1); }
  manifest.user = TEST_UID;
  ok(`TEST USER user_id = ${TEST_UID}`);

  console.log('\n[2] Authenticated member login');
  await sleep(6000);
  let login = await prodApi('/api/login', { method: 'POST', body: { uid: TEST_USER, login_method: 'PASSWORD', credential: 'admin123', device: { device_id: DEVICE } } });
  if (login.status !== 200 || !login.json?.session) { bad(`member login failed: ${login.status} ${login.text.slice(0, 200)}`); process.exit(1); }
  const token = login.json.session;
  manifest.session = true;
  ok(`login ok (token ${String(token).slice(0, 8)}…, len ${String(token).length})`);

  console.log('\n[3] Open task details (must be DB-rendered)');
  await sleep(3000);
  const t = {};
  for (const id of [T1, T2, T3, T4, T5, T6]) {
    t[id] = await prodApi(`/api/tasks/${id}`, { token });
    if (t[id].status !== 200) bad(`open ${id} failed: ${t[id].status}`);
  }
  const c1cfg = t[T1].json?.config || {};
  if (t[T1].json?.task_type === 'MINI_GAME' && t[T1].json?.evaluation_type === 'AUTO' && c1cfg.game === 'DECISION_PRIORITY' && c1cfg.scenario?.events?.length === 6 && t[T1].json?.reward_coins === 50 && t[T1].json?.reward_xp === 100) ok('T1 MINI_GAME DECISION_PRIORITY 6 events AUTO 50c/100xp (DB config)');
  else bad(`T1 config mismatch: ${JSON.stringify(c1cfg).slice(0, 200)}`);
  if (!/correct_answer|answer_key|expected_answer|is_correct/.test(JSON.stringify(c1cfg))) ok('T1 no answer-key fields leak through member API');
  else bad('T1 answer-key leakage through member API');
  const c2cfg = t[T2].json?.config || {};
  if (t[T2].json?.task_type === 'PHOTO_UPLOAD' && t[T2].json?.evaluation_type === 'ADMIN_REVIEW' && c2cfg.max_files === 1 && c2cfg.camera_only === false && /before -> action -> after/.test(c2cfg.instruction || '') && t[T2].json?.reward_coins === 40) ok('T2 PHOTO max_files=1 camera_only=false ADMIN_REVIEW + before->action->after 40c/100xp (DB config)');
  else bad(`T2 config mismatch: ${JSON.stringify(c2cfg).slice(0, 200)}`);
  const c3cfg = t[T3].json?.config || {};
  if (t[T3].json?.task_type === 'TEXT_RESPONSE' && t[T3].json?.evaluation_type === 'ADMIN_REVIEW' && c3cfg.minimum_characters === 200 && c3cfg.maximum_characters === 1000 && t[T3].json?.reward_coins === 40) ok('T3 TEXT bounds 200/1000 ADMIN_REVIEW 40c/100xp (DB config)');
  else bad(`T3 config mismatch: ${JSON.stringify(c3cfg).slice(0, 200)}`);
  const c4cfg = t[T4].json?.config || {};
  if (t[T4].json?.task_type === 'VIDEO' && t[T4].json?.evaluation_type === 'ADMIN_REVIEW' && c4cfg.recording?.enabled === true && c4cfg.recording?.max_duration_seconds === 60 && /DEMONSTRASIKAN/.test(c4cfg.recording?.instruction || '') && t[T4].json?.reward_coins === 40) ok('T4 VIDEO recording enabled max 60s ADMIN_REVIEW + demo instruction 40c/100xp (DB config)');
  else bad(`T4 config mismatch: ${JSON.stringify(c4cfg).slice(0, 200)}`);
  const c5cfg = t[T5].json?.config || {};
  if (t[T5].json?.task_type === 'MINI_GAME' && t[T5].json?.evaluation_type === 'AUTO' && c5cfg.game === 'DECISION_PRIORITY' && c5cfg.scenario?.events?.length === 6 && t[T5].json?.reward_coins === 50) ok('T5 MINI_GAME DECISION_PRIORITY (not ANOMALY) 6 events AUTO 50c/100xp (DB config)');
  else bad(`T5 config mismatch: ${JSON.stringify(c5cfg).slice(0, 200)}`);
  if (!/janggal/i.test(JSON.stringify(c5cfg))) ok('T5 no anomaly wording in scenario');
  else bad('T5 contains anomaly wording');
  const c6cfg = t[T6].json?.config || {};
  if (t[T6].json?.task_type === 'DOCUMENT_UPLOAD' && t[T6].json?.evaluation_type === 'ADMIN_REVIEW' && Array.isArray(c6cfg.allowed_extensions) && Array.isArray(c6cfg.accepted_extensions) && c6cfg.max_file_size_mb === 4 && /Maksimal 10 checklist items/.test(c6cfg.instruction || '') && t[T6].json?.reward_coins === 40) ok('T6 DOCUMENT dual-key ext + 4MB + checklist<=10 ADMIN_REVIEW 40c/100xp (DB config)');
  else bad(`T6 config mismatch: ${JSON.stringify(c6cfg).slice(0, 200)}`);

  const pngB64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';
  const pngBytes = Buffer.from(pngB64, 'base64');

  console.log('\n[4] Task T1 MINI_GAME (invalid rejected, spoof ignored, valid APPROVED)');
  await sleep(3000);
  const bad1 = await prodApi(`/api/tasks/${T1}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: { ev_1: 'zzz' }, final_balance: 999999, score: 100 } } });
  if (bad1.status === 200 && bad1.json?.success === true) bad('T1 accepted spoofed/invalid choices — server recompute BYPASSED');
  else ok(`T1 rejected invalid/spoofed choices (status ${bad1.status})`);
  await sleep(3000);
  const s1 = await prodApi(`/api/tasks/${T1}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: GAME_CHOICES_6, final_balance: 0, score: 0 } } });
  if (s1.status !== 200 || s1.json?.success !== true) bad(`T1 valid submit failed: ${s1.status} ${s1.text.slice(0, 200)}`);
  else ok(`T1 valid choices APPROVED despite spoofed score=0/balance=0 (server recompute authoritative, coins_earned=${s1.json?.coins_earned})`);
  const rows1 = await dbGet(`odyssey_task_submissions?task_id=eq.${T1}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows1.length !== 1 || rows1[0].status !== 'APPROVED') bad(`T1 submission state wrong: ${JSON.stringify(rows1).slice(0, 300)}`);
  else {
    manifest.submissions[T1] = rows1[0].id;
    const fb = rows1[0].payload?.final_balance;
    ok(`submission_id = ${rows1[0].id} (APPROVED, server final_balance=${fb}, expected ${EXPECTED_BALANCE})`);
    if (fb !== EXPECTED_BALANCE) bad(`T1 server balance wrong: ${fb} (expected ${EXPECTED_BALANCE})`);
  }
  const led1 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows1[0]?.id}`);
  console.log(`       ledger rows for T1 submission: ${JSON.stringify(led1)}`);
  if (led1.length === 1 && led1[0].amount === 50) ok('T1 ledger TASK_REWARD +50');
  else bad(`T1 ledger wrong: ${JSON.stringify(led1)}`);
  manifest.ledger.push(...led1.map((l) => l.id));

  console.log('\n[5] Task T2 photo (upload + PENDING submit, 3-part note)');
  const form2 = new FormData();
  form2.append('file', new Blob([pngBytes], { type: 'image/png' }), 'SMOKE-TEST-hasil-akhir.png');
  const up2 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form2 });
  if (up2.status !== 200 || !up2.json?.file_url) { bad(`T2 upload failed: ${up2.status} ${up2.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up2.json.storage_path);
    ok(`uploaded storage_path = ${up2.json.storage_path}`);
    const chk = await fetch(up2.json.file_url, { method: 'HEAD' });
    if (chk.ok) ok(`storage file reachable via file_url (HEAD ${chk.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk.status})`);
    await sleep(3000);
    const note2 = 'SMOKE-TEST (data uji, akan dihapus). (a) Apa: contoh perubahan uji — merapikan meja contoh yang sebelumnya berantakan contoh. (b) Kenapa: contoh alasan uji — agar mudah mencari barang contoh. (c) Manfaat: contoh manfaat uji — bekerja lebih fokus contoh.';
    const s2 = await prodApi(`/api/tasks/${T2}/submit`, { method: 'POST', token, body: { payload: { file_url: up2.json.file_url, file_name: up2.json.file_name, file_size: up2.json.file_size, note: note2, submitted_at: new Date().toISOString() } } });
    if (s2.status !== 200 || s2.json?.success !== true) bad(`T2 submit failed: ${s2.status} ${s2.text.slice(0, 160)}`);
    else ok('T2 submit accepted (manual review path)');
    const rows2 = await dbGet(`odyssey_task_submissions?task_id=eq.${T2}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows2.length !== 1 || rows2[0].status !== 'PENDING') bad(`T2 submission state wrong: ${JSON.stringify(rows2).slice(0, 300)}`);
    else { manifest.submissions[T2] = rows2[0].id; ok(`submission_id = ${rows2[0].id} (PENDING, file_url: ${!!rows2[0].payload?.file_url}, note: ${!!rows2[0].payload?.note})`); }
  }

  console.log('\n[6] Task T3 text (50ch rejected P0008, >=200ch 3-part PENDING)');
  await sleep(3000);
  const short3 = await prodApi(`/api/tasks/${T3}/submit`, { method: 'POST', token, body: { payload: { text: 'SMOKE-TEST pendek contoh data uji', submitted_at: new Date().toISOString() } } });
  if (short3.status === 200 && short3.json?.success === true) bad('T3 accepted too-short text — min-length validation BYPASSED');
  else if (/minimal 200/.test(short3.text)) ok('T3 server rejected short text with min-200 message (P0008, DB-configured validation)');
  else bad(`T3 short-text rejection unclear: ${short3.status} ${short3.text.slice(0, 160)}`);
  const text3 = 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Kondisi sekarang: contoh kondisi uji — kebiasaan menunda pekerjaan contoh yang membuat tugas menumpuk contoh dan menimbulkan stres contoh setiap minggu contoh. 2) Perubahan Oktober: contoh target uji — menyelesaikan setiap tugas di hari yang sama contoh secara konsisten contoh sampai akhir Oktober contoh. 3) Langkah pertama: contoh langkah uji — membuat daftar tugas harian contoh mulai besok pagi contoh dan mencentang satu per satu contoh.';
  console.log(`       evidence len = ${text3.length} (bounds 200/1000)`);
  await sleep(3000);
  const s3 = await prodApi(`/api/tasks/${T3}/submit`, { method: 'POST', token, body: { payload: { text: text3, submitted_at: new Date().toISOString() } } });
  if (s3.status !== 200 || s3.json?.success !== true) bad(`T3 submit failed: ${s3.status} ${s3.text.slice(0, 160)}`);
  else ok('T3 submit accepted (manual review path)');
  const rows3 = await dbGet(`odyssey_task_submissions?task_id=eq.${T3}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows3.length !== 1 || rows3[0].status !== 'PENDING') bad(`T3 submission state wrong: ${JSON.stringify(rows3).slice(0, 300)}`);
  else { manifest.submissions[T3] = rows3[0].id; ok(`submission_id = ${rows3[0].id} (PENDING, text present: ${!!rows3[0].payload?.text})`); }

  console.log('\n[7] Task T4 video (real playable MP4 fixture + PENDING submit)');
  if (!fs.existsSync(MP4_FIXTURE)) { bad(`MP4 fixture missing: ${MP4_FIXTURE}`); }
  else {
    const mp4Bytes = fs.readFileSync(MP4_FIXTURE);
    console.log(`       fixture bytes = ${mp4Bytes.length} (expect 60672)`);
    const form4 = new FormData();
    form4.append('file', new Blob([mp4Bytes], { type: 'video/mp4' }), 'SMOKE-TEST-trik.mp4');
    const up4 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form4 });
    if (up4.status !== 200 || !up4.json?.file_url) { bad(`T4 upload failed: ${up4.status} ${up4.text.slice(0, 160)}`); }
    else {
      manifest.storage.push(up4.json.storage_path);
      ok(`uploaded storage_path = ${up4.json.storage_path}`);
      const chk4 = await fetch(up4.json.file_url, { method: 'HEAD' });
      if (chk4.ok) ok(`storage file reachable via file_url (HEAD ${chk4.status})`);
      else bad(`storage file NOT reachable (HEAD ${chk4.status})`);
      await sleep(3000);
      const s4 = await prodApi(`/api/tasks/${T4}/submit`, {
        method: 'POST', token,
        body: { payload: { file_url: up4.json.file_url, file_name: up4.json.file_name, file_size: up4.json.file_size, duration_seconds: 5, mime_type: 'video/mp4', note: 'SMOKE-TEST: video uji disposable 5 detik (fixture sintetis, bukan demo asli), mohon abaikan', captured_at: new Date().toISOString() } },
      });
      if (s4.status !== 200 || s4.json?.success !== true) bad(`T4 submit failed: ${s4.status} ${s4.text.slice(0, 160)}`);
      else ok('T4 submit accepted (manual review path, duration 5s within 60s max)');
      const rows4 = await dbGet(`odyssey_task_submissions?task_id=eq.${T4}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
      if (rows4.length !== 1 || rows4[0].status !== 'PENDING') bad(`T4 submission state wrong: ${JSON.stringify(rows4).slice(0, 300)}`);
      else { manifest.submissions[T4] = rows4[0].id; ok(`submission_id = ${rows4[0].id} (PENDING, file_url present: ${!!rows4[0].payload?.file_url})`); }
    }
  }

  console.log('\n[8] Task T5 MINI_GAME (invalid -> 400, valid -> APPROVED +50)');
  await sleep(3000);
  const bad5 = await prodApi(`/api/tasks/${T5}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: { ev_1: 'zzz' }, final_balance: 999999, score: 100 } } });
  if (bad5.status === 200 && bad5.json?.success === true) bad('T5 accepted spoofed/invalid choices — server recompute BYPASSED');
  else ok(`T5 rejected invalid/spoofed choices (status ${bad5.status})`);
  await sleep(3000);
  const s5 = await prodApi(`/api/tasks/${T5}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: GAME_CHOICES_6, final_balance: 0, score: 0 } } });
  if (s5.status !== 200 || s5.json?.success !== true) bad(`T5 valid submit failed: ${s5.status} ${s5.text.slice(0, 200)}`);
  else ok(`T5 valid choices APPROVED despite spoofed score=0/balance=0 (server recompute authoritative, coins_earned=${s5.json?.coins_earned})`);
  const rows5 = await dbGet(`odyssey_task_submissions?task_id=eq.${T5}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows5.length !== 1 || rows5[0].status !== 'APPROVED') bad(`T5 submission state wrong: ${JSON.stringify(rows5).slice(0, 300)}`);
  else {
    manifest.submissions[T5] = rows5[0].id;
    const fb = rows5[0].payload?.final_balance;
    ok(`submission_id = ${rows5[0].id} (APPROVED, server final_balance=${fb}, expected ${EXPECTED_BALANCE})`);
    if (fb !== EXPECTED_BALANCE) bad(`T5 server balance wrong: ${fb} (expected ${EXPECTED_BALANCE})`);
  }
  const led5 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows5[0]?.id}`);
  if (led5.length === 1 && led5[0].amount === 50) ok('T5 ledger TASK_REWARD +50');
  else bad(`T5 ledger wrong: ${JSON.stringify(led5)}`);
  manifest.ledger.push(...led5.map((l) => l.id));

  console.log('\n[9] Task T6 document (real .txt checklist + PENDING submit)');
  const checklist = 'SMOKE-TEST CHECKLIST (data uji, akan dihapus)\nChecklist sebelum berangkat kerja:\n[ ] Cek dompet dan kunci sudah dibawa contoh\n[ ] Cek HP sudah terisi daya contoh\n[ ] Cek bekal sudah dikemas contoh\n[ ] Cek pintu terkunci contoh\n[ ] Cek kompor mati contoh\n';
  const form6 = new FormData();
  form6.append('file', new Blob([Buffer.from(checklist, 'utf8')], { type: 'text/plain' }), 'SMOKE-TEST-checklist.txt');
  const up6 = await prodApi('/api/tasks/upload', { method: 'POST', token, form: form6 });
  if (up6.status !== 200 || !up6.json?.file_url) { bad(`T6 upload failed: ${up6.status} ${up6.text.slice(0, 160)}`); }
  else {
    manifest.storage.push(up6.json.storage_path);
    ok(`uploaded storage_path = ${up6.json.storage_path}`);
    const chk6 = await fetch(up6.json.file_url, { method: 'HEAD' });
    if (chk6.ok) ok(`storage file reachable via file_url (HEAD ${chk6.status})`);
    else bad(`storage file NOT reachable (HEAD ${chk6.status})`);
    await sleep(3000);
    const s6 = await prodApi(`/api/tasks/${T6}/submit`, { method: 'POST', token, body: { payload: { file_url: up6.json.file_url, file_name: up6.json.file_name, file_size: up6.json.file_size, note: 'SMOKE-TEST: checklist uji disposable (checklist sebelum berangkat kerja, 5 item), mohon abaikan', submitted_at: new Date().toISOString() } } });
    if (s6.status !== 200 || s6.json?.success !== true) bad(`T6 submit failed: ${s6.status} ${s6.text.slice(0, 160)}`);
    else ok('T6 submit accepted (manual review path)');
    const rows6 = await dbGet(`odyssey_task_submissions?task_id=eq.${T6}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
    if (rows6.length !== 1 || rows6[0].status !== 'PENDING') bad(`T6 submission state wrong: ${JSON.stringify(rows6).slice(0, 300)}`);
    else { manifest.submissions[T6] = rows6[0].id; ok(`submission_id = ${rows6[0].id} (PENDING, file_url present: ${!!rows6[0].payload?.file_url})`); }
  }

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
  if (pendCount === 4 && apprCount === 2) ok('4 PENDING (ADMIN_REVIEW) + 2 APPROVED (AUTO mini-games) as designed');
  else bad(`status split wrong: PENDING=${pendCount} APPROVED=${apprCount}`);
  if (claims.length) bad('unexpected claim rows');
  else ok('no claim artifacts (nothing redeemed)');
  const prof = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid,coins,xp`);
  console.log(`       profile: ${JSON.stringify(prof)}`);

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

  console.log('\n[12] Post-cleanup verification');
  const r1 = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id`);
  const r2 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const r3 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  const r4 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id`);
  const r5 = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  if (r1.length || r2.length || r3.length || r4.length || r5.length) bad(`residual artifacts: subs=${r1.length} uid=${r2.length} uname=${r3.length} ledger=${r4.length} claims=${r5.length}`);
  else ok('zero remaining test artifacts (submissions/profile/ledger/claims)');
  const t101 = await dbGet('odyssey_tasks?active_date=eq.2026-10-01&is_active=eq.true&select=id,title&order=step_order.asc');
  const t102 = await dbGet('odyssey_tasks?active_date=eq.2026-10-02&is_active=eq.true&select=id,title&order=step_order.asc');
  if (t101.length === 3 && JSON.stringify(t101.map((x) => x.title)) === JSON.stringify(TITLES_01)) ok(`10-01 intact (exactly 3 active: ${t101.map((x) => x.id).join(',')})`);
  else bad(`10-01 tasks changed: ${JSON.stringify(t101)}`);
  if (t102.length === 3 && JSON.stringify(t102.map((x) => x.title)) === JSON.stringify(TITLES_02)) ok(`10-02 intact (exactly 3 active: ${t102.map((x) => x.id).join(',')})`);
  else bad(`10-02 tasks changed: ${JSON.stringify(t102)}`);

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
