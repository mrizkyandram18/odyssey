require('dotenv').config();

// Applies Migration 088: Replace tasks for 18 + 19 September 2026.
// 1. Re-verifies 546/547 (2026-09-18) and 548/549 (2026-09-19) carry 0
//    submissions + 0 ledger rows, then exact-ID hard-deletes ONLY those
//    rows on their exact dates (ABORTS otherwise).
// 2. Idempotent pre-clean: deletes ONLY the 6 exact new titles on their
//    exact dates (previous partial run of THIS migration).
// 3. Inserts the 6 tasks (config mirrors
//    supabase/migrations/088_tasks_replace_2026_09_18_19.sql exactly).
// 4. Bumps odyssey_schema_version to 088_tasks_replace_2026_09_18_19.
// Does NOT touch any other date, task, or config.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const NEW_18 = [
  'Kenapa Barang Ini Bisa Begitu?',
  'Uji Coba Sederhana',
  'Mitos atau Fakta?',
];
const NEW_19 = [
  'Hitung Mana yang Lebih Hemat?',
  'Cari Jalan Keluar dari Masalah',
  'Ajarkan Satu Skill',
];

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

async function run() {
  console.log('=== Applying Migration 088: Replace tasks for 2026-09-18 + 2026-09-19 ===');

  // 0. Re-audit legacy tasks immediately before mutating.
  console.log('[0] Re-audit legacy tasks 546/547/548/549...');
  const legacy = await get('odyssey_tasks?id=in.(546,547,548,549)&select=id,title,active_date,is_active');
  console.log('   legacy rows: ' + JSON.stringify(legacy.map((t) => [t.id, t.title, t.active_date])));
  const subIds = [];
  for (const id of [546, 547, 548, 549]) {
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
  const rows18 = await get('odyssey_tasks?active_date=eq.2026-09-18&select=id');
  const rows19 = await get('odyssey_tasks?active_date=eq.2026-09-19&select=id');
  console.log('   all rows on 2026-09-18: ' + JSON.stringify(rows18.map((t) => t.id).sort()));
  console.log('   all rows on 2026-09-19: ' + JSON.stringify(rows19.map((t) => t.id).sort()));
  if (JSON.stringify(rows18.map((t) => t.id).sort((a, b) => a - b)) !== '[546,547]') {
    throw new Error('ABORT: unexpected rows on 2026-09-18 — audit before proceeding.');
  }
  if (JSON.stringify(rows19.map((t) => t.id).sort((a, b) => a - b)) !== '[548,549]') {
    throw new Error('ABORT: unexpected rows on 2026-09-19 — audit before proceeding.');
  }

  // 1. Exact-ID hard delete of legacy tasks.
  console.log('[1] Exact-ID delete of 546/547 (2026-09-18)...');
  const del1 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(546,547)&active_date=eq.2026-09-18`,
    { method: 'DELETE', headers }
  );
  if (!del1.ok) throw new Error(`DELETE 546/547 failed: ${del1.status} ${await del1.text()}`);
  console.log('[1] Exact-ID delete of 548/549 (2026-09-19)...');
  const del2 = await fetch(
    `${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(548,549)&active_date=eq.2026-09-19`,
    { method: 'DELETE', headers }
  );
  if (!del2.ok) throw new Error(`DELETE 548/549 failed: ${del2.status} ${await del2.text()}`);

  // 2. Idempotent pre-clean of THIS migration's titles only.
  console.log('[2] Idempotent pre-clean (exact new titles, exact dates)...');
  const enc = (s) => encodeURIComponent(s);
  const pre18 = await get(`odyssey_tasks?active_date=eq.2026-09-18&select=id,title&title=in.(${NEW_18.map(enc).join(',')})`);
  const pre19 = await get(`odyssey_tasks?active_date=eq.2026-09-19&select=id,title&title=in.(${NEW_19.map(enc).join(',')})`);
  for (const rows of [pre18, pre19]) {
    if (rows.length) {
      const ids = rows.map((t) => t.id).join(',');
      const d = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
      if (!d.ok) throw new Error(`Pre-clean DELETE failed: ${d.status} ${await d.text()}`);
      console.log(`   removed ${rows.length} partial-run row(s).`);
    }
  }
  if (!pre18.length && !pre19.length) console.log('   no partial-run rows found.');

  // 3. Insert 6 tasks (payload mirrors 088 SQL migration).
  console.log('[3] Inserting 6 tasks...');
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Kenapa Barang Ini Bisa Begitu?',
      description:
        'Pilih SATU benda biasa yang ada di sekitarmu — misalnya charger yang menghangat, botol dengan bentuk tertentu, ban dengan pola kembang, remote yang harus diarahkan, atau benda sehari-hari lainnya. Cari tahu kenapa benda itu bekerja atau didesain seperti itu, lalu jelaskan temuanmu dengan kata-katamu sendiri. Tidak perlu ilmu tingkat lanjut dan tidak perlu membeli apa pun — cukup rasa ingin tahu dan kemauan mencari tahu.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 1,
      active_date: '2026-09-18',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt:
          'Tulis hasil penyelidikanmu dengan format:\n1) Benda yang kamu pilih (sebutkan nama bendanya):\n2) Pertanyaanmu tentang benda itu (contoh: kenapa charger jadi hangat? kenapa ban ada pola kembangnya?):\n3) Penjelasan yang kamu temukan (dengan kata-katamu sendiri, minimal 3 kalimat):\n4) Dari mana kamu mendapatkan penjelasan itu (misalnya: label kemasan, buku pelajaran, video edukasi, bertanya ke orang yang paham, artikel — tulis sejelas mungkin):\n5) Satu hal baru yang kamu pelajari:\n\nWajib memilih benda nyata yang bisa diamati. Jangan menyalin mentah-mentah dari internet — tulis ulang dengan bahasamu. Hindari klaim kesehatan/medis dan hindari percobaan berbahaya (tidak perlu membongkar atau mengubah apa pun). Admin akan menolak jawaban tanpa benda yang jelas, tanpa penjelasan dengan kata sendiri, tanpa sumber belajar, atau tanpa hal yang dipelajari.',
        minimum_characters: 150,
        maximum_characters: 2000,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Uji Coba Sederhana',
      description:
        'Lakukan SATU percobaan kecil yang aman dengan benda-benda yang sudah ada di rumah — misalnya menguji bahan mana yang paling cepat menyerap air, di mana es batu paling cepat mencair, atau susunan mana yang membuat benda lebih stabil. Tulis prediksimu SEBELUM mencoba, lakukan percobaannya, foto kondisi atau hasilnya, lalu tulis kesimpulanmu. Tidak boleh ada api, bahan kimia, listrik, benda tajam berbahaya, atau apa pun yang berisiko melukai.',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-18',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction:
          '1) Pilih SATU uji coba sederhana yang AMAN dari contoh di deskripsi, atau idemu sendiri yang setara amannya. DILARANG: api, bahan kimia/pembersih, mengutak-atik listrik/stopkontak, benda tajam, memakan/mencicipi bahan percobaan, atau apa pun yang bisa melukai dirimu, orang lain, atau merusak barang. 2) Tulis dulu PREDIKSImu (tebakan hasil sebelum mencoba). 3) Lakukan percobaannya dan ambil/unggah 1 foto kondisi percobaan atau hasilnya (boleh kamera langsung atau galeri). 4) Wajib tulis catatan dengan 4 bagian: (a) Prediksi, (b) Percobaan — apa yang kamu lakukan langkah per langkah, (c) Hasil — apa yang benar-benar terjadi, (d) Kesimpulan — apakah prediksimu benar dan apa yang kamu pelajari. Admin akan menolak submission tanpa foto, tanpa 4 bagian catatan, atau percobaan yang tidak aman.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Mitos atau Fakta?',
      description:
        'Uji ketajaman berpikirmu: ada 5 pernyataan sehari-hari yang sering beredar. Tentukan masing-masing MITOS atau FAKTA, dan baca penjelasannya setelah menjawab. Semua harus benar agar lolos — pikirkan baik-baik sebelum mengunci jawaban!',
      task_type: 'QUIZ',
      evaluation_type: 'AUTO',
      step_order: 3,
      active_date: '2026-09-18',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        questions: [
          {
            id: 'q1',
            question: 'Pernyataan: Petir tidak pernah menyambar tempat yang sama dua kali. Ini MITOS atau FAKTA?',
            options: ['A. MITOS', 'B. FAKTA'],
            correct_answer: 'A',
            explanation:
              'MITOS. Petir justru sering menyambar tempat yang sama berulang kali — gedung-gedung tinggi dan menara disambar petir berkali-kali setiap tahun. Tidak ada aturan alam yang melarang sambaran ulang.',
          },
          {
            id: 'q2',
            question: 'Pernyataan: Air selalu mendidih pada suhu 100 derajat Celsius, di mana pun. Ini MITOS atau FAKTA?',
            options: ['A. MITOS', 'B. FAKTA'],
            correct_answer: 'A',
            explanation:
              'MITOS. Air mendidih pada 100 derajat hanya pada tekanan udara normal di permukaan laut. Di dataran tinggi (misalnya daerah pegunungan) tekanan udara lebih rendah sehingga air mendidih di bawah 100 derajat.',
          },
          {
            id: 'q3',
            question: 'Pernyataan: Benda yang lebih berat selalu jatuh lebih cepat daripada benda yang ringan. Ini MITOS atau FAKTA?',
            options: ['A. MITOS', 'B. FAKTA'],
            correct_answer: 'A',
            explanation:
              'MITOS. Tanpa hambatan udara, semua benda jatuh sama cepatnya apa pun beratnya — sudah dibuktikan sejak percobaan Galileo dan eksperimen palu-bulu di Bulan. Di udara, yang berpengaruh adalah bentuk dan hambatan udara, bukan sekadar berat.',
          },
          {
            id: 'q4',
            question: 'Pernyataan: Mencuci tangan dengan sabun sambil menggosok minimal 20 detik membersihkan kuman lebih efektif daripada sekadar membilas dengan air. Ini MITOS atau FAKTA?',
            options: ['A. MITOS', 'B. FAKTA'],
            correct_answer: 'B',
            explanation:
              'FAKTA. Sabun melarutkan minyak dan kotoran tempat kuman menempel, dan gosokan selama sekitar 20 detik memberi waktu untuk mengangkatnya. Inilah anjuran standar organisasi kesehatan dunia.',
          },
          {
            id: 'q5',
            question: 'Pernyataan: Baterai HP harus selalu dikosongkan sampai 0% dulu sebelum diisi ulang agar awet. Ini MITOS atau FAKTA?',
            options: ['A. MITOS', 'B. FAKTA'],
            correct_answer: 'A',
            explanation:
              'MITOS. Aturan itu berlaku untuk baterai lama jenis nikel. Baterai lithium-ion modern justru lebih awet bila diisi sebelum habis total — pengosongan sampai 0% yang berulang malah mempercepat penurunan kapasitasnya.',
          },
        ],
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Hitung Mana yang Lebih Hemat?',
      description:
        'Harga termurah di label belum tentu paling hemat! Ada 4 perbandingan belanja sehari-hari: kemasan beda ukuran, diskon beda persen, dan paket beda isi. Hitung harga per satuannya (per ml, per liter, per kg, atau harga akhir setelah diskon) lalu pilih jawaban yang benar. Semua soal harus benar agar lolos.',
      task_type: 'QUIZ',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-19',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        questions: [
          {
            id: 'q1',
            question: 'Sabun cair A: Rp18.000 untuk 250 ml. Sabun cair B: Rp25.000 untuk 400 ml. Isinya sama bagusnya. Mana yang lebih hemat per ml?',
            options: ['A. Produk A (Rp72 per ml)', 'B. Produk B (Rp62,50 per ml)', 'C. Sama hematnya', 'D. Tidak bisa dibandingkan'],
            correct_answer: 'B',
            explanation:
              'Hitung harga per ml: A = 18.000 / 250 = Rp72 per ml. B = 25.000 / 400 = Rp62,50 per ml. Produk B lebih murah per ml, jadi lebih hemat meski harga labelnya lebih mahal.',
          },
          {
            id: 'q2',
            question: 'Minyak goreng A: 1 liter seharga Rp22.000. Minyak goreng B: 2 liter seharga Rp40.000. Mana yang lebih hemat per liter?',
            options: ['A. Minyak A (Rp22.000 per liter)', 'B. Minyak B (Rp20.000 per liter)', 'C. Sama hematnya', 'D. Tidak bisa dibandingkan'],
            correct_answer: 'B',
            explanation:
              'Harga per liter: A = Rp22.000 per liter. B = 40.000 / 2 = Rp20.000 per liter. Kemasan besar B lebih hemat Rp2.000 per liter.',
          },
          {
            id: 'q3',
            question: 'Tas A: Rp100.000 diskon 20%. Tas B: Rp90.000 diskon 10%. Kualitas setara. Mana harga akhir yang lebih murah?',
            options: ['A. Tas A (Rp80.000)', 'B. Tas B (Rp81.000)', 'C. Sama murahnya', 'D. Tidak bisa dibandingkan'],
            correct_answer: 'A',
            explanation:
              'Harga akhir: A = 100.000 x 0,8 = Rp80.000. B = 90.000 x 0,9 = Rp81.000. Diskon persen lebih besar tidak selalu berarti harga akhir lebih murah — selalu hitung angka akhirnya.',
          },
          {
            id: 'q4',
            question: 'Beras A: kemasan 5 kg seharga Rp65.000. Beras B: kemasan 10 kg seharga Rp120.000. Mana yang lebih hemat per kg?',
            options: ['A. Beras A (Rp13.000 per kg)', 'B. Beras B (Rp12.000 per kg)', 'C. Sama hematnya', 'D. Tidak bisa dibandingkan'],
            correct_answer: 'B',
            explanation:
              'Harga per kg: A = 65.000 / 5 = Rp13.000 per kg. B = 120.000 / 10 = Rp12.000 per kg. Beras B lebih hemat Rp1.000 per kg, total hemat Rp10.000 untuk 10 kg.',
          },
        ],
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Cari Jalan Keluar dari Masalah',
      description:
        'Simulasi keadaan nyata: malam hari kamu berada di tempat yang belum kamu kenal. Baterai HP tinggal 15%, kamu harus mengirim satu kabar penting ke keluarga, dan sampai di rumah dengan selamat dalam sekitar 30 menit. Hadapi 4 situasi berurutan dan pilih tindakan yang paling tepat untuk kondisimu. Setiap pilihan ada konsekuensinya — bijaklah seperti dalam kehidupan nyata!',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 2,
      active_date: '2026-09-19',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        game: 'DECISION_PRIORITY',
        target_score: 100,
        scenario: {
          currency: 'POINTS',
          initial_balance: 0,
          events: [
            {
              id: 'ev_1',
              title: 'Situasi 1 — Baterai Tinggal 15%',
              description: 'Kamu baru sadar baterai HP tinggal 15% di tempat asing, malam hari. Satu pesan penting belum terkirim. Apa langkah pertamamu?',
              options: [
                { id: 'a', label: 'Aktifkan mode hemat daya, kecilkan kecerahan layar, dan tutup aplikasi yang tidak perlu', delta: 20, hint: 'Menghemat setiap persen baterai untuk hal yang benar-benar penting' },
                { id: 'b', label: 'Biarkan semua pengaturan seperti biasa dan langsung buka peta online dengan kecerahan penuh', delta: 5, hint: 'Peta memang berguna, tetapi layar penuh dan data menyedot baterai dengan cepat' },
                { id: 'c', label: 'Buka media sosial dulu untuk mengisi waktu sambil berpikir', delta: 0, hint: 'Hiburan menghabiskan baterai tanpa mendekatkanmu ke solusi' },
              ],
            },
            {
              id: 'ev_2',
              title: 'Situasi 2 — Mengirim Kabar Penting',
              description: 'Keluarga menunggumu di rumah dan mulai khawatir. Baterai makin menipis. Bagaimana kamu mengabari mereka?',
              options: [
                { id: 'a', label: 'Kirim satu pesan teks singkat dan padat (lokasi, kondisimu, rencanamu), lalu matikan data seluler', delta: 20, hint: 'Satu pesan lengkap menenangkan keluarga sekaligus menghemat baterai' },
                { id: 'b', label: 'Telepon video panjang agar keluarga bisa melihat keadaanmu secara detail', delta: 5, hint: 'Video call menenangkan sesaat tetapi menguras baterai paling besar' },
                { id: 'c', label: 'Tidak mengabari sama sekali dan berharap sampai rumah sebelum mereka panik', delta: 0, hint: 'Tanpa kabar, keluarga khawatir dan tidak bisa membantumu jika terjadi sesuatu' },
              ],
            },
            {
              id: 'ev_3',
              title: 'Situasi 3 — Butuh Arah Jalan Pulang',
              description: 'Kamu tidak hafal jalan di daerah ini. Baterai tinggal sedikit. Bagaimana cara mendapatkan arah yang benar?',
              options: [
                { id: 'a', label: 'Lihat rute di peta selagi masih ada baterai, simpan tangkapan layarnya, lalu matikan data dan ikuti petunjuk itu', delta: 20, hint: 'Satu kali akses peta yang direncanakan memberi panduan tanpa baterai terus menyala' },
                { id: 'b', label: 'Bertanya arah ke petugas, penjaga, atau penjual di sekitar dan mencatat patokan jalannya', delta: 15, hint: 'Bertanya ke orang setempat itu cerdas — catat patokannya agar tidak lupa' },
                { id: 'c', label: 'Jalan mengikuti firasat tanpa panduan apa pun', delta: 0, hint: 'Di tempat asing, firasat tanpa informasi justru berisiko tersesat lebih jauh' },
              ],
            },
            {
              id: 'ev_4',
              title: 'Situasi 4 — Baterai Tinggal 8%',
              description: 'Peringatan baterai lemah muncul: tinggal 8%, perjalanan masih sekitar 15 menit. Apa yang kamu lakukan?',
              options: [
                { id: 'a', label: 'Kunci layar, simpan HP di saku, ikuti patokan yang sudah dicatat, nyalakan HP hanya bila benar-benar perlu', delta: 20, hint: 'Menjaga sisa baterai sebagai cadangan darurat sampai kamu tiba' },
                { id: 'b', label: 'Nyalakan senter dan putar musik agar perjalanan terasa lebih tenang', delta: 5, hint: 'Nyaman sesaat, tetapi senter dan musik menghabiskan sisa baterai dengan cepat' },
                { id: 'c', label: 'Panik dan terus-menerus membuka tutup peta online untuk memastikan posisi', delta: 0, hint: 'Kepanikan dan refresh berulang menghabiskan baterai tanpa menambah kepastian' },
              ],
            },
          ],
        },
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Ajarkan Satu Skill',
      description:
        'Pilih SATU keahlian nyata yang benar-benar kamu kuasai — misalnya fitur HP yang jarang orang tahu, cara memasak sederhana, cara memperbaiki hal kecil, shortcut kerja, atau teknik praktis apa pun yang aman. Rekam video sekitar 60 detik yang MENGAJARKANNYA sambil MEMPRAKTIKKANNYA langsung: tunjukkan masalahnya, tunjukkan caranya langkah per langkah, dan tunjukkan hasilnya. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya sambil menjelaskan dengan suaramu.',
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-19',
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
            'Rekam video maksimal 60 detik yang mengajarkan SATU skill nyata milikmu DENGAN demonstrasi langsung (bukan sekadar bercerita):\n1. Masalah/Tujuan — sebutkan skill apa yang kamu ajarkan dan untuk apa gunanya\n2. Cara — jelaskan langkahnya satu per satu dengan bahasa sederhana\n3. Demonstrasi — PRAKTIKKAN langsung di depan kamera selagi menjelaskan\n4. Hasil — tunjukkan hasil akhirnya agar penonton yakin caranya berhasil\n\nWajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya. Pilih hal yang aman (tanpa api, bahan kimia, listrik, atau benda berbahaya). Jangan menyebut data pribadi sensitif. Bicaralah senatural mungkin!',
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
  console.log('[4] Verifying 2026-09-18 + 2026-09-19 state...');
  for (const [date, titles] of [['2026-09-18', NEW_18], ['2026-09-19', NEW_19]]) {
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
    body: JSON.stringify([{ key: 'schema_version', value: '088_tasks_replace_2026_09_18_19', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 088 Applied Successfully ===');
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
