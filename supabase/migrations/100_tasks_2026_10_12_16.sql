-- ============================================================
-- Migration 100: Fresh tasks for 12-16 October 2026 (15 tasks)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction/question/scenario text may be
-- hardcoded in frontend/backend source. Admin can edit every value below
-- through the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. No legacy rows exist on 2026-10-12..2026-10-16 (verified live:
--    zero rows on all five dates; pre-apply script re-audits live and
--    ABORTS if any row carries submissions/ledger). odyssey_claims has
--    no task_id column, so no claim linkage is possible. Fresh INSERT
--    only; no delete/deactivate of other dates.
-- 2. Inserts 15 fresh tasks (3 per date, step 1-3, existing
--    architecture and config shapes only, no new framework, no new
--    tables, no new engines, no new upload system, no economy change):
--    2026-10-12:
--    - Step 1 PHOTO_UPLOAD (Rekam Satu Detail Kecil lewat Foto) -
--      DIVERGENCE NOTE vs task 650 (Cari Hal yang Sering Diabaikan):
--      650 asks for a neglected problem spot needing attention; this
--      task asks for a close-up DETAIL on one familiar object (texture,
--      pola aus, sambungan, tanda pakai) plus why it is often missed.
--      Observation, not repair. max_files 1, camera_only = false;
--      ADMIN_REVIEW. Reward 20c/100xp.
--    - Step 2 MINI_GAME (Tentukan Langkah Paling Tepat) - DIVERGENCE
--      NOTE vs tasks 666 (awal-bulan), 670 (pre-action checks), 673
--      (reasonable choice under constraints), 676 (which-step-first),
--      648 (unexpected events): this set is about the single most
--      correct IMMEDIATE action (lampu mati, galon habis, kunci
--      nyangkut, paket datang, pinjaman barang, colokan penuh), each
--      with 3 options + deterministic delta + hint. DECISION_PRIORITY
--      + POINTS, initial 0, target 100. AUTO via existing
--      ValidateDecisionChoices server recompute (client score never
--      trusted). Reward 25c/100xp.
--    - Step 3 TEXT_RESPONSE (Uraikan Satu Hal dengan Bahasamu Sendiri)
--      - DIVERGENCE NOTE vs tasks 641/653 (video explanations): this
--      is WRITTEN explanation with a required analogy + example for a
--      beginner friend. min 150, max 800 (server-enforced P0008 from
--      config); ADMIN_REVIEW. Reward 20c/100xp.
--    2026-10-13:
--    - Step 1 VIDEO (Peragakan Cara Memakai Satu Benda) - DIVERGENCE
--      NOTE vs tasks 669 (trik) and 672 (cara lebih praktis): this is
--      BASIC CORRECT USE of one everyday object (not a trick, not an
--      improvement). recording.enabled, max_duration_seconds = 60,
--      camera user; ADMIN_REVIEW. Reward 20c/100xp.
--    - Step 2 TEXT_RESPONSE (Temukan Fungsi Lain Satu Barang) - member
--      picks one item, states its main function, then ONE alternative
--      use + how + limits. min 150, max 800; ADMIN_REVIEW. 20c/100xp.
--    - Step 3 DOCUMENT_UPLOAD (Susun Lembar Panduan untuk Teman Baru)
--      - DIVERGENCE NOTE vs task 677 (general short guide) and 671
--      (checklist): this is a BEGINNER sheet (assumes zero knowledge,
--      explains terms, lists common mistakes). Checklist-only
--      documents ([ ] items) are rejected. allowed_extensions +
--      accepted_extensions (dual-key intentional), max_file_size_mb =
--      4; ADMIN_REVIEW. Reward 20c/100xp.
--    2026-10-14:
--    - Step 1 TEXT_RESPONSE (Timbang Dua Opsi lalu Pilih Satu) -
--      DIVERGENCE NOTE vs task 640 (pre-purchase research of 2-3 real
--      products): this is ANY everyday choice between two options,
--      compared on 2 criteria, ending with a choice + reason. Not a
--      purchase. min 150, max 800; ADMIN_REVIEW. 20c/100xp.
--    - Step 2 MINI_GAME (Atasi Keterbatasan dengan yang Ada) -
--      DIVERGENCE NOTE vs task 673 (constraint situations like
--      baterai/tas/ajakan/bahan/hujan): this set is strictly about
--      creative WORKAROUNDS with what is at hand, no buying (tali
--      putus, paku kurang, wadah kurang, cahaya redup, air sedikit,
--      kantong sobek). DECISION_PRIORITY + POINTS, initial 0, target
--      100. AUTO via existing server recompute. 25c/100xp.
--    - Step 3 DOCUMENT_UPLOAD (Tulis Lembar Catatan yang Bisa Dibuka
--      Lagi) - DIVERGENCE NOTE vs 671 (checklist) and 677 (guide):
--      this is a reusable REFERENCE NOTE (rumus, langkah ringkas,
--      daftar istilah, template) the member can reopen later.
--      Checklist-only or multi-page essay documents are rejected.
--      Dual-key extensions, 4MB; ADMIN_REVIEW. 20c/100xp.
--    2026-10-15:
--    - Step 1 VIDEO (Ajarkan Satu Keterampilan Sederhana) - DIVERGENCE
--      NOTE vs task 659 (video story) and 09-19 skill tasks: this is a
--      foundational MICRO-SKILL for a beginner (melipat, mengikat,
--      mencatat, menyusun), demo + steps + common mistake. 60s max;
--      ADMIN_REVIEW. Reward 20c/100xp.
--    - Step 2 PHOTO_UPLOAD (Abadikan Satu Perubahan Kecil Hari Ini) -
--      DIVERGENCE NOTE vs task 667 (member's own before->action->
--      after): this is OBSERVATION of one small change noticed today
--      (plant growth, repaired item, rearranged corner), not
--      necessarily the member's own action. max_files 1, camera_only
--      = false; ADMIN_REVIEW. Reward 20c/100xp.
--    - Step 3 TEXT_RESPONSE (Bagikan Satu Pengalaman dan Pelajaran
--      Praktis) - DIVERGENCE NOTE vs task 659 (video storytelling):
--      WRITTEN structure (kejadian / pelajaran / penerapan) with a
--      required next-application step. min 150, max 800; ADMIN_REVIEW.
--      Reward 20c/100xp.
--    2026-10-16:
--    - Step 1 DOCUMENT_UPLOAD (Tuangkan Satu Ide Jadi Rencana Singkat)
--      - one IDEA turned into a one-page ACTIONABLE plan (masalah /
--      ide / langkah pertama). Vague essays without an actionable
--      first step and checklist-only docs are rejected. Dual-key
--      extensions, 4MB; ADMIN_REVIEW. Reward 20c/100xp.
--    - Step 2 MINI_GAME (Pilih Tanggapan Terbaik Sehari-hari) -
--      DIVERGENCE NOTE vs task 648 (unexpected events: tumpahan,
--      antrean, hujan, dompet, teman hilang): this set is SOCIAL
--      everyday responses with fresh specifics (kembalian lebih, salah
--      kirim pesan, teman terlambat, tumpahan di tempat umum, diminta
--      tolong saat sibuk, parkir sempit). DECISION_PRIORITY + POINTS,
--      initial 0, target 100. AUTO via existing server recompute.
--      Reward 25c/100xp.
--    - Step 3 PHOTO_UPLOAD (Pamerkan Satu Hasil Buatan atau
--      Perbaikanmu) - DIVERGENCE NOTE vs task 639 (fix an annoying
--      thing) and 667: this shows the member's OWN creation or fix
--      result (dibuat atau diperbaiki sendiri) with photo + 3-part
--      note. max_files 1, camera_only = false; ADMIN_REVIEW. Reward
--      20c/100xp.
-- 3. Idempotency: the pre-clean DELETEs below match ONLY the 15 exact
--    new titles on their exact dates (narrow exact-title guard for dev
--    re-runs, same spirit as migrations 083-099). No broad date/title
--    matching.
-- 4. Does NOT touch tasks 608-677 (610 camera_only stays true),
--    payout/economy config, announcements, RLS, RPCs, triggers,
--    claims, shop, or any other date. Batch total 315 coins
--    (12x20 + 3x25); task batch brief states 310 but the per-task
--    rewards it lists (20 normal / 25 challenging x3) sum to 315 -
--    per-task values win, applied exactly as specified. Economy
--    (cap/target/rate/payout) unchanged.
-- ============================================================

-- 1. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact dates). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-12'
  AND title IN (
    'Rekam Satu Detail Kecil lewat Foto',
    'Tentukan Langkah Paling Tepat',
    'Uraikan Satu Hal dengan Bahasamu Sendiri'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-13'
  AND title IN (
    'Peragakan Cara Memakai Satu Benda',
    'Temukan Fungsi Lain Satu Barang',
    'Susun Lembar Panduan untuk Teman Baru'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-14'
  AND title IN (
    'Timbang Dua Opsi lalu Pilih Satu',
    'Atasi Keterbatasan dengan yang Ada',
    'Tulis Lembar Catatan yang Bisa Dibuka Lagi'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-15'
  AND title IN (
    'Ajarkan Satu Keterampilan Sederhana',
    'Abadikan Satu Perubahan Kecil Hari Ini',
    'Bagikan Satu Pengalaman dan Pelajaran Praktis'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-16'
  AND title IN (
    'Tuangkan Satu Ide Jadi Rencana Singkat',
    'Pilih Tanggapan Terbaik Sehari-hari',
    'Pamerkan Satu Hasil Buatan atau Perbaikanmu'
  );

-- 2. Insert 3 fresh tasks for 2026-10-12.
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
    'Rekam Satu Detail Kecil lewat Foto',
    'Benda yang setiap hari kamu pegang menyimpan detail kecil yang jarang disadari! Pilih SATU benda yang akrab bagimu, foto SATU detail kecilnya dari dekat (tekstur, pola aus, sambungan, atau tanda pemakaian), lalu jelaskan detail itu dan kenapa sering terlewat.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-12',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Pilih SATU benda sehari-hari yang akrab bagimu (contoh: tas, sepatu, botol minum, alat dapur, atau perabot kecil), lalu ambil atau unggah tepat 1 foto CLOSE-UP satu detail kecil pada benda itu (boleh kamera langsung atau galeri, maksimal 1 file). Wajib tulis catatan dengan 3 bagian: (a) Detail apa - jelaskan detail kecil yang kamu foto dan letaknya di bagian mana, (b) Pada benda apa - sebutkan bendanya dan seberapa sering kamu memakainya, (c) Kenapa sering terlewat - jelaskan kenapa detail itu jarang disadari orang. Ini tugas pengamatan, bukan perbaikan. Admin akan menolak submission tanpa foto close-up detail, tanpa 3 bagian catatan, atau foto yang tidak menunjukkan satu detail yang spesifik."
    }'::jsonb
),
(
    'demo-crew-1',
    'Tentukan Langkah Paling Tepat',
    'Hari biasa penuh momen yang menuntut satu langkah yang tepat! Hadapi 6 situasi rumah tangga yang umum - lampu mati, galon habis, kunci nyangkut, paket datang, pinjaman barang, dan colokan penuh. Pada setiap situasi, tentukan SATU langkah paling tepat yang harus dilakukan lebih dulu!',
    'MINI_GAME',
    'AUTO',
    2,
    '2026-10-12',
    25,
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
            "title": "Situasi 1 - Lampu Kamar Mati",
            "description": "Malam hari lampu kamarmu tiba-tiba mati. Kamu perlu cahaya untuk belajar. Apa langkah paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Cek dulu saklar dan bohlamnya, coba bohlam cadangan atau pindahkan ke ruangan terang",
                "delta": 20,
                "hint": "Mengecek sumber masalah dulu sebelum bertindak menghemat waktu dan biaya"
              },
              {
                "id": "b",
                "label": "Langsung beli lampu baru malam itu juga tanpa mengecek apa pun",
                "delta": 5,
                "hint": "Membeli pengganti sebelum tahu rusaknya di mana berisiko salah beli"
              },
              {
                "id": "c",
                "label": "Biarkan gelap dan berharap besok menyala sendiri",
                "delta": 0,
                "hint": "Mendiamkan masalah tanpa langkah apa pun tidak mengubah keadaan"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 - Galon Habis Saat Haus",
            "description": "Kamu sangat haus tetapi galon air habis dan toko sudah tutup. Apa langkah paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Minta sedikit air ke tetangga atau gunakan persediaan darurat, isi galon besok pagi",
                "delta": 20,
                "hint": "Kebutuhan mendesak diatasi dengan sumber terdekat, sisanya dijadwalkan besok"
              },
              {
                "id": "b",
                "label": "Beli minuman kemasan mahal dalam jumlah banyak lewat ojek online",
                "delta": 5,
                "hint": "Solusi darurat yang mahal untuk kebutuhan sesaat menguras uang tanpa perlu"
              },
              {
                "id": "c",
                "label": "Tahan haus sampai besok tanpa melakukan apa pun",
                "delta": 0,
                "hint": "Menahan kebutuhan dasar tanpa usaha mencari alternatif merugikan dirimu sendiri"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 - Kunci Nyangkut",
            "description": "Kunci motormu nyangkut dan sulit diputar. Kamu terburu-buru. Apa langkah paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Tenang, goyangkan kunci perlahan sambil mengatur posisi setang, beri pelumas bila ada",
                "delta": 20,
                "hint": "Kunci yang dipaksa berisiko patah; gerakan perlahan dan pelumas menyelesaikan tanpa merusak"
              },
              {
                "id": "b",
                "label": "Putar paksa sekuat tenaga agar cepat terbuka",
                "delta": 0,
                "hint": "Memaksa kunci yang nyangkut justru membuatnya patah dan masalahnya berlipat"
              },
              {
                "id": "c",
                "label": "Congkel rumah kunci dengan obeng atau benda tajam",
                "delta": 5,
                "hint": "Mencongkel merusak rumah kunci dan biayanya jauh lebih besar daripada bersabar sebentar"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 - Paket Datang Saat Sibuk",
            "description": "Kurir mengantar paket saat kedua tanganmu sibuk memasak dan tidak bisa ke depan. Apa langkah paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Minta kurir menitipkan di tempat aman atau meletakkannya di teras lalu kabari keluarga",
                "delta": 20,
                "hint": "Solusi titip yang aman menyelamatkan paket tanpa mengorbankan masakanmu"
              },
              {
                "id": "b",
                "label": "Biarkan paket tergeletak di depan pagar tanpa kepastian",
                "delta": 0,
                "hint": "Paket tanpa kepastian penyimpanan berisiko hilang atau rusak"
              },
              {
                "id": "c",
                "label": "Suruh kurir menunggu lama sampai kamu selesai memasak",
                "delta": 5,
                "hint": "Menahan kurir terlalu lama merugikan orang lain padahal ada solusi titip"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 - Teman Meminjam Barang Penting",
            "description": "Teman ingin meminjam barang pentingmu padahal besok kamu juga membutuhkannya. Apa langkah paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Jelaskan kamu juga membutuhkannya besok, tawarkan jadwal pinjam bergantian atau barang pengganti",
                "delta": 20,
                "hint": "Jujur soal kebutuhanmu sambil menawarkan alternatif menjaga barang dan pertemanan"
              },
              {
                "id": "b",
                "label": "Langsung berikan tanpa mencatat kapan kembali",
                "delta": 5,
                "hint": "Memberi tanpa kesepakatan waktu kembali berisiko kamu sendiri yang kesulitan besok"
              },
              {
                "id": "c",
                "label": "Tolak mentah-mentah tanpa penjelasan apa pun",
                "delta": 0,
                "hint": "Penolakan tanpa penjelasan merusak hubungan padahal ada jalan tengah"
              }
            ]
          },
          {
            "id": "ev_6",
            "title": "Situasi 6 - Colokan Penuh",
            "description": "Semua colokan terisi dan HP-mu hampir habis baterai padahal akan dipakai menelepon penting. Apa langkah paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Cabut perangkat yang sudah penuh dayanya, lalu cas HP-mu untuk keperluan penting itu",
                "delta": 20,
                "hint": "Memprioritaskan colokan berdasarkan kebutuhan mendesak adalah langkah yang adil dan masuk akal"
              },
              {
                "id": "b",
                "label": "Tumpuk beberapa steker dengan paksa agar HP-mu ikut tercolok",
                "delta": 0,
                "hint": "Menumpuk steker secara paksa berbahaya dan bisa merusak perangkat"
              },
              {
                "id": "c",
                "label": "Biarkan HP habis dan batalkan telepon penting itu",
                "delta": 5,
                "hint": "Menyerah tanpa mengatur ulang colokan mengorbankan keperluan penting yang sebenarnya bisa diselamatkan"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Uraikan Satu Hal dengan Bahasamu Sendiri',
    'Bisa menjelaskan dengan kata-kata sendiri tandanya benar-benar paham! Pilih SATU hal sederhana yang kamu kuasai, lalu uraikan secara tertulis dengan bahasamu sendiri lengkap dengan satu perumpamaan dan satu contoh agar teman pemula bisa memahaminya.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-10-12',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Hal yang kamu jelaskan - sebutkan SATU hal sederhana yang kamu kuasai (contoh: kenapa nasi basi, cara kerja payung, kenapa kaca berembun, atau cara mengatur sesuatu). Jelaskan dengan kata-katamu sendiri, bukan menyalin dari internet:\n2) Perumpamaan - buat SATU perumpamaan sederhana agar mudah dibayangkan (contoh: seperti ..., ibarat ...):\n3) Contoh nyata - berikan SATU contoh nyata dari pengalamanmu yang menunjukkan hal itu benar:\n\nDitulis untuk teman pemula yang belum tahu apa-apa. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, tanpa perumpamaan, atau penjelasan yang disalin dan sulit dipahami pemula.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
)
;

-- 3. Insert 3 fresh tasks for 2026-10-13.
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
    'Peragakan Cara Memakai Satu Benda',
    'Benda sehari-hari baru berguna kalau dipakai dengan benar! Pilih SATU benda di rumahmu, lalu rekam video maksimal 60 detik yang MEMPERAGAKAN cara memakai benda itu dengan benar dan aman. Orang lain harus bisa menirunya setelah menonton!',
    'VIDEO',
    'ADMIN_REVIEW',
    1,
    '2026-10-13',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik dengan struktur:\n1. Tunjukkan SATU benda sehari-hari yang spesifik (contoh: payung lipat, setrika, blender, gembok, dispenser, atau kipas angin) - fokus pada satu benda saja\n2. PERAGAKAN langsung di depan kamera cara memakai benda itu dengan benar dan aman (wajib ada peragaan nyata, bukan hanya bicara) - tunjukkan bagian-bagian benda yang relevan\n3. Sebutkan SATU kesalahan umum yang sering dilakukan orang saat memakai benda itu dan cara menghindarinya\n\nCatatan: ini peragaan pemakaian dasar yang benar, bukan trik dan bukan cara yang lebih praktis. Admin akan menolak video tanpa peragaan nyata, tanpa satu benda yang spesifik, tanpa kesalahan umum yang dijelaskan, durasi melebihi 60 detik, atau cara pakai yang tidak aman."
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Temukan Fungsi Lain Satu Barang',
    'Banyak barang bisa dipakai untuk hal lain selain fungsi utamanya! Pilih SATU barang di rumahmu, sebutkan fungsi utamanya, lalu temukan SATU fungsi lain yang masuk akal lengkap dengan cara melakukannya dan batasannya.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    2,
    '2026-10-13',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Barang dan fungsi utama - sebutkan SATU barang dan fungsi utamanya (contoh: handuk untuk mengeringkan badan, kardus untuk menyimpan barang, atau jepit jemuran untuk menjepit pakaian):\n2) Fungsi lain - jelaskan SATU fungsi lain yang masuk akal untuk barang itu dan cara melakukannya langkah demi langkah:\n3) Batasan - jelaskan kapan fungsi lain itu TIDAK cocok dipakai dan kenapa:\n\nFokus pada satu barang dan satu fungsi lain saja. Fungsi lain harus aman dan tidak merusak barang. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, atau fungsi lain yang berbahaya atau tidak masuk akal.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
),
(
    'demo-crew-1',
    'Susun Lembar Panduan untuk Teman Baru',
    'Bayangkan temanmu baru pertama kali melakukan sesuatu dan tidak tahu apa-apa! Pilih SATU aktivitas sehari-hari yang spesifik, lalu buat SATU dokumen lembar panduan satu halaman untuk pemula lengkap dengan istilah yang dijelaskan dan kesalahan umum yang harus dihindari. Ini panduan bacaan, bukan daftar centang.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-10-13',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "accepted_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4,
      "instruction": "Buat SATU dokumen lembar panduan untuk pemula dengan aturan: 1) Maksimal sekitar 1 halaman - singkat dan langsung ke intinya. 2) Fokus pada SATU aktivitas sehari-hari yang spesifik (contoh: cara mencuci sepatu, cara menyimpan beras, cara merawat tanaman, atau cara mengatur dompet). 3) Ditulis untuk orang yang belum pernah melakukannya: jelaskan istilah yang mungkin asing dan sebutkan minimal 2 kesalahan umum pemula beserta cara menghindarinya. 4) Ditulis sebagai panduan bacaan yang runut - BUKAN berupa daftar centang (dokumen berisi kotak centang seperti [ ] ... akan ditolak karena itu jenis tugas yang berbeda). 5) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: aktivitas apa dan untuk pemula seperti apa panduan itu. Admin akan menolak dokumen berupa daftar centang, dokumen tanpa penjelasan istilah, dokumen tanpa kesalahan umum, atau topik yang terlalu umum."
    }'::jsonb
)
;

-- 4. Insert 3 fresh tasks for 2026-10-14.
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
    'Timbang Dua Opsi lalu Pilih Satu',
    'Setiap hari kita dihadapkan pada dua pilihan! Pilih SATU keputusan nyata yang sedang kamu hadapi dengan dua opsi saja, bandingkan keduanya dengan dua kriteria yang jelas, lalu tentukan pilihanmu beserta alasannya.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    1,
    '2026-10-14',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Keputusan dan dua opsi - jelaskan SATU keputusan nyata sehari-hari dengan tepat DUA opsi (contoh: memperbaiki sendiri atau minta bantuan, berangkat pagi atau siang, menyimpan di lemari atau di kardus). Bukan soal belanja barang:\n2) Perbandingan dua kriteria - bandingkan kedua opsi dengan DUA kriteria yang jelas (contoh: biaya dan waktu, usaha dan hasil, atau cocok dan risiko). Jelaskan kelebihan dan kekurangan tiap opsi:\n3) Pilihan dan alasan - tentukan SATU pilihanmu dan jelaskan kenapa opsi itu menang berdasarkan perbandingan di atas:\n\nFokus pada satu keputusan dengan dua opsi saja. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, membahas lebih dari dua opsi, atau pilihan tanpa alasan dari perbandingan.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
),
(
    'demo-crew-1',
    'Atasi Keterbatasan dengan yang Ada',
    'Tidak selalu bisa membeli yang baru - kadang harus beres dengan yang ada! Hadapi 6 situasi serba terbatas di mana kamu tidak boleh membeli apa pun: tali putus, paku kurang, wadah kurang, cahaya redup, air sedikit, dan kantong sobek. Pada setiap situasi, pilih akal terbaik dengan bahan yang ada!',
    'MINI_GAME',
    'AUTO',
    2,
    '2026-10-14',
    25,
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
            "title": "Situasi 1 - Tali Putus",
            "description": "Tali untuk mengikat barang putus dan tidak ada tali cadangan. Kamu tidak boleh membeli. Apa akal terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Sambung dengan kain perca atau plastik yang dipilin menjadi tali darurat yang kuat",
                "delta": 20,
                "hint": "Bahan lentur di sekitar bisa dipilin menjadi pengikat darurat yang cukup kuat"
              },
              {
                "id": "b",
                "label": "Ikat seadanya dengan sisa tali pendek walau longgar",
                "delta": 5,
                "hint": "Ikatan longgar yang dipaksakan akan lepas di jalan"
              },
              {
                "id": "c",
                "label": "Biarkan barang tanpa ikatan dan berharap tidak jatuh",
                "delta": 0,
                "hint": "Barang tanpa ikatan pasti bermasalah saat dibawa"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 - Paku Kurang",
            "description": "Paku untuk menggantung pajangan kurang dan toko jauh. Apa akal terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Manfaatkan pengait tempel, paku yang masih menancap, atau sandarkan pajangan dengan aman",
                "delta": 20,
                "hint": "Titik gantung yang sudah ada atau pengait tempel menyelesaikan tanpa paku baru"
              },
              {
                "id": "b",
                "label": "Paksakan paku yang bengkok untuk dipakai lagi",
                "delta": 5,
                "hint": "Paku bengkok yang dipaksakan tidak kuat menahan beban"
              },
              {
                "id": "c",
                "label": "Tempel pajangan dengan lem kertas biasa",
                "delta": 0,
                "hint": "Lem kertas tidak dirancang menahan beban pajangan dan pasti jatuh"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 - Wadah Kurang",
            "description": "Makanan sisa cukup banyak tetapi wadah tertutup kurang. Apa akal terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Gunakan mangkuk lalu tutup rapat dengan piring dan ikat dengan kain atau karet",
                "delta": 20,
                "hint": "Wadah darurat yang tertutup rapat menjaga makanan sama baiknya"
              },
              {
                "id": "b",
                "label": "Simpan dalam piring terbuka di kulkas",
                "delta": 5,
                "hint": "Makanan terbuka menyerap bau dan mengering walau di kulkas"
              },
              {
                "id": "c",
                "label": "Biarkan di meja tanpa tutup sampai besok",
                "delta": 0,
                "hint": "Makanan tanpa tutup di suhu ruang cepat basi"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 - Cahaya Redup",
            "description": "Lampu belajar redup dan kamu harus membaca malam ini juga. Apa akal terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Pindah ke dekat sumber cahaya yang masih kuat atau gunakan lampu HP sebagai penerangan fokus",
                "delta": 20,
                "hint": "Mendekati cahaya yang ada jauh lebih efektif daripada memaksakan mata dalam gelap"
              },
              {
                "id": "b",
                "label": "Tetap membaca di cahaya redup sambil menyipitkan mata",
                "delta": 0,
                "hint": "Membaca dalam cahaya redup melelahkan mata dan tidak efektif"
              },
              {
                "id": "c",
                "label": "Nyalakan semua lampu rumah walau tidak membantu bacaanmu",
                "delta": 5,
                "hint": "Menyalakan semua lampu tanpa mengarahkan cahaya ke bacaan membuang listrik tanpa hasil"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 - Air Sedikit",
            "description": "Air tinggal sedikit tetapi tumpukan piring kotor banyak. Apa akal terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Rendam yang berlemak dulu agar lunak, lalu cuci hemat dengan spons dan bilas seperlunya",
                "delta": 20,
                "hint": "Merendam dulu membuat kotoran lunak sehingga sedikit air pun cukup"
              },
              {
                "id": "b",
                "label": "Guyur semua piring dengan air seperti biasa sampai air habis",
                "delta": 0,
                "hint": "Mengguyur boros saat air sedikit membuat sebagian piring tidak kebagian"
              },
              {
                "id": "c",
                "label": "Tumpuk saja piring kotornya dan tunggu air banyak besok",
                "delta": 5,
                "hint": "Menumpuk tanpa merendam membuat lemak mengeras dan besok jauh lebih sulit dicuci"
              }
            ]
          },
          {
            "id": "ev_6",
            "title": "Situasi 6 - Kantong Sobek",
            "description": "Kantong belanjaan sobek di jalan dan tidak ada kantong cadangan. Apa akal terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Rangkap bagian sobek dengan lipatan, bagi beban ke dua tangan atau minta tolong bawakan sebagian",
                "delta": 20,
                "hint": "Mendistribusikan beban dan melipat bagian sobek menyelamatkan belanjaan sampai rumah"
              },
              {
                "id": "b",
                "label": "Tetap jinjing seperti biasa walau sobeknya makin lebar",
                "delta": 0,
                "hint": "Beban penuh pada kantong sobek menjamin isinya tumpah di jalan"
              },
              {
                "id": "c",
                "label": "Tinggalkan sebagian belanjaan di jalan agar kantong ringan",
                "delta": 5,
                "hint": "Meninggalkan belanjaan merugikan; membagi beban adalah jalan yang lebih baik"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Tulis Lembar Catatan yang Bisa Dibuka Lagi',
    'Catatan yang baik menyelamatkanmu di kemudian hari! Pilih SATU hal yang sering kamu lupa atau perlu diingat (langkah kerja, rumus sederhana, daftar istilah, atau template), lalu buat SATU dokumen lembar catatan ringkas maksimal satu halaman yang bisa kamu buka lagi kapan pun. Ini catatan rujukan, bukan daftar centang dan bukan karangan panjang.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-10-14',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "accepted_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4,
      "instruction": "Buat SATU dokumen lembar catatan rujukan dengan aturan: 1) Maksimal sekitar 1 halaman - ringkas agar enak dibuka lagi. 2) Berisi SATU topik rujukan yang spesifik (contoh: langkah mencuci sepatu, daftar istilah HP, template pesan izin, atau rumus belanja hemat). 3) Ditulis sebagai catatan rujukan yang runut (penjelasan singkat plus poin-poin ingatan) sehingga kamu sendiri bisa memakainya lagi bulan depan tanpa penjelasan tambahan - BUKAN berupa daftar centang saja (dokumen berisi kotak centang seperti [ ] ... akan ditolak) dan BUKAN karangan panjang berlembar-lembar. 4) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: topik catatanmu dan kapan kamu akan membukanya lagi. Admin akan menolak dokumen berupa daftar centang saja, karangan panjang tanpa struktur rujukan, atau topik yang terlalu umum."
    }'::jsonb
)
;

-- 5. Insert 3 fresh tasks for 2026-10-15.
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
    'Ajarkan Satu Keterampilan Sederhana',
    'Punya keterampilan kecil yang berguna? Saatnya mengajarkannya! Pilih SATU keterampilan sederhana yang kamu kuasai (misalnya melipat baju, mengikat tali, mencatat tugas, atau menyusun barang), lalu rekam video maksimal 60 detik yang MENGAJARKANNYA langkah demi langkah. Pemula harus bisa mengikutinya!',
    'VIDEO',
    'ADMIN_REVIEW',
    1,
    '2026-10-15',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik dengan struktur:\n1. Sebutkan SATU keterampilan sederhana yang spesifik dan benar-benar kamu kuasai (contoh: melipat baju, mengikat tali sepatu, mencatat tugas harian, atau menyusun barang) - fokus pada satu keterampilan untuk pemula\n2. AJARKAN langsung di depan kamera langkah demi langkah (wajib ada peragaan nyata, bukan hanya bicara) - tunjukkan objek atau alat yang dipakai jika relevan\n3. Sebutkan SATU kesalahan umum pemula dan cara memperbaikinya\n\nCatatan: ini pengajaran keterampilan dasar untuk pemula, bukan trik dan bukan sekadar cara memakai benda. Admin akan menolak video tanpa peragaan mengajar, tanpa satu keterampilan yang spesifik, tanpa kesalahan umum pemula, durasi melebihi 60 detik, atau langkah yang tidak bisa diikuti pemula."
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Abadikan Satu Perubahan Kecil Hari Ini',
    'Sekitarmu berubah sedikit setiap hari - sadarkah kamu? Temukan SATU perubahan kecil yang terjadi hari ini di sekitarmu (tanaman bertambah daun, barang yang diperbaiki, sudut yang ditata ulang, atau hasil kerja seseorang), foto, lalu ceritakan perubahan itu.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-10-15',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Temukan SATU perubahan kecil yang nyata terjadi hari ini di sekitarmu (contoh: tanaman yang bertunas, barang yang baru diperbaiki, sudut ruangan yang ditata ulang, atau hasil kerja seseorang hari ini), lalu ambil atau unggah tepat 1 foto perubahan itu (boleh kamera langsung atau galeri, maksimal 1 file). Wajib tulis catatan dengan 3 bagian: (a) Apa yang berubah - jelaskan perubahannya dan di mana, (b) Sejak kapan - jelaskan kapan kamu menyadari perubahan itu, (c) Kenapa berarti - jelaskan kenapa perubahan kecil itu penting atau menyenangkan. Ini tugas pengamatan perubahan, kamu tidak wajib menjadi pelakunya. Admin akan menolak submission tanpa foto perubahan, tanpa 3 bagian catatan, atau perubahan yang tidak terlihat nyata."
    }'::jsonb
),
(
    'demo-crew-1',
    'Bagikan Satu Pengalaman dan Pelajaran Praktis',
    'Pengalaman jadi berharga saat pelajarannya bisa dipakai lagi! Tulis SATU pengalaman nyatamu, pelajaran praktis yang kamu ambil, dan SATU cara kamu akan menerapkan pelajaran itu ke depan. Ditulis dari pengalaman sendiri, bukan karangan umum.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-10-15',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Pengalaman - ceritakan SATU pengalaman nyata yang kamu alami sendiri (contoh: memperbaiki sesuatu, mengatur sesuatu, menghadapi situasi sulit, atau mencoba hal baru). Ceritakan kejadiannya secara konkret:\n2) Pelajaran praktis - jelaskan SATU pelajaran praktis yang kamu ambil dari pengalaman itu:\n3) Penerapan ke depan - tulis SATU cara konkret kamu akan menerapkan pelajaran itu dalam 7 hari ke depan:\n\nDitulis sebagai teks, bukan video. Fokus pada satu pengalaman dan satu pelajaran saja. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, pengalaman yang tidak konkret, atau penerapan yang tidak bisa dilakukan dalam 7 hari.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
)
;

-- 6. Insert 3 fresh tasks for 2026-10-16.
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
    'Tuangkan Satu Ide Jadi Rencana Singkat',
    'Punya ide bagus? Jangan biarkan hanya di kepala! Pilih SATU ide sederhana yang bisa membuat hidup sehari-hari lebih baik, lalu buat SATU dokumen rencana singkat satu halaman: masalahnya apa, idemu apa, dan langkah pertama yang bisa dilakukan minggu ini. Rencana harus bisa langsung dijalankan.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-16',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "accepted_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4,
      "instruction": "Buat SATU dokumen rencana singkat dengan aturan: 1) Maksimal sekitar 1 halaman - singkat dan langsung ke intinya. 2) Berisi SATU ide sederhana untuk kehidupan sehari-hari (contoh: cara mengatur sudut ruangan, cara menghemat sesuatu, cara memperbaiki kebiasaan, atau cara memanfaatkan barang yang ada). 3) Struktur wajib: masalah yang ingin diatasi, ide dan cara kerjanya, serta SATU langkah pertama yang konkret dan bisa dilakukan dalam 7 hari. 4) BUKAN berupa daftar centang saja (dokumen berisi kotak centang seperti [ ] ... akan ditolak) dan BUKAN karangan tanpa langkah yang bisa dijalankan. 5) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: ide apa dan untuk siapa rencana itu. Admin akan menolak dokumen tanpa langkah pertama yang konkret, dokumen berupa daftar centang saja, atau ide yang tidak bisa dijalankan."
    }'::jsonb
),
(
    'demo-crew-1',
    'Pilih Tanggapan Terbaik Sehari-hari',
    'Cara kita menanggapi hal kecil menunjukkan kedewasaan! Hadapi 6 situasi sosial sehari-hari - kembalian lebih, salah kirim pesan, teman terlambat, tumpahan di tempat umum, diminta tolong saat sibuk, dan parkir sempit. Pada setiap situasi, pilih tanggapan yang paling baik!',
    'MINI_GAME',
    'AUTO',
    2,
    '2026-10-16',
    25,
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
            "title": "Situasi 1 - Kembalian Lebih",
            "description": "Kasir mengembalikan uang lebih dari seharusnya dan tidak sadar. Antrean di belakangmu panjang. Apa tanggapan terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Segera kembalikan kelebihannya dengan sopan walau antrean menunggu sebentar",
                "delta": 20,
                "hint": "Kelebihan yang dikembalikan menjaga kejujuran; antrean sebentar bisa dimaklumi"
              },
              {
                "id": "b",
                "label": "Ambil diam-diam dan anggap rezeki",
                "delta": 0,
                "hint": "Mengambil kelebihan yang bukan hakmu merugikan kasir yang harus mengganti"
              },
              {
                "id": "c",
                "label": "Belanjakan lagi kelebihannya di toko yang sama agar tidak kentara",
                "delta": 5,
                "hint": "Menghabiskan kelebihan tetap berarti mengambil yang bukan hakmu"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 - Salah Kirim Pesan",
            "description": "Kamu salah mengirim pesan pribadi ke grup keluarga besar. Isinya tidak pantas dibaca umum. Apa tanggapan terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Segera akui kesalahan, minta maaf singkat, lalu kirim koreksi yang benar",
                "delta": 20,
                "hint": "Mengakui dan mengoreksi dengan cepat memulihkan keadaan lebih baik daripada menghilang"
              },
              {
                "id": "b",
                "label": "Hapus pesan diam-diam dan berharap tidak ada yang sempat membaca",
                "delta": 5,
                "hint": "Menghapus tanpa penjelasan meninggalkan kebingungan bagi yang sudah terlanjur membaca"
              },
              {
                "id": "c",
                "label": "Biarkan saja dan pura-pura tidak terjadi apa-apa",
                "delta": 0,
                "hint": "Membiarkan tanpa tanggapan membuat kesalahpahaman berkembang"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 - Teman Terlambat",
            "description": "Teman janjian denganmu terlambat 30 menit tanpa kabar. Kamu sudah menunggu. Apa tanggapan terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Hubungi dengan tenang, tanyakan kabarnya, sepakati batas tunggu yang wajar",
                "delta": 20,
                "hint": "Menanyakan kabar dengan tenang memberi kesempatan penjelasan sekaligus batas yang jelas"
              },
              {
                "id": "b",
                "label": "Marahi langsung lewat pesan panjang penuh emosi",
                "delta": 0,
                "hint": "Luapan emosi merusak pertemuan yang sebenarnya masih bisa diselamatkan"
              },
              {
                "id": "c",
                "label": "Pergi diam-diam tanpa kabar dan biarkan dia mencari",
                "delta": 5,
                "hint": "Pergi tanpa kabar membalas kekecewaan dengan kekecewaan baru"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 - Menumpahkan Minuman",
            "description": "Kamu tidak sengaja menumpahkan minuman di tempat umum dan mengotori lantai. Orang melihat. Apa tanggapan terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Minta maaf singkat lalu segera lap dan bereskan sampai bersih",
                "delta": 20,
                "hint": "Bertanggung jawab membersihkan langsung menunjukkan kedewasaan"
              },
              {
                "id": "b",
                "label": "Pura-pura tidak melihat dan pergi menjauh",
                "delta": 0,
                "hint": "Menghindar meninggalkan kotoran untuk orang lain dan membahayakan yang lewat"
              },
              {
                "id": "c",
                "label": "Salahkan lantai yang licin atau gelas yang terlalu penuh",
                "delta": 5,
                "hint": "Mencari kambing hitam tidak membersihkan tumpahan yang tetap harus dibereskan"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 - Diminta Tolong Saat Sibuk",
            "description": "Keluargamu meminta tolong saat kamu sedang mengerjakan hal penting yang mendesak. Apa tanggapan terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Jelaskan kamu butuh beberapa menit menyelesaikan ini, lalu bantu setelahnya",
                "delta": 20,
                "hint": "Memberi kepastian waktu membantu jauh lebih baik daripada menolak atau mengorbankan pekerjaan"
              },
              {
                "id": "b",
                "label": "Tinggalkan pekerjaan pentingmu seketika walau berantakan",
                "delta": 5,
                "hint": "Meninggalkan pekerjaan mendesak tanpa kepastian membuat dua-duanya berantakan"
              },
              {
                "id": "c",
                "label": "Tolak dengan kasar agar tidak diganggu lagi",
                "delta": 0,
                "hint": "Penolakan kasar melukai orang yang sebenarnya hanya butuh bantuan kecil"
              }
            ]
          },
          {
            "id": "ev_6",
            "title": "Situasi 6 - Parkir Sempit",
            "description": "Motor tetangga menghalangi jalan keluar motormu di parkiran sempit. Kamu terburu-buru. Apa tanggapan terbaikmu?",
            "options": [
              {
                "id": "a",
                "label": "Geser perlahan dengan hati-hati lalu kabari tetangga agar lain kali parkir lebih rapi",
                "delta": 20,
                "hint": "Menggeser hati-hati plus kabar baik menyelesaikan masalah tanpa merusak dan mencegah terulang"
              },
              {
                "id": "b",
                "label": "Dorong paksa sampai motormu bisa keluar walau motor tetangga tergores",
                "delta": 0,
                "hint": "Mendorong paksa merusak barang orang lain demi kepentinganmu sendiri"
              },
              {
                "id": "c",
                "label": "Klakson dan marah-marah sampai tetangga keluar",
                "delta": 5,
                "hint": "Keributan tidak mengeluarkan motormu lebih cepat daripada menggesernya baik-baik"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Pamerkan Satu Hasil Buatan atau Perbaikanmu',
    'Bangga dengan sesuatu yang kamu buat atau perbaiki sendiri? Saatnya memamerkannya! Foto SATU hasil buatan tanganmu atau barang yang kamu perbaiki sendiri, lalu ceritakan cara singkatnya dan apa gunanya sekarang.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-10-16',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Pilih SATU hasil karya tanganmu sendiri atau SATU barang yang kamu perbaiki sendiri (contoh: masakan, kerajinan, tulisan, susunan, atau barang yang tadinya rusak lalu kamu betulkan), lalu ambil atau unggah tepat 1 foto hasilnya (boleh kamera langsung atau galeri, maksimal 1 file). Wajib tulis catatan dengan 3 bagian: (a) Hasil apa - sebutkan apa yang kamu buat atau perbaiki, (b) Cara singkat - jelaskan cara singkat pembuatannya atau apa yang kamu perbaiki, (c) Guna sekarang - jelaskan apa gunanya hasil itu sekarang. Karya atau perbaikan harus hasil tanganmu sendiri. Admin akan menolak submission tanpa foto hasil, tanpa 3 bagian catatan, atau hasil yang jelas bukan buatan atau perbaikanmu sendiri."
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '100_tasks_2026_10_12_16')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
