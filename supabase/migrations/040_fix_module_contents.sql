-- Migration 040: Fix Module Contents
-- Overwrites fantasy missions with realistic family learning modules and assigns them to the correct journeys/courses.

-- 1. Komunikasi Dasar (Literasi Keluarga)
UPDATE odyssey_mission_definitions 
SET journey = 'literasi-keluarga', course = 'komunikasi-dasar',
    title = 'Menulis Pesan Sopan', 
    description = 'Pelajari cara menulis pesan yang baik dan sopan di WhatsApp atau SMS.',
    exercise_defs = '[{"slug":"baca-contoh","description":"Baca contoh pesan yang sopan kepada guru atau orang tua.","type":"OBSERVATION"},{"slug":"tulis-pesan","description":"Tuliskan satu contoh draf pesan izin tidak masuk sekolah karena sakit.","type":"WRITE"}]'::jsonb
WHERE slug = 'morning-light';

UPDATE odyssey_mission_definitions 
SET journey = 'literasi-keluarga', course = 'komunikasi-dasar',
    title = 'Berbicara dengan Jelas', 
    description = 'Latih kejelasan intonasi dan pilihan kata saat berbicara.',
    exercise_defs = '[{"slug":"latih-intonasi","description":"Ucapkan kalimat halo dengan tiga intonasi berbeda.","type":"OBSERVATION"},{"slug":"rekam-suara","description":"Tuliskan apa perbedaan dari ketiga intonasi tersebut.","type":"WRITE"}]'::jsonb
WHERE slug = 'gather-herbs';

-- 2. Kebiasaan Baik (Literasi Keluarga)
UPDATE odyssey_mission_definitions 
SET journey = 'literasi-keluarga', course = 'kebiasaan-baik',
    title = 'Merapikan Tempat Tidur', 
    description = 'Kebiasaan kecil di pagi hari yang berdampak besar sepanjang hari.',
    exercise_defs = '[{"slug":"rapi-kasur","description":"Rapikan tempat tidurmu pagi ini.","type":"OBSERVATION"},{"slug":"cerita-pagi","description":"Tuliskan bagaimana perasaanmu setelah melihat kamar yang rapi.","type":"WRITE"}]'::jsonb
WHERE slug = 'riddle-of-the-stones';

UPDATE odyssey_mission_definitions 
SET journey = 'literasi-keluarga', course = 'kebiasaan-baik',
    title = 'Jadwal Harian', 
    description = 'Membuat rutinitas agar waktu luang tidak terbuang percuma.',
    exercise_defs = '[{"slug":"buat-jadwal","description":"Buat daftar 3 aktivitas penting yang akan kamu lakukan hari ini.","type":"WRITE"},{"slug":"prioritas","description":"Pilih 1 dari 3 aktivitas tersebut yang paling penting untuk diselesaikan pertama.","type":"OBSERVATION"}]'::jsonb
WHERE slug = 'shadow-trail';

-- 3. Matematika Praktis (Literasi Finansial)
UPDATE odyssey_mission_definitions 
SET journey = 'literasi-finansial', course = 'matematika-praktis',
    title = 'Menghitung Diskon', 
    description = 'Belajar menghitung potongan harga saat berbelanja.',
    exercise_defs = '[{"slug":"hitung-diskon","description":"Berapa harga barang Rp100.000 jika diskon 20%? Tuliskan jawabannya.","type":"WRITE"}]'::jsonb
WHERE slug = 'the-old-growth';

UPDATE odyssey_mission_definitions 
SET journey = 'literasi-finansial', course = 'matematika-praktis',
    title = 'Uang Kembalian', 
    description = 'Melatih ketelitian menghitung uang kembalian kasir.',
    exercise_defs = '[{"slug":"kembalian","description":"Kamu beli barang seharga Rp32.000 dengan uang Rp50.000. Berapa kembaliannya?","type":"WRITE"}]'::jsonb
WHERE slug = 'forest-riddle';

-- 4. Manajemen Uang (Literasi Finansial)
UPDATE odyssey_mission_definitions 
SET journey = 'literasi-finansial', course = 'manajemen-uang',
    title = 'Kebutuhan vs Keinginan', 
    description = 'Bisa membedakan mana yang wajib dibeli dan mana yang hanya keinginan.',
    exercise_defs = '[{"slug":"kebutuhan","description":"Sebutkan 3 barang yang termasuk Kebutuhan.","type":"WRITE"},{"slug":"keinginan","description":"Sebutkan 2 barang yang termasuk Keinginan.","type":"WRITE"}]'::jsonb
WHERE slug = 'clockwork-intro';

UPDATE odyssey_mission_definitions 
SET journey = 'literasi-finansial', course = 'manajemen-uang',
    title = 'Pentingnya Menabung', 
    description = 'Mengumpulkan sisa uang untuk tujuan di masa depan.',
    exercise_defs = '[{"slug":"target-nabung","description":"Tuliskan 1 barang yang ingin kamu beli dengan uang tabunganmu.","type":"WRITE"}]'::jsonb
WHERE slug = 'gear-hunt';

-- 5. Dunia Digital (Persiapan Karier)
UPDATE odyssey_mission_definitions 
SET journey = 'persiapan-karier', course = 'dunia-digital',
    title = 'Waspada Phishing', 
    description = 'Cara mengenali link penipuan yang dikirim via SMS atau WhatsApp.',
    exercise_defs = '[{"slug":"cek-link","description":"Sebutkan satu ciri-ciri link penipuan atau phishing.","type":"WRITE"}]'::jsonb
WHERE slug = 'the-copper-key';

UPDATE odyssey_mission_definitions 
SET journey = 'persiapan-karier', course = 'dunia-digital',
    title = 'Password yang Aman', 
    description = 'Membuat kata sandi yang sulit ditebak hacker.',
    exercise_defs = '[{"slug":"buat-password","description":"Tuliskan kombinasi password yang kuat (huruf besar, kecil, angka, simbol) tanpa memakai nama aslimu.","type":"WRITE"}]'::jsonb
WHERE slug = 'clockwork-expedition';

-- 6. Dunia Kerja (Persiapan Karier)
UPDATE odyssey_mission_definitions 
SET journey = 'persiapan-karier', course = 'dunia-kerja',
    title = 'Menulis CV Dasar', 
    description = 'Apa saja yang harus dimasukkan ke dalam daftar riwayat hidup.',
    exercise_defs = '[{"slug":"isi-cv","description":"Sebutkan 3 komponen utama yang harus ada di dalam CV.","type":"WRITE"}]'::jsonb
WHERE slug = 'star-observation';

UPDATE odyssey_mission_definitions 
SET journey = 'persiapan-karier', course = 'dunia-kerja',
    title = 'Persiapan Wawancara', 
    description = 'Bagaimana menjawab pertanyaan wawancara dengan percaya diri.',
    exercise_defs = '[{"slug":"jawab-wawancara","description":"Tuliskan satu kalimat untuk memperkenalkan dirimu di sesi wawancara.","type":"WRITE"}]'::jsonb
WHERE slug = 'constellation-map';

-- The last one
UPDATE odyssey_mission_definitions 
SET journey = 'persiapan-karier', course = 'dunia-kerja',
    title = 'Etika Bekerja', 
    description = 'Aturan tidak tertulis di tempat kerja.',
    exercise_defs = '[{"slug":"etika-kerja","description":"Sebutkan satu contoh etika yang baik saat bekerja dalam tim.","type":"WRITE"}]'::jsonb
WHERE slug = 'library-lore';
