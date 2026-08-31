require('dotenv').config();

const SUPABASE_URL = process.env.SUPABASE_URL;
const SUPABASE_KEY = process.env.SUPABASE_SERVICE_KEY;

const headers = {
  'apikey': SUPABASE_KEY,
  'Authorization': 'Bearer ' + SUPABASE_KEY,
  'Content-Type': 'application/json',
  'Prefer': 'return=representation'
};

async function runSeed() {
  console.log('Seeding September 2026 Antigravity Program (Exact 2690 / 510 distribution)...');

  // 1. Ensure demo-crew-1 family exists
  await fetch(SUPABASE_URL + '/rest/v1/odyssey_families', {
    method: 'POST',
    headers: { ...headers, 'Prefer': 'resolution=ignore-duplicates' },
    body: JSON.stringify([{ id: 'demo-crew-1', name: 'Keluarga Demo' }])
  });

  // 2. Deactivate legacy tasks
  await fetch(SUPABASE_URL + '/rest/v1/odyssey_tasks?active_date=lt.2026-09-01', {
    method: 'PATCH',
    headers: { ...headers, 'Prefer': 'return=minimal' },
    body: JSON.stringify({ is_active: false })
  });
  await fetch(SUPABASE_URL + '/rest/v1/odyssey_tasks?active_date=gt.2026-09-30', {
    method: 'PATCH',
    headers: { ...headers, 'Prefer': 'return=minimal' },
    body: JSON.stringify({ is_active: false })
  });

  // 3. Delete previous ANTIGRAVITY tasks in Sep 2026
  await fetch(SUPABASE_URL + '/rest/v1/odyssey_tasks?title=like.ANTIGRAVITY:*&active_date=gte.2026-09-01&active_date=lte.2026-09-30', {
    method: 'DELETE',
    headers: { ...headers, 'Prefer': 'return=minimal' }
  });

  // 4. Define all September tasks (Days 1–24 = 2,690 coins, Days 25–30 = 510 coins, Total = 3,200 coins)
  const tasks = [
    // Day 1 (160)
    { title: 'ANTIGRAVITY: Kenalan dengan CV — Apa yang Harus Ada di CV?', description: 'Tonton video 5 menit tentang CV (YouTube) lalu jawab quiz 5 pertanyaan. HP only. Pelajari bagian: data diri, pendidikan, pengalaman, skill.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-01', reward_coins: 90, reward_xp: 60, config: { youtube_url: 'https://www.youtube.com/watch?v=example-cv-01', questions: [{ id: 'q1', question: 'Bagian CV yang berisi riwayat pendidikan disebut?', options: ['A. Profil', 'B. Pendidikan', 'C. Hobi', 'D. Alamat'], correct_answer: 'B' }, { id: 'q2', question: 'Untuk lulusan SMA tanpa pengalaman, bagian yang paling penting adalah?', options: ['A. Pengalaman kerja 10 tahun', 'B. Skill dan pendidikan', 'C. Foto liburan', 'D. Nilai rapor SD'], correct_answer: 'B' }, { id: 'q3', question: 'CV sebaiknya berapa halaman untuk pemula?', options: ['A. 1-2 halaman', 'B. 10 halaman', 'C. 5 halaman penuh warna', 'D. Tidak perlu CV'], correct_answer: 'A' }, { id: 'q4', question: 'Yang TIDAK perlu ada di CV?', options: ['A. Nama dan kontak', 'B. Pendidikan', 'C. Nomor PIN ATM', 'D. Skill'], correct_answer: 'C' }, { id: 'q5', question: 'Skill yang relevan untuk admin toko?', options: ['A. Menguasai kasir & komunikasi', 'B. Jago main game', 'C. Tidur siang', 'D. Tidak ada skill'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Tulis Profil Singkat Tentang Dirimu', description: 'Bayangkan memperkenalkan diri ke perusahaan. Tulis 3–5 kalimat: siapa kamu, pendidikan, hal yang disukai, pekerjaan yang ingin dicoba. Kerjakan di HP.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-01', reward_coins: 55, reward_xp: 40, config: { prompt: 'Tulis 3–5 kalimat tentang dirimu, pendidikanmu, hal yang kamu sukai, dan pekerjaan yang ingin kamu coba.', minimum_characters: 80, maximum_characters: 2000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Memory Challenge — Skill ↔ Pekerjaan', description: 'Mainkan mini game mencocokkan skill dengan pekerjaan (2 menit). Skor minimal 60 untuk lulus.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 3, active_date: '2026-09-01', reward_coins: 15, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 60 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 2 (150)
    { title: 'ANTIGRAVITY: Bikin CV Sederhana di Google Docs (HP)', description: 'Download template Google Docs di HP, isi dengan data dirimu dari Day 1, lalu upload screenshot atau link. Tidak perlu laptop.', task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-02', reward_coins: 80, reward_xp: 55, config: { instruction: 'Buka Google Docs di HP, buat CV 1 halaman sederhana, upload foto/screenshot hasil.', max_files: 2 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Cari 1 Lowongan untuk Lulusan SMA', description: 'Cari di Jobstreet/Glints/Instagram lowongan yang menerima SMA. Tulis: posisi, perusahaan, lokasi, pendidikan minimal, 3 skill diminta + skill apa yang belum kamu punya?', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-02', reward_coins: 70, reward_xp: 45, config: { prompt: 'Posisi, perusahaan, lokasi, pendidikan minimal, 3 skill diminta, dan skill apa yang ingin kamu pelajari.', minimum_characters: 100, maximum_characters: 2000 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 3 (90)
    { title: 'ANTIGRAVITY: Simpan & Bagikan CV di Google Drive (HP)', description: 'Simpan CV dari Day 2 ke Google Drive, atur link bisa dilihat, tulis link dan jelaskan 2 langkah cara membagikan.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-03', reward_coins: 70, reward_xp: 45, config: { prompt: 'Tulis link Google Drive CV-mu dan 2 langkah cara membagikan file.', minimum_characters: 60, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Puzzle Kata — PROFESIONAL', description: 'Susun huruf acak menjadi kata terkait dunia kerja. Mini puzzle 2-3 menit.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-03', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 50 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 4 (130)
    { title: 'ANTIGRAVITY: Kebutuhan vs Keinginan', description: 'Pilih mana kebutuhan vs keinginan dari 5 contoh belanja. Quiz 5 pertanyaan.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-04', reward_coins: 70, reward_xp: 45, config: { questions: [{ id: 'q1', question: 'Membeli beras untuk makan sehari-hari adalah?', options: ['A. Kebutuhan', 'B. Keinginan', 'C. Tidak penting', 'D. Hiburan'], correct_answer: 'A' }, { id: 'q2', question: 'Membeli skin game seharga 200rb padahal uang pas-pasan adalah?', options: ['A. Kebutuhan', 'B. Keinginan', 'C. Kewajiban', 'D. Investasi'], correct_answer: 'B' }, { id: 'q3', question: 'Bayar kos/transport untuk kerja termasuk?', options: ['A. Kebutuhan', 'B. Keinginan', 'C. Hadiah', 'D. Bonus'], correct_answer: 'A' }, { id: 'q4', question: 'Nongkrong kopi tiap hari demi gaya adalah?', options: ['A. Kebutuhan', 'B. Keinginan', 'C. Kebutuhan pokok', 'D. Wajib'], correct_answer: 'B' }, { id: 'q5', question: 'Menabung untuk dana darurat adalah?', options: ['A. Keinginan', 'B. Kebutuhan masa depan', 'C. Pemborosan', 'D. Tidak perlu'], correct_answer: 'B' }] }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Catat Pengeluaran Hari Ini', description: 'Tulis minimal 5 pengeluaran hari ini (makan, transport, dll) + totalnya. Gunakan notes HP.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-04', reward_coins: 60, reward_xp: 40, config: { prompt: 'Tulis minimal 5 pengeluaran hari ini dan totalnya.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 5 (130)
    { title: 'ANTIGRAVITY: Tulis Email Lamaran Singkat (HP)', description: 'Tulis email profesional 4–6 kalimat untuk melamar posisi dari Day 2. Perhatikan subjek, salam, isi, penutup.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-05', reward_coins: 80, reward_xp: 55, config: { prompt: 'Tulis subjek + isi email lamaran 4–6 kalimat yang sopan.', minimum_characters: 120, maximum_characters: 2000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Rapikan Nama File di HP', description: 'Rapikan 3 file di HP (beri nama jelas: CV_Nama.pdf, dsb). Upload foto sebelum & sesudah atau jelaskan perubahan nama.', task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-05', reward_coins: 50, reward_xp: 35, config: { instruction: 'Rapikan nama file di HP dan upload bukti foto/screenshot.', max_files: 2 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 6 (90)
    { title: 'ANTIGRAVITY: Pahami Syarat Lowongan — Skill Gap', description: 'Dari lowongan Day 2, tulis 3 skill wajib + 1 skill yang belum kamu punya dan cara mempelajarinya.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-06', reward_coins: 70, reward_xp: 45, config: { prompt: '3 skill wajib + 1 skill yang belum kamu punya dan rencanamu mempelajarinya.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Reaction Tap — Uji Kecepatan', description: 'Mini game ketuk cepat untuk melatih fokus dan refleks. Skor minimal 50.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-06', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 50 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 7 (60)
    { title: 'ANTIGRAVITY: Waspada Hoax Lowongan', description: 'Ciri lowongan palsu: minta uang, gaji tidak wajar, tanpa alamat jelas. Quiz 4Q.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-07', reward_coins: 60, reward_xp: 40, config: { questions: [{ id: 'q1', question: 'Lowongan yang meminta transfer uang di awal adalah?', options: ['A. Penipuan', 'B. Resmi', 'C. Wajar', 'D. Bonus'], correct_answer: 'A' }, { id: 'q2', question: 'Ciri lowongan terpercaya?', options: ['A. Alamat jelas & proses gratis', 'B. Minta uang jaminan', 'C. Janji gaji fantastis tanpa kerja', 'D. Tanpa kontak'], correct_answer: 'A' }, { id: 'q3', question: 'Sebelum melamar sebaiknya?', options: ['A. Cek alamat & ulasan perusahaan', 'B. Langsung transfer', 'C. Kirim KTP ke sembarang orang', 'D. Tidak perlu cek'], correct_answer: 'A' }, { id: 'q4', question: 'Jika ragu, kamu harus?', options: ['A. Tanya admin / cari info resmi', 'B. Tetap bayar', 'C. Diam saja', 'D. Sebar hoax'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },

    // Day 8 (120)
    { title: 'ANTIGRAVITY: Perbaiki Profil Diri (Revisi Day 1)', description: 'Perbaiki tulisan profil Day 1 setelah belajar CV & lowongan. Buat lebih meyakinkan dan sesuai pekerjaan incaran.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-08', reward_coins: 70, reward_xp: 45, config: { prompt: 'Tulis ulang profil 3–5 kalimat yang lebih baik dan meyakinkan.', minimum_characters: 80, maximum_characters: 2000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Formatting Dokumen Rapi di HP', description: 'Di Google Docs, praktik bold, bullet, dan rapikan CV. Upload screenshot hasil formatting.', task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-08', reward_coins: 50, reward_xp: 35, config: { instruction: 'Praktik bold/bullet di Docs dan upload screenshot.', max_files: 2 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 9 (130)
    { title: 'ANTIGRAVITY: Hitung Gaji Bersih', description: 'Gaji 2.500.000 - potongan 300.000 + lembur 200.000 = berapa? + 2 pertanyaan tunjangan. Quiz 4Q.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-09', reward_coins: 70, reward_xp: 45, config: { questions: [{ id: 'q1', question: '2.500.000 - 300.000 + 200.000 = ?', options: ['A. 2.400.000', 'B. 2.200.000', 'C. 2.000.000', 'D. 2.700.000'], correct_answer: 'A' }, { id: 'q2', question: 'Potongan BPJS termasuk?', options: ['A. Potongan wajib', 'B. Bonus', 'C. Hadiah', 'D. Tidak ada'], correct_answer: 'A' }, { id: 'q3', question: 'Gaji bersih adalah gaji kotor dikurangi?', options: ['A. Potongan', 'B. Bonus', 'C. Hadiah', 'D. THR'], correct_answer: 'A' }, { id: 'q4', question: 'Jika lembur tidak dibayar, gaji bersih akan?', options: ['A. Lebih kecil', 'B. Lebih besar', 'C. Sama saja', 'D. Tidak terpengaruh'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Perkenalkan Diri 30 Detik', description: 'Tulis naskah perkenalan 30 detik untuk interview: nama, pendidikan, kelebihan, motivasi kerja.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-09', reward_coins: 60, reward_xp: 40, config: { prompt: 'Tulis naskah perkenalan 30 detik yang percaya diri.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 10 (110)
    { title: 'ANTIGRAVITY: Bikin Anggaran Bulanan di Google Sheets (HP)', description: 'Buat sheet sederhana: pemasukan, pengeluaran, sisa. Upload screenshot atau link sheet.', task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-10', reward_coins: 90, reward_xp: 60, config: { instruction: 'Buat anggaran bulanan di Sheets (HP) dan upload bukti.', max_files: 2 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Memory — Harga vs Barang', description: 'Cocokkan harga dengan barang yang benar. Game memori 2 menit.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-10', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 50 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 11 (120)
    { title: 'ANTIGRAVITY: Simulasi Diskon — Pilih Paling Hemat', description: '3 barang diskon 20%, 30%, 50%. Hitung harga akhir dan pilih paling hemat. Quiz 4Q.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-11', reward_coins: 70, reward_xp: 45, config: { questions: [{ id: 'q1', question: 'Barang 100.000 diskon 20% menjadi?', options: ['A. 80.000', 'B. 90.000', 'C. 70.000', 'D. 100.000'], correct_answer: 'A' }, { id: 'q2', question: 'Barang 200.000 diskon 50% menjadi?', options: ['A. 100.000', 'B. 150.000', 'C. 50.000', 'D. 200.000'], correct_answer: 'A' }, { id: 'q3', question: 'Lebih hemat: A 100k diskon 20% vs B 100k diskon 30%?', options: ['A. B lebih hemat', 'B. A lebih hemat', 'C. Sama', 'D. Tidak ada diskon'], correct_answer: 'A' }, { id: 'q4', question: 'Diskon 10% dari 50.000 adalah?', options: ['A. 5.000', 'B. 10.000', 'C. 15.000', 'D. 500'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Rencana Tabungan 1 Bulan', description: 'Tulis target tabungan 1 bulan: nominal, tujuan, cara mencapainya (kurangi jajan, dll).', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-11', reward_coins: 50, reward_xp: 35, config: { prompt: 'Target tabungan, nominal, dan cara mencapainya.', minimum_characters: 60, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 12 (120)
    { title: 'ANTIGRAVITY: Jawab Pertanyaan Interview Dasar', description: 'Latihan jawab: "Ceritakan tentang diri Anda" — tulis jawaban 4–6 kalimat yang meyakinkan.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-12', reward_coins: 70, reward_xp: 45, config: { prompt: 'Jawab pertanyaan interview dengan percaya diri 4–6 kalimat.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Bikin Poster Diri di Canva (HP)', description: 'Buka Canva di HP, bikin poster profil diri sederhana (nama, skill, cita-cita). Upload screenshot.', task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-12', reward_coins: 50, reward_xp: 35, config: { instruction: 'Buat poster diri di Canva HP dan upload bukti.', max_files: 2 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 13 (100)
    { title: 'ANTIGRAVITY: Balas Komplain Pelanggan dengan Sopan', description: 'Kamu CS toko online. Pelanggan komplain barang telat. Tulis balasan WA yang sopan, solusi, dan meminta maaf.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-13', reward_coins: 80, reward_xp: 50, config: { prompt: 'Tulis balasan komplain yang sopan dan solutif.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Puzzle Prioritas Kerja', description: 'Urutkan 4 tugas kerja berdasarkan prioritas (deadline, penting). Mini game logic.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-13', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'MEDIUM', target_score: 60 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 14 (60)
    { title: 'ANTIGRAVITY: Rangkum Video Etika Kerja (3 poin)', description: 'Tonton video 3 menit tentang etika kerja (link YouTube), lalu rangkum 3 poin penting dengan bahasa sendiri.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-14', reward_coins: 60, reward_xp: 40, config: { prompt: 'Rangkum 3 poin penting tentang etika kerja.', minimum_characters: 60, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 15 (120)
    { title: 'ANTIGRAVITY: Bandingkan 2 Lowongan Sejenis', description: 'Cari 2 lowongan posisi sama beda perusahaan, bandingkan gaji, syarat, lokasi, dan pilih mana lebih cocok untukmu + alasan.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-15', reward_coins: 70, reward_xp: 45, config: { prompt: 'Bandingkan 2 lowongan dan jelaskan pilihanmu.', minimum_characters: 100, maximum_characters: 2000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Latihan Kasir — Hitung Kembalian', description: 'Kembalian 50.000 - 17.500 = ? dan 2 soal kasir lain. Quiz 3Q.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-15', reward_coins: 50, reward_xp: 35, config: { questions: [{ id: 'q1', question: 'Kembalian 50.000 - 17.500 = ?', options: ['A. 32.500', 'B. 33.500', 'C. 30.000', 'D. 35.000'], correct_answer: 'A' }, { id: 'q2', question: 'Jika barang 25.000 beli 2, total?', options: ['A. 50.000', 'B. 25.000', 'C. 75.000', 'D. 100.000'], correct_answer: 'A' }, { id: 'q3', question: 'Kembalian 100.000 - 45.000 = ?', options: ['A. 55.000', 'B. 45.000', 'C. 65.000', 'D. 50.000'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },

    // Day 16 (90)
    { title: 'ANTIGRAVITY: Email Follow-up Lamaran', description: 'Tulis email follow-up sopan 3–5 hari setelah melamar: tanyakan status, ucapkan terima kasih.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-16', reward_coins: 70, reward_xp: 45, config: { prompt: 'Tulis email follow-up lamaran yang sopan 3–5 kalimat.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Puzzle Jadwal Harian', description: 'Susun jadwal harian produktif dengan drag & drop. Game 2 menit.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-16', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 50 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 17 (120)
    { title: 'ANTIGRAVITY: Bersihkan Inbox Email (HP)', description: 'Buka Gmail di HP, hapus/archieve 10 email tidak penting, buat folder/label Lamaran. Upload screenshot.', task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-17', reward_coins: 60, reward_xp: 40, config: { instruction: 'Bersihkan inbox Gmail dan upload screenshot folder Lamaran.', max_files: 2 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Ikuti Instruksi SOP Packing', description: 'Baca 5 langkah SOP packing barang, lalu jawab quiz urutan yang benar.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-17', reward_coins: 60, reward_xp: 40, config: { questions: [{ id: 'q1', question: 'Langkah pertama packing yang benar?', options: ['A. Cek barang & jumlah', 'B. Langsung kirim', 'C. Tulis alamat asal', 'D. Tutup kardus dulu'], correct_answer: 'A' }, { id: 'q2', question: 'Setelah cek barang, selanjutnya?', options: ['A. Bungkus bubble wrap', 'B. Tidur', 'C. Buang barang', 'D. Foto selfie'], correct_answer: 'A' }, { id: 'q3', question: 'Langkah terakhir sebelum kirim?', options: ['A. Tempel alamat & cek kembali', 'B. Buka kembali', 'C. Buang kardus', 'D. Tidak perlu cek'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },

    // Day 18 (130)
    { title: 'ANTIGRAVITY: Budget 1 Minggu 500rb', description: 'Buat anggaran mingguan 500rb: makan, transport, tabungan, hiburan. Upload foto/sheet.', task_type: 'PHOTO_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-18', reward_coins: 80, reward_xp: 55, config: { instruction: 'Buat anggaran 1 minggu 500rb, upload bukti.', max_files: 2 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Tulis Pesan WA ke HR', description: 'Tulis pesan WA profesional ke HR: perkenalkan diri, tanya status lamaran, ucapkan terima kasih.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-18', reward_coins: 50, reward_xp: 35, config: { prompt: 'Pesan WA ke HR yang sopan dan profesional.', minimum_characters: 60, maximum_characters: 1000 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 19 (130)
    { title: 'ANTIGRAVITY: Simulasi Interview 3 Pertanyaan Sulit', description: 'Jawab: kelebihan, kekurangan, dan kenapa ingin bekerja di sini. Quiz pilihan ganda situasional.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-19', reward_coins: 80, reward_xp: 50, config: { questions: [{ id: 'q1', question: 'Saat ditanya kekurangan, jawaban terbaik?', options: ['A. Jujur + tunjukkan usaha perbaikan', 'B. Saya tidak punya kekurangan', 'C. Malas dan sering telat', 'D. Tidak jawab'], correct_answer: 'A' }, { id: 'q2', question: 'Kenapa ingin bekerja di sini? Jawab?', options: ['A. Sesuai skill & ingin belajar', 'B. Karena dekat rumah saja', 'C. Iseng', 'D. Tidak tahu'], correct_answer: 'A' }, { id: 'q3', question: 'Ceritakan kelebihanmu?', options: ['A. Sebutkan skill relevan + contoh', 'B. Saya sempurna', 'C. Tidak ada kelebihan', 'D. Diam'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Cari Cara Buat NPWP Online (Rangkum 3 Langkah)', description: 'Riset di HP cara daftar NPWP online, rangkum 3 langkah penting dengan bahasamu.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-19', reward_coins: 50, reward_xp: 35, config: { prompt: 'Rangkum 3 langkah daftar NPWP online.', minimum_characters: 60, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 20 (110)
    { title: 'ANTIGRAVITY: Menabung vs Cicilan', description: 'Pilih: cicilan 12x bunga vs nabung 6 bulan. Quiz hitung bunga cicilan.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-20', reward_coins: 90, reward_xp: 60, config: { questions: [{ id: 'q1', question: 'Cicilan 1.200.000 bunga 10% total jadi?', options: ['A. 1.320.000', 'B. 1.200.000', 'C. 1.100.000', 'D. 1.500.000'], correct_answer: 'A' }, { id: 'q2', question: 'Lebih hemat jika tidak urgent?', options: ['A. Nabung dulu', 'B. Cicilan bunga tinggi', 'C. Pinjam teman bunga tinggi', 'D. Tidak bayar'], correct_answer: 'A' }, { id: 'q3', question: 'Dana darurat ideal berapa bulan pengeluaran?', options: ['A. 3-6 bulan', 'B. 1 hari', 'C. Tidak perlu', 'D. 1 tahun penuh gaji'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Memory Challenge Lv2', description: 'Tingkat kesulitan naik — cocokkan 8 pasang kartu. Target 70.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-20', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'MEDIUM', target_score: 70 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 21 (120)
    { title: 'ANTIGRAVITY: Rangkuman Minggu 1–3', description: 'Tulis 5 hal penting yang kamu pelajari selama 3 minggu: CV, digital, finansial, komunikasi.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-21', reward_coins: 100, reward_xp: 65, config: { prompt: 'Tulis 5 hal penting yang kamu pelajari.', minimum_characters: 100, maximum_characters: 2000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Skor Challenge', description: 'Kumpulkan poin tertinggi di mini game. Target 70.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-21', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'MEDIUM', target_score: 70 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 22 (130)
    { title: 'ANTIGRAVITY: Final CV — Upload CV Lengkap', description: 'Upload CV final versi terbaik (PDF/DOCX) — gabungan perbaikan dari Day 8 & 1. Ini milestone utama!', task_type: 'DOCUMENT_UPLOAD', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-22', reward_coins: 90, reward_xp: 60, config: { instruction: 'Upload CV final terbaikmu (PDF/DOCX).', attachment_name: 'CV Final', max_file_size_mb: 5 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Naskah Perkenalan 60 Detik', description: 'Tulis naskah perkenalan 60 detik untuk interview/video profil: sapa, latar, skill, tujuan.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-22', reward_coins: 40, reward_xp: 30, config: { prompt: 'Naskah perkenalan 60 detik yang percaya diri.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 23 (70: 20+50)
    { title: 'ANTIGRAVITY: Portfolio HP — Kumpulkan Karya di Drive', description: 'Kumpulkan link/file karya (CV, poster Canva, sheet anggaran) ke 1 folder Drive. Tulis link folder.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-23', reward_coins: 20, reward_xp: 45, config: { prompt: 'Tulis link folder Drive portfolio dan isi 3 file di dalamnya.', minimum_characters: 60, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Rencana Keuangan Gaji Pertama', description: 'Gaji pertama 3jt — alokasikan: kebutuhan, tabungan, orang tua, pengembangan diri. Tulis persentasenya.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-23', reward_coins: 50, reward_xp: 35, config: { prompt: 'Alokasi gaji pertama 100% dengan persentase.', minimum_characters: 60, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 24 (100: 50+50) -> Days 1–24 sum = 2,690 Coins!
    { title: 'ANTIGRAVITY: Refleksi 24 Hari — Skill yang Ingin Dilanjutkan', description: 'Tulis refleksi: 3 skill yang paling berguna, 1 yang ingin kamu dalami setelah program, dan langkah selanjutnya.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-24', reward_coins: 50, reward_xp: 55, config: { prompt: '3 skill berguna, 1 yang ingin didalami, langkah selanjutnya.', minimum_characters: 100, maximum_characters: 2000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Celebration Mini Game', description: 'Rayakan perjalanan! Mainkan game perayaan — tidak perlu target tinggi, have fun!', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-24', reward_coins: 50, reward_xp: 30, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 40 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 25 (170: 90+80)
    { title: 'ANTIGRAVITY: Rencana Karir 3 Bulan', description: 'Tulis target 3 bulan ke depan: posisi impian, skill yang harus dipelajari, dan langkah konkrit.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-25', reward_coins: 90, reward_xp: 45, config: { prompt: 'Target karir 3 bulan dan langkah konkritnya.', minimum_characters: 80, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Evaluasi Portofolio', description: 'Periksa kembali file portofolio di Drive, pastikan semua izin akses sudah umum.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-25', reward_coins: 80, reward_xp: 30, config: { prompt: 'Tulis ringkasan link portofolio dan kelengkapannya.', minimum_characters: 50, maximum_characters: 1000 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 26 (110: 80+30)
    { title: 'ANTIGRAVITY: Simulasi Presentasi Diri', description: 'Buat rangkuman singkat perkenalan 1 menit untuk calon atasan.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-26', reward_coins: 80, reward_xp: 40, config: { prompt: 'Rangkuman perkenalan 1 menit yang efektif.', minimum_characters: 60, maximum_characters: 1000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Mini Game Fokus Refleks', description: 'Uji ketangkasan dan fokus dengan mini game memori.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-26', reward_coins: 30, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 50 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 27 (100: 70+30)
    { title: 'ANTIGRAVITY: Etika Berkomunikasi di Tempat Kerja', description: 'Quiz 4Q tentang komunikasi profesional via chat, email, dan tatap muka.', task_type: 'QUIZ', evaluation_type: 'AUTO', step_order: 1, active_date: '2026-09-27', reward_coins: 70, reward_xp: 45, config: { questions: [{ id: 'q1', question: 'Sikap terbaik saat menerima masukan dari atasan?', options: ['A. Mendengarkan dan belajar', 'B. Marah', 'C. Acuh', 'D. Keluar kerja'], correct_answer: 'A' }] }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Catat Pembelajaran Penting', description: 'Tulis 3 prinsip kerja profesional yang paling ingin kamu terapkan.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 2, active_date: '2026-09-27', reward_coins: 30, reward_xp: 20, config: { prompt: '3 prinsip kerja profesional pilihanmu.', minimum_characters: 50, maximum_characters: 1000 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 28 (70: 50+20)
    { title: 'ANTIGRAVITY: Persiapan Menghadapi Hari Pertama Kerja', description: 'Rangkum 5 hal yang harus disiapkan sebelum hari pertama kerja.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-28', reward_coins: 50, reward_xp: 35, config: { prompt: '5 hal persiapan hari pertama kerja.', minimum_characters: 60, maximum_characters: 1000 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Memory Refresh Challenge', description: 'Latihan daya ingat mini game 2 menit.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-28', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 50 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 29 (60: 40+20) -> Days 25–30 sum = 510 Coins!
    { title: 'ANTIGRAVITY: Checklist Dokumen Lamaran', description: 'Pastikan KTP, CV, dan berkas penting siap dikirim kapan saja.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-29', reward_coins: 40, reward_xp: 30, config: { prompt: 'Tulis checklist berkas yang sudah siap.', minimum_characters: 40, maximum_characters: 800 }, is_active: true, family_id: 'demo-crew-1' },
    { title: 'ANTIGRAVITY: Mini Game Kelulusan', description: 'Mainkan mini game santai penutup periode.', task_type: 'MINI_GAME', evaluation_type: 'AUTO', step_order: 2, active_date: '2026-09-29', reward_coins: 20, reward_xp: 15, config: { game: 'MEMORY', difficulty: 'EASY', target_score: 40 }, is_active: true, family_id: 'demo-crew-1' },

    // Day 30 (0 coins, 100 xp)
    { title: 'ANTIGRAVITY: Kelulusan Periode — Selamat!', description: 'Kamu berhasil menyelesaikan program 30 hari! Tulis kesan dan pesan akhirmu.', task_type: 'TEXT_RESPONSE', evaluation_type: 'ADMIN_REVIEW', step_order: 1, active_date: '2026-09-30', reward_coins: 0, reward_xp: 100, config: { prompt: 'Tulis kesan dan pesan akhirmu setelah 30 hari.', minimum_characters: 50, maximum_characters: 1500 }, is_active: true, family_id: 'demo-crew-1' }
  ];

  // Batch insert tasks (chunks of 10)
  for (let i = 0; i < tasks.length; i += 10) {
    const chunk = tasks.slice(i, i + 10);
    const res = await fetch(SUPABASE_URL + '/rest/v1/odyssey_tasks', {
      method: 'POST',
      headers,
      body: JSON.stringify(chunk)
    });
    if (!res.ok) {
      console.error('Error inserting task chunk:', await res.text());
      return;
    }
  }

  // Update schema version
  const verRes = await fetch(SUPABASE_URL + '/rest/v1/odyssey_schema_version', {
    method: 'POST',
    headers: { ...headers, 'Prefer': 'resolution=merge-duplicates' },
    body: JSON.stringify([{ key: 'schema_version', value: '052_seed_antigravity_september_2026', updated_at: new Date().toISOString() }])
  });
  console.log('Schema version updated:', verRes.status);

  console.log('September seed applied successfully to Production DB!');
}

runSeed().catch(console.error);
