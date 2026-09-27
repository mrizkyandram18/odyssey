require('dotenv').config();

// ============================================================
// Comprehensive Verification & Gate Tests for Migration 094
// Tests all 14 points of GATE 9 and all requirements of GATE 12
// ============================================================

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;
const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
};

const TARGET_USER_UID = 'usr_1788196798_d46e4385da497091';
const ADMIN_UID = 'demo-uid-2';

let failures = 0;
const ok = (m) => console.log(`   PASS ${m}`);
const bad = (m) => { console.log(`   FAIL ${m}`); failures++; };

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

async function rpc(fn, body) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/rpc/${fn}`, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });
  const text = await r.text();
  if (!r.ok) {
    const err = new Error(`RPC ${fn}: ${r.status} ${text}`);
    err.status = r.status;
    err.body = text;
    throw err;
  }
  return text ? JSON.parse(text) : null;
}

async function run() {
  console.log('=== GATE 9 & GATE 12: Comprehensive 094 Verification ===\n');

  // [0] Schema Version Check
  console.log('[Test 0] Schema Version Check');
  const ver = await get('odyssey_schema_version?key=eq.schema_version&select=value');
  if (ver[0]?.value === '094_historical_compensation_sep25_26') {
    ok('schema_version = 094_historical_compensation_sep25_26');
  } else {
    bad(`Unexpected schema_version: ${JSON.stringify(ver)}`);
  }

  // [1] Negative: Wrong User Rejected (Point 6)
  console.log('\n[Test 1] Negative: Wrong User Rejected (Gate 6)');
  try {
    await rpc('odyssey_apply_historical_compensation', {
      p_submission_id: 193,
      p_user_uid: 'usr_wrong_user_uid_123',
      p_amount: 83,
      p_admin_uid: ADMIN_UID,
    });
    bad('Expected error for wrong user UID, but call succeeded');
  } catch (err) {
    if (err.body?.includes('P0003') || err.body?.includes('hanya untuk user yang tercatat')) {
      ok('Wrong user rejected with P0003 permission error');
    } else {
      bad(`Wrong user error unexpected: ${err.message}`);
    }
  }

  // [2] Negative: Unknown Submission Rejected (Point 10)
  console.log('\n[Test 2] Negative: Unknown Submission Rejected (Point 10)');
  try {
    await rpc('odyssey_apply_historical_compensation', {
      p_submission_id: 999999,
      p_user_uid: TARGET_USER_UID,
      p_amount: 50,
      p_admin_uid: ADMIN_UID,
    });
    bad('Expected error for unknown submission, but call succeeded');
  } catch (err) {
    if (err.body?.includes('P0002') || err.body?.includes('tidak termasuk dalam kompensasi')) {
      ok('Unknown submission rejected with P0002 not found error');
    } else {
      bad(`Unknown submission error unexpected: ${err.message}`);
    }
  }

  // [3] Negative: Invalid Amount Rejected (Point 9)
  console.log('\n[Test 3] Negative: Invalid Amount Rejected (Point 9)');
  try {
    await rpc('odyssey_apply_historical_compensation', {
      p_submission_id: 193,
      p_user_uid: TARGET_USER_UID,
      p_amount: 999, // Should be 83
      p_admin_uid: ADMIN_UID,
    });
    bad('Expected error for mismatching amount, but call succeeded');
  } catch (err) {
    if (err.body?.includes('P0005') || err.body?.includes('tidak sesuai dengan data hotfix')) {
      ok('Invalid amount rejected with P0005 data mismatch error');
    } else {
      bad(`Invalid amount error unexpected: ${err.message}`);
    }
  }

  // [4] Negative: Non-admin rejected
  console.log('\n[Test 4] Negative: Non-admin Rejected');
  try {
    await rpc('odyssey_apply_historical_compensation', {
      p_submission_id: 193,
      p_user_uid: TARGET_USER_UID,
      p_amount: 83,
      p_admin_uid: TARGET_USER_UID, // Selvi is MEMBER, not ADMIN
    });
    bad('Expected error for non-admin caller, but call succeeded');
  } catch (err) {
    if (err.body?.includes('P0003') || err.body?.includes('Hanya admin keluarga')) {
      ok('Non-admin caller rejected with P0003');
    } else {
      bad(`Non-admin error unexpected: ${err.message}`);
    }
  }

  // [5] Ledger and Grants table integrity
  console.log('\n[Test 5] Allowlist Grants Table');
  const grants = await get('odyssey_historical_compensation_grants?order=submission_id.asc');
  if (grants.length === 4) {
    ok(`Found exactly 4 allowlist grants: ${grants.map(g => `${g.submission_id}:${g.amount}`).join(', ')}`);
  } else {
    bad(`Expected 4 grants, found ${grants.length}`);
  }

  // [6] Ledger Check for 193, 194, 195, 196
  console.log('\n[Test 6] Ledger Transactions Check');
  const txs = await get('odyssey_coin_transactions?type=eq.HISTORICAL_COMPENSATION&order=reference_id.asc');
  if (txs.length === 4) {
    ok(`Found exactly 4 HISTORICAL_COMPENSATION ledger rows`);
    const expected = { '193': 83, '194': 74, '195': 74, '196': 28 };
    let amountsMatch = true;
    for (const t of txs) {
      if (t.amount !== expected[t.reference_id]) {
        bad(`Ledger ref ${t.reference_id} amount ${t.amount} != expected ${expected[t.reference_id]}`);
        amountsMatch = false;
      }
      if (!t.description.includes('Kompensasi historis')) {
        bad(`Ledger ref ${t.reference_id} missing auditable description prefix: ${t.description}`);
        amountsMatch = false;
      }
    }
    if (amountsMatch) ok('All 4 ledger rows have exact reference IDs, amounts, and auditable descriptions');
  } else {
    bad(`Expected 4 HISTORICAL_COMPENSATION rows, found ${txs.length}`);
  }

  // [7] Idempotency: Duplicate Invocation Returns already_applied=true, credited=0 (Point 4, 8)
  console.log('\n[Test 7] Idempotency: Re-invocation Check (Point 4, 8)');
  const re193 = await rpc('odyssey_apply_historical_compensation', {
    p_submission_id: 193,
    p_user_uid: TARGET_USER_UID,
    p_amount: 83,
    p_admin_uid: ADMIN_UID,
  });
  if (re193.already_applied === true && re193.credited === 0) {
    ok(`Re-invocation for 193 returned already_applied=true, credited=0, balance=${re193.new_balance}`);
  } else {
    bad(`Re-invocation failed idempotency: ${JSON.stringify(re193)}`);
  }

  // [8] Concurrent duplicate invocation test (Point 5)
  console.log('\n[Test 8] Concurrent Duplicate Invocation Test (Point 5)');
  const parallelCalls = await Promise.all([
    rpc('odyssey_apply_historical_compensation', { p_submission_id: 194, p_user_uid: TARGET_USER_UID, p_amount: 74, p_admin_uid: ADMIN_UID }),
    rpc('odyssey_apply_historical_compensation', { p_submission_id: 194, p_user_uid: TARGET_USER_UID, p_amount: 74, p_admin_uid: ADMIN_UID }),
    rpc('odyssey_apply_historical_compensation', { p_submission_id: 194, p_user_uid: TARGET_USER_UID, p_amount: 74, p_admin_uid: ADMIN_UID }),
  ]);
  const totalParallelCredited = parallelCalls.reduce((sum, res) => sum + res.credited, 0);
  if (totalParallelCredited === 0) {
    ok('3 concurrent duplicate calls for already-compensated 194 credited 0 additional coins');
  } else {
    bad(`Concurrent duplicate credited non-zero coins: ${totalParallelCredited}`);
  }

  // [9] Profile Balance Integrity (Point 3, GATE 12.C)
  console.log('\n[Test 9] Profile Balance Integrity (Gate 12.C)');
  const prof = await get(`odyssey_user_profiles?uid=eq.${TARGET_USER_UID}&select=coins`);
  const currentCoins = prof[0]?.coins;
  // All transactions sum
  const allTxs = await get(`odyssey_coin_transactions?user_uid=eq.${TARGET_USER_UID}&select=amount`);
  const ledgerSum = allTxs.reduce((sum, t) => sum + t.amount, 0);
  if (currentCoins === ledgerSum) {
    ok(`Profile coins (${currentCoins}) exactly equals sum of all ledger transactions (${ledgerSum})`);
  } else {
    bad(`Profile coins (${currentCoins}) != sum of ledger transactions (${ledgerSum})`);
  }
  if (currentCoins === 600) {
    ok(`Profile coins matches expected after-balance: 341 (before) + 259 (hotfix) = 600`);
  } else {
    bad(`Expected profile balance 600, got ${currentCoins}`);
  }

  // [10] Submissions Intact (Gate 12.A, Point 7, 8)
  console.log('\n[Test 10] Submissions Intact (Gate 12.A)');
  const subs = await get('odyssey_task_submissions?id=in.(193,194,195,196)&select=id,status,coins_earned');
  const subsIntact = subs.length === 4 && subs.every(s => s.status === 'APPROVED' && s.coins_earned === 0);
  if (subsIntact) {
    ok('Submissions 193, 194, 195, 196 remain APPROVED with coins_earned = 0 (historical preservation)');
  } else {
    bad(`Submissions modified unexpectedly: ${JSON.stringify(subs)}`);
  }

  // [11] Cap Isolation: HISTORICAL_COMPENSATION does NOT count toward TASK_REWARD cap (Gate 8, Point 13)
  console.log('\n[Test 11] Cap Isolation Check (Gate 8)');
  const earned = await rpc('odyssey_earned_this_period', { p_user_uid: TARGET_USER_UID });
  if (earned === 3480) {
    ok(`odyssey_earned_this_period remains exactly 3480 (cap headroom 120 preserved, undistorted by compensation)`);
  } else {
    bad(`odyssey_earned_this_period distorted: ${earned}, expected 3480`);
  }

  // [12] Normal Reward Calculation Flow (Gate 12.E, Point 13)
  console.log('\n[Test 12] Normal Reward Calculation Flow (Gate 12.E)');
  for (const [taskId, expectedReward, label] of [
    [660, 37, 'task 660 date 2026-09-27 (base 40 scaled to 37 for target 3320)'],
    [561, 74, 'task 561 date 2026-09-25'],
    [563, 28, 'task 563 date 2026-09-26'],
  ]) {
    const r = await rpc('odyssey_calc_target_reward', {
      p_target: 3320,
      p_task_id: taskId,
      p_family_id: 'demo-crew-1',
      p_user_uid: TARGET_USER_UID,
    });
    if (r === expectedReward) {
      ok(`Normal reward for ${label} = ${r} (expected ${expectedReward})`);
    } else {
      bad(`Reward for ${label} = ${r}, expected ${expectedReward}`);
    }
  }

  // [13] Claim / Redeem Flow Unchanged (Point 14)
  console.log('\n[Test 13] Claim / Redeem Policy Intact (Point 14)');
  const sysConfig = await get('odyssey_system_config?key=eq.default_minimum_withdrawal_coins&select=value');
  if (sysConfig[0]?.value === '500') {
    ok('default_minimum_withdrawal_coins = 500 unchanged');
  } else {
    bad(`default_minimum_withdrawal_coins = ${JSON.stringify(sysConfig)}`);
  }

  console.log(`\n==================================================`);
  console.log(`VERIFICATION RESULT: ${failures === 0 ? 'ALL PASS (14/14 checks passed)' : `FAILED (${failures})`}`);
  console.log(`==================================================`);
  process.exit(failures === 0 ? 0 : 1);
}

run().catch(err => {
  console.error('\nVerification failed with exception:', err);
  process.exit(1);
});
