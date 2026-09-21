require('dotenv').config();

// Applies Migration 091: Replace tasks for 22 + 23 September 2026.
// 1. Re-verifies 554/555 (2026-09-22) and 556/557 (2026-09-23) carry 0
//    submissions + 0 ledger rows, then exact-ID hard-deletes ONLY those
//    rows on their exact dates (ABORTS otherwise).
// 2. Idempotent pre-clean: deletes ONLY the 6 exact new titles on their
//    exact dates (previous partial run of THIS migration).
// 3. Inserts the 6 tasks (config mirrors
//    supabase/migrations/091_tasks_replace_2026_09_22_23.sql exactly).
// 4. Bumps odyssey_schema_version to 091_tasks_replace_2026_09_22_23.
// Does NOT touch any other date, task, or config.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const NEW_22 = [
  'Bongkar Cara Kerjanya',
  'Bikin Petunjuk yang Nggak Bikin Bingung',
  'Apa yang Bisa Kamu Pelajari dari Ini?',
];
const NEW_23 = [
  'Uangnya Lari ke Mana?',
  'Cek Sebelum Percaya',
  'Cerita dari Pengalamanmu',
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
        ev('ev_1', 'Situasi 1 — Rice Cooker Jeglek Sebelum Matang',
          'Nasi belum matang tetapi tombol rice cooker sudah jeglek ke posisi warm. Kamu butuh nasinya matang untuk makan bersama. Apa yang kamu lakukan?',
          [
            o('a', 'Cek dulu penyebabnya: pastikan takaran air cukup dan bersihkan kerak di dasar panci, lalu masak ulang dengan benar', 20, 'Tombol jeglek itu mekanisme pengaman — memahami cara kerjanya berarti mencari penyebabnya, bukan memaksa alatnya'),
            o('b', 'Ganjal atau tahan tombol cook-nya agar tidak jeglek', 5, 'Memaksa mekanisme pengaman bisa merusak alat dan berbahaya — masalah aslinya tidak selesai'),
            o('c', 'Buang rice cooker-nya dan beli yang baru', 0, 'Tanpa tahu penyebabnya, alat baru bisa mengalami hal yang sama'),
          ]),
        ev('ev_2', 'Situasi 2 — Kamar Gerah Padahal Kipas Kencang',
          'Kamar terasa gerah dan pengap meski kipas angin sudah menyala paling kencang. Apa yang kamu lakukan?',
          [
            o('a', 'Buka jendela atau ventilasi agar udara bersirkulasi — kipas hanya menggerakkan udara, bukan mendinginkan ruangan', 20, 'Memahami cara kerja kipas: angin terasa sejuk karena membantu penguapan keringat, tetapi butuh udara segar yang masuk'),
            o('b', 'Dekatkan kipas ke wajah sedekat mungkin dan tutup rapat semua ventilasi', 5, 'Udara pengap hanya diputar-putar di tempat — ruangan tetap panas dan pengap'),
            o('c', 'Biarkan kipas menyala seharian penuh tanpa henti agar lama-lama dingin', 0, 'Kipas tidak menurunkan suhu ruangan; menyala terus hanya menambah tagihan listrik'),
          ]),
        ev('ev_3', 'Situasi 3 — Got Depan Rumah Meluap Tiap Hujan',
          'Setiap hujan deras, saluran air depan rumah meluap dan menggenangi jalan. Warga mulai resah. Apa tindakan yang paling tepat?',
          [
            o('a', 'Periksa salurannya: angkat sampah dan sumbatan yang menghalangi aliran, lalu usulkan jadwal bersih rutin', 20, 'Air meluap karena alirannya tersumbat — atasi penyebabnya, bukan gejalanya'),
            o('b', 'Tinggikan pagar dan teras rumahmu sendiri agar air tidak masuk', 10, 'Rumahmu aman sementara, tetapi saluran tetap mampet dan air makin meluap ke tetangga'),
            o('c', 'Biarkan saja, nanti juga surut sendiri kalau hujan berhenti', 0, 'Sumbatan tidak hilang sendiri; hujan berikutnya meluap lagi, bahkan bisa lebih parah'),
          ]),
        ev('ev_4', 'Situasi 4 — HP Panas Saat Main Sambil Ngecas',
          'HP terasa sangat panas karena dipakai bermain game sambil diisi daya, dan diletakkan di atas bantal. Apa yang kamu lakukan?',
          [
            o('a', 'Berhenti memakai HP untuk hal berat selama mengisi daya, dan pindahkan dari bantal agar panasnya bisa keluar', 20, 'Dua sumber panas sekaligus ditambah ventilasi tertutup — memahami sebabnya berarti menghilangkan penyebabnya'),
            o('b', 'Masukkan HP ke kulkas atau freezer sebentar agar cepat dingin', 5, 'Perubahan suhu ekstrem bisa menimbulkan embun di dalam HP dan merusak komponen'),
            o('c', 'Teruskan bermain, panas itu normal dan tidak apa-apa', 0, 'Panas berlebih yang dibiarkan mempercepat kerusakan baterai dan bisa berbahaya'),
          ]),
        ev('ev_5', 'Situasi 5 — Tagihan Listrik Naik Tanpa Sebab Jelas',
          'Tagihan listrik bulan ini naik padahal kamu merasa pemakaiannya sama saja. Keluarga ingin tahu uangnya lari ke mana. Apa yang kamu lakukan?',
          [
            o('a', 'Catat pemakaian tiap alat (berapa watt, berapa jam menyala), cabut yang standby, lalu atur prioritas pemakaian', 20, 'Memahami ke mana listrik pergi berarti mengukur dulu, bukan menebak-nebak'),
            o('b', 'Matikan semua lampu setiap hari meski masih dipakai', 5, 'Hemat sesaat tetapi penyebab sebenarnya tidak ketemu — bisa jadi alat lain yang boros'),
            o('c', 'Marahi anggota keluarga agar tidak memakai listrik', 0, 'Tanpa data, hemat jadi tebak-tebakan dan menyalahkan orang tanpa solusi'),
          ]),
      ],
    },
  };
}

function quiz23() {
  return {
    questions: [
      {
        id: 'q1',
        question: 'Kamu membawa uang Rp50.000 ke sekolah. Pengeluaran hari ini: ongkos Rp12.000, jajan Rp15.000, fotokopi Rp8.000. Berapa sisa uangmu?',
        options: [
          'A. Rp15.000',
          'B. Rp20.000',
          'C. Rp25.000',
          'D. Rp35.000',
        ],
        correct_answer: 'A',
        explanation:
          'Total pengeluaran = 12.000 + 15.000 + 8.000 = Rp35.000. Sisa = 50.000 - 35.000 = Rp15.000. Selalu jumlahkan dulu semua pengeluaran sebelum menghitung sisa.',
      },
      {
        id: 'q2',
        question: 'Pengeluaran bulan ini: (A) jajan Rp10.000 x 20 hari, (B) langganan Rp75.000, (C) bensin Rp50.000 x 2 kali, (D) pulsa Rp60.000. Mana pengeluaran TERBESAR?',
        options: [
          'A. Jajan (Rp200.000)',
          'B. Langganan (Rp75.000)',
          'C. Bensin (Rp100.000)',
          'D. Pulsa (Rp60.000)',
        ],
        correct_answer: 'A',
        explanation:
          'Hitung masing-masing: A = 10.000 x 20 = Rp200.000; B = Rp75.000; C = 50.000 x 2 = Rp100.000; D = Rp60.000. Pengeluaran kecil yang berulang (jajan) ternyata paling besar totalnya.',
      },
      {
        id: 'q3',
        question: 'Mi instan A: 5 bungkus seharga Rp15.000. Mi instan B: 12 bungkus seharga Rp33.600. Isinya sama. Mana yang lebih hemat per bungkus?',
        options: [
          'A. Produk A (Rp3.000 per bungkus)',
          'B. Produk B (Rp2.800 per bungkus)',
          'C. Sama hematnya',
          'D. Tidak bisa dibandingkan',
        ],
        correct_answer: 'B',
        explanation:
          'Harga per bungkus: A = 15.000 / 5 = Rp3.000. B = 33.600 / 12 = Rp2.800. Produk B lebih murah Rp200 per bungkus, jadi lebih hemat meski harga labelnya lebih mahal.',
      },
      {
        id: 'q4',
        question: 'Kamu terbiasa beli minum kemasan Rp8.000 setiap hari. Jika kebiasaan itu berlangsung 30 hari, berapa total uang yang keluar?',
        options: [
          'A. Rp240.000',
          'B. Rp180.000',
          'C. Rp80.000',
          'D. Rp8.000',
        ],
        correct_answer: 'A',
        explanation:
          'Total = 8.000 x 30 = Rp240.000 dalam sebulan. Pengeluaran kecil harian yang terlihat sepele bisa menjadi besar saat dikalikan waktu — membawa botol isi ulang bisa menghemat hampir semuanya.',
      },
    ],
  };
}

async function run() {
  console.log('=== Applying Migration 091: Replace tasks for 2026-09-22 + 2026-09-23 ===');

  // 0. Re-audit legacy tasks immediately before mutating.
  console.log('[0] Re-audit legacy tasks 554/555/556/557...');
  const legacy = await get('odyssey_tasks?id=in.(554,555,556,557)&select=id,title,active_date,is_active');
  console.log('   legacy rows: ' + JSON.stringify(legacy.map((t) => [t.id, t.title, t.active_date])));
  const subIds = [];
  for (const id of [554, 555, 556, 557]) {
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
  const rows22 = await get('odyssey_tasks?active_date=eq.2026-09-22&select=id');
  const rows23 = await get('odyssey_tasks?active_date=eq.2026-09-23&select=id');
  console.log('   all rows on 2026-09-22: ' + JSON.stringify(rows22.map((t) => t.id).sort()));
  console.log('   all rows on 2026-09-23: ' + JSON.stringify(rows23.map((t) => t.id).sort()));
  if (JSON.stringify(rows22.map((t) => t.id).sort((a, b) => a - b)) !== '[554,555]') {
    throw new Error('ABORT: unexpected rows on 2026-09-22 — audit before proceeding.');
  }
  if (JSON.stringify(rows23.map((t) => t.id).sort((a, b) => a - b)) !== '[556,557]') {
    throw new Error('ABORT: unexpected rows on 2026-09-23 — audit before proceeding.');
  }

  // 1. Exact-ID hard delete of legacy tasks.
  console.log('[1] Exact-ID delete of 554/555 (2026-09-22)...');
  const del1 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(554,555)&active_date=eq.2026-09-22`,
    { method: 'DELETE', headers }
  );
  if (!del1.ok) throw new Error(`DELETE 554/555 failed: ${del1.status} ${await del1.text()}`);
  console.log('[1] Exact-ID delete of 556/557 (2026-09-23)...');
  const del2 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(556,557)&active_date=eq.2026-09-23`,
    { method: 'DELETE', headers }
  );
  if (!del2.ok) throw new Error(`DELETE 556/557 failed: ${del2.status} ${await del2.text()}`);

  // 2. Idempotent pre-clean of THIS migration's titles only.
  console.log('[2] Idempotent pre-clean (exact new titles, exact dates)...');
  const enc = (s) => encodeURIComponent(s);
  const pre22 = await get(`odyssey_tasks?active_date=eq.2026-09-22&select=id,title&title=in.(${NEW_22.map(enc).join(',')})`);
  const pre23 = await get(`odyssey_tasks?active_date=eq.2026-09-23&select=id,title&title=in.(${NEW_23.map(enc).join(',')})`);
  for (const rows of [pre22, pre23]) {
    if (rows.length) {
      const ids = rows.map((t) => t.id).join(',');
      const d = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
      if (!d.ok) throw new Error(`Pre-clean DELETE failed: ${d.status} ${await d.text()}`);
      console.log(`   removed ${rows.length} partial-run row(s).`);
    }
  }
  if (!pre22.length && !pre23.length) console.log('   no partial-run rows found.');

  // 3. Insert 6 tasks (payload mirrors 091 SQL migration).
  console.log('[3] Inserting 6 tasks...');
  const docExt = ['pdf', 'docx', 'txt', 'jpg', 'jpeg', 'png'];
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Bongkar Cara Kerjanya',
      description:
        'Kenapa rice cooker bisa jeglek sendiri? Kenapa kipas tidak membuat ruangan dingin? Kenapa got meluap tiap hujan? Semua ada cara kerjanya! Hadapi 5 situasi sehari-hari, pilih tindakan yang menunjukkan kamu paham BAGAIMANA dan MENGAPA sesuatu bekerja seperti itu. Setiap pilihan ada konsekuensinya — yang paham prosesnya akan memilih dengan tepat!',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-22',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: scenario(),
    },
    {
      family_id: 'demo-crew-1',
      title: 'Bikin Petunjuk yang Nggak Bikin Bingung',
      description:
        'Pernah membaca petunjuk yang malah bikin bingung? Sekarang giliranmu membuat yang benar! Buat SATU panduan singkat yang jelas, berurutan, dan mudah diikuti orang lain — misalnya cara packing barang, cara memakai alat sederhana, cara melakukan aktivitas kerja atau aktivitas rutin. Panduan yang bagus membuat orang lain bisa berhasil tanpa bertanya-tanya.',
      task_type: 'DOCUMENT_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-22',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        instruction:
          'Buat 1 panduan praktis yang bisa langsung dipakai orang lain tanpa penjelasan tambahan. Pilih SATU topik yang kamu kuasai (contoh: cara packing barang, cara memakai alat sederhana, cara melakukan aktivitas kerja sederhana, atau cara melakukan aktivitas rutin). Isi panduan WAJIB memuat: (1) judul yang jelas, (2) tujuan panduan + siapa yang bisa mengikutinya, (3) daftar bahan, alat, atau syarat sebelum mulai (jika ada), (4) langkah-langkah BERNOMOR yang berurutan dan konkret (setiap langkah bisa langsung dilakukan tanpa menebak), (5) 1 tips agar berhasil atau kesalahan umum yang harus dihindari. Boleh diketik di HP lalu disimpan sebagai PDF/DOCX/TXT, atau ditulis tangan rapi lalu difoto jelas (JPG/PNG). Maksimal 4 MB. Admin akan menolak file yang tidak bisa dibuka, tulisan yang bukan panduan (misalnya curhat atau essay tanpa langkah), panduan tanpa langkah bernomor yang konkret, atau langkah yang tidak berurutan.',
        allowed_extensions: docExt,
        accepted_extensions: docExt,
        max_file_size_mb: 4,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Apa yang Bisa Kamu Pelajari dari Ini?',
      description:
        'Belajar tidak selalu dari buku — benda dan situasi di sekitarmu menyimpan pelajaran! Temukan SATU hal nyata di sekitar (benda, tanaman, hewan, bangunan, atau situasi sehari-hari), foto, lalu jelaskan apa yang kamu temukan, apa yang bisa dipelajari darinya, dan bagaimana pengetahuan itu bisa berguna dalam hidupmu.',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-22',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction:
          '1) Temukan SATU benda atau situasi nyata di sekitarmu (contoh: retakan di tembok, sarang semut, tanaman liar yang tumbuh di celah, tumpukan barang, atau situasi sehari-hari yang menarik). Jangan memotret orang lain tanpa izin dan jangan memotret hal yang memalukan atau merugikan pihak tertentu. 2) Ambil atau unggah 1 foto hal tersebut (boleh kamera langsung atau galeri). 3) Wajib tulis catatan dengan 3 bagian: (a) Temuan — apa yang kamu temukan dan di mana kamu menemukannya, (b) Pelajaran — apa yang bisa dipelajari dari temuan itu (bagaimana hal itu bisa terjadi atau bekerja menurut pengamatanmu), (c) Manfaat — bagaimana pengetahuan tersebut bisa berguna untukmu atau orang lain. Admin akan menolak submission tanpa foto, tanpa 3 bagian catatan, atau temuan yang tidak pantas.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Uangnya Lari ke Mana?',
      description:
        'Uang saku habis tapi bingung ke mana perginya? Saatnya jadi detektif keuangan! Ada 4 situasi pengeluaran sehari-hari: hitung totalnya, temukan sisa uangnya, bandingkan pilihan yang ada, dan lihat dampak kebiasaan kecil. Semua soal harus benar agar lolos — kalkulator boleh, yang penting paham cara menghitungnya!',
      task_type: 'QUIZ',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-23',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: quiz23(),
    },
    {
      family_id: 'demo-crew-1',
      title: 'Cek Sebelum Percaya',
      description:
        'Diskon 90 persen? Hadiah undian? Kabar yang bikin panik? Tidak semua informasi bisa langsung dipercaya! Pilih SATU contoh informasi, iklan, atau penawaran sederhana yang pernah kamu lihat, lalu jelaskan hal apa yang perlu diperiksa, tanda apa yang perlu dicurigai, informasi apa yang perlu diverifikasi, dan kenapa pemeriksaan itu penting.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-23',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt:
          'Tulis hasil pemeriksaanmu dengan format:\n1) Informasi pilihanmu (contoh informasi, iklan, atau penawaran sederhana yang pernah kamu lihat — tulis isi atau bunyinya dengan bahasamu):\n2) Hal yang perlu diperiksa — tulis minimal 2 hal yang harus dicek sebelum percaya (contoh: siapa pengirimnya, apakah ada syarat tersembunyi, apakah harganya masuk akal):\n3) Tanda yang perlu dicurigai — tulis minimal 2 tanda bahaya (contoh: disuruh buru-buru, diminta data pribadi atau uang muka, tidak ada alamat atau kontak yang jelas):\n4) Cara memverifikasi — jelaskan bagaimana kamu memastikan kebenarannya (contoh: cek ke sumber resmi, tanya orang dewasa yang paham, bandingkan dengan info lain):\n5) Alasan — jelaskan kenapa pemeriksaan seperti ini penting dan apa risikonya jika langsung percaya:\n\nWajib contoh yang nyata dan sederhana. Jangan menyebar ulang kabar bohong — tulis dengan bahasamu sebagai bahan latihan. Jangan mencantumkan data pribadi siapa pun. Admin akan menolak jawaban tanpa contoh yang jelas, tanpa hal yang diperiksa, tanpa tanda curiga, tanpa cara verifikasi, atau tanpa alasan yang masuk akal.',
        minimum_characters: 150,
        maximum_characters: 2000,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Cerita dari Pengalamanmu',
      description:
        'Setiap pengalaman menyimpan pelajaran — dan pelajaran itu lebih berharga saat dibagikan! Ceritakan SATU pengalaman nyata dari kerja, sekolah, kehidupan sehari-hari, organisasi, atau aktivitas lainnya. Jelaskan apa yang terjadi, apa yang kamu pelajari, dan apa yang akan kamu lakukan berbeda sekarang. Rekam dalam video maksimal 60 detik, senatural mungkin seperti bercerita ke teman!',
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-23',
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
            'Rekam video maksimal 60 detik yang menceritakan SATU pengalaman nyatamu dengan struktur:\n1. Kejadian — ceritakan apa yang terjadi (kapan, di mana, siapa yang terlibat, bagaimana kejadiannya)\n2. Pelajaran — apa yang kamu pelajari dari pengalaman itu\n3. Perubahan — apa yang akan kamu lakukan berbeda sekarang jika mengalami hal serupa, dan kenapa\n\nPilih pengalaman yang aman dan pantas untuk dibagikan. Jangan menyebut data pribadi sensitif (milikmu maupun orang lain) dan jangan menjelekkan pihak tertentu. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau gambarmu sambil bercerita dengan suaramu. Bicaralah senatural mungkin seperti bercerita ke seorang teman!',
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
  console.log('[4] Verifying 2026-09-22 + 2026-09-23 state...');
  for (const [date, titles] of [['2026-09-22', NEW_22], ['2026-09-23', NEW_23]]) {
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
    body: JSON.stringify([{ key: 'schema_version', value: '091_tasks_replace_2026_09_22_23', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 091 Applied Successfully ===');
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
