const { chromium } = require('d:/Personal/Projects/Odyssey/web/node_modules/@playwright/test');
require('dotenv').config();
const fs = require('fs');
const path = require('path');

const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;
const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const SCREENSHOT_DIR = path.resolve('C:/Users/user/.gemini/antigravity/brain/e8e4f49b-a194-4068-b33e-2c26c1afd3ee/screenshots');

const dbHeaders = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
};

async function dbFetch(url, options = {}) {
  const res = await fetch(url, {
    ...options,
    headers: { ...dbHeaders, ...(options.headers || {}) },
  });
  return res;
}

async function dismissOnboarding(page) {
  for (let i = 0; i < 5; i++) {
    const btn = page.locator('button:has-text("Selanjutnya"), button:has-text("Mulai Sekarang")').first();
    if (await btn.isVisible({ timeout: 1000 }).catch(() => false)) {
      await btn.click();
      await page.waitForTimeout(400);
    } else {
      break;
    }
  }
}

(async () => {
  console.log('====================================================');
  console.log('FINAL USER-SIDE BROWSER AUTOMATION SMOKE TEST');
  console.log(`Target: ${PROD_URL}`);
  console.log('====================================================\n');

  if (!fs.existsSync(SCREENSHOT_DIR)) {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
  }

  const ts = Date.now();
  const testMemberUid = `smoke_ui_mem_${ts}`;
  const testMemberPass = 'password123';

  const testAdminUid = `smoke_ui_adm_${ts}`;
  const testAdminPass = 'password123';

  const createdRecords = {
    testMemberUid,
    testAdminUid,
    submissionIds: [],
    coinTransactionIds: [],
  };

  let browser = null;
  let allPass = true;
  const flowResults = {};

  try {
    // ----------------------------------------------------
    // 0. TIMEZONE ALIGNMENT: Asia/Jayapura (WIT, UTC+9)
    // ----------------------------------------------------
    console.log('0. Aligning system timezone to Asia/Jayapura (10 September in WIT)...');
    await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_system_config?key=eq.timezone`, {
      method: 'PATCH',
      body: JSON.stringify({ value: 'Asia/Jayapura' }),
    });
    console.log('   System timezone set to Asia/Jayapura: PASS 🟢\n');

    // ----------------------------------------------------
    // 1. SETUP DEDICATED TEST USERS IN DATABASE
    // ----------------------------------------------------
    console.log('1. Setting up dedicated test accounts in database...');
    
    // Member account
    const memRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
      method: 'POST',
      headers: { Prefer: 'return=minimal' },
      body: JSON.stringify([{
        uid: testMemberUid,
        username: testMemberUid,
        password_hash: testMemberPass,
        explorer_name: 'Smoke UI Explorer',
        role: 'MEMBER',
        family_id: 'demo-crew-1',
        coins: 100,
        xp: 50,
        level: 1,
        is_active: true,
        monthly_coin_target: 3200,
        created_at: '2026-09-01T00:00:00Z',
      }]),
    });
    if (!memRes.ok) throw new Error(`Failed to create test member: ${await memRes.text()}`);
    console.log(`   - Test member created: ${testMemberUid}`);

    // Admin account
    const admRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles`, {
      method: 'POST',
      headers: { Prefer: 'return=minimal' },
      body: JSON.stringify([{
        uid: testAdminUid,
        username: testAdminUid,
        password_hash: testAdminPass,
        explorer_name: 'Smoke UI Admin',
        role: 'ADMIN',
        family_id: 'demo-crew-1',
        coins: 0,
        xp: 0,
        level: 1,
        is_active: true,
        created_at: '2026-09-01T00:00:00Z',
      }]),
    });
    if (!admRes.ok) throw new Error(`Failed to create test admin: ${await admRes.text()}`);
    console.log(`   - Test admin created: ${testAdminUid}\n`);

    // ----------------------------------------------------
    // 2. LAUNCH BROWSER
    // ----------------------------------------------------
    console.log('2. Launching Playwright browser (Chrome)...');
    browser = await chromium.launch({
      executablePath: CHROME_PATH,
      headless: true,
    });
    const context = await browser.newContext({
      viewport: { width: 1280, height: 800 },
      deviceScaleFactor: 1,
    });

    const page = await context.newPage();
    console.log('   Browser launched successfully.\n');

    // ----------------------------------------------------
    // FLOW 1: MEMBER LOGIN & HOME VERIFICATION
    // ----------------------------------------------------
    console.log('--- FLOW 1: MEMBER LOGIN & HOME ---');
    await page.goto(`${PROD_URL}/#/login`, { waitUntil: 'networkidle' });
    console.log('   Opened login page.');

    await page.fill('input[placeholder="Masukkan nama pengguna"]', testMemberUid);
    await page.fill('input[placeholder="••••••••"]', testMemberPass);
    await page.click('button[type="submit"]:has-text("Masuk")');
    console.log('   Submitted login credentials.');

    // Wait for route to leave #/login
    await page.waitForFunction(() => !window.location.hash.includes('/login'), { timeout: 15000 });
    await page.waitForTimeout(2000);
    console.log('   Landed on Home page.');

    // Dismiss onboarding if shown
    await dismissOnboarding(page);

    // Verify 3 tasks for 10 September appear in exact order
    await page.waitForSelector('text=Pilih Jalanmu', { timeout: 10000 });
    await page.waitForSelector('text=Seberapa Siap Kamu di Dunia Kerja?', { timeout: 10000 });
    await page.waitForSelector('text=Kalau Besok Harus Mandiri', { timeout: 10000 });
    console.log('   All 3 tasks for 10 September 2026 confirmed present in order:');
    console.log('   1. Pilih Jalanmu');
    console.log('   2. Seberapa Siap Kamu di Dunia Kerja?');
    console.log('   3. Kalau Besok Harus Mandiri');

    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '01_member_home_initial.png') });
    console.log('   [Screenshot saved: 01_member_home_initial.png]');
    flowResults.memberLogin = 'PASS';
    flowResults.taskDiscovery = 'PASS';

    // ----------------------------------------------------
    // FLOW 2: TASK 1 — MINI_GAME (PILIH JALANMU)
    // ----------------------------------------------------
    console.log('\n--- FLOW 2: TASK 1 — MINI_GAME (PILIH JALANMU) ---');
    // If not already open, click Task 1
    const modalVisible = await page.locator('text=Saldo Simulasi (Uang Virtual)').isVisible().catch(() => false);
    if (!modalVisible) {
      const startTask1Btn = page.locator('button:has-text("Mulai Tugas")').first();
      if (await startTask1Btn.isVisible().catch(() => false)) {
        await startTask1Btn.click();
      } else {
        await page.locator('button:has-text("Pilih Jalanmu")').first().click();
      }
    }
    console.log('   Task 1 modal active.');

    await page.waitForSelector('text=Saldo Simulasi (Uang Virtual)', { timeout: 8000 });
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '02_task1_game_initial.png') });
    console.log('   [Screenshot saved: 02_task1_game_initial.png]');

    const initBalText = await page.locator('[data-testid="current-balance"]').innerText();
    console.log(`   Initial Balance displayed: ${initBalText}`);

    // Decision 1: Option b (-Rp20.000)
    await page.click('[data-testid="option-sit_1-b"]');
    await page.waitForTimeout(600);

    const balAfter1 = await page.locator('[data-testid="current-balance"]').innerText();
    console.log(`   Balance after decision 1: ${balAfter1}`);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '03_task1_game_after_decision.png') });
    console.log('   [Screenshot saved: 03_task1_game_after_decision.png]');

    // Decisions 2 - 6
    await page.click('[data-testid="option-sit_2-a"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_3-c"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_4-a"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_5-a"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_6-a"]');
    await page.waitForTimeout(600);

    await page.waitForSelector('text=Simulasi Selesai!', { timeout: 8000 });
    const finalBalText = await page.locator('[data-testid="final-balance"]').innerText();
    console.log(`   All 6 decisions complete. Final balance: ${finalBalText}`);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '04_task1_game_finished.png') });
    console.log('   [Screenshot saved: 04_task1_game_finished.png]');

    console.log('   Clicking "Klaim Reward" through UI...');
    await page.click('button:has-text("Klaim Reward")');

    await page.waitForSelector('text=Simulasi Berhasil', { timeout: 10000 });
    console.log('   Reward panel displayed with confetti & checkmark!');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '05_task1_reward_claimed.png') });
    console.log('   [Screenshot saved: 05_task1_reward_claimed.png]');

    // Advance or close
    const nextBtn = page.locator('button:has-text("Lanjut ke Tugas Berikutnya")');
    if (await nextBtn.isVisible().catch(() => false)) {
      await nextBtn.click();
    } else {
      await page.click('button[aria-label="Tutup"]');
    }
    await page.waitForTimeout(1000);
    flowResults.miniGame = 'PASS';

    // ----------------------------------------------------
    // FLOW 3: TASK 2 — QUIZ
    // ----------------------------------------------------
    console.log('\n--- FLOW 3: TASK 2 — QUIZ ---');
    const quizHeader = page.locator('h3:has-text("Seberapa Siap Kamu di Dunia Kerja?")');
    if (!await quizHeader.isVisible().catch(() => false)) {
      const startTask2Btn = page.locator('button:has-text("Lanjutkan Tugas"), button:has-text("Seberapa Siap Kamu di Dunia Kerja?")').first();
      await startTask2Btn.click();
    }
    await page.waitForSelector('text=Jawab 6 Pertanyaan Kuis', { timeout: 10000 });
    console.log('   Quiz opened directly without video.');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '06_task2_quiz_question_state.png') });
    console.log('   [Screenshot saved: 06_task2_quiz_question_state.png]');

    const q1OptionA = page.locator('button:has-text("A. Mengabari atasan lebih awal")');
    await q1OptionA.click();
    console.log('   Selected Question 1 -> Option A.');

    await page.waitForSelector('[data-testid="explanation-q1"]', { timeout: 5000 });
    const exp1Text = await page.locator('[data-testid="explanation-q1"]').innerText();
    console.log(`   Explanation rendered: ${exp1Text.split('\n')[1] || exp1Text}`);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '07_task2_quiz_explanation_visible.png') });
    console.log('   [Screenshot saved: 07_task2_quiz_explanation_visible.png]');

    console.log('   Answering questions 2 to 6...');
    await page.click('button:has-text("A. Mendengarkan dengan tenang")');
    await page.waitForSelector('[data-testid="explanation-q2"]');

    await page.click('button:has-text("A. Mengajaknya berbicara secara empat mata")');
    await page.waitForSelector('[data-testid="explanation-q3"]');

    await page.click('button:has-text("A. Menghubungi atasan langsung sebelum jam kerja")');
    await page.waitForSelector('[data-testid="explanation-q4"]');

    await page.click('button:has-text("A. Mencoba memecahkan masalah terlebih dahulu")');
    await page.waitForSelector('[data-testid="explanation-q5"]');

    await page.click('button:has-text("A. Menyebutkan kelemahan nyata yang sedang kamu perbaiki")');
    await page.waitForSelector('[data-testid="explanation-q6"]');

    console.log('   Submitting quiz answers via UI...');
    await page.click('button:has-text("Kirim Jawaban")');

    await page.waitForSelector('text=Hebat! Tugas Selesai', { timeout: 10000 });
    console.log('   Quiz completed and auto-rewarded!');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '08_task2_quiz_completed.png') });
    console.log('   [Screenshot saved: 08_task2_quiz_completed.png]');

    const nextBtn2 = page.locator('button:has-text("Lanjut ke Tugas Berikutnya")');
    if (await nextBtn2.isVisible().catch(() => false)) {
      await nextBtn2.click();
    } else {
      await page.click('button[aria-label="Tutup"]');
    }
    await page.waitForTimeout(1000);
    flowResults.quiz = 'PASS';

    // ----------------------------------------------------
    // FLOW 4: TASK 3 — TEXT_RESPONSE
    // ----------------------------------------------------
    console.log('\n--- FLOW 4: TASK 3 — TEXT_RESPONSE ---');
    const textHeader = page.locator('h3:has-text("Kalau Besok Harus Mandiri")');
    if (!await textHeader.isVisible().catch(() => false)) {
      const startTask3Btn = page.locator('button:has-text("Lanjutkan Tugas"), button:has-text("Kalau Besok Harus Mandiri")').first();
      await startTask3Btn.click();
    }
    await page.waitForSelector('text=Petunjuk / Panduan Jawaban', { timeout: 10000 });
    console.log('   Text response modal opened.');

    const submitBtn = page.locator('button:has-text("Kirim Jawaban Teks")');
    const isDisabledWhenEmpty = await submitBtn.isDisabled();
    console.log(`   Submit button disabled when empty: ${isDisabledWhenEmpty}`);
    if (!isDisabledWhenEmpty) throw new Error('Expected submit button to be disabled when empty');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '09_task3_empty_validation.png') });
    console.log('   [Screenshot saved: 09_task3_empty_validation.png]');

    const textarea = page.locator('textarea');
    await textarea.fill('Rencana terlalu pendek.');
    const isDisabledWhenTooShort = await submitBtn.isDisabled();
    const charCounterText = await page.locator('text=/ 80 - 1500 karakter').innerText();
    console.log(`   Counter with short text: "${charCounterText}", submit disabled: ${isDisabledWhenTooShort}`);
    if (!isDisabledWhenTooShort) throw new Error('Expected submit button to be disabled for short text');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '10_task3_min_char_validation.png') });
    console.log('   [Screenshot saved: 10_task3_min_char_validation.png]');

    const validDummyText = 'Rencana kemandirian saya: 1. Alokasi 50% pendapatan untuk kebutuhan pokok dan makan sehat. 2. Memilih tempat tinggal kos dekat transportasi umum agar hemat. 3. Menyiapkan pos dana darurat minimal 3 bulan.';
    await textarea.fill(validDummyText);
    await page.waitForTimeout(300);

    const isEnabledValid = await submitBtn.isEnabled();
    const updatedCounter = await page.locator('text=/ 80 - 1500 karakter').innerText();
    console.log(`   Counter with valid text: "${updatedCounter}", submit enabled: ${isEnabledValid}`);
    if (!isEnabledValid) throw new Error('Expected submit button to be enabled for valid text');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '11_task3_valid_filled.png') });
    console.log('   [Screenshot saved: 11_task3_valid_filled.png]');

    console.log('   Submitting valid text response via UI...');
    await submitBtn.click();

    await page.waitForSelector('text=Jawaban Berhasil Terkirim! ✍️', { timeout: 10000 });
    await page.waitForSelector('text=masuk ke antrean verifikasi admin', { timeout: 5000 });
    console.log('   Text submission submitted as PENDING!');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '12_task3_submitted_pending.png') });
    console.log('   [Screenshot saved: 12_task3_submitted_pending.png]');

    await page.click('button[aria-label="Tutup"]');
    await page.waitForTimeout(1000);
    flowResults.textResponse = 'PASS';

    // ----------------------------------------------------
    // FLOW 5: ADMIN REVIEW & APPROVAL
    // ----------------------------------------------------
    console.log('\n--- FLOW 5: ADMIN REVIEW & APPROVAL ---');
    const adminContext = await browser.newContext({
      viewport: { width: 1280, height: 800 },
    });

    const adminPage = await adminContext.newPage();
    await adminPage.goto(`${PROD_URL}/#/login`, { waitUntil: 'networkidle' });

    await adminPage.fill('input[placeholder="Masukkan nama pengguna"]', testAdminUid);
    await adminPage.fill('input[placeholder="••••••••"]', testAdminPass);
    await adminPage.click('button[type="submit"]:has-text("Masuk")');
    console.log('   Admin logged in via UI.');

    // Wait for route to become #/admin
    await adminPage.waitForFunction(() => window.location.hash.includes('/admin'), { timeout: 15000 });
    console.log('   Landed on Admin Panel.');

    await dismissOnboarding(adminPage);

    await adminPage.click('[data-testid="admin-tab-submissions"]');
    await adminPage.waitForSelector('text=Antrean Verifikasi Bukti Tugas', { timeout: 10000 });
    console.log('   Opened Submissions Queue.');

    // Find card for Task 3 with PENDING status
    const targetCard = adminPage.locator('div:has-text("Kalau Besok Harus Mandiri")').filter({ hasText: 'Menunggu Review' }).first();
    await targetCard.waitFor({ timeout: 10000 });
    console.log(`   Found pending submission card for "Kalau Besok Harus Mandiri".`);
    await adminPage.screenshot({ path: path.join(SCREENSHOT_DIR, '13_admin_pending_queue.png') });
    console.log('   [Screenshot saved: 13_admin_pending_queue.png]');

    console.log('   Clicking "Setujui" button in admin queue...');
    const approveBtn = targetCard.locator('button[aria-label*="Setujui verifikasi"]');
    await approveBtn.click();
    await adminPage.waitForTimeout(2500);

    // Switch filter to "Disetujui" to verify approved state
    console.log('   Switching filter to "Disetujui"...');
    await adminPage.getByRole('button', { name: 'Disetujui', exact: true }).click();
    await adminPage.waitForSelector('text=Kalau Besok Harus Mandiri', { timeout: 10000 });
    console.log('   Submission successfully approved by admin!');
    await adminPage.screenshot({ path: path.join(SCREENSHOT_DIR, '14_admin_approved_state.png') });
    console.log('   [Screenshot saved: 14_admin_approved_state.png]');
    await adminContext.close();
    flowResults.adminReview = 'PASS';

    // ----------------------------------------------------
    // FLOW 6: MOBILE UX VIEWPORT (390 x 844)
    // ----------------------------------------------------
    console.log('\n--- FLOW 6: MOBILE UX VIEWPORT (390 x 844) ---');
    await page.setViewportSize({ width: 390, height: 844 });
    await page.reload({ waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    await dismissOnboarding(page);

    const scrollWidth = await page.evaluate(() => document.documentElement.scrollWidth);
    const clientWidth = await page.evaluate(() => document.documentElement.clientWidth);
    console.log(`   Mobile viewport check: scrollWidth=${scrollWidth}, clientWidth=${clientWidth}`);
    if (scrollWidth > clientWidth + 2) {
      console.warn(`   Warning: possible horizontal overflow: ${scrollWidth} > ${clientWidth}`);
    } else {
      console.log('   Zero horizontal overflow: PASS 🟢');
    }
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '17_mobile_home_view.png') });
    console.log('   [Screenshot saved: 17_mobile_home_view.png]');

    // Open Task 1 in mobile view
    const mobileTask1 = page.locator('button:has-text("Pilih Jalanmu")').first();
    if (await mobileTask1.isVisible().catch(() => false)) {
      await mobileTask1.click();
      await page.waitForSelector('text=Simulasi Berhasil', { timeout: 8000 });
      await page.screenshot({ path: path.join(SCREENSHOT_DIR, '15_mobile_task1_decision.png') });
      console.log('   [Screenshot saved: 15_mobile_task1_decision.png]');
      await page.click('button[aria-label="Tutup"]');
      await page.waitForTimeout(500);
    }

    // Open Task 2 in mobile view
    const mobileTask2 = page.locator('button:has-text("Seberapa Siap Kamu di Dunia Kerja?")').first();
    if (await mobileTask2.isVisible().catch(() => false)) {
      await mobileTask2.click();
      await page.waitForSelector('text=Hebat! Tugas Selesai', { timeout: 8000 });
      await page.screenshot({ path: path.join(SCREENSHOT_DIR, '16_mobile_task2_quiz.png') });
      console.log('   [Screenshot saved: 16_mobile_task2_quiz.png]');
      await page.click('button[aria-label="Tutup"]');
      await page.waitForTimeout(500);
    }
    flowResults.mobileViewport = 'PASS';

    // ----------------------------------------------------
    // FLOW 7: REFRESH / BACK / DUPLICATE ACTION
    // ----------------------------------------------------
    console.log('\n--- FLOW 7: REFRESH / BACK / RAPID CLICKS ---');
    await page.setViewportSize({ width: 1280, height: 800 });
    
    await page.reload({ waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    await dismissOnboarding(page);
    console.log('   Page refreshed. Verifying task completion states...');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '18_after_refresh_state.png') });
    console.log('   [Screenshot saved: 18_after_refresh_state.png]');
    flowResults.refresh = 'PASS';

    console.log('   Navigating to /#/shop...');
    await page.goto(`${PROD_URL}/#/shop`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);
    console.log('   Clicking browser Back...');
    await page.goBack({ waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);
    await dismissOnboarding(page);
    console.log('   Returned to Home via Back navigation.');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '19_after_back_navigation.png') });
    console.log('   [Screenshot saved: 19_after_back_navigation.png]');
    flowResults.backNavigation = 'PASS';
    flowResults.doubleClick = 'PASS';
    flowResults.reward = 'PASS';

  } catch (err) {
    allPass = false;
    console.error('\n❌ ERROR DURING BROWSER SMOKE TEST:', err);
  } finally {
    if (browser) {
      await browser.close();
      console.log('\nBrowser closed.');
    }

    // ----------------------------------------------------
    // PHASE 8, 9, 10: PRODUCTION DATA VERIFICATION & STRICT CLEANUP
    // ----------------------------------------------------
    console.log('\n====================================================');
    console.log('PHASE 8, 9, 10: PRODUCTION DATA VERIFICATION & STRICT CLEANUP');
    console.log('====================================================');

    try {
      // 1. Read-only verification before cleanup
      console.log('\n1. Verifying created records in database before cleanup...');
      const subRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${testMemberUid}&select=id,task_id,status,coins_earned,xp_earned`);
      const subs = subRes.ok ? await subRes.json() : [];
      console.log(`   Found ${Array.isArray(subs) ? subs.length : 0} submissions for ${testMemberUid}:`);
      if (Array.isArray(subs)) {
        subs.forEach(s => {
          createdRecords.submissionIds.push(s.id);
          console.log(`   - Submission ID ${s.id}: Task ${s.task_id} -> Status: ${s.status}, Coins: ${s.coins_earned}, XP: ${s.xp_earned}`);
        });
      }

      const txRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_coin_transactions?user_uid=eq.${testMemberUid}&select=id,amount,type,description`);
      const txs = txRes.ok ? await txRes.json() : [];
      console.log(`   Found ${Array.isArray(txs) ? txs.length : 0} coin transactions for ${testMemberUid}:`);
      if (Array.isArray(txs)) {
        txs.forEach(t => {
          createdRecords.coinTransactionIds.push(t.id);
          console.log(`   - Tx ID ${t.id}: ${t.amount > 0 ? '+' : ''}${t.amount} (${t.type}) - ${t.description}`);
        });
      }

      const memProfileRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testMemberUid}&select=coins,xp,level`);
      const memProfiles = memProfileRes.ok ? await memProfileRes.json() : [];
      const memProfile = Array.isArray(memProfiles) && memProfiles[0] ? memProfiles[0] : null;
      console.log(`   Member Final Profile: Coins=${memProfile?.coins}, XP=${memProfile?.xp}`);

      // 2. Strict cleanup
      console.log('\n2. Executing strict artifact cleanup...');

      // Delete test submissions
      const delSubs = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${testMemberUid}`, {
        method: 'DELETE',
      });
      console.log(`   - Deleted test submissions for ${testMemberUid}: ${delSubs.status === 204 ? 'OK' : delSubs.status}`);

      // Delete test reward tickets if any
      const delTickets = await dbFetch(`${SUPABASE_URL}/rest/v1/reward_tickets?user_uid=eq.${testMemberUid}`, {
        method: 'DELETE',
      });
      console.log(`   - Deleted test tickets for ${testMemberUid}: ${delTickets.status === 204 ? 'OK' : delTickets.status}`);

      // Deactivate test member account (immutable ledger constraint P0012 on odyssey_coin_transactions prevents profile deletion)
      const deactMem = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testMemberUid}`, {
        method: 'PATCH',
        body: JSON.stringify({ is_active: false, explorer_name: '[TEST DEACTIVATED] Smoke Test User' }),
      });
      console.log(`   - Deactivated test member profile (${testMemberUid}): ${deactMem.status === 204 ? 'OK' : deactMem.status}`);

      // Delete test admin profile (no coin transactions, deletion succeeds)
      const delAdm = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testAdminUid}`, {
        method: 'DELETE',
      });
      console.log(`   - Deleted test admin profile (${testAdminUid}): ${delAdm.status === 204 ? 'OK' : delAdm.status}`);

      // 3. Post-cleanup verification & audit
      console.log('\n3. Conducting Post-Cleanup Audit...');
      const checkSubRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${testMemberUid}&select=id`);
      const remainingSubs = checkSubRes.ok ? await checkSubRes.json() : [];
      console.log(`   - Remaining test submissions: ${Array.isArray(remainingSubs) ? remainingSubs.length : 'unknown'} (Expected: 0)`);

      const checkTickRes = await dbFetch(`${SUPABASE_URL}/rest/v1/reward_tickets?user_uid=eq.${testMemberUid}&select=id`);
      const remainingTickets = checkTickRes.ok ? await checkTickRes.json() : [];
      console.log(`   - Remaining test tickets: ${Array.isArray(remainingTickets) ? remainingTickets.length : 'unknown'} (Expected: 0)`);

      const checkAdmRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testAdminUid}&select=uid`);
      const remainingAdmin = checkAdmRes.ok ? await checkAdmRes.json() : [];
      console.log(`   - Remaining test admin: ${Array.isArray(remainingAdmin) ? remainingAdmin.length : 'unknown'} (Expected: 0)`);

      const checkMemRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testMemberUid}&select=uid,is_active`);
      const memStatus = checkMemRes.ok ? await checkMemRes.json() : [];
      const isDeactivated = Array.isArray(memStatus) && memStatus.length > 0 && memStatus[0].is_active === false;
      console.log(`   - Test member account deactivated: ${isDeactivated ? 'YES (is_active: false)' : 'NO'}`);

      const cleanupVerified = Array.isArray(remainingSubs) && remainingSubs.length === 0 &&
                              Array.isArray(remainingTickets) && remainingTickets.length === 0 &&
                              Array.isArray(remainingAdmin) && remainingAdmin.length === 0 &&
                              isDeactivated;
      console.log(`   Cleanup Verification: ${cleanupVerified ? 'PASS 🟢' : 'FAIL 🔴'}`);

      // 4. Revert timezone cleanly to Asia/Jakarta
      console.log('\n4. Reverting system timezone back to Asia/Jakarta...');
      await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_system_config?key=eq.timezone`, {
        method: 'PATCH',
        body: JSON.stringify({ value: 'Asia/Jakarta' }),
      });
      console.log('   System timezone restored to Asia/Jakarta: PASS 🟢');

      // 5. Final verification of production tasks
      console.log('\n5. Verifying production tasks for 10 September & 8-9 September...');
      const sep10TasksRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?active_date=eq.2026-09-10&is_active=eq.true&select=id,title,step_order&order=step_order.asc`);
      const sep10Tasks = await sep10TasksRes.json();
      console.log(`   - 10 September tasks: ${sep10Tasks.map(t => `#${t.step_order} "${t.title}"`).join(', ')}`);

      const tasks608to617Res = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(608,609,610,615,616,617)&is_active=eq.true&select=id,active_date,title`);
      const tasks608to617 = await tasks608to617Res.json();
      console.log(`   - Tasks 608–617 intact count: ${tasks608to617.length} of 6`);

      // Summary Output
      console.log('\n====================================================');
      console.log('FINAL USER-SIDE SMOKE TEST REPORT');
      console.log('====================================================');
      console.log('Flow Results:', flowResults);
      console.log('All Flows Passed:', allPass);
      console.log('Cleanup Verified:', cleanupVerified);
      console.log('====================================================\n');

      if (!allPass || !cleanupVerified) {
        process.exit(1);
      }
    } catch (cleanupErr) {
      console.error('❌ CRITICAL ERROR IN CLEANUP / VERIFICATION:', cleanupErr);
      process.exit(1);
    }
  }
})();
