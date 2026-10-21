-- ============================================================
-- Migration 101: Fresh tasks for 21-25 October 2026 (15 tasks)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction text may be hardcoded in
-- frontend/backend source. Admin can edit every value below through
-- the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. No rows exist on 2026-10-21..2026-10-25 (verified live:
--    zero rows on all five dates). Fresh INSERT only; the apply
--    script re-audits live and ABORTS if any row already exists
--    with submissions/ledger. No delete/deactivate of other dates.
-- 2. Inserts 15 fresh tasks (3 per date, step 1-3, existing
--    architecture and config shapes only, no new framework, no new
--    tables, no new engines, no new upload system, no economy change).
-- 3. Reward budget: 20 + 30 + 40 = 90 coins per date (<= 100 cap).
-- 4. Uniform rule: 2026-10-22 and 2026-10-23 tasks that photo the
--    member's own appearance require proper school uniform.
-- 5. camera_only stays false for this batch; the camera-only
--    mechanism belongs to tasks 600/610 (anchor preserved, untouched).
-- Themes (all 5 dates deliberately different):
--   2026-10-21 Daily Lifestyle      -> Daily Moment, Detail of the Day, Mini Journal
--   2026-10-22 School Day           -> School Look, School Essentials, My School Moment
--   2026-10-23 School Creativity    -> Different School Pose, Study Corner, School Day Reflection
--   2026-10-24 Creative Expression  -> Creative Self Portrait, Around Me, 15-Second Creativity
--   2026-10-25 Relaxed Weekend      -> Weekend Snapshot, Little Achievement, Weekend Recap
-- ============================================================

-- 1. Idempotent pre-clean: delete ONLY the 15 exact new titles on
--    their exact dates (previous partial run of THIS migration).
DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-21'
  AND title IN ('Daily Moment', 'Detail of the Day', 'Mini Journal');

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-22'
  AND title IN ('School Look', 'School Essentials', 'My School Moment');

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-23'
  AND title IN ('Different School Pose', 'Study Corner', 'School Day Reflection');

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-24'
  AND title IN ('Creative Self Portrait', 'Around Me', '15-Second Creativity');

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-25'
  AND title IN ('Weekend Snapshot', 'Little Achievement', 'Weekend Recap');

-- 2. Insert 3 fresh tasks for 2026-10-21 (Daily Lifestyle).
INSERT INTO odyssey_tasks (
    family_id, title, description, task_type, evaluation_type,
    step_order, active_date, reward_coins, reward_xp, target_scope,
    is_active, config
) VALUES
(
    'demo-crew-1',
    'Daily Moment',
    'Hari biasa menyimpan momen yang layak direkam! Foto SATU aktivitas sehari-hari yang nyata hari ini dengan komposisi natural apa adanya - bukan foto yang dipanggungkan berlebihan. Tambahkan catatan singkat yang menceritakan momen itu.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-21',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Ambil atau unggah tepat 1 foto (maksimal 1 file) yang menunjukkan SATU aktivitas sehari-hari yang nyata hari ini (contoh: sarapan, membereskan kamar, bermain bersama keluarga, menyiram tanaman, atau istirahat di kamar) dengan komposisi natural - biarkan tempat dan keadaannya sesuai aslinya, tidak perlu mengatur ulang segalanya demi foto. Wajib tulis catatan dengan 3 bagian: (a) Momen apa - jelaskan aktivitas yang direkam dan kira-kira jam berapa, (b) Di mana dan bersama siapa - jelaskan lingkungannya secara umum tanpa menuliskan alamat lengkap, (c) Kenapa layak direkam - jelaskan apa yang membuat momen itu terasa khas hari ini. Jangan memotret orang lain tanpa izin. Outfit bebas asalkan sopan dan sesuai aktivitas. Admin akan menolak submission tanpa foto aktivitas nyata, tanpa 3 bagian catatan, atau foto yang terlihat jelas dipanggungkan berlebihan."
    }'::jsonb
),
(
    'demo-crew-1',
    'Detail of the Day',
    'Setiap hari ada detail kecil yang sering lewat begitu saja! Foto SATU detail menarik yang kamu temukan di sekitarmu hari ini - bisa dari benda, sudut ruangan, cahaya, atau suasana yang sedang terjadi. Ceritakan detail itu dengan catatan singkat.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-10-21',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Temukan SATU detail menarik di sekitarmu HARI INI (contoh: bayangan sore di dinding, embun di jendela, suasana meja makan petang, susunan buku di rak, atau sudut favorit di rumah yang jarang diperhatikan), lalu ambil atau unggah tepat 1 foto detail itu (maksimal 1 file). Ambil dari jarak dekat atau sudut yang biasa terlewat orang. Wajib tulis catatan dengan 3 bagian: (a) Detail apa - jelaskan apa yang kamu rekam, (b) Dari mana - sebutkan objek atau tempat detail itu berada dan bagian mananya, (c) Cerita di baliknya - hubungkan detail itu dengan kegiatan atau suasana hari ini. Fokusnya adalah menangkap detail suasana atau sekitar yang ditemukan hari ini, bukan foto tanda pemakaian atau kerusakan benda lama. Admin akan menolak submission tanpa foto detail, tanpa 3 bagian catatan, atau detail yang tidak ada hubungannya dengan hari ini."
    }'::jsonb
),
(
    'demo-crew-1',
    'Mini Journal',
    'Catatan singkat hari ini bisa jadi kenangan berharga! Tulis SATU hal menarik yang kamu alami hari ini dengan bahasamu sendiri - singkat, jujur, dan jelas.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-10-21',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Hal menarik hari ini - ceritakan SATU kejadian atau pengalaman hari ini yang menarik bagimu secara singkat dari awal sampai akhirnya:\n2) Kenapa menarik - jelaskan apa yang membuatnya menarik, lucu, mengejutkan, atau tidak akan kamu lupakan:\n3) Kesan hari ini - tulis SATU kesan atau pelajaran kecil yang kamu ambil dari hari ini:\n\nGunakan bahasamu sendiri dan jujur. Tidak perlu menuliskan informasi pribadi seperti alamat, nama sekolah lengkap, atau nomor telepon. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, atau cerita yang bukan pengalamanmu sendiri.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
);

-- 3. Insert 3 fresh tasks for 2026-10-22 (School Day).
--    Uniform rule: tasks displaying the member's own appearance
--    (School Look, My School Moment) require complete school uniform.
INSERT INTO odyssey_tasks (
    family_id, title, description, task_type, evaluation_type,
    step_order, active_date, reward_coins, reward_xp, target_scope,
    is_active, config
) VALUES
(
    'demo-crew-1',
    'School Look',
    'Hari sekolah, baju sekolah yang rapi adalah awal yang baik! Kenakan baju sekolahmu lengkap dan rapi, lalu foto dirimu dalam pose natural yang menunjukkan penampilan siap sekolah hari ini.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-22',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Kenakan baju sekolah lengkap dan rapi (seragam atau setelan sekolah sesuai aturan sekolahmu), lalu ambil atau unggah tepat 1 foto (maksimal 1 file) yang menampilkan penampilan sekolahmu hari ini. Pose natural dan percaya diri - berdiri atau berjalan seperti biasanya, bukan sengaja melihat ke bawah. Wajib tulis catatan dengan 3 bagian: (a) Outfit - sebutkan bagian baju sekolah yang kamu kenakan (atasan, bawahan, alas kaki), (b) Persiapan - jelaskan SATU hal yang kamu lakukan supaya tampil rapi hari ini, (c) Perasaan - tulis singkat bagaimana perasaanmu menyambut hari sekolah ini. Baju sekolah wajib pada foto ini; foto dengan pakaian lain akan ditolak. Jangan memperlihatkan alamat, papan nama rumah, atau informasi pribadi di dalam foto."
    }'::jsonb
),
(
    'demo-crew-1',
    'School Essentials',
    'Perlengkapan sekolah yang tertata membuat pagi tidak tersendat! Foto perlengkapan sekolahmu yang tertata rapi - tas yang sudah diisi, buku dan kotak pensil tersusun, atau perlengkapan hari ini yang sudah siap dibawa.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-10-22',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Bereskan dan susun perlengkapan sekolah pas untuk hari ini (contoh: isi tas untuk hari ini, buku pelajaran hari ini, kotak pensil, seragam lengkap, alat tulis, atau botol minum), lalu ambil atau unggah tepat 1 foto (maksimal 1 file) yang memperlihatkan perlengkapan itu tersusun rapi. Ambil dari atas (top view) atau sudut yang memperlihatkan keseluruhan susunannya. Wajib tulis catatan dengan 3 bagian: (a) Perlengkapan apa saja yang tersusun rapi, (b) Cara menata - jelaskan langkah singkat yang kamu lakukan untuk menyiapkannya, (c) Kenapa penting - jelaskan kesulitan apa yang terhindar berkat perlengkapan yang tertata. Ini tugas soal perlengkapan sekolah dan keterampilan menata, bukan foto outfit diri. Admin akan menolak submission tanpa foto perlengkapan yang tersusun rapi, tanpa 3 bagian catatan, atau foto yang tidak memperlihatkan perlengkapan sekolah sama sekali."
    }'::jsonb
),
(
    'demo-crew-1',
    'My School Moment',
    'Di sekolah ada suasana yang tidak terjadi dua kali! Foto SATU suasana aktivitas di sekolahmu hari ini yang sesuai, aman, dan menunjukkan kemeriahan hari sekolah.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-10-22',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Ambil atau unggah tepat 1 foto (maksimal 1 file) yang menunjukkan SATU suasana atau aktivitas nyata di lingkungan sekolahmu hari ini (contoh: suasana belajar di kelas, kegiatan istirahat, piket kelas yang sudah rapi, atau area sekolah yang terlihat hidup pagi itu). Foto harus aman dan sesuai - memotret orang lain hanya jika kamu sudah meminta izinnya; kalau tidak, pilih suasana yang tidak menampilkan wajah orang lain atau ambil dari belakang. Apabila kamu ikut terlihat di foto, kamu wajib memakai baju sekolah. Wajib tulis catatan dengan 3 bagian: (a) Suasana apa yang kamu rekam, (b) Di bagian sekolah mana suasana itu terjadi (sebutkan secara umum saja), (c) Kenapa berkesan - jelaskan apa yang membuat suasana ini layak diingat. Tidak perlu menyebutkan nama lengkap orang lain, nama sekolah, atau alamat. Admin akan menolak submission tanpa foto suasana sekolah, tanpa 3 bagian catatan, atau foto tanpa izin orang yang tampil di dalamnya."
    }'::jsonb
);

-- 4. Insert 3 fresh tasks for 2026-10-23 (School Creativity).
--    Uniform rule: School Day Reflection is TEXT; Different School Pose
--    requires complete school uniform and MUST differ from 2026-10-22.
INSERT INTO odyssey_tasks (
    family_id, title, description, task_type, evaluation_type,
    step_order, active_date, reward_coins, reward_xp, target_scope,
    is_active, config
) VALUES
(
    'demo-crew-1',
    'Different School Pose',
    'Kemarin kamu menampilkan penampilan sekolahmu dengan pose natural; hari ini tampil bedanya! Kenakan lagi baju sekolahmu, lalu foto dirimu dengan pose atau komposisi yang BERBEDA dari foto sekolah kemarin.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-23',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Kenakan baju sekolah lengkap dan rapi, lalu ambil atau unggah tepat 1 foto (maksimal 1 file) dirimu dengan SATU pose atau komposisi yang baru - berbeda dari foto penampilan sekolah kemarin (contoh: duduk sambil membaca buku, foto dari samping, membawa tas di satu bahu, atau berdiri di tempat yang berbeda). Cukup dengan yang ada di sekitar, tidak perlu membeli atau berpindah tempat jauh. Wajib tulis catatan dengan 3 bagian: (a) Pose atau komposisi baru apa yang kamu coba hari ini, (b) Beda dari kemarin - jelaskan apa yang berbeda dari foto sekolah kemarin, (c) Perasaan - tulis singkat bagaimana rasanya mencoba tampil dengan cara baru. Baju sekolah wajib pada foto ini; foto dengan pakaian lain akan ditolak. Admin akan menolak submission tanpa baju sekolah, tanpa 3 bagian catatan, atau foto yang ternyata sama persis dengan foto penampilan sekolah kemarin."
    }'::jsonb
),
(
    'demo-crew-1',
    'Study Corner',
    'Sudut belajar yang nyaman membuat belajar lebih semangat! Foto meja belajar, tumpukan buku, atau perlengkapan belajarmu dalam keadaan tertata.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-10-23',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Ambil atau unggah tepat 1 foto (maksimal 1 file) yang memperlihatkan sudut belajarmu: meja belajar, buku, catatan, atau perlengkapan belajar lain. Boleh membereskan dulu sebelum difoto, boleh juga kondisi asli yang sudah rapi. Ambil dari sudut yang memperlihatkan keseluruhan tempat belajarmu. Wajib tulis catatan dengan 3 bagian: (a) Apa saja yang terlihat di sudut belajarmu, (b) Kebiasaan - jelaskan bagaimana biasanya kamu menata atau memakai sudut itu untuk belajar, (c) Satu perbaikan - tulis SATU hal yang akan kamu tata atau tambahkan supaya belajar makin nyaman. Ini tugas soal tempat dan kebiasaan belajar di rumah, bukan foto perlengkapan yang dibawa ke sekolah. Admin akan menolak submission tanpa foto sudut belajar, tanpa 3 bagian catatan, atau foto yang tidak memperlihatkan suasana belajar sama sekali."
    }'::jsonb
),
(
    'demo-crew-1',
    'School Day Reflection',
    'Apa yang kamu pelajari hari ini akan lebih melekat kalau direfleksikan! Tulis SATU hal yang kamu pelajari atau pengalaman dari hari sekolah ini beserta maknanya bagimu.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-10-23',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Hal yang dipelajari - sebutkan SATU pelajaran atau materi yang kamu pelajari di sekolah hari ini dan ceritakan singkat apa isinya:\n2) Pemahamanmu - jelaskan dengan bahasamu sendiri apa yang kamu pahami dari hal itu dan menurutmu kenapa hal itu berguna:\n3) Refleksi - ceritakan SATU pengalaman kecil dari hari sekolah ini yang membuatmu penasaran, terinspirasi, atau ingin memperbaiki sesuatu:\n\nGunakan bahasamu sendiri. Tidak perlu menuliskan nama sekolah atau nama guru lengkap. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, atau isi yang tidak berkaitan dengan hari sekolah ini.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
);

-- 5. Insert 3 fresh tasks for 2026-10-24 (Creative Expression).
INSERT INTO odyssey_tasks (
    family_id, title, description, task_type, evaluation_type,
    step_order, active_date, reward_coins, reward_xp, target_scope,
    is_active, config
) VALUES
(
    'demo-crew-1',
    'Creative Self Portrait',
    'Sebentar kreatif boleh banget! Foto dirimu dengan komposisi kreatif sederhana yang bisa dibuat langsung dengan barang di sekitar - gunakan outfit bebas yang tetap sopan.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-24',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Ambil atau unggah tepat 1 foto (maksimal 1 file) dirimu dengan komposisi kreatif sederhana (contoh: pantulan di cermin, bayangan di dinding, bingkai dari pintu atau daun, foto dengan barang-barang warna-warni sebagai latar, atau ekspresi dan pembawaan yang kamu suka). Gunakan hanya barang yang sudah ada di sekitar - tidak perlu membeli apa pun atau membuat instalasi rumit. Outfit bebas asalkan sopan dan nyaman. Wajib tulis catatan dengan 3 bagian: (a) Ide kreatif yang kamu coba, (b) Cara - bagaimana kamu membuat foto itu dengan barang yang ada, (c) Kesan - tulis singkat bagaimana hasilnya menurutmu. Jangan memotret orang lain tanpa izin dan hindari situasi yang berisiko atau memalukan. Admin akan menolak submission tanpa sentuhan kreatif, tanpa 3 bagian catatan, atau foto yang menempatkan dirimu dalam situasi yang berbahaya."
    }'::jsonb
),
(
    'demo-crew-1',
    'Around Me',
    'Tidak perlu ke tempat jauh - di sekitar juga penuh komposisi unik! Foto SATU objek, tempat, atau suasana di sekitarmu dari sudut yang tidak biasa.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-10-24',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Ambil atau unggah tepat 1 foto (maksimal 1 file) dari SATU sudut yang tidak biasa untuk memotret objek, tempat, atau suasana di sekitarmu (contoh: ditembak dari bawah, dari atas dekat, melalui celah pintu, fokus ke satu objek kecil di tengah, atau simetri yang tidak sengaja terbentuk). Tidak perlu membeli apa pun dan tidak perlu berpindah tempat yang jauh - cukup tetangga rumah atau sekitar. Wajib tulis catatan dengan 3 bagian: (a) Objek atau suasana apa yang kamu tangkap, (b) Sudut - jelaskan dari posisi atau sudut mana kamu memotret dan kenapa memilihnya, (c) Efek - tulis singkat apa yang berbeda terasa ketika dilihat dari sudut itu. Admin akan menolak submission tanpa sudut yang tidak biasa, tanpa 3 bagian catatan, atau foto yang ternyata hanya mengambil sudut biasa."
    }'::jsonb
),
(
    'demo-crew-1',
    '15-Second Creativity',
    'Kreativitas tidak harus rumit atau mahal! Rekam video singkat 10-15 detik yang menunjukkan SATU kreativitas sederhanamu - dari kegiatan harian, gerakan, karya kecil, atau ekspresimu sendiri.',
    'VIDEO',
    'ADMIN_REVIEW',
    3,
    '2026-10-24',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 15,
        "camera_facing": "user",
        "instruction": "Rekam video 10-15 detik yang menunjukkan SATU kreativitas sederhana buatanmu sendiri (contoh: menggambar atau menulis dengan cepat, menyusun barang jadi satu pola menarik, gerakan kegemaran singkat, atau membaca dan menceritakan sesuatu dengan ekspresi). Gunakan barang yang sudah ada di sekitar - cukup rekam langsung tanpa perlu editing. Wajib tulis catatan dengan 3 bagian: (a) Kreativitas apa yang kamu tunjukkan, (b) Cara - bagaimana kamu menyiapkan dan menjalankannya, (c) Cerita - bagaimana perasaanmu membuat video itu. Durasi maksimal 15 detik. Admin akan menolak video melebihi 15 detik, tanpa kreativitas yang jelas, tanpa 3 bagian catatan, atau yang membutuhkan pembelian barang atau berbahaya."
      }
    }'::jsonb
);

-- 6. Insert 3 fresh tasks for 2026-10-25 (Relaxed Weekend).
INSERT INTO odyssey_tasks (
    family_id, title, description, task_type, evaluation_type,
    step_order, active_date, reward_coins, reward_xp, target_scope,
    is_active, config
) VALUES
(
    'demo-crew-1',
    'Weekend Snapshot',
    'Akhir pekan adalah waktunya bersantai dan jadi diri sendiri! Foto SATU aktivitas akhir pekan yang nyata dan natural yang sedang kamu jalani hari ini.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-25',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Ambil atau unggah tepat 1 foto (maksimal 1 file) yang memperlihatkan SATU aktivitasmu di akhir pekan (contoh: sarapan yang lebih santai, membantu beres-beres rumah, bermain bersama keluarga, menata barang kesukaan, atau sekadar menikmati sore). Foto apa adanya seperti kegiatan yang sedang berjalan, tanpa disiap khusus demi foto. Wajib tulis catatan dengan 3 bagian: (a) Aktivitas apa yang direkam dan kira-kira jam berapa, (b) Suasana hari ini - jelaskan bagaimana terasanya menjalani akhir pekan ini, (c) Satu hal yang membuat momen ini menyenangkan. Outfit bebas, sesuai aktivitas, sopan, dan nyaman. Jangan memotret orang lain tanpa izin. Admin akan menolak submission tanpa foto aktivitas nyata, tanpa 3 bagian catatan, atau foto yang jelas dibuat berlebihan."
    }'::jsonb
),
(
    'demo-crew-1',
    'Little Achievement',
    'Di akhir pekan yang santai ada hasil kecil yang layak dipuji! Foto SATU kegiatan sederhana, karya, atau sesuatu yang berhasil kamu selesaikan sendiri.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-10-25',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Ambil atau unggah tepat 1 foto (maksimal 1 file) yang memperlihatkan hasil dari SATU kegiatan sederhana yang kamu selesaikan sendiri (contoh: tanaman yang berhasil disiram dan ditata, kerajinan kecil, tugas yang selesai, rak buku yang sudah dibereskan, atau masakan sederhana). Tidak harus hasil besar - usaha tanganmu sendirilah yang dinilai. Wajib tulis catatan dengan 3 bagian: (a) Hasil apa yang berhasil kamu selesaikan, (b) Proses - jelaskan langkah singkat yang kamu lakukan untuk sampai ke situ, (c) Perasaan - tulis apa yang kamu rasakan setelah selesai. Admin akan menolak submission tanpa foto hasil, tanpa 3 bagian catatan, atau hasil yang jelas bukan usaha atau buatanmu sendiri."
    }'::jsonb
),
(
    'demo-crew-1',
    'Weekend Recap',
    'Tutup akhir pekanmu dengan merangkumnya sendiri! Tulis singkat kegiatan dan pengalaman akhir pekanmu - apa yang terjadi, yang paling berkesan, dan bekal masuk pekan baru.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-10-25',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Kegiatan akhir pekan - ceritakan secara singkat apa saja yang kamu lakukan pada akhir pekan ini (contoh: beres-beres, bermain, kegiatan keluarga, kegiatan di rumah, atau istirahat):\n2) Yang paling berkesan - pilih SATU momen yang paling berkesan dan jelaskan kenapa:\n3) Siap pekan baru - tulis SATU hal yang membuatmu siap memulai pekan baru (contoh: pekerjaan yang sudah selesai, semangat baru, atau rencana kecil):\n\nGunakan bahasamu sendiri dan jujur. Tidak perlu menuliskan informasi pribadi seperti alamat atau nomor telepon. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, atau cerita yang tidak berkaitan dengan akhir pekan ini.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
);

-- 7. Bump schema version.
INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '101_tasks_2026_10_21_25')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
