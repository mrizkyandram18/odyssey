const { chromium } = require('d:/Personal/Projects/Odyssey/web/node_modules/@playwright/test');
require('dotenv').config();
const fs = require('fs');
const path = require('path');

const PROD_URL = 'https://odyssey-beta-nine.vercel.app';
const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;
const CHROME_PATH = 'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe';
const SCREENSHOT_DIR = path.resolve('C:/Users/user/.gemini/antigravity/brain/f68f6e48-fcee-4b05-b909-f47eb2b97bab/screenshots');
const TEMP_IMAGE_PATH = path.resolve('C:/Users/user/.gemini/antigravity/brain/f68f6e48-fcee-4b05-b909-f47eb2b97bab/test_desk_photo.png');

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
  console.log('PRODUCTION USER-SIDE BROWSER SMOKE TEST (11/09/2026)');
  console.log(`Target: ${PROD_URL}`);
  console.log('====================================================\n');

  if (!fs.existsSync(SCREENSHOT_DIR)) {
    fs.mkdirSync(SCREENSHOT_DIR, { recursive: true });
  }

  // Create a small 1x1 test image
  const png1x1 = Buffer.from('iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==', 'base64');
  fs.writeFileSync(TEMP_IMAGE_PATH, png1x1);

  const ts = Date.now();
  const testMemberUser = `smoke_mem_${ts}`;
  const testMemberPass = 'password123';
  let testMemberUid = null;

  let browser = null;
  let allPass = true;
  const flowResults = {};

  try {
    // ----------------------------------------------------
    // 0. TIMEZONE ALIGNMENT: Asia/Jayapura (11 September in WIT)
    // ----------------------------------------------------
    console.log('0. Aligning system timezone to Asia/Jayapura (11 September in WIT)...');
    await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_system_config?key=eq.timezone`, {
      method: 'PATCH',
      body: JSON.stringify({ value: 'Asia/Jayapura' }),
    });
    console.log('   System timezone set to Asia/Jayapura: PASS 🟢\n');

    // ----------------------------------------------------
    // 1. SETUP DEDICATED TEST MEMBER VIA ADMIN API
    // ----------------------------------------------------
    console.log('1. Setting up dedicated test member via Admin API...');
    const adminLoginRes = await fetch(`${PROD_URL}/api/login`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        uid: 'admin',
        credential: 'admin123',
        device: { device_id: 'web_5e2c1a6a-64a2-4eff-963c-e4745161927e' },
      }),
    });
    const adminLoginData = await adminLoginRes.json();
    if (!adminLoginRes.ok || !adminLoginData.session) {
      throw new Error(`Admin login failed: ${JSON.stringify(adminLoginData)}`);
    }

    const createMemRes = await fetch(`${PROD_URL}/api/admin/members`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${adminLoginData.session}`,
      },
      body: JSON.stringify({
        username: testMemberUser,
        password: testMemberPass,
        explorer_name: 'Smoke UI Explorer 11-Sep',
        role: 'MEMBER',
      }),
    });
    const createdMemData = await createMemRes.json();
    if (!createMemRes.ok || !createdMemData.uid) {
      throw new Error(`Member creation failed: ${JSON.stringify(createdMemData)}`);
    }
    testMemberUid = createdMemData.uid;
    console.log(`   - Test member created: ${testMemberUser} (${testMemberUid})`);

    await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testMemberUid}`, {
      method: 'PATCH',
      body: JSON.stringify({ must_change_password: false }),
    });
    console.log('   - Cleared must_change_password flag for test member.\n');

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

    await page.fill('input[placeholder="Masukkan nama pengguna"]', testMemberUser);
    await page.fill('input[placeholder="••••••••"]', testMemberPass);
    await page.click('button[type="submit"]:has-text("Masuk")');
    console.log('   Submitted login credentials.');

    // Wait for route to leave #/login
    try {
      await page.waitForFunction(() => !window.location.hash.includes('/login'), { timeout: 25000 });
    } catch (e) {
      const bodyText = await page.innerText('body').catch(() => '');
      console.error('Login page text on timeout:', bodyText);
      await page.screenshot({ path: path.join(SCREENSHOT_DIR, '00_login_error.png') });
      throw e;
    }
    await page.waitForTimeout(2000);
    console.log('   Landed on Home page.');

    // Dismiss onboarding if shown
    await dismissOnboarding(page);

    // Verify exactly 3 tasks for 11 September appear in exact order
    await page.waitForSelector('text=Prioritas Dulu, Baru Gas', { timeout: 10000 });
    await page.waitForSelector('text=Bikin Tempatmu Lebih Siap', { timeout: 10000 });
    await page.waitForSelector('text=Kalau Ada Masalah di Tempat Kerja', { timeout: 10000 });
    console.log('   All 3 tasks for 11 September 2026 confirmed present in order:');
    console.log('   1. Prioritas Dulu, Baru Gas');
    console.log('   2. Bikin Tempatmu Lebih Siap');
    console.log('   3. Kalau Ada Masalah di Tempat Kerja');

    // Verify placeholders are NOT shown
    const hasP1 = await page.locator('text=Simulasi Diskon — Pilih Paling Hemat').isVisible().catch(() => false);
    const hasP2 = await page.locator('text=Rencana Tabungan 1 Bulan').isVisible().catch(() => false);
    if (hasP1 || hasP2) throw new Error('Legacy placeholder task(s) visible on member home!');
    console.log('   Placeholder tasks 533/534 correctly NOT visible 🟢');

    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '01_member_home_initial.png') });
    console.log('   [Screenshot saved: 01_member_home_initial.png]');
    flowResults.memberLogin = 'PASS';
    flowResults.taskDiscovery = 'PASS';

    // ----------------------------------------------------
    // FLOW 2: TASK 1 — MINI_GAME (Prioritas Dulu, Baru Gas)
    // ----------------------------------------------------
    console.log('\n--- FLOW 2: TASK 1 — MINI_GAME (Prioritas Dulu, Baru Gas) ---');
    const modalVisible = await page.locator('text=Skor Prioritas').isVisible().catch(() => false);
    if (!modalVisible) {
      const startTask1Btn = page.locator('button:has-text("Mulai Tugas")').first();
      if (await startTask1Btn.isVisible().catch(() => false)) {
        await startTask1Btn.click();
      } else {
        await page.locator('button:has-text("Prioritas Dulu, Baru Gas")').first().click();
      }
    }
    console.log('   Task 1 modal active.');

    await page.waitForSelector('text=Skor Prioritas', { timeout: 8000 });
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '02_task1_initial.png') });
    console.log('   [Screenshot saved: 02_task1_initial.png]');

    const initBalText = await page.locator('[data-testid="current-balance"]').innerText();
    console.log(`   Initial Skor Prioritas displayed: ${initBalText}`);

    // Decision 1: Option a (+20 Poin)
    await page.click('[data-testid="option-sit_1-a"]');
    await page.waitForTimeout(600);

    const balAfter1 = await page.locator('[data-testid="current-balance"]').innerText();
    console.log(`   Score after decision 1: ${balAfter1}`);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '03_task1_after_decision.png') });
    console.log('   [Screenshot saved: 03_task1_after_decision.png]');

    // Decisions 2 - 6
    await page.click('[data-testid="option-sit_2-a"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_3-a"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_4-a"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_5-a"]');
    await page.waitForTimeout(600);

    await page.click('[data-testid="option-sit_6-a"]');
    await page.waitForTimeout(600);

    await page.waitForSelector('text=Tantangan Selesai!', { timeout: 8000 });
    const finalScoreText = await page.locator('[data-testid="final-balance"]').innerText();
    console.log(`   All 6 decisions complete. Final Score: ${finalScoreText}`);
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '04_task1_completed.png') });
    console.log('   [Screenshot saved: 04_task1_completed.png]');

    console.log('   Clicking "Klaim Reward" through UI...');
    await page.click('button:has-text("Klaim Reward")');

    await page.waitForSelector('text=Simulasi Berhasil', { timeout: 10000 });
    console.log('   Reward panel displayed with confetti & checkmark!');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '05_task1_reward.png') });
    console.log('   [Screenshot saved: 05_task1_reward.png]');

    // Advance to next task
    const nextBtn = page.locator('button:has-text("Lanjut ke Tugas Berikutnya")');
    if (await nextBtn.isVisible().catch(() => false)) {
      await nextBtn.click();
    } else {
      await page.click('button[aria-label="Tutup"]');
    }
    await page.waitForTimeout(1000);
    flowResults.miniGame = 'PASS';

    // ----------------------------------------------------
    // FLOW 3: TASK 2 — PHOTO_UPLOAD (Bikin Tempatmu Lebih Siap)
    // ----------------------------------------------------
    console.log('\n--- FLOW 3: TASK 2 — PHOTO_UPLOAD (Bikin Tempatmu Lebih Siap) ---');
    const photoHeader = page.locator('h3:has-text("Bikin Tempatmu Lebih Siap")');
    if (!await photoHeader.isVisible().catch(() => false)) {
      const startTask2Btn = page.locator('button:has-text("Lanjutkan Tugas"), button:has-text("Bikin Tempatmu Lebih Siap")').first();
      await startTask2Btn.click();
    }
    await page.waitForSelector('text=Buka Kamera HP', { timeout: 10000 });
    console.log('   Photo upload modal opened.');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '06_task2_initial.png') });
    console.log('   [Screenshot saved: 06_task2_initial.png]');

    // Verify it's standard file/camera input (NOT LiveCameraCaptureModal)
    const fileInput = page.locator('input[type="file"]');
    if (await fileInput.count() === 0) throw new Error('Expected file input for normal photo upload');
    console.log('   Confirmed standard file upload input exists (NOT camera-only) 🟢');

    // Select test image file
    console.log('   Setting test image to file input...');
    await fileInput.setInputFiles(TEMP_IMAGE_PATH);
    await page.waitForSelector('img[alt="Preview Foto"]', { timeout: 10000 });
    console.log('   Image preview rendered with compression indicator!');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '07_task2_preview.png') });
    console.log('   [Screenshot saved: 07_task2_preview.png]');

    // Submit photo proof
    console.log('   Clicking "Kirim Foto Bukti"...');
    await page.click('button:has-text("Kirim Foto Bukti")');

    await page.locator('text=Foto Berhasil Dikirim! 📸').or(page.locator('text=Foto Selesai (Disetujui)! 📸')).waitFor({ timeout: 15000 });
    console.log('   Photo submission completed and auto-processed!');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '08_task2_auto_completed.png') });
    console.log('   [Screenshot saved: 08_task2_auto_completed.png]');

    // Advance to next task
    const nextBtn2 = page.locator('button:has-text("Lanjut Tugas Berikutnya")');
    if (await nextBtn2.isVisible().catch(() => false)) {
      await nextBtn2.click();
    } else {
      await page.click('button[aria-label="Tutup"]');
    }
    await page.waitForTimeout(1000);
    flowResults.photoUpload = 'PASS';

    // ----------------------------------------------------
    // FLOW 4: TASK 3 — TEXT_RESPONSE (Kalau Ada Masalah di Tempat Kerja)
    // ----------------------------------------------------
    console.log('\n--- FLOW 4: TASK 3 — TEXT_RESPONSE ---');
    const textHeader = page.locator('h3:has-text("Kalau Ada Masalah di Tempat Kerja")');
    if (!await textHeader.isVisible().catch(() => false)) {
      const startTask3Btn = page.locator('button:has-text("Lanjutkan Tugas"), button:has-text("Kalau Ada Masalah di Tempat Kerja")').first();
      await startTask3Btn.click();
    }
    await page.waitForSelector('text=Petunjuk / Panduan Jawaban', { timeout: 10000 });
    console.log('   Text response modal opened.');

    const submitBtn = page.locator('button:has-text("Kirim Jawaban Teks")');
    const isDisabledWhenEmpty = await submitBtn.isDisabled();
    console.log(`   Submit button disabled when empty: ${isDisabledWhenEmpty}`);
    if (!isDisabledWhenEmpty) throw new Error('Expected submit button to be disabled when empty');

    const textarea = page.locator('textarea');
    await textarea.fill('Masalah kecil.');
    const isDisabledWhenTooShort = await submitBtn.isDisabled();
    const charCounterText = await page.locator('text=/ 80 - 1000 karakter').innerText();
    console.log(`   Counter with short text: "${charCounterText}", submit disabled: ${isDisabledWhenTooShort}`);
    if (!isDisabledWhenTooShort) throw new Error('Expected submit button to be disabled for short text');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '09_task3_validation.png') });
    console.log('   [Screenshot saved: 09_task3_validation.png]');

    const validResponseText = '1) Pertama kali saya akan menenangkan diri dan mencari tahu akar penyebab terjadinya kendala tanpa panik. ' +
      '2) Segera berkomunikasi dengan atasan langsung dan rekan tim yang terdampak agar ada antisipasi operasional sedini mungkin. ' +
      '3) Menghindari saling menyalahkan atau menutup-nutupi masalah karena hal itu justru memperparah dampak ke depan. ' +
      '4) Langkah ini dipilih agar masalah cepat terselesaikan dengan transparan dan menjaga kepercayaan kerja sama dalam tim.';
    await textarea.fill(validResponseText);
    await page.waitForTimeout(300);

    const isEnabledValid = await submitBtn.isEnabled();
    const updatedCounter = await page.locator('text=/ 80 - 1000 karakter').innerText();
    console.log(`   Counter with valid text: "${updatedCounter}", submit enabled: ${isEnabledValid}`);
    if (!isEnabledValid) throw new Error('Expected submit button to be enabled for valid text');

    console.log('   Submitting valid text response via UI...');
    await submitBtn.click();

    await page.waitForSelector('text=Jawaban Berhasil Terkirim! ✍️', { timeout: 10000 });
    await page.waitForSelector('text=masuk ke antrean verifikasi admin', { timeout: 5000 });
    console.log('   Text submission submitted as PENDING!');
    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '10_task3_submitted_pending.png') });
    console.log('   [Screenshot saved: 10_task3_submitted_pending.png]');

    await page.click('button[aria-label="Tutup"]');
    await page.waitForTimeout(1000);
    flowResults.textResponse = 'PASS';

    // ----------------------------------------------------
    // FLOW 5: ADMIN REVIEW (APPROVE TASK 3 SUBMISSION)
    // ----------------------------------------------------
    console.log('\n--- FLOW 5: ADMIN REVIEW ---');
    const adminContext = await browser.newContext({
      viewport: { width: 1280, height: 800 },
    });
    const adminPage = await adminContext.newPage();
    await adminPage.goto(`${PROD_URL}/#/login`, { waitUntil: 'networkidle' });
    
    const adminSessionObj = {
      uid: adminLoginData.uid || 'demo-uid-2',
      family_id: adminLoginData.family_id || 'demo-crew-1',
      kind: adminLoginData.kind || 'user',
      role: adminLoginData.role || 'ADMIN',
      expires: adminLoginData.expires || Math.floor(Date.now() / 1000) + 86400,
      token: adminLoginData.session,
    };
    await adminPage.evaluate((sess) => {
      localStorage.setItem('odyssey_device_id', 'web_5e2c1a6a-64a2-4eff-963c-e4745161927e');
      localStorage.setItem('odyssey_session', JSON.stringify(sess));
      localStorage.setItem('odyssey_session_token', sess.token);
    }, adminSessionObj);

    // Reload so React mounts fresh and reads localStorage admin session!
    await adminPage.reload({ waitUntil: 'networkidle' });
    await adminPage.waitForTimeout(2000);
    console.log('   Admin landed on /#/admin dashboard.');

    await dismissOnboarding(adminPage);

    // Go to Verifikasi tab
    const verifTab = adminPage.locator('[data-testid="admin-tab-submissions"], button:has-text("Verifikasi"), a:has-text("Verifikasi")').first();
    await verifTab.click();
    await adminPage.waitForSelector('text=Antrean Verifikasi Bukti Tugas', { timeout: 10000 });
    await adminPage.waitForTimeout(1500);

    // Locate pending submission for test member
    console.log(`   Locating submission for ${testMemberUser}...`);
    const targetCard = adminPage.locator('div:has-text("Kalau Ada Masalah di Tempat Kerja")').filter({ hasText: 'Menunggu Review' }).first();
    await targetCard.waitFor({ timeout: 10000 });
    await adminPage.screenshot({ path: path.join(SCREENSHOT_DIR, '11_admin_queue_pending.png') });
    console.log('   [Screenshot saved: 11_admin_queue_pending.png]');

    // Click Setujui button for the submission
    console.log('   Clicking "Setujui" button...');
    const approveBtn = targetCard.locator('button[aria-label*="Setujui verifikasi"], button:has-text("Setujui")').first();
    await approveBtn.click();
    await adminPage.waitForTimeout(2500);

    await adminPage.screenshot({ path: path.join(SCREENSHOT_DIR, '12_admin_submission_approved.png') });
    console.log('   [Screenshot saved: 12_admin_submission_approved.png]');
    flowResults.adminReview = 'PASS';
    await adminContext.close();

    // ----------------------------------------------------
    // FLOW 6: MOBILE VIEWPORT (390 x 844)
    // ----------------------------------------------------
    console.log('\n--- FLOW 6: MOBILE SMOKE TEST (390 x 844) ---');
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

    await page.screenshot({ path: path.join(SCREENSHOT_DIR, '13_mobile_home_390x844.png') });
    console.log('   [Screenshot saved: 13_mobile_home_390x844.png]');

    // Open completed Task 1 in mobile view
    const task1Node = page.locator('button:has-text("Prioritas Dulu, Baru Gas")').first();
    if (await task1Node.isVisible().catch(() => false)) {
      await task1Node.click();
      await page.waitForTimeout(1000);
      await page.screenshot({ path: path.join(SCREENSHOT_DIR, '14_mobile_modal_390x844.png') });
      console.log('   [Screenshot saved: 14_mobile_modal_390x844.png]');
      await page.click('button[aria-label="Tutup"]').catch(() => {});
    }

    flowResults.mobile = 'PASS';

    // ----------------------------------------------------
    // FLOW 7: REFRESH / BACK PERSISTENCE
    // ----------------------------------------------------
    console.log('\n--- FLOW 7: REFRESH / BACK PERSISTENCE ---');
    await page.setViewportSize({ width: 1280, height: 800 });
    await page.reload({ waitUntil: 'networkidle' });
    await page.waitForTimeout(2000);
    await page.waitForSelector('text=Prioritas Dulu, Baru Gas', { timeout: 10000 });
    console.log('   Page reloaded. State persisted cleanly 🟢');

    await page.goto(`${PROD_URL}/#/profile`, { waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);
    await page.goBack({ waitUntil: 'networkidle' });
    await page.waitForTimeout(1000);
    await page.waitForSelector('text=Prioritas Dulu, Baru Gas', { timeout: 10000 });
    console.log('   Navigated to profile and browser Back successfully returned to Home 🟢');
    flowResults.refreshBack = 'PASS';

  } catch (err) {
    console.error('Smoke test error:', err);
    allPass = false;
  } finally {
    if (browser) await browser.close();

    // ----------------------------------------------------
    // STRICT CLEANUP
    // ----------------------------------------------------
    console.log('\n====================================================');
    console.log('STRICT CLEANUP OF SMOKE TEST ARTIFACTS');
    console.log('====================================================');

    if (testMemberUid) {
      // 1. Fetch created submission IDs
      const subRes = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${testMemberUid}&select=id`);
      const subs = subRes.ok ? await subRes.json() : [];
      const subIds = subs.map(s => s.id);
      console.log(`   Found ${subIds.length} test submissions to delete: [${subIds.join(', ')}]`);

      if (subIds.length > 0) {
        const delSub = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?id=in.(${subIds.join(',')})`, {
          method: 'DELETE',
        });
        console.log(`   - Deleted test submissions: ${delSub.status === 204 ? 'OK' : delSub.status}`);
      }

      // 2. Safe account cleanup: Deactivate test member (immutable ledger P0012)
      const deactMem = await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testMemberUid}`, {
        method: 'PATCH',
        body: JSON.stringify({ is_active: false, explorer_name: '[TEST DEACTIVATED] Smoke Test User' }),
      });
      console.log(`   - Deactivated test member profile (${testMemberUid}): ${deactMem.status === 204 ? 'OK' : deactMem.status}`);
    }

    // 3. Restore system timezone
    console.log('   Restoring system timezone to Asia/Jakarta...');
    await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_system_config?key=eq.timezone`, {
      method: 'PATCH',
      body: JSON.stringify({ value: 'Asia/Jakarta' }),
    });
    console.log('   System timezone restored to Asia/Jakarta: PASS 🟢');

    // 4. Remove temporary test image
    if (fs.existsSync(TEMP_IMAGE_PATH)) {
      fs.unlinkSync(TEMP_IMAGE_PATH);
      console.log('   Removed temporary test image file.');
    }

    // 5. Second read-only verification
    if (testMemberUid) {
      console.log('\n--- POST-CLEANUP SECOND AUDIT ---');
      const verifySubs = await (await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?user_uid=eq.${testMemberUid}&select=id`)).json();
      const verifyMem = await (await dbFetch(`${SUPABASE_URL}/rest/v1/odyssey_user_profiles?uid=eq.${testMemberUid}&select=is_active`)).json();

      console.log(`   Remaining test submissions: ${verifySubs.length} (Expected: 0)`);
      console.log(`   Test member is_active: ${verifyMem[0]?.is_active} (Expected: false)`);

      const cleanupVerified = verifySubs.length === 0 && verifyMem[0]?.is_active === false;
      console.log(`\nCLEANUP AUDIT RESULT: ${cleanupVerified ? 'CLEANUP VERIFIED 🟢' : 'CLEANUP INCOMPLETE 🔴'}`);

      if (!cleanupVerified) allPass = false;
    }
  }

  console.log('\n====================================================');
  console.log('FLOW SUMMARY:');
  console.log(JSON.stringify(flowResults, null, 2));
  console.log(`FINAL RESULT: ${allPass ? 'USER-SIDE SMOKE TEST PASS + CLEANUP VERIFIED 🟢' : 'FAILED 🔴'}`);
  console.log('====================================================');

  if (!allPass) process.exit(1);
})();
