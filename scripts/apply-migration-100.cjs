require('dotenv').config();

// Applies Migration 100: Fresh tasks for 12-16 October 2026 (15 tasks).
// 1. Pre-audit: expects 0 rows on each of the 5 dates. ABORTS if any row
//    exists with submissions/ledger, or any unexpected row at all.
// 2. Idempotent pre-clean: deletes ONLY the 15 exact new titles on their
//    exact dates (previous partial run of THIS migration).
// 3. Inserts the 15 tasks (payload mirrors
//    supabase/migrations/100_tasks_2026_10_12_16.sql exactly).
// 4. Verifies exactly 3 active tasks per date, step 1-3, exact titles.
// 5. Bumps odyssey_schema_version to 100_tasks_2026_10_12_16.
// Does NOT touch any other date, task, or config. No economy change.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const DATES = ['2026-10-12', '2026-10-13', '2026-10-14', '2026-10-15', '2026-10-16'];
const NEW_TITLES = {
  '2026-10-12': [
    'Rekam Satu Detail Kecil lewat Foto',
    'Tentukan Langkah Paling Tepat',
    'Uraikan Satu Hal dengan Bahasamu Sendiri',
  ],
  '2026-10-13': [
    'Peragakan Cara Memakai Satu Benda',
    'Temukan Fungsi Lain Satu Barang',
    'Susun Lembar Panduan untuk Teman Baru',
  ],
  '2026-10-14': [
    'Timbang Dua Opsi lalu Pilih Satu',
    'Atasi Keterbatasan dengan yang Ada',
    'Tulis Lembar Catatan yang Bisa Dibuka Lagi',
  ],
  '2026-10-15': [
    'Ajarkan Satu Keterampilan Sederhana',
    'Abadikan Satu Perubahan Kecil Hari Ini',
    'Bagikan Satu Pengalaman dan Pelajaran Praktis',
  ],
  '2026-10-16': [
    'Tuangkan Satu Ide Jadi Rencana Singkat',
    'Pilih Tanggapan Terbaik Sehari-hari',
    'Pamerkan Satu Hasil Buatan atau Perbaikanmu',
  ],
};

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

function scenario12() {
  const ev = (id, title, description, opts) => ({ id, title, description, options: opts });
  const o = (id, label, delta, hint) => ({ id, label, delta, hint });
  return {
    game: 'DECISION_PRIORITY',
    target_score: 100,
    scenario: {
      currency: 'POINTS',
      initial_balance: 0,
      events: [
        ev('ev_1', 'Situasi 1 - Lampu Kamar Mati',
          'Malam hari lampu kamarmu tiba-tiba mati. Kamu perlu cahaya untuk belajar. Apa langkah paling tepat?',
          [
            o('a', 'Cek dulu saklar dan bohlamnya, coba bohlam cadangan atau pindahkan ke ruangan terang', 20, 'Mengecek sumber masalah dulu sebelum bertindak menghemat waktu dan biaya'),
            o('b', 'Langsung beli lampu baru malam itu juga tanpa mengecek apa pun', 5, 'Membeli pengganti sebelum tahu rusaknya di mana berisiko salah beli'),
            o('c', 'Biarkan gelap dan berharap besok menyala sendiri', 0, 'Mendiamkan masalah tanpa langkah apa pun tidak mengubah keadaan'),
          ]),
        ev('ev_2', 'Situasi 2 - Galon Habis Saat Haus',
          'Kamu sangat haus tetapi galon air habis dan toko sudah tutup. Apa langkah paling tepat?',
          [
            o('a', 'Minta sedikit air ke tetangga atau gunakan persediaan darurat, isi galon besok pagi', 20, 'Kebutuhan mendesak diatasi dengan sumber terdekat, sisanya dijadwalkan besok'),
            o('b', 'Beli minuman kemasan mahal dalam jumlah banyak lewat ojek online', 5, 'Solusi darurat yang mahal untuk kebutuhan sesaat menguras uang tanpa perlu'),
            o('c', 'Tahan haus sampai besok tanpa melakukan apa pun', 0, 'Menahan kebutuhan dasar tanpa usaha mencari alternatif merugikan dirimu sendiri'),
          ]),
        ev('ev_3', 'Situasi 3 - Kunci Nyangkut',
          'Kunci motormu nyangkut dan sulit diputar. Kamu terburu-buru. Apa langkah paling tepat?',
          [
            o('a', 'Tenang, goyangkan kunci perlahan sambil mengatur posisi setang, beri pelumas bila ada', 20, 'Kunci yang dipaksa berisiko patah; gerakan perlahan dan pelumas menyelesaikan tanpa merusak'),
            o('b', 'Putar paksa sekuat tenaga agar cepat terbuka', 0, 'Memaksa kunci yang nyangkut justru membuatnya patah dan masalahnya berlipat'),
            o('c', 'Congkel rumah kunci dengan obeng atau benda tajam', 5, 'Mencongkel merusak rumah kunci dan biayanya jauh lebih besar daripada bersabar sebentar'),
          ]),
        ev('ev_4', 'Situasi 4 - Paket Datang Saat Sibuk',
          'Kurir mengantar paket saat kedua tanganmu sibuk memasak dan tidak bisa ke depan. Apa langkah paling tepat?',
          [
            o('a', 'Minta kurir menitipkan di tempat aman atau meletakkannya di teras lalu kabari keluarga', 20, 'Solusi titip yang aman menyelamatkan paket tanpa mengorbankan masakanmu'),
            o('b', 'Biarkan paket tergeletak di depan pagar tanpa kepastian', 0, 'Paket tanpa kepastian penyimpanan berisiko hilang atau rusak'),
            o('c', 'Suruh kurir menunggu lama sampai kamu selesai memasak', 5, 'Menahan kurir terlalu lama merugikan orang lain padahal ada solusi titip'),
          ]),
        ev('ev_5', 'Situasi 5 - Teman Meminjam Barang Penting',
          'Teman ingin meminjam barang pentingmu padahal besok kamu juga membutuhkannya. Apa langkah paling tepat?',
          [
            o('a', 'Jelaskan kamu juga membutuhkannya besok, tawarkan jadwal pinjam bergantian atau barang pengganti', 20, 'Jujur soal kebutuhanmu sambil menawarkan alternatif menjaga barang dan pertemanan'),
            o('b', 'Langsung berikan tanpa mencatat kapan kembali', 5, 'Memberi tanpa kesepakatan waktu kembali berisiko kamu sendiri yang kesulitan besok'),
            o('c', 'Tolak mentah-mentah tanpa penjelasan apa pun', 0, 'Penolakan tanpa penjelasan merusak hubungan padahal ada jalan tengah'),
          ]),
        ev('ev_6', 'Situasi 6 - Colokan Penuh',
          'Semua colokan terisi dan HP-mu hampir habis baterai padahal akan dipakai menelepon penting. Apa langkah paling tepat?',
          [
            o('a', 'Cabut perangkat yang sudah penuh dayanya, lalu cas HP-mu untuk keperluan penting itu', 20, 'Memprioritaskan colokan berdasarkan kebutuhan mendesak adalah langkah yang adil dan masuk akal'),
            o('b', 'Tumpuk beberapa steker dengan paksa agar HP-mu ikut tercolok', 0, 'Menumpuk steker secara paksa berbahaya dan bisa merusak perangkat'),
            o('c', 'Biarkan HP habis dan batalkan telepon penting itu', 5, 'Menyerah tanpa mengatur ulang colokan mengorbankan keperluan penting yang sebenarnya bisa diselamatkan'),
          ]),
      ],
    },
  };
}

function scenario14() {
  const ev = (id, title, description, opts) => ({ id, title, description, options: opts });
  const o = (id, label, delta, hint) => ({ id, label, delta, hint });
  return {
    game: 'DECISION_PRIORITY',
    target_score: 100,
    scenario: {
      currency: 'POINTS',
      initial_balance: 0,
      events: [
        ev('ev_1', 'Situasi 1 - Tali Putus',
          'Tali untuk mengikat barang putus dan tidak ada tali cadangan. Kamu tidak boleh membeli. Apa akal terbaikmu?',
          [
            o('a', 'Sambung dengan kain perca atau plastik yang dipilin menjadi tali darurat yang kuat', 20, 'Bahan lentur di sekitar bisa dipilin menjadi pengikat darurat yang cukup kuat'),
            o('b', 'Ikat seadanya dengan sisa tali pendek walau longgar', 5, 'Ikatan longgar yang dipaksakan akan lepas di jalan'),
            o('c', 'Biarkan barang tanpa ikatan dan berharap tidak jatuh', 0, 'Barang tanpa ikatan pasti bermasalah saat dibawa'),
          ]),
        ev('ev_2', 'Situasi 2 - Paku Kurang',
          'Paku untuk menggantung pajangan kurang dan toko jauh. Apa akal terbaikmu?',
          [
            o('a', 'Manfaatkan pengait tempel, paku yang masih menancap, atau sandarkan pajangan dengan aman', 20, 'Titik gantung yang sudah ada atau pengait tempel menyelesaikan tanpa paku baru'),
            o('b', 'Paksakan paku yang bengkok untuk dipakai lagi', 5, 'Paku bengkok yang dipaksakan tidak kuat menahan beban'),
            o('c', 'Tempel pajangan dengan lem kertas biasa', 0, 'Lem kertas tidak dirancang menahan beban pajangan dan pasti jatuh'),
          ]),
        ev('ev_3', 'Situasi 3 - Wadah Kurang',
          'Makanan sisa cukup banyak tetapi wadah tertutup kurang. Apa akal terbaikmu?',
          [
            o('a', 'Gunakan mangkuk lalu tutup rapat dengan piring dan ikat dengan kain atau karet', 20, 'Wadah darurat yang tertutup rapat menjaga makanan sama baiknya'),
            o('b', 'Simpan dalam piring terbuka di kulkas', 5, 'Makanan terbuka menyerap bau dan mengering walau di kulkas'),
            o('c', 'Biarkan di meja tanpa tutup sampai besok', 0, 'Makanan tanpa tutup di suhu ruang cepat basi'),
          ]),
        ev('ev_4', 'Situasi 4 - Cahaya Redup',
          'Lampu belajar redup dan kamu harus membaca malam ini juga. Apa akal terbaikmu?',
          [
            o('a', 'Pindah ke dekat sumber cahaya yang masih kuat atau gunakan lampu HP sebagai penerangan fokus', 20, 'Mendekati cahaya yang ada jauh lebih efektif daripada memaksakan mata dalam gelap'),
            o('b', 'Tetap membaca di cahaya redup sambil menyipitkan mata', 0, 'Membaca dalam cahaya redup melelahkan mata dan tidak efektif'),
            o('c', 'Nyalakan semua lampu rumah walau tidak membantu bacaanmu', 5, 'Menyalakan semua lampu tanpa mengarahkan cahaya ke bacaan membuang listrik tanpa hasil'),
          ]),
        ev('ev_5', 'Situasi 5 - Air Sedikit',
          'Air tinggal sedikit tetapi tumpukan piring kotor banyak. Apa akal terbaikmu?',
          [
            o('a', 'Rendam yang berlemak dulu agar lunak, lalu cuci hemat dengan spons dan bilas seperlunya', 20, 'Merendam dulu membuat kotoran lunak sehingga sedikit air pun cukup'),
            o('b', 'Guyur semua piring dengan air seperti biasa sampai air habis', 0, 'Mengguyur boros saat air sedikit membuat sebagian piring tidak kebagian'),
            o('c', 'Tumpuk saja piring kotornya dan tunggu air banyak besok', 5, 'Menumpuk tanpa merendam membuat lemak mengeras dan besok jauh lebih sulit dicuci'),
          ]),
        ev('ev_6', 'Situasi 6 - Kantong Sobek',
          'Kantong belanjaan sobek di jalan dan tidak ada kantong cadangan. Apa akal terbaikmu?',
          [
            o('a', 'Rangkap bagian sobek dengan lipatan, bagi beban ke dua tangan atau minta tolong bawakan sebagian', 20, 'Mendistribusikan beban dan melipat bagian sobek menyelamatkan belanjaan sampai rumah'),
            o('b', 'Tetap jinjing seperti biasa walau sobeknya makin lebar', 0, 'Beban penuh pada kantong sobek menjamin isinya tumpah di jalan'),
            o('c', 'Tinggalkan sebagian belanjaan di jalan agar kantong ringan', 5, 'Meninggalkan belanjaan merugikan; membagi beban adalah jalan yang lebih baik'),
          ]),
      ],
    },
  };
}

function scenario16() {
  const ev = (id, title, description, opts) => ({ id, title, description, options: opts });
  const o = (id, label, delta, hint) => ({ id, label, delta, hint });
  return {
    game: 'DECISION_PRIORITY',
    target_score: 100,
    scenario: {
      currency: 'POINTS',
      initial_balance: 0,
      events: [
        ev('ev_1', 'Situasi 1 - Kembalian Lebih',
          'Kasir mengembalikan uang lebih dari seharusnya dan tidak sadar. Antrean di belakangmu panjang. Apa tanggapan terbaikmu?',
          [
            o('a', 'Segera kembalikan kelebihannya dengan sopan walau antrean menunggu sebentar', 20, 'Kelebihan yang dikembalikan menjaga kejujuran; antrean sebentar bisa dimaklumi'),
            o('b', 'Ambil diam-diam dan anggap rezeki', 0, 'Mengambil kelebihan yang bukan hakmu merugikan kasir yang harus mengganti'),
            o('c', 'Belanjakan lagi kelebihannya di toko yang sama agar tidak kentara', 5, 'Menghabiskan kelebihan tetap berarti mengambil yang bukan hakmu'),
          ]),
        ev('ev_2', 'Situasi 2 - Salah Kirim Pesan',
          'Kamu salah mengirim pesan pribadi ke grup keluarga besar. Isinya tidak pantas dibaca umum. Apa tanggapan terbaikmu?',
          [
            o('a', 'Segera akui kesalahan, minta maaf singkat, lalu kirim koreksi yang benar', 20, 'Mengakui dan mengoreksi dengan cepat memulihkan keadaan lebih baik daripada menghilang'),
            o('b', 'Hapus pesan diam-diam dan berharap tidak ada yang sempat membaca', 5, 'Menghapus tanpa penjelasan meninggalkan kebingungan bagi yang sudah terlanjur membaca'),
            o('c', 'Biarkan saja dan pura-pura tidak terjadi apa-apa', 0, 'Membiarkan tanpa tanggapan membuat kesalahpahaman berkembang'),
          ]),
        ev('ev_3', 'Situasi 3 - Teman Terlambat',
          'Teman janjian denganmu terlambat 30 menit tanpa kabar. Kamu sudah menunggu. Apa tanggapan terbaikmu?',
          [
            o('a', 'Hubungi dengan tenang, tanyakan kabarnya, sepakati batas tunggu yang wajar', 20, 'Menanyakan kabar dengan tenang memberi kesempatan penjelasan sekaligus batas yang jelas'),
            o('b', 'Marahi langsung lewat pesan panjang penuh emosi', 0, 'Luapan emosi merusak pertemuan yang sebenarnya masih bisa diselamatkan'),
            o('c', 'Pergi diam-diam tanpa kabar dan biarkan dia mencari', 5, 'Pergi tanpa kabar membalas kekecewaan dengan kekecewaan baru'),
          ]),
        ev('ev_4', 'Situasi 4 - Menumpahkan Minuman',
          'Kamu tidak sengaja menumpahkan minuman di tempat umum dan mengotori lantai. Orang melihat. Apa tanggapan terbaikmu?',
          [
            o('a', 'Minta maaf singkat lalu segera lap dan bereskan sampai bersih', 20, 'Bertanggung jawab membersihkan langsung menunjukkan kedewasaan'),
            o('b', 'Pura-pura tidak melihat dan pergi menjauh', 0, 'Menghindar meninggalkan kotoran untuk orang lain dan membahayakan yang lewat'),
            o('c', 'Salahkan lantai yang licin atau gelas yang terlalu penuh', 5, 'Mencari kambing hitam tidak membersihkan tumpahan yang tetap harus dibereskan'),
          ]),
        ev('ev_5', 'Situasi 5 - Diminta Tolong Saat Sibuk',
          'Keluargamu meminta tolong saat kamu sedang mengerjakan hal penting yang mendesak. Apa tanggapan terbaikmu?',
          [
            o('a', 'Jelaskan kamu butuh beberapa menit menyelesaikan ini, lalu bantu setelahnya', 20, 'Memberi kepastian waktu membantu jauh lebih baik daripada menolak atau mengorbankan pekerjaan'),
            o('b', 'Tinggalkan pekerjaan pentingmu seketika walau berantakan', 5, 'Meninggalkan pekerjaan mendesak tanpa kepastian membuat dua-duanya berantakan'),
            o('c', 'Tolak dengan kasar agar tidak diganggu lagi', 0, 'Penolakan kasar melukai orang yang sebenarnya hanya butuh bantuan kecil'),
          ]),
        ev('ev_6', 'Situasi 6 - Parkir Sempit',
          'Motor tetangga menghalangi jalan keluar motormu di parkiran sempit. Kamu terburu-buru. Apa tanggapan terbaikmu?',
          [
            o('a', 'Geser perlahan dengan hati-hati lalu kabari tetangga agar lain kali parkir lebih rapi', 20, 'Menggeser hati-hati plus kabar baik menyelesaikan masalah tanpa merusak dan mencegah terulang'),
            o('b', 'Dorong paksa sampai motormu bisa keluar walau motor tetangga tergores', 0, 'Mendorong paksa merusak barang orang lain demi kepentinganmu sendiri'),
            o('c', 'Klakson dan marah-marah sampai tetangga keluar', 5, 'Keributan tidak mengeluarkan motormu lebih cepat daripada menggesernya baik-baik'),
          ]),
      ],
    },
  };
}

const DOC_EXT = ['pdf', 'docx', 'txt', 'jpg', 'jpeg', 'png'];

async function run() {
  console.log('=== Applying Migration 100: Fresh tasks for 2026-10-12..16 ===');

  console.log('[0] Economy fingerprint (before)...');
  const cfgBefore = await get('odyssey_system_config?select=key,value');
  const keep = ['coin_conversion_rate', 'payout_target_rupiah', 'max_payout_coins', 'default_monthly_earning_cap', 'default_monthly_coin_target', 'default_minimum_withdrawal_coins', 'default_payout_frequency', 'timezone'];
  const fp = {};
  for (const row of cfgBefore) if (keep.includes(row.key)) fp[row.key] = row.value;
  console.log('   BEFORE: ' + JSON.stringify(fp));

  console.log('[1] Pre-apply audit of target dates...');
  for (const d of DATES) {
    const rows = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title,step_order,is_active`);
    console.log(`   rows on ${d}: ` + JSON.stringify(rows));
    if (rows.length) {
      for (const t of rows) {
        const s = await get(`odyssey_task_submissions?task_id=eq.${t.id}&select=id,status`);
        if (s.length) throw new Error(`ABORT: task ${t.id} carries ${s.length} submissions — deactivate instead of delete.`);
      }
      throw new Error(`ABORT: unexpected rows on ${d} without submissions — audit manually before proceeding.`);
    }
  }

  console.log('[2] Idempotent pre-clean (exact new titles, exact dates)...');
  const enc = (s) => encodeURIComponent(`"${s}"`);
  let cleaned = 0;
  for (const d of DATES) {
    const pre = await get(`odyssey_tasks?active_date=eq.${d}&select=id,title&title=in.(${NEW_TITLES[d].map(enc).join(',')})`);
    if (pre.length) {
      const ids = pre.map((t) => t.id).join(',');
      const del = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
      if (!del.ok) throw new Error(`Pre-clean DELETE failed: ${del.status} ${await del.text()}`);
      cleaned += pre.length;
      console.log(`   removed ${pre.length} partial-run row(s) on ${d}.`);
    }
  }
  if (!cleaned) console.log('   no partial-run rows found.');

  console.log('[3] Inserting 15 tasks...');
  const FAM = 'demo-crew-1';
  const tasks = [
    {
      family_id: FAM, title: 'Rekam Satu Detail Kecil lewat Foto',
      description: 'Benda yang setiap hari kamu pegang menyimpan detail kecil yang jarang disadari! Pilih SATU benda yang akrab bagimu, foto SATU detail kecilnya dari dekat (tekstur, pola aus, sambungan, atau tanda pemakaian), lalu jelaskan detail itu dan kenapa sering terlewat.',
      task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-10-12',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { max_files: 1, camera_only: false, instruction: 'Pilih SATU benda sehari-hari yang akrab bagimu (contoh: tas, sepatu, botol minum, alat dapur, atau perabot kecil), lalu ambil atau unggah tepat 1 foto CLOSE-UP satu detail kecil pada benda itu (boleh kamera langsung atau galeri, maksimal 1 file). Wajib tulis catatan dengan 3 bagian: (a) Detail apa - jelaskan detail kecil yang kamu foto dan letaknya di bagian mana, (b) Pada benda apa - sebutkan bendanya dan seberapa sering kamu memakainya, (c) Kenapa sering terlewat - jelaskan kenapa detail itu jarang disadari orang. Ini tugas pengamatan, bukan perbaikan. Admin akan menolak submission tanpa foto close-up detail, tanpa 3 bagian catatan, atau foto yang tidak menunjukkan satu detail yang spesifik.' },
    },
    {
      family_id: FAM, title: 'Tentukan Langkah Paling Tepat',
      description: 'Hari biasa penuh momen yang menuntut satu langkah yang tepat! Hadapi 6 situasi rumah tangga yang umum - lampu mati, galon habis, kunci nyangkut, paket datang, pinjaman barang, dan colokan penuh. Pada setiap situasi, tentukan SATU langkah paling tepat yang harus dilakukan lebih dulu!',
      task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-10-12',
      reward_coins: 25, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: scenario12(),
    },
    {
      family_id: FAM, title: 'Uraikan Satu Hal dengan Bahasamu Sendiri',
      description: 'Bisa menjelaskan dengan kata-kata sendiri tandanya benar-benar paham! Pilih SATU hal sederhana yang kamu kuasai, lalu uraikan secara tertulis dengan bahasamu sendiri lengkap dengan satu perumpamaan dan satu contoh agar teman pemula bisa memahaminya.',
      task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 3, active_date: '2026-10-12',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { prompt: 'Tulis jawabanmu dengan format 3 bagian:\n1) Hal yang kamu jelaskan - sebutkan SATU hal sederhana yang kamu kuasai (contoh: kenapa nasi basi, cara kerja payung, kenapa kaca berembun, atau cara mengatur sesuatu). Jelaskan dengan kata-katamu sendiri, bukan menyalin dari internet:\n2) Perumpamaan - buat SATU perumpamaan sederhana agar mudah dibayangkan (contoh: seperti ..., ibarat ...):\n3) Contoh nyata - berikan SATU contoh nyata dari pengalamanmu yang menunjukkan hal itu benar:\n\nDitulis untuk teman pemula yang belum tahu apa-apa. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, tanpa perumpamaan, atau penjelasan yang disalin dan sulit dipahami pemula.', minimum_characters: 150, maximum_characters: 800 },
    },
    {
      family_id: FAM, title: 'Peragakan Cara Memakai Satu Benda',
      description: 'Benda sehari-hari baru berguna kalau dipakai dengan benar! Pilih SATU benda di rumahmu, lalu rekam video maksimal 60 detik yang MEMPERAGAKAN cara memakai benda itu dengan benar dan aman. Orang lain harus bisa menirunya setelah menonton!',
      task_type: 'VIDEO', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-10-13',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { recording: { enabled: true, max_duration_seconds: 60, camera_facing: 'user', instruction: 'Rekam video maksimal 60 detik dengan struktur:\n1. Tunjukkan SATU benda sehari-hari yang spesifik (contoh: payung lipat, setrika, blender, gembok, dispenser, atau kipas angin) - fokus pada satu benda saja\n2. PERAGAKAN langsung di depan kamera cara memakai benda itu dengan benar dan aman (wajib ada peragaan nyata, bukan hanya bicara) - tunjukkan bagian-bagian benda yang relevan\n3. Sebutkan SATU kesalahan umum yang sering dilakukan orang saat memakai benda itu dan cara menghindarinya\n\nCatatan: ini peragaan pemakaian dasar yang benar, bukan trik dan bukan cara yang lebih praktis. Admin akan menolak video tanpa peragaan nyata, tanpa satu benda yang spesifik, tanpa kesalahan umum yang dijelaskan, durasi melebihi 60 detik, atau cara pakai yang tidak aman.' } },
    },
    {
      family_id: FAM, title: 'Temukan Fungsi Lain Satu Barang',
      description: 'Banyak barang bisa dipakai untuk hal lain selain fungsi utamanya! Pilih SATU barang di rumahmu, sebutkan fungsi utamanya, lalu temukan SATU fungsi lain yang masuk akal lengkap dengan cara melakukannya dan batasannya.',
      task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-10-13',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { prompt: 'Tulis jawabanmu dengan format 3 bagian:\n1) Barang dan fungsi utama - sebutkan SATU barang dan fungsi utamanya (contoh: handuk untuk mengeringkan badan, kardus untuk menyimpan barang, atau jepit jemuran untuk menjepit pakaian):\n2) Fungsi lain - jelaskan SATU fungsi lain yang masuk akal untuk barang itu dan cara melakukannya langkah demi langkah:\n3) Batasan - jelaskan kapan fungsi lain itu TIDAK cocok dipakai dan kenapa:\n\nFokus pada satu barang dan satu fungsi lain saja. Fungsi lain harus aman dan tidak merusak barang. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, atau fungsi lain yang berbahaya atau tidak masuk akal.', minimum_characters: 150, maximum_characters: 800 },
    },
    {
      family_id: FAM, title: 'Susun Lembar Panduan untuk Teman Baru',
      description: 'Bayangkan temanmu baru pertama kali melakukan sesuatu dan tidak tahu apa-apa! Pilih SATU aktivitas sehari-hari yang spesifik, lalu buat SATU dokumen lembar panduan satu halaman untuk pemula lengkap dengan istilah yang dijelaskan dan kesalahan umum yang harus dihindari. Ini panduan bacaan, bukan daftar centang.',
      task_type: 'DOCUMENT_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 3, active_date: '2026-10-13',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { allowed_extensions: DOC_EXT, accepted_extensions: DOC_EXT, max_file_size_mb: 4, instruction: 'Buat SATU dokumen lembar panduan untuk pemula dengan aturan: 1) Maksimal sekitar 1 halaman - singkat dan langsung ke intinya. 2) Fokus pada SATU aktivitas sehari-hari yang spesifik (contoh: cara mencuci sepatu, cara menyimpan beras, cara merawat tanaman, atau cara mengatur dompet). 3) Ditulis untuk orang yang belum pernah melakukannya: jelaskan istilah yang mungkin asing dan sebutkan minimal 2 kesalahan umum pemula beserta cara menghindarinya. 4) Ditulis sebagai panduan bacaan yang runut - BUKAN berupa daftar centang (dokumen berisi kotak centang seperti [ ] ... akan ditolak karena itu jenis tugas yang berbeda). 5) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: aktivitas apa dan untuk pemula seperti apa panduan itu. Admin akan menolak dokumen berupa daftar centang, dokumen tanpa penjelasan istilah, dokumen tanpa kesalahan umum, atau topik yang terlalu umum.' },
    },
    {
      family_id: FAM, title: 'Timbang Dua Opsi lalu Pilih Satu',
      description: 'Setiap hari kita dihadapkan pada dua pilihan! Pilih SATU keputusan nyata yang sedang kamu hadapi dengan dua opsi saja, bandingkan keduanya dengan dua kriteria yang jelas, lalu tentukan pilihanmu beserta alasannya.',
      task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-10-14',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { prompt: 'Tulis jawabanmu dengan format 3 bagian:\n1) Keputusan dan dua opsi - jelaskan SATU keputusan nyata sehari-hari dengan tepat DUA opsi (contoh: memperbaiki sendiri atau minta bantuan, berangkat pagi atau siang, menyimpan di lemari atau di kardus). Bukan soal belanja barang:\n2) Perbandingan dua kriteria - bandingkan kedua opsi dengan DUA kriteria yang jelas (contoh: biaya dan waktu, usaha dan hasil, atau cocok dan risiko). Jelaskan kelebihan dan kekurangan tiap opsi:\n3) Pilihan dan alasan - tentukan SATU pilihanmu dan jelaskan kenapa opsi itu menang berdasarkan perbandingan di atas:\n\nFokus pada satu keputusan dengan dua opsi saja. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, membahas lebih dari dua opsi, atau pilihan tanpa alasan dari perbandingan.', minimum_characters: 150, maximum_characters: 800 },
    },
    {
      family_id: FAM, title: 'Atasi Keterbatasan dengan yang Ada',
      description: 'Tidak selalu bisa membeli yang baru - kadang harus beres dengan yang ada! Hadapi 6 situasi serba terbatas di mana kamu tidak boleh membeli apa pun: tali putus, paku kurang, wadah kurang, cahaya redup, air sedikit, dan kantong sobek. Pada setiap situasi, pilih akal terbaik dengan bahan yang ada!',
      task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-10-14',
      reward_coins: 25, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: scenario14(),
    },
    {
      family_id: FAM, title: 'Tulis Lembar Catatan yang Bisa Dibuka Lagi',
      description: 'Catatan yang baik menyelamatkanmu di kemudian hari! Pilih SATU hal yang sering kamu lupa atau perlu diingat (langkah kerja, rumus sederhana, daftar istilah, atau template), lalu buat SATU dokumen lembar catatan ringkas maksimal satu halaman yang bisa kamu buka lagi kapan pun. Ini catatan rujukan, bukan daftar centang dan bukan karangan panjang.',
      task_type: 'DOCUMENT_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 3, active_date: '2026-10-14',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { allowed_extensions: DOC_EXT, accepted_extensions: DOC_EXT, max_file_size_mb: 4, instruction: 'Buat SATU dokumen lembar catatan rujukan dengan aturan: 1) Maksimal sekitar 1 halaman - ringkas agar enak dibuka lagi. 2) Berisi SATU topik rujukan yang spesifik (contoh: langkah mencuci sepatu, daftar istilah HP, template pesan izin, atau rumus belanja hemat). 3) Ditulis sebagai catatan rujukan yang runut (penjelasan singkat plus poin-poin ingatan) sehingga kamu sendiri bisa memakainya lagi bulan depan tanpa penjelasan tambahan - BUKAN berupa daftar centang saja (dokumen berisi kotak centang seperti [ ] ... akan ditolak) dan BUKAN karangan panjang berlembar-lembar. 4) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: topik catatanmu dan kapan kamu akan membukanya lagi. Admin akan menolak dokumen berupa daftar centang saja, karangan panjang tanpa struktur rujukan, atau topik yang terlalu umum.' },
    },
    {
      family_id: FAM, title: 'Ajarkan Satu Keterampilan Sederhana',
      description: 'Punya keterampilan kecil yang berguna? Saatnya mengajarkannya! Pilih SATU keterampilan sederhana yang kamu kuasai (misalnya melipat baju, mengikat tali, mencatat tugas, atau menyusun barang), lalu rekam video maksimal 60 detik yang MENGAJARKANNYA langkah demi langkah. Pemula harus bisa mengikutinya!',
      task_type: 'VIDEO', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-10-15',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { recording: { enabled: true, max_duration_seconds: 60, camera_facing: 'user', instruction: 'Rekam video maksimal 60 detik dengan struktur:\n1. Sebutkan SATU keterampilan sederhana yang spesifik dan benar-benar kamu kuasai (contoh: melipat baju, mengikat tali sepatu, mencatat tugas harian, atau menyusun barang) - fokus pada satu keterampilan untuk pemula\n2. AJARKAN langsung di depan kamera langkah demi langkah (wajib ada peragaan nyata, bukan hanya bicara) - tunjukkan objek atau alat yang dipakai jika relevan\n3. Sebutkan SATU kesalahan umum pemula dan cara memperbaikinya\n\nCatatan: ini pengajaran keterampilan dasar untuk pemula, bukan trik dan bukan sekadar cara memakai benda. Admin akan menolak video tanpa peragaan mengajar, tanpa satu keterampilan yang spesifik, tanpa kesalahan umum pemula, durasi melebihi 60 detik, atau langkah yang tidak bisa diikuti pemula.' } },
    },
    {
      family_id: FAM, title: 'Abadikan Satu Perubahan Kecil Hari Ini',
      description: 'Sekitarmu berubah sedikit setiap hari - sadarkah kamu? Temukan SATU perubahan kecil yang terjadi hari ini di sekitarmu (tanaman bertambah daun, barang yang diperbaiki, sudut yang ditata ulang, atau hasil kerja seseorang), foto, lalu ceritakan perubahan itu.',
      task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-10-15',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { max_files: 1, camera_only: false, instruction: 'Temukan SATU perubahan kecil yang nyata terjadi hari ini di sekitarmu (contoh: tanaman yang bertunas, barang yang baru diperbaiki, sudut ruangan yang ditata ulang, atau hasil kerja seseorang hari ini), lalu ambil atau unggah tepat 1 foto perubahan itu (boleh kamera langsung atau galeri, maksimal 1 file). Wajib tulis catatan dengan 3 bagian: (a) Apa yang berubah - jelaskan perubahannya dan di mana, (b) Sejak kapan - jelaskan kapan kamu menyadari perubahan itu, (c) Kenapa berarti - jelaskan kenapa perubahan kecil itu penting atau menyenangkan. Ini tugas pengamatan perubahan, kamu tidak wajib menjadi pelakunya. Admin akan menolak submission tanpa foto perubahan, tanpa 3 bagian catatan, atau perubahan yang tidak terlihat nyata.' },
    },
    {
      family_id: FAM, title: 'Bagikan Satu Pengalaman dan Pelajaran Praktis',
      description: 'Pengalaman jadi berharga saat pelajarannya bisa dipakai lagi! Tulis SATU pengalaman nyatamu, pelajaran praktis yang kamu ambil, dan SATU cara kamu akan menerapkan pelajaran itu ke depan. Ditulis dari pengalaman sendiri, bukan karangan umum.',
      task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 3, active_date: '2026-10-15',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { prompt: 'Tulis jawabanmu dengan format 3 bagian:\n1) Pengalaman - ceritakan SATU pengalaman nyata yang kamu alami sendiri (contoh: memperbaiki sesuatu, mengatur sesuatu, menghadapi situasi sulit, atau mencoba hal baru). Ceritakan kejadiannya secara konkret:\n2) Pelajaran praktis - jelaskan SATU pelajaran praktis yang kamu ambil dari pengalaman itu:\n3) Penerapan ke depan - tulis SATU cara konkret kamu akan menerapkan pelajaran itu dalam 7 hari ke depan:\n\nDitulis sebagai teks, bukan video. Fokus pada satu pengalaman dan satu pelajaran saja. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, pengalaman yang tidak konkret, atau penerapan yang tidak bisa dilakukan dalam 7 hari.', minimum_characters: 150, maximum_characters: 800 },
    },
    {
      family_id: FAM, title: 'Tuangkan Satu Ide Jadi Rencana Singkat',
      description: 'Punya ide bagus? Jangan biarkan hanya di kepala! Pilih SATU ide sederhana yang bisa membuat hidup sehari-hari lebih baik, lalu buat SATU dokumen rencana singkat satu halaman: masalahnya apa, idemu apa, dan langkah pertama yang bisa dilakukan minggu ini. Rencana harus bisa langsung dijalankan.',
      task_type: 'DOCUMENT_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-10-16',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { allowed_extensions: DOC_EXT, accepted_extensions: DOC_EXT, max_file_size_mb: 4, instruction: 'Buat SATU dokumen rencana singkat dengan aturan: 1) Maksimal sekitar 1 halaman - singkat dan langsung ke intinya. 2) Berisi SATU ide sederhana untuk kehidupan sehari-hari (contoh: cara mengatur sudut ruangan, cara menghemat sesuatu, cara memperbaiki kebiasaan, atau cara memanfaatkan barang yang ada). 3) Struktur wajib: masalah yang ingin diatasi, ide dan cara kerjanya, serta SATU langkah pertama yang konkret dan bisa dilakukan dalam 7 hari. 4) BUKAN berupa daftar centang saja (dokumen berisi kotak centang seperti [ ] ... akan ditolak) dan BUKAN karangan tanpa langkah yang bisa dijalankan. 5) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: ide apa dan untuk siapa rencana itu. Admin akan menolak dokumen tanpa langkah pertama yang konkret, dokumen berupa daftar centang saja, atau ide yang tidak bisa dijalankan.' },
    },
    {
      family_id: FAM, title: 'Pilih Tanggapan Terbaik Sehari-hari',
      description: 'Cara kita menanggapi hal kecil menunjukkan kedewasaan! Hadapi 6 situasi sosial sehari-hari - kembalian lebih, salah kirim pesan, teman terlambat, tumpahan di tempat umum, diminta tolong saat sibuk, dan parkir sempit. Pada setiap situasi, pilih tanggapan yang paling baik!',
      task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-10-16',
      reward_coins: 25, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: scenario16(),
    },
    {
      family_id: FAM, title: 'Pamerkan Satu Hasil Buatan atau Perbaikanmu',
      description: 'Bangga dengan sesuatu yang kamu buat atau perbaiki sendiri? Saatnya memamerkannya! Foto SATU hasil buatan tanganmu atau barang yang kamu perbaiki sendiri, lalu ceritakan cara singkatnya dan apa gunanya sekarang.',
      task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 3, active_date: '2026-10-16',
      reward_coins: 20, reward_xp: 100, target_scope: 'ALL', is_active: true,
      config: { max_files: 1, camera_only: false, instruction: 'Pilih SATU hasil karya tanganmu sendiri atau SATU barang yang kamu perbaiki sendiri (contoh: masakan, kerajinan, tulisan, susunan, atau barang yang tadinya rusak lalu kamu betulkan), lalu ambil atau unggah tepat 1 foto hasilnya (boleh kamera langsung atau galeri, maksimal 1 file). Wajib tulis catatan dengan 3 bagian: (a) Hasil apa - sebutkan apa yang kamu buat atau perbaiki, (b) Cara singkat - jelaskan cara singkat pembuatannya atau apa yang kamu perbaiki, (c) Guna sekarang - jelaskan apa gunanya hasil itu sekarang. Karya atau perbaikan harus hasil tanganmu sendiri. Admin akan menolak submission tanpa foto hasil, tanpa 3 bagian catatan, atau hasil yang jelas bukan buatan atau perbaikanmu sendiri.' },
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

  console.log('[4] Verifying target dates...');
  for (const d of DATES) {
    const after = await get(
      `odyssey_tasks?active_date=eq.${d}&select=id,title,task_type,evaluation_type,step_order,reward_coins,reward_xp,target_scope,family_id,is_active,config&order=step_order.asc`
    );
    const active = after.filter((t) => t.is_active);
    console.log(`   ${d}: total rows=${after.length}, active=${active.length}`);
    active.forEach((t) => console.log(`   - [ID ${t.id}] #${t.step_order} "${t.title}" ${t.task_type}/${t.evaluation_type} ${t.reward_coins}c/${t.reward_xp}xp scope=${t.target_scope} fam=${t.family_id}`));
    if (active.length !== 3) throw new Error(`Expected exactly 3 active tasks on ${d}, found ${active.length}`);
    if (JSON.stringify(active.map((t) => t.title)) !== JSON.stringify(NEW_TITLES[d])) {
      throw new Error('Title/step mismatch on ' + d + ': ' + JSON.stringify(active.map((t) => t.title)));
    }
  }

  console.log('[5] Economy fingerprint (after)...');
  const cfgAfter = await get('odyssey_system_config?select=key,value');
  const fpAfter = {};
  for (const row of cfgAfter) if (keep.includes(row.key)) fpAfter[row.key] = row.value;
  console.log('   AFTER:  ' + JSON.stringify(fpAfter));
  if (JSON.stringify(fp) !== JSON.stringify(fpAfter)) throw new Error('ECONOMY FINGERPRINT CHANGED — investigate before proceeding.');

  console.log('[6] Updating odyssey_schema_version...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_schema_version`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '100_tasks_2026_10_12_16', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 100 Applied Successfully ===');
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
