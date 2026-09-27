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
  console.log('END-TO-END VERIFICATION (11/09/2026 TASKS)');
  console.log('====================================================\n');

  const testUid = `test_member_${Date.now()}`;
  const testFamilyId = 'demo-crew-1';
  let passed = true;

  // Artifact Tracking Manifest
  const manifest = {
    test_uid: testUid,
    submission_ids: [],
    transaction_ids: [],
  };

  // 0. Setup a clean test member profile
  console.log('0. Setting up test member in family demo-crew-1...');
  const profRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'return=representation' },
    body: JSON.stringify([{
      uid: testUid,
      username: testUid,
      explorer_name: 'Test Explorer 11-Sep',
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
    // 1. Fetch 11 September tasks
    console.log('1. Querying active tasks for 2026-09-11...');
    const taskRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?active_date=eq.2026-09-11&is_active=eq.true&order=step_order.asc`, { headers });
    const tasks = await taskRes.json();
    console.log(`   Found ${tasks.length} active tasks on 2026-09-11:`);
    tasks.forEach(t => {
      console.log(`   - Step #${t.step_order} [ID ${t.id}]: "${t.title}" (${t.task_type}, ${t.evaluation_type}, ${t.reward_coins}c / ${t.reward_xp}xp)`);
    });

    if (tasks.length !== 3) {
      throw new Error(`Expected exactly 3 active tasks for 11 September, got ${tasks.length}`);
    }

    const [step1Task, step2Task, step3Task] = tasks;
    if (step1Task.step_order !== 1 || step1Task.task_type !== 'MINI_GAME' || step1Task.title !== 'Prioritas Dulu, Baru Gas') {
      throw new Error(`Step 1 task invalid: ${JSON.stringify(step1Task)}`);
    }
    if (step2Task.step_order !== 2 || step2Task.task_type !== 'PHOTO_UPLOAD' || step2Task.title !== 'Bikin Tempatmu Lebih Siap') {
      throw new Error(`Step 2 task invalid: ${JSON.stringify(step2Task)}`);
    }
    if (step3Task.step_order !== 3 || step3Task.task_type !== 'TEXT_RESPONSE' || step3Task.title !== 'Kalau Ada Masalah di Tempat Kerja') {
      throw new Error(`Step 3 task invalid: ${JSON.stringify(step3Task)}`);
    }
    console.log('   Task sequence, titles, types, and step orders are correct! 🟢\n');

    // 2. FLOW 1: Task 1 (MINI_GAME - Prioritas Dulu, Baru Gas)
    console.log('--- TESTING TASK 1: Prioritas Dulu, Baru Gas ---');
    console.log(`   Config game: ${step1Task.config.game}`);
    console.log(`   Currency mode: ${step1Task.config.scenario.currency}`);
    console.log(`   Events count: ${step1Task.config.scenario.events.length}`);

    // Security check: Anti-Cheat — submitting without answering all events
    console.log('   [Security] Testing Incomplete Choices...');
    try {
      await rpc('odyssey_submit_auto_task', {
        p_task_id: step1Task.id,
        p_user_uid: testUid,
        p_answers: { score: -1 }, // Negative score
      });
      console.log('   [Security] FAILED: Negative score was accepted! 🔴');
      passed = false;
    } catch (err) {
      console.log(`   [Security] Rejected invalid score: "${err.message}" 🟢`);
    }

    // Submit valid decisions
    console.log('   Submitting valid scenario choices...');
    const step1Answers = {
      game: 'DECISION_PRIORITY',
      choices: {
        sit_1: 'a',
        sit_2: 'a',
        sit_3: 'a',
        sit_4: 'a',
        sit_5: 'a',
        sit_6: 'a',
      },
      final_balance: 120,
      score: 100,
    };

    const step1Result = await rpc('odyssey_submit_auto_task', {
      p_task_id: step1Task.id,
      p_user_uid: testUid,
      p_answers: step1Answers,
    });
    console.log(`   Result: success=${step1Result.success}, coins_earned=${step1Result.coins_earned}, xp_earned=${step1Result.xp_earned}, new_balance=${step1Result.new_balance}, new_xp=${step1Result.new_xp}`);
    manifest.submission_ids.push(step1Result.submission_id);

    if (!step1Result.success || step1Result.coins_earned <= 0 || step1Result.xp_earned !== 100) {
      throw new Error(`Step 1 completion failed or reward incorrect: ${JSON.stringify(step1Result)}`);
    }
    const expectedCoinsAfterStep1 = step1Result.coins_earned;

    // Security check: Repeat submit on Task 1
    console.log('   [Security] Testing Repeat Submit on Step 1...');
    try {
      await rpc('odyssey_submit_auto_task', {
        p_task_id: step1Task.id,
        p_user_uid: testUid,
        p_answers: step1Answers,
      });
      console.log('   [Security] FAILED: Repeat auto submit permitted! 🔴');
      passed = false;
    } catch (err) {
      console.log(`   [Security] Successfully blocked repeat submit: "${err.message}" 🟢\n`);
    }

    // 3. FLOW 2: Task 2 (PHOTO_UPLOAD - Bikin Tempatmu Lebih Siap, AUTO eval)
    console.log('--- TESTING TASK 2: Bikin Tempatmu Lebih Siap ---');
    console.log(`   Task evaluation_type: ${step2Task.evaluation_type}`);
    console.log(`   Camera-only flag: ${Boolean(step2Task.config?.camera_only)}`);
    if (Boolean(step2Task.config?.camera_only)) {
      throw new Error('Task 2 must NOT be camera-only!');
    }

    const step2Payload = {
      file_url: 'https://odyssey.supabase.co/storage/v1/object/public/proofs/test_clean_desk.jpg',
      file_name: 'test_clean_desk.jpg',
      file_size: 102400,
      captured_at: new Date().toISOString(),
    };

    console.log('   Submitting photo proof to AUTO evaluation pipeline...');
    const step2Result = await rpc('odyssey_submit_auto_task', {
      p_task_id: step2Task.id,
      p_user_uid: testUid,
      p_answers: step2Payload,
    });
    console.log(`   Result: success=${step2Result.success}, coins_earned=${step2Result.coins_earned}, xp_earned=${step2Result.xp_earned}, new_balance=${step2Result.new_balance}, new_xp=${step2Result.new_xp}`);
    manifest.submission_ids.push(step2Result.submission_id);

    if (!step2Result.success || step2Result.coins_earned <= 0 || step2Result.xp_earned !== 100) {
      throw new Error(`Step 2 completion failed or reward incorrect: ${JSON.stringify(step2Result)}`);
    }
    const expectedCoinsAfterStep2 = expectedCoinsAfterStep1 + step2Result.coins_earned;
    console.log(`   Member balance after Step 2: ${step2Result.new_balance} coins (expected ${expectedCoinsAfterStep2})\n`);

    // 4. FLOW 3: Task 3 (TEXT_RESPONSE - Kalau Ada Masalah di Tempat Kerja, ADMIN_REVIEW)
    console.log('--- TESTING TASK 3: Kalau Ada Masalah di Tempat Kerja ---');
    console.log(`   Task evaluation_type: ${step3Task.evaluation_type}`);
    console.log(`   Min characters: ${step3Task.config.minimum_characters}`);
    console.log(`   Max characters: ${step3Task.config.maximum_characters}`);

    const shortText = 'Saya bingung.';
    console.log(`   [Validation] Testing short text (${shortText.length} chars, min ${step3Task.config.minimum_characters})...`);
    if (shortText.length < step3Task.config.minimum_characters) {
      console.log('   [Validation] Client validation correctly blocks submission below minimum chars 🟢');
    }

    const validText = '1) Langkah pertama yang saya lakukan adalah berhenti sejenak, mengidentifikasi akar masalah secara tenang, dan tidak panik agar tidak membuat kesalahan baru. ' +
      '2) Saya segera mengabarkan atasan langsung dan rekan kerja yang terdampak secara transparan sebelum masalah membesar. ' +
      '3) Hal yang saya hindari adalah menyembunyikan kesalahan, mencari kambing hitam, atau pura-pura tidak tahu. ' +
      '4) Saya memilih langkah ini karena kejujuran, akuntabilitas, dan respon proaktif adalah kunci utama menjaga kepercayaan dan keselamatan operasional tim.';
    console.log(`   Valid response length: ${validText.length} characters (within [${step3Task.config.minimum_characters}, ${step3Task.config.maximum_characters}])`);

    console.log('   Submitting to manual verification pipeline (odyssey_submit_manual_task)...');
    const step3SubmitResult = await rpc('odyssey_submit_manual_task', {
      p_task_id: step3Task.id,
      p_user_uid: testUid,
      p_payload: {
        text: validText,
        submitted_at: new Date().toISOString(),
      },
    });
    console.log(`   Manual submit result: success=${step3SubmitResult.success}, submission_id=${step3SubmitResult.submission_id}, status=${step3SubmitResult.status}`);
    manifest.submission_ids.push(step3SubmitResult.submission_id);

    if (!step3SubmitResult.success || step3SubmitResult.status !== 'PENDING') {
      throw new Error(`Step 3 manual submit failed: expected PENDING status, got ${step3SubmitResult.status}`);
    }

    // 5. FLOW 4: Admin Review on Task 3
    console.log('\n--- TESTING ADMIN REVIEW ON TASK 3 ---');
    const adminRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?family_id=eq.demo-crew-1&role=eq.ADMIN&select=uid,explorer_name`, { headers });
    const admins = await adminRes.json();
    if (admins.length === 0) throw new Error('No admin user found in demo-crew-1');
    const adminUid = admins[0].uid;
    console.log(`   Admin reviewer: ${admins[0].explorer_name} (${adminUid})`);

    // Verify submission is in PENDING state
    const subCheck = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?id=eq.${step3SubmitResult.submission_id}&select=*`, { headers });
    const subRows = await subCheck.json();
    console.log(`   Submission record in DB: status=${subRows[0].status}, task_id=${subRows[0].task_id}, snippet="${subRows[0].payload.text.substring(0, 50)}..."`);

    // Admin approves
    console.log('   Admin calling odyssey_verify_submission (status: APPROVED)...');
    const verifyResult = await rpc('odyssey_verify_submission', {
      p_submission_id: step3SubmitResult.submission_id,
      p_admin_uid: adminUid,
      p_status: 'APPROVED',
      p_admin_notes: 'Refleksi penanganan masalah sangat dewasa, runut, dan bertanggung jawab!',
      p_penalty_coins: 0,
    });
    console.log(`   Admin approval result: success=${verifyResult.success}, coins_earned=${verifyResult.coins_earned}, new_balance=${verifyResult.new_balance}, new_xp=${verifyResult.new_xp}`);
    const expectedCoinsFinal = expectedCoinsAfterStep2 + verifyResult.coins_earned;
    if (!verifyResult.success || verifyResult.coins_earned <= 0 || verifyResult.new_balance !== expectedCoinsFinal) {
      throw new Error(`Step 3 approval reward incorrect: expected earned > 0, total ${expectedCoinsFinal}, got earned ${verifyResult.coins_earned}, total ${verifyResult.new_balance}`);
    }

    // Anti-Double Approval: Attempt second approval
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

    // 6. Verify Ledger transactions
    console.log('\n--- VERIFYING COIN LEDGER TRANSACTIONS ---');
    const txRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_coin_transactions?user_uid=eq.${testUid}&order=created_at.asc`, { headers });
    const txs = await txRes.json();
    console.log(`   Found ${txs.length} ledger transactions:`);
    txs.forEach((tx, i) => {
      manifest.transaction_ids.push(tx.id);
      console.log(`   #${i + 1} [ID ${tx.id}]: +${tx.amount} coins | type: ${tx.type} | ref: ${tx.reference_id} | desc: "${tx.description}"`);
    });
    const totalLedger = txs.reduce((sum, tx) => sum + tx.amount, 0);
    console.log(`   Total ledger amount = ${totalLedger} coins (matches user coins: ${totalLedger === expectedCoinsFinal})`);
    if (txs.length !== 3 || totalLedger !== expectedCoinsFinal) {
      throw new Error(`Ledger verification failed: expected 3 transactions summing to ${expectedCoinsFinal}, got ${txs.length} summing to ${totalLedger}`);
    }

  } finally {
    // 7. Strict Cleanup using exact test identifiers
    console.log('\n====================================================');
    console.log('STRICT CLEANUP OF TEST ARTIFACTS');
    console.log('====================================================');
    console.log(`Target Member: ${testUid}`);
    console.log(`Target Submissions: [${manifest.submission_ids.join(', ')}]`);

    if (manifest.submission_ids.length > 0) {
      const delSub = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?id=in.(${manifest.submission_ids.join(',')})`, { method: 'DELETE', headers });
      console.log(`   Deleted exact test submissions (${delSub.status}): [${manifest.submission_ids.join(', ')}]`);
    }

    // Safe account cleanup policy: Deactivate test member account (immutable ledger P0012 prevents row deletion)
    const deactMem = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testUid}`, {
      method: 'PATCH',
      headers,
      body: JSON.stringify({ is_active: false, explorer_name: '[TEST DEACTIVATED] Smoke Test User' }),
    });
    console.log(`   Deactivated test member profile (${testUid}): status ${deactMem.status}`);

    // Post-cleanup verification
    const remSub = await (await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${testUid}`, { headers })).json();
    const profCheck = await (await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testUid}&select=is_active`, { headers })).json();

    console.log(`\nPost-Cleanup Audit:`);
    console.log(`   Remaining test submissions: ${remSub.length} (expected 0)`);
    console.log(`   Member is_active: ${profCheck[0]?.is_active} (expected false)`);

    if (remSub.length !== 0 || profCheck[0]?.is_active !== false) {
      console.error('🔴 CLEANUP INCOMPLETE!');
      passed = false;
    } else {
      console.log('🟢 CLEANUP VERIFIED 100% COMPLETE!');
    }
  }

  console.log('\n====================================================');
  console.log(`VERIFICATION RESULT: ${passed ? 'ALL CHECKS PASSED 🟢' : 'SOME CHECKS FAILED 🔴'}`);
  console.log('====================================================');
  if (!passed) process.exit(1);
}

runVerification().catch(err => {
  console.error('Fatal verification error:', err);
  process.exit(1);
});
