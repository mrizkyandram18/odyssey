-- ============================================================
-- Migration 081: Tasks for 11 September 2026
-- 1. Deactivates legacy placeholder tasks (ID 533 & 534)
-- 2. Inserts 3 canonical tasks:
--    - Step 1: MINI_GAME (Prioritas Dulu, Baru Gas)
--    - Step 2: PHOTO_UPLOAD (Bikin Tempatmu Lebih Siap)
--    - Step 3: TEXT_RESPONSE (Kalau Ada Masalah di Tempat Kerja)
-- ============================================================

-- 1. Safely deactivate legacy placeholder tasks on 2026-09-11
UPDATE odyssey_tasks
SET is_active = FALSE
WHERE id IN (533, 534)
  AND active_date = '2026-09-11';

-- 2. Insert 3 new canonical tasks for 2026-09-11
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
    'Prioritas Dulu, Baru Gas',
    'Hadapi 6 situasi nyata di dunia kerja dan kehidupan sehari-hari. Latih kemampuan membedakan hal mendesak vs penting, mengelola distraksi, dan mengambil keputusan praktis dengan bijak.',
    'MINI_GAME',
    'AUTO',
    1,
    '2026-09-11',
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
            "title": "Situasi 1 — Deadline Mendekat vs Pesan Obrolan Santai",
            "description": "Laporan penting harus dikirim dalam waktu 45 menit ke atasan. Tiba-tiba ponselmu bergetar karena pesan obrolan santai di grup teman yang sedang ramai membicarakan topik seru.",
            "options": [
              {
                "id": "a",
                "label": "Tutup notifikasi ponsel, fokus penuh menyelesaikan dan mengecek ulang laporan hingga terkirim",
                "delta": 20,
                "hint": "Prioritas utama terlindungi dari distraksi"
              },
              {
                "id": "b",
                "label": "Buka chat sebentar untuk membalas beberapa kalimat sambil tetap membuka dokumen laporan",
                "delta": 5,
                "hint": "Multitasking memecah fokus dan berisiko membuat laporan telat atau keliru"
              },
              {
                "id": "c",
                "label": "Asyik ikut nimbrung obrolan seru di grup dan menunda laporan hingga 10 menit sebelum batas akhir",
                "delta": 0,
                "hint": "Sangat berisiko melewati deadline dan menghasilkan pekerjaan terburu-buru"
              }
            ]
          },
          {
            "id": "sit_2",
            "title": "Situasi 2 — Rekan Minta Bantuan vs Tugas Sendiri Belum Beres",
            "description": "Rekan kerja meminta bantuanmu untuk merapikan file presentasinya. Di saat yang sama, tugas pribadimu yang berprioritas tinggi dan bermuatan deadline ketat masih tersisa separuh.",
            "options": [
              {
                "id": "a",
                "label": "Jelaskan dengan sopan bahwa tugas utamamu sedang mendesak, lalu tawarkan bantuan setelah tugasmu rampung",
                "delta": 20,
                "hint": "Komunikasi transparan dan tetap bertanggung jawab atas amanah utama"
              },
              {
                "id": "b",
                "label": "Langsung tinggalkan tugas pribadimu demi membantu rekan kerja agar dianggap tidak enakan",
                "delta": 5,
                "hint": "Membantu itu baik, namun menelantarkan tanggung jawab utama dapat merugikan tim"
              },
              {
                "id": "c",
                "label": "Menolak ketus tanpa penjelasan apa pun sehingga menimbulkan ketegangan antar rekan tim",
                "delta": 0,
                "hint": "Merusak relasi profesional tanpa memberikan alternatif solusi"
              }
            ]
          },
          {
            "id": "sit_3",
            "title": "Situasi 3 — Masalah Operasional Mendadak vs Tugas Rutin Harian",
            "description": "Saat kamu sedang mengerjakan rekap data rutin bulanan, terjadi gangguan pada alur layanan pelanggan yang membutuhkan penanganan segera agar operasional tidak mandek.",
            "options": [
              {
                "id": "a",
                "label": "Tunda sementara rekap rutin, tangani gangguan operasional mendesak, lalu lanjutkan rekap setelah normal",
                "delta": 20,
                "hint": "Merespons masalah berdampak besar (urgent & important) terlebih dahulu"
              },
              {
                "id": "b",
                "label": "Tetap fokus pada rekap rutin dan membiarkan gangguan pelanggan menunggu berjam-jam",
                "delta": 5,
                "hint": "Kaku pada rutinitas tanpa peka terhadap eskalasi krisis operasional"
              },
              {
                "id": "c",
                "label": "Panik dan meninggalkan semua pekerjaan tanpa berkoordinasi dengan tim lain",
                "delta": 0,
                "hint": "Reaktif tanpa kendali memperburuk situasi tim"
              }
            ]
          },
          {
            "id": "sit_4",
            "title": "Situasi 4 — Instruksi Tambahan Atasan vs Beban Kerja Sedang Penuh",
            "description": "Atasan memberikan satu tugas mendadak baru di siang hari, padahal jadwalmu hari ini sudah padat dengan target yang disepakati sebelumnya.",
            "options": [
              {
                "id": "a",
                "label": "Sampaikan daftar tugas yang sedang berjalan secara objektif dan tanyakan urutan prioritas yang diinginkan atasan",
                "delta": 20,
                "hint": "Membantu pimpinan menyelaraskan ekspektasi berdasarkan realitas kapasitas"
              },
              {
                "id": "b",
                "label": "Langsung mengiyakan semua tugas tanpa konfirmasi lalu menyelesaikan semuanya secara asal-asalan",
                "delta": 5,
                "hint": "Overpromising berujung pada penurunan mutu kerja dan keterlambatan"
              },
              {
                "id": "c",
                "label": "Mengeluh di media sosial bahwa beban kerjamu tidak manusiawi sebelum berdiskusi dengan atasan",
                "delta": 0,
                "hint": "Tindakan tidak etis yang merusak reputasi profesional diri sendiri"
              }
            ]
          },
          {
            "id": "sit_5",
            "title": "Situasi 5 — Menemukan Detail Minor vs Fitur Kritis Siap Rilis",
            "description": "Dua jam sebelum sistem/dokumen utama dirilis ke klien, kamu menemukan detail estetika minor yang kurang sempurna, sementara fungsi utama sistem sudah lolos uji menyeluruh.",
            "options": [
              {
                "id": "a",
                "label": "Catat temuan minor ke daftar perbaikan tahap berikutnya (backlog) dan amankan jadwal rilis utama",
                "delta": 20,
                "hint": "Prinsip tepat waktu tanpa mengorbankan fungsi utama demi detail kosmetik"
              },
              {
                "id": "b",
                "label": "Membongkar ulang tata letak di menit-menit terakhir hingga rilis terlambat dan berisiko memicu bug baru",
                "delta": 5,
                "hint": "Perfeksionisme tanpa pertimbangan risiko membahayakan komitmen rilis"
              },
              {
                "id": "c",
                "label": "Mengabaikan sepenuhnya dan tidak mencatat apa pun untuk perbaikan di masa depan",
                "delta": 0,
                "hint": "Kehilangan kesempatan untuk evaluasi dan peningkatan mutu berkelanjutan"
              }
            ]
          },
          {
            "id": "sit_6",
            "title": "Situasi 6 — Mengatur Rencana Kerja Pagi Hari",
            "description": "Kamu baru tiba di tempat kerja dan memiliki 5 email, 2 dokumen harus ditandatangani, 1 rapat siang, dan persiapan materi besok. Bagaimana kamu memulai harimu?",
            "options": [
              {
                "id": "a",
                "label": "Luangkan 10 menit memilah matriks prioritas (mendesak/penting), lalu kerjakan tugas berdampak tertinggi terlebih dahulu",
                "delta": 20,
                "hint": "Perencanaan strategis di awal hari melipatgandakan efektivitas kerja"
              },
              {
                "id": "b",
                "label": "Kerjakan apa pun yang paling mudah dulu tanpa melihat tingkat kepentingannya",
                "delta": 10,
                "hint": "Cepat selesai namun berisiko menunda tugas-tugas kritis berbobot besar"
              },
              {
                "id": "c",
                "label": "Bekerja secara acak sesuai email mana yang paling atas masuk ke inbox",
                "delta": 0,
                "hint": "Kerja tanpa arah membuatmu sekadar reaktif terhadap notifikasi orang lain"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Bikin Tempatmu Lebih Siap',
    'Pilih satu tempat yang biasa kamu pakai untuk bekerja, belajar, atau menyimpan barang penting (misalnya meja kerja, meja belajar, rak buku, atau ruang kerja pribadi). Rapikan dan bersihkan agar lebih nyaman dan siap pakai, lalu unggah 1 foto bukti hasilnya.',
    'PHOTO_UPLOAD',
    'AUTO',
    2,
    '2026-09-11',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Pilih satu tempat yang biasa kamu pakai untuk bekerja, belajar, atau menyimpan perlengkapan penting. Rapikan dan tata tempat tersebut agar bersih, teratur, dan nyaman digunakan. Ambil 1 foto bukti kondisi tempat yang sudah rapi lalu unggah sebagai bukti."
    }'::jsonb
),
(
    'demo-crew-1',
    'Kalau Ada Masalah di Tempat Kerja',
    'Bayangkan kamu menghadapi kendala atau kesalahan di tempat kerja yang memperlambat pekerjaanmu atau berdampak pada rekan lain. Refleksikan tindakan profesional yang akan kamu ambil secara bertanggung jawab.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-09-11',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Bayangkan kamu menghadapi masalah di tempat kerja yang menghambat pekerjaanmu atau berdampak pada orang lain. Jelaskan: 1) Apa yang kamu lakukan pertama kali? 2) Siapa yang kamu hubungi/ajak komunikasi? 3) Apa yang sebaiknya dihindari saat menghadapi masalah tersebut? 4) Mengapa kamu memilih langkah tersebut?",
      "minimum_characters": 80,
      "maximum_characters": 1000
    }'::jsonb
);

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '081_tasks_2026_09_11')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
