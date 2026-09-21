-- ============================================================
-- Migration 091: Replace tasks for 22 + 23 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction/question/scenario text may be
-- hardcoded in frontend/backend source. Admin can edit every value below
-- through the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. DELETES exact tasks 554/555 on 2026-09-22 and 556/557 on
--    2026-09-23. Production evidence (audit at apply time): all four
--    carry 0 submissions in odyssey_task_submissions and 0 linked rows
--    in odyssey_coin_transactions (reference_id); odyssey_claims has no
--    task_id column, so no claim linkage is possible. Exact-ID
--    hard-delete is safe. No other rows exist on either date. History
--    of all other dates is untouched.
-- 2. Inserts 6 fresh tasks (3 per date, existing architecture and
--    config shapes only, no new framework, no new tables, no new
--    engines, no new upload system):
--    2026-09-22 (How things work / Clear instructions / Observation):
--    - Step 1: MINI_GAME (Bongkar Cara Kerjanya) — 5 everyday process
--      situations (rice cooker jeglek, stuffy room with fan, clogged
--      drain, overheating phone, rising electricity bill), each with
--      3 actions + deterministic delta + hint. DECISION_PRIORITY +
--      POINTS, initial 0, target 100; best path = 5 x 20 = 100. AUTO
--      via the existing ValidateDecisionChoices server recompute
--      (client score never trusted; any complete choice set accepted
--      per engine contract). No drag/drop or ordering mechanic is
--      assumed — deterministic scenario/choice only. Reward 50c/100xp.
--    - Step 2: DOCUMENT_UPLOAD (Bikin Petunjuk yang Nggak Bikin
--      Bingung) — member writes ONE clear step-by-step guide others
--      can follow (title, materials/requirements, numbered concrete
--      steps, 1 success tip). Same config shape as tasks 634/636
--      (instruction + allowed_extensions + accepted_extensions +
--      max_file_size_mb 4); existing DocUploadModal path only.
--      ADMIN_REVIEW. Reward 40c/100xp.
--    - Step 3: PHOTO_UPLOAD (Apa yang Bisa Kamu Pelajari dari Ini?)
--      — member photographs ONE real thing/situation nearby and
--      explains via payload.note what it is, what can be learned,
--      and how that knowledge is useful. max_files 1, camera_only =
--      false (gallery/file-picker allowed via existing
--      CameraCaptureModal path, same as tasks 628/632/643/650);
--      ADMIN_REVIEW via the existing approve/reject flow. Reward
--      40c/100xp.
--    2026-09-23 (Money tracking / Verify-before-trust / Storytelling):
--    - Step 1: QUIZ (Uangnya Lari ke Mana?) — 4 practical money
--      scenarios (total spending + remainder, biggest expense,
--      cheaper-per-unit choice, impact of a daily habit) with
--      deterministic letter answers. All numbers were independently
--      calculated (see EDUCATIONAL VALIDATION note below). AUTO via
--      the existing odyssey_submit_auto_task (all-correct-required)
--      grading; explanations render from config. Reward 50c/100xp.
--    - Step 2: TEXT_RESPONSE (Cek Sebelum Percaya) — member picks ONE
--      simple info/ad/offer, explains what to check, what warning
--      signs to watch, what to verify, and why checking matters.
--      min 150 chars, max 2000 (server-enforced by
--      odyssey_submit_manual_task from config); ADMIN_REVIEW. No
--      URL-verification engine, no fact-check service. Reward
--      40c/100xp.
--    - Step 3: VIDEO (Cerita dari Pengalamanmu) — member tells ONE
--      real experience (what happened -> what was learned -> what
--      they would do differently now). recording.enabled,
--      max_duration_seconds = 60, camera user. Resolves to
--      ADMIN_REVIEW via existing ResolveEvaluationTypeForConfig
--      (same shape as tasks 616, 629, 631, 641, 647, 653). Uses the
--      existing VIDEO submission infrastructure only. Reward
--      40c/100xp.
--    EDUCATIONAL VALIDATION (independent checks, 2026-09-21):
--    - Q1 remainder: 12000 + 15000 + 8000 = 35000; 50000 - 35000 =
--      15000. Correct A.
--    - Q2 biggest expense: A = 10000 x 20 = 200000; B = 75000;
--      C = 50000 x 2 = 100000; D = 60000. Biggest is A. Correct A.
--    - Q3 unit price: A = 15000 / 5 = 3000 per pack; B = 33600 / 12
--      = 2800 per pack. B cheaper per pack. Correct B.
--    - Q4 habit impact: 8000 x 30 = 240000 per month. Correct A.
--    - Scenario E: every event has 3 valid options; any complete
--      choice set is accepted (engine contract) with delta feedback;
--      best-path total = 100 points, worst = 0. No opinion grading.
-- 3. Idempotency: the pre-clean DELETEs below match ONLY the 6 exact
--    new titles on their exact dates (narrow exact-title guard for dev
--    re-runs, same spirit as migrations 052/083/084/085/086/087/088/
--    090). No broad date/title matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), 2026-09-13
--    (630/631/632), 2026-09-15 (633/634/635), 2026-09-16 (636/637/638),
--    2026-09-17 (639/640/641), 2026-09-18 (642/643/644), 2026-09-19
--    (645/646/647), 2026-09-20 (648/649/650), 2026-09-21 (651/652/653),
--    tasks 608-620 (610 camera_only stays true), payout/economy
--    config, announcements, RLS, RPCs, triggers, claims, shop, or any
--    other date.
-- ============================================================

-- 1. Remove exact legacy tasks for 2026-09-22 / 2026-09-23 (proven dependency-free).
DELETE FROM odyssey_tasks
WHERE id IN (554, 555)
  AND active_date = '2026-09-22';

DELETE FROM odyssey_tasks
WHERE id IN (556, 557)
  AND active_date = '2026-09-23';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact dates). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-22'
  AND title IN (
    'Bongkar Cara Kerjanya',
    'Bikin Petunjuk yang Nggak Bikin Bingung',
    'Apa yang Bisa Kamu Pelajari dari Ini?'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-23'
  AND title IN (
    'Uangnya Lari ke Mana?',
    'Cek Sebelum Percaya',
    'Cerita dari Pengalamanmu'
  );

-- 3. Insert 3 fresh tasks for 2026-09-22.
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
    'Bongkar Cara Kerjanya',
    'Kenapa rice cooker bisa jeglek sendiri? Kenapa kipas tidak membuat ruangan dingin? Kenapa got meluap tiap hujan? Semua ada cara kerjanya! Hadapi 5 situasi sehari-hari, pilih tindakan yang menunjukkan kamu paham BAGAIMANA dan MENGAPA sesuatu bekerja seperti itu. Setiap pilihan ada konsekuensinya — yang paham prosesnya akan memilih dengan tepat!',
    'MINI_GAME',
    'AUTO',
    1,
    '2026-09-22',
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
            "title": "Situasi 1 — Rice Cooker Jeglek Sebelum Matang",
            "description": "Nasi belum matang tetapi tombol rice cooker sudah jeglek ke posisi warm. Kamu butuh nasinya matang untuk makan bersama. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Cek dulu penyebabnya: pastikan takaran air cukup dan bersihkan kerak di dasar panci, lalu masak ulang dengan benar",
                "delta": 20,
                "hint": "Tombol jeglek itu mekanisme pengaman — memahami cara kerjanya berarti mencari penyebabnya, bukan memaksa alatnya"
              },
              {
                "id": "b",
                "label": "Ganjal atau tahan tombol cook-nya agar tidak jeglek",
                "delta": 5,
                "hint": "Memaksa mekanisme pengaman bisa merusak alat dan berbahaya — masalah aslinya tidak selesai"
              },
              {
                "id": "c",
                "label": "Buang rice cooker-nya dan beli yang baru",
                "delta": 0,
                "hint": "Tanpa tahu penyebabnya, alat baru bisa mengalami hal yang sama"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 — Kamar Gerah Padahal Kipas Kencang",
            "description": "Kamar terasa gerah dan pengap meski kipas angin sudah menyala paling kencang. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Buka jendela atau ventilasi agar udara bersirkulasi — kipas hanya menggerakkan udara, bukan mendinginkan ruangan",
                "delta": 20,
                "hint": "Memahami cara kerja kipas: angin terasa sejuk karena membantu penguapan keringat, tetapi butuh udara segar yang masuk"
              },
              {
                "id": "b",
                "label": "Dekatkan kipas ke wajah sedekat mungkin dan tutup rapat semua ventilasi",
                "delta": 5,
                "hint": "Udara pengap hanya diputar-putar di tempat — ruangan tetap panas dan pengap"
              },
              {
                "id": "c",
                "label": "Biarkan kipas menyala seharian penuh tanpa henti agar lama-lama dingin",
                "delta": 0,
                "hint": "Kipas tidak menurunkan suhu ruangan; menyala terus hanya menambah tagihan listrik"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 — Got Depan Rumah Meluap Tiap Hujan",
            "description": "Setiap hujan deras, saluran air depan rumah meluap dan menggenangi jalan. Warga mulai resah. Apa tindakan yang paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Periksa salurannya: angkat sampah dan sumbatan yang menghalangi aliran, lalu usulkan jadwal bersih rutin",
                "delta": 20,
                "hint": "Air meluap karena alirannya tersumbat — atasi penyebabnya, bukan gejalanya"
              },
              {
                "id": "b",
                "label": "Tinggikan pagar dan teras rumahmu sendiri agar air tidak masuk",
                "delta": 10,
                "hint": "Rumahmu aman sementara, tetapi saluran tetap mampet dan air makin meluap ke tetangga"
              },
              {
                "id": "c",
                "label": "Biarkan saja, nanti juga surut sendiri kalau hujan berhenti",
                "delta": 0,
                "hint": "Sumbatan tidak hilang sendiri; hujan berikutnya meluap lagi, bahkan bisa lebih parah"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 — HP Panas Saat Main Sambil Ngecas",
            "description": "HP terasa sangat panas karena dipakai bermain game sambil diisi daya, dan diletakkan di atas bantal. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Berhenti memakai HP untuk hal berat selama mengisi daya, dan pindahkan dari bantal agar panasnya bisa keluar",
                "delta": 20,
                "hint": "Dua sumber panas sekaligus ditambah ventilasi tertutup — memahami sebabnya berarti menghilangkan penyebabnya"
              },
              {
                "id": "b",
                "label": "Masukkan HP ke kulkas atau freezer sebentar agar cepat dingin",
                "delta": 5,
                "hint": "Perubahan suhu ekstrem bisa menimbulkan embun di dalam HP dan merusak komponen"
              },
              {
                "id": "c",
                "label": "Teruskan bermain, panas itu normal dan tidak apa-apa",
                "delta": 0,
                "hint": "Panas berlebih yang dibiarkan mempercepat kerusakan baterai dan bisa berbahaya"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 — Tagihan Listrik Naik Tanpa Sebab Jelas",
            "description": "Tagihan listrik bulan ini naik padahal kamu merasa pemakaiannya sama saja. Keluarga ingin tahu uangnya lari ke mana. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Catat pemakaian tiap alat (berapa watt, berapa jam menyala), cabut yang standby, lalu atur prioritas pemakaian",
                "delta": 20,
                "hint": "Memahami ke mana listrik pergi berarti mengukur dulu, bukan menebak-nebak"
              },
              {
                "id": "b",
                "label": "Matikan semua lampu setiap hari meski masih dipakai",
                "delta": 5,
                "hint": "Hemat sesaat tetapi penyebab sebenarnya tidak ketemu — bisa jadi alat lain yang boros"
              },
              {
                "id": "c",
                "label": "Marahi anggota keluarga agar tidak memakai listrik",
                "delta": 0,
                "hint": "Tanpa data, hemat jadi tebak-tebakan dan menyalahkan orang tanpa solusi"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Bikin Petunjuk yang Nggak Bikin Bingung',
    'Pernah membaca petunjuk yang malah bikin bingung? Sekarang giliranmu membuat yang benar! Buat SATU panduan singkat yang jelas, berurutan, dan mudah diikuti orang lain — misalnya cara packing barang, cara memakai alat sederhana, cara melakukan aktivitas kerja atau aktivitas rutin. Panduan yang bagus membuat orang lain bisa berhasil tanpa bertanya-tanya.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-09-22',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "instruction": "Buat 1 panduan praktis yang bisa langsung dipakai orang lain tanpa penjelasan tambahan. Pilih SATU topik yang kamu kuasai (contoh: cara packing barang, cara memakai alat sederhana, cara melakukan aktivitas kerja sederhana, atau cara melakukan aktivitas rutin). Isi panduan WAJIB memuat: (1) judul yang jelas, (2) tujuan panduan + siapa yang bisa mengikutinya, (3) daftar bahan, alat, atau syarat sebelum mulai (jika ada), (4) langkah-langkah BERNOMOR yang berurutan dan konkret (setiap langkah bisa langsung dilakukan tanpa menebak), (5) 1 tips agar berhasil atau kesalahan umum yang harus dihindari. Boleh diketik di HP lalu disimpan sebagai PDF/DOCX/TXT, atau ditulis tangan rapi lalu difoto jelas (JPG/PNG). Maksimal 4 MB. Admin akan menolak file yang tidak bisa dibuka, tulisan yang bukan panduan (misalnya curhat atau essay tanpa langkah), panduan tanpa langkah bernomor yang konkret, atau langkah yang tidak berurutan.",
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "accepted_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4
    }'::jsonb
),
(
    'demo-crew-1',
    'Apa yang Bisa Kamu Pelajari dari Ini?',
    'Belajar tidak selalu dari buku — benda dan situasi di sekitarmu menyimpan pelajaran! Temukan SATU hal nyata di sekitar (benda, tanaman, hewan, bangunan, atau situasi sehari-hari), foto, lalu jelaskan apa yang kamu temukan, apa yang bisa dipelajari darinya, dan bagaimana pengetahuan itu bisa berguna dalam hidupmu.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-09-22',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Temukan SATU benda atau situasi nyata di sekitarmu (contoh: retakan di tembok, sarang semut, tanaman liar yang tumbuh di celah, tumpukan barang, atau situasi sehari-hari yang menarik). Jangan memotret orang lain tanpa izin dan jangan memotret hal yang memalukan atau merugikan pihak tertentu. 2) Ambil atau unggah 1 foto hal tersebut (boleh kamera langsung atau galeri). 3) Wajib tulis catatan dengan 3 bagian: (a) Temuan — apa yang kamu temukan dan di mana kamu menemukannya, (b) Pelajaran — apa yang bisa dipelajari dari temuan itu (bagaimana hal itu bisa terjadi atau bekerja menurut pengamatanmu), (c) Manfaat — bagaimana pengetahuan tersebut bisa berguna untukmu atau orang lain. Admin akan menolak submission tanpa foto, tanpa 3 bagian catatan, atau temuan yang tidak pantas."
    }'::jsonb
);

-- 4. Insert 3 fresh tasks for 2026-09-23.
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
    'Uangnya Lari ke Mana?',
    'Uang saku habis tapi bingung ke mana perginya? Saatnya jadi detektif keuangan! Ada 4 situasi pengeluaran sehari-hari: hitung totalnya, temukan sisa uangnya, bandingkan pilihan yang ada, dan lihat dampak kebiasaan kecil. Semua soal harus benar agar lolos — kalkulator boleh, yang penting paham cara menghitungnya!',
    'QUIZ',
    'AUTO',
    1,
    '2026-09-23',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "questions": [
        {
          "id": "q1",
          "question": "Kamu membawa uang Rp50.000 ke sekolah. Pengeluaran hari ini: ongkos Rp12.000, jajan Rp15.000, fotokopi Rp8.000. Berapa sisa uangmu?",
          "options": [
            "A. Rp15.000",
            "B. Rp20.000",
            "C. Rp25.000",
            "D. Rp35.000"
          ],
          "correct_answer": "A",
          "explanation": "Total pengeluaran = 12.000 + 15.000 + 8.000 = Rp35.000. Sisa = 50.000 - 35.000 = Rp15.000. Selalu jumlahkan dulu semua pengeluaran sebelum menghitung sisa."
        },
        {
          "id": "q2",
          "question": "Pengeluaran bulan ini: (A) jajan Rp10.000 x 20 hari, (B) langganan Rp75.000, (C) bensin Rp50.000 x 2 kali, (D) pulsa Rp60.000. Mana pengeluaran TERBESAR?",
          "options": [
            "A. Jajan (Rp200.000)",
            "B. Langganan (Rp75.000)",
            "C. Bensin (Rp100.000)",
            "D. Pulsa (Rp60.000)"
          ],
          "correct_answer": "A",
          "explanation": "Hitung masing-masing: A = 10.000 x 20 = Rp200.000; B = Rp75.000; C = 50.000 x 2 = Rp100.000; D = Rp60.000. Pengeluaran kecil yang berulang (jajan) ternyata paling besar totalnya."
        },
        {
          "id": "q3",
          "question": "Mi instan A: 5 bungkus seharga Rp15.000. Mi instan B: 12 bungkus seharga Rp33.600. Isinya sama. Mana yang lebih hemat per bungkus?",
          "options": [
            "A. Produk A (Rp3.000 per bungkus)",
            "B. Produk B (Rp2.800 per bungkus)",
            "C. Sama hematnya",
            "D. Tidak bisa dibandingkan"
          ],
          "correct_answer": "B",
          "explanation": "Harga per bungkus: A = 15.000 / 5 = Rp3.000. B = 33.600 / 12 = Rp2.800. Produk B lebih murah Rp200 per bungkus, jadi lebih hemat meski harga labelnya lebih mahal."
        },
        {
          "id": "q4",
          "question": "Kamu terbiasa beli minum kemasan Rp8.000 setiap hari. Jika kebiasaan itu berlangsung 30 hari, berapa total uang yang keluar?",
          "options": [
            "A. Rp240.000",
            "B. Rp180.000",
            "C. Rp80.000",
            "D. Rp8.000"
          ],
          "correct_answer": "A",
          "explanation": "Total = 8.000 x 30 = Rp240.000 dalam sebulan. Pengeluaran kecil harian yang terlihat sepele bisa menjadi besar saat dikalikan waktu — membawa botol isi ulang bisa menghemat hampir semuanya."
        }
      ]
    }'::jsonb
),
(
    'demo-crew-1',
    'Cek Sebelum Percaya',
    'Diskon 90 persen? Hadiah undian? Kabar yang bikin panik? Tidak semua informasi bisa langsung dipercaya! Pilih SATU contoh informasi, iklan, atau penawaran sederhana yang pernah kamu lihat, lalu jelaskan hal apa yang perlu diperiksa, tanda apa yang perlu dicurigai, informasi apa yang perlu diverifikasi, dan kenapa pemeriksaan itu penting.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    2,
    '2026-09-23',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis hasil pemeriksaanmu dengan format:\n1) Informasi pilihanmu (contoh informasi, iklan, atau penawaran sederhana yang pernah kamu lihat — tulis isi atau bunyinya dengan bahasamu):\n2) Hal yang perlu diperiksa — tulis minimal 2 hal yang harus dicek sebelum percaya (contoh: siapa pengirimnya, apakah ada syarat tersembunyi, apakah harganya masuk akal):\n3) Tanda yang perlu dicurigai — tulis minimal 2 tanda bahaya (contoh: disuruh buru-buru, diminta data pribadi atau uang muka, tidak ada alamat atau kontak yang jelas):\n4) Cara memverifikasi — jelaskan bagaimana kamu memastikan kebenarannya (contoh: cek ke sumber resmi, tanya orang dewasa yang paham, bandingkan dengan info lain):\n5) Alasan — jelaskan kenapa pemeriksaan seperti ini penting dan apa risikonya jika langsung percaya:\n\nWajib contoh yang nyata dan sederhana. Jangan menyebar ulang kabar bohong — tulis dengan bahasamu sebagai bahan latihan. Jangan mencantumkan data pribadi siapa pun. Admin akan menolak jawaban tanpa contoh yang jelas, tanpa hal yang diperiksa, tanpa tanda curiga, tanpa cara verifikasi, atau tanpa alasan yang masuk akal.",
      "minimum_characters": 150,
      "maximum_characters": 2000
    }'::jsonb
),
(
    'demo-crew-1',
    'Cerita dari Pengalamanmu',
    'Setiap pengalaman menyimpan pelajaran — dan pelajaran itu lebih berharga saat dibagikan! Ceritakan SATU pengalaman nyata dari kerja, sekolah, kehidupan sehari-hari, organisasi, atau aktivitas lainnya. Jelaskan apa yang terjadi, apa yang kamu pelajari, dan apa yang akan kamu lakukan berbeda sekarang. Rekam dalam video maksimal 60 detik, senatural mungkin seperti bercerita ke teman!',
    'VIDEO',
    'ADMIN_REVIEW',
    3,
    '2026-09-23',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik yang menceritakan SATU pengalaman nyatamu dengan struktur:\n1. Kejadian — ceritakan apa yang terjadi (kapan, di mana, siapa yang terlibat, bagaimana kejadiannya)\n2. Pelajaran — apa yang kamu pelajari dari pengalaman itu\n3. Perubahan — apa yang akan kamu lakukan berbeda sekarang jika mengalami hal serupa, dan kenapa\n\nPilih pengalaman yang aman dan pantas untuk dibagikan. Jangan menyebut data pribadi sensitif (milikmu maupun orang lain) dan jangan menjelekkan pihak tertentu. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau gambarmu sambil bercerita dengan suaramu. Bicaralah senatural mungkin seperti bercerita ke seorang teman!"
      }
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '091_tasks_replace_2026_09_22_23')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
