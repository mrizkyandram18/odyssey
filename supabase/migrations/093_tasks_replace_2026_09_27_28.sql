-- ============================================================
-- Migration 093: Replace tasks for 27 + 28 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction/question/scenario text may be
-- hardcoded in frontend/backend source. Admin can edit every value below
-- through the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. DELETES exact tasks 564/565 on 2026-09-27 and 566/567 on
--    2026-09-28. Production evidence (audit at apply time): all four
--    carry 0 submissions in odyssey_task_submissions and therefore 0
--    linked rows in odyssey_coin_transactions (reference_id);
--    odyssey_claims has no task_id column, so no claim linkage is
--    possible. Exact-ID hard-delete is safe. No other rows exist on
--    either date. History of all other dates is untouched.
-- 2. Inserts 6 fresh tasks (3 per date, existing architecture and
--    config shapes only, no new framework, no new tables, no new
--    engines, no new upload system):
--    2026-09-27 (school-uniform day):
--    - Step 1: PHOTO_UPLOAD (Sehari Kembali Pakai Seragam) - member
--      wears an existing school uniform they still own, photographs
--      themselves neatly, and writes a short note about what they
--      remember most or what it feels like wearing it again.
--      max_files 1, camera_only = false (gallery/file-picker allowed
--      via existing CameraCaptureModal path, same as tasks 628/632/
--      643/650); ADMIN_REVIEW via the existing approve/reject flow.
--      Reward 40c/100xp.
--    - Step 2: VIDEO (Kalau Sekarang Masih Pakai Seragam) - member
--      wears the uniform and records max 60 seconds pretending to get
--      ready for school again, explaining one thing that is different
--      about themselves now versus back then. recording.enabled,
--      max_duration_seconds = 60, camera user. Resolves to
--      ADMIN_REVIEW via existing ResolveEvaluationTypeForConfig
--      (same shape as tasks 616, 629, 631, 641, 647, 653, 658).
--      Reward 40c/100xp.
--    - Step 3: MINI_GAME (Balik Jadi Anak Sekolah Sehari) - 5
--      practical back-to-school situations (waking up late, uniform
--      not ready, forgotten supplies, almost late, item left behind),
--      each with 3 actions + deterministic delta + hint. DECISION_
--      PRIORITY + POINTS, initial 0, target 100; best path = 5 x 20 =
--      100. AUTO via the existing ValidateDecisionChoices server
--      recompute (client score never trusted; any complete choice set
--      accepted per engine contract). Reward 50c/100xp.
--    2026-09-28 (uniform then vs now):
--    - Step 1: PHOTO_UPLOAD (Seragam vs Dirimu Sekarang) - member
--      wears the uniform, poses showing the difference between now
--      and school days, plus a short note about that change.
--      max_files 1, camera_only = false; ADMIN_REVIEW. Reward
--      40c/100xp.
--    - Step 2: TEXT_RESPONSE (Kalau Bisa Balik ke Hari Pertama
--      Sekolah) - member imagines returning to school tomorrow and
--      writes what they would do differently plus one piece of advice
--      for their younger self. min 150 chars, max 2000
--      (server-enforced by odyssey_submit_manual_task from config);
--      ADMIN_REVIEW. Reward 40c/100xp.
--    - Step 3: VIDEO (Dari Seragam Sekolah ke Seragam Kerja) -
--      member wears the uniform and records max 60 seconds about the
--      journey from school to now: what changed, what stayed the
--      same, one lesson learned. recording.enabled,
--      max_duration_seconds = 60, camera user; ADMIN_REVIEW. Reward
--      40c/100xp.
--    UNIFORM FLEXIBILITY (all photo/video tasks above): instructions
--    say to use a school uniform the member still owns and that is
--    still wearable. Members must NOT buy a new uniform, are NOT
--    required to show any specific school logo, and must NOT share
--    school identity information. No cost is required.
-- 3. Idempotency: the pre-clean DELETEs below match ONLY the 6 exact
--    new titles on their exact dates (narrow exact-title guard for dev
--    re-runs, same spirit as migrations 052/083/084/085/086/087/088/
--    090/091). No broad date/title matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), 2026-09-13
--    (630/631/632), 2026-09-15 (633/634/635), 2026-09-16 (636/637/638),
--    2026-09-17 (639/640/641), 2026-09-18 (642/643/644), 2026-09-19
--    (645/646/647), 2026-09-20 (648/649/650), 2026-09-21 (651/652/653),
--    2026-09-22 (654/655/656), 2026-09-23 (657/658/659), tasks 608-620
--    (610 camera_only stays true), payout/economy config,
--    announcements, RLS, RPCs, triggers, claims, shop, or any other
--    date.
-- ============================================================

-- 1. Remove exact legacy tasks for 2026-09-27 / 2026-09-28 (proven dependency-free).
DELETE FROM odyssey_tasks
WHERE id IN (564, 565)
  AND active_date = '2026-09-27';

DELETE FROM odyssey_tasks
WHERE id IN (566, 567)
  AND active_date = '2026-09-28';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact dates). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-27'
  AND title IN (
    'Sehari Kembali Pakai Seragam',
    'Kalau Sekarang Masih Pakai Seragam',
    'Balik Jadi Anak Sekolah Sehari'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-28'
  AND title IN (
    'Seragam vs Dirimu Sekarang',
    'Kalau Bisa Balik ke Hari Pertama Sekolah',
    'Dari Seragam Sekolah ke Seragam Kerja'
  );

-- 3. Insert 3 fresh tasks for 2026-09-27.
INSERT INTO odyssey_tasks (
    family_id,
    title,
    description,
    task_type,
    evaluation_type,
    step_order,
    active_date,
    reward_coins,
    reward_xp,
    target_scope,
    is_active,
    config
) VALUES
(
    'demo-crew-1',
    'Sehari Kembali Pakai Seragam',
    'Masih menyimpan seragam sekolah yang layak dipakai? Hari ini waktunya memakainya lagi! Kenakan seragammu dengan rapi, ambil satu foto dirimu, lalu tulis catatan singkat: apa yang paling kamu ingat ketika memakai seragam itu, atau kesan apa yang muncul saat memakainya kembali.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-09-27',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). Pastikan penampilanmu rapi. 2) Ambil atau unggah 1 foto dirimu memakai seragam tersebut (boleh kamera langsung atau galeri). 3) Wajib tulis catatan singkat dengan 2 bagian: (a) Kenangan - apa yang paling kamu ingat ketika memakai seragam tersebut dulu, (b) Kesan - pengalaman atau perasaan apa yang muncul ketika memakainya kembali sekarang. Admin akan menolak submission tanpa foto memakai seragam, tanpa 2 bagian catatan, atau foto yang tidak pantas."
    }'::jsonb
),
(
    'demo-crew-1',
    'Kalau Sekarang Masih Pakai Seragam',
    'Bayangkan pagi ini kamu harus berangkat sekolah lagi! Kenakan seragam sekolahmu, lalu rekam video maksimal 60 detik seolah-olah kamu sedang bersiap berangkat sekolah. Dalam video, jelaskan satu hal yang berbeda antara dirimu sekarang dengan dirimu ketika masih sekolah.',
    'VIDEO',
    'ADMIN_REVIEW',
    2,
    '2026-09-27',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). Rekam video maksimal 60 detik dengan struktur:\n1. Persiapan - perankan seolah-olah kamu sedang bersiap berangkat sekolah lagi pagi ini (misalnya merapikan seragam, menyiapkan tas, atau pamit berangkat)\n2. Perbedaan - jelaskan SATU hal yang berbeda antara dirimu sekarang dengan dirimu ketika masih sekolah (misalnya kebiasaan, cara berpikir, keberanian, atau tanggung jawab)\n\nBicaralah senatural mungkin seperti bercerita ke seorang teman. Jangan menyebut data pribadi sensitif dan jangan menjelekkan pihak tertentu."
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Balik Jadi Anak Sekolah Sehari',
    'Hari ini kamu balik jadi anak sekolah sehari! Alarm berbunyi, seragam belum siap, tas belum dikemas, dan jam terus berjalan. Hadapi 5 situasi pagi yang kacau ini dan pilih tindakan terbaik pada setiap situasi. Setiap pilihan ada konsekuensinya - yang paling siap dan tenang akan meraih poin tertinggi!',
    'MINI_GAME',
    'AUTO',
    3,
    '2026-09-27',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "game": "DECISION_PRIORITY",
      "target_score": 100,
      "scenario": {
        "currency": "POINTS",
        "initial_balance": 0,
        "events": [
          {
            "id": "ev_1",
            "title": "Situasi 1 - Bangun Kesiangan",
            "description": "Alarm tidak berbunyi dan kamu terbangun 30 menit lebih siang dari biasanya. Seragam masih di jemuran, dan sekolah dimulai 45 menit lagi. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Tetap tenang: mandi cepat, kenakan seragam yang paling rapi tersedia, dan prioritaskan hal yang paling penting saja",
                "delta": 20,
                "hint": "Panik hanya membuang waktu - yang kesiangan tapi tetap sistematis akan lebih siap berangkat"
              },
              {
                "id": "b",
                "label": "Berlama-lama memilih seragam terbaik agar tetap tampil sempurna meski terlambat",
                "delta": 5,
                "hint": "Penampilan penting, tetapi mengorbankan waktu berangkat membuatmu makin terlambat"
              },
              {
                "id": "c",
                "label": "Putuskan bolos saja karena sudah kesiangan dan percuma berangkat",
                "delta": 0,
                "hint": "Terlambat sedikit masih jauh lebih baik daripada tidak masuk sama sekali"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 - Seragam Belum Siap",
            "description": "Seragam yang ingin kamu pakai ternyata belum disetrika dan ada noda kecil di lengannya. Waktu tersisa 20 menit. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Rapikan secepatnya: kibas dan gantung agar rapi, bersihkan nodanya, atau pakai seragam cadangan yang bersih",
                "delta": 20,
                "hint": "Solusi praktis mengalahkan kesempurnaan - seragam cadangan yang bersih lebih baik dari seragam favorit yang kotor"
              },
              {
                "id": "b",
                "label": "Pakai saja apa adanya tanpa dirapikan, yang penting berangkat",
                "delta": 10,
                "hint": "Berangkat itu bagus, tetapi kerapian seragam adalah bagian dari kedisiplinan siswa"
              },
              {
                "id": "c",
                "label": "Menunggu orang tua pulang untuk menyiapkan seragammu",
                "delta": 0,
                "hint": "Menunggu tanpa berbuat apa-apa hanya menghabiskan waktu yang tersisa"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 - Lupa Menyiapkan Perlengkapan",
            "description": "Semalam kamu lupa menyiapkan perlengkapan: buku pelajaran, alat tulis, dan bekal belum dikemas. Bel sekolah 30 menit lagi. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Buat daftar cepat sesuai jadwal hari ini, kemas yang wajib dulu (buku dan alat tulis), lalu bekal seadanya",
                "delta": 20,
                "hint": "Daftar prioritas membuat persiapan kilat tetap lengkap untuk hal yang paling penting"
              },
              {
                "id": "b",
                "label": "Masukkan semua buku yang ada ke tas tanpa mengecek jadwal agar aman",
                "delta": 5,
                "hint": "Tas jadi berat dan jadwal tetap bisa terlewat - mengecek jadwal hanya butuh satu menit"
              },
              {
                "id": "c",
                "label": "Berangkat tanpa perlengkapan dan berencana meminjam semua ke teman di sekolah",
                "delta": 0,
                "hint": "Mengandalkan pinjaman merepotkan teman dan menunjukkan kamu tidak bertanggung jawab"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 - Hampir Terlambat di Jalan",
            "description": "Kamu sudah di jalan tetapi macet, dan gerbang sekolah tutup 10 menit lagi. Ada jalan pintas yang sepi dan tidak aman. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Tetap di jalur aman yang biasa, jalan cepat tapi tertib, dan terima konsekuensinya dengan jujur jika terlambat",
                "delta": 20,
                "hint": "Keselamatan tidak bisa ditukar dengan kecepatan - keterlambatan bisa dijelaskan, kecelakaan tidak"
              },
              {
                "id": "b",
                "label": "Meminta tumpangan orang asing yang lewat agar cepat sampai",
                "delta": 0,
                "hint": "Naik kendaraan orang yang tidak dikenal sangat berbahaya, seberapa pun mendesaknya"
              },
              {
                "id": "c",
                "label": "Memotong jalan lewat area sepi dan berbahaya demi mengejar waktu",
                "delta": 0,
                "hint": "Jalan pintas yang tidak aman bisa berakibat jauh lebih buruk daripada terlambat"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 - Ada Perlengkapan yang Tertinggal",
            "description": "Sampai di sekolah, kamu sadar ada satu perlengkapan penting yang tertinggal di rumah (misalnya tugas atau alat untuk pelajaran pertama). Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Jujur ke guru tentang apa yang terjadi, minta solusi, dan buat pengingat agar besok tidak terulang",
                "delta": 20,
                "hint": "Kejujuran plus rencana perbaikan menunjukkan kedewasaan - guru menghargai siswa yang bertanggung jawab"
              },
              {
                "id": "b",
                "label": "Menyalahkan adik atau orang tua di rumah karena tidak mengingatkanmu",
                "delta": 5,
                "hint": "Menyalahkan orang lain tidak menyelesaikan masalah dan merusak hubungan"
              },
              {
                "id": "c",
                "label": "Berbohong dengan alasan yang dibuat-buat agar tidak dimarahi",
                "delta": 0,
                "hint": "Kebohongan merusak kepercayaan - sekali ketahuan, penjelasan jujur berikutnya sulit dipercaya"
              }
            ]
          }
        ]
      }
    }'::jsonb
);

-- 4. Insert 3 fresh tasks for 2026-09-28.
INSERT INTO odyssey_tasks (
    family_id,
    title,
    description,
    task_type,
    evaluation_type,
    step_order,
    active_date,
    reward_coins,
    reward_xp,
    target_scope,
    is_active,
    config
) VALUES
(
    'demo-crew-1',
    'Seragam vs Dirimu Sekarang',
    'Seragam yang sama, orang yang berbeda! Kenakan lagi seragam sekolahmu, lalu ambil satu foto dengan pose yang menunjukkan perbedaan dirimu sekarang dibanding masa sekolah. Tambahkan catatan singkat tentang perubahan tersebut - apa yang berubah darimu sejak dulu?',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-09-28',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). 2) Ambil atau unggah 1 foto dirimu memakai seragam tersebut dengan pose yang menunjukkan perbedaan dirimu sekarang dibanding masa sekolah (misalnya pose percaya diri seperti pekerja, pose dengan barang khas aktivitasmu sekarang, atau pose yang membandingkan dulu dan kini). 3) Wajib tulis catatan singkat dengan 2 bagian: (a) Pose - jelaskan pose yang kamu pilih dan perbedaan apa yang ingin ditunjukkannya, (b) Perubahan - ceritakan satu perubahan terbesar dalam dirimu sejak masa sekolah. Admin akan menolak submission tanpa foto memakai seragam, tanpa 2 bagian catatan, atau foto yang tidak pantas."
    }'::jsonb
),
(
    'demo-crew-1',
    'Kalau Bisa Balik ke Hari Pertama Sekolah',
    'Bayangkan besok kamu harus memakai seragam dan kembali ke hari pertama sekolah! Tulis dua hal: apa yang akan kamu lakukan berbeda dibanding dulu, dan satu nasihat yang akan kamu berikan kepada dirimu saat itu. Jujur, reflektif, dan dari hati.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    2,
    '2026-09-28',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format:\n1) Yang akan dilakukan berbeda - jelaskan minimal 2 hal yang akan kamu lakukan berbeda jika kembali ke hari pertama sekolah (contoh: cara belajar, cara bergaul, keberanian mencoba hal baru, atau cara mengatur waktu), beserta alasannya:\n2) Nasihat untuk diriku dulu - tulis SATU nasihat yang akan kamu berikan kepada dirimu pada hari pertama sekolah itu, dan jelaskan kenapa nasihat itu penting:\n\nTulis dengan jujur dari pengalamanmu sendiri. Jangan mencantumkan data pribadi siapa pun. Admin akan menolak jawaban yang terlalu pendek, tanpa hal yang berbeda, tanpa nasihat, atau tanpa alasan yang masuk akal.",
      "minimum_characters": 150,
      "maximum_characters": 2000
    }'::jsonb
),
(
    'demo-crew-1',
    'Dari Seragam Sekolah ke Seragam Kerja',
    'Dari seragam sekolah sampai ke titik ini - perjalanan yang panjang! Kenakan seragam sekolahmu, lalu rekam video maksimal 60 detik tentang perjalananmu dari masa sekolah sampai sekarang: apa yang berubah, apa yang tetap sama, dan satu hal yang kamu pelajari.',
    'VIDEO',
    'ADMIN_REVIEW',
    3,
    '2026-09-28',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Kenakan seragam sekolah yang masih kamu miliki dan masih layak dipakai (tidak perlu membeli yang baru, tidak perlu logo sekolah tertentu, dan jangan mencantumkan identitas sekolah). Rekam video maksimal 60 detik dengan struktur:\n1. Yang berubah - ceritakan satu hal terbesar yang berubah dari masa sekolah sampai sekarang\n2. Yang tetap sama - ceritakan satu hal dari dirimu yang tidak berubah sampai sekarang\n3. Pelajaran - ceritakan satu hal terpenting yang kamu pelajari sepanjang perjalanan itu\n\nBicaralah senatural mungkin seperti bercerita ke seorang teman. Jangan menyebut data pribadi sensitif dan jangan menjelekkan pihak tertentu."
      }
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '093_tasks_replace_2026_09_27_28')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
