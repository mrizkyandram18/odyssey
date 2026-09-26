require('dotenv').config();

// Applies Migration 093: Replace tasks for 27 + 28 September 2026.
// 1. Re-verifies 564/565 (2026-09-27) and 566/567 (2026-09-28) carry 0
//    submissions + 0 ledger rows, then exact-ID hard-deletes ONLY those
//    rows on their exact dates (ABORTS otherwise).
// 2. Idempotent pre-clean: deletes ONLY the 6 exact new titles on their
//    exact dates (previous partial run of THIS migration).
// 3. Inserts the 6 tasks (config mirrors
//    supabase/migrations/093_tasks_replace_2026_09_27_28.sql exactly).
// 4. Bumps odyssey_schema_version to 093_tasks_replace_2026_09_27_28.
// Does NOT touch any other date, task, or config.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const NEW_27 = [
  'Sehari Kembali Pakai Seragam',
  'Kalau Sekarang Masih Pakai Seragam',
  'Balik Jadi Anak Sekolah Sehari',
];
const NEW_28 = [
  'Seragam vs Dirimu Sekarang',
  'Kalau Bisa Balik ke Hari Pertama Sekolah',
  'Dari Seragam Sekolah ke Seragam Kerja',
];

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

function scenario() {
  const ev = (id, title, description, opts) => ({ id, title, description, options: opts });
  const o = (id, label, delta, hint) => ({ id, label, delta, hint });
  return {
    game: 'DECISION_PRIORITY',
    target_score: 100,
    scenario: {
      currency: 'POINTS',
      initial_balance: 0,
      events: [
        ev('ev_1', 'Situasi 1 - Bangun Kesiangan',
          'Alarm tidak berbunyi dan kamu terbangun 30 menit lebih siang dari biasanya. Seragam masih di jemuran, dan sekolah dimulai 45 menit lagi. Apa yang kamu lakukan?',
          [
            o('a', 'Tetap tenang: mandi cepat, kenakan seragam yang paling rapi tersedia, dan prioritaskan hal yang paling penting saja', 20, 'Panik hanya membuang waktu - yang kesiangan tapi tetap sistematis akan lebih siap berangkat'),
            o('b', 'Berlama-lama memilih seragam terbaik agar tetap tampil sempurna meski terlambat', 5, 'Penampilan penting, tetapi mengorbankan waktu berangkat membuatmu makin terlambat'),
            o('c', 'Putuskan bolos saja karena sudah kesiangan dan percuma berangkat', 0, 'Terlambat sedikit masih jauh lebih baik daripada tidak masuk sama sekali'),
          ]),
        ev('ev_2', 'Situasi 2 - Seragam Belum Siap',
          'Seragam yang ingin kamu pakai ternyata belum disetrika dan ada noda kecil di lengannya. Waktu tersisa 20 menit. Apa yang kamu lakukan?',
          [
            o('a', 'Rapikan secepatnya: kibas dan gantung agar rapi, bersihkan nodanya, atau pakai seragam cadangan yang bersih', 20, 'Solusi praktis mengalahkan kesempurnaan - seragam cadangan yang bersih lebih baik dari seragam favorit yang kotor'),
            o('b', 'Pakai saja apa adanya tanpa dirapikan, yang penting berangkat', 10, 'Berangkat itu bagus, tetapi kerapian seragam adalah bagian dari kedisiplinan siswa'),
            o('c', 'Menunggu orang tua pulang untuk menyiapkan seragammu', 0, 'Menunggu tanpa berbuat apa-apa hanya menghabiskan waktu yang tersisa'),
          ]),
        ev('ev_3', 'Situasi 3 - Lupa Menyiapkan Perlengkapan',
          'Semalam kamu lupa menyiapkan perlengkapan: buku pelajaran, alat tulis, dan bekal belum dikemas. Bel sekolah 30 menit lagi. Apa yang kamu lakukan?',
          [
            o('a', 'Buat daftar cepat sesuai jadwal hari ini, kemas yang wajib dulu (buku dan alat tulis), lalu bekal seadanya', 20, 'Daftar prioritas membuat persiapan kilat tetap lengkap untuk hal yang paling penting'),
            o('b', 'Masukkan semua buku yang ada ke tas tanpa mengecek jadwal agar aman', 5, 'Tas jadi berat dan jadwal tetap bisa terlewat - mengecek jadwal hanya butuh satu menit'),
            o('c', 'Berangkat tanpa perlengkapan dan berencana meminjam semua ke teman di sekolah', 0, 'Mengandalkan pinjaman merepotkan teman dan menunjukkan kamu tidak bertanggung jawab'),
          ]),
        ev('ev_4', 'Situasi 4 - Hampir Terlambat di Jalan',
          'Kamu sudah di jalan tetapi macet, dan gerbang sekolah tutup 10 menit lagi. Ada jalan pintas yang sepi dan tidak aman. Apa yang kamu lakukan?',
          [
            o('a', 'Tetap di jalur aman yang biasa, jalan cepat tapi tertib, dan terima konsekuensinya dengan jujur jika terlambat', 20, 'Keselamatan tidak bisa ditukar dengan kecepatan - keterlambatan bisa dijelaskan, kecelakaan tidak'),
            o('b', 'Meminta tumpangan orang asing yang lewat agar cepat sampai', 0, 'Naik kendaraan orang yang tidak dikenal sangat berbahaya, seberapa pun mendesaknya'),
            o('c', 'Memotong jalan lewat area sepi dan berbahaya demi mengejar waktu', 0, 'Jalan pintas yang tidak aman bisa berakibat jauh lebih buruk daripada terlambat'),
          ]),
        ev('ev_5', 'Situasi 5 - Ada Perlengkapan yang Tertinggal',
          'Sampai di sekolah, kamu sadar ada satu perlengkapan penting yang tertinggal di rumah (misalnya tugas atau alat untuk pelajaran pertama). Apa yang kamu lakukan?',
          [
            o('a', 'Jujur ke guru tentang apa yang terjadi, minta solusi, dan buat pengingat agar besok tidak terulang', 20, 'Kejujuran plus rencana perbaikan menunjukkan kedewasaan - guru menghargai siswa yang bertanggung jawab'),
            o('b', 'Menyalahkan adik atau orang tua di rumah karena tidak mengingatkanmu', 5, 'Menyalahkan orang lain tidak menyelesaikan masalah dan merusak hubungan'),
            o('c', 'Berbohong dengan alasan yang dibuat-buat agar tidak dimarahi', 0, 'Kebohongan merusak kepercayaan - sekali ketahuan, penjelasan jujur berikutnya sulit dipercaya'),
          ]),
      ],
    },
  };
}

async function run() {
  console.log('=== Applying Migration 093: Replace tasks for 2026-09-27 + 2026-09-28 ===');

  // 0. Re-audit legacy tasks immediately before mutating.
  console.log('[0] Re-audit legacy tasks 564/565/566/567...');
  const legacy = await get('odyssey_tasks?id=in.(564,565,566,567)&select=id,title,active_date,is_active');
  console.log('   legacy rows: ' + JSON.stringify(legacy.map((t) => [t.id, t.title, t.active_date])));
  const subIds = [];
  for (const id of [564, 565, 566, 567]) {
    const s = await get(`odyssey_task_submissions?task_id=eq.${id}&select=id,status`);
    console.log(`   submissions task ${id}: ${s.length}`);
    subIds.push(...s.map((x) => x.id));
  }
  let ledger = [];
  if (subIds.length) {
    ledger = await get(`odyssey_coin_transactions?select=id&reference_id=in.(${subIds.join(',')})`);
  }
  console.log(`   linked ledger rows: ${ledger.length}`);
  if (subIds.length || ledger.length) {
    throw new Error('ABORT: legacy tasks carry submissions/ledger — hard-delete unsafe. Deactivate instead.');
  }
  const rows27 = await get('odyssey_tasks?active_date=eq.2026-09-27&select=id');
  const rows28 = await get('odyssey_tasks?active_date=eq.2026-09-28&select=id');
  console.log('   all rows on 2026-09-27: ' + JSON.stringify(rows27.map((t) => t.id).sort()));
  console.log('   all rows on 2026-09-28: ' + JSON.stringify(rows28.map((t) => t.id).sort()));
  if (JSON.stringify(rows27.map((t) => t.id).sort((a, b) => a - b)) !== '[564,565]') {
    throw new Error('ABORT: unexpected rows on 2026-09-27 — audit before proceeding.');
  }
  if (JSON.stringify(rows28.map((t) => t.id).sort((a, b) => a - b)) !== '[566,567]') {
    throw new Error('ABORT: unexpected rows on 2026-09-28 — audit before proceeding.');
  }

  // 1. Exact-ID hard delete of legacy tasks.
  console.log('[1] Exact-ID delete of 564/565 (2026-09-27)...');
  const del1 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(564,565)&active_date=eq.2026-09-27`,
    { method: 'DELETE', headers }
  );
  if (!del1.ok) throw new Error(`DELETE 564/565 failed: ${del1.status} ${await del1.text()}`);
  console.log('[1] Exact-ID delete of 566/567 (2026-09-28)...');
  const del2 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(566,567)&active_date=eq.2026-09-28`,
    { method: 'DELETE', headers }
  );
  if (!del2.ok) throw new Error(`DELETE 566/567 failed: ${del2.status} ${await del2.text()}`);

  // 2. Idempotent pre-clean of THIS migration's titles only.
  console.log('[2] Idempotent pre-clean (exact new titles, exact dates)...');
  const enc = (s) => encodeURIComponent(s);
  const pre27 = await get(`odyssey_tasks?active_date=eq.2026-09-27&select=id,title&title=in.(${NEW_27.map(enc).join(',')})`);
  const pre28 = await get(`odyssey_tasks?active_date=eq.2026-09-28&select=id,title&title=in.(${NEW_28.map(enc).join(',')})`);
  for (const rows of [pre27, pre28]) {
    if (rows.length) {
      const ids = rows.map((t) => t.id).join(',');
      const d = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
      if (!d.ok) throw new Error(`Pre-clean DELETE failed: ${d.status} ${await d.text()}`);
      console.log(`   removed ${rows.length} partial-run row(s).`);
    }
  }
  if (!pre27.length && !pre28.length) console.log('   no partial-run rows found.');

  // 3. Insert 6 tasks (payload mirrors 093 SQL migration).
  console.log('[3] Inserting 6 tasks...');
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Sehari Kembali Pakai Seragam',
      description:
        'Masih menyimpan seragam sekolah yang layak dipakai? Hari ini waktunya memakainya lagi! Kenakan seragammu dengan rapi, ambil satu foto dirimu, lalu tulis catatan singkat: apa yang paling kamu ingat ketika memakai seragam itu, atau kesan apa yang muncul saat memakainya kembali.',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 1,
      active_date: '2026-09-27',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction:
          '1) Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). Pastikan penampilanmu rapi. 2) Ambil atau unggah 1 foto dirimu memakai seragam tersebut (boleh kamera langsung atau galeri). 3) Wajib tulis catatan singkat dengan 2 bagian: (a) Kenangan - apa yang paling kamu ingat ketika memakai seragam tersebut dulu, (b) Kesan - pengalaman atau perasaan apa yang muncul ketika memakainya kembali sekarang. Admin akan menolak submission tanpa foto memakai seragam, tanpa 2 bagian catatan, atau foto yang tidak pantas.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Kalau Sekarang Masih Pakai Seragam',
      description:
        'Bayangkan pagi ini kamu harus berangkat sekolah lagi! Kenakan seragam sekolahmu, lalu rekam video maksimal 60 detik seolah-olah kamu sedang bersiap berangkat sekolah. Dalam video, jelaskan satu hal yang berbeda antara dirimu sekarang dengan dirimu ketika masih sekolah.',
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-27',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        recording: {
          enabled: true,
          max_duration_seconds: 60,
          camera_facing: 'user',
          instruction:
            'Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). Rekam video maksimal 60 detik dengan struktur:\n1. Persiapan - perankan seolah-olah kamu sedang bersiap berangkat sekolah lagi pagi ini (misalnya merapikan seragam, menyiapkan tas, atau pamit berangkat)\n2. Perbedaan - jelaskan SATU hal yang berbeda antara dirimu sekarang dengan dirimu ketika masih sekolah (misalnya kebiasaan, cara berpikir, keberanian, atau tanggung jawab)\n\nBicaralah senatural mungkin seperti bercerita ke seorang teman. Jangan menyebut data pribadi sensitif dan jangan menjelekkan pihak tertentu.',
        },
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Balik Jadi Anak Sekolah Sehari',
      description:
        'Hari ini kamu balik jadi anak sekolah sehari! Alarm berbunyi, seragam belum siap, tas belum dikemas, dan jam terus berjalan. Hadapi 5 situasi pagi yang kacau ini dan pilih tindakan terbaik pada setiap situasi. Setiap pilihan ada konsekuensinya - yang paling siap dan tenang akan meraih poin tertinggi!',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 3,
      active_date: '2026-09-27',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: scenario(),
    },
    {
      family_id: 'demo-crew-1',
      title: 'Seragam vs Dirimu Sekarang',
      description:
        'Seragam yang sama, orang yang berbeda! Kenakan lagi seragam sekolahmu, lalu ambil satu foto dengan pose yang menunjukkan perbedaan dirimu sekarang dibanding masa sekolah. Tambahkan catatan singkat tentang perubahan tersebut - apa yang berubah darimu sejak dulu?',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 1,
      active_date: '2026-09-28',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction:
          '1) Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). 2) Ambil atau unggah 1 foto dirimu memakai seragam tersebut dengan pose yang menunjukkan perbedaan dirimu sekarang dibanding masa sekolah (misalnya pose percaya diri seperti pekerja, pose dengan barang khas aktivitasmu sekarang, atau pose yang membandingkan dulu dan kini). 3) Wajib tulis catatan singkat dengan 2 bagian: (a) Pose - jelaskan pose yang kamu pilih dan perbedaan apa yang ingin ditunjukkannya, (b) Perubahan - ceritakan satu perubahan terbesar dalam dirimu sejak masa sekolah. Admin akan menolak submission tanpa foto memakai seragam, tanpa 2 bagian catatan, atau foto yang tidak pantas.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Kalau Bisa Balik ke Hari Pertama Sekolah',
      description:
        'Bayangkan besok kamu harus memakai seragam dan kembali ke hari pertama sekolah! Tulis dua hal: apa yang akan kamu lakukan berbeda dibanding dulu, dan satu nasihat yang akan kamu berikan kepada dirimu saat itu. Jujur, reflektif, dan dari hati.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-28',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt:
          'Tulis jawabanmu dengan format:\n1) Yang akan dilakukan berbeda - jelaskan minimal 2 hal yang akan kamu lakukan berbeda jika kembali ke hari pertama sekolah (contoh: cara belajar, cara bergaul, keberanian mencoba hal baru, atau cara mengatur waktu), beserta alasannya:\n2) Nasihat untuk diriku dulu - tulis SATU nasihat yang akan kamu berikan kepada dirimu pada hari pertama sekolah itu, dan jelaskan kenapa nasihat itu penting:\n\nTulis dengan jujur dari pengalamanmu sendiri. Jangan mencantumkan data pribadi siapa pun. Admin akan menolak jawaban yang terlalu pendek, tanpa hal yang berbeda, tanpa nasihat, atau tanpa alasan yang masuk akal.',
        minimum_characters: 150,
        maximum_characters: 2000,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Dari Seragam Sekolah ke Seragam Kerja',
      description:
        'Dari seragam sekolah sampai ke titik ini - perjalanan yang panjang! Kenakan seragam sekolahmu, lalu rekam video maksimal 60 detik tentang perjalananmu dari masa sekolah sampai sekarang: apa yang berubah, apa yang tetap sama, dan satu hal yang kamu pelajari.',
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-28',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        recording: {
          enabled: true,
          max_duration_seconds: 60,
          camera_facing: 'user',
          instruction:
            'Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). Rekam video maksimal 60 detik dengan struktur:\n1. Yang berubah - ceritakan satu hal terbesar yang berubah dari masa sekolah sampai sekarang\n2. Yang tetap sama - ceritakan satu hal dari dirimu yang tidak berubah sampai sekarang\n3. Pelajaran - ceritakan satu hal terpenting yang kamu pelajari sepanjang perjalanan itu\n\nBicaralah senatural mungkin seperti bercerita ke seorang teman. Jangan menyebut data pribadi sensitif dan jangan menjelekkan pihak tertentu.',
        },
      },
    },
  ];
  const postRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks`, {
    method: 'POST',
    headers,
    body: JSON.stringify(tasks),
  });
  if (!postRes.ok) throw new Error(`INSERT failed: ${postRes.status} ${await postRes.text()}`);
  const created = await postRes.json();
  console.log(`   inserted ${created.length} tasks:`);
  created.forEach((t) => {
    console.log(`   - [ID ${t.id}] ${t.active_date} #${t.step_order}: "${t.title}" (${t.task_type}, ${t.evaluation_type}, ${t.reward_coins}c/${t.reward_xp}xp)`);
  });

  // 4. Verify exactly 3 active tasks per date.
  console.log('[4] Verifying 2026-09-27 + 2026-09-28 state...');
  for (const [date, titles] of [['2026-09-27', NEW_27], ['2026-09-28', NEW_28]]) {
    const after = await get(
      `odyssey_tasks?active_date=eq.${date}&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,target_scope,family_id,is_active,config&order=step_order.asc`
    );
    const active = after.filter((t) => t.is_active);
    console.log(`   ${date}: total rows=${after.length}, active=${active.length}`);
    active.forEach((t) => console.log(`   - [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp scope=${t.target_scope} fam=${t.family_id}`));
    if (active.length !== 3) throw new Error(`Expected exactly 3 active tasks on ${date}, found ${active.length}`);
    if (JSON.stringify(active.map((t) => t.title)) !== JSON.stringify(titles)) {
      throw new Error(`Title/step mismatch on ${date}: ` + JSON.stringify(active.map((t) => t.title)));
    }
  }

  // 5. Bump schema version.
  console.log('[5] Updating odyssey_schema_version...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_schema_version`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '093_tasks_replace_2026_09_27_28', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 093 Applied Successfully ===');
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
