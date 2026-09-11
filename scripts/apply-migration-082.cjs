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
  console.log('=== Applying Migration 082: Tasks for 12 September 2026 + payout announcement seed ===');

  // 1. Deactivate old placeholder tasks 535 & 536 on 2026-09-12
  console.log('1. Deactivating placeholder tasks 535 and 536...');
  const patchRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(535,536)&active_date=eq.2026-09-12`, {
    method: 'PATCH',
    headers,
    body: JSON.stringify({ is_active: false }),
  });
  const patched = await patchRes.json();
  console.log(`   Deactivated ${Array.isArray(patched) ? patched.length : 0} placeholder tasks.`);

  // 2. Check if new tasks already exist (idempotency check)
  const existingRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?family_id=eq.demo-crew-1&active_date=eq.2026-09-12&is_active=eq.true&select=id,title,step_order`, {
    headers,
  });
  const existing = await existingRes.json();
  if (existing.length > 0) {
    console.log(`   Found ${existing.length} existing active tasks on 2026-09-12. Deactivating them first for clean idempotent run...`);
    const ids = existing.map(t => t.id).join(',');
    await fetch(`${SUPABASE_URL}/rest/v1/odyssey_tasks?id=in.(${ids})`, {
      method: 'PATCH',
      headers,
      body: JSON.stringify({ is_active: false }),
    });
  }

  // 3. Insert 3 new tasks (initial admin configuration; all values editable via Admin UI)
  const tasks = [
    {
      family_id: 'demo-crew-1',
      title: 'Pilih yang Paling Penting',
      description: 'Hadapi 6 situasi nyata tentang mengatur prioritas dan membuat pilihan bijak. Setiap keputusan melatih kemampuan membedakan hal penting vs mendesak dan fokus pada yang berdampak terbesar.',
      task_type: 'MINI_GAME',
      evaluation_type: 'AUTO',
      step_order: 1,
      active_date: '2026-09-12',
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
              title: 'Situasi 1 — Pagi yang Penuh Tugas',
              description: 'Kamu bangun pagi dan melihat daftar tugas: PR yang dikumpulkan hari ini, kamar berantakan, dan ajakan main game dari teman. Waktu sebelum berangkat terbatas.',
              options: [
                { id: 'a', label: 'Kerjakan PR yang dikumpulkan hari ini terlebih dahulu hingga selesai', delta: 20, hint: 'Mendahulukan kewajiban dengan tenggat waktu yang jelas' },
                { id: 'b', label: 'Bereskan kamar sebentar lalu kerjakan PR dengan sisa waktu yang ada', delta: 10, hint: 'Baik untuk kerapian, namun PR berisiko tidak selesai maksimal' },
                { id: 'c', label: 'Main game dulu bersama teman dan mengerjakan PR nanti jika sempat', delta: 0, hint: 'Hiburan menggeser kewajiban utama yang berbatas waktu' },
              ],
            },
            {
              id: 'sit_2',
              title: 'Situasi 2 — Uang Saku Terbatas',
              description: 'Kamu memiliki uang saku untuk seminggu. Di saat yang sama ada kebutuhan alat tulis yang habis dan ada jajanan viral yang ingin dicoba bersama teman.',
              options: [
                { id: 'a', label: 'Beli alat tulis yang dibutuhkan untuk belajar, sisihkan sisanya sebagai cadangan', delta: 20, hint: 'Kebutuhan belajar didahulukan sebelum keinginan sesaat' },
                { id: 'b', label: 'Beli jajanan viral porsi kecil, lalu beli alat tulis yang paling murah', delta: 5, hint: 'Masih mempertimbangkan kebutuhan meski tergoda tren' },
                { id: 'c', label: 'Habiskan untuk jajanan viral dan meminjam alat tulis ke teman setiap hari', delta: 0, hint: 'Ketergantungan pada orang lain karena keinginan sesaat' },
              ],
            },
            {
              id: 'sit_3',
              title: 'Situasi 3 — Belajar vs Hiburan Malam',
              description: 'Besok ada ulangan penting, tetapi serial favoritmu baru saja merilis episode terakhir yang sedang ramai dibicarakan semua teman.',
              options: [
                { id: 'a', label: 'Belajar terlebih dahulu dengan fokus, tonton serial setelah ulangan selesai', delta: 20, hint: 'Menunda kesenangan demi hasil jangka panjang' },
                { id: 'b', label: 'Nonton satu episode sambil membuka catatan belajar di sampingnya', delta: 5, hint: 'Multitasking membuat belajar dan hiburan sama-sama setengah-setengah' },
                { id: 'c', label: 'Maraton semua episode malam ini dan begadang, belajar seadanya menjelang pagi', delta: 0, hint: 'Mengorbankan persiapan dan kondisi tubuh sekaligus' },
              ],
            },
            {
              id: 'sit_4',
              title: 'Situasi 4 — Dua Janji di Waktu yang Sama',
              description: 'Kamu sudah berjanji membantu orang tua di rumah pada sore hari, lalu teman mengajak pergi ke acara seru di jam yang sama.',
              options: [
                { id: 'a', label: 'Tepati janji membantu orang tua, lalu ajak teman bertemu di lain waktu', delta: 20, hint: 'Menjaga komitmen yang sudah dibuat terlebih dahulu' },
                { id: 'b', label: 'Minta izin orang tua dengan sopan dan tawarkan mengganti bantuan di pagi atau malam hari', delta: 10, hint: 'Bernegosiasi dengan jujur dan tetap bertanggung jawab' },
                { id: 'c', label: 'Pergi diam-diam ke acara teman tanpa memberi kabar ke rumah', delta: 0, hint: 'Mengingkari janji dan merusak kepercayaan keluarga' },
              ],
            },
            {
              id: 'sit_5',
              title: 'Situasi 5 — Tugas Kelompok yang Macet',
              description: 'Tugas kelompok harus dikumpulkan lusa, tetapi dua anggota belum mengerjakan bagiannya dan grup chat sepi tanpa kabar.',
              options: [
                { id: 'a', label: 'Sampaikan pembagian tugas yang jelas dengan tenggat tiap bagian, lalu tawarkan bantuan untuk bagian yang sulit', delta: 20, hint: 'Kepemimpinan praktis: kejelasan peran plus dukungan nyata' },
                { id: 'b', label: 'Kerjakan seluruh tugas sendirian agar cepat selesai tanpa berkoordinasi lagi', delta: 5, hint: 'Cepat selesai tetapi melelahkan dan tidak melatih kerja sama' },
                { id: 'c', label: 'Biarkan saja dan salahkan anggota lain jika nilai kelompok jelek', delta: 0, hint: 'Pasif dan menyalahkan tanpa usaha memperbaiki keadaan' },
              ],
            },
            {
              id: 'sit_6',
              title: 'Situasi 6 — Menyusun Rencana Satu Minggu',
              description: 'Minggu depan padat: ulangan, latihan kegiatan, dan acara keluarga. Kamu ingin semua berjalan tanpa ada yang terbengkalai.',
              options: [
                { id: 'a', label: 'Tulis semua kegiatan, tandai yang paling penting dan mendesak, lalu susun jadwal harian yang realistis', delta: 20, hint: 'Perencanaan tertulis membuat prioritas terlihat dan bisa dijalankan' },
                { id: 'b', label: 'Ingat-ingat saja semua jadwal di kepala dan jalani hari apa adanya', delta: 10, hint: 'Berisiko lupa dan kewalahan saat semua datang bersamaan' },
                { id: 'c', label: 'Tidak membuat rencana apa pun dan menangani apa pun yang datang paling akhir', delta: 0, hint: 'Hidup reaktif membuat hal penting mudah terlewat' },
              ],
            },
          ],
        },
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Bikin Rencana Sederhana',
      description: 'Susun rencana sederhana untuk satu minggu ke depan: tulis tujuan, langkah, dan jadwal harianmu. Tuangkan ke dalam dokumen lalu unggah sebagai bukti.',
      task_type: 'DOCUMENT_UPLOAD',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 2,
      active_date: '2026-09-12',
      reward_coins: 40,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        allowed_extensions: ['pdf', 'docx', 'xlsx', 'txt'],
        max_file_size_mb: 10,
        instruction: 'Buat dokumen rencana satu minggu: 1) Tulis 2-3 tujuan minggumu. 2) Jabarkan langkah kecil untuk tiap tujuan. 3) Susun jadwal harian yang realistis. 4) Simpan sebagai PDF/DOCX/XLSX/TXT (maks 10 MB) lalu unggah.',
      },
    },
    {
      family_id: 'demo-crew-1',
      title: 'Kalau Rencana Berubah',
      description: 'Rencana tidak selalu berjalan mulus. Bayangkan satu bagian penting dari rencanamu minggu ini gagal atau berubah mendadak. Refleksikan sikap dan langkah penyesuaianmu secara dewasa.',
      task_type: 'TEXT_RESPONSE',
      evaluation_type: 'ADMIN_REVIEW',
      step_order: 3,
      active_date: '2026-09-12',
      reward_coins: 30,
      reward_xp: 100,
      target_scope: 'ALL',
      is_active: true,
      config: {
        prompt: 'Bayangkan satu bagian penting dari rencanamu minggu ini gagal atau berubah mendadak. Jelaskan: 1) Apa yang berubah dan apa dampaknya? 2) Apa langkah penyesuaian yang kamu ambil? 3) Apa yang sebaiknya dihindari saat rencana berubah? 4) Pelajaran apa yang kamu petik tentang fleksibilitas?',
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

  // 4. Seed payout announcement via existing odyssey_system_config keys (no new table).
  console.log('3. Seeding payout announcement config (initial values; editable via Admin Pengaturan)...');
  const announcement = [
    { key: 'announcement_enabled', value: 'true' },
    { key: 'announcement_title', value: 'Pengumuman Pencairan' },
    { key: 'announcement_body', value: 'Pencairan saat ini mengalami penyesuaian jadwal karena sedang dilakukan proses verifikasi dan pengecekan transaksi. Hal ini dilakukan agar setiap pencairan dapat diproses dengan aman dan akurat. Mohon menunggu dan tidak perlu mengajukan ulang pencairan yang sudah masuk. Terima kasih atas pengertiannya.' },
    { key: 'announcement_audience', value: 'ALL' },
    { key: 'announcement_start_at', value: '2026-09-12T00:00:00+07:00' },
    { key: 'announcement_end_at', value: '2026-09-30T23:59:59+07:00' },
    { key: 'announcement_priority', value: 'normal' },
  ];
  const annRes = await fetch(`${SUPABASE_URL}/rest/v1/odyssey_system_config`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify(announcement),
  });
  if (!annRes.ok) {
    const errText = await annRes.text();
    throw new Error(`Failed to seed announcement config: ${annRes.status} ${errText}`);
  }
  console.log('   Announcement config seeded (7 keys).');

  // 5. Update schema version
  console.log('4. Updating odyssey_schema_version...');
  await fetch(`${SUPABASE_URL}/rest/v1/odyssey_schema_version`, {
    method: 'POST',
    headers: { ...headers, Prefer: 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '082_tasks_2026_09_12', updated_at: new Date().toISOString() }]),
  });
  console.log('=== Migration 082 Applied Successfully ===');
}

run().catch(err => {
  console.error('Migration failed:', err);
  process.exit(1);
});
