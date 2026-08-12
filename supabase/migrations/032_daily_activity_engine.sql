-- 032_daily_activity_engine.sql

-- odyssey_daily_activities
CREATE TABLE IF NOT EXISTS odyssey_daily_activities (
    id BIGSERIAL PRIMARY KEY,
    slug TEXT UNIQUE NOT NULL,
    title TEXT NOT NULL,
    question TEXT NOT NULL,
    type TEXT NOT NULL, -- 'MCQ' or 'TRUE_FALSE'
    options JSONB NOT NULL DEFAULT '[]'::jsonb,
    correct_answer TEXT NOT NULL,
    explanation TEXT NOT NULL,
    xp_reward INTEGER NOT NULL DEFAULT 10,
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- odyssey_daily_activity_completions
CREATE TABLE IF NOT EXISTS odyssey_daily_activity_completions (
    id BIGSERIAL PRIMARY KEY,
    user_id TEXT NOT NULL,
    activity_date DATE NOT NULL,
    activity_id BIGINT NOT NULL REFERENCES odyssey_daily_activities(id),
    completed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Constraint to ensure a user can only complete one activity per date
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_daily_activity_completions_uid_date 
ON odyssey_daily_activity_completions(user_id, activity_date);

-- RLS for odyssey_daily_activities
ALTER TABLE odyssey_daily_activities ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow public read access on daily activities" ON odyssey_daily_activities;
CREATE POLICY "Allow public read access on daily activities" ON odyssey_daily_activities FOR SELECT USING (true);
DROP POLICY IF EXISTS "Allow service_role full access on daily activities" ON odyssey_daily_activities;
CREATE POLICY "Allow service_role full access on daily activities" ON odyssey_daily_activities FOR ALL USING (true);

-- RLS for odyssey_daily_activity_completions
ALTER TABLE odyssey_daily_activity_completions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on completions" ON odyssey_daily_activity_completions;
CREATE POLICY "Allow service_role full access on completions" ON odyssey_daily_activity_completions FOR ALL USING (true);

-- Seed data for testing (14 activities)
INSERT INTO odyssey_daily_activities (slug, title, question, type, options, correct_answer, explanation, xp_reward, active) VALUES
('safety-1', 'Digital Safety: OTP', 'Bank menelepon dan meminta OTP. Apa yang harus kamu lakukan?', 'MCQ', '["Memberikan OTP", "Tutup telepon segera", "Tanya alasan mereka", "Minta mereka mengirim ulang"]', 'Tutup telepon segera', 'OTP adalah rahasia pribadi yang tidak boleh dibagikan kepada siapapun, bahkan pihak bank.', 10, true),
('finance-1', 'Finance: Menabung', 'Kamu memiliki Rp100.000 dan ingin menabung Rp20.000. Berapa yang boleh kamu gunakan?', 'MCQ', '["Rp100.000", "Rp80.000", "Rp20.000", "Rp120.000"]', 'Rp80.000', 'Sisihkan uang untuk ditabung di awal, lalu gunakan sisanya.', 10, true),
('career-1', 'Career: Hubungi Recruiter', 'Mana yang lebih profesional untuk menghubungi recruiter?', 'MCQ', '["Halo kak, ada loker?", "Selamat pagi, saya tertarik dengan lowongan yang dibuka...", "P, info loker", "Saya butuh kerjaan nih"]', 'Selamat pagi, saya tertarik dengan lowongan yang dibuka...', 'Selalu gunakan bahasa yang sopan dan profesional saat menghubungi recruiter.', 10, true),
('prod-1', 'Productivity: Prioritas', 'Kamu punya tiga tugas. Mana yang sebaiknya dikerjakan lebih dulu?', 'MCQ', '["Tugas paling mudah", "Tugas yang paling penting dan mendesak", "Tugas yang disukai saja", "Dikerjakan bersamaan semua"]', 'Tugas yang paling penting dan mendesak', 'Fokuslah pada hal yang paling penting dan memiliki tenggat waktu terdekat (Kuadran Eisenhower).', 10, true),
('comm-1', 'Communication: Klarifikasi', 'Apa respons yang paling baik ketika kamu tidak memahami instruksi seseorang?', 'MCQ', '["Diam saja", "Pura-pura paham", "Maaf, bisa tolong jelaskan lagi bagian ini?", "Kamu ngomong apa sih?"]', 'Maaf, bisa tolong jelaskan lagi bagian ini?', 'Bertanya untuk klarifikasi mencegah kesalahan fatal di kemudian hari.', 10, true),
('safety-2', 'Digital Safety: Password', 'Apakah menggunakan nama kucing sebagai password aman?', 'TRUE_FALSE', '["Benar", "Salah"]', 'Salah', 'Password harus unik dan kombinasi huruf, angka, serta simbol agar sulit ditebak.', 10, true),
('finance-2', 'Finance: Utang Konsumtif', 'Membeli HP baru dengan cicilan yang menghabiskan 50% gajimu adalah keputusan finansial yang bijak.', 'TRUE_FALSE', '["Benar", "Salah"]', 'Salah', 'Utang konsumtif sebaiknya tidak melebihi 20-30% dari total pendapatan.', 10, true),
('career-2', 'Career: CV', 'CV (Curriculum Vitae) sebaiknya dibuat sepanjang 5 halaman agar terlihat berpengalaman.', 'TRUE_FALSE', '["Benar", "Salah"]', 'Salah', 'CV yang baik biasanya ringkas (1-2 halaman) dan langsung menyoroti pengalaman relevan.', 10, true),
('prod-2', 'Productivity: Multitasking', 'Melakukan banyak pekerjaan sekaligus (multitasking) selalu membuat kita lebih cepat selesai.', 'TRUE_FALSE', '["Benar", "Salah"]', 'Salah', 'Multitasking sering kali menurunkan fokus dan kualitas hasil kerja.', 10, true),
('comm-2', 'Communication: Feedback', 'Saat menerima kritik membangun dari teman satu tim, sikap terbaik adalah marah.', 'TRUE_FALSE', '["Benar", "Salah"]', 'Salah', 'Kritik yang membangun sebaiknya diterima dengan pikiran terbuka untuk perbaikan diri.', 10, true),
('safety-3', 'Digital Safety: Link', 'Kamu menerima email berisi link undian berhadiah dari alamat yang tidak dikenal.', 'MCQ', '["Klik langsung", "Forward ke teman", "Abaikan dan hapus", "Balas emailnya"]', 'Abaikan dan hapus', 'Link semacam itu sering kali merupakan upaya phishing untuk mencuri data pribadi.', 10, true),
('finance-3', 'Finance: Dana Darurat', 'Apa tujuan utama dari memiliki dana darurat?', 'MCQ', '["Untuk jalan-jalan", "Persiapan kejadian tak terduga (sakit/PHK)", "Beli barang diskon", "Investasi saham"]', 'Persiapan kejadian tak terduga (sakit/PHK)', 'Dana darurat melindungi kondisi finansialmu saat terjadi hal yang tak terduga.', 10, true),
('career-3', 'Career: Wawancara', 'Apa yang sebaiknya dilakukan sebelum wawancara kerja?', 'MCQ', '["Tidur saja", "Mencari tahu tentang perusahaan tersebut", "Datang terlambat", "Tidak menyiapkan apa-apa"]', 'Mencari tahu tentang perusahaan tersebut', 'Riset tentang perusahaan menunjukkan keseriusan dan inisiatifmu sebagai kandidat.', 10, true),
('prod-3', 'Productivity: Istirahat', 'Bekerja non-stop tanpa istirahat adalah cara terbaik untuk produktif.', 'TRUE_FALSE', '["Benar", "Salah"]', 'Salah', 'Istirahat sejenak (misal teknik Pomodoro) sangat penting untuk menjaga fokus dan energi.', 10, true)
ON CONFLICT (slug) DO NOTHING;
