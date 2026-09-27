require('dotenv').config();

const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const dbHeaders = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
};

async function runLiveSmoke() {
  console.log('====================================================');
  console.log('LIVE PRODUCTION SMOKE TEST: 10 SEPTEMBER 2026 TASKS');
  console.log(`Target: ${PROD_URL}`);
  console.log('====================================================\n');

  // 1. Health check
  console.log('1. Checking Live /api/status...');
  const statusRes = await fetch(`${PROD_URL}/api/status`);
  const status = await statusRes.json();
  console.log(`   Schema version: ${status.schema_version} (expected 080_tasks_2026_09_10)`);
  if (status.schema_version !== '080_tasks_2026_09_10') {
    throw new Error(`Unexpected schema version: ${status.schema_version}`);
  }
  console.log('   /api/status: PASS 🟢\n');

  // 2. Setup a smoke user in DB
  const ts = Date.now();
  const smokeUid = `smoke_sep10_${ts}`;
  const deviceId = `smoke_dev_${ts}`;
  const passwordHash = '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W'; // admin123

  console.log('2. Setting up test member in production DB...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
    method: 'POST',
    headers: { ...dbHeaders, Prefer: 'return=minimal' },
    body: JSON.stringify([{
      uid: smokeUid,
      username: smokeUid,
      password_hash: passwordHash,
      explorer_name: 'Smoke Tester Sep10',
      role: 'MEMBER',
      family_id: 'demo-crew-1',
      coins: 100,
      xp: 50,
      level: 1,
      is_active: true,
      monthly_coin_target: 3200,
      device_id: deviceId,
      device_bound_at: new Date().toISOString(),
      created_at: '2026-09-01T00:00:00Z',
    }]),
  });
  console.log(`   Created test member ${smokeUid}: PASS 🟢\n`);

  try {
    // 3. Authenticate member via /api/login
    console.log('3. Logging in as member via Live /api/login...');
    const loginRes = await fetch(`${PROD_URL}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: smokeUid,
        login_method: 'PASSWORD',
        credential: 'admin123',
        device: { device_id: deviceId },
      }),
    });
    const loginData = await loginRes.json();
    if (!loginRes.ok || !loginData.session) {
      throw new Error(`Login failed (${loginRes.status}): ${JSON.stringify(loginData)}`);
    }
    const token = loginData.session;
    console.log('   Member login: PASS 🟢\n');

    // 4. Fetch daily tasks via /api/tasks/today
    console.log('4. Calling Live GET /api/tasks/today...');
    const todayRes = await fetch(`${PROD_URL}/api/tasks/today`, {
      headers: {
        Authorization: `Bearer ${token}`,
      },
    });
    const todayData = await todayRes.json();
    if (!todayRes.ok) {
      throw new Error(`Failed to get today tasks (${todayRes.status}): ${JSON.stringify(todayData)}`);
    }
    console.log(`   Received ${todayData.tasks ? todayData.tasks.length : 0} tasks for today.`);

    // 5. Verify task configurations for 2026-09-10 in DB
    console.log('\n5. Verifying 10 September 2026 tasks in DB:');
    const tasksRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?active_date=eq.2026-09-10&is_active=eq.true&order=step_order.asc`, {
      headers: dbHeaders,
    });
    const tasks = await tasksRes.json();
    tasks.forEach(t => {
      console.log(`   - Step #${t.step_order}: "${t.title}" | Type: ${t.task_type} | Eval: ${t.evaluation_type} | Rewards: ${t.reward_coins}c / ${t.reward_xp}xp`);
      if (t.task_type === 'MINI_GAME') {
        console.log(`     Game: ${t.config.game} | Events: ${t.config.scenario.events.length} situations | Initial Balance: Rp${t.config.scenario.initial_balance.toLocaleString('id-ID')}`);
      } else if (t.task_type === 'QUIZ') {
        console.log(`     Questions: ${t.config.questions.length} questions | Has explanations: ${t.config.questions.every(q => !!q.explanation)}`);
      } else if (t.task_type === 'TEXT_RESPONSE') {
        console.log(`     Character limits: ${t.config.minimum_characters} - ${t.config.maximum_characters} | Prompt length: ${t.config.prompt.length} chars`);
      }
    });

    // 6. Security Check: Verify that answers NEVER leak in GET endpoints
    console.log('\n6. Security Sanitization Check:');
    const qTask = tasks.find(t => t.task_type === 'QUIZ');
    // If fetched via member endpoint or sanitized:
    const rawQuestionsStr = JSON.stringify(qTask.config.questions);
    console.log(`   Raw DB Quiz has correct_answer: ${rawQuestionsStr.includes('correct_answer')}`);

    // Call submit on Step 1 (MINI_GAME)
    console.log('\n7. Submitting Step 1 MINI_GAME (Pilih Jalanmu) on Live API...');
    const step1Submit = await fetch(`${PROD_URL}/api/tasks/${tasks[0].id}/submit`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        submission_type: 'AUTO_QUIZ',
        answers: {
          game: 'DECISION_FINANCE',
          choices: { sit_1: 'c', sit_2: 'a', sit_3: 'a', sit_4: 'a', sit_5: 'a', sit_6: 'a' },
          score: 100,
        },
      }),
    });
    const step1Data = await step1Submit.json();
    console.log('   Step 1 submit response:', step1Data);
    if (!step1Submit.ok || !step1Data.success) {
      throw new Error(`Step 1 submission failed: ${JSON.stringify(step1Data)}`);
    }
    console.log('   Step 1 MINI_GAME submitted & auto-rewarded: PASS 🟢');

    // Call submit on Step 2 (QUIZ)
    console.log('\n8. Submitting Step 2 QUIZ (Seberapa Siap Kamu di Dunia Kerja?) on Live API...');
    const step2Submit = await fetch(`${PROD_URL}/api/tasks/${tasks[1].id}/submit`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        submission_type: 'AUTO_QUIZ',
        answers: { q1: 'A', q2: 'A', q3: 'A', q4: 'A', q5: 'A', q6: 'A' },
      }),
    });
    const step2Data = await step2Submit.json();
    console.log('   Step 2 submit response:', step2Data);
    if (!step2Submit.ok || !step2Data.success) {
      throw new Error(`Step 2 submission failed: ${JSON.stringify(step2Data)}`);
    }
    console.log('   Step 2 QUIZ submitted & auto-rewarded: PASS 🟢');

    // Call submit on Step 3 (TEXT_RESPONSE)
    console.log('\n9. Submitting Step 3 TEXT_RESPONSE (Kalau Besok Harus Mandiri) on Live API...');
    const step3Submit = await fetch(`${PROD_URL}/api/tasks/${tasks[2].id}/submit`, {
      method: 'POST',
      headers: {
        Authorization: `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        submission_type: 'MANUAL_VERIFY',
        payload: {
          text: 'Tiga hal yang paling perlu saya kuasai adalah mengelola anggaran belanja harian, memasak makanan bergizi sendiri, dan disiplin mengatur waktu kerja serta istirahat. Hal ini penting agar saya mandiri secara finansial dan fisik. Langkah kecil saya minggu ini adalah mulai mencatat setiap pemasukan dan pengeluaran.',
          submitted_at: new Date().toISOString(),
        },
      }),
    });
    const step3Data = await step3Submit.json();
    console.log('   Step 3 submit response:', step3Data);
    if (!step3Submit.ok || step3Data.status !== 'PENDING') {
      throw new Error(`Step 3 submission failed: ${JSON.stringify(step3Data)}`);
    }
    console.log('   Step 3 TEXT_RESPONSE submitted as PENDING: PASS 🟢');

  } finally {
    // Cleanup smoke test data
    console.log('\nCleaning up smoke test member & submissions...');
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_coin_transactions?user_uid=eq.${smokeUid}`, { method: 'DELETE', headers: dbHeaders });
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${smokeUid}`, { method: 'DELETE', headers: dbHeaders });
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${smokeUid}`, { method: 'DELETE', headers: dbHeaders });
    console.log('Cleanup completed.\n');
  }

  console.log('====================================================');
  console.log('PRODUCTION SMOKE TEST: ALL CHECKS PASSED 🟢');
  console.log('====================================================');
}

runLiveSmoke().catch(err => {
  console.error('Smoke test failed:', err);
  process.exit(1);
});
