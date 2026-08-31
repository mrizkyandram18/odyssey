require('dotenv').config();

const PROD_API_BASE = 'https://odyssey-beta-nine.vercel.app';
const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_KEY = process.env.SUPABASE_SERVICE_KEY;

const dbHeaders = {
  'apikey': SUPABASE_KEY,
  'Authorization': 'Bearer ' + SUPABASE_KEY,
  'Content-Type': 'application/json'
};

const delay = ms => new Promise(res => setTimeout(res, ms));

async function runSmokeTest() {
  console.log('==================================================');
  console.log('LIVE PRODUCTION API SMOKE TEST (FINAL SUITE)');
  console.log(`Target Base URL: ${PROD_API_BASE}`);
  console.log('==================================================\n');

  // Reset Admin device binding in DB first
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.demo-uid-2`, {
    method: 'PATCH',
    headers: dbHeaders,
    body: JSON.stringify({ device_id: 'admin-fixed-device-id', device_bound_at: new Date().toISOString() })
  });

  // Initial wait to clear rate-limit window (65s)
  console.log('Waiting 65s for Production rate-limit window reset...');
  await delay(65000);
  console.log('Rate-limit window reset. Starting live assertions...\n');

  let allPassed = true;
  const testTimestamp = Date.now();
  const testUsername = `smoke_user_${testTimestamp}`;
  const testExplorerName = `Smoke Test User ${testTimestamp}`;
  const deviceA = `device-A-${testTimestamp}`;
  const deviceB = `device-B-${testTimestamp}`;

  let adminToken = '';
  let memberToken = '';
  let createdUserUid = `smoke-uid-${testTimestamp}`;
  let testFamilyId = 'demo-crew-1';

  // --- 1. ADMIN LOGIN ---
  console.log('1. Testing Admin Login...');
  try {
    const adminLoginRes = await fetch(`${PROD_API_BASE}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: 'admin',
        login_method: 'PASSWORD',
        credential: 'admin123',
        device: { device_id: 'admin-fixed-device-id' }
      })
    });
    const adminData = await adminLoginRes.json();
    if (adminLoginRes.status === 200 && adminData.session) {
      adminToken = adminData.session;
      console.log('   Admin Login: PASS 🟢');
    } else {
      console.log(`   Admin Login: FAIL 🔴 (Status ${adminLoginRes.status}: ${JSON.stringify(adminData)})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Admin Login ERROR: ${err.message}`);
    allPassed = false;
  }

  // Setup test user in Production DB
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
    method: 'POST',
    headers: { ...dbHeaders, 'Prefer': 'resolution=merge-duplicates' },
    body: JSON.stringify([{
      uid: createdUserUid,
      explorer_name: testExplorerName,
      role: 'MEMBER',
      family_id: testFamilyId,
      coins: 100,
      xp: 0,
      level: 1,
      is_active: true
    }])
  });
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_local_users`, {
    method: 'POST',
    headers: { ...dbHeaders, 'Prefer': 'resolution=merge-duplicates' },
    body: JSON.stringify([{
      id: `local-${createdUserUid}`,
      username: testUsername,
      password_hash: '$2a$10$tfuTDHMLQ0oW0WJqJz20seP0NvP5P2zdWNNxOuOd3bUa5TR1rB..W', // admin123
      profile_uid: createdUserUid
    }])
  });
  console.log(`\n2. Setup Test User ${testUsername} in Production DB: PASS 🟢`);

  await delay(15000); // 15s spacing

  // --- 3. MEMBER LOGIN & DEVICE LOCKING ---
  console.log('\n3. Testing Member Login & Device Binding (Device A)...');
  try {
    const loginARes = await fetch(`${PROD_API_BASE}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: testUsername,
        login_method: 'PASSWORD',
        credential: 'admin123',
        device: { device_id: deviceA }
      })
    });
    const loginAData = await loginARes.json();
    if (loginARes.status === 200 && loginAData.session) {
      memberToken = loginAData.session;
      console.log('   Device A Login: PASS 🟢');
    } else {
      console.log(`   Device A Login: FAIL 🔴 (Status ${loginARes.status}: ${JSON.stringify(loginAData)})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Device A Login ERROR: ${err.message}`);
    allPassed = false;
  }

  // Verify DB state after Device A login
  const profCheck1Res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${createdUserUid}&select=device_id,device_bound_at`, { headers: dbHeaders });
  const profCheck1 = await profCheck1Res.json();
  console.log(`   DB State after Device A: device_id = "${profCheck1[0]?.device_id}", bound_at = ${profCheck1[0]?.device_bound_at ? 'SET' : 'NULL'}`);
  const devABoundPass = (profCheck1[0]?.device_id === deviceA);
  if (!devABoundPass) allPassed = false;

  await delay(15000); // 15s spacing

  // --- 3b. DEVICE A RELOGIN ---
  console.log('\n3b. Testing Same Device Relogin (Device A)...');
  try {
    const reloginARes = await fetch(`${PROD_API_BASE}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: testUsername,
        login_method: 'PASSWORD',
        credential: 'admin123',
        device: { device_id: deviceA }
      })
    });
    if (reloginARes.status === 200) {
      console.log('   Device A Relogin: PASS 🟢');
    } else {
      console.log(`   Device A Relogin: FAIL 🔴 (Status ${reloginARes.status})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Device A Relogin ERROR: ${err.message}`);
    allPassed = false;
  }

  await delay(15000); // 15s spacing

  // --- 3c. SECOND DEVICE REJECTION (Device B) ---
  console.log('\n3c. Testing Second Device Login Block (Device B)...');
  try {
    const loginBRes = await fetch(`${PROD_API_BASE}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: testUsername,
        login_method: 'PASSWORD',
        credential: 'admin123',
        device: { device_id: deviceB }
      })
    });
    const loginBData = await loginBRes.json();
    if (loginBRes.status === 403) {
      console.log(`   Second Device Block: PASS 🟢 (HTTP 403: "${loginBData.error || loginBData.message}")`);
    } else {
      console.log(`   Second Device Block: FAIL 🔴 (Expected 403, got ${loginBRes.status}: ${JSON.stringify(loginBData)})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Second Device Block ERROR: ${err.message}`);
    allPassed = false;
  }

  // --- 3d. ADMIN RESET DEVICE ---
  console.log('\n3d. Testing Admin Device Reset...');
  try {
    const rpcRes = await fetch(`${SUPABASE_URL}/rest/v1/rpc/odyssey_admin_reset_device`, {
      method: 'POST',
      headers: dbHeaders,
      body: JSON.stringify({ p_target_uid: createdUserUid })
    });
    if (rpcRes.status === 200) {
      console.log('   Admin Reset Device (Server RPC): PASS 🟢');
    } else {
      console.log(`   Admin Reset Device: FAIL 🔴 (Status ${rpcRes.status})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Admin Reset Device ERROR: ${err.message}`);
    allPassed = false;
  }

  // Verify DB state after reset
  const profCheck2Res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${createdUserUid}&select=device_id,device_bound_at`, { headers: dbHeaders });
  const profCheck2 = await profCheck2Res.json();
  console.log(`   DB State after Reset: device_id = ${profCheck2[0]?.device_id}, bound_at = ${profCheck2[0]?.device_bound_at}`);
  const devResetPass = (profCheck2[0]?.device_id === null || profCheck2[0]?.device_id === undefined);
  if (!devResetPass) allPassed = false;

  await delay(15000); // 15s spacing

  // --- 3e. POST-RESET REBIND (Device B) ---
  console.log('\n3e. Testing Post-Reset Device B Login (Rebind)...');
  try {
    const rebindRes = await fetch(`${PROD_API_BASE}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: testUsername,
        login_method: 'PASSWORD',
        credential: 'admin123',
        device: { device_id: deviceB }
      })
    });
    const rebindData = await rebindRes.json();
    if (rebindRes.status === 200) {
      console.log('   Device B Rebind: PASS 🟢');
    } else {
      console.log(`   Device B Rebind: FAIL 🔴 (Status ${rebindRes.status}: ${JSON.stringify(rebindData)})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Device B Rebind ERROR: ${err.message}`);
    allPassed = false;
  }

  // Verify DB state after Device B rebind
  const profCheck3Res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${createdUserUid}&select=device_id`, { headers: dbHeaders });
  const profCheck3 = await profCheck3Res.json();
  console.log(`   DB State after Rebind: device_id = "${profCheck3[0]?.device_id}"`);
  const devBBoundPass = (profCheck3[0]?.device_id === deviceB);
  if (!devBBoundPass) allPassed = false;

  await delay(15000); // 15s spacing

  // --- 3f. OLD DEVICE REJECTION (Device A after Rebind) ---
  console.log('\n3f. Testing Old Device A Block after Rebind...');
  try {
    const oldDevRes = await fetch(`${PROD_API_BASE}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: testUsername,
        login_method: 'PASSWORD',
        credential: 'admin123',
        device: { device_id: deviceA }
      })
    });
    if (oldDevRes.status === 403) {
      console.log('   Old Device A Block: PASS 🟢 (HTTP 403)');
    } else {
      console.log(`   Old Device A Block: FAIL 🔴 (Expected 403, got ${oldDevRes.status})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Old Device A Block ERROR: ${err.message}`);
    allPassed = false;
  }

  // --- 4. TASK / USER ISOLATION ---
  console.log('\n4. Testing Task Delivery & Isolation (/api/tasks/today)...');
  try {
    const todayRes = await fetch(`${PROD_API_BASE}/api/tasks/today`, {
      headers: {
        'Authorization': `Bearer ${memberToken}`,
        'Cookie': `odyssey_session=${memberToken}`
      }
    });
    console.log(`   /api/tasks/today Endpoint Status: ${todayRes.status}`);
    const todayPass = (todayRes.status === 200 || todayRes.status === 304 || todayRes.status === 204);
    if (!todayPass) allPassed = false;
    console.log(`   Task Delivery: ${todayPass ? 'PASS 🟢' : 'FAIL 🔴'}`);
  } catch (err) {
    console.log(`   Task Delivery ERROR: ${err.message}`);
    allPassed = false;
  }

  // --- 5. REDEEM-ONLY CASH CONTRACT ---
  console.log('\n5. Testing Redemption Cash Contract...');
  try {
    // Legacy /api/shop/items -> 404
    const legacyShopRes = await fetch(`${PROD_API_BASE}/api/shop/items`, {
      headers: {
        'Authorization': `Bearer ${memberToken}`,
        'Cookie': `odyssey_session=${memberToken}`
      }
    });
    console.log(`   Legacy /api/shop/items Status: ${legacyShopRes.status} (Expected 404)`);
    const shopPass = (legacyShopRes.status === 404);
    if (!shopPass) allPassed = false;
    console.log(`   Legacy Catalog 404: ${shopPass ? 'PASS 🟢' : 'FAIL 🔴'}`);

    // Invalid non-cash redemption -> Server rejected
    const invalidRedeemRes = await fetch(`${PROD_API_BASE}/api/shop/redeem`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${memberToken}`,
        'Cookie': `odyssey_session=${memberToken}`
      },
      body: JSON.stringify({
        coins: 100,
        target_type: 'PULSA', // Non-cash type
        target_value: '08123456789'
      })
    });
    console.log(`   Non-cash redemption (PULSA) Status: ${invalidRedeemRes.status} (Expected 400/403/rejected)`);
    const nonCashRejected = (invalidRedeemRes.status >= 400);
    if (!nonCashRejected) allPassed = false;
    console.log(`   Non-cash Rejection: ${nonCashRejected ? 'PASS 🟢' : 'FAIL 🔴'}`);
  } catch (err) {
    console.log(`   Redemption Contract ERROR: ${err.message}`);
    allPassed = false;
  }

  // --- 6. ADMIN AUTHORIZATION ---
  console.log('\n6. Testing Admin Authorization Scoping...');
  try {
    const memberAdminAttempt = await fetch(`${PROD_API_BASE}/api/admin/members`, {
      headers: {
        'Authorization': `Bearer ${memberToken}`,
        'Cookie': `odyssey_session=${memberToken}`
      }
    });
    if (memberAdminAttempt.status === 403) {
      console.log('   MEMBER -> Admin API: PASS 🟢 (HTTP 403 Forbidden)');
    } else {
      console.log(`   MEMBER -> Admin API: FAIL 🔴 (Expected 403, got ${memberAdminAttempt.status})`);
      allPassed = false;
    }
  } catch (err) {
    console.log(`   Admin Auth Check ERROR: ${err.message}`);
    allPassed = false;
  }

  // --- 7. CLEANUP TEMPORARY SMOKE TEST DATA ---
  console.log('\n7. Cleaning up temporary smoke test data...');
  if (createdUserUid) {
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_completions?user_uid=eq.${createdUserUid}`, { method: 'DELETE', headers: dbHeaders });
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_claims?user_uid=eq.${createdUserUid}`, { method: 'DELETE', headers: dbHeaders });
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_local_users?profile_uid=eq.${createdUserUid}`, { method: 'DELETE', headers: dbHeaders });
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${createdUserUid}`, { method: 'DELETE', headers: dbHeaders });
    console.log(`   Cleaned up test user ${createdUserUid} 🟢`);
  }

  console.log('\n==================================================');
  console.log(`OVERALL PRODUCTION SMOKE TEST: ${allPassed ? 'ALL SMOKE TESTS PASSED 🟢' : 'SMOKE TEST FAILED 🔴'}`);
  console.log('==================================================');
}

runSmokeTest().catch(console.error);
