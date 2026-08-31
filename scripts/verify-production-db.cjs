require('dotenv').config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  'apikey': SUPABASE_KEY,
  'Authorization': 'Bearer ' + SUPABASE_KEY
};

async function verify() {
  console.log('==================================================');
  console.log('VERIFYING PRODUCTION SUPABASE DB');
  console.log('==================================================\n');

  let allPassed = true;

  // --- 1. ROLE MODEL PURGE (055) ---
  console.log('1. Checking Role Model Purge (Migration 055)...');
  const profRes = await fetch(SUPABASE_URL + '/rest/v1/odyssey_user_profiles?select=role', { headers });
  const profiles = await profRes.json();
  
  const seekerCount = profiles.filter(p => p.role === 'SEEKER').length;
  const guideCount = profiles.filter(p => p.role === 'GUIDE').length;
  const builderCount = profiles.filter(p => p.role === 'BUILDER').length;
  const adminCount = profiles.filter(p => p.role === 'ADMIN').length;
  const memberCount = profiles.filter(p => p.role === 'MEMBER').length;
  const otherRoles = profiles.filter(p => !['ADMIN', 'MEMBER'].includes(p.role)).length;

  console.log(`   SEEKER  = ${seekerCount} (expected 0)`);
  console.log(`   GUIDE   = ${guideCount} (expected 0)`);
  console.log(`   BUILDER = ${builderCount} (expected 0)`);
  console.log(`   ADMIN   = ${adminCount}`);
  console.log(`   MEMBER  = ${memberCount}`);
  console.log(`   Other   = ${otherRoles} (expected 0)`);

  const rolePurgePass = (seekerCount === 0 && guideCount === 0 && builderCount === 0 && otherRoles === 0);
  if (!rolePurgePass) allPassed = false;
  console.log(`   ROLE PURGE STATUS: ${rolePurgePass ? 'PASS' : 'FAIL'}\n`);

  // --- 2. SEPTEMBER 2026 SEED ---
  console.log('2. Checking September 2026 Seed Data in Production DB...');
  const taskRes = await fetch(SUPABASE_URL + '/rest/v1/odyssey_tasks?select=active_date,reward_coins,reward_xp,title,family_id&active_date=gte.2026-09-01&active_date=lte.2026-09-30&title=like.ANTIGRAVITY:*', { headers });
  const tasks = await taskRes.json();

  const activeDates = new Set(tasks.map(t => t.active_date));
  const distinctDateCount = activeDates.size;
  console.log(`   Distinct active dates (2026-09-01 .. 2026-09-30) = ${distinctDateCount} (expected 30)`);

  let missingDates = [];
  for (let d = 1; d <= 30; d++) {
    const dateStr = `2026-09-${d.toString().padStart(2, '0')}`;
    if (!activeDates.has(dateStr)) missingDates.push(dateStr);
  }
  console.log(`   Missing dates count = ${missingDates.length}`);

  // Sum Day 1-24 and Day 25-30
  let day1_24_coins = 0;
  let day25_30_coins = 0;

  tasks.forEach(t => {
    const day = parseInt(t.active_date.split('-')[2], 10);
    if (day >= 1 && day <= 24) {
      day1_24_coins += t.reward_coins;
    } else if (day >= 25 && day <= 30) {
      day25_30_coins += t.reward_coins;
    }
  });
  const totalCoins = day1_24_coins + day25_30_coins;

  console.log(`   Day 1–24 Coins  = ${day1_24_coins} (expected 2690)`);
  console.log(`   Day 25–30 Coins = ${day25_30_coins} (expected 510)`);
  console.log(`   Total Coins     = ${totalCoins} (expected 3200)`);

  // Check Day 30 task
  const day30Tasks = tasks.filter(t => t.active_date === '2026-09-30');
  const day30Coins = day30Tasks.reduce((acc, t) => acc + t.reward_coins, 0);
  const day30XP = day30Tasks.reduce((acc, t) => acc + t.reward_xp, 0);

  console.log(`   Day 30 Coins    = ${day30Coins} (expected 0)`);
  console.log(`   Day 30 XP       = ${day30XP} (expected 100)`);

  const seedPass = (distinctDateCount === 30 && missingDates.length === 0 && day1_24_coins === 2690 && day25_30_coins === 510 && totalCoins === 3200 && day30Coins === 0 && day30XP === 100);
  if (!seedPass) allPassed = false;
  console.log(`   SEPTEMBER SEED STATUS: ${seedPass ? 'PASS' : 'FAIL'}\n`);

  // --- 3. PRODUCTION ECONOMY CONFIG ---
  console.log('3. Checking Production Economy Configuration...');
  const cfgRes = await fetch(SUPABASE_URL + '/rest/v1/odyssey_system_config?select=*', { headers });
  const cfgRows = await cfgRes.json();
  const configMap = {};
  cfgRows.forEach(row => { configMap[row.key] = row.value; });

  console.log(`   coin_conversion_rate = ${configMap['coin_conversion_rate']} (expected 100)`);
  console.log(`   payout_target_rupiah = ${configMap['payout_target_rupiah']} (expected 320000)`);
  console.log(`   max_payout_coins     = ${configMap['max_payout_coins']} (expected 3200)`);
  console.log(`   payout_day           = ${configMap['payout_day']} (expected 24)`);
  console.log(`   redemption_start_day = ${configMap['redemption_start_day']} (expected 24)`);
  console.log(`   redemption_end_day   = ${configMap['redemption_end_day']} (expected 26)`);
  console.log(`   earning_period_days  = ${configMap['earning_period_days']} (expected 30)`);
  console.log(`   timezone             = ${configMap['timezone']} (expected Asia/Jakarta)`);

  const derivedCoins = parseInt(configMap['payout_target_rupiah'], 10) / parseInt(configMap['coin_conversion_rate'], 10);
  console.log(`   Derived invariant: 320000 / 100 = ${derivedCoins} (matches max_payout_coins: ${derivedCoins === parseInt(configMap['max_payout_coins'], 10)})`);

  const econPass = (
    configMap['coin_conversion_rate'] === '100' &&
    configMap['payout_target_rupiah'] === '320000' &&
    configMap['max_payout_coins'] === '3200' &&
    configMap['payout_day'] === '24' &&
    configMap['redemption_start_day'] === '24' &&
    configMap['redemption_end_day'] === '26' &&
    configMap['earning_period_days'] === '30' &&
    configMap['timezone'] === 'Asia/Jakarta' &&
    derivedCoins === 3200
  );
  if (!econPass) allPassed = false;
  console.log(`   ECONOMY CONFIG STATUS: ${econPass ? 'PASS' : 'FAIL'}\n`);

  console.log('==================================================');
  console.log(`OVERALL DATABASE VERIFICATION: ${allPassed ? 'ALL VERIFICATIONS PASSED 🟢' : 'VERIFICATION FAILED 🔴'}`);
  console.log('==================================================');
}

verify().catch(console.error);
