require('dotenv').config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
};

async function rpc(name, body) {
  const res = await fetch(`${SUPABASE_URL}/rest/v1/rpc/${name}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });
  const data = await res.json();
  if (!res.ok) {
    const msg = data.message || data.error || JSON.stringify(data);
    const err = new Error(msg);
    err.status = res.status;
    err.data = data;
    throw err;
  }
  return data;
}

async function runVerification() {
  console.log('====================================================');
  console.log('PHASE 6, 7, 8 — END-TO-END VERIFICATION (10/09/2026)');
  console.log('====================================================\n');

  const testUid = `test_member_${Date.now()}`;
  const testFamilyId = 'demo-crew-1';
  let passed = true;

  // 0. Setup a clean test member profile
  console.log('0. Setting up test member in family demo-crew-1...');
  const profRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'return=representation' },
    body: JSON.stringify([{
      uid: testUid,
      username: testUid,
      explorer_name: 'Test Explorer 10-Sep',
      role: 'MEMBER',
      family_id: testFamilyId,
      coins: 0,
      xp: 0,
      level: 1,
      is_active: true,
      monthly_coin_target: 3200,
      created_at: '2026-09-01T00:00:00Z',
    }]),
  });
  if (!profRes.ok) throw new Error(`Failed to create test profile: ${await profRes.text()}`);
  console.log(`   Created test member: ${testUid}\n`);

  try {
    // 1. Fetch 10-Sep Tasks from DB
    console.log('1. Verifying Tasks for 2026-09-10 in DB...');
    const taskRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?active_date=eq.2026-09-10&is_active=eq.true&order=step_order.asc`, { headers });
    const tasks = await taskRes.json();
    console.log(`   Found ${tasks.length} active tasks:`);
    tasks.forEach(t => {
      console.log(`   - Step ${t.step_order}: "${t.title}" (${t.task_type}, ${t.evaluation_type}, ${t.reward_coins}c / ${t.reward_xp}xp) [ID: ${t.id}]`);
    });

    if (tasks.length !== 3) {
      throw new Error(`Expected exactly 3 active tasks for 2026-09-10, got ${tasks.length}`);
    }
    const [task1, task2, task3] = tasks;

    // --- STEP 1: MINI_GAME ---
    console.log('\n--- VERIFYING STEP 1: MINI_GAME (Pilih Jalanmu) ---');
    console.log(`   Title: ${task1.title}`);
    console.log(`   Config game: ${task1.config.game}`);
    console.log(`   Events count: ${task1.config.scenario.events.length}`);
    console.log(`   Initial balance: ${task1.config.scenario.initial_balance}`);

    // Test choices:
    const step1Choices = {
      sit_1: 'b', // -20,000
      sit_2: 'a', // -50,000
      sit_3: 'c', // -25,000
      sit_4: 'a', // +90,000
      sit_5: 'a', // -45,000
      sit_6: 'a', // -50,000
    };

    console.log('   Submitting valid choices to odyssey_submit_auto_task...');
    const step1Result = await rpc('odyssey_submit_auto_task', {
      p_task_id: task1.id,
      p_user_uid: testUid,
      p_answers: {
        game: 'DECISION_FINANCE',
        choices: step1Choices,
        score: 100,
        final_balance: 400000,
      },
    });
    console.log(`   Result: success=${step1Result.success}, coins_earned=${step1Result.coins_earned}, base_reward=${step1Result.base_reward}, new_balance=${step1Result.new_balance}, new_xp=${step1Result.new_xp}`);
    if (!step1Result.success || step1Result.base_reward !== 50 || step1Result.xp_earned !== 100 || step1Result.coins_earned <= 0) {
      throw new Error(`Step 1 reward incorrect: expected base_reward=50 and coins_earned > 0, got base=${step1Result.base_reward}, earned=${step1Result.coins_earned}`);
    }
    console.log('   Step 1 MINI_GAME: PASS 🟢');

    // Anti-Double Reward: Attempt to submit Step 1 again
    console.log('\n   [Security] Testing Anti-Double Claim on Step 1...');
    try {
      await rpc('odyssey_submit_auto_task', {
        p_task_id: task1.id,
        p_user_uid: testUid,
        p_answers: { choices: step1Choices, score: 100 },
      });
      console.log('   [Security] FAILED: Double claim was permitted! 🔴');
      passed = false;
    } catch (err) {
      console.log(`   [Security] Successfully blocked duplicate claim: "${err.message}" 🟢`);
    }

    // --- STEP 2: QUIZ ---
    console.log('\n--- VERIFYING STEP 2: QUIZ (Seberapa Siap Kamu di Dunia Kerja?) ---');
    console.log(`   Title: ${task2.title}`);
    console.log(`   Questions count: ${task2.config.questions.length}`);

    // Security check on Quiz: Try submitting wrong answers first
    console.log('   [Security] Testing rejection of incorrect quiz answer (q1 = B)...');
    try {
      await rpc('odyssey_submit_auto_task', {
        p_task_id: task2.id,
        p_user_uid: testUid,
        p_answers: { q1: 'B', q2: 'A', q3: 'A', q4: 'A', q5: 'A', q6: 'A' },
      });
      console.log('   [Security] FAILED: Incorrect answer was accepted! 🔴');
      passed = false;
    } catch (err) {
      console.log(`   [Security] Successfully rejected incorrect answer: "${err.message}" 🟢`);
    }

    // Now submit correct answers for all 6 questions
    console.log('   Submitting all correct answers (q1..q6 = A)...');
    const step2Result = await rpc('odyssey_submit_auto_task', {
      p_task_id: task2.id,
      p_user_uid: testUid,
      p_answers: { q1: 'A', q2: 'A', q3: 'A', q4: 'A', q5: 'A', q6: 'A' },
    });
    console.log(`   Result: success=${step2Result.success}, coins_earned=${step2Result.coins_earned}, base_reward=${step2Result.base_reward}, new_balance=${step2Result.new_balance}, new_xp=${step2Result.new_xp}`);
    if (!step2Result.success || step2Result.base_reward !== 40 || step2Result.xp_earned !== 100 || step2Result.coins_earned <= 0) {
      throw new Error(`Step 2 reward incorrect: expected base_reward=40 and coins_earned > 0, got base=${step2Result.base_reward}, earned=${step2Result.coins_earned}`);
    }
    console.log('   Step 2 QUIZ: PASS 🟢');

    // Anti-Double Reward: Attempt to submit Step 2 again
    console.log('\n   [Security] Testing Anti-Double Claim on Step 2...');
    try {
      await rpc('odyssey_submit_auto_task', {
        p_task_id: task2.id,
        p_user_uid: testUid,
        p_answers: { q1: 'A', q2: 'A', q3: 'A', q4: 'A', q5: 'A', q6: 'A' },
      });
      console.log('   [Security] FAILED: Double claim was permitted on Quiz! 🔴');
      passed = false;
    } catch (err) {
      console.log(`   [Security] Successfully blocked duplicate quiz claim: "${err.message}" 🟢`);
    }

    // --- STEP 3: TEXT_RESPONSE ---
    console.log('\n--- VERIFYING STEP 3: TEXT_RESPONSE (Kalau Besok Harus Mandiri) ---');
    console.log(`   Title: ${task3.title}`);
    console.log(`   Min chars: ${task3.config.minimum_characters}, Max chars: ${task3.config.maximum_characters}`);

    // Submit text response via odyssey_submit_manual_task
    const validText = 'Tiga hal yang paling perlu saya kuasai adalah manajemen keuangan pribadi, memasak makanan sehat sederhana, dan kemampuan menyelesaikan masalah secara mandiri tanpa panik. Hal ini penting agar saya dapat bertahan dalam kondisi apa pun secara bertanggung jawab. Langkah kecil yang saya mulai minggu ini adalah mencatat seluruh pengeluaran harian dan memasak menu sendiri.';
    console.log(`   Text response length: ${validText.length} characters (valid: ${validText.length >= 80 && validText.length <= 1500})`);

    const step3SubmitResult = await rpc('odyssey_submit_manual_task', {
      p_task_id: task3.id,
      p_user_uid: testUid,
      p_payload: {
        text: validText,
        submitted_at: new Date().toISOString(),
      },
    });
    console.log(`   Manual submit result: submission_id=${step3SubmitResult.submission_id}, status=${step3SubmitResult.status}`);
    if (step3SubmitResult.status !== 'PENDING') {
      throw new Error(`Expected submission status PENDING, got ${step3SubmitResult.status}`);
    }

    // Verify member coins didn't increase yet
    const expectedCoinsAfterStep2 = step1Result.coins_earned + step2Result.coins_earned;
    const profCheck1 = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testUid}&select=coins,xp`, { headers });
    const profData1 = (await profCheck1.json())[0];
    console.log(`   Member balance before admin review: coins=${profData1.coins} (expected ${expectedCoinsAfterStep2}), xp=${profData1.xp} (expected 200)`);
    if (profData1.coins !== expectedCoinsAfterStep2) throw new Error(`Premature reward granted for pending task! coins=${profData1.coins}`);

    // --- ADMIN REVIEW ---
    console.log('\n--- VERIFYING ADMIN REVIEW (Step 3 Approval) ---');
    // Find admin user in demo-crew-1
    const adminRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?family_id=eq.demo-crew-1&role=eq.ADMIN&select=uid,explorer_name`, { headers });
    const admins = await adminRes.json();
    if (admins.length === 0) throw new Error('No admin user found in demo-crew-1');
    const adminUid = admins[0].uid;
    console.log(`   Admin reviewer: ${admins[0].explorer_name} (${adminUid})`);

    // Verify submission is visible in submissions queue
    const subCheck = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?id=eq.${step3SubmitResult.submission_id}&select=*`, { headers });
    const subRows = await subCheck.json();
    console.log(`   Submission record in DB: status=${subRows[0].status}, task_id=${subRows[0].task_id}, payload.text snippet="${subRows[0].payload.text.substring(0, 50)}..."`);

    // Admin approves submission
    console.log('   Admin calling odyssey_verify_submission (status: APPROVED)...');
    const verifyResult = await rpc('odyssey_verify_submission', {
      p_submission_id: step3SubmitResult.submission_id,
      p_admin_uid: adminUid,
      p_status: 'APPROVED',
      p_admin_notes: 'Refleksi sangat baik dan realistis!',
      p_penalty_coins: 0,
    });
    console.log(`   Admin approval result: success=${verifyResult.success}, coins_earned=${verifyResult.coins_earned}, new_balance=${verifyResult.new_balance}, new_xp=${verifyResult.new_xp}`);
    const expectedCoinsFinal = expectedCoinsAfterStep2 + verifyResult.coins_earned;
    if (!verifyResult.success || verifyResult.coins_earned <= 0 || verifyResult.new_balance !== expectedCoinsFinal) {
      throw new Error(`Step 3 approval reward incorrect: expected earned > 0, total ${expectedCoinsFinal}, got earned ${verifyResult.coins_earned}, total ${verifyResult.new_balance}`);
    }

    // Verify member final balance and XP
    const profCheck2 = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testUid}&select=coins,xp`, { headers });
    const profData2 = (await profCheck2.json())[0];
    console.log(`   Final Member Profile: coins=${profData2.coins} (expected ${expectedCoinsFinal}), xp=${profData2.xp} (expected 300)`);
    if (profData2.coins !== expectedCoinsFinal || profData2.xp !== 300) {
      throw new Error(`Final balance mismatch: expected ${expectedCoinsFinal} coins and 300 XP, got ${profData2.coins}c / ${profData2.xp}xp`);
    }

    // Anti-Double Approval: Attempt to approve again
    console.log('\n   [Security] Testing Repeat Approval on Step 3...');
    try {
      await rpc('odyssey_verify_submission', {
        p_submission_id: step3SubmitResult.submission_id,
        p_admin_uid: adminUid,
        p_status: 'APPROVED',
      });
      console.log('   [Security] FAILED: Repeat approval was permitted! 🔴');
      passed = false;
    } catch (err) {
      console.log(`   [Security] Successfully blocked repeat approval: "${err.message}" 🟢`);
    }

    // Verify Ledger transactions
    console.log('\n--- VERIFYING COIN LEDGER TRANSACTIONS ---');
    const txRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_coin_transactions?user_uid=eq.${testUid}&order=created_at.asc`, { headers });
    const txs = await txRes.json();
    console.log(`   Found ${txs.length} ledger transactions:`);
    txs.forEach((tx, i) => {
      console.log(`   #${i + 1}: +${tx.amount} coins | type: ${tx.type} | ref: ${tx.reference_id} | desc: "${tx.description}"`);
    });
    const totalLedger = txs.reduce((sum, tx) => sum + tx.amount, 0);
    console.log(`   Total ledger amount = ${totalLedger} coins (matches user coins: ${totalLedger === expectedCoinsFinal})`);
    if (txs.length !== 3 || totalLedger !== expectedCoinsFinal) {
      throw new Error(`Ledger verification failed: expected 3 transactions summing to ${expectedCoinsFinal}, got ${txs.length} summing to ${totalLedger}`);
    }

  } finally {
    // Cleanup test user
    console.log('\nCleaning up test user submissions and profile...');
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_coin_transactions?user_uid=eq.${testUid}`, { method: 'DELETE', headers });
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${testUid}`, { method: 'DELETE', headers });
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testUid}`, { method: 'DELETE', headers });
    console.log('Cleanup completed.\n');
  }

  console.log('====================================================');
  console.log(`VERIFICATION RESULT: ${passed ? 'ALL CHECKS PASSED 🟢' : 'SOME CHECKS FAILED 🔴'}`);
  console.log('====================================================');
}

runVerification().catch(err => {
  console.error('Fatal verification error:', err);
  process.exit(1);
});
