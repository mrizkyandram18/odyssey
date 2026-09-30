require('dotenv').config();

// Applies Migration 098: Fresh tasks for 1 + 2 October 2026.
// 1. Pre-audit: expects 0 rows on 2026-10-01 / 2026-10-02. ABORTS if any
//    row exists with submissions/ledger (deactivate instead). Also ABORTS
//    on any unexpected row on those dates.
// 2. Idempotent pre-clean: deletes ONLY the 6 exact new titles on their
//    exact dates (previous partial run of THIS migration).
// 3. Inserts the 6 tasks (config mirrors
//    supabase/migrations/098_tasks_replace_2026_10_01_02.sql exactly).
// 4. Bumps odyssey_schema_version to 098_tasks_replace_2026_10_01_02.
// Does NOT touch any other date, task, or config. No economy change.

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

const NEW_01 = [
  'Kalau Mau Mulai Lagi, Pilih yang Mana?',
  'Sebelum & Sesudah Awal Bulan',
  'Satu Hal yang Mau Kamu Benerin Bulan Ini',
];
const NEW_02 = [
  'Tunjukkan Satu Trik yang Sering Kamu Pakai',
  'Sebelum Dipakai, Cek Dulu',
  'Bikin Checklist yang Bisa Langsung Dipakai',
];

async function get(path) {
  const r = await fetch(`${SUPABASE_URL}/rest/v1/${path}`, { headers });
  if (!r.ok) throw new Error(`GET ${path}: ${r.status} ${await r.text()}`);
  return r.json();
}

function octStartScenario() {
  const ev = (id, title, description, opts) => ({ id, title, description, options: opts });
  const o = (id, label, delta, hint) => ({ id, label, delta, hint });
  return {
    game: 'DECISION_PRIORITY',
    target_score: 100,
    scenario: {
      currency: 'POINTS',
      initial_balance: 0,
      events: [
        ev('ev_1', 'Situasi 1 - Uang Mulai Menipis',
          'Baru tanggal 1, kamu sadar sisa uang tinggal sedikit sementara kebutuhan sebulan masih panjang. Apa langkah pertamamu?',
          [
            o('a', 'Catat sisa uang dan susun anggaran prioritas: kebutuhan wajib dulu, keinginan ditunda', 20, 'Awal bulan yang baik dimulai dari tahu persis posisi keuanganmu sebelum uang habis tanpa jejak'),
            o('b', 'Langsung berhemat tanpa mencatat, kurangi saja semua pengeluaran secara merata', 10, 'Berhemat itu bagus, tetapi tanpa catatan kamu tidak tahu pos mana yang sebenarnya bocor'),
            o('c', 'Belanja dulu mumpung awal bulan, urusan kehabisan dipikir nanti saja', 0, 'Mengulang pola bulan lalu yang membuat uang menipis justru menjamin masalah yang sama terulang'),
          ]),
        ev('ev_2', 'Situasi 2 - Jadwal Berantakan',
          'Jadwal minggu pertamamu bertabrakan di mana-mana: janji, pekerjaan, dan urusan keluarga saling tumpuk. Apa yang kamu lakukan?',
          [
            o('a', 'Tulis semua jadwal di satu tempat, tentukan prioritas, lalu atur ulang yang bisa digeser', 20, 'Jadwal yang terlihat semua dalam satu daftar bisa diatur; yang hanya di kepala pasti ada yang terlewat'),
            o('b', 'Kerjakan yang paling mendesak saja hari ini, sisanya dipikir besok', 5, 'Reaktif tanpa peta jadwal membuat tabrakan berikutnya tidak terhindarkan'),
            o('c', 'Batalkan semuanya sekaligus agar tenang, nanti disusun ulang kalau sempat', 0, 'Membatalkan semua tanpa memilah merusak komitmen yang sebenarnya masih bisa dipenuhi'),
          ]),
        ev('ev_3', 'Situasi 3 - Perlengkapan Belum Siap',
          'Barang dan perlengkapan yang kamu butuhkan untuk beraktivitas bulan ini belum lengkap dan berserakan. Apa langkahmu?',
          [
            o('a', 'Daftar dulu apa yang benar-benar dibutuhkan, kumpulkan yang sudah ada, baru lengkapi yang kurang', 20, 'Daftar kebutuhan mencegah beli ganda dan memastikan yang paling penting siap duluan'),
            o('b', 'Beli semua yang terasa kurang sekaligus agar langsung lengkap', 5, 'Belanja tanpa daftar di awal bulan berisiko menguras anggaran untuk barang yang ternyata sudah ada'),
            o('c', 'Jalani saja apa adanya tanpa menyiapkan apa-apa', 0, 'Bulan baru tanpa persiapan perlengkapan hanya memindahkan kerepotan ke setiap harinya'),
          ]),
        ev('ev_4', 'Situasi 4 - Target Bulan Lalu Belum Selesai',
          'Ada target bulan September yang belum selesai dan terbawa ke Oktober. Bagaimana kamu menyikapinya?',
          [
            o('a', 'Pilih satu sisa target yang paling penting, beri tenggat baru yang realistis, dan mulai dari langkah terkecil', 20, 'Satu sisa target yang diselesaikan tuntas lebih berharga daripada semua sisa target yang terus dibawa-bawa'),
            o('b', 'Lanjutkan semua sisa target sekaligus bersamaan dengan target baru Oktober', 5, 'Menumpuk semua sisa target dengan target baru tanpa memilah membuat tidak ada yang selesai'),
            o('c', 'Lupakan semua sisa target bulan lalu dan anggap Oktober mulai dari nol', 0, 'Membuang sisa target tanpa evaluasi menjamin pelajaran yang sama hilang begitu saja'),
          ]),
        ev('ev_5', 'Situasi 5 - Pekerjaan Tertunda Menumpuk',
          'Beberapa pekerjaan tertunda dari minggu-minggu sebelumnya menumpuk dan semuanya terasa mendesak. Apa yang kamu lakukan?',
          [
            o('a', 'Urutkan pekerjaan tertunda berdasarkan dampak dan tenggat, selesaikan satu per satu mulai dari yang paling penting', 20, 'Tumpukan yang diurutkan berubah menjadi antrean yang bisa diselesaikan; yang tidak diurutkan tetap menjadi beban'),
            o('b', 'Kerjakan yang paling mudah dulu semuanya agar tumpukan cepat berkurang', 10, 'Yang mudah selesai duluan memang melegakan, tetapi yang penting dan mendesak bisa kedaluwarsa menunggu'),
            o('c', 'Tunggu sampai benar-benar mepet baru kerjakan semuanya sekaligus', 0, 'Menunda yang tertunda hanya melipatgandakan tekanan tanpa menambah waktu'),
          ]),
        ev('ev_6', 'Situasi 6 - Ruang Kerja Berantakan',
          'Meja dan area kerjamu berantakan setelah sebulan penuh dipakai. Awal bulan rasanya ingin yang segar. Apa langkahmu?',
          [
            o('a', 'Luangkan waktu khusus merapikan: buang yang tidak perlu, kembalikan tiap barang ke tempatnya', 20, 'Ruang kerja yang rapi di awal bulan menghemat waktu mencari barang selama sebulan penuh'),
            o('b', 'Geser barang-barang ke satu sudut agar meja terlihat rapi sekilas', 5, 'Rapi sekilas tanpa memilah hanya memindahkan kekacauan ke sudut lain yang akan meledak lagi'),
            o('c', 'Biarkan saja, yang penting langsung mulai bekerja', 0, 'Memulai bulan baru di atas kekacauan lama membuat setiap pekerjaan kecil terasa lebih berat'),
          ]),
      ],
    },
  };
}

function preCheckScenario() {
  const ev = (id, title, description, opts) => ({ id, title, description, options: opts });
  const o = (id, label, delta, hint) => ({ id, label, delta, hint });
  return {
    game: 'DECISION_PRIORITY',
    target_score: 100,
    scenario: {
      currency: 'POINTS',
      initial_balance: 0,
      events: [
        ev('ev_1', 'Situasi 1 - Sebelum Menggunakan Alat',
          'Kamu akan memakai sebuah alat (misalnya setrika, blender, atau perkakas) yang sudah lama tidak dipakai. Apa urutan tindakan yang paling tepat?',
          [
            o('a', 'Periksa dulu kondisi alat dan kabelnya, pastikan aman, baru gunakan sesuai fungsinya', 20, 'Pengecekan kondisi sebelum memakai alat mencegah kecelakaan yang sebenarnya bisa dihindari dalam satu menit'),
            o('b', 'Langsung pakai saja, kalau rusak baru berhenti', 0, 'Menunggu sampai rusak atau celaka untuk berhenti justru adalah cara paling mahal untuk mengecek'),
            o('c', 'Minta orang lain mencoba dulu, kalau aman baru kamu pakai', 5, 'Memindahkan risiko ke orang lain bukan pengecekan - tanggung jawab memastikan keamanan ada padamu'),
          ]),
        ev('ev_2', 'Situasi 2 - Sebelum Mengirim Dokumen',
          'Kamu akan mengirim dokumen penting (lamaran, tugas, atau berkas). Tenggatnya mepet. Apa yang kamu lakukan?',
          [
            o('a', 'Cek ulang isi, nama file, dan alamat tujuan, baru kirim walau butuh beberapa menit ekstra', 20, 'Dokumen penting yang salah kirim tidak bisa ditarik kembali - beberapa menit pengecekan jauh lebih murah daripada mengulang'),
            o('b', 'Kirim secepatnya agar tidak telat, cek belakangan kalau sempat', 0, 'Kecepatan tanpa pengecekan pada dokumen penting berisiko mengirim kesalahan yang permanen'),
            o('c', 'Kirim dulu versi seadanya, nanti kirim revisi kalau ada yang salah', 5, 'Mengirim versi seadanya dengan rencana revisi menunjukkan kamu tidak memeriksa sebelum bertindak'),
          ]),
        ev('ev_3', 'Situasi 3 - Sebelum Meninggalkan Rumah',
          'Kamu akan pergi dan rumah kosong seharian. Apa urutan pengecekan yang paling tepat sebelum berangkat?',
          [
            o('a', 'Cek kompor dan listrik sudah mati atau aman, kunci pintu dan jendela, bawa kunci serta barang penting, baru berangkat', 20, 'Urutan cek sebelum meninggalkan rumah menyelamatkan dari bahaya dan dari keharusan pulang di tengah jalan'),
            o('b', 'Langsung berangkat agar tidak telat, rumah aman-aman saja biasanya', 0, 'Mengandalkan kebiasaan aman tanpa mengecek adalah cara paling umum musibah rumah kosong terjadi'),
            o('c', 'Kunci pintu saja, urusan kompor dan listrik dipikir nanti lewat telepon ke keluarga', 5, 'Mengunci saja tanpa mengecek sumber bahaya meninggalkan risiko terbesar tanpa pengawasan'),
          ]),
        ev('ev_4', 'Situasi 4 - Sebelum Menyimpan Makanan',
          'Ada makanan sisa yang ingin kamu simpan untuk besok. Apa tindakan yang paling tepat?',
          [
            o('a', 'Pastikan makanan masih layak, simpan dalam wadah bersih tertutup, dan beri tanda kapan menyimpannya', 20, 'Makanan yang dicek kelayakannya dan disimpan tertutup dengan penanda waktu aman dikonsumsi besok'),
            o('b', 'Langsung masukkan ke kulkas apa adanya dalam piring terbuka', 5, 'Menyimpan tanpa wadah tertutup dan tanpa mengecek kelayakan mengundang kontaminasi dan kebusukan'),
            o('c', 'Biarkan saja di meja tertutup serbet, besok masih bisa dimakan', 0, 'Makanan yang dibiarkan di suhu ruang semalaman berisiko basi dan berbahaya bagi kesehatan'),
          ]),
        ev('ev_5', 'Situasi 5 - Sebelum Menggunakan Perangkat',
          'Kamu akan memakai HP atau laptop untuk pekerjaan penting yang lama. Apa persiapan yang paling tepat?',
          [
            o('a', 'Cek daya baterai atau colokan, pastikan data penting tersimpan, siapkan koneksi, baru mulai bekerja', 20, 'Pekerjaan penting yang disiapkan daya, data, dan koneksinya tidak akan terhenti di tengah jalan'),
            o('b', 'Langsung mulai bekerja, urusan baterai dan data dipikir kalau ada masalah', 0, 'Bekerja tanpa memastikan daya dan simpanan data mempertaruhkan seluruh hasil pekerjaanmu'),
            o('c', 'Pinjam perangkat orang lain saja agar tidak repot menyiapkan perangkat sendiri', 5, 'Bergantung pada perangkat pinjaman tanpa persiapan sendiri menambah risiko yang tidak perlu'),
          ]),
        ev('ev_6', 'Situasi 6 - Sebelum Menyelesaikan Pekerjaan',
          'Pekerjaanmu sudah hampir selesai dan tinggal dinyatakan beres. Apa langkah terakhirmu?',
          [
            o('a', 'Cek ulang hasilnya satu per satu, perbaiki yang kurang, rapikan, baru nyatakan selesai', 20, 'Pengecekan akhir sebelum menyatakan selesai membedakan pekerjaan yang rapi dari yang asal beres'),
            o('b', 'Langsung nyatakan selesai agar cepat, toh sudah dikerjakan', 0, 'Menyatakan selesai tanpa cek ulang menyerahkan penilaian kualitas pada keberuntungan'),
            o('c', 'Minta orang lain yang mengecek semuanya, kamu langsung lanjut ke hal lain', 5, 'Melempar seluruh pengecekan akhir ke orang lain berarti kamu tidak memastikan sendiri hasil kerjamu'),
          ]),
      ],
    },
  };
}

async function run() {
  console.log('=== Applying Migration 098: Fresh tasks for 2026-10-01 + 2026-10-02 ===');

  // 0. Pre-apply audit: expect zero rows on both dates.
  console.log('[0] Pre-apply audit of 2026-10-01 / 2026-10-02...');
  const rows01 = await get('odyssey_tasks?active_date=eq.2026-10-01&select=id,title,step_order,is_active');
  const rows02 = await get('odyssey_tasks?active_date=eq.2026-10-02&select=id,title,step_order,is_active');
  console.log('   rows on 2026-10-01: ' + JSON.stringify(rows01));
  console.log('   rows on 2026-10-02: ' + JSON.stringify(rows02));
  const existing = [...rows01, ...rows02];
  if (existing.length) {
    for (const t of existing) {
      const s = await get(`odyssey_task_submissions?task_id=eq.${t.id}&select=id,status`);
      if (s.length) throw new Error(`ABORT: task ${t.id} carries ${s.length} submissions — deactivate instead of delete.`);
    }
    throw new Error('ABORT: unexpected rows on Oct dates without submissions — audit manually before proceeding.');
  }

  // 1. Idempotent pre-clean of THIS migration's titles only.
  console.log('[1] Idempotent pre-clean (exact new titles, exact dates)...');
  const enc = (s) => encodeURIComponent(`"${s}"`);
  const pre01 = await get(`odyssey_tasks?active_date=eq.2026-10-01&select=id,title&title=in.(${NEW_01.map(enc).join(',')})`);
  const pre02 = await get(`odyssey_tasks?active_date=eq.2026-10-02&select=id,title&title=in.(${NEW_02.map(enc).join(',')})`);
  for (const rows of [pre01, pre02]) {
    if (rows.length) {
      const ids = rows.map((t) => t.id).join(',');
      const d = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, { method: 'DELETE', headers });
      if (!d.ok) throw new Error(`Pre-clean DELETE failed: ${d.status} ${await d.text()}`);
      console.log(`   removed ${rows.length} partial-run row(s).`);
    }
  }
  if (!pre01.length && !pre02.length) console.log('   no partial-run rows found.');

  // 2. Insert 6 tasks (payload mirrors 098 SQL migration).
  console.log('[2] Inserting 6 tasks...');
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Kalau Mau Mulai Lagi, Pilih yang Mana?',
      description:
        'Bulan baru, awal baru! Oktober baru saja dimulai dan ada 6 situasi khas awal bulan yang harus kamu hadapi: uang mulai menipis, jadwal berantakan, perlengkapan belum siap, target bulan lalu belum selesai, pekerjaan tertunda, dan ruang kerja yang berantakan. Pada setiap situasi, pilih tindakan yang paling masuk akal untuk memulai bulan baru dengan benar. Setiap pilihan ada konsekuensinya!',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-10-01',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: octStartScenario(),
    },
    {
      family_id: 'demo-crew-1',
      title: 'Sebelum & Sesudah Awal Bulan',
      description:
        'Buktikan awal bulanmu dimulai dengan perubahan nyata! Pilih satu tempat di sekitarmu (meja, tas kerja, lemari, area belajar, atau tempat menyimpan perlengkapan), lakukan satu perubahan kecil yang benar-benar nyata, lalu foto HASIL AKHIRNYA dan ceritakan perubahanmu.',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-10-01',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction:
          'Lakukan alur before -> action -> after: 1) BEFORE - amati dan pilih SATU tempat yang perlu diubah (contoh: meja, tas kerja, lemari, area belajar, atau tempat menyimpan perlengkapan). 2) ACTION - lakukan satu perubahan kecil yang benar-benar nyata di tempat itu (merapikan, menyusun ulang, membersihkan, atau menata ulang). 3) AFTER - ambil atau unggah tepat 1 foto HASIL AKHIR perubahanmu (boleh kamera langsung atau galeri, maksimal 1 file). 4) Wajib tulis catatan dengan 3 bagian: (a) Apa - jelaskan apa yang kamu ubah dan bagaimana kondisi sebelumnya, (b) Kenapa - jelaskan kenapa tempat itu perlu diubah, (c) Manfaat - jelaskan apa manfaat perubahan itu untuk aktivitasmu. Admin akan menolak submission tanpa foto hasil akhir, tanpa 3 bagian catatan (apa/kenapa/manfaat), atau perubahan yang tidak terlihat nyata.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Satu Hal yang Mau Kamu Benerin Bulan Ini',
      description:
        'Oktober punya 31 hari - cukup untuk memperbaiki satu hal dengan sungguh-sungguh! Pilih SATU kebiasaan atau hal praktis yang ingin kamu perbaiki selama bulan Oktober, lalu tulis kondisimu sekarang, perubahan yang kamu inginkan, dan langkah pertama yang realistis.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-10-01',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt:
          'Tulis jawabanmu dengan format 3 bagian:\n1) Kondisi sekarang - jelaskan kondisi saat ini dari SATU hal yang ingin kamu perbaiki (contoh: kebiasaan bangun siang, meja yang selalu berantakan, keuangan yang bocor, atau cara mengatur waktu). Ceritakan sejujurnya seperti apa kondisinya sekarang dan apa dampaknya buatmu:\n2) Perubahan yang ingin dilakukan selama Oktober - jelaskan perubahan seperti apa yang kamu targetkan sampai akhir Oktober. Buat spesifik dan terukur, bukan sekadar ingin lebih baik:\n3) Langkah pertama yang realistis - tulis SATU langkah pertama yang kecil, konkret, dan bisa kamu mulai dalam 24 jam ke depan:\n\nFokus pada satu hal saja, jangan membahas banyak hal sekaligus. Tulis dari pengalamanmu sendiri. Admin akan menolak jawaban di bawah 200 karakter, tanpa 3 bagian di atas, membahas terlalu banyak hal sekaligus, atau langkah pertama yang tidak realistis.',
        minimum_characters: 200,
        maximum_characters: 1000,
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Tunjukkan Satu Trik yang Sering Kamu Pakai',
      description:
        'Punya trik andalan yang hampir setiap hari kamu pakai? Saatnya menunjukkannya! Rekam video maksimal 60 detik yang MENDEMONSTRASIKAN satu trik praktis yang benar-benar sering kamu gunakan - misalnya shortcut HP, cara merapikan kabel, trik packing, cara menyimpan file, atau trik sederhana saat bekerja. Orang lain harus bisa menirunya setelah menonton videomu!',
      task_type: 'VIDEO',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 1,
      active_date: '2026-10-02',
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
            'Rekam video maksimal 60 detik dengan struktur:\n1. Sebutkan SATU trik spesifik yang benar-benar sering kamu pakai (contoh: shortcut HP, cara merapikan kabel, trik packing, cara menyimpan file, atau trik sederhana saat bekerja) - fokus pada satu trik saja, jangan banyak-banyak\n2. DEMONSTRASIKAN trik itu secara langsung di depan kamera (wajib ada demonstrasi nyata, bukan hanya bicara atau bercerita) - tunjukkan objek atau perangkat yang kamu gunakan jika relevan (HP, kabel, tas, aplikasi, peralatan kerja)\n3. Jelaskan kenapa trik itu berguna dan dalam situasi apa orang lain bisa menirunya\n\nCatatan: submission yang hanya berbicara atau bercerita tanpa demonstrasi TIDAK memenuhi task ini. Admin akan menolak video tanpa demonstrasi nyata, tanpa satu trik yang spesifik, tanpa objek atau perangkat yang ditunjukkan padahal relevan, durasi melebihi 60 detik, atau trik yang tidak bisa ditiru orang lain.',
        },
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Sebelum Dipakai, Cek Dulu',
      description:
        'Orang yang teliti selalu mengecek dulu sebelum bertindak! Hadapi 6 situasi sehari-hari di mana kamu harus memastikan sesuatu aman dan siap dipakai: sebelum menggunakan alat, mengirim dokumen, meninggalkan rumah, menyimpan makanan, menggunakan perangkat, dan menyelesaikan pekerjaan. Pada setiap situasi, tentukan tindakan pengecekan yang paling tepat. Ini soal urutan dan prioritas tindakan - bukan mencari yang janggal!',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 2,
      active_date: '2026-10-02',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: preCheckScenario(),
    },
    {
      family_id: 'demo-crew-1',
      title: 'Bikin Checklist yang Bisa Langsung Dipakai',
      description:
        'Tutorial panjang jarang dibaca sampai habis - checklist praktis selalu dipakai! Buat SATU dokumen checklist (maksimal 10 item) yang bisa langsung digunakan orang lain tanpa membaca tutorial panjang. Pilih topik praktis: checklist sebelum berangkat kerja, sebelum packing, sebelum mengirim dokumen, sebelum membersihkan sesuatu, atau sebelum melakukan pekerjaan tertentu.',
      task_type: 'DOCUMENT_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-10-02',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        allowed_extensions: ['pdf', 'docx', 'txt', 'jpg', 'jpeg', 'png'],
        accepted_extensions: ['pdf', 'docx', 'txt', 'jpg', 'jpeg', 'png'],
        max_file_size_mb: 4,
        instruction:
          'Buat SATU dokumen checklist dengan aturan: 1) Maksimal 10 checklist items - tidak boleh lebih dari 10. 2) Setiap item harus berupa tindakan atau cek yang konkret (contoh format: [ ] Cek ..., [ ] Pastikan ..., [ ] Bawa ...) - bukan paragraf penjelasan panjang. 3) Dokumen harus bisa LANGSUNG DIGUNAKAN orang lain sebagai quick checklist tanpa perlu membaca tutorial (orang lain cukup mencentang satu per satu). 4) Pilih SATU topik praktis (contoh: checklist sebelum berangkat kerja, sebelum packing, sebelum mengirim dokumen, sebelum membersihkan sesuatu, atau sebelum melakukan pekerjaan tertentu). 5) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: topik checklistmu dan untuk siapa checklist itu berguna. Admin akan menolak dokumen berupa tutorial atau guide panjang tanpa format checklist, dokumen dengan lebih dari 10 item, item yang bukan tindakan atau cek konkret, atau dokumen yang tidak bisa langsung dipakai.',
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

  // 3. Verify exactly 3 active tasks per date.
  console.log('[3] Verifying 2026-10-01 + 2026-10-02 state...');
  for (const [date, titles] of [['2026-10-01', NEW_01], ['2026-10-02', NEW_02]]) {
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

  // 4. Bump schema version.
  console.log('[4] Updating odyssey_schema_version...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_schema_version`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '098_tasks_replace_2026_10_01_02', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 098 Applied Successfully ===');
}

run().catch((err) => {
  console.error('Migration failed:', err);
  process.exit(1);
});
