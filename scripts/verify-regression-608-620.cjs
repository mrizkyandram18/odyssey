require('dotenv').config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
};

async function runRegression() {
  console.log('====================================================');
  console.log('REGRESSION VERIFICATION (TASKS 608–620)');
  console.log('====================================================\n');

  const ids = [608, 609, 610, 615, 616, 617, 618, 619, 620];
  const res = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids.join(',')})&order=id.asc`, { headers });
  const tasks = await res.json();

  const expected = {
    608: {
      title: 'Komunikasi yang Baik di Dunia Kerja',
      active_date: '2026-09-08',
      step_order: 1,
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      reward_coins: 40,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
    609: {
      title: 'Cari Sudut Pandang Lain',
      active_date: '2026-09-08',
      step_order: 2,
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      reward_coins: 30,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
    610: {
      title: 'Bukti Diskusi',
      active_date: '2026-09-08',
      step_order: 3,
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      reward_coins: 30,
      reward_xp: 100,
      is_active: true,
      camera_only: true,
    },
    615: {
      title: 'Kamu Jadi Bos Keuanganmu',
      active_date: '2026-09-09',
      step_order: 1,
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      reward_coins: 50,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
    616: {
      title: 'Kenalin Dirimu dalam 60 Detik',
      active_date: '2026-09-09',
      step_order: 2,
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      reward_coins: 40,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
    617: {
      title: 'Cari Masalah, Jangan Cuma Mengeluh',
      active_date: '2026-09-09',
      step_order: 3,
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      reward_coins: 30,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
    618: {
      title: 'Pilih Jalanmu',
      active_date: '2026-09-10',
      step_order: 1,
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      reward_coins: 50,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
    619: {
      title: 'Seberapa Siap Kamu di Dunia Kerja?',
      active_date: '2026-09-10',
      step_order: 2,
      task_type: 'QUIZ',
      evaluation_type: 'AUTO',
      reward_coins: 40,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
    620: {
      title: 'Kalau Besok Harus Mandiri',
      active_date: '2026-09-10',
      step_order: 3,
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      reward_coins: 30,
      reward_xp: 100,
      is_active: true,
      camera_only: false,
    },
  };

  let allOk = true;
  for (const t of tasks) {
    const exp = expected[t.id];
    if (!exp) {
      console.log(`[?] Unexpected task ${t.id}`);
      continue;
    }

    const actualCamOnly = Boolean(t.config?.camera_only);
    const subRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_task_submissions?task_id=eq.${t.id}&select=id`, { headers });
    const subs = await subRes.json();

    const matches =
      t.title === exp.title &&
      t.active_date === exp.active_date &&
      t.step_order === exp.step_order &&
      t.task_type === exp.task_type &&
      t.evaluation_type === exp.evaluation_type &&
      t.reward_coins === exp.reward_coins &&
      t.reward_xp === exp.reward_xp &&
      t.is_active === exp.is_active &&
      actualCamOnly === exp.camera_only &&
      subs.length >= 1; // Legitimate existing submissions must be preserved

    if (matches) {
      console.log(`   [ID ${t.id}] ${t.active_date} #${t.step_order} "${t.title}": PASS 🟢`);
      console.log(`         type: ${t.task_type} | eval: ${t.evaluation_type} | rewards: ${t.reward_coins}c/${t.reward_xp}xp | camera_only: ${actualCamOnly} | submissions: ${subs.length}`);
    } else {
      console.log(`   [ID ${t.id}] MISMATCH 🔴:`);
      console.log('   Expected:', exp);
      console.log('   Actual:  ', { ...t, actualCamOnly, subCount: subs.length });
      allOk = false;
    }
  }

  if (tasks.length !== Object.keys(expected).length) {
    console.log(`🔴 Expected ${Object.keys(expected).length} tasks, found ${tasks.length}`);
    allOk = false;
  }

  console.log('\n====================================================');
  console.log(`REGRESSION CHECK: ${allOk ? 'NO REGRESSIONS DETECTED 🟢' : 'REGRESSION DETECTED 🔴'}`);
  console.log('====================================================');
  if (!allOk) process.exit(1);
}

runRegression().catch(err => {
  console.error(err);
  process.exit(1);
});
