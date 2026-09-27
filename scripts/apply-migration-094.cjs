require('dotenv').config();

// ============================================================
// Script: apply-migration-094.cjs
// Applies historical compensation for Selvi (Sep 25-26 zero-coin hotfix)
// Uses RPC odyssey_apply_historical_compensation
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

const GRANTS = [
  { submission_id: 193, task_id: 560, active_date: '2026-09-25', amount: 83 },
  { submission_id: 194, task_id: 561, active_date: '2026-09-25', amount: 74 },
  { submission_id: 195, task_id: 562, active_date: '2026-09-26', amount: 74 },
  { submission_id: 196, task_id: 563, active_date: '2026-09-26', amount: 28 },
];

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
  console.log('==================================================');
  console.log('HOTFIX: Apply Migration 094 Historical Compensation');
  console.log('==================================================');

  // GATE 11: Display exact target submission IDs, target UID, amounts, total
  console.log('\n[PRE-FLIGHT TARGETS]');
  console.log(`Target User UID: ${TARGET_USER_UID} (selvicahyani)`);
  console.log(`Admin UID:       ${ADMIN_UID}`);
  console.log('Grants to apply:');
  let totalAmount = 0;
  for (const g of GRANTS) {
    console.log(`  - Sub ${g.submission_id} (Task ${g.task_id}, ${g.active_date}): ${g.amount} Coins`);
    totalAmount += g.amount;
  }
  console.log(`Total Compensation: ${totalAmount} Coins`);

  // Verify before balance
  const profBefore = await get(`odyssey_user_profiles?uid=eq.${TARGET_USER_UID}&select=coins`);
  const balanceBefore = profBefore[0]?.coins ?? 0;
  console.log(`\nBalance before: ${balanceBefore} Coins`);

  // Verify submissions are APPROVED and coins_earned == 0
  const subIds = GRANTS.map(g => g.submission_id).join(',');
  const subs = await get(`odyssey_task_submissions?id=in.(${subIds})&select=id,status,coins_earned,user_uid,task_id`);
  if (subs.length !== 4) {
    throw new Error(`Expected 4 submissions, found ${subs.length}`);
  }
  for (const s of subs) {
    if (s.user_uid !== TARGET_USER_UID) throw new Error(`Sub ${s.id} owner mismatch: ${s.user_uid}`);
    if (s.status !== 'APPROVED') throw new Error(`Sub ${s.id} status not APPROVED: ${s.status}`);
    if (s.coins_earned !== 0) throw new Error(`Sub ${s.id} coins_earned not 0: ${s.coins_earned}`);
  }
  console.log('Pre-check passed: all 4 submissions verified APPROVED with coins_earned = 0.');

  // Apply compensations via RPC
  console.log('\n[APPLYING COMPENSATIONS]');
  let totalCredited = 0;
  for (const g of GRANTS) {
    console.log(`Applying Sub ${g.submission_id} (${g.amount} Coins)...`);
    const res = await rpc('odyssey_apply_historical_compensation', {
      p_submission_id: g.submission_id,
      p_user_uid: TARGET_USER_UID,
      p_amount: g.amount,
      p_admin_uid: ADMIN_UID,
      p_reference: 'HOTFIX-092-WINDOW',
    });
    console.log(`  Result: credited=${res.credited}, already_applied=${res.already_applied}, new_balance=${res.new_balance}`);
    totalCredited += res.credited;
  }

  // Verify after balance
  const profAfter = await get(`odyssey_user_profiles?uid=eq.${TARGET_USER_UID}&select=coins`);
  const balanceAfter = profAfter[0]?.coins ?? 0;
  console.log(`\nBalance after: ${balanceAfter} Coins`);
  console.log(`Total credited: ${totalCredited} Coins`);

  if (balanceAfter !== balanceBefore + totalCredited) {
    throw new Error(`Balance mismatch: ${balanceBefore} + ${totalCredited} != ${balanceAfter}`);
  }

  console.log('\n[IDEMPOTENCY TEST: RE-RUN]');
  let rerunCredited = 0;
  for (const g of GRANTS) {
    const res = await rpc('odyssey_apply_historical_compensation', {
      p_submission_id: g.submission_id,
      p_user_uid: TARGET_USER_UID,
      p_amount: g.amount,
      p_admin_uid: ADMIN_UID,
      p_reference: 'HOTFIX-092-WINDOW',
    });
    rerunCredited += res.credited;
  }
  console.log(`Re-run credited: ${rerunCredited} Coins (expected 0)`);
  if (rerunCredited !== 0) {
    throw new Error(`Idempotency failed: second run credited ${rerunCredited}`);
  }

  console.log('\n=== MIGRATION 094 APPLIED SUCCESSFULLY ===');
}

run().catch(err => {
  console.error('\nFATAL ERROR:', err);
  process.exit(1);
});
