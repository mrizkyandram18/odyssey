-- ============================================================
-- Migration 083: Replace tasks for 12 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/scenario/instruction text may be hardcoded
-- in frontend/backend source. Admin can edit every value below through
-- the existing Admin UI (Tasks tab).
-- 1. DEACTIVATES (never deletes) exact tasks 624/625/626 on 2026-09-12.
--    Production evidence (2026-09-12 audit): each of 624/625/626 carries
--    1 APPROVED submission + 1 immutable odyssey_coin_transactions row
--    from a real user. DELETE would violate FK history and the immutable
--    ledger trigger, so deactivation is the only safe semantics.
--    Deactivation is naturally idempotent.
-- 2. Inserts 3 fresh canonical tasks for 2026-09-12 (same architecture
--    and config shapes as migrations 080/081/082, no new framework):
--    - Step 1: MINI_GAME (Kalau Kamu Jadi Pengambil Keputusan) — 6 fresh
--      everyday decision scenarios, POINTS currency, target_score 100.
--    - Step 2: PHOTO_UPLOAD (Cari Sesuatu yang Bisa Dibikin Lebih Baik)
--      camera_only = false (gallery/file-picker allowed via existing
--      CameraCaptureModal path). Member writes a short explanation in the
--      existing payload.note field (same convention as DOCUMENT_UPLOAD);
--      ADMIN_REVIEW enforcement is via the existing approve/reject flow.
--      Global/camera-only semantics untouched (task 610 unaffected).
--    - Step 3: VIDEO (60 Detik Jadi Versi Dirimu yang Lebih Percaya Diri)
--      recording.enabled, max_duration_seconds = 60, camera user.
--      Resolves to ADMIN_REVIEW via existing
--      ResolveEvaluationTypeForConfig rule. Same shape as task 616.
-- 3. Idempotency: the pre-clean DELETE below matches ONLY the 3 exact
--    new titles on 2026-09-12 (narrow exact-title guard for dev re-runs,
--    same spirit as migration 052). No broad date/title matching.
-- 4. Does NOT touch tasks 608-620, 535/536, payout/economy config,
--    announcements, RLS, RPCs, triggers, or any other date.
-- ============================================================

-- 1. Deactivate exact legacy tasks for 2026-09-12 (history-preserving).
UPDATE odyssey_tasks
SET is_active = FALSE
WHERE id IN (624, 625, 626)
  AND active_date = '2026-09-12';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact date). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-12'
  AND title IN (
    'Kalau Kamu Jadi Pengambil Keputusan',
    'Cari Sesuatu yang Bisa Dibikin Lebih Baik',
    '60 Detik Jadi Versi Dirimu yang Lebih Percaya Diri'
  );

-- 3. Insert 3 fresh canonical tasks for 2026-09-12.
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
    'Kalau Kamu Jadi Pengambil Keputusan',
    'Hadapi 6 situasi sehari-hari dan pilih tindakan yang menurutmu paling tepat. Setiap pilihan melatih keberanian mengambil keputusan dan melihat konsekuensinya secara jujur.',
    'MINI_GAME',
    'AUTO',
    1,
    '2026-09-12',
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
            "id": "sit_1",
            "title": "Situasi 1 — Teman Minta Bantuan Saat Kamu Sibuk",
            "description": "Kamu sedang mengerjakan urusan penting yang harus selesai hari ini. Tiba-tiba teman dekat meminta bantuanmu untuk hal yang menurutnya mendesak.",
            "options": [
              {
                "id": "a",
                "label": "Jelaskan dengan jujur bahwa urusanmu mendesak, lalu tawarkan waktu spesifik kapan kamu bisa membantu",
                "delta": 20,
                "hint": "Jujur soal prioritas namun tetap peduli dan memberi kepastian"
              },
              {
                "id": "b",
                "label": "Langsung tinggalkan urusanmu dan bantu teman agar tidak disebut tidak peduli",
                "delta": 5,
                "hint": "Niatan baik, tetapi mengorbankan tanggung jawab sendiri tanpa komunikasi"
              },
              {
                "id": "c",
                "label": "Menolak dengan ketus tanpa penjelasan sehingga teman tersinggung",
                "delta": 0,
                "hint": "Menutup pintu komunikasi dan merusak hubungan pertemanan"
              }
            ]
          },
          {
            "id": "sit_2",
            "title": "Situasi 2 — Beli Sekarang atau Simpan Dulu",
            "description": "Kamu melihat barang yang sangat diinginkan sedang diskon. Di sisi lain, uangmu sebenarnya disiapkan untuk kebutuhan penting minggu depan.",
            "options": [
              {
                "id": "a",
                "label": "Tahan keinginan, amankan dulu uang untuk kebutuhan penting, beli hanya jika sisa mencukupi",
                "delta": 20,
                "hint": "Kebutuhan yang sudah direncanakan didahulukan dari keinginan sesaat"
              },
              {
                "id": "b",
                "label": "Beli versi yang paling murah secukupnya, lalu simpan sisanya baik-baik",
                "delta": 10,
                "hint": "Kompromi yang masih menjaga sebagian rencana keuangan"
              },
              {
                "id": "c",
                "label": "Langsung beli mumpung diskon, urusan minggu depan dipikirkan nanti saja",
                "delta": 0,
                "hint": "Diskon terasa hemat padahal mengorbankan kebutuhan yang pasti"
              }
            ]
          },
          {
            "id": "sit_3",
            "title": "Situasi 3 — Salah Paham dengan Seseorang",
            "description": "Seseorang mengucapkan sesuatu yang kamu anggap menyinggung. Kamu belum yakin apakah ia bermaksud buruk atau hanya salah penyampaian.",
            "options": [
              {
                "id": "a",
                "label": "Tenangkan diri, lalu tanyakan maksudnya baik-baik secara langsung dan dengarkan penjelasannya",
                "delta": 20,
                "hint": "Klarifikasi langsung mencegah prasangka yang membesar"
              },
              {
                "id": "b",
                "label": "Langsung membalas dengan nada emosi di chat grup agar ia tahu kamu tersinggung",
                "delta": 5,
                "hint": "Emosi sesaat yang bisa memperkeruh keadaan dan sulit ditarik kembali"
              },
              {
                "id": "c",
                "label": "Diam-diam mendiamkannya berhari-hari sambil menceritakan keburukannya ke orang lain",
                "delta": 0,
                "hint": "Menghindari masalah sambil merusak reputasi orang lain"
              }
            ]
          },
          {
            "id": "sit_4",
            "title": "Situasi 4 — Kesempatan Menarik tetapi Waktu Terbatas",
            "description": "Ada kegiatan menarik yang hanya tersedia sore ini selama satu jam, misalnya sesi berbagi ilmu gratis. Namun kamu sudah punya janji lain di jam yang berdekatan.",
            "options": [
              {
                "id": "a",
                "label": "Cek jadwal dengan teliti, ikuti bagian kegiatan yang sempat dihadiri, dan beri kabar baik ke pihak janji lainmu",
                "delta": 20,
                "hint": "Keputusan sadar dengan komunikasi yang menjaga semua pihak"
              },
              {
                "id": "b",
                "label": "Paksakan mengikuti keduanya sekaligus meski harus terburu-buru dan setengah fokus",
                "delta": 5,
                "hint": "Serakah waktu membuat kedua kegiatan tidak maksimal"
              },
              {
                "id": "c",
                "label": "Bimbang sampai menit terakhir hingga akhirnya tidak mengikuti apa pun",
                "delta": 0,
                "hint": "Keraguan tanpa keputusan menghilangkan semua kesempatan"
              }
            ]
          },
          {
            "id": "sit_5",
            "title": "Situasi 5 — Menemukan Barang Tertinggal",
            "description": "Di tempat umum kamu menemukan dompet yang tertinggal. Di dalamnya ada uang dan kartu identitas pemiliknya yang kamu kenali.",
            "options": [
              {
                "id": "a",
                "label": "Amankan dompet itu lalu kembalikan langsung ke pemiliknya atau titipkan ke petugas dengan catatan jelas",
                "delta": 20,
                "hint": "Tanggung jawab penuh atas barang orang lain yang dipercayakan keadaan padamu"
              },
              {
                "id": "b",
                "label": "Biarkan saja di tempatnya dengan harapan pemiliknya kembali sendiri",
                "delta": 5,
                "hint": "Pasif dan membiarkan risiko barang hilang diambil orang lain"
              },
              {
                "id": "c",
                "label": "Ambil uang di dalamnya lalu buang dompetnya agar tidak ketahuan",
                "delta": 0,
                "hint": "Memanfaatkan keadaan untuk keuntungan pribadi yang merugikan orang lain"
              }
            ]
          },
          {
            "id": "sit_6",
            "title": "Situasi 6 — Merusak Barang Pinjaman",
            "description": "Kamu tidak sengaja merusak barang yang dipinjam dari teman. Kerusakannya terlihat jelas dan tidak mungkin disembunyikan lama.",
            "options": [
              {
                "id": "a",
                "label": "Akui dengan jujur, minta maaf, dan tawarkan mengganti atau memperbaiki bersama",
                "delta": 20,
                "hint": "Keberanian mengakui kesalahan adalah fondasi kepercayaan"
              },
              {
                "id": "b",
                "label": "Kembalikan diam-diam dan berharap teman tidak menyadarinya",
                "delta": 5,
                "hint": "Menghindari tanggung jawab dengan mempertaruhkan kepercayaan"
              },
              {
                "id": "c",
                "label": "Bersikeras barang itu sudah rusak sejak awal dan menyalahkan orang lain",
                "delta": 0,
                "hint": "Kebohongan untuk menyelamatkan diri yang menghancurkan kepercayaan"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Cari Sesuatu yang Bisa Dibikin Lebih Baik',
    'Lihat sekitarmu dan temukan satu hal kecil yang bisa dibuat lebih baik: meja berantakan, kabel tidak rapi, tempat penyimpanan kurang teratur, atau sudut rumah yang bisa dibuat lebih nyaman. Ambil 1 foto kondisinya dan tulis penjelasan singkat tentang apa yang bisa diperbaiki beserta alasannya.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-09-12',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Temukan satu hal di lingkungan sekitarmu yang menurutmu bisa dibuat lebih baik (misalnya meja berantakan, kabel tidak rapi, atau sudut rumah yang kurang nyaman). 2) Ambil atau unggah 1 foto kondisi tersebut (boleh dari kamera langsung atau galeri). 3) Wajib tulis penjelasan singkat di kolom catatan: apa yang bisa diperbaiki dan mengapa hal itu layak dibuat lebih baik. Admin akan menolak submission tanpa foto atau tanpa penjelasan."
    }'::jsonb
),
(
    'demo-crew-1',
    '60 Detik Jadi Versi Dirimu yang Lebih Percaya Diri',
    'Rekam video pendek sekitar 60 detik: ceritakan satu hal yang sekarang sudah lebih bisa kamu lakukan dibanding beberapa tahun lalu. Boleh tentang kerja, belajar, komunikasi, mengatur uang, membantu keluarga, mengatasi masalah, atau skill sehari-hari. Tidak perlu menyebut data pribadi atau informasi sensitif.',
    'VIDEO',
    'ADMIN_REVIEW',
    3,
    '2026-09-12',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam sendiri lewat kamera depan, santai seperti video story (maksimal 60 detik). Ceritakan:\n1. Satu hal yang sekarang sudah lebih bisa kamu lakukan dibanding beberapa tahun lalu\n2. Boleh pilih topik: kerja, belajar, komunikasi, mengatur uang, membantu keluarga, mengatasi masalah, atau skill sehari-hari\n3. Ceritakan proses singkatnya: dulu seperti apa, sekarang seperti apa\n\nTidak perlu menyebut nama lengkap, alamat, atau data pribadi sensitif lainnya. Bicaralah senatural mungkin!"
      }
    }'::jsonb
);

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '083_tasks_replace_2026_09_12')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
