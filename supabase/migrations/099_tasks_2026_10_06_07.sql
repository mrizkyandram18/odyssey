-- ============================================================
-- Migration 099: Fresh tasks for 6 + 7 October 2026
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction/question/scenario text may be
-- hardcoded in frontend/backend source. Admin can edit every value below
-- through the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. No legacy rows exist on 2026-10-06 / 2026-10-07 (verified: zero
--    grep hits for 2026-10-0[67] in supabase/migrations; pre-apply script
--    re-audits live and ABORTS if any row carries submissions/ledger).
--    odyssey_claims has no task_id column, so no claim linkage is
--    possible. Fresh INSERT only; no delete/deactivate of other dates.
-- 2. Inserts 6 fresh tasks (3 per date, existing architecture and
--    config shapes only, no new framework, no new tables, no new
--    engines, no new upload system, no economy change):
--    2026-10-06:
--    - Step 1: VIDEO (Temukan Cara yang Lebih Praktis) - member picks
--      ONE everyday activity they do regularly, then DEMONSTRATES a
--      more practical/efficient way to do it (demo + improvement +
--      imitable result). recording.enabled, max_duration_seconds = 60,
--      camera user; ADMIN_REVIEW. Reward 20c/100xp.
--    - Step 2: MINI_GAME (Pilih yang Paling Masuk Akal) - 6 everyday
--      decision situations under practical constraints (mepet waktu,
--      baterai menipis, tas terbatas, dua ajakan bersamaan, bahan
--      terbatas, hujan tiba-tiba), each with 3 options + deterministic
--      delta + hint. DECISION_PRIORITY + POINTS, initial 0, target 100.
--      Scenario set is materially different from task 666 (awal-bulan
--      situations) and task 670 (pre-action readiness checks); this is
--      about choosing the most reasonable option, NOT anomaly
--      detection and NOT before-using-something checks. AUTO via the
--      existing ValidateDecisionChoices server recompute (client score
--      never trusted). Reward 25c/100xp.
--    - Step 3: TEXT_RESPONSE (Catat Satu Hal yang Jangan Sampai
--      Kelupaan) - member writes what they often forget / when it
--      usually happens / one simple reminder method. min 150 chars,
--      max 800 (server-enforced by odyssey_submit_manual_task P0008
--      from config); ADMIN_REVIEW. Reward 20c/100xp.
--    2026-10-07:
--    - Step 1: PHOTO_UPLOAD (Buktikan Barang Ini Masih Berguna) -
--      member submits exactly 1 photo of an item they still regularly
--      use plus a note with 3 parts (barang apa / dipakai untuk apa /
--      kenapa masih layak dan belum perlu diganti). max_files 1,
--      camera_only = false (gallery/file-picker allowed via existing
--      CameraCaptureModal path, same as tasks 628/632/643/650);
--      ADMIN_REVIEW. Reward 20c/100xp.
--    - Step 2: MINI_GAME (Susun Langkahnya) - CAPABILITY NOTE: the
--      engine has NO ordering/sequencing mechanic (verified: no
--      drag-sort UI, no order validator; only validateDecisionScenario
--      + target_score; same gap documented in migration 085). NO new
--      mechanic is introduced. Adapted with the existing
--      DECISION_PRIORITY mechanism, preserving the concept as closely
--      as possible: 6 everyday routines (berangkat pagi, mencuci baju,
--      memasak mie, merapikan kamar, mengirim pesan penting, mencuci
--      piring), each asking which step is most correct to do FIRST -
--      sequencing taught as ordered priority decisions. POINTS,
--      initial 0, target 100. AUTO via existing server recompute.
--      Reward 25c/100xp.
--    - Step 3: DOCUMENT_UPLOAD (Bikin Panduan Singkat Versi Kamu) -
--      member uploads a SHORT PRACTICAL GUIDE (max ~1 page, directly
--      actionable, one everyday activity, understandable without extra
--      explanation). Explicitly NOT a checklist (checklist documents
--      are rejected; this distinguishes it from task 671 which is a
--      checklist task). allowed_extensions + accepted_extensions
--      (dual-key intentional: DocUploadModal reads accepted_
--      extensions while admin builder writes allowed_extensions),
--      max_file_size_mb = 4; ADMIN_REVIEW. Reward 20c/100xp.
-- 3. Idempotency: the pre-clean DELETEs below match ONLY the 6 exact
--    new titles on their exact dates (narrow exact-title guard for dev
--    re-runs, same spirit as migrations 083/084/085/086/087/088/090/
--    091/093/098). No broad date/title matching.
-- 4. Does NOT touch tasks 608-671 (610 camera_only stays true;
--    628/632/643/650 stay false), payout/economy config,
--    announcements, RLS, RPCs, triggers, claims, shop, or any other
--    date. Batch total 130 coins (20+25+20+20+25+20); economy
--    (cap/target/rate/payout) unchanged.
-- ============================================================

-- 1. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact dates). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-06'
  AND title IN (
    'Temukan Cara yang Lebih Praktis',
    'Pilih yang Paling Masuk Akal',
    'Catat Satu Hal yang Jangan Sampai Kelupaan'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-07'
  AND title IN (
    'Buktikan Barang Ini Masih Berguna',
    'Susun Langkahnya',
    'Bikin Panduan Singkat Versi Kamu'
  );

-- 2. Insert 3 fresh tasks for 2026-10-06.
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
    'Temukan Cara yang Lebih Praktis',
    'Setiap orang punya cara sendiri melakukan hal sehari-hari - dan sering ada cara yang lebih praktis! Pilih SATU aktivitas sederhana yang biasa kamu lakukan, lalu rekam video maksimal 60 detik yang MENDEMONSTRASIKAN cara yang menurutmu lebih praktis atau efisien. Orang lain harus bisa memahami dan menirunya setelah menonton videomu!',
    'VIDEO',
    'ADMIN_REVIEW',
    1,
    '2026-10-06',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik dengan struktur:\n1. Sebutkan SATU aktivitas sehari-hari yang spesifik dan biasa kamu lakukan (contoh: melipat baju, menyimpan kabel, menyiapkan bekal, membersihkan sesuatu, atau mengatur barang) - fokus pada satu aktivitas saja\n2. DEMONSTRASIKAN langsung di depan kamera cara yang menurutmu lebih praktis atau efisien untuk melakukan aktivitas itu (wajib ada demonstrasi nyata, bukan hanya bicara atau bercerita) - tunjukkan objek atau alat yang kamu gunakan jika relevan\n3. Jelaskan apa bedanya dengan cara biasa dan kenapa caramu lebih praktis, serta dalam situasi apa orang lain bisa menirunya\n\nCatatan: submission yang hanya berbicara tanpa demonstrasi TIDAK memenuhi task ini. Admin akan menolak video tanpa demonstrasi nyata, tanpa satu aktivitas yang spesifik, tanpa cara praktis yang jelas terlihat, durasi melebihi 60 detik, atau hasil yang tidak bisa dipahami atau ditiru orang lain."
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Pilih yang Paling Masuk Akal',
    'Hidup sehari-hari penuh pilihan kecil yang harus diputuskan dengan akal sehat! Hadapi 6 situasi biasa dengan keterbatasan nyata - waktu mepet, baterai menipis, tempat terbatas, ajakan bersamaan, bahan terbatas, dan hujan tiba-tiba. Pada setiap situasi, pilih tindakan yang paling masuk akal. Ini soal menimbang pilihan dan memutuskan dengan akal sehat!',
    'MINI_GAME',
    'AUTO',
    2,
    '2026-10-06',
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
            "title": "Situasi 1 - Sore yang Mepet",
            "description": "Tinggal 45 menit sebelum acara keluarga dimulai, kamu belum mandi, belum siap, dan masih ada pekerjaan kecil yang belum selesai. Apa yang paling masuk akal kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Segera mandi dan bersiap, tunda pekerjaan kecil itu sampai acara selesai",
                "delta": 20,
                "hint": "Hal yang terikat waktu dan melibatkan orang lain pantas didahulukan; pekerjaan kecil bisa menunggu tanpa merugikan siapa pun"
              },
              {
                "id": "b",
                "label": "Selesaikan dulu pekerjaan kecilnya, lalu bersiap secepat mungkin",
                "delta": 5,
                "hint": "Mendahulukan yang kecil saat waktu mepet berisiko membuatmu terlambat untuk hal yang lebih penting"
              },
              {
                "id": "c",
                "label": "Santap camilan dulu sambil main HP sebentar agar tenang, baru bersiap",
                "delta": 0,
                "hint": "Menambah kegiatan santai saat waktu tinggal 45 menit hanya mengorbankan persiapan yang sebenarnya mendesak"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 - Baterai Tinggal 10 Persen",
            "description": "Baterai HP-mu tinggal 10 persen dan tidak ada colokan di dekatmu. Kamu perlu mengabari orang rumah, memesan kendaraan, dan menyimpan nomor penting. Apa yang paling masuk akal?",
            "options": [
              {
                "id": "a",
                "label": "Kirim kabar singkat ke orang rumah dan pesan kendaraan dulu, hal lain menyusul kalau baterai masih ada",
                "delta": 20,
                "hint": "Daya terbatas harus dipakai untuk yang paling penting dulu: keselamatan kabar dan kepastian pulang"
              },
              {
                "id": "b",
                "label": "Simpan dulu semua nomor dan data penting, urusan kabar dan kendaraan nanti saja",
                "delta": 5,
                "hint": "Menyimpan data memang berguna, tetapi kabar ke rumah dan kendaraan pulang jauh lebih mendesak daripada arsip"
              },
              {
                "id": "c",
                "label": "Pakai sisa baterai untuk scroll media sosial sebentar, siapa tahu ada colokan nanti",
                "delta": 0,
                "hint": "Menghabiskan daya terakhir untuk hiburan berarti kebutuhan pentingnya tidak terpenuhi sama sekali"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 - Tas Kecil, Barang Banyak",
            "description": "Kamu akan pergi seharian dan hanya bisa membawa satu tas kecil. Barang yang ingin dibawa jauh lebih banyak daripada muatnya tas. Apa yang paling masuk akal?",
            "options": [
              {
                "id": "a",
                "label": "Pilih barang berdasarkan kebutuhan hari ini: dompet, HP, dan yang wajib untuk kegiatanmu, sisanya tinggalkan",
                "delta": 20,
                "hint": "Tas terbatas memaksa memilih berdasarkan kebutuhan nyata hari itu, bukan berdasarkan keinginan membawa semuanya"
              },
              {
                "id": "b",
                "label": "Bawa barang favoritmu saja, kebutuhan lain bisa pinjam atau beli di jalan",
                "delta": 5,
                "hint": "Memilih berdasarkan favorit tanpa menimbang kebutuhan berisiko meninggalkan justru yang paling wajib"
              },
              {
                "id": "c",
                "label": "Paksakan semua barang masuk dengan cara menjejalkan dan menggantung di luar tas",
                "delta": 0,
                "hint": "Memaksakan semua barang merusak tas dan membuat barang penting justru sulit diambil saat dibutuhkan"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 - Dua Ajakan di Waktu yang Sama",
            "description": "Teman mengajakmu jalan dadakan di jam yang sama dengan janji membantu keluarga yang sudah disepakati sejak kemarin. Apa yang paling masuk akal?",
            "options": [
              {
                "id": "a",
                "label": "Tepati janji membantu keluarga yang sudah disepakati, ajak teman di lain waktu",
                "delta": 20,
                "hint": "Komitmen yang sudah disepakati lebih dulu pantas ditepati; ajakan dadakan masih bisa diatur ulang tanpa mengecewakan siapa pun"
              },
              {
                "id": "b",
                "label": "Ikut teman dulu sebentar, urusan keluarga dikerjakan nanti kalau sempat",
                "delta": 5,
                "hint": "Menunda janji yang sudah disepakati demi ajakan dadakan mengecewakan orang yang sudah mengandalkanmu"
              },
              {
                "id": "c",
                "label": "Terima dua-duanya dan berharap waktunya ternyata tidak bertabrakan",
                "delta": 0,
                "hint": "Menerima dua komitmen berbenturan tanpa memilih hanya menjamin salah satunya terbengkalai"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 - Bahan Makanan Sisa Sedikit",
            "description": "Di dapur hanya tersisa sedikit bahan makanan, cukup untuk satu atau dua porsi, sementara yang perlu makan lebih dari itu. Apa yang paling masuk akal?",
            "options": [
              {
                "id": "a",
                "label": "Masak bahan yang ada menjadi hidangan sederhana yang bisa dibagi rata untuk semua",
                "delta": 20,
                "hint": "Bahan terbatas paling masuk akal diolah menjadi sesuatu yang cukup untuk semua, bukan yang enak untuk sebagian"
              },
              {
                "id": "b",
                "label": "Masak makanan favoritmu dari bahan itu, sisanya beli di luar kalau ada yang kurang",
                "delta": 5,
                "hint": "Mengutamakan selera sendiri saat bahan terbatas berarti ada yang tidak kebagian tanpa kepastian pengganti"
              },
              {
                "id": "c",
                "label": "Simpan bahannya untuk nanti dan biarkan semua orang mencari makan sendiri-sendiri",
                "delta": 0,
                "hint": "Menyimpan bahan yang sebenarnya cukup untuk dibagi hanya memindahkan masalah lapar ke semua orang"
              }
            ]
          },
          {
            "id": "ev_6",
            "title": "Situasi 6 - Hujan Turun Tiba-tiba",
            "description": "Hujan deras turun tiba-tiba: jemuran masih di luar, adik belum pulang dan tidak bawa payung, dan kamu sendiri harus segera pergi. Apa yang paling masuk akal kamu lakukan dulu?",
            "options": [
              {
                "id": "a",
                "label": "Amankan jemuran dengan cepat lalu kabari atau jemput adik sebelum kamu berangkat",
                "delta": 20,
                "hint": "Yang bisa rusak (jemuran) dan yang menyangkut orang (adik) didahulukan sebelum urusan keberangkatanmu sendiri"
              },
              {
                "id": "b",
                "label": "Langsung berangkat agar tidak telat, jemuran dan adik biar diurus nanti",
                "delta": 0,
                "hint": "Mendahulukan keberangkatan sendiri sambil membiarkan jemuran basah dan adik kehujanan mengorbankan dua hal sekaligus"
              },
              {
                "id": "c",
                "label": "Tunggu hujan reda sambil menonton dari jendela, semua urusan menyusul setelahnya",
                "delta": 5,
                "hint": "Menunggu tanpa bertindak membiarkan jemuran basah kuyup dan adik menunggu tanpa kabar"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Catat Satu Hal yang Jangan Sampai Kelupaan',
    'Semua orang punya satu hal yang sering kelupaan! Tulis SATU hal yang paling sering kamu lupakan, ceritakan kapan biasanya itu terjadi, dan tentukan satu cara sederhana yang akan kamu pakai agar tidak lupa lagi.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-10-06',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Hal yang sering kelupaan - jelaskan SATU hal yang paling sering kamu lupakan (contoh: membawa barang tertentu, mengabari seseorang, mematikan sesuatu, atau mengerjakan sesuatu). Ceritakan seperti apa kejadiannya:\n2) Kapan biasanya terjadi - jelaskan kapan atau dalam situasi apa kamu biasanya melupakannya (contoh: saat terburu-buru pagi hari, saat sedang asyik melakukan hal lain, atau saat jadwal berubah mendadak):\n3) Cara sederhana agar tidak lupa lagi - tulis SATU cara sederhana dan konkret yang akan kamu pakai mulai sekarang (contoh: menempel catatan, memasang pengingat, menaruh barang di tempat yang terlihat, atau kebiasaan kecil tertentu):\n\nFokus pada satu hal saja. Tulis dari pengalamanmu sendiri. Admin akan menolak jawaban di bawah 150 karakter, tanpa 3 bagian di atas, atau cara yang tidak konkret dan tidak bisa langsung dipraktikkan.",
      "minimum_characters": 150,
      "maximum_characters": 800
    }'::jsonb
)
;

-- 3. Insert 3 fresh tasks for 2026-10-07.
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
    'Buktikan Barang Ini Masih Berguna',
    'Sebelum membeli yang baru, lihat dulu sekitarmu! Pilih SATU barang yang masih rutin kamu pakai sampai sekarang, foto barangnya, dan buktikan kenapa barang itu masih berguna dan belum perlu diganti.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-10-07',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Pilih SATU barang yang masih rutin kamu pakai (contoh: tas, botol minum, alat tulis, peralatan dapur, pakaian, atau perabot kecil), lalu ambil atau unggah tepat 1 foto barang tersebut (boleh kamera langsung atau galeri, maksimal 1 file). Wajib tulis catatan dengan 3 bagian: (a) Barang apa - sebutkan barangnya dan sudah berapa lama kamu memakainya, (b) Dipakai untuk apa - jelaskan apa yang masih kamu lakukan dengan barang itu sehari-hari, (c) Kenapa masih layak - jelaskan kenapa barang itu masih berguna dan belum perlu diganti. Admin akan menolak submission tanpa foto barang, tanpa 3 bagian catatan (barang apa / dipakai untuk apa / kenapa masih layak), atau barang yang jelas sudah tidak dipakai lagi."
    }'::jsonb
),
(
    'demo-crew-1',
    'Susun Langkahnya',
    'Kegiatan sehari-hari punya urutan yang benar! Hadapi 6 rutinitas biasa - berangkat pagi, mencuci baju, memasak mie, merapikan kamar, mengirim pesan penting, dan mencuci piring. Pada setiap rutinitas, langkah-langkahnya teracak: tentukan langkah mana yang paling tepat dilakukan LEBIH DULU agar semuanya berjalan lancar!',
    'MINI_GAME',
    'AUTO',
    2,
    '2026-10-07',
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
            "title": "Rutinitas 1 - Berangkat Pagi",
            "description": "Besok pagi kamu harus berangkat tepat waktu. Langkah-langkahnya teracak: menyiapkan barang, mandi dan sarapan, lalu berangkat. Mana yang paling tepat dilakukan LEBIH DULU malam ini?",
            "options": [
              {
                "id": "a",
                "label": "Kumpulkan dan cek semua barang yang harus dibawa besok sejak malam ini",
                "delta": 20,
                "hint": "Menyiapkan barang malam sebelumnya adalah langkah pertama yang benar: pagi harimu tinggal mandi, sarapan, dan berangkat tanpa panik mencari barang"
              },
              {
                "id": "b",
                "label": "Tidur lebih awal saja, barang-barang disiapkan pagi-pagi sebelum berangkat",
                "delta": 5,
                "hint": "Menyiapkan semuanya di pagi hari yang mepet berisiko ada barang tertinggal dan membuatmu terburu-buru"
              },
              {
                "id": "c",
                "label": "Langsung tidur tanpa persiapan apa pun, besok bawa saja apa yang terlihat",
                "delta": 0,
                "hint": "Berangkat tanpa langkah persiapan sama sekali berarti urutannya rusak sejak awal"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Rutinitas 2 - Mencuci Baju",
            "description": "Tumpukan baju kotor menumpuk. Langkah-langkahnya teracak: memisahkan, mencuci, lalu menjemur. Mana yang paling tepat dilakukan PERTAMA?",
            "options": [
              {
                "id": "a",
                "label": "Pisahkan dulu baju berdasarkan warna dan jenis bahannya sebelum dicuci",
                "delta": 20,
                "hint": "Memisahkan adalah langkah pertama yang benar: warna tidak luntur ke baju lain dan tiap bahan mendapat perlakuan yang tepat"
              },
              {
                "id": "b",
                "label": "Langsung cuci semuanya sekaligus agar cepat selesai",
                "delta": 0,
                "hint": "Mencuci tanpa memisahkan dulu membuat warna luntur dan merusak urutan kerja yang benar"
              },
              {
                "id": "c",
                "label": "Jemur dulu baju yang paling basah oleh keringat, urusan cuci nanti",
                "delta": 5,
                "hint": "Menjemur sebelum mencuci membalik urutan: baju kotor yang dijemur tetap kotor dan harus dicuci juga akhirnya"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Rutinitas 3 - Memasak Mie",
            "description": "Kamu ingin memasak mie. Langkah-langkahnya teracak: menyiapkan bahan, memasak, lalu menyajikan. Mana yang paling tepat dilakukan PERTAMA?",
            "options": [
              {
                "id": "a",
                "label": "Siapkan dulu semua bahan dan bumbunya, lalu didihkan air sebelum mie dimasukkan",
                "delta": 20,
                "hint": "Menyiapkan bahan dan mendidihkan air dulu adalah urutan pertama yang benar: mie masuk tepat saat air siap dan bumbu sudah tersedia"
              },
              {
                "id": "b",
                "label": "Masukkan mie ke air dingin lalu nyalakan kompor, bumbunya dicari sambil menunggu",
                "delta": 0,
                "hint": "Mie yang masuk sebelum air mendidih menjadi lembek, dan mencari bumbu sambil memasak membuat langkahnya berantakan"
              },
              {
                "id": "c",
                "label": "Tata dulu mangkuk saji dengan cantik, urusan bahan dan air menyusul",
                "delta": 5,
                "hint": "Menata sajian sebelum bahan siap berarti mendahulukan langkah terakhir sebelum langkah pertamanya dikerjakan"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Rutinitas 4 - Merapikan Kamar",
            "description": "Kamarmu berantakan. Langkah-langkahnya teracak: mengumpulkan, mengelompokkan, lalu mengembalikan ke tempatnya. Mana yang paling tepat dilakukan PERTAMA?",
            "options": [
              {
                "id": "a",
                "label": "Kumpulkan dulu semua barang yang berserakan ke satu tempat agar terlihat semuanya",
                "delta": 20,
                "hint": "Mengumpulkan dulu adalah langkah pertama yang benar: kamu baru bisa mengelompokkan dan mengembalikan setelah tahu semua yang berserakan"
              },
              {
                "id": "b",
                "label": "Langsung masukkan barang-barang ke lemari satu per satu begitu terlihat",
                "delta": 5,
                "hint": "Memasukkan tanpa mengumpulkan dan mengelompokkan dulu membuat lemari ikut berantakan dan ada barang yang terlewat"
              },
              {
                "id": "c",
                "label": "Sapu dan pel dulu lantainya, barang-barang digeser saja ke sudut",
                "delta": 0,
                "hint": "Mengepel sebelum barang dikumpulkan berarti membersihkan lantai yang masih tertutup barang berserakan"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Rutinitas 5 - Mengirim Pesan Penting",
            "description": "Kamu harus mengirim pesan penting ke seseorang. Langkah-langkahnya teracak: menyiapkan isi, memeriksa, lalu mengirim. Mana yang paling tepat dilakukan PERTAMA?",
            "options": [
              {
                "id": "a",
                "label": "Tulis dulu isi pesan dengan lengkap dan jelas sebelum melakukan apa pun",
                "delta": 20,
                "hint": "Menyiapkan isi dengan lengkap adalah langkah pertama yang benar: tidak ada yang bisa diperiksa atau dikirim sebelum isinya ada"
              },
              {
                "id": "b",
                "label": "Tekan kirim dulu agar cepat sampai, isinya dilengkapi lewat pesan susulan",
                "delta": 0,
                "hint": "Mengirim sebelum isi disiapkan membalik urutan sepenuhnya dan berisiko mengirim pesan kosong atau salah"
              },
              {
                "id": "c",
                "label": "Cek dulu nama penerima berkali-kali, isi pesan dipikir sambil jalan",
                "delta": 5,
                "hint": "Memeriksa penerima memang penting, tetapi itu langkah kedua setelah isi pesan disiapkan dengan lengkap"
              }
            ]
          },
          {
            "id": "ev_6",
            "title": "Rutinitas 6 - Mencuci Piring",
            "description": "Piring kotor menumpuk setelah makan. Langkah-langkahnya teracak: mengumpulkan, mencuci, lalu meniriskan. Mana yang paling tepat dilakukan PERTAMA?",
            "options": [
              {
                "id": "a",
                "label": "Kumpulkan semua piring kotor, buang sisa makanannya, dan rendam yang berlemak",
                "delta": 20,
                "hint": "Mengumpulkan dan merendam dulu adalah langkah pertama yang benar: kotoran melunak sehingga tahap mencuci jadi jauh lebih mudah"
              },
              {
                "id": "b",
                "label": "Langsung gosok satu per satu begitu diambil dari meja tanpa dikumpulkan dulu",
                "delta": 5,
                "hint": "Mencuci tanpa mengumpulkan dulu membuatmu bolak-balik dan yang berlemak sulit bersih karena tidak direndam"
              },
              {
                "id": "c",
                "label": "Susun dulu rak peniris dengan rapi, piring kotornya dicuci belakangan",
                "delta": 0,
                "hint": "Menyiapkan rak peniris sebelum ada piring yang dicuci berarti mendahulukan langkah terakhir"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Bikin Panduan Singkat Versi Kamu',
    'Punya cara andalan melakukan sesuatu? Tulis panduan singkat versimu! Pilih SATU aktivitas sehari-hari yang spesifik, lalu buat SATU dokumen panduan singkat (maksimal sekitar 1 halaman) yang bisa langsung dipraktikkan orang lain tanpa penjelasan tambahan. Ini panduan bacaan, bukan daftar centang.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-10-07',
    20,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "accepted_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4,
      "instruction": "Buat SATU dokumen panduan singkat dengan aturan: 1) Maksimal sekitar 1 halaman - singkat dan langsung ke intinya, bukan tulisan panjang berlembar-lembar. 2) Fokus pada SATU aktivitas sehari-hari yang spesifik (contoh: cara melipat baju agar rapi, cara menyimpan kabel, cara menyiapkan bekal, cara membersihkan sesuatu, atau cara mengatur barang). 3) Ditulis sebagai panduan bacaan yang runut (penjelasan langkah demi langkah dalam bentuk uraian atau paragraf bernomor) sehingga orang lain bisa langsung mempraktikkannya tanpa penjelasan tambahan - BUKAN berupa daftar centang (dokumen berisi kotak centang seperti [ ] ... akan ditolak karena itu jenis tugas yang berbeda). 4) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: aktivitas apa yang kamu pandu dan untuk siapa panduan itu berguna. Admin akan menolak dokumen berupa daftar centang, dokumen lebih dari sekitar 1 halaman, panduan yang tidak bisa langsung dipraktikkan, atau topik yang terlalu umum tanpa satu aktivitas yang spesifik."
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '099_tasks_2026_10_06_07')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
