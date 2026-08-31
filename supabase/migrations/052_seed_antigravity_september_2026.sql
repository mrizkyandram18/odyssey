-- ============================================================
-- Migration 052: Seed September 1–24 Antigravity Program (HP-only, Learn→Practice→Earn)
-- Total earnable ≈ 2770 coins (≤3000). 1 Koin = Rp100 → max Rp277.000 (< Rp300.000 cap)
-- Timezone: Asia/Jakarta. Family-scoped per existing families.
-- Legacy chores disabled (is_active=false outside Sep 1–24) without destructive delete.
-- ============================================================

-- 0. Ensure at least one family exists for seeding (demo crew renamed to families)
INSERT INTO odyssey_families (id, name)
VALUES ('demo-crew-1', 'Keluarga Demo')
ON CONFLICT (id) DO NOTHING;

-- 1. Disable legacy tasks that would mix with September program (but keep history)
--    Deactivate any task whose active_date is outside 2026-09-01..2026-09-30
UPDATE odyssey_tasks
SET is_active = false
WHERE active_date < DATE '2026-09-01' OR active_date > DATE '2026-09-30';

-- 1b. Also deactivate tasks with NULL active_date or non-HP chore keywords
UPDATE odyssey_tasks
SET is_active = false
WHERE title ILIKE '%menyapu%' OR title ILIKE '%mencuci%' OR title ILIKE '%memasak%'
   OR title ILIKE '%merapikan kamar%' OR title ILIKE '%membersihkan rumah%'
   OR title ILIKE '%cuci piring%' OR title ILIKE '%bantu memasak%';

-- 2. Remove any previously seeded September antigravity tasks to allow re-run idempotently
DELETE FROM odyssey_tasks WHERE active_date BETWEEN DATE '2026-09-01' AND DATE '2026-09-30' AND title LIKE 'ANTIGRAVITY:%';

-- 3. Helper: seed tasks for each existing family (demo-crew-1 + any other)
--    We insert via SELECT from odyssey_families to ensure family scoping is correct.
--    Each row is templated; family_id is substituted per family.

DO $$
DECLARE
    fam RECORD;
BEGIN
    FOR fam IN SELECT id FROM odyssey_families LOOP
        -- === DAY 1 — 2026-09-01 — Mulai Siap Kerja (3 tasks) total 160 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Kenalan dengan CV — Apa yang Harus Ada di CV?', 'Tonton video 5 menit tentang CV (YouTube) lalu jawab quiz 5 pertanyaan. HP only. Pelajari bagian: data diri, pendidikan, pengalaman, skill.', 'QUIZ', 'AUTO', 1, DATE '2026-09-01', 90, 60, '{"youtube_url":"https://www.youtube.com/watch?v=example-cv-01","questions":[{"id":"q1","question":"Bagian CV yang berisi riwayat pendidikan disebut?","options":["A. Profil","B. Pendidikan","C. Hobi","D. Alamat"],"correct_answer":"B"},{"id":"q2","question":"Untuk lulusan SMA tanpa pengalaman, bagian yang paling penting adalah?","options":["A. Pengalaman kerja 10 tahun","B. Skill dan pendidikan","C. Foto liburan","D. Nilai rapor SD"],"correct_answer":"B"},{"id":"q3","question":"CV sebaiknya berapa halaman untuk pemula?","options":["A. 1-2 halaman","B. 10 halaman","C. 5 halaman penuh warna","D. Tidak perlu CV"],"correct_answer":"A"},{"id":"q4","question":"Yang TIDAK perlu ada di CV?","options":["A. Nama dan kontak","B. Pendidikan","C. Nomor PIN ATM","D. Skill"],"correct_answer":"C"},{"id":"q5","question":"Skill yang relevan untuk admin toko?","options":["A. Menguasai kasir & komunikasi","B. Jago main game","C. Tidur siang","D. Tidak ada skill"],"correct_answer":"A"}]}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Tulis Profil Singkat Tentang Dirimu', 'Bayangkan memperkenalkan diri ke perusahaan. Tulis 3–5 kalimat: siapa kamu, pendidikan, hal yang disukai, pekerjaan yang ingin dicoba. Kerjakan di HP.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-01', 55, 40, '{"prompt":"Tulis 3–5 kalimat tentang dirimu, pendidikanmu, hal yang kamu sukai, dan pekerjaan yang ingin kamu coba.","minimum_characters":80,"maximum_characters":2000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Memory Challenge — Skill ↔ Pekerjaan', 'Mainkan mini game mencocokkan skill dengan pekerjaan (2 menit). Skor minimal 60 untuk lulus.', 'MINI_GAME', 'AUTO', 3, DATE '2026-09-01', 15, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":60}'::jsonb, true);

        -- === DAY 2 — 2026-09-02 — total 150 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Bikin CV Sederhana di Google Docs (HP)', 'Download template Google Docs di HP, isi dengan data dirimu dari Day 1, lalu upload screenshot atau link. Tidak perlu laptop.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-09-02', 80, 55, '{"instruction":"Buka Google Docs di HP, buat CV 1 halaman sederhana, upload foto/screenshot hasil.","max_files":2}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Cari 1 Lowongan untuk Lulusan SMA', 'Cari di Jobstreet/Glints/Instagram lowongan yang menerima SMA. Tulis: posisi, perusahaan, lokasi, pendidikan minimal, 3 skill diminta + skill apa yang belum kamu punya?', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-02', 70, 45, '{"prompt":"Posisi, perusahaan, lokasi, pendidikan minimal, 3 skill diminta, dan skill apa yang ingin kamu pelajari.","minimum_characters":100,"maximum_characters":2000}'::jsonb, true);

        -- === DAY 3 — 2026-09-03 — total 90 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Simpan & Bagikan CV di Google Drive (HP)', 'Simpan CV dari Day 2 ke Google Drive, atur link bisa dilihat, tulis link dan jelaskan 2 langkah cara membagikan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-03', 70, 45, '{"prompt":"Tulis link Google Drive CV-mu dan 2 langkah cara membagikan file.","minimum_characters":60,"maximum_characters":1500}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Puzzle Kata — PROFESIONAL', 'Susun huruf acak menjadi kata terkait dunia kerja. Mini puzzle 2-3 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-03', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}'::jsonb, true);

        -- === DAY 4 — 2026-09-04 — total 130 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Kebutuhan vs Keinginan', 'Pilih mana kebutuhan vs keinginan dari 5 contoh belanja. Quiz 5 pertanyaan.', 'QUIZ', 'AUTO', 1, DATE '2026-09-04', 70, 45, '{"questions":[{"id":"q1","question":"Membeli beras untuk makan sehari-hari adalah?","options":["A. Kebutuhan","B. Keinginan","C. Tidak penting","D. Hiburan"],"correct_answer":"A"},{"id":"q2","question":"Membeli skin game seharga 200rb padahal uang pas-pasan adalah?","options":["A. Kebutuhan","B. Keinginan","C. Kewajiban","D. Investasi"],"correct_answer":"B"},{"id":"q3","question":"Bayar kos/transport untuk kerja termasuk?","options":["A. Kebutuhan","B. Keinginan","C. Hadiah","D. Bonus"],"correct_answer":"A"},{"id":"q4","question":"Nongkrong kopi tiap hari demi gaya adalah?","options":["A. Kebutuhan","B. Keinginan","C. Kebutuhan pokok","D. Wajib"],"correct_answer":"B"},{"id":"q5","question":"Menabung untuk dana darurat adalah?","options":["A. Keinginan","B. Kebutuhan masa depan","C. Pemborosan","D. Tidak perlu"],"correct_answer":"B"}]}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Catat Pengeluaran Hari Ini', 'Tulis minimal 5 pengeluaran hari ini (makan, transport, dll) + totalnya. Gunakan notes HP.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-04', 60, 40, '{"prompt":"Tulis minimal 5 pengeluaran hari ini dan totalnya.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true);

        -- === DAY 5 — 2026-09-05 — total 130 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Tulis Email Lamaran Singkat (HP)', 'Tulis email profesional 4–6 kalimat untuk melamar posisi dari Day 2. Perhatikan subjek, salam, isi, penutup.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-05', 80, 55, '{"prompt":"Tulis subjek + isi email lamaran 4–6 kalimat yang sopan.","minimum_characters":120,"maximum_characters":2000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Rapikan Nama File di HP', 'Rapikan 3 file di HP (beri nama jelas: CV_Nama.pdf, dsb). Upload foto sebelum & sesudah atau jelaskan perubahan nama.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 2, DATE '2026-09-05', 50, 35, '{"instruction":"Rapikan nama file di HP dan upload bukti foto/screenshot.","max_files":2}'::jsonb, true);

        -- === DAY 6 — 2026-09-06 — total 90 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Pahami Syarat Lowongan — Skill Gap', 'Dari lowongan Day 2, tulis 3 skill wajib + 1 skill yang belum kamu punya dan cara mempelajarinya.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-06', 70, 45, '{"prompt":"3 skill wajib + 1 skill yang belum kamu punya dan rencanamu mempelajarinya.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Reaction Tap — Uji Kecepatan', 'Mini game ketuk cepat untuk melatih fokus dan refleks. Skor minimal 50.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-06', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}'::jsonb, true);

        -- === DAY 7 — 2026-09-07 — Minggu Ringan — 60 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Waspada Hoax Lowongan', 'Ciri lowongan palsu: minta uang, gaji tidak wajar, tanpa alamat jelas. Quiz 4Q.', 'QUIZ', 'AUTO', 1, DATE '2026-09-07', 60, 40, '{"questions":[{"id":"q1","question":"Lowongan yang meminta transfer uang di awal adalah?","options":["A. Penipuan","B. Resmi","C. Wajar","D. Bonus"],"correct_answer":"A"},{"id":"q2","question":"Ciri lowongan terpercaya?","options":["A. Alamat jelas & proses gratis","B. Minta uang jaminan","C. Janji gaji fantastis tanpa kerja","D. Tanpa kontak"],"correct_answer":"A"},{"id":"q3","question":"Sebelum melamar sebaiknya?","options":["A. Cek alamat & ulasan perusahaan","B. Langsung transfer","C. Kirim KTP ke sembarang orang","D. Tidak perlu cek"],"correct_answer":"A"},{"id":"q4","question":"Jika ragu, kamu harus?","options":["A. Tanya admin / cari info resmi","B. Tetap bayar","C. Diam saja","D. Sebar hoax"],"correct_answer":"A"}]}'::jsonb, true);

        -- === DAY 8 — 2026-09-08 — 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Perbaiki Profil Diri (Revisi Day 1)', 'Perbaiki tulisan profil Day 1 setelah belajar CV & lowongan. Buat lebih meyakinkan dan sesuai pekerjaan incaran.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-08', 70, 45, '{"prompt":"Tulis ulang profil 3–5 kalimat yang lebih baik dan meyakinkan.","minimum_characters":80,"maximum_characters":2000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Formatting Dokumen Rapi di HP', 'Di Google Docs, praktik bold, bullet, dan rapikan CV. Upload screenshot hasil formatting.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 2, DATE '2026-09-08', 50, 35, '{"instruction":"Praktik bold/bullet di Docs dan upload screenshot.","max_files":2}'::jsonb, true);

        -- === DAY 9 — 2026-09-09 — 130 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Hitung Gaji Bersih', 'Gaji 2.500.000 - potongan 300.000 + lembur 200.000 = berapa? + 2 pertanyaan tunjangan. Quiz 4Q.', 'QUIZ', 'AUTO', 1, DATE '2026-09-09', 70, 45, '{"questions":[{"id":"q1","question":"2.500.000 - 300.000 + 200.000 = ?","options":["A. 2.400.000","B. 2.200.000","C. 2.000.000","D. 2.700.000"],"correct_answer":"A"},{"id":"q2","question":"Potongan BPJS termasuk?","options":["A. Potongan wajib","B. Bonus","C. Hadiah","D. Tidak ada"],"correct_answer":"A"},{"id":"q3","question":"Gaji bersih adalah gaji kotor dikurangi?","options":["A. Potongan","B. Bonus","C. Hadiah","D. THR"],"correct_answer":"A"},{"id":"q4","question":"Jika lembur tidak dibayar, gaji bersih akan?","options":["A. Lebih kecil","B. Lebih besar","C. Sama saja","D. Tidak terpengaruh"],"correct_answer":"A"}]}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Perkenalkan Diri 30 Detik', 'Tulis naskah perkenalan 30 detik untuk interview: nama, pendidikan, kelebihan, motivasi kerja.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-09', 60, 40, '{"prompt":"Tulis naskah perkenalan 30 detik yang percaya diri.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true);

        -- === DAY 10 — 2026-09-10 — 110 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Bikin Anggaran Bulanan di Google Sheets (HP)', 'Buat sheet sederhana: pemasukan, pengeluaran, sisa. Upload screenshot atau link sheet.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-09-10', 90, 60, '{"instruction":"Buat anggaran bulanan di Sheets (HP) dan upload bukti.","max_files":2}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Memory — Harga vs Barang', 'Cocokkan harga dengan barang yang benar. Game memori 2 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-10', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}'::jsonb, true);

        -- === DAY 11 — 2026-09-11 — 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Simulasi Diskon — Pilih Paling Hemat', '3 barang diskon 20%, 30%, 50%. Hitung harga akhir dan pilih paling hemat. Quiz 4Q.', 'QUIZ', 'AUTO', 1, DATE '2026-09-11', 70, 45, '{"questions":[{"id":"q1","question":"Barang 100.000 diskon 20% menjadi?","options":["A. 80.000","B. 90.000","C. 70.000","D. 100.000"],"correct_answer":"A"},{"id":"q2","question":"Barang 200.000 diskon 50% menjadi?","options":["A. 100.000","B. 150.000","C. 50.000","D. 200.000"],"correct_answer":"A"},{"id":"q3","question":"Lebih hemat: A 100k diskon 20% vs B 100k diskon 30%?","options":["A. B lebih hemat","B. A lebih hemat","C. Sama","D. Tidak ada diskon"],"correct_answer":"A"},{"id":"q4","question":"Diskon 10% dari 50.000 adalah?","options":["A. 5.000","B. 10.000","C. 15.000","D. 500"],"correct_answer":"A"}]}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Rencana Tabungan 1 Bulan', 'Tulis target tabungan 1 bulan: nominal, tujuan, cara mencapainya (kurangi jajan, dll).', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-11', 50, 35, '{"prompt":"Target tabungan, nominal, dan cara mencapainya.","minimum_characters":60,"maximum_characters":1500}'::jsonb, true);

        -- === DAY 12 — 2026-09-12 — 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Jawab Pertanyaan Interview Dasar', 'Latihan jawab: "Ceritakan tentang diri Anda" — tulis jawaban 4–6 kalimat yang meyakinkan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-12', 70, 45, '{"prompt":"Jawab pertanyaan interview dengan percaya diri 4–6 kalimat.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Bikin Poster Diri di Canva (HP)', 'Buka Canva di HP, bikin poster profil diri sederhana (nama, skill, cita-cita). Upload screenshot.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 2, DATE '2026-09-12', 50, 35, '{"instruction":"Buat poster diri di Canva HP dan upload bukti.","max_files":2}'::jsonb, true);

        -- === DAY 13 — 2026-09-13 — 100 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Balas Komplain Pelanggan dengan Sopan', 'Kamu CS toko online. Pelanggan komplain barang telat. Tulis balasan WA yang sopan, solusi, dan meminta maaf.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-13', 80, 50, '{"prompt":"Tulis balasan komplain yang sopan dan solutif.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Puzzle Prioritas Kerja', 'Urutkan 4 tugas kerja berdasarkan prioritas (deadline, penting). Mini game logic.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-13', 20, 15, '{"game":"MEMORY","difficulty":"MEDIUM","target_score":60}'::jsonb, true);

        -- === DAY 14 — 2026-09-14 — Minggu Ringan 60 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Rangkum Video Etika Kerja (3 poin)', 'Tonton video 3 menit tentang etika kerja (link YouTube), lalu rangkum 3 poin penting dengan bahasa sendiri.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-14', 60, 40, '{"prompt":"Rangkum 3 poin penting tentang etika kerja.","minimum_characters":60,"maximum_characters":1500}'::jsonb, true);

        -- === DAY 15 — 2026-09-15 — 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Bandingkan 2 Lowongan Sejenis', 'Cari 2 lowongan posisi sama beda perusahaan, bandingkan gaji, syarat, lokasi, dan pilih mana lebih cocok untukmu + alasan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-15', 70, 45, '{"prompt":"Bandingkan 2 lowongan dan jelaskan pilihanmu.","minimum_characters":100,"maximum_characters":2000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Latihan Kasir — Hitung Kembalian', 'Kembalian 50.000 - 17.500 = ? dan 2 soal kasir lain. Quiz 3Q.', 'QUIZ', 'AUTO', 2, DATE '2026-09-15', 50, 35, '{"questions":[{"id":"q1","question":"Kembalian 50.000 - 17.500 = ?","options":["A. 32.500","B. 33.500","C. 30.000","D. 35.000"],"correct_answer":"A"},{"id":"q2","question":"Jika barang 25.000 beli 2, total?","options":["A. 50.000","B. 25.000","C. 75.000","D. 100.000"],"correct_answer":"A"},{"id":"q3","question":"Kembalian 100.000 - 45.000 = ?","options":["A. 55.000","B. 45.000","C. 65.000","D. 50.000"],"correct_answer":"A"}]}'::jsonb, true);

        -- === DAY 16 — 2026-09-16 — 90 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Email Follow-up Lamaran', 'Tulis email follow-up sopan 3–5 hari setelah melamar: tanyakan status, ucapkan terima kasih.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-16', 70, 45, '{"prompt":"Tulis email follow-up lamaran yang sopan 3–5 kalimat.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Puzzle Jadwal Harian', 'Susun jadwal harian produktif dengan drag & drop. Game 2 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-16', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}'::jsonb, true);

        -- === DAY 17 — 2026-09-17 — 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Bersihkan Inbox Email (HP)', 'Buka Gmail di HP, hapus/archieve 10 email tidak penting, buat folder/label Lamaran. Upload screenshot.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-09-17', 60, 40, '{"instruction":"Bersihkan inbox Gmail dan upload screenshot folder Lamaran.","max_files":2}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Ikuti Instruksi SOP Packing', 'Baca 5 langkah SOP packing barang, lalu jawab quiz urutan yang benar.', 'QUIZ', 'AUTO', 2, DATE '2026-09-17', 60, 40, '{"questions":[{"id":"q1","question":"Langkah pertama packing yang benar?","options":["A. Cek barang & jumlah","B. Langsung kirim","C. Tulis alamat asal","D. Tutup kardus dulu"],"correct_answer":"A"},{"id":"q2","question":"Setelah cek barang, selanjutnya?","options":["A. Bungkus bubble wrap","B. Tidur","C. Buang barang","D. Foto selfie"],"correct_answer":"A"},{"id":"q3","question":"Langkah terakhir sebelum kirim?","options":["A. Tempel alamat & cek kembali","B. Buka kembali","C. Buang kardus","D. Tidak perlu cek"],"correct_answer":"A"}]}'::jsonb, true);

        -- === DAY 18 — 2026-09-18 — 130 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Budget 1 Minggu 500rb', 'Buat anggaran mingguan 500rb: makan, transport, tabungan, hiburan. Upload foto/sheet.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-09-18', 80, 55, '{"instruction":"Buat anggaran 1 minggu 500rb, upload bukti.","max_files":2}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Tulis Pesan WA ke HR', 'Tulis pesan WA profesional ke HR: perkenalkan diri, tanya status lamaran, ucapkan terima kasih.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-18', 50, 35, '{"prompt":"Pesan WA ke HR yang sopan dan profesional.","minimum_characters":60,"maximum_characters":1000}'::jsonb, true);

        -- === DAY 19 — 2026-09-19 — 130 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Simulasi Interview 3 Pertanyaan Sulit', 'Jawab: kelebihan, kekurangan, dan kenapa ingin bekerja di sini. Quiz pilihan ganda situasional.', 'QUIZ', 'AUTO', 1, DATE '2026-09-19', 80, 50, '{"questions":[{"id":"q1","question":"Saat ditanya kekurangan, jawaban terbaik?","options":["A. Jujur + tunjukkan usaha perbaikan","B. Saya tidak punya kekurangan","C. Malas dan sering telat","D. Tidak jawab"],"correct_answer":"A"},{"id":"q2","question":"Kenapa ingin bekerja di sini? Jawab?","options":["A. Sesuai skill & ingin belajar","B. Karena dekat rumah saja","C. Iseng","D. Tidak tahu"],"correct_answer":"A"},{"id":"q3","question":"Ceritakan kelebihanmu?","options":["A. Sebutkan skill relevan + contoh","B. Saya sempurna","C. Tidak ada kelebihan","D. Diam"],"correct_answer":"A"}]}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Cari Cara Buat NPWP Online (Rangkum 3 Langkah)', 'Riset di HP cara daftar NPWP online, rangkum 3 langkah penting dengan bahasamu.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-19', 50, 35, '{"prompt":"Rangkum 3 langkah daftar NPWP online.","minimum_characters":60,"maximum_characters":1500}'::jsonb, true);

        -- === DAY 20 — 2026-09-20 — 110 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Menabung vs Cicilan', 'Pilih: cicilan 12x bunga vs nabung 6 bulan. Quiz hitung bunga cicilan.', 'QUIZ', 'AUTO', 1, DATE '2026-09-20', 90, 60, '{"questions":[{"id":"q1","question":"Cicilan 1.200.000 bunga 10% total jadi?","options":["A. 1.320.000","B. 1.200.000","C. 1.100.000","D. 1.500.000"],"correct_answer":"A"},{"id":"q2","question":"Lebih hemat jika tidak urgent?","options":["A. Nabung dulu","B. Cicilan bunga tinggi","C. Pinjam teman bunga tinggi","D. Tidak bayar"],"correct_answer":"A"},{"id":"q3","question":"Dana darurat ideal berapa bulan pengeluaran?","options":["A. 3-6 bulan","B. 1 hari","C. Tidak perlu","D. 1 tahun penuh gaji"],"correct_answer":"A"}]}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Memory Challenge Lv2', 'Tingkat kesulitan naik — cocokkan 8 pasang kartu. Target 70.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-20', 20, 15, '{"game":"MEMORY","difficulty":"MEDIUM","target_score":70}'::jsonb, true);

        -- === DAY 21 — 2026-09-21 — Review 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Rangkuman Minggu 1–3', 'Tulis 5 hal penting yang kamu pelajari selama 3 minggu: CV, digital, finansial, komunikasi.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-21', 100, 65, '{"prompt":"Tulis 5 hal penting yang kamu pelajari.","minimum_characters":100,"maximum_characters":2000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Skor Challenge', 'Kumpulkan poin tertinggi di mini game. Target 70.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-21', 20, 15, '{"game":"MEMORY","difficulty":"MEDIUM","target_score":70}'::jsonb, true);

        -- === DAY 22 — 2026-09-22 — Final CV 130 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Final CV — Upload CV Lengkap', 'Upload CV final versi terbaik (PDF/DOCX) — gabungan perbaikan dari Day 8 & 1. Ini milestone utama!', 'DOCUMENT_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-09-22', 90, 60, '{"instruction":"Upload CV final terbaikmu (PDF/DOCX).","attachment_name":"CV Final","max_file_size_mb":5}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Naskah Perkenalan 60 Detik', 'Tulis naskah perkenalan 60 detik untuk interview/video profil: sapa, latar, skill, tujuan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-22', 40, 30, '{"prompt":"Naskah perkenalan 60 detik yang percaya diri.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true);

        -- === DAY 23 — 2026-09-23 — Portfolio 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Portfolio HP — Kumpulkan Karya di Drive', 'Kumpulkan link/file karya (CV, poster Canva, sheet anggaran) ke 1 folder Drive. Tulis link folder.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-23', 70, 45, '{"prompt":"Tulis link folder Drive portfolio dan isi 3 file di dalamnya.","minimum_characters":60,"maximum_characters":1500}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Rencana Keuangan Gaji Pertama', 'Gaji pertama 3jt — alokasikan: kebutuhan, tabungan, orang tua, pengembangan diri. Tulis persentasenya.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-23', 50, 35, '{"prompt":"Alokasi gaji pertama 100% dengan persentase.","minimum_characters":60,"maximum_characters":1500}'::jsonb, true);

        -- === DAY 24 — 2026-09-24 — PAYDAY 130 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Refleksi 24 Hari — Skill yang Ingin Dilanjutkan', 'Tulis refleksi: 3 skill yang paling berguna, 1 yang ingin kamu dalami setelah program, dan langkah selanjutnya.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-24', 80, 55, '{"prompt":"3 skill berguna, 1 yang ingin didalami, langkah selanjutnya.","minimum_characters":100,"maximum_characters":2000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Celebration Mini Game', 'Rayakan perjalanan! Mainkan game perayaan — tidak perlu target tinggi, have fun!', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-24', 50, 30, '{"game":"MEMORY","difficulty":"EASY","target_score":40}'::jsonb, true);

        -- === DAY 25 — 2026-09-25 — 120 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Rencana Karir 3 Bulan', 'Tulis target 3 bulan ke depan: posisi impian, skill yang harus dipelajari, dan langkah konkrit.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-25', 70, 45, '{"prompt":"Target karir 3 bulan dan langkah konkritnya.","minimum_characters":80,"maximum_characters":1500}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Evaluasi Portofolio', 'Periksa kembali file portofolio di Drive, pastikan semua izin akses sudah umum.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-25', 50, 30, '{"prompt":"Tulis ringkasan link portofolio dan kelengkapannya.","minimum_characters":50,"maximum_characters":1000}'::jsonb, true);

        -- === DAY 26 — 2026-09-26 — 80 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Simulasi Presentasi Diri', 'Buat rangkuman singkat perkenalan 1 menit untuk calon atasan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-26', 60, 40, '{"prompt":"Rangkuman perkenalan 1 menit yang efektif.","minimum_characters":60,"maximum_characters":1000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Mini Game Fokus Refleks', 'Uji ketangkasan dan fokus dengan mini game memori.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-26', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}'::jsonb, true);

        -- === DAY 27 — 2026-09-27 — 100 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Etika Berkomunikasi di Tempat Kerja', 'Quiz 4Q tentang komunikasi profesional via chat, email, dan tatap muka.', 'QUIZ', 'AUTO', 1, DATE '2026-09-27', 70, 45, '{"questions":[{"id":"q1","question":"Sikap terbaik saat menerima masukan dari atasan?","options":["A. Mendengarkan dan belajar","B. Marah","C. Acuh","D. Keluar kerja"],"correct_answer":"A"}]}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Catat Pembelajaran Penting', 'Tulis 3 prinsip kerja profesional yang paling ingin kamu terapkan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-09-27', 30, 20, '{"prompt":"3 prinsip kerja profesional pilihanmu.","minimum_characters":50,"maximum_characters":1000}'::jsonb, true);

        -- === DAY 28 — 2026-09-28 — 70 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Persiapan Menghadapi Hari Pertama Kerja', 'Rangkum 5 hal yang harus disiapkan sebelum hari pertama kerja.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-28', 50, 35, '{"prompt":"5 hal persiapan hari pertama kerja.","minimum_characters":60,"maximum_characters":1000}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Memory Refresh Challenge', 'Latihan daya ingat mini game 2 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-28', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}'::jsonb, true);

        -- === DAY 29 — 2026-09-29 — 60 ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Checklist Dokumen Lamaran', 'Pastikan KTP, CV, dan berkas penting siap dikirim kapan saja.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-29', 40, 30, '{"prompt":"Tulis checklist berkas yang sudah siap.","minimum_characters":40,"maximum_characters":800}'::jsonb, true),
        (fam.id, 'ANTIGRAVITY: Mini Game Kelulusan', 'Mainkan mini game santai penutup periode.', 'MINI_GAME', 'AUTO', 2, DATE '2026-09-29', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":40}'::jsonb, true);

        -- === DAY 30 — 2026-09-30 — 0 coins (Graduation & Final Badging) ===
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
        VALUES
        (fam.id, 'ANTIGRAVITY: Kelulusan Periode — Selamat!', 'Kamu berhasil menyelesaikan program 30 hari! Tulis kesan dan pesan akhirmu.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-09-30', 0, 100, '{"prompt":"Tulis kesan dan pesan akhirmu setelah 30 hari.","minimum_characters":50,"maximum_characters":1500}'::jsonb, true);

    END LOOP;
END $$;

-- 4. Verify total coins per family (target ~3200 coins per 30-day period)
DO $$
DECLARE
    v_total INT;
BEGIN
    SELECT SUM(reward_coins) INTO v_total FROM odyssey_tasks WHERE title LIKE 'ANTIGRAVITY:%' AND active_date BETWEEN DATE '2026-09-01' AND DATE '2026-09-30';
    IF v_total IS NOT NULL THEN
        RAISE NOTICE 'Antigravity total coins (all families combined): % (per family ~ %)', v_total, v_total / GREATEST((SELECT COUNT(*) FROM odyssey_families),1);
    END IF;
END $$;

INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '052_seed_antigravity_september_2026')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
