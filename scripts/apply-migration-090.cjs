require('dotenv').config();

// Applies Migration 090: Replace tasks for 20 + 21 September 2026.
// 1. Re-verifies 550/551 (2026-09-20) and 552/553 (2026-09-21) carry 0
//    submissions + 0 ledger rows, then exact-ID hard-deletes ONLY those
//    rows on their exact dates (ABORTS otherwise).
// 2. Idempotent pre-clean: deletes ONLY the 6 exact new titles on their
//    exact dates (previous partial run of THIS migration).
// 3. Inserts the 6 tasks (config mirrors
//    supabase/migrations/090_tasks_replace_2026_09_20_21.sql exactly).
// 4. Bumps odyssey_schema_version to 090_tasks_replace_2026_09_20_21.
// Does NOT touch any other date, task, or config.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const NEW_20 = [
  'Kalau Ini Terjadi, Kamu Ngapain?',
  'Temukan Penyebabnya',
  'Cari Hal yang Sering Diabaikan',
];
const NEW_21 = [
  'Mana yang Lebih Masuk Akal?',
  'Bikin Cara yang Lebih Praktis',
  'Jelaskan Tanpa Membaca',
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
        ev('ev_1', 'Situasi 1 — Air Tumpah Dekat Stopkontak',
          'Kamu tidak sengaja menumpahkan segelas air di meja, dan genangannya mengalir ke arah stopkontak dan kabel charger yang masih tercolok. Apa tindakan pertamamu?',
          [
            o('a', 'Cabut charger/matikan listrik di bagian itu dulu, baru lap genangan air sampai kering', 20, 'Listrik + air itu berbahaya — amankan sumber bahayanya dulu sebelum membersihkan'),
            o('b', 'Langsung lap airnya dengan kain tanpa mencabut charger', 5, 'Membersihkan itu benar, tetapi tangan basah di dekat listrik yang masih menyala berisiko tersengat'),
            o('c', 'Biarkan saja, nanti juga kering sendiri', 0, 'Air yang dibiarkan dekat listrik bisa menyebabkan korsleting sebelum sempat kering'),
          ]),
        ev('ev_2', 'Situasi 2 — Antrean Panjang',
          'Kamu harus mengurus dokumen di loket. Antrean panjang dan nomor antreanmu baru dipanggil sekitar 15 menit lagi. Apa yang kamu lakukan?',
          [
            o('a', 'Siapkan semua dokumen dan syarat selagi antre, catat nomormu, dan perhatikan panggilan', 20, 'Waktu menunggu jadi produktif dan kamu siap saat giliran tiba'),
            o('b', 'Main HP untuk mengisi waktu tanpa menyiapkan apa pun', 5, 'Waktu terbuang dan kamu bisa kelabakan mencari dokumen saat dipanggil'),
            o('c', 'Tinggalkan antrean untuk jajan tanpa titip ke siapa pun', 0, 'Nomormu bisa terlewat dan kamu harus antre ulang dari awal'),
          ]),
        ev('ev_3', 'Situasi 3 — Dompet Ketinggalan',
          'Di pintu keluar parkir kamu baru sadar dompet ketinggalan di rumah, sedangkan kamu harus membayar parkir sekarang. Apa yang kamu lakukan?',
          [
            o('a', 'Jelaskan dengan jujur ke petugas, hubungi keluarga/teman terdekat untuk bantuan, dan catat sebagai pengingat', 20, 'Kejujuran + meminta bantuan dengan cara yang benar menyelesaikan masalah tanpa merugikan siapa pun'),
            o('b', 'Berdebat dengan petugas agar digratiskan', 5, 'Petugas hanya menjalankan tugasnya — berdebat membuang waktu dan tidak menyelesaikan apa pun'),
            o('c', 'Kabur melewati palang tanpa membayar', 0, 'Itu merugikan pengelola dan bisa berujung masalah yang jauh lebih besar'),
          ]),
        ev('ev_4', 'Situasi 4 — Teman Kelompok Hilang Kabar',
          'Tugas kelompok dikumpulkan besok. Satu anggota belum mengirim bagiannya dan tidak membalas chat sejak kemarin. Apa yang kamu lakukan?',
          [
            o('a', 'Hubungi dia baik-baik untuk memastikan kabarnya, tawarkan bantuan, siapkan rencana cadangan, dan kabari ketua/guru', 20, 'Komunikasi yang baik + rencana cadangan menyelamatkan tugas tanpa menyalahkan siapa pun'),
            o('b', 'Diam-diam kerjakan bagiannya sendiri tanpa memberi tahu siapa pun', 10, 'Tugas selesai, tetapi tanpa komunikasi masalah yang sama akan terulang dan pembagian jadi tidak adil'),
            o('c', 'Marah-marah di grup tanpa memberi solusi', 0, 'Emosi tanpa solusi merusak kerja sama dan tidak mendekatkan tugas ke selesai'),
          ]),
        ev('ev_5', 'Situasi 5 — Hujan Deras Tanpa Jas Hujan',
          'Sepulang sekolah hujan turun sangat deras, kamu tidak membawa jas hujan maupun payung, dan keluarga belum bisa menjemput. Apa yang kamu lakukan?',
          [
            o('a', 'Berteduh di tempat yang aman, kabari keluarga tentang posisimu, dan tunggu hujan reda', 20, 'Keselamatan dulu — memberi kabar membuat keluarga tenang dan tahu di mana mencarimu'),
            o('b', 'Terobos hujan deras berjalan kaki sampai rumah', 5, 'Bisa sampai lebih cepat, tetapi berisiko sakit, terpeleset, dan barang-barangmu rusak'),
            o('c', 'Berteduh di bawah pohon besar atau baliho saat ada petir', 0, 'Pohon dan baliho justru berbahaya saat petir dan angin kencang — pilih bangunan yang kokoh'),
          ]),
      ],
    },
  };
}

function quiz21() {
  return {
    questions: [
      {
        id: 'q1',
        question: 'Air galon A: isi 19 liter seharga Rp20.000. Air galon B: isi 15 liter seharga Rp17.000. Kualitas air setara. Mana yang lebih hemat per liter?',
        options: [
          'A. Galon A (sekitar Rp1.053 per liter)',
          'B. Galon B (sekitar Rp1.133 per liter)',
          'C. Sama hematnya',
          'D. Tidak bisa dibandingkan',
        ],
        correct_answer: 'A',
        explanation:
          'Hitung harga per liter: A = 20.000 / 19 = sekitar Rp1.053 per liter. B = 17.000 / 15 = sekitar Rp1.133 per liter. Galon A lebih murah per liter, jadi lebih hemat meski harga labelnya lebih mahal.',
      },
      {
        id: 'q2',
        question: 'Les dimulai pukul 15:00 dan perjalanan dari rumah memakan waktu 45 menit. Kapan waktu berangkat yang paling masuk akal?',
        options: [
          'A. Pukul 14:00 (45 menit perjalanan + 15 menit cadangan)',
          'B. Pukul 14:15 (tiba tepat 15:00, tanpa cadangan)',
          'C. Pukul 14:30 (tiba 15:15, terlambat 15 menit)',
          'D. Pukul 13:00 (tiba 1 jam lebih awal)',
        ],
        correct_answer: 'A',
        explanation:
          'Berangkat 14:15 memang tiba tepat 15:00 di atas kertas, tetapi tanpa cadangan sedikit pun — macet sebentar saja langsung terlambat. Berangkat 14:00 memberi 15 menit cadangan menghadapi hal tak terduga. Berangkat 14:30 jelas terlambat, dan 13:00 membuang satu jam tanpa perlu.',
      },
      {
        id: 'q3',
        question: 'Lantai koridor baru saja dipel dan masih licin, tetapi kamu harus lewat untuk ke kelas. Tindakan mana yang paling masuk akal?',
        options: [
          'A. Minta dipasangi tanda licin / tunggu sebentar sampai kering, lalu lewat sisi yang kering dengan pelan',
          'B. Berlari kecil agar cepat sampai ke sisi lain',
          'C. Melompat jauh melewati area yang basah',
          'D. Menyuruh teman lewat duluan sebagai percobaan',
        ],
        correct_answer: 'A',
        explanation:
          'Satu-satunya pilihan yang menghilangkan bahayanya dulu sebelum bergerak. Berlari dan melompat di lantai licin justru memperbesar risiko terpeleset, dan menjadikan teman sebagai percobaan membahayakan orang lain.',
      },
      {
        id: 'q4',
        question: 'Lampu LED: 10 watt, harga Rp35.000, tahan 15.000 jam. Lampu pijar: 60 watt, harga Rp10.000, tahan 1.000 jam. Dipakai 5 jam sehari. Mana yang lebih hemat untuk jangka panjang?',
        options: [
          'A. Lampu LED (lebih awet dan jauh lebih hemat listrik)',
          'B. Lampu pijar (harga belinya lebih murah)',
          'C. Sama hematnya',
          'D. Tidak bisa dibandingkan',
        ],
        correct_answer: 'A',
        explanation:
          'Dipakai 5 jam sehari, LED tahan 15.000 / 5 = 3.000 hari (sekitar 8 tahun). Lampu pijar hanya tahan 1.000 / 5 = 200 hari, sehingga butuh 15 kali ganti (Rp150.000) plus listrik sekitar 6 kali lipat (60 watt vs 10 watt). Harga beli yang murah tidak berarti hemat jangka panjang.',
      },
    ],
  };
}

async function run() {
  console.log('=== Applying Migration 090: Replace tasks for 2026-09-20 + 2026-09-21 ===');

  // 0. Re-audit legacy tasks immediately before mutating.
  console.log('[0] Re-audit legacy tasks 550/551/552/553...');
  const legacy = await get('odyssey_tasks?id=in.(550,551,552,553)&select=id,title,active_date,is_active');
  console.log('   legacy rows: ' + JSON.stringify(legacy.map((t) => [t.id, t.title, t.active_date])));
  const subIds = [];
  for (const id of [550, 551, 552, 553]) {
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
  const rows20 = await get('odyssey_tasks?active_date=eq.2026-09-20&select=id');
  const rows21 = await get('odyssey_tasks?active_date=eq.2026-09-21&select=id');
  console.log('   all rows on 2026-09-20: ' + JSON.stringify(rows20.map((t) => t.id).sort()));
  console.log('   all rows on 2026-09-21: ' + JSON.stringify(rows21.map((t) => t.id).sort()));
  if (JSON.stringify(rows20.map((t) => t.id).sort((a, b) => a - b)) !== '[550,551]') {
    throw new Error('ABORT: unexpected rows on 2026-09-20 — audit before proceeding.');
  }
  if (JSON.stringify(rows21.map((t) => t.id).sort((a, b) => a - b)) !== '[552,553]') {
    throw new Error('ABORT: unexpected rows on 2026-09-21 — audit before proceeding.');
  }

  // 1. Exact-ID hard delete of legacy tasks.
  console.log('[1] Exact-ID delete of 550/551 (2026-09-20)...');
  const del1 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(550,551)&active_date=eq.2026-09-20`,
    { method: 'DELETE', headers }
  );
  if (!del1.ok) throw new Error(`DELETE 550/551 failed: ${del1.status} ${await del1.text()}`);
  console.log('[1] Exact-ID delete of 552/553 (2026-09-21)...');
  const del2 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(552,553)&active_date=eq.2026-09-21`,
    { method: 'DELETE', headers }
  );
  if (!del2.ok) throw new Error(`DELETE 552/553 failed: ${del2.status} ${await del2.text()}`);

  // 2. Idempotent pre-clean of THIS migration's titles only.
  console.log('[2] Idempotent pre-clean (exact new titles, exact dates)...');
  const enc = (s) => encodeURIComponent(s);
  const pre20 = await get(`odyssey_tasks?active_date=eq.2026-09-20&select=id,title&title=in.(${NEW_20.map(enc).join(',')})`);
  const pre21 = await get(`odyssey_tasks?active_date=eq.2026-09-21&select=id,title&title=in.(${NEW_21.map(enc).join(',')})`);
  for (const rows of [pre20, pre21]) {
    if (rows.length) {
      const ids = rows.map((t) => t.id).join(',');
      const d = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
      if (!d.ok) throw new Error(`Pre-clean DELETE failed: ${d.status} ${await d.text()}`);
      console.log(`   removed ${rows.length} partial-run row(s).`);
    }
  }
  if (!pre20.length && !pre21.length) console.log('   no partial-run rows found.');

  // 3. Insert 6 tasks (payload mirrors 090 SQL migration).
  console.log('[3] Inserting 6 tasks...');
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Kalau Ini Terjadi, Kamu Ngapain?',
      description:
        'Hidup penuh kejadian tak terduga — yang membedakan adalah keputusanmu! Ada 5 situasi sehari-hari yang realistis: air tumpah dekat listrik, antrean panjang, dompet ketinggalan, teman kelompok yang hilang kabar, dan hujan deras. Pilih tindakan yang paling tepat untuk setiap situasi. Setiap pilihan ada konsekuensinya — bijaklah seperti dalam kehidupan nyata!',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-20',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: scenario(),
    },
    {
      family_id: 'demo-crew-1',
      title: 'Temukan Penyebabnya',
      description:
        'Kenapa kran menetes terus padahal sudah diputar kencang? Kenapa nilai turun padahal belajar lebih lama? Setiap masalah punya penyebab — dan orang yang bisa menemukan penyebabnya bisa memperbaikinya. Pilih SATU masalah nyata di sekitarmu, selidiki kemungkinan penyebabnya, jelaskan alasanmu, dan tawarkan solusinya.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-20',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt:
          'Tulis hasil penyelidikanmu dengan format:\n1) Masalah yang kamu amati (sebutkan masalah nyata yang bisa diamati, bukan khayalan):\n2) Kemungkinan penyebab — tulis minimal 2 dugaan penyebab yang berbeda:\n3) Alasan — untuk setiap penyebab, jelaskan mengapa kamu menduga itu (fakta/pengamatan apa yang mendukung dugaanmu?):\n4) Solusi — untuk setiap penyebab, tulis solusi yang masuk akal:\n5) Solusi utama pilihanmu — penyebab mana yang paling mungkin dan solusi mana yang akan kamu jalankan duluan, beserta alasannya:\n\nWajib masalah nyata yang kamu amati sendiri. Jangan menyalin dari internet — tulis dengan bahasamu. Hindari masalah medis/kesehatan, dan jangan menyarankan tindakan berbahaya. Admin akan menolak jawaban tanpa masalah yang jelas, tanpa minimal 2 penyebab, tanpa alasan yang masuk akal, atau tanpa solusi.',
        minimum_characters: 150,
        maximum_characters: 2000,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Cari Hal yang Sering Diabaikan',
      description:
        'Setiap hari kita melewati hal-hal kecil yang sebenarnya butuh perhatian: keran yang menetes, lampu yang kedip, sampah yang menumpuk di sudut, cat yang mengelupas, atau kabel yang berserakan. Temukan SATU hal kecil yang sering diabaikan di sekitarmu, foto, lalu jelaskan kenapa hal itu layak diperhatikan dan apa usulan perbaikanmu.',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-20',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction:
          '1) Cari SATU hal kecil yang sering diabaikan di rumah, sekolah, atau lingkungan sekitarmu (contoh: keran menetes, lampu kedip, sampah menumpuk, cat mengelupas, kabel berserakan). Jangan memotret orang lain tanpa izin dan jangan memotret hal yang memalukan/merugikan pihak tertentu. 2) Ambil/unggah 1 foto hal tersebut (boleh kamera langsung atau galeri). 3) Wajib tulis catatan dengan 4 bagian: (a) Temuan — apa yang kamu temukan dan di mana, (b) Penjelasan — kenapa hal ini bisa terjadi menurut pengamatanmu, (c) Kenapa penting — apa akibatnya jika terus diabaikan, (d) Usulan — perbaikan sederhana apa yang kamu sarankan dan siapa yang bisa melakukannya. Admin akan menolak submission tanpa foto, tanpa 4 bagian catatan, atau temuan yang tidak pantas.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Mana yang Lebih Masuk Akal?',
      description:
        'Jawaban yang terlihat murah, cepat, atau mudah belum tentu paling masuk akal! Ada 4 situasi sehari-hari soal harga, waktu, risiko, dan efisiensi. Hitung dan pikirkan baik-baik sebelum memilih — semua soal harus benar agar lolos. Bukan soal hafalan, tapi soal nalar!',
      task_type: 'QUIZ',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-21',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: quiz21(),
    },
    {
      family_id: 'demo-crew-1',
      title: 'Bikin Cara yang Lebih Praktis',
      description:
        'Setiap hari kita melakukan hal yang sama berulang-ulang: menyiapkan bekal, merapikan kamar, mencatat tugas, mencuci piring, berangkat sekolah. Pasti ada yang terasa ribet atau lambat! Pilih SATU aktivitas rutinmu, jelaskan caramu sekarang, lalu rancang cara yang lebih praktis beserta alasannya.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-21',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt:
          'Tulis usulan perbaikanmu dengan format:\n1) Aktivitas rutin pilihanmu (kegiatan nyata yang sering kamu lakukan):\n2) Cara sekarang — jelaskan langkah per langkah bagaimana kamu melakukannya selama ini:\n3) Masalahnya — bagian mana yang terasa ribet, lambat, atau boros (waktu/tenaga/biaya)?\n4) Cara baru usulmu — jelaskan langkah per langkah cara yang lebih praktis:\n5) Alasan — jelaskan kenapa cara barumu lebih praktis (hemat waktu berapa, hemat tenaga/biaya apa, atau lebih rapi/aman bagaimana):\n\nWajib aktivitas nyata yang kamu lakukan sendiri. Jangan menyalin dari internet — tulis dengan bahasamu. Usulan harus aman dan bisa benar-benar dilakukan. Admin akan menolak jawaban tanpa aktivitas yang jelas, tanpa cara sekarang, tanpa cara baru yang konkret, atau tanpa alasan yang masuk akal.',
        minimum_characters: 150,
        maximum_characters: 2000,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Jelaskan Tanpa Membaca',
      description:
        'Orang yang benar-benar paham bisa menjelaskan tanpa mencontek teks! Pilih SATU pengetahuan atau hal sederhana yang kamu kuasai — misalnya kenapa langit biru, cara menghemat baterai HP, cara mencuci tangan yang benar, atau pengetahuan praktis lainnya. Rekam video maksimal 60 detik yang MENGAJARKANNYA dari ingatanmu, seolah kamu mengajari seorang teman. Dilarang membaca teks/catatan selama merekam!',
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-21',
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
            'Rekam video maksimal 60 detik yang mengajarkan SATU pengetahuan sederhana SEPENUHNYA DARI INGATAN (tanpa membaca teks, catatan, atau layar):\n1. Pembuka — sebutkan apa yang akan kamu jelaskan dan kenapa itu berguna\n2. Penjelasan inti — jelaskan dengan kata-katamu sendiri + beri SATU contoh nyata agar mudah dipahami\n3. Penutup — rangkum dalam satu kalimat kunci yang mudah diingat\n\nDILARANG membaca teks/catatan/layar selama merekam — admin akan menolak video yang terlihat membaca. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau gambarmu sambil menjelaskan dengan suaramu. Pilih topik yang aman dan pantas. Jangan menyebut data pribadi sensitif. Bicaralah senatural mungkin seperti mengajari teman!',
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
  console.log('[4] Verifying 2026-09-20 + 2026-09-21 state...');
  for (const [date, titles] of [['2026-09-20', NEW_20], ['2026-09-21', NEW_21]]) {
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
    body: JSON.stringify([{ key: 'schema_version', value: '090_tasks_replace_2026_09_20_21', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 090 Applied Successfully ===');
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
