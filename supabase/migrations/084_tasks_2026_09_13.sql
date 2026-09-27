-- ============================================================
-- Migration 084: Replace tasks for 13 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/question/scenario/instruction text may be hardcoded
-- in frontend/backend source. Admin can edit every value below through
-- the existing Admin UI (Tasks tab).
-- 1. DELETES exact tasks 537/538 on 2026-09-13. Production evidence
--    (audit at apply time): both carry 0 submissions and 0 coin-ledger
--    rows, so exact-ID hard-delete is safe. No other rows exist on
--    2026-09-13. History of other dates is untouched.
-- 2. Inserts 3 fresh canonical tasks for 2026-09-13 (existing
--    architecture and config shapes only, no new framework):
--    - Step 1: QUIZ (Siaga Sehari-hari) — 5 everyday emergency-readiness
--      questions with correct_answer + explanation, AUTO evaluation.
--    - Step 2: VIDEO (60 Detik Jadi Guru Kecil) — recording.enabled,
--      max_duration_seconds = 60, camera user. Resolves to ADMIN_REVIEW
--      via existing ResolveEvaluationTypeForConfig rule (same shape as
--      tasks 616 and 629).
--    - Step 3: PHOTO_UPLOAD (Bekal Sehat Versimu) — camera_only = false
--      (gallery/file-picker allowed via existing CameraCaptureModal
--      path). Member writes a short note in the existing payload.note
--      field (same convention as DOCUMENT_UPLOAD and task 628);
--      ADMIN_REVIEW enforcement is via the existing approve/reject flow.
--      Global/camera-only semantics untouched (task 610 unaffected).
-- 3. Idempotency: the pre-clean DELETE below matches ONLY the 3 exact
--    new titles on 2026-09-13 (narrow exact-title guard for dev re-runs,
--    same spirit as migrations 052/083). No broad date/title matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), tasks 608-620,
--    payout/economy config, announcements, RLS, RPCs, triggers, or any
--    other date.
-- ============================================================

-- 1. Remove exact legacy tasks for 2026-09-13 (proven dependency-free).
DELETE FROM odyssey_tasks
WHERE id IN (537, 538)
  AND active_date = '2026-09-13';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact date). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-13'
  AND title IN (
    'Siaga Sehari-hari',
    '60 Detik Jadi Guru Kecil',
    'Bekal Sehat Versimu'
  );

-- 3. Insert 3 fresh canonical tasks for 2026-09-13.
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
    'Siaga Sehari-hari',
    'Uji kesiapanmu menghadapi kejadian tak terduga sehari-hari: api kecil di dapur, mimisan, gempa, kabel listrik rusak, dan orang pingsan. Lima pertanyaan praktis lengkap dengan penjelasan — tanpa perlu panik saat kejadian nyata.',
    'QUIZ',
    'AUTO',
    1,
    '2026-09-13',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "questions": [
        {
          "id": "q1",
          "question": "Api kecil menyala di wajan saat kamu memasak. Tindakan pertama yang paling tepat adalah?",
          "options": [
            "A. Menyiram wajan dengan air agar api cepat padam",
            "B. Menutup wajan dengan tutup panci dan mematikan kompor",
            "C. Membawa wajan berapi keluar rumah sambil berlari",
            "D. Meniup api sekuat tenaga dari jarak dekat"
          ],
          "correct_answer": "B",
          "explanation": "Menutup wajan memutus pasokan oksigen sehingga api padam. Menyiram minyak panas dengan air justru membuat api menyembur dan melukai."
        },
        {
          "id": "q2",
          "question": "Temanmu mimisan (darah keluar dari hidung). Pertolongan yang benar adalah?",
          "options": [
            "A. Mendongakkan kepala ke belakang agar darah berhenti",
            "B. Duduk tegak, condongkan badan sedikit ke depan, dan tekan hidung selama 10-15 menit",
            "C. Memasukkan tisu jauh ke dalam hidung lalu berbaring",
            "D. Menyuruhnya berlari kecil agar peredaran darah lancar"
          ],
          "correct_answer": "B",
          "explanation": "Posisi tegak condong ke depan mencegah darah mengalir ke tenggorokan, dan tekanan pada hidung membantu pembekuan alami."
        },
        {
          "id": "q3",
          "question": "Gempa terasa cukup kuat saat kamu berada di dalam rumah. Yang sebaiknya kamu lakukan adalah?",
          "options": [
            "A. Langsung lari menuruni tangga atau masuk lift",
            "B. Berdiri di dekat jendela kaca agar mudah terlihat",
            "C. Berlindung di bawah meja yang kokoh sambil melindungi kepala, jauhi kaca dan barang yang bisa jatuh",
            "D. Tetap tidur dan menunggu guncangan berhenti sendiri"
          ],
          "correct_answer": "C",
          "explanation": "Sebagian besar luka saat gempa berasal dari barang jatuh dan pecahan kaca. Berlindung di bawah furnitur kokoh adalah respons teraman di dalam ruangan."
        },
        {
          "id": "q4",
          "question": "Kamu melihat kabel listrik terkelupas di rumah. Sikap yang tepat adalah?",
          "options": [
            "A. Menyentuh kabel untuk memastikan masih ada arusnya",
            "B. Menutup kabel dengan lakban sendiri dalam keadaan listrik menyala",
            "C. Tidak menyentuh kabel, menjauhkan orang lain, dan melapor ke orang dewasa atau teknisi",
            "D. Menyiram kabel dengan air agar tidak panas"
          ],
          "correct_answer": "C",
          "explanation": "Kabel terkelupas bisa menyetrum. Jangan ditangani sendiri; amankan area dan serahkan ke orang yang berwenang."
        },
        {
          "id": "q5",
          "question": "Seseorang pingsan di tempat umum. Tindakan awal yang benar adalah?",
          "options": [
            "A. Mengangkat dan mendudukkannya paksa sambil memberinya minum",
            "B. Mengerumuninya ramai-ramai agar ia cepat sadar",
            "C. Mengecek responsnya, memberi ruang udara, meminta orang menghubungi bantuan darurat, dan menemaninya",
            "D. Meninggalkannya karena pasti akan sadar sendiri"
          ],
          "correct_answer": "C",
          "explanation": "Beri jalan napas dan ruang, pastikan bantuan profesional dihubungi, dan jangan memberi minum kepada orang yang tidak sadar."
        }
      ]
    }'::jsonb
),
(
    'demo-crew-1',
    '60 Detik Jadi Guru Kecil',
    'Rekam video pendek sekitar 60 detik: ajarkan satu hal sederhana yang kamu kuasai, misalnya melipat baju, menyeduh teh, mengikat tali sepatu, atau merapikan tas. Jelaskan langkahnya dengan jelas dan santai. Tidak perlu menyebut data pribadi atau informasi sensitif.',
    'VIDEO',
    'ADMIN_REVIEW',
    2,
    '2026-09-13',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam sendiri lewat kamera depan, santai seperti video story (maksimal 60 detik). Ajarkan:\n1. Satu hal sederhana yang kamu kuasai (misalnya melipat baju, menyeduh teh, mengikat tali sepatu, merapikan tas)\n2. Jelaskan langkahnya satu per satu dengan jelas\n3. Akhiri dengan satu tips agar berhasil\n\nTidak perlu menyebut nama lengkap, alamat, atau data pribadi sensitif lainnya. Bicaralah senatural mungkin!"
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Bekal Sehat Versimu',
    'Siapkan satu piring sarapan atau bekal yang menurutmu cukup sehat dan mengenyangkan. Ambil 1 foto hasilnya dan tulis penjelasan singkat: apa isinya dan kenapa kamu memilihnya. Tidak perlu makanan mahal — yang penting niatnya peduli diri.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-09-13',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Siapkan satu piring sarapan atau bekal sederhana yang menurutmu cukup sehat (misalnya nasi + telur + sayur, atau roti + buah). 2) Ambil atau unggah 1 foto hasilnya (boleh dari kamera langsung atau galeri). 3) Wajib tulis penjelasan singkat di kolom catatan: apa isi piringmu dan kenapa kamu memilihnya. Admin akan menolak submission tanpa foto atau tanpa penjelasan."
    }'::jsonb
);

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '084_tasks_2026_09_13')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
