-- ============================================================
-- Migration 082: Tasks for 12 September 2026 + payout announcement seed
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/scenario/announcement text may be hardcoded
-- in frontend/backend source. Admin can edit every value below through
-- the existing Admin UI (Tasks tab + Pengaturan / announcement section).
-- 1. Deactivates legacy placeholder tasks (ID 535 & 536) on 2026-09-12
--    (ANTIGRAVITY seed rows from migration 052, same pattern as 080/081).
-- 2. Inserts 3 canonical tasks for 2026-09-12:
--    - Step 1: MINI_GAME (Pilih yang Paling Penting) — 6 decision scenarios
--    - Step 2: DOCUMENT_UPLOAD (Bikin Rencana Sederhana)
--    - Step 3: TEXT_RESPONSE (Kalau Rencana Berubah, min 80 / max 1000)
-- 3. Seeds payout announcement via existing odyssey_system_config
--    announcement_* keys (no new table; same config pattern as economy).
-- 4. Does NOT touch tasks 608-620. Does NOT modify payout logic.
-- ============================================================

-- 1. Safely deactivate legacy placeholder tasks on 2026-09-12
UPDATE odyssey_tasks
SET is_active = FALSE
WHERE id IN (535, 536)
  AND active_date = '2026-09-12';

-- 2. Insert 3 new canonical tasks for 2026-09-12
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
    'Pilih yang Paling Penting',
    'Hadapi 6 situasi nyata tentang mengatur prioritas dan membuat pilihan bijak. Setiap keputusan melatih kemampuan membedakan hal penting vs mendesak dan fokus pada yang berdampak terbesar.',
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
            "title": "Situasi 1 — Pagi yang Penuh Tugas",
            "description": "Kamu bangun pagi dan melihat daftar tugas: PR yang dikumpulkan hari ini, kamar berantakan, dan ajakan main game dari teman. Waktu sebelum berangkat terbatas.",
            "options": [
              {
                "id": "a",
                "label": "Kerjakan PR yang dikumpulkan hari ini terlebih dahulu hingga selesai",
                "delta": 20,
                "hint": "Mendahulukan kewajiban dengan tenggat waktu yang jelas"
              },
              {
                "id": "b",
                "label": "Bereskan kamar sebentar lalu kerjakan PR dengan sisa waktu yang ada",
                "delta": 10,
                "hint": "Baik untuk kerapian, namun PR berisiko tidak selesai maksimal"
              },
              {
                "id": "c",
                "label": "Main game dulu bersama teman dan mengerjakan PR nanti jika sempat",
                "delta": 0,
                "hint": "Hiburan menggeser kewajiban utama yang berbatas waktu"
              }
            ]
          },
          {
            "id": "sit_2",
            "title": "Situasi 2 — Uang Saku Terbatas",
            "description": "Kamu memiliki uang saku untuk seminggu. Di saat yang sama ada kebutuhan alat tulis yang habis dan ada jajanan viral yang ingin dicoba bersama teman.",
            "options": [
              {
                "id": "a",
                "label": "Beli alat tulis yang dibutuhkan untuk belajar, sisihkan sisanya sebagai cadangan",
                "delta": 20,
                "hint": "Kebutuhan belajar didahulukan sebelum keinginan sesaat"
              },
              {
                "id": "b",
                "label": "Beli jajanan viral porsi kecil, lalu beli alat tulis yang paling murah",
                "delta": 5,
                "hint": "Masih mempertimbangkan kebutuhan meski tergoda tren"
              },
              {
                "id": "c",
                "label": "Habiskan untuk jajanan viral dan meminjam alat tulis ke teman setiap hari",
                "delta": 0,
                "hint": "Ketergantungan pada orang lain karena keinginan sesaat"
              }
            ]
          },
          {
            "id": "sit_3",
            "title": "Situasi 3 — Belajar vs Hiburan Malam",
            "description": "Besok ada ulangan penting, tetapi serial favoritmu baru saja merilis episode terakhir yang sedang ramai dibicarakan semua teman.",
            "options": [
              {
                "id": "a",
                "label": "Belajar terlebih dahulu dengan fokus, tonton serial setelah ulangan selesai",
                "delta": 20,
                "hint": "Menunda kesenangan demi hasil jangka panjang"
              },
              {
                "id": "b",
                "label": "Nonton satu episode sambil membuka catatan belajar di sampingnya",
                "delta": 5,
                "hint": "Multitasking membuat belajar dan hiburan sama-sama setengah-setengah"
              },
              {
                "id": "c",
                "label": "Maraton semua episode malam ini dan begadang, belajar seadanya menjelang pagi",
                "delta": 0,
                "hint": "Mengorbankan persiapan dan kondisi tubuh sekaligus"
              }
            ]
          },
          {
            "id": "sit_4",
            "title": "Situasi 4 — Dua Janji di Waktu yang Sama",
            "description": "Kamu sudah berjanji membantu orang tua di rumah pada sore hari, lalu teman mengajak pergi ke acara seru di jam yang sama.",
            "options": [
              {
                "id": "a",
                "label": "Tepati janji membantu orang tua, lalu ajak teman bertemu di lain waktu",
                "delta": 20,
                "hint": "Menjaga komitmen yang sudah dibuat terlebih dahulu"
              },
              {
                "id": "b",
                "label": "Minta izin orang tua dengan sopan dan tawarkan mengganti bantuan di pagi atau malam hari",
                "delta": 10,
                "hint": "Bernegosiasi dengan jujur dan tetap bertanggung jawab"
              },
              {
                "id": "c",
                "label": "Pergi diam-diam ke acara teman tanpa memberi kabar ke rumah",
                "delta": 0,
                "hint": "Mengingkari janji dan merusak kepercayaan keluarga"
              }
            ]
          },
          {
            "id": "sit_5",
            "title": "Situasi 5 — Tugas Kelompok yang Macet",
            "description": "Tugas kelompok harus dikumpulkan lusa, tetapi dua anggota belum mengerjakan bagiannya dan grup chat sepi tanpa kabar.",
            "options": [
              {
                "id": "a",
                "label": "Sampaikan pembagian tugas yang jelas dengan tenggat tiap bagian, lalu tawarkan bantuan untuk bagian yang sulit",
                "delta": 20,
                "hint": "Kepemimpinan praktis: kejelasan peran plus dukungan nyata"
              },
              {
                "id": "b",
                "label": "Kerjakan seluruh tugas sendirian agar cepat selesai tanpa berkoordinasi lagi",
                "delta": 5,
                "hint": "Cepat selesai tetapi melelahkan dan tidak melatih kerja sama"
              },
              {
                "id": "c",
                "label": "Biarkan saja dan salahkan anggota lain jika nilai kelompok jelek",
                "delta": 0,
                "hint": "Pasif dan menyalahkan tanpa usaha memperbaiki keadaan"
              }
            ]
          },
          {
            "id": "sit_6",
            "title": "Situasi 6 — Menyusun Rencana Satu Minggu",
            "description": "Minggu depan padat: ulangan, latihan kegiatan, dan acara keluarga. Kamu ingin semua berjalan tanpa ada yang terbengkalai.",
            "options": [
              {
                "id": "a",
                "label": "Tulis semua kegiatan, tandai yang paling penting dan mendesak, lalu susun jadwal harian yang realistis",
                "delta": 20,
                "hint": "Perencanaan tertulis membuat prioritas terlihat dan bisa dijalankan"
              },
              {
                "id": "b",
                "label": "Ingat-ingat saja semua jadwal di kepala dan jalani hari apa adanya",
                "delta": 10,
                "hint": "Berisiko lupa dan kewalahan saat semua datang bersamaan"
              },
              {
                "id": "c",
                "label": "Tidak membuat rencana apa pun dan menangani apa pun yang datang paling akhir",
                "delta": 0,
                "hint": "Hidup reaktif membuat hal penting mudah terlewat"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Bikin Rencana Sederhana',
    'Susun rencana sederhana untuk satu minggu ke depan: tulis tujuan, langkah, dan jadwal harianmu. Tuangkan ke dalam dokumen lalu unggah sebagai bukti.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-09-12',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "xlsx", "txt"],
      "max_file_size_mb": 10,
      "instruction": "Buat dokumen rencana satu minggu: 1) Tulis 2-3 tujuan minggumu. 2) Jabarkan langkah kecil untuk tiap tujuan. 3) Susun jadwal harian yang realistis. 4) Simpan sebagai PDF/DOCX/XLSX/TXT (maks 10 MB) lalu unggah."
    }'::jsonb
),
(
    'demo-crew-1',
    'Kalau Rencana Berubah',
    'Rencana tidak selalu berjalan mulus. Bayangkan satu bagian penting dari rencanamu minggu ini gagal atau berubah mendadak. Refleksikan sikap dan langkah penyesuaianmu secara dewasa.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-09-12',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Bayangkan satu bagian penting dari rencanamu minggu ini gagal atau berubah mendadak. Jelaskan: 1) Apa yang berubah dan apa dampaknya? 2) Apa langkah penyesuaian yang kamu ambil? 3) Apa yang sebaiknya dihindari saat rencana berubah? 4) Pelajaran apa yang kamu petik tentang fleksibilitas?",
      "minimum_characters": 80,
      "maximum_characters": 1000
    }'::jsonb
);

-- 3. Seed payout announcement via existing odyssey_system_config keys.
-- Initial configured content only; Admin can edit/disable/reschedule anytime
-- through Pengaturan > Pengumuman Sistem. Member UI reads it from
-- /api/shop/config (announcement.visible) — never from hardcoded source.
INSERT INTO odyssey_system_config (key, value)
VALUES
    ('announcement_enabled', 'true'),
    ('announcement_title', 'Pengumuman Pencairan'),
    ('announcement_body', 'Pencairan saat ini mengalami penyesuaian jadwal karena sedang dilakukan proses verifikasi dan pengecekan transaksi. Hal ini dilakukan agar setiap pencairan dapat diproses dengan aman dan akurat. Mohon menunggu dan tidak perlu mengajukan ulang pencairan yang sudah masuk. Terima kasih atas pengertiannya.'),
    ('announcement_audience', 'ALL'),
    ('announcement_start_at', '2026-09-12T00:00:00+07:00'),
    ('announcement_end_at', '2026-09-30T23:59:59+07:00'),
    ('announcement_priority', 'normal')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '082_tasks_2026_09_12')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
