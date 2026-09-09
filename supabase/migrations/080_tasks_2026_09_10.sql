-- ============================================================
-- Migration 080: Tasks for 10 September 2026
-- 1. Deactivates legacy placeholder tasks (ID 531 & 532)
-- 2. Inserts 3 canonical tasks:
--    - Step 1: MINI_GAME (Pilih Jalanmu)
--    - Step 2: QUIZ (Seberapa Siap Kamu di Dunia Kerja?)
--    - Step 3: TEXT_RESPONSE (Kalau Besok Harus Mandiri)
-- ============================================================

-- 1. Safely deactivate legacy placeholder tasks on 2026-09-10
UPDATE odyssey_tasks
SET is_active = FALSE
WHERE id IN (531, 532)
  AND active_date = '2026-09-10';

-- 2. Insert 3 new canonical tasks for 2026-09-10
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
    'Pilih Jalanmu',
    'Hadapi 6 situasi nyata sehari-hari dengan saldo virtual awal Rp500.000. Setiap pilihan memiliki konsekuensi nyata terhadap kondisi finansialmu. Buat keputusan terbaik dan capai hasil yang bijak!',
    'MINI_GAME',
    'AUTO',
    1,
    '2026-09-10',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "game": "DECISION_FINANCE",
      "target_score": 100,
      "scenario": {
        "currency": "IDR",
        "initial_balance": 500000,
        "events": [
          {
            "id": "sit_1",
            "title": "Situasi 1 — Ajakan Nongkrong Saat Anggaran Pas-Pasan",
            "description": "Teman-teman mengajak kumpul di kafe kekinian. Kamu ingin menjaga pertemanan, tetapi saldo mingguanmu harus dihemat.",
            "options": [
              {
                "id": "a",
                "label": "Ikut nongkrong dan memesan paket makanan lengkap (-Rp60.000)",
                "delta": -60000
              },
              {
                "id": "b",
                "label": "Ikut kumpul hanya memesan minuman terjangkau (-Rp20.000)",
                "delta": -20000
              },
              {
                "id": "c",
                "label": "Menolak halus dan ajak bertemu di ruang publik gratis (Rp0)",
                "delta": 0
              }
            ]
          },
          {
            "id": "sit_2",
            "title": "Situasi 2 — Investasi Diri & Buku Panduan Skill Kerja",
            "description": "Ada kesempatan membeli buku panduan skill kerja praktis yang sedang diskon dan sangat berguna untuk melamar kerja.",
            "options": [
              {
                "id": "a",
                "label": "Beli buku panduan kerja praktis untuk investasi ilmu (-Rp50.000)",
                "delta": -50000
              },
              {
                "id": "b",
                "label": "Cari modul dan tutorial gratis di internet/perpustakaan (Rp0)",
                "delta": 0
              },
              {
                "id": "c",
                "label": "Tunda belajar dan gunakan uang untuk jajan santai (-Rp35.000)",
                "delta": -35000
              }
            ]
          },
          {
            "id": "sit_3",
            "title": "Situasi 3 — Sepatu Kerja Mulai Rusak vs Model Terbaru",
            "description": "Sepatu yang biasa kamu pakai solnya mulai terbuka. Di toko ada sepatu kerja standar yang awet dan sepatu branded model terbaru yang mahal.",
            "options": [
              {
                "id": "a",
                "label": "Beli sepatu kerja standar yang rapi, awet, dan fungsional (-Rp120.000)",
                "delta": -120000
              },
              {
                "id": "b",
                "label": "Beli sepatu branded mahal demi gengsi (-Rp250.000)",
                "delta": -250000
              },
              {
                "id": "c",
                "label": "Bawa ke tukang sol sepatu untuk diperbaiki dulu (-Rp25.000)",
                "delta": -25000
              }
            ]
          },
          {
            "id": "sit_4",
            "title": "Situasi 4 — Peluang Bantuan Toko Tetangga",
            "description": "Seorang tetangga butuh bantuan membukukan stok dagangan dan merapikan barang selama setengah hari akhir pekan.",
            "options": [
              {
                "id": "a",
                "label": "Ambil tawaran merapikan stok dagangan toko 3 jam (+Rp90.000)",
                "delta": 90000
              },
              {
                "id": "b",
                "label": "Tolak karena ingin menghabiskan waktu rebahan seharian (Rp0)",
                "delta": 0
              }
            ]
          },
          {
            "id": "sit_5",
            "title": "Situasi 5 — Biaya Darurat: Ban Bocor & Rem Aus",
            "description": "Saat perjalanan, ban motor kempes terkena paku dan rem terasa aus sehingga perlu diservis agar aman di jalan.",
            "options": [
              {
                "id": "a",
                "label": "Tambal ban dan ganti kampas rem demi keselamatan berkendara (-Rp45.000)",
                "delta": -45000
              },
              {
                "id": "b",
                "label": "Hanya tambal ban dan tunda servis rem yang aus (berisiko) (-Rp15.000)",
                "delta": -15000
              },
              {
                "id": "c",
                "label": "Ganti pelek dan variasi baru yang tidak mendesak (-Rp160.000)",
                "delta": -160000
              }
            ]
          },
          {
            "id": "sit_6",
            "title": "Situasi 6 — Komitmen Menabung & Mengatur Sisa Saldo",
            "description": "Setelah melewati berbagai situasi, kamu melihat saldo tersisa. Apa komitmen finansial yang kamu ambil sekarang?",
            "options": [
              {
                "id": "a",
                "label": "Kunci 20% sisa uang ke pos tabungan darurat terpisah (-Rp50.000)",
                "delta": -50000
              },
              {
                "id": "b",
                "label": "Habiskan sisa uang untuk belanja hiburan tanpa rencana (-Rp100.000)",
                "delta": -100000
              },
              {
                "id": "c",
                "label": "Simpan sisa saldo sebagai cadangan tunai tanpa belanja impulsif (Rp0)",
                "delta": 0
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Seberapa Siap Kamu di Dunia Kerja?',
    'Uji kesiapanmu menghadapi situasi nyata di tempat kerja: komunikasi profesional, manajemen deadline, menerima kritik, teamwork, meminta bantuan, hingga wawancara.',
    'QUIZ',
    'AUTO',
    2,
    '2026-09-10',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "questions": [
        {
          "id": "q1",
          "question": "Kamu diberi dua tugas penting oleh atasan dengan tenggat waktu (deadline) yang sama di sore hari, dan kamu sadar tidak mungkin menyelesaikan keduanya sendirian. Apa tindakan yang paling profesional?",
          "options": [
            "A. Mengabari atasan lebih awal, menjelaskan estimasi waktu, dan meminta arahan tugas mana yang harus diprioritaskan",
            "B. Mengerjakan salah satu tugas diam-diam dan membiarkan tugas lainnya melewati deadline tanpa kabar",
            "C. Menyelesaikan keduanya secara tergesa-gesa meskipun kualitasnya berantakan dan banyak kesalahan",
            "D. Mengeluh kepada rekan kerja bahwa pembagian beban kerjamu tidak adil"
          ],
          "correct_answer": "A",
          "explanation": "Komunikasi proaktif sebelum deadline tiba jauh lebih profesional daripada membiarkan pekerjaan terbengkalai. Atasan dapat membantu menentukan prioritas atau membagi beban kerja."
        },
        {
          "id": "q2",
          "question": "Saat evaluasi hasil kerja, atasan menyampaikan kritik bahwa laporanmu kurang teliti dan ada data yang keliru. Bagaimana respon terbaikmu?",
          "options": [
            "A. Mendengarkan dengan tenang, mencatat poin perbaikan, dan menanyakan saran konkret untuk memperbaikinya",
            "B. Langsung membantah dan menyalahkan rekan kerja yang memberikan data awal",
            "C. Merasa tersinggung dan mendiamkan atasan selama beberapa hari ke depan",
            "D. Mengabaikan masukan tersebut karena merasa cara kerjamu sudah paling benar"
          ],
          "correct_answer": "A",
          "explanation": "Kritik di tempat kerja adalah bahan evaluasi untuk tumbuh. Menerima dengan tenang dan mencari solusi konkret mencerminkan kedewasaan dan profesionalisme."
        },
        {
          "id": "q3",
          "question": "Dalam satu tim proyek, salah satu rekan kerjamu lambat merespons sehingga menghambat kemajuan pekerjaan bersama. Apa langkah pertama yang sebaiknya kamu ambil?",
          "options": [
            "A. Mengajaknya berbicara secara empat mata dengan sopan untuk memahami kendalanya dan mencari solusi bersama",
            "B. Menyindir rekan tersebut di grup chat tim agar merasa malu dan segera menyelesaikan bagiannya",
            "C. Melaporkannya langsung ke pimpinan puncak tanpa konfirmasi terlebih dahulu",
            "D. Mengambil alih semua pekerjaannya sambil menggerutu di belakang rekan tersebut"
          ],
          "correct_answer": "A",
          "explanation": "Pendekatan empat mata secara profesional memungkinkan kita memahami akar masalah tanpa merusak hubungan kerja tim."
        },
        {
          "id": "q4",
          "question": "Suatu pagi kamu mendadak sakit dan tidak bisa masuk bekerja. Prosedur etika kerja apa yang wajib kamu lakukan pertama kali?",
          "options": [
            "A. Menghubungi atasan langsung sebelum jam kerja dimulai, mengabarkan kondisi, dan menginfokan tugas mendesak hari itu",
            "B. Mematikan ponsel dan baru memberikan penjelasan saat sudah sembuh beberapa hari kemudian",
            "C. Hanya mengunggah status sedang sakit di media sosial pribadi",
            "D. Menitipkan pesan kepada teman kerja tanpa menghubungi atasan langsung"
          ],
          "correct_answer": "A",
          "explanation": "Memberi kabar sebelum jam operasional dimulai adalah bentuk tanggung jawab kerja agar tim dapat mengantisipasi tugas-tugas mendesakmu."
        },
        {
          "id": "q5",
          "question": "Kamu menghadapi kendala teknis saat mengerjakan tugas baru yang belum pernah kamu lakukan, padahal panduan tertulis sudah dibaca. Apa yang sebaiknya dilakukan?",
          "options": [
            "A. Mencoba memecahkan masalah terlebih dahulu, lalu bertanya kepada rekan senior dengan pertanyaan terstruktur dan menunjukkan apa yang sudah dicoba",
            "B. Langsung meminta rekan kerja lain mengerjakan seluruh tugasmu dari awal",
            "C. Pura-pura mengerti dan menunggu sampai ditegur oleh atasan",
            "D. Menghentikan pekerjaan dan meninggalkan tugas tersebut tanpa penjelasan"
          ],
          "correct_answer": "A",
          "explanation": "Menunjukkan usaha mandiri terlebih dahulu lalu bertanya secara terstruktur menunjukkan inisiatif, menghargai waktu rekan senior, dan komitmen belajar."
        },
        {
          "id": "q6",
          "question": "Saat wawancara kerja, pewawancara menanyakan apa kelemahan terbesarmu. Jawaban mana yang paling tepat dan profesional?",
          "options": [
            "A. Menyebutkan kelemahan nyata yang sedang kamu perbaiki beserta langkah konkret yang sudah kamu lakukan",
            "B. Menjawab bahwa kamu adalah orang yang sempurna dan tidak memiliki kelemahan apa pun",
            "C. Mengatakan bahwa kamu sering malas dan tidak suka bangun pagi tanpa ada penjelasan lanjut",
            "D. Menolak menjawab karena menganggap kelemahan adalah rahasia pribadi"
          ],
          "correct_answer": "A",
          "explanation": "Pewawancara ingin melihat kejujuran, kesadaran diri (self-awareness), dan kemauanmu untuk terus belajar memperbaiki diri."
        }
      ]
    }'::jsonb
),
(
    'demo-crew-1',
    'Kalau Besok Harus Mandiri',
    'Bayangkan mulai besok kamu harus mengatur hidupmu sendiri secara mandiri tanpa bergantung pada orang lain. Refleksikan kemampuan dasar yang paling kamu butuhkan untuk bertahan dan berkembang.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-09-10',
    30,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Bayangkan mulai besok kamu harus mengatur hidupmu sendiri. Sebutkan 3 hal yang menurutmu paling perlu kamu kuasai, jelaskan kenapa hal itu penting, dan tuliskan satu langkah kecil yang bisa kamu mulai minggu ini.",
      "minimum_characters": 80,
      "maximum_characters": 1500
    }'::jsonb
);

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '080_tasks_2026_09_10')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
