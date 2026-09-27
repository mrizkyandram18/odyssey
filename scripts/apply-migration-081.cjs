require('dotenv').config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SERVICE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  apikey: SERVICE_KEY,
  Authorization: 'Bearer ' + SERVICE_KEY,
  'Content-Type': 'application/json',
  Prefer: 'return=representation',
};

async function run() {
  console.log('=== Applying Migration 081: Tasks for 11 September 2026 ===');

  // 1. Deactivate old placeholder tasks 533 & 534 on 2026-09-11
  console.log('1. Deactivating placeholder tasks 533 and 534...');
  const patchRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(533,534)&active_date=eq.2026-09-11`, {
    method: 'PATCH',
    headers,
    body: JSON.stringify({ is_active: false }),
  });
  const patched = await patchRes.json();
  console.log(`   Deactivated ${patched.length} placeholder tasks.`);

  // 2. Check if new tasks already exist (idempotency check)
  const existingRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?family_id=eq.demo-crew-1&active_date=eq.2026-09-11&is_active=eq.true&select=id,title,step_order`, {
    headers,
  });
  const existing = await existingRes.json();
  if (existing.length > 0) {
    console.log(`   Found ${existing.length} existing active tasks on 2026-09-11. Deactivating them first for clean idempotent run...`);
    const ids = existing.map(t => t.id).join(',');
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, {
      method: 'PATCH',
      headers,
      body: JSON.stringify({ is_active: false }),
    });
  }

  // 3. Insert 3 new tasks
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Prioritas Dulu, Baru Gas',
      description: 'Hadapi 6 situasi nyata di dunia kerja dan kehidupan sehari-hari. Latih kemampuan membedakan hal mendesak vs penting, mengelola distraksi, dan mengambil keputusan praktis dengan bijak.',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-11',
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
              id: 'sit_1',
              title: 'Situasi 1 — Deadline Mendekat vs Pesan Obrolan Santai',
              description: 'Laporan penting harus dikirim dalam waktu 45 menit ke atasan. Tiba-tiba ponselmu bergetar karena pesan obrolan santai di grup teman yang sedang ramai membicarakan topik seru.',
              options: [
                {
                  id: 'a',
                  label: 'Tutup notifikasi ponsel, fokus penuh menyelesaikan dan mengecek ulang laporan hingga terkirim',
                  delta: 20,
                  hint: 'Prioritas utama terlindungi dari distraksi',
                },
                {
                  id: 'b',
                  label: 'Buka chat sebentar untuk membalas beberapa kalimat sambil tetap membuka dokumen laporan',
                  delta: 5,
                  hint: 'Multitasking memecah fokus dan berisiko membuat laporan telat atau keliru',
                },
                {
                  id: 'c',
                  label: 'Asyik ikut nimbrung obrolan seru di grup dan menunda laporan hingga 10 menit sebelum batas akhir',
                  delta: 0,
                  hint: 'Sangat berisiko melewati deadline dan menghasilkan pekerjaan terburu-buru',
                },
              ],
            },
            {
              id: 'sit_2',
              title: 'Situasi 2 — Rekan Minta Bantuan vs Tugas Sendiri Belum Beres',
              description: 'Rekan kerja meminta bantuanmu untuk merapikan file presentasinya. Di saat yang sama, tugas pribadimu yang berprioritas tinggi dan bermuatan deadline ketat masih tersisa separuh.',
              options: [
                {
                  id: 'a',
                  label: 'Jelaskan dengan sopan bahwa tugas utamamu sedang mendesak, lalu tawarkan bantuan setelah tugasmu rampung',
                  delta: 20,
                  hint: 'Komunikasi transparan dan tetap bertanggung jawab atas amanah utama',
                },
                {
                  id: 'b',
                  label: 'Langsung tinggalkan tugas pribadimu demi membantu rekan kerja agar dianggap tidak enakan',
                  delta: 5,
                  hint: 'Membantu itu baik, namun menelantarkan tanggung jawab utama dapat merugikan tim',
                },
                {
                  id: 'c',
                  label: 'Menolak ketus tanpa penjelasan apa pun sehingga menimbulkan ketegangan antar rekan tim',
                  delta: 0,
                  hint: 'Merusak relasi profesional tanpa memberikan alternatif solusi',
                },
              ],
            },
            {
              id: 'sit_3',
              title: 'Situasi 3 — Masalah Operasional Mendadak vs Tugas Rutin Harian',
              description: 'Saat kamu sedang mengerjakan rekap data rutin bulanan, terjadi gangguan pada alur layanan pelanggan yang membutuhkan penanganan segera agar operasional tidak mandek.',
              options: [
                {
                  id: 'a',
                  label: 'Tunda sementara rekap rutin, tangani gangguan operasional mendesak, lalu lanjutkan rekap setelah normal',
                  delta: 20,
                  hint: 'Merespons masalah berdampak besar (urgent & important) terlebih dahulu',
                },
                {
                  id: 'b',
                  label: 'Tetap fokus pada rekap rutin dan membiarkan gangguan pelanggan menunggu berjam-jam',
                  delta: 5,
                  hint: 'Kaku pada rutinitas tanpa peka terhadap eskalasi krisis operasional',
                },
                {
                  id: 'c',
                  label: 'Panik dan meninggalkan semua pekerjaan tanpa berkoordinasi dengan tim lain',
                  delta: 0,
                  hint: 'Reaktif tanpa kendali memperburuk situasi tim',
                },
              ],
            },
            {
              id: 'sit_4',
              title: 'Situasi 4 — Instruksi Tambahan Atasan vs Beban Kerja Sedang Penuh',
              description: 'Atasan memberikan satu tugas mendadak baru di siang hari, padahal jadwalmu hari ini sudah padat dengan target yang disepakati sebelumnya.',
              options: [
                {
                  id: 'a',
                  label: 'Sampaikan daftar tugas yang sedang berjalan secara objektif dan tanyakan urutan prioritas yang diinginkan atasan',
                  delta: 20,
                  hint: 'Membantu pimpinan menyelaraskan ekspektasi berdasarkan realitas kapasitas',
                },
                {
                  id: 'b',
                  label: 'Langsung mengiyakan semua tugas tanpa konfirmasi lalu menyelesaikan semuanya secara asal-asalan',
                  delta: 5,
                  hint: 'Overpromising berujung pada penurunan mutu kerja dan keterlambatan',
                },
                {
                  id: 'c',
                  label: 'Mengeluh di media sosial bahwa beban kerjamu tidak manusiawi sebelum berdiskusi dengan atasan',
                  delta: 0,
                  hint: 'Tindakan tidak etis yang merusak reputasi profesional diri sendiri',
                },
              ],
            },
            {
              id: 'sit_5',
              title: 'Situasi 5 — Menemukan Detail Minor vs Fitur Kritis Siap Rilis',
              description: 'Dua jam sebelum sistem/dokumen utama dirilis ke klien, kamu menemukan detail estetika minor yang kurang sempurna, sementara fungsi utama sistem sudah lolos uji menyeluruh.',
              options: [
                {
                  id: 'a',
                  label: 'Catat temuan minor ke daftar perbaikan tahap berikutnya (backlog) dan amankan jadwal rilis utama',
                  delta: 20,
                  hint: 'Prinsip tepat waktu tanpa mengorbankan fungsi utama demi detail kosmetik',
                },
                {
                  id: 'b',
                  label: 'Membongkar ulang tata letak di menit-menit terakhir hingga rilis terlambat dan berisiko memicu bug baru',
                  delta: 5,
                  hint: 'Perfeksionisme tanpa pertimbangan risiko membahayakan komitmen rilis',
                },
                {
                  id: 'c',
                  label: 'Mengabaikan sepenuhnya dan tidak mencatat apa pun untuk perbaikan di masa depan',
                  delta: 0,
                  hint: 'Kehilangan kesempatan untuk evaluasi dan peningkatan mutu berkelanjutan',
                },
              ],
            },
            {
              id: 'sit_6',
              title: 'Situasi 6 — Mengatur Rencana Kerja Pagi Hari',
              description: 'Kamu baru tiba di tempat kerja dan memiliki 5 email, 2 dokumen harus ditandatangani, 1 rapat siang, dan persiapan materi besok. Bagaimana kamu memulai harimu?',
              options: [
                {
                  id: 'a',
                  label: 'Luangkan 10 menit memilah matriks prioritas (mendesak/penting), lalu kerjakan tugas berdampak tertinggi terlebih dahulu',
                  delta: 20,
                  hint: 'Perencanaan strategis di awal hari melipatgandakan efektivitas kerja',
                },
                {
                  id: 'b',
                  label: 'Kerjakan apa pun yang paling mudah dulu tanpa melihat tingkat kepentingannya',
                  delta: 10,
                  hint: 'Cepat selesai namun berisiko menunda tugas-tugas kritis berbobot besar',
                },
                {
                  id: 'c',
                  label: 'Bekerja secara acak sesuai email mana yang paling atas masuk ke inbox',
                  delta: 0,
                  hint: 'Kerja tanpa arah membuatmu sekadar reaktif terhadap notifikasi orang lain',
                },
              ],
            },
          ],
        },
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Bikin Tempatmu Lebih Siap',
      description: 'Pilih satu tempat yang biasa kamu pakai untuk bekerja, belajar, atau menyimpan barang penting (misalnya meja kerja, meja belajar, rak buku, atau ruang kerja pribadi). Rapikan dan bersihkan agar lebih nyaman dan siap pakai, lalu unggah 1 foto bukti hasilnya.',
      task_type: 'PHOTO_UPLOAD',
      evaluation_type: 'AUTO',
      step_order: 2,
      active_date: '2026-09-11',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        max_files: 1,
        camera_only: false,
        instruction: 'Pilih satu tempat yang biasa kamu pakai untuk bekerja, belajar, atau menyimpan perlengkapan penting. Rapikan dan tata tempat tersebut agar bersih, teratur, dan nyaman digunakan. Ambil 1 foto bukti kondisi tempat yang sudah rapi lalu unggah sebagai bukti.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Kalau Ada Masalah di Tempat Kerja',
      description: 'Bayangkan kamu menghadapi kendala atau kesalahan di tempat kerja yang memperlambat pekerjaanmu atau berdampak pada rekan lain. Refleksikan tindakan profesional yang akan kamu ambil secara bertanggung jawab.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-11',
      reward_coins: 30,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt: 'Bayangkan kamu menghadapi masalah di tempat kerja yang menghambat pekerjaanmu atau berdampak pada orang lain. Jelaskan: 1) Apa yang kamu lakukan pertama kali? 2) Siapa yang kamu hubungi/ajak komunikasi? 3) Apa yang sebaiknya dihindari saat menghadapi masalah tersebut? 4) Mengapa kamu memilih langkah tersebut?',
        minimum_characters: 80,
        maximum_characters: 1000,
      },
    },
  ];

  console.log('2. Inserting 3 new canonical tasks...');
  const postRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks`, {
    method: 'POST',
    headers,
    body: JSON.stringify(tasks),
  });
  if (!postRes.ok) {
    const errText = await postRes.text();
    throw new Error(`Failed to insert tasks: ${postRes.status} ${errText}`);
  }
  const created = await postRes.json();
  console.log(`   Successfully inserted ${created.length} tasks!`);
  created.forEach(t => {
    console.log(`   - [ID ${t.id}] Step #${t.step_order}: "${t.title}" (${t.task_type}, ${t.evaluation_type}, ${t.reward_coins}c / ${t.reward_xp}xp)`);
  });

  // 4. Update schema version
  console.log('3. Updating odyssey_schema_version...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_schema_version`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '081_tasks_2026_09_11', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 081 Applied Successfully ===');
}

run().catch(err => {
  console.error('Migration failed:', err);
  process.exit(1);
});
