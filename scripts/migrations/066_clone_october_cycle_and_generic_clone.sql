-- ============================================================
-- Migration 066: Generic clone + October 2026 seeds
-- DEFERRED CUTOVER: First rolling cycle is 2026-10-25 -> 2026-11-25 (31 days)
-- Keeps Oct1-24 for current 1-24 October period before cutover (legacy)
-- Prepares Oct25-31 (7 days, 510) + Nov1-24 (24 days, 2690) = 3200 for first rolling cycle
-- Business-valid: Oct25-31 uses safe recurring content (no graduation/Day24 milestone)
-- Idempotent, per-family via CROSS JOIN, never deletes history
-- ============================================================

-- 1. Generic idempotent clone helper (for Oct1-24 and Nov1-24)
CREATE OR REPLACE FUNCTION odyssey_clone_cycle(
    p_source_start DATE,
    p_source_end DATE,
    p_target_start DATE
) RETURNS INT
LANGUAGE plpgsql SECURITY DEFINER SET search_path = public AS $$
DECLARE v_offset INT; v_inserted INT;
BEGIN
    IF p_source_start IS NULL OR p_source_end IS NULL OR p_target_start IS NULL THEN
        RAISE EXCEPTION 'clone_cycle: source/target dates required' USING ERRCODE='P0005';
    END IF;
    IF p_source_start >= p_source_end THEN
        RAISE EXCEPTION 'clone_cycle: source_start must be < source_end' USING ERRCODE='P0005';
    END IF;
    v_offset := (p_target_start - p_source_start);
    WITH inserted AS (
        INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active, created_by, target_scope, target_user_uid)
        SELECT s.family_id, s.title, s.description, s.task_type, s.evaluation_type, s.step_order, s.active_date + v_offset, s.reward_coins, s.reward_xp, s.config, s.is_active, s.created_by, s.target_scope, s.target_user_uid
        FROM odyssey_tasks s
        WHERE s.active_date >= p_source_start
          AND s.active_date < p_source_end
          AND s.is_active = true
          AND NOT EXISTS (
              SELECT 1 FROM odyssey_tasks t
              WHERE t.family_id = s.family_id
                AND t.active_date = s.active_date + v_offset
                AND t.step_order = s.step_order
                AND COALESCE(t.target_scope,'ALL') = COALESCE(s.target_scope,'ALL')
                AND COALESCE(t.target_user_uid,'') = COALESCE(s.target_user_uid,'')
                AND t.title = s.title
          )
        RETURNING 1
    )
    SELECT COUNT(*) INTO v_inserted FROM inserted;
    RETURN v_inserted;
END; $$;
REVOKE ALL ON FUNCTION odyssey_clone_cycle(DATE,DATE,DATE) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_clone_cycle(DATE,DATE,DATE) TO service_role;

-- 2. Seed October 1-24 2026 for CURRENT 1-24 October period before cutover (legacy)
-- Source Sep1-24 (46 tasks, 2690) -> Oct1-24, per-family via clone
SELECT odyssey_clone_cycle('2026-09-01'::DATE, '2026-09-25'::DATE, '2026-10-01'::DATE);

-- 3. Seed first rolling cycle Oct25-Nov24 (31 days, 3200) — explicit per-family VALUES
-- Oct25-31: 7 days, 510 total (safe recurring, no milestone/graduation)
-- Nov1-24: 24 days, 2690 total (cloned from Sep1-24 via explicit VALUES for per-family guarantee)
-- We use CROSS JOIN odyssey_families to ensure EVERY active family gets complete schedule,
-- not just families that had September source rows (fixes per-family blocker).

-- 3a. Oct25-31 explicit (7 days, 14 tasks, 510)
INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
SELECT fam.id, v.title, v.description, v.task_type, v.evaluation_type, v.step_order, v.active_date, v.reward_coins, v.reward_xp, v.config::jsonb, v.is_active
FROM odyssey_families fam
CROSS JOIN (VALUES
    -- Oct25 (80 = 50+30)
    ('ANTIGRAVITY: Rencana Karir Mingguan — Evaluasi Progress', 'Tulis 3 pencapaian minggu ini dan 1 hal yang ingin ditingkatkan minggu depan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-10-25', 50, 30, '{"prompt":"3 pencapaian dan 1 hal yang ingin ditingkatkan","minimum_characters":60,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Mini Game Fokus Harian', 'Latihan fokus 2 menit, target 50.', 'MINI_GAME', 'AUTO', 2, DATE '2026-10-25', 30, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Oct26 (80 = 50+30)
    ('ANTIGRAVITY: Simulasi Presentasi Diri — Latihan Percaya Diri', 'Tulis naskah perkenalan 1 menit untuk calon atasan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-10-26', 50, 30, '{"prompt":"Naskah perkenalan 1 menit","minimum_characters":60,"maximum_characters":1000}', true),
    ('ANTIGRAVITY: Puzzle Prioritas Harian', 'Urutkan 4 tugas harian berdasarkan urgensi.', 'MINI_GAME', 'AUTO', 2, DATE '2026-10-26', 30, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Oct27 (80 = 50+30)
    ('ANTIGRAVITY: Etika Komunikasi — Chat Profesional', 'Quiz 4 pertanyaan tentang komunikasi kerja via chat/email.', 'QUIZ', 'AUTO', 1, DATE '2026-10-27', 50, 35, '{"questions":[{"id":"q1","question":"Balasan chat profesional sebaiknya?","options":["A. Singkat sopan","B. Mengabaikan","C. Marah","D. Spam"],"correct_answer":"A"}]}', true),
    ('ANTIGRAVITY: Catat Pembelajaran Hari Ini', 'Tulis 3 hal baru yang dipelajari hari ini.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-10-27', 30, 20, '{"prompt":"3 hal baru yang dipelajari","minimum_characters":50,"maximum_characters":1000}', true),
    -- Oct28 (70 = 40+30)
    ('ANTIGRAVITY: Persiapan Hari Kerja — Checklist', 'Tulis 5 hal yang disiapkan sebelum hari kerja.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-10-28', 40, 25, '{"prompt":"5 hal persiapan hari kerja","minimum_characters":60,"maximum_characters":1000}', true),
    ('ANTIGRAVITY: Memory Harian — Latihan Ingatan', 'Mini game 2 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-10-28', 30, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Oct29 (70 = 40+30)
    ('ANTIGRAVITY: Simulasi Anggaran Harian', 'Buat anggaran harian sederhana dan upload bukti.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-10-29', 40, 25, '{"instruction":"Buat anggaran harian dan upload foto","max_files":2}', true),
    ('ANTIGRAVITY: Mini Game Ketangkasan', 'Uji reflek 2 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-10-29', 30, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Oct30 (70 = 40+30)
    ('ANTIGRAVITY: Komunikasi Efektif — Studi Kasus', 'Tulis cara menyampaikan ide dengan jelas dalam rapat.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-10-30', 40, 25, '{"prompt":"Cara menyampaikan ide dengan jelas","minimum_characters":60,"maximum_characters":1000}', true),
    ('ANTIGRAVITY: Puzzle Konsentrasi', 'Susun puzzle 2 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-10-30', 30, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Oct31 (60 = 30+30)
    ('ANTIGRAVITY: Evaluasi Mingguan — Apa yang Berjalan Baik?', 'Tulis 2 hal yang berjalan baik minggu ini dan 1 yang perlu diperbaiki.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-10-31', 30, 20, '{"prompt":"2 hal baik dan 1 perbaikan","minimum_characters":50,"maximum_characters":1000}', true),
    ('ANTIGRAVITY: Mini Game Refleksi Mingguan', 'Mini game santai penutup minggu.', 'MINI_GAME', 'AUTO', 2, DATE '2026-10-31', 30, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":40}', true)
) AS v(title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
WHERE NOT EXISTS (
    SELECT 1 FROM odyssey_tasks t2
    WHERE t2.family_id = fam.id AND t2.active_date = v.active_date AND t2.step_order = v.step_order AND t2.title = v.title
);

-- 3b. Nov1-24 explicit (24 days, 48 tasks, 2690) — mirrors Sep1-24 safe pattern, per-family via CROSS JOIN
-- Weights per day identical to Sep1-24 to preserve proven 2690 distribution, titles dated for November
INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
SELECT fam.id, v.title, v.description, v.task_type, v.evaluation_type, v.step_order, v.active_date, v.reward_coins, v.reward_xp, v.config::jsonb, v.is_active
FROM odyssey_families fam
CROSS JOIN (VALUES
    -- Nov1 (160 =90+55+15) mirrors Sep1
    ('ANTIGRAVITY: Pengenalan CV — Dasar untuk Pemula (Nov)', 'Tonton video CV dan jawab quiz.', 'QUIZ', 'AUTO', 1, DATE '2026-11-01', 90, 60, '{"youtube_url":"https://www.youtube.com/watch?v=example-cv-01","questions":[{"id":"q1","question":"Bagian CV pendidikan?","options":["A. Profil","B. Pendidikan","C. Hobi","D. Alamat"],"correct_answer":"B"}]}', true),
    ('ANTIGRAVITY: Profil Diri Singkat (Nov)', 'Tulis 3-5 kalimat tentang diri.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-01', 55, 40, '{"prompt":"3-5 kalimat tentang diri","minimum_characters":80,"maximum_characters":2000}', true),
    ('ANTIGRAVITY: Memory Skill-Pekerjaan (Nov)', 'Mini game 2 menit target 60.', 'MINI_GAME', 'AUTO', 3, DATE '2026-11-01', 15, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":60}', true),
    -- Nov2 (150 =80+70)
    ('ANTIGRAVITY: CV Sederhana di HP (Nov)', 'Buat CV 1 halaman di Docs.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-11-02', 80, 55, '{"instruction":"Buat CV di Docs HP","max_files":2}', true),
    ('ANTIGRAVITY: Cari Lowongan SMA (Nov)', 'Cari lowongan SMA dan tulis detail.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-02', 70, 45, '{"prompt":"Posisi, perusahaan, lokasi","minimum_characters":100,"maximum_characters":2000}', true),
    -- Nov3 (90 =70+20)
    ('ANTIGRAVITY: Simpan CV di Drive (Nov)', 'Simpan CV ke Drive dan tulis link.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-03', 70, 45, '{"prompt":"Link Drive dan 2 langkah","minimum_characters":60,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Puzzle Profesional (Nov)', 'Puzzle kata 2 menit.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-03', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Nov4 (130 =70+60)
    ('ANTIGRAVITY: Kebutuhan vs Keinginan (Nov)', 'Quiz 5 pertanyaan.', 'QUIZ', 'AUTO', 1, DATE '2026-11-04', 70, 45, '{"questions":[{"id":"q1","question":"Beras adalah?","options":["A. Kebutuhan","B. Keinginan","C. Tidak penting","D. Hiburan"],"correct_answer":"A"}]}', true),
    ('ANTIGRAVITY: Catat Pengeluaran (Nov)', 'Tulis 5 pengeluaran hari ini.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-04', 60, 40, '{"prompt":"5 pengeluaran dan total","minimum_characters":80,"maximum_characters":1500}', true),
    -- Nov5 (130 =80+50)
    ('ANTIGRAVITY: Email Lamaran Singkat (Nov)', 'Tulis email lamaran 4-6 kalimat.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-05', 80, 55, '{"prompt":"Email lamaran sopan","minimum_characters":120,"maximum_characters":2000}', true),
    ('ANTIGRAVITY: Rapikan File di HP (Nov)', 'Rapikan 3 file dan upload bukti.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 2, DATE '2026-11-05', 50, 35, '{"instruction":"Rapikan file HP","max_files":2}', true),
    -- Nov6 (90 =70+20)
    ('ANTIGRAVITY: Skill Gap — Syarat Lowongan (Nov)', 'Tulis 3 skill wajib + cara belajar.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-06', 70, 45, '{"prompt":"3 skill wajib","minimum_characters":80,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Reaction Tap (Nov)', 'Mini game target 50.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-06', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Nov7 (60)
    ('ANTIGRAVITY: Waspada Hoax Lowongan (Nov)', 'Quiz 4Q hoax.', 'QUIZ', 'AUTO', 1, DATE '2026-11-07', 60, 40, '{"questions":[{"id":"q1","question":"Minta transfer adalah?","options":["A. Penipuan","B. Resmi","C. Wajar","D. Bonus"],"correct_answer":"A"}]}', true),
    -- Nov8 (120 =70+50)
    ('ANTIGRAVITY: Perbaiki Profil Diri (Nov)', 'Perbaiki profil Day1.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-08', 70, 45, '{"prompt":"Profil 3-5 kalimat","minimum_characters":80,"maximum_characters":2000}', true),
    ('ANTIGRAVITY: Formatting Dokumen (Nov)', 'Praktik bold/bullet.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 2, DATE '2026-11-08', 50, 35, '{"instruction":"Formatting Docs","max_files":2}', true),
    -- Nov9 (130 =70+60)
    ('ANTIGRAVITY: Hitung Gaji Bersih (Nov)', 'Quiz gaji.', 'QUIZ', 'AUTO', 1, DATE '2026-11-09', 70, 45, '{"questions":[{"id":"q1","question":"2.500.000-300.000+200.000=?","options":["A. 2.400.000","B. 2.200.000","C. 2.000.000","D. 2.700.000"],"correct_answer":"A"}]}', true),
    ('ANTIGRAVITY: Perkenalan 30 Detik (Nov)', 'Naskah perkenalan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-09', 60, 40, '{"prompt":"Naskah 30 detik","minimum_characters":80,"maximum_characters":1500}', true),
    -- Nov10 (110 =90+20)
    ('ANTIGRAVITY: Anggaran Bulanan Sheets (Nov)', 'Buat anggaran di Sheets.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-11-10', 90, 60, '{"instruction":"Anggaran Sheets","max_files":2}', true),
    ('ANTIGRAVITY: Memory Harga-Barang (Nov)', 'Game memori.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-10', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Nov11 (120 =70+50)
    ('ANTIGRAVITY: Simulasi Diskon (Nov)', 'Quiz diskon.', 'QUIZ', 'AUTO', 1, DATE '2026-11-11', 70, 45, '{"questions":[{"id":"q1","question":"100.000 diskon 20%=?","options":["A. 80.000","B. 90.000","C. 70.000","D. 100.000"],"correct_answer":"A"}]}', true),
    ('ANTIGRAVITY: Rencana Tabungan (Nov)', 'Target tabungan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-11', 50, 35, '{"prompt":"Target tabungan","minimum_characters":60,"maximum_characters":1500}', true),
    -- Nov12 (120 =70+50)
    ('ANTIGRAVITY: Interview Dasar (Nov)', 'Jawab interview.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-12', 70, 45, '{"prompt":"Jawab interview 4-6 kalimat","minimum_characters":80,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Poster Diri Canva (Nov)', 'Poster di Canva.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 2, DATE '2026-11-12', 50, 35, '{"instruction":"Poster Canva","max_files":2}', true),
    -- Nov13 (100 =80+20)
    ('ANTIGRAVITY: Balas Komplain Sopan (Nov)', 'Tulis balasan WA.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-13', 80, 50, '{"prompt":"Balasan komplain","minimum_characters":80,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Puzzle Prioritas (Nov)', 'Puzzle prioritas.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-13', 20, 15, '{"game":"MEMORY","difficulty":"MEDIUM","target_score":60}', true),
    -- Nov14 (60)
    ('ANTIGRAVITY: Etika Kerja Video (Nov)', 'Rangkum 3 poin etika.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-14', 60, 40, '{"prompt":"3 poin etika","minimum_characters":60,"maximum_characters":1500}', true),
    -- Nov15 (120 =70+50)
    ('ANTIGRAVITY: Bandingkan Lowongan (Nov)', 'Bandingkan 2 lowongan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-15', 70, 45, '{"prompt":"Bandingkan lowongan","minimum_characters":100,"maximum_characters":2000}', true),
    ('ANTIGRAVITY: Latihan Kasir (Nov)', 'Quiz kembalian.', 'QUIZ', 'AUTO', 2, DATE '2026-11-15', 50, 35, '{"questions":[{"id":"q1","question":"Kembalian 50.000-17.500=?","options":["A. 32.500","B. 33.500","C. 30.000","D. 35.000"],"correct_answer":"A"}]}', true),
    -- Nov16 (90 =70+20)
    ('ANTIGRAVITY: Email Follow-up (Nov)', 'Email follow-up.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-16', 70, 45, '{"prompt":"Email follow-up","minimum_characters":80,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Puzzle Jadwal (Nov)', 'Puzzle jadwal.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-16', 20, 15, '{"game":"MEMORY","difficulty":"EASY","target_score":50}', true),
    -- Nov17 (120 =60+60)
    ('ANTIGRAVITY: Bersihkan Inbox (Nov)', 'Bersihkan Gmail.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-11-17', 60, 40, '{"instruction":"Bersihkan inbox","max_files":2}', true),
    ('ANTIGRAVITY: SOP Packing (Nov)', 'Quiz SOP.', 'QUIZ', 'AUTO', 2, DATE '2026-11-17', 60, 40, '{"questions":[{"id":"q1","question":"Langkah pertama packing?","options":["A. Cek barang","B. Kirim","C. Alamat","D. Tutup"],"correct_answer":"A"}]}', true),
    -- Nov18 (130 =80+50)
    ('ANTIGRAVITY: Budget 1 Minggu 500rb (Nov)', 'Anggaran mingguan.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-11-18', 80, 55, '{"instruction":"Anggaran 500rb","max_files":2}', true),
    ('ANTIGRAVITY: Pesan WA HR (Nov)', 'Pesan WA HR.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-18', 50, 35, '{"prompt":"Pesan WA HR","minimum_characters":60,"maximum_characters":1000}', true),
    -- Nov19 (130 =80+50)
    ('ANTIGRAVITY: Interview Sulit (Nov)', 'Quiz interview.', 'QUIZ', 'AUTO', 1, DATE '2026-11-19', 80, 50, '{"questions":[{"id":"q1","question":"Kekurangan dijawab?","options":["A. Jujur+perbaikan","B. Tidak punya","C. Malas","D. Diam"],"correct_answer":"A"}]}', true),
    ('ANTIGRAVITY: Cara Buat NPWP (Nov)', 'Rangkum 3 langkah NPWP.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-19', 50, 35, '{"prompt":"3 langkah NPWP","minimum_characters":60,"maximum_characters":1500}', true),
    -- Nov20 (110 =90+20)
    ('ANTIGRAVITY: Menabung vs Cicilan (Nov)', 'Quiz cicilan.', 'QUIZ', 'AUTO', 1, DATE '2026-11-20', 90, 60, '{"questions":[{"id":"q1","question":"1.200.000 bunga 10%=?","options":["A. 1.320.000","B. 1.200.000","C. 1.100.000","D. 1.500.000"],"correct_answer":"A"}]}', true),
    ('ANTIGRAVITY: Memory Lv2 (Nov)', 'Memory medium.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-20', 20, 15, '{"game":"MEMORY","difficulty":"MEDIUM","target_score":70}', true),
    -- Nov21 (120 =100+20) — safe generic, not final reflection
    ('ANTIGRAVITY: Evaluasi Skill Mingguan (Nov)', 'Tulis 3 skill yang meningkat minggu ini.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-21', 100, 55, '{"prompt":"3 skill meningkat","minimum_characters":80,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Mini Game Mingguan (Nov)', 'Game target 70.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-21', 20, 15, '{"game":"MEMORY","difficulty":"MEDIUM","target_score":70}', true),
    -- Nov22 (130 =90+40) — generic, not Final CV milestone
    ('ANTIGRAVITY: Latihan Dokumen Profesional (Nov)', 'Upload contoh dokumen rapi (bukan Final CV).', 'DOCUMENT_UPLOAD', 'ADMIN_REVIEW', 1, DATE '2026-11-22', 90, 50, '{"instruction":"Upload dokumen rapi","max_file_size_mb":5}', true),
    ('ANTIGRAVITY: Latihan Perkenalan 60 Detik (Nov)', 'Naskah perkenalan.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-22', 40, 30, '{"prompt":"Naskah 60 detik","minimum_characters":80,"maximum_characters":1500}', true),
    -- Nov23 (70 =40+30) — generic portfolio, not final portfolio milestone
    ('ANTIGRAVITY: Koleksi Karya — Organisasi Drive (Nov)', 'Kumpulkan link karya ke folder Drive.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-23', 20, 25, '{"prompt":"Link folder Drive","minimum_characters":60,"maximum_characters":1500}', true),
    ('ANTIGRAVITY: Simulasi Gaji Pertama (Nov)', 'Alokasi gaji 3jt.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, DATE '2026-11-23', 20, 20, '{"prompt":"Alokasi gaji","minimum_characters":60,"maximum_characters":1500}', true),
    -- Nov24 (130 =80+50) — generic reflection, not Day24 final celebration
    ('ANTIGRAVITY: Refleksi Harian — Pelajaran Hari Ini (Nov)', 'Tulis 2 pelajaran hari ini dan rencana besok.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 1, DATE '2026-11-24', 80, 40, '{"prompt":"2 pelajaran dan rencana","minimum_characters":60,"maximum_characters":1000}', true),
    ('ANTIGRAVITY: Mini Game Harian (Nov)', 'Game santai.', 'MINI_GAME', 'AUTO', 2, DATE '2026-11-24', 50, 30, '{"game":"MEMORY","difficulty":"EASY","target_score":40}', true)
) AS v(title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
WHERE NOT EXISTS (
    SELECT 1 FROM odyssey_tasks t2
    WHERE t2.family_id = fam.id AND t2.active_date = v.active_date AND t2.step_order = v.step_order AND t2.title = v.title
);

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','066_clone_october_cycle_and_generic_clone')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
