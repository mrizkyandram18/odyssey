require('dotenv').config();

// Real authenticated production smoke for 2026-10-12..16 tasks
// from Migration 100. Task IDs are resolved at runtime by exact
// title+date (no hardcoded ID assumptions).
// Strategy: disposable test user -> real flows -> exact-ID manifest ->
// exact-ID cleanup -> post-cleanup verification. No broad deletes.
// No redeem/claims. MINI_GAME x3 (AUTO): invalid rejected, spoof
// ignored, valid APPROVED (server final_balance 120). PHOTO x3 /
// VIDEO x2 / DOCUMENT x3: real uploads + PENDING submits. TEXT x4:
// short text rejected (P0008), >=150ch 3-part text -> PENDING.
// NOTE: the disposable user keeps the default monthly_coin_target=0,
// so AUTO approvals yield coins_earned=0 (v_actual=0, proportional
// share of target 0) and create NO ledger rows (odyssey_coin_
// transactions is append-only, P0012 rejects DELETE; a coin-bearing
// probe would leave an indelible ledger row). coins_earned=0 here is
// EXPECTED engine semantics (target 0 = no target), NOT a bug.
// XP is still awarded (+100 per AUTO approval).

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
const TEST_EXPLORER = `Smoke Oct12-16 ${ts}`;
const DEVICE = `smoke-dev-${ts}`;
const PASS_HASH = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123 (repo smoke pattern)

const PLAN = [
  { date: '2026-10-12', titles: ['Rekam Satu Detail Kecil lewat Foto', 'Tentukan Langkah Paling Tepat', 'Uraikan Satu Hal dengan Bahasamu Sendiri'], types: ['PHOTO_UPLOAD', 'MINI_GAME', 'TEXT_RESPONSE'], coins: [20, 25, 20] },
  { date: '2026-10-13', titles: ['Peragakan Cara Memakai Satu Benda', 'Temukan Fungsi Lain Satu Barang', 'Susun Lembar Panduan untuk Teman Baru'], types: ['VIDEO', 'TEXT_RESPONSE', 'DOCUMENT_UPLOAD'], coins: [20, 20, 20] },
  { date: '2026-10-14', titles: ['Timbang Dua Opsi lalu Pilih Satu', 'Atasi Keterbatasan dengan yang Ada', 'Tulis Lembar Catatan yang Bisa Dibuka Lagi'], types: ['TEXT_RESPONSE', 'MINI_GAME', 'DOCUMENT_UPLOAD'], coins: [20, 25, 20] },
  { date: '2026-10-15', titles: ['Ajarkan Satu Keterampilan Sederhana', 'Abadikan Satu Perubahan Kecil Hari Ini', 'Bagikan Satu Pengalaman dan Pelajaran Praktis'], types: ['VIDEO', 'PHOTO_UPLOAD', 'TEXT_RESPONSE'], coins: [20, 20, 20] },
  { date: '2026-10-16', titles: ['Tuangkan Satu Ide Jadi Rencana Singkat', 'Pilih Tanggapan Terbaik Sehari-hari', 'Pamerkan Satu Hasil Buatan atau Perbaikanmu'], types: ['DOCUMENT_UPLOAD', 'MINI_GAME', 'PHOTO_UPLOAD'], coins: [20, 25, 20] },
];

const GAME_CHOICES_6 = { ev_1: 'a', ev_2: 'a', ev_3: 'a', ev_4: 'a', ev_5: 'a', ev_6: 'a' };
const EXPECTED_BALANCE = 120; // 6 events x delta 20

const MP4_FIXTURE = 'C:\\Users\\user\\AppData\\Local\\Temp\\opencode\\smoke-skill-5s.mp4';

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

const pngB64 = 'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==';

async function submitMiniGame(token, taskId, label) {
  const badR = await prodApi(`/api/tasks/${taskId}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: { ev_1: 'zzz' }, final_balance: 999999, score: 100 } } });
  if (badR.status === 200 && badR.json?.success === true) bad(`${label} accepted spoofed/invalid choices — server recompute BYPASSED`);
  else ok(`${label} rejected invalid/spoofed choices (status ${badR.status})`);
  await sleep(2500);
  const s = await prodApi(`/api/tasks/${taskId}/submit`, { method: 'POST', token, body: { answers: { game: 'DECISION_PRIORITY', choices: GAME_CHOICES_6, final_balance: 0, score: 0 } } });
  if (s.status !== 200 || s.json?.success !== true) { bad(`${label} valid submit failed: ${s.status} ${s.text.slice(0, 200)}`); return; }
  if (s.json?.coins_earned !== 0) bad(`${label} coins_earned=${s.json?.coins_earned}, expected 0 (target-0 semantics)`);
  else ok(`${label} valid choices APPROVED despite spoofed score=0/balance=0 (server recompute authoritative, coins_earned=0 as expected for target 0)`);
  const rows = await dbGet(`odyssey_task_submissions?task_id=eq.${taskId}&user_uid=eq.${TEST_UID}&select=id,status,payload,coins_earned,xp_earned`);
  if (rows.length !== 1 || rows[0].status !== 'APPROVED') bad(`${label} submission state wrong: ${JSON.stringify(rows).slice(0, 300)}`);
  else {
    manifest.submissions[taskId] = rows[0].id;
    const fb = rows[0].payload?.final_balance;
    ok(`submission_id = ${rows[0].id} (APPROVED, server final_balance=${fb}, expected ${EXPECTED_BALANCE}, xp_earned=${rows[0].xp_earned})`);
    if (fb !== EXPECTED_BALANCE) bad(`${label} server balance wrong: ${fb} (expected ${EXPECTED_BALANCE})`);
    if (rows[0].coins_earned !== 0) bad(`${label} DB coins_earned=${rows[0].coins_earned}, expected 0`);
  }
  const led = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id&reference_id=eq.${rows[0]?.id}`);
  if (led.length === 0) ok(`${label} no ledger rows (expected: v_actual=0 inserts nothing)`);
  else bad(`${label} unexpected ledger rows: ${JSON.stringify(led)}`);
}

async function uploadFile(token, bytes, filename, mime) {
  const form = new FormData();
  form.append('file', new Blob([bytes], { type: mime }), filename);
  const up = await prodApi('/api/tasks/upload', { method: 'POST', token, form });
  if (up.status !== 200 || !up.json?.file_url) return { err: `upload failed: ${up.status} ${up.text.slice(0, 160)}` };
  manifest.storage.push(up.json.storage_path);
  const chk = await fetch(up.json.file_url, { method: 'HEAD' });
  return { up, reachable: chk.ok, headStatus: chk.status };
}

async function submitPhoto(token, taskId, label, note) {
  const pngBytes = Buffer.from(pngB64, 'base64');
  const { up, err, reachable, headStatus } = await uploadFile(token, pngBytes, 'SMOKE-TEST-detail.png', 'image/png');
  if (err) { bad(`${label} ${err}`); return; }
  ok(`uploaded storage_path = ${up.json.storage_path}`);
  if (reachable) ok(`storage file reachable via file_url (HEAD ${headStatus})`);
  else bad(`storage file NOT reachable (HEAD ${headStatus})`);
  await sleep(2500);
  const s = await prodApi(`/api/tasks/${taskId}/submit`, { method: 'POST', token, body: { payload: { file_url: up.json.file_url, file_name: up.json.file_name, file_size: up.json.file_size, note, submitted_at: new Date().toISOString() } } });
  if (s.status !== 200 || s.json?.success !== true) bad(`${label} submit failed: ${s.status} ${s.text.slice(0, 160)}`);
  else ok(`${label} submit accepted (manual review path)`);
  const rows = await dbGet(`odyssey_task_submissions?task_id=eq.${taskId}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows.length !== 1 || rows[0].status !== 'PENDING') bad(`${label} submission state wrong: ${JSON.stringify(rows).slice(0, 300)}`);
  else { manifest.submissions[taskId] = rows[0].id; ok(`submission_id = ${rows[0].id} (PENDING, file_url: ${!!rows[0].payload?.file_url}, note: ${!!rows[0].payload?.note})`); }
}

async function submitVideo(token, taskId, label) {
  if (!fs.existsSync(MP4_FIXTURE)) { bad(`MP4 fixture missing: ${MP4_FIXTURE}`); return; }
  const mp4Bytes = fs.readFileSync(MP4_FIXTURE);
  console.log(`       fixture bytes = ${mp4Bytes.length}`);
  const { up, err, reachable, headStatus } = await uploadFile(token, mp4Bytes, 'SMOKE-TEST-skill.mp4', 'video/mp4');
  if (err) { bad(`${label} ${err}`); return; }
  ok(`uploaded storage_path = ${up.json.storage_path}`);
  if (reachable) ok(`storage file reachable via file_url (HEAD ${headStatus})`);
  else bad(`storage file NOT reachable (HEAD ${headStatus})`);
  await sleep(2500);
  const s = await prodApi(`/api/tasks/${taskId}/submit`, {
    method: 'POST', token,
    body: { payload: { file_url: up.json.file_url, file_name: up.json.file_name, file_size: up.json.file_size, duration_seconds: 5, mime_type: 'video/mp4', note: 'SMOKE-TEST: video uji disposable 5 detik (fixture sintetis, bukan demo asli), mohon abaikan', captured_at: new Date().toISOString() } },
  });
  if (s.status !== 200 || s.json?.success !== true) bad(`${label} submit failed: ${s.status} ${s.text.slice(0, 160)}`);
  else ok(`${label} submit accepted (manual review path, duration 5s within 60s max)`);
  const rows = await dbGet(`odyssey_task_submissions?task_id=eq.${taskId}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows.length !== 1 || rows[0].status !== 'PENDING') bad(`${label} submission state wrong: ${JSON.stringify(rows).slice(0, 300)}`);
  else { manifest.submissions[taskId] = rows[0].id; ok(`submission_id = ${rows[0].id} (PENDING, file_url present: ${!!rows[0].payload?.file_url})`); }
}

async function submitText(token, taskId, label, validText, testShort) {
  if (testShort) {
    await sleep(2000);
    const short = await prodApi(`/api/tasks/${taskId}/submit`, { method: 'POST', token, body: { payload: { text: 'SMOKE-TEST pendek contoh data uji', submitted_at: new Date().toISOString() } } });
    if (short.status === 200 && short.json?.success === true) bad(`${label} accepted too-short text — min-length validation BYPASSED`);
    else if (/minimal 150/.test(short.text)) ok(`${label} server rejected short text with min-150 message (P0008, DB-configured validation)`);
    else bad(`${label} short-text rejection unclear: ${short.status} ${short.text.slice(0, 160)}`);
  }
  console.log(`       evidence len = ${validText.length} (bounds 150/800)`);
  await sleep(2500);
  const s = await prodApi(`/api/tasks/${taskId}/submit`, { method: 'POST', token, body: { payload: { text: validText, submitted_at: new Date().toISOString() } } });
  if (s.status !== 200 || s.json?.success !== true) bad(`${label} submit failed: ${s.status} ${s.text.slice(0, 160)}`);
  else ok(`${label} submit accepted (manual review path)`);
  const rows = await dbGet(`odyssey_task_submissions?task_id=eq.${taskId}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows.length !== 1 || rows[0].status !== 'PENDING') bad(`${label} submission state wrong: ${JSON.stringify(rows).slice(0, 300)}`);
  else { manifest.submissions[taskId] = rows[0].id; ok(`submission_id = ${rows[0].id} (PENDING, text present: ${!!rows[0].payload?.text})`); }
}

async function submitDoc(token, taskId, label, content, filename, note) {
  const { up, err, reachable, headStatus } = await uploadFile(token, Buffer.from(content, 'utf8'), filename, 'text/plain');
  if (err) { bad(`${label} ${err}`); return; }
  ok(`uploaded storage_path = ${up.json.storage_path}`);
  if (reachable) ok(`storage file reachable via file_url (HEAD ${headStatus})`);
  else bad(`storage file NOT reachable (HEAD ${headStatus})`);
  await sleep(2500);
  const s = await prodApi(`/api/tasks/${taskId}/submit`, { method: 'POST', token, body: { payload: { file_url: up.json.file_url, file_name: up.json.file_name, file_size: up.json.file_size, note, submitted_at: new Date().toISOString() } } });
  if (s.status !== 200 || s.json?.success !== true) bad(`${label} submit failed: ${s.status} ${s.text.slice(0, 160)}`);
  else ok(`${label} submit accepted (manual review path)`);
  const rows = await dbGet(`odyssey_task_submissions?task_id=eq.${taskId}&user_uid=eq.${TEST_UID}&select=id,status,payload`);
  if (rows.length !== 1 || rows[0].status !== 'PENDING') bad(`${label} submission state wrong: ${JSON.stringify(rows).slice(0, 300)}`);
  else { manifest.submissions[taskId] = rows[0].id; ok(`submission_id = ${rows[0].id} (PENDING, file_url present: ${!!rows[0].payload?.file_url})`); }
}

const TEXTS = {
  uraikan: 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Hal yang dijelaskan: contoh hal uji — kenapa kaca berembun contoh yang sering terlihat pagi hari contoh dan membingungkan contoh. 2) Perumpamaan: ibarat contoh cermin kamar mandi yang berkabut contoh setelah air panas contoh sehingga mudah dibayangkan contoh. 3) Contoh nyata: contoh pengalaman uji — melihat kaca jendela berembun contoh saat udara dingin contoh dan menetes contoh.',
  fungsi: 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Barang dan fungsi utama: contoh barang uji — handuk contoh yang fungsi utamanya mengeringkan badan contoh setelah mandi contoh. 2) Fungsi lain: contoh fungsi uji — dipakai sebagai alas darurat contoh saat mengganti sesuatu contoh dengan cara dilipat contoh lalu digelar contoh. 3) Batasan: contoh batasan uji — tidak cocok untuk menahan beban berat contoh karena bahannya lentur contoh dan mudah bergeser contoh.',
  timbang: 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Keputusan dan dua opsi: contoh keputusan uji — menyimpan barang contoh dengan opsi lemari contoh atau kardus contoh untuk kerapian contoh. 2) Perbandingan: contoh kriteria uji — dari sisi biaya contoh lemari gratis contoh dan dari sisi waktu contoh kardus lebih cepat contoh namun kurang rapi contoh. 3) Pilihan: contoh pilihan uji — memilih lemari contoh karena lebih rapi contoh dan tahan lama contoh untuk jangka panjang contoh.',
  pengalaman: 'SMOKE-TEST DATA (fiktif, akan dihapus). 1) Pengalaman: contoh pengalaman uji — mengatur sudut ruangan contoh yang berantakan contoh selama akhir pekan contoh sampai rapi contoh. 2) Pelajaran praktis: contoh pelajaran uji — mengelompokkan barang sejenis contoh membuat pencarian contoh jauh lebih cepat contoh setiap hari contoh. 3) Penerapan: contoh penerapan uji — akan merapikan satu laci contoh dalam 7 hari contoh dengan cara yang sama contoh dan mencatat hasilnya contoh.',
};

const DOCS = {
  panduan: 'SMOKE-TEST PANDUAN PEMULA (data uji, akan dihapus)\nPanduan mencuci sepatu untuk pemula:\nIstilah: sol adalah bagian bawah sepatu contoh; deterjen ringan contoh.\nLangkah: lepas tali sepatu contoh, sikat kotoran kering contoh, cuci dengan sabun ringan contoh, bilas secukupnya contoh, keringkan di tempat teduh contoh.\nKesalahan umum pemula: 1) Merendam terlalu lama contoh — sepatu jadi bau contoh. 2) Menjemur di terik langsung contoh — lem mengelupas contoh.\n',
  catatan: 'SMOKE-TEST CATATAN RUJUKAN (data uji, akan dihapus)\nCatatan langkah mencuci sepatu (rujukan pribadi):\nRingkasan: lepas tali contoh, sikat kering contoh, sabun ringan contoh, bilas contoh, teduhkan contoh.\nIngatan: jangan rendam lama contoh; jangan jemur terik contoh; simpan dengan ganjal contoh.\nDibuka lagi saat: setiap akan mencuci sepatu contoh, sekitar sebulan sekali contoh.\n',
  ide: 'SMOKE-TEST RENCANA IDE (data uji, akan dihapus)\nMasalah: contoh masalah uji — sudut ruangan berantakan contoh sehingga sulit mencari barang contoh setiap pagi contoh.\nIde: contoh ide uji — menata sudut dengan tiga wadah berlabel contoh agar tiap barang punya tempat contoh dan mudah dikembalikan contoh.\nLangkah pertama (7 hari): contoh langkah uji — membeli tiga wadah sederhana contoh dan memberi label contoh akhir pekan ini contoh.\n',
};

async function run() {
  console.log('====================================================');
  console.log('REAL AUTHENTICATED PROD SMOKE — 2026-10-12..16 (100)');
  console.log(` disposable user: ${TEST_USER} / ${TEST_UID}`);
  console.log('====================================================\n');

  console.log('[0] Collision check (must be empty)');
  const c1 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const c2 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  if (c1.length || c2.length) { bad('test identity collides with existing user — ABORT'); process.exit(1); }
  ok('no collision for disposable identity');

  console.log('\n[0b] Resolve task IDs (exact title + date, 3 active per date)');
  const IDS = {};
  for (const day of PLAN) {
    const rows = await dbGet(`odyssey_tasks?active_date=eq.${day.date}&is_active=eq.true&select=id,title,step_order,task_type,evaluation_type,reward_coins,reward_xp&order=step_order.asc`);
    if (rows.length !== 3 || JSON.stringify(rows.map((t) => t.title)) !== JSON.stringify(day.titles)) {
      bad(`${day.date} task set wrong: ${JSON.stringify(rows.map((t) => t.title))}`); process.exit(1);
    }
    rows.forEach((t, i) => {
      if (t.task_type !== day.types[i]) { bad(`${day.date} step${i + 1} type ${t.task_type} != ${day.types[i]}`); }
      if (t.reward_coins !== day.coins[i] || t.reward_xp !== 100) { bad(`${day.date} step${i + 1} reward ${t.reward_coins}c/${t.reward_xp}xp wrong`); }
      IDS[`${day.date}#${i + 1}`] = t.id;
    });
    console.log(`   ${day.date}: ${rows.map((t) => `[ID ${t.id}]`).join(' ')}`);
  }
  ok('resolved 15 task IDs by exact title+date, types+rewards match plan');

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

  console.log('\n[3] Open task details (must be DB-rendered, no answer-key leak)');
  await sleep(3000);
  const allIds = Object.values(IDS);
  const detail = {};
  for (const id of allIds) {
    detail[id] = await prodApi(`/api/tasks/${id}`, { token });
    if (detail[id].status !== 200) bad(`open ${id} failed: ${detail[id].status}`);
  }
  for (const id of allIds) {
    const cfg = detail[id].json?.config || {};
    if (/correct_answer|answer_key|expected_answer|is_correct/.test(JSON.stringify(cfg))) bad(`task ${id} answer-key leakage through member API`);
  }
  ok('no answer-key fields leak through member API (15/15)');
  const spot = [
    [IDS['2026-10-12#1'], 'PHOTO_UPLOAD', 'ADMIN_REVIEW', (c) => c.max_files === 1 && c.camera_only === false],
    [IDS['2026-10-12#2'], 'MINI_GAME', 'AUTO', (c) => c.game === 'DECISION_PRIORITY' && c.scenario?.events?.length === 6],
    [IDS['2026-10-12#3'], 'TEXT_RESPONSE', 'ADMIN_REVIEW', (c) => c.minimum_characters === 150 && c.maximum_characters === 800],
    [IDS['2026-10-13#1'], 'VIDEO', 'ADMIN_REVIEW', (c) => c.recording?.enabled === true && c.recording?.max_duration_seconds === 60],
    [IDS['2026-10-14#2'], 'MINI_GAME', 'AUTO', (c) => c.game === 'DECISION_PRIORITY' && c.scenario?.events?.length === 6],
    [IDS['2026-10-16#2'], 'MINI_GAME', 'AUTO', (c) => c.game === 'DECISION_PRIORITY' && c.scenario?.events?.length === 6],
    [IDS['2026-10-16#3'], 'PHOTO_UPLOAD', 'ADMIN_REVIEW', (c) => c.max_files === 1 && c.camera_only === false],
  ];
  for (const [id, ty, ev, fn] of spot) {
    const j = detail[id].json || {};
    if (j.task_type === ty && j.evaluation_type === ev && fn(j.config || {})) ok(`task ${id} ${ty}/${ev} DB config shape ok`);
    else bad(`task ${id} config mismatch: ${JSON.stringify(j.config).slice(0, 160)}`);
  }

  console.log('\n[4] MINI_GAME x3 (invalid rejected, spoof ignored, valid APPROVED)');
  await sleep(3000);
  await submitMiniGame(token, IDS['2026-10-12#2'], 'T12-MG');
  await sleep(2000);
  await submitMiniGame(token, IDS['2026-10-14#2'], 'T14-MG');
  await sleep(2000);
  await submitMiniGame(token, IDS['2026-10-16#2'], 'T16-MG');

  console.log('\n[5] PHOTO x3 (upload + PENDING, 3-part notes)');
  await submitPhoto(token, IDS['2026-10-12#1'], 'T12-PHOTO', 'SMOKE-TEST (data uji, akan dihapus). (a) Detail apa: contoh detail uji — pola aus contoh di bagian sudut contoh. (b) Pada benda apa: contoh benda uji — tas contoh yang dipakai harian contoh. (c) Kenapa sering terlewat: contoh alasan uji — karena kecil dan tersembunyi contoh sehingga jarang disadari contoh.');
  await submitPhoto(token, IDS['2026-10-15#2'], 'T15-PHOTO', 'SMOKE-TEST (data uji, akan dihapus). (a) Apa yang berubah: contoh perubahan uji — sudut contoh yang ditata ulang contoh hari ini contoh. (b) Sejak kapan: contoh waktu uji — disadari pagi ini contoh setelah melihat contoh. (c) Kenapa berarti: contoh makna uji — membuat mencari barang contoh lebih mudah contoh setiap hari contoh.');
  await submitPhoto(token, IDS['2026-10-16#3'], 'T16-PHOTO', 'SMOKE-TEST (data uji, akan dihapus). (a) Hasil apa: contoh hasil uji — susunan contoh buatan sendiri contoh hari ini contoh. (b) Cara singkat: contoh cara uji — mengelompokkan lalu menyusun contoh dalam satu jam contoh. (c) Guna sekarang: contoh guna uji — dipakai harian contoh agar rapi contoh dan mudah diambil contoh.');

  console.log('\n[6] TEXT x4 (short rejected once, valid PENDING)');
  await submitText(token, IDS['2026-10-12#3'], 'T12-TEXT', TEXTS.uraikan, true);
  await submitText(token, IDS['2026-10-13#2'], 'T13-TEXT', TEXTS.fungsi, false);
  await submitText(token, IDS['2026-10-14#1'], 'T14-TEXT', TEXTS.timbang, false);
  await submitText(token, IDS['2026-10-15#3'], 'T15-TEXT', TEXTS.pengalaman, false);

  console.log('\n[7] VIDEO x2 (real MP4 fixture + PENDING)');
  await submitVideo(token, IDS['2026-10-13#1'], 'T13-VIDEO');
  await submitVideo(token, IDS['2026-10-15#1'], 'T15-VIDEO');

  console.log('\n[8] DOCUMENT x3 (real .txt + PENDING)');
  await submitDoc(token, IDS['2026-10-13#3'], 'T13-DOC', DOCS.panduan, 'SMOKE-TEST-panduan.txt', 'SMOKE-TEST: lembar panduan pemula uji disposable, mohon abaikan');
  await submitDoc(token, IDS['2026-10-14#3'], 'T14-DOC', DOCS.catatan, 'SMOKE-TEST-catatan.txt', 'SMOKE-TEST: lembar catatan rujukan uji disposable, mohon abaikan');
  await submitDoc(token, IDS['2026-10-16#1'], 'T16-DOC', DOCS.ide, 'SMOKE-TEST-ide.txt', 'SMOKE-TEST: rencana ide uji disposable, mohon abaikan');

  console.log('\n[9] Artifact enumeration (ownership = disposable user only)');
  const allSubs = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id,task_id,status,coins_earned`);
  const ledgerAll = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id,amount,reference_id`);
  const claims = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  manifest.claims = claims;
  console.log(`       submissions: ${allSubs.length}, ledger: ${ledgerAll.length}, claims: ${claims.length}`);
  if (allSubs.length !== 15) bad(`expected exactly 15 submissions, found ${allSubs.length}`);
  else ok('exactly 15 submission artifacts (all tasks)');
  const pendCount = allSubs.filter((s) => s.status === 'PENDING').length;
  const apprCount = allSubs.filter((s) => s.status === 'APPROVED').length;
  if (pendCount === 12 && apprCount === 3) ok('12 PENDING (ADMIN_REVIEW) + 3 APPROVED (AUTO mini-games) as designed');
  else bad(`status split wrong: PENDING=${pendCount} APPROVED=${apprCount}`);
  if (allSubs.some((s) => (s.coins_earned || 0) !== 0)) bad('unexpected nonzero coins_earned (target-0 semantics expect 0)');
  else ok('all coins_earned=0 (target-0 semantics, no coin payout)');
  if (ledgerAll.length) bad(`unexpected ledger rows: ${JSON.stringify(ledgerAll)}`);
  else ok('no ledger artifacts (v_actual=0 inserts nothing; nothing to clean in append-only ledger)');
  if (claims.length) bad('unexpected claim rows');
  else ok('no claim artifacts (nothing redeemed)');
  const prof = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid,coins,xp`);
  console.log(`       profile: ${JSON.stringify(prof)}`);
  if (prof.length === 1 && prof[0].coins === 0 && prof[0].xp === 300) ok('profile coins=0/xp=300 (3 AUTO x 100xp, no coin payout as designed)');
  else bad(`profile state unexpected: ${JSON.stringify(prof)}`);

  console.log('\n[10] Exact-ID cleanup (only manifest artifacts)');
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

  console.log('\n[11] Post-cleanup verification');
  const r1 = await dbGet(`odyssey_task_submissions?user_uid=eq.${TEST_UID}&select=id`);
  const r2 = await dbGet(`odyssey_user_profiles?uid=eq.${TEST_UID}&select=uid`);
  const r3 = await dbGet(`odyssey_user_profiles?username=eq.${TEST_USER}&select=uid`);
  const r4 = await dbGet(`odyssey_coin_transactions?user_uid=eq.${TEST_UID}&select=id`);
  const r5 = await dbGet(`odyssey_claims?user_uid=eq.${TEST_UID}&select=id`);
  if (r1.length || r2.length || r3.length || r4.length || r5.length) bad(`residual artifacts: subs=${r1.length} uid=${r2.length} uname=${r3.length} ledger=${r4.length} claims=${r5.length}`);
  else ok('zero remaining test artifacts (submissions/profile/ledger/claims)');
  for (const day of PLAN) {
    const t = await dbGet(`odyssey_tasks?active_date=eq.${day.date}&is_active=eq.true&select=id,title&order=step_order.asc`);
    if (t.length === 3 && JSON.stringify(t.map((x) => x.title)) === JSON.stringify(day.titles)) ok(`${day.date} intact (exactly 3 active: ${t.map((x) => x.id).join(',')})`);
    else bad(`${day.date} tasks changed: ${JSON.stringify(t)}`);
  }
  const a610 = await dbGet('odyssey_tasks?id=eq.610&select=id,title,config');
  if (a610.length === 1 && a610[0].config?.camera_only === true) ok('anchor 610 camera_only=true intact');
  else bad(`anchor 610 changed: ${JSON.stringify(a610).slice(0, 200)}`);

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
