-- ============================================================
-- Migration 087: Replace tasks for 17 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction text may be hardcoded
-- in frontend/backend source. Admin can edit every value below through
-- the existing Admin UI (Tasks tab).
-- 1. DELETES exact tasks 544/545 on 2026-09-17. Production evidence
--    (audit at apply time): both carry 0 submissions in
--    odyssey_task_submissions and 0 linked rows in
--    odyssey_coin_transactions (reference_id), so exact-ID hard-delete
--    is safe. No other rows exist on 2026-09-17. History of other
--    dates is untouched.
-- 2. Inserts 3 fresh canonical tasks for 2026-09-17 forming one
--    practical progression (existing architecture and config shapes
--    only, no new framework, no new tables, no shopping integration,
--    no product API, no new upload system):
--    - Step 1: PHOTO_UPLOAD (Benerin Sesuatu yang Sering Bikin Ribet)
--      — member finds ONE real everyday friction point (home, work,
--      or surroundings), uploads max 1 photo of the actual condition,
--      and writes a short note: what makes it ribet, why it is a
--      problem, and a simple fix idea. camera_only = false (gallery/
--      file-picker allowed via existing CameraCaptureModal path);
--      note via existing payload.note field (same convention as
--      DOCUMENT_UPLOAD and task 628); ADMIN_REVIEW via the existing
--      approve/reject flow. No physical change or purchase required.
--    - Step 2: TEXT_RESPONSE (Bandingkan Sebelum Beli) — member picks
--      ONE real item they might actually buy, researches 2-3 real
--      options themselves, and reports: item name, each option with
--      price + key features + pros/cons, extra costs if relevant,
--      final choice + reason tied to their need. minimum 200 chars
--      so a bare item name cannot pass (server-enforced by
--      odyssey_submit_manual_task from config); ADMIN_REVIEW. No
--      purchase required, no personal data, URLs are plain text
--      evidence (no new URL validator).
--    - Step 3: VIDEO (Jelaskan Sampai Orang Lain Paham) — member
--      records ~60 seconds teaching one thing they genuinely know or
--      do often, as if explaining to a beginner. recording.enabled,
--      max_duration_seconds = 60, camera user. Resolves to ADMIN_REVIEW
--      via existing ResolveEvaluationTypeForConfig rule (same shape as
--      tasks 616, 629, and 631). Face need not be visible. Uses the
--      existing VIDEO submission infrastructure only.
-- 3. Idempotency: the pre-clean DELETE below matches ONLY the 3 exact
--    new titles on 2026-09-17 (narrow exact-title guard for dev re-runs,
--    same spirit as migrations 052/083/084/085/086). No broad
--    date/title matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), 2026-09-13
--    (630/631/632), 2026-09-15 (633/634/635), 2026-09-16 (636/637/638),
--    tasks 608-620 (610 camera_only stays true), payout/economy config,
--    announcements, RLS, RPCs, triggers, or any other date.
-- ============================================================

-- 1. Remove exact legacy tasks for 2026-09-17 (proven dependency-free).
DELETE FROM odyssey_tasks
WHERE id IN (544, 545)
  AND active_date = '2026-09-17';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact date). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-17'
  AND title IN (
    'Benerin Sesuatu yang Sering Bikin Ribet',
    'Bandingkan Sebelum Beli',
    'Jelaskan Sampai Orang Lain Paham'
  );

-- 3. Insert 3 fresh canonical tasks for 2026-09-17.
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
    'Benerin Sesuatu yang Sering Bikin Ribet',
    'Temukan SATU hal kecil yang nyata di rumah, tempat kerja, atau lingkungan sekitarmu yang membuat aktivitas sehari-hari jadi kurang praktis atau kurang nyaman — misalnya tumpukan barang yang menghalangi jalan, kabel berbelit yang susah dirapikan, wadah tanpa label yang bikin salah ambil, atau sudut ruangan yang gelap dan rawan. Ambil 1 foto kondisi aslinya dan jelaskan singkat: apa yang membuatnya ribet, kenapa itu masalah bagimu, dan ide sederhana untuk memperbaikinya. Tidak perlu mengubah apa pun secara fisik dan tidak perlu membeli sesuatu — cukup amati, foto, dan jelaskan idemu.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-09-17',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Temukan satu masalah NYATA yang bisa difoto di lingkungan sekitarmu — sesuatu yang benar-benar membuat aktivitas jadi ribet atau tidak nyaman (bukan sekadar foto benda biasa). 2) Ambil atau unggah 1 foto kondisi tersebut (boleh dari kamera langsung atau galeri). 3) Wajib tulis penjelasan singkat di kolom catatan dengan 3 bagian: (a) apa yang membuatnya ribet, (b) kenapa menurutmu itu masalah, (c) ide sederhana untuk memperbaikinya. Tidak wajib mengubah atau membeli apa pun. Admin akan menolak submission tanpa foto, tanpa penjelasan 3 bagian, atau foto yang tidak menunjukkan masalah yang bisa diamati."
    }'::jsonb
),
(
    'demo-crew-1',
    'Bandingkan Sebelum Beli',
    'Pilih SATU barang yang memang mungkin kamu beli dalam kehidupan nyata — misalnya charger, sepatu, alat kerja, perlengkapan rumah, aksesori HP, atau kebutuhan sehari-hari lainnya. Riset sendiri 2-3 pilihan NYATA barang tersebut (boleh dari toko, marketplace, atau katalog), lalu tulis hasil perbandinganmu: nama barang, setiap pilihan beserta harga dan fungsi/fitur utama, kelebihan dan kekurangan masing-masing, biaya tambahan jika ada, pilihan akhirmu, dan alasanmu berdasarkan kebutuhanmu. Tidak perlu membeli apa pun — cukup teliti sebelum memutuskan.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    2,
    '2026-09-17',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis hasil perbandinganmu dengan format:\n1) Nama barang yang ingin dibeli:\n2) Pilihan 1 — nama/merk, harga, fungsi atau fitur utama, kelebihan, kekurangan:\n3) Pilihan 2 — nama/merk, harga, fungsi atau fitur utama, kelebihan, kekurangan:\n4) (Opsional) Pilihan 3 dengan format yang sama:\n5) Biaya tambahan jika ada (misalnya ongkir, garansi, aksesori):\n6) Pilihan akhirmu:\n7) Alasan pilihanmu berdasarkan kebutuhanmu (2-3 kalimat):\n\nWajib membandingkan minimal 2 pilihan nyata dengan harga masing-masing. Boleh mencantumkan link atau sumber sebagai teks biasa. Jangan cantumkan data pribadi. Admin akan menolak jawaban tanpa harga, dengan kurang dari 2 pilihan, tanpa pilihan akhir, atau tanpa alasan.",
      "minimum_characters": 200,
      "maximum_characters": 2000
    }'::jsonb
),
(
    'demo-crew-1',
    'Jelaskan Sampai Orang Lain Paham',
    'Pilih SATU hal yang benar-benar kamu pahami atau sering kamu lakukan — misalnya cara melipat baju dengan cepat, cara menyeduh minuman favoritmu, cara merapikan file di HP, tips mengatur uang belanja, cara melakukan bagian dari pekerjaanmu, atau keahlian praktis lainnya. Rekam video sekitar 60 detik seolah-olah kamu menjelaskannya kepada seseorang yang belum memahaminya sama sekali: tunjukkan atau jelaskan langkahnya satu per satu dengan bahasa sederhana. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya sambil menjelaskan dengan suaramu.',
    'VIDEO',
    'ADMIN_REVIEW',
    3,
    '2026-09-17',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik, santai seperti video story. Jelaskan SATU hal yang benar-benar kamu kuasai dari pengalamanmu sendiri (misalnya demonstrasi sederhana, tutorial singkat, penjelasan proses, tips praktis, atau cara melakukan sesuatu):\n1. Sebutkan dulu hal apa yang akan kamu jelaskan\n2. Jelaskan langkahnya satu per satu dengan bahasa sederhana, seolah pendengarmu belum tahu sama sekali\n3. Akhiri dengan satu tips agar berhasil\n\nWajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya sambil menjelaskan. Bukan kuis dan tidak perlu topik akademik. Jangan menyebut data pribadi sensitif. Bicaralah senatural mungkin!"
      }
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '087_tasks_replace_2026_09_17')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
