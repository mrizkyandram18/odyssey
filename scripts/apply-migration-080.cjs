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
  console.log('=== Applying Migration 080: Tasks for 10 September 2026 ===');

  // 1. Deactivate old placeholder tasks 531 & 532 on 2026-09-10
  console.log('1. Deactivating placeholder tasks 531 and 532...');
  const patchRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(531,532)&active_date=eq.2026-09-10`, {
    method: 'PATCH',
    headers,
    body: JSON.stringify({ is_active: false }),
  });
  const patched = await patchRes.json();
  console.log(`   Deactivated ${patched.length} placeholder tasks.`);

  // 2. Check if new tasks already exist (idempotency check)
  const existingRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?family_id=eq.demo-crew-1&active_date=eq.2026-09-10&is_active=eq.true&select=id,title,step_order`, {
    headers,
  });
  const existing = await existingRes.json();
  if (existing.length > 0) {
    console.log(`   Found ${existing.length} existing active tasks on 2026-09-10. Deactivating them first for clean idempotent run...`);
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
      title: 'Pilih Jalanmu',
      description: 'Hadapi 6 situasi nyata sehari-hari dengan saldo virtual awal Rp500.000. Setiap pilihan memiliki konsekuensi nyata terhadap kondisi finansialmu. Buat keputusan terbaik dan capai hasil yang bijak!',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-10',
      reward_coins: 50,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        game: 'DECISION_FINANCE',
        target_score: 100,
        scenario: {
          currency: 'IDR',
          initial_balance: 500000,
          events: [
            {
              id: 'sit_1',
              title: 'Situasi 1 — Ajakan Nongkrong Saat Anggaran Pas-Pasan',
              description: 'Teman-teman mengajak kumpul di kafe kekinian. Kamu ingin menjaga pertemanan, tetapi saldo mingguanmu harus dihemat.',
              options: [
                {
                  id: 'a',
                  label: 'Ikut nongkrong dan memesan paket makanan lengkap (-Rp60.000)',
                  delta: -60000,
                },
                {
                  id: 'b',
                  label: 'Ikut kumpul hanya memesan minuman terjangkau (-Rp20.000)',
                  delta: -20000,
                },
                {
                  id: 'c',
                  label: 'Menolak halus dan ajak bertemu di ruang publik gratis (Rp0)',
                  delta: 0,
                },
              ],
            },
            {
              id: 'sit_2',
              title: 'Situasi 2 — Investasi Diri & Buku Panduan Skill Kerja',
              description: 'Ada kesempatan membeli buku panduan skill kerja praktis yang sedang diskon dan sangat berguna untuk melamar kerja.',
              options: [
                {
                  id: 'a',
                  label: 'Beli buku panduan kerja praktis untuk investasi ilmu (-Rp50.000)',
                  delta: -50000,
                },
                {
                  id: 'b',
                  label: 'Cari modul dan tutorial gratis di internet/perpustakaan (Rp0)',
                  delta: 0,
                },
                {
                  id: 'c',
                  label: 'Tunda belajar dan gunakan uang untuk jajan santai (-Rp35.000)',
                  delta: -35000,
                },
              ],
            },
            {
              id: 'sit_3',
              title: 'Situasi 3 — Sepatu Kerja Mulai Rusak vs Model Terbaru',
              description: 'Sepatu yang biasa kamu pakai solnya mulai terbuka. Di toko ada sepatu kerja standar yang awet dan sepatu branded model terbaru yang mahal.',
              options: [
                {
                  id: 'a',
                  label: 'Beli sepatu kerja standar yang rapi, awet, dan fungsional (-Rp120.000)',
                  delta: -120000,
                },
                {
                  id: 'b',
                  label: 'Beli sepatu branded mahal demi gengsi (-Rp250.000)',
                  delta: -250000,
                },
                {
                  id: 'c',
                  label: 'Bawa ke tukang sol sepatu untuk diperbaiki dulu (-Rp25.000)',
                  delta: -25000,
                },
              ],
            },
            {
              id: 'sit_4',
              title: 'Situasi 4 — Peluang Bantuan Toko Tetangga',
              description: 'Seorang tetangga butuh bantuan membukukan stok dagangan dan merapikan barang selama setengah hari akhir pekan.',
              options: [
                {
                  id: 'a',
                  label: 'Ambil tawaran merapikan stok dagangan toko 3 jam (+Rp90.000)',
                  delta: 90000,
                },
                {
                  id: 'b',
                  label: 'Tolak karena ingin menghabiskan waktu rebahan seharian (Rp0)',
                  delta: 0,
                },
              ],
            },
            {
              id: 'sit_5',
              title: 'Situasi 5 — Biaya Darurat: Ban Bocor & Rem Aus',
              description: 'Saat perjalanan, ban motor kempes terkena paku dan rem terasa aus sehingga perlu diservis agar aman di jalan.',
              options: [
                {
                  id: 'a',
                  label: 'Tambal ban dan ganti kampas rem demi keselamatan berkendara (-Rp45.000)',
                  delta: -45000,
                },
                {
                  id: 'b',
                  label: 'Hanya tambal ban dan tunda servis rem yang aus (berisiko) (-Rp15.000)',
                  delta: -15000,
                },
                {
                  id: 'c',
                  label: 'Ganti pelek dan variasi baru yang tidak mendesak (-Rp160.000)',
                  delta: -160000,
                },
              ],
            },
            {
              id: 'sit_6',
              title: 'Situasi 6 — Komitmen Menabung & Mengatur Sisa Saldo',
              description: 'Setelah melewati berbagai situasi, kamu melihat saldo tersisa. Apa komitmen finansial yang kamu ambil sekarang?',
              options: [
                {
                  id: 'a',
                  label: 'Kunci 20% sisa uang ke pos tabungan darurat terpisah (-Rp50.000)',
                  delta: -50000,
                },
                {
                  id: 'b',
                  label: 'Habiskan sisa uang untuk belanja hiburan tanpa rencana (-Rp100.000)',
                  delta: -100000,
                },
                {
                  id: 'c',
                  label: 'Simpan sisa saldo sebagai cadangan tunai tanpa belanja impulsif (Rp0)',
                  delta: 0,
                },
              ],
            },
          ],
        },
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Seberapa Siap Kamu di Dunia Kerja?',
      description: 'Uji kesiapanmu menghadapi situasi nyata di tempat kerja: komunikasi profesional, manajemen deadline, menerima kritik, teamwork, meminta bantuan, hingga wawancara.',
      task_type: 'QUIZ',
      evaluation_type: 'AUTO',
      step_order: 2,
      active_date: '2026-09-10',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        questions: [
          {
            id: 'q1',
            question: 'Kamu diberi dua tugas penting oleh atasan dengan tenggat waktu (deadline) yang sama di sore hari, dan kamu sadar tidak mungkin menyelesaikan keduanya sendirian. Apa tindakan yang paling profesional?',
            options: [
              'A. Mengabari atasan lebih awal, menjelaskan estimasi waktu, dan meminta arahan tugas mana yang harus diprioritaskan',
              'B. Mengerjakan salah satu tugas diam-diam dan membiarkan tugas lainnya melewati deadline tanpa kabar',
              'C. Menyelesaikan keduanya secara tergesa-gesa meskipun kualitasnya berantakan dan banyak kesalahan',
              'D. Mengeluh kepada rekan kerja bahwa pembagian beban kerjamu tidak adil',
            ],
            correct_answer: 'A',
            explanation: 'Komunikasi proaktif sebelum deadline tiba jauh lebih profesional daripada membiarkan pekerjaan terbengkalai. Atasan dapat membantu menentukan prioritas atau membagi beban kerja.',
          },
          {
            id: 'q2',
            question: 'Saat evaluasi hasil kerja, atasan menyampaikan kritik bahwa laporanmu kurang teliti dan ada data yang keliru. Bagaimana respon terbaikmu?',
            options: [
              'A. Mendengarkan dengan tenang, mencatat poin perbaikan, dan menanyakan saran konkret untuk memperbaikinya',
              'B. Langsung membantah dan menyalahkan rekan kerja yang memberikan data awal',
              'C. Merasa tersinggung dan mendiamkan atasan selama beberapa hari ke depan',
              'D. Mengabaikan masukan tersebut karena merasa cara kerjamu sudah paling benar',
            ],
            correct_answer: 'A',
            explanation: 'Kritik di tempat kerja adalah bahan evaluasi untuk tumbuh. Menerima dengan tenang dan mencari solusi konkret mencerminkan kedewasaan dan profesionalisme.',
          },
          {
            id: 'q3',
            question: 'Dalam satu tim proyek, salah satu rekan kerjamu lambat merespons sehingga menghambat kemajuan pekerjaan bersama. Apa langkah pertama yang sebaiknya kamu ambil?',
            options: [
              'A. Mengajaknya berbicara secara empat mata dengan sopan untuk memahami kendalanya dan mencari solusi bersama',
              'B. Menyindir rekan tersebut di grup chat tim agar merasa malu dan segera menyelesaikan bagiannya',
              'C. Melaporkannya langsung ke pimpinan puncak tanpa konfirmasi terlebih dahulu',
              'D. Mengambil alih semua pekerjaannya sambil menggerutu di belakang rekan tersebut',
            ],
            correct_answer: 'A',
            explanation: 'Pendekatan empat mata secara profesional memungkinkan kita memahami akar masalah tanpa merusak hubungan kerja tim.',
          },
          {
            id: 'q4',
            question: 'Suatu pagi kamu mendadak sakit dan tidak bisa masuk bekerja. Prosedur etika kerja apa yang wajib kamu lakukan pertama kali?',
            options: [
              'A. Menghubungi atasan langsung sebelum jam kerja dimulai, mengabarkan kondisi, dan menginfokan tugas mendesak hari itu',
              'B. Mematikan ponsel dan baru memberikan penjelasan saat sudah sembuh beberapa hari kemudian',
              'C. Hanya mengunggah status sedang sakit di media sosial pribadi',
              'D. Menitipkan pesan kepada teman kerja tanpa menghubungi atasan langsung',
            ],
            correct_answer: 'A',
            explanation: 'Memberi kabar sebelum jam operasional dimulai adalah bentuk tanggung jawab kerja agar tim dapat mengantisipasi tugas-tugas mendesakmu.',
          },
          {
            id: 'q5',
            question: 'Kamu menghadapi kendala teknis saat mengerjakan tugas baru yang belum pernah kamu lakukan, padahal panduan tertulis sudah dibaca. Apa yang sebaiknya dilakukan?',
            options: [
              'A. Mencoba memecahkan masalah terlebih dahulu, lalu bertanya kepada rekan senior dengan pertanyaan terstruktur dan menunjukkan apa yang sudah dicoba',
              'B. Langsung meminta rekan kerja lain mengerjakan seluruh tugasmu dari awal',
              'C. Pura-pura mengerti dan menunggu sampai ditegur oleh atasan',
              'D. Menghentikan pekerjaan dan meninggalkan tugas tersebut tanpa penjelasan',
            ],
            correct_answer: 'A',
            explanation: 'Menunjukkan usaha mandiri terlebih dahulu lalu bertanya secara terstruktur menunjukkan inisiatif, menghargai waktu rekan senior, dan komitmen belajar.',
          },
          {
            id: 'q6',
            question: 'Saat wawancara kerja, pewawancara menanyakan apa kelemahan terbesarmu. Jawaban mana yang paling tepat dan profesional?',
            options: [
              'A. Menyebutkan kelemahan nyata yang sedang kamu perbaiki beserta langkah konkret yang sudah kamu lakukan',
              'B. Menjawab bahwa kamu adalah orang yang sempurna dan tidak memiliki kelemahan apa pun',
              'C. Mengatakan bahwa kamu sering malas dan tidak suka bangun pagi tanpa ada penjelasan lanjut',
              'D. Menolak menjawab karena menganggap kelemahan adalah rahasia pribadi',
            ],
            correct_answer: 'A',
            explanation: 'Pewawancara ingin melihat kejujuran, kesadaran diri (self-awareness), dan kemauanmu untuk terus belajar memperbaiki diri.',
          },
        ],
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Kalau Besok Harus Mandiri',
      description: 'Bayangkan mulai besok kamu harus mengatur hidupmu sendiri secara mandiri tanpa bergantung pada orang lain. Refleksikan kemampuan dasar yang paling kamu butuhkan untuk bertahan dan berkembang.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-10',
      reward_coins: 30,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt: 'Bayangkan mulai besok kamu harus mengatur hidupmu sendiri. Sebutkan 3 hal yang menurutmu paling perlu kamu kuasai, jelaskan kenapa hal itu penting, dan tuliskan satu langkah kecil yang bisa kamu mulai minggu ini.',
        minimum_characters: 80,
        maximum_characters: 1500,
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
    body: JSON.stringify([{ key: 'schema_version', value: '080_tasks_2026_09_10', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 080 Applied Successfully ===');
}

run().catch(err => {
  console.error('Migration failed:', err);
  process.exit(1);
});
