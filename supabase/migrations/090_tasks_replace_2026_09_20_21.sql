-- ============================================================
-- Migration 090: Replace tasks for 20 + 21 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction/question/scenario text may be
-- hardcoded in frontend/backend source. Admin can edit every value below
-- through the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. DELETES exact tasks 550/551 on 2026-09-20 and 552/553 on
--    2026-09-21. Production evidence (audit at apply time): all four
--    carry 0 submissions in odyssey_task_submissions and 0 linked rows
--    in odyssey_coin_transactions (reference_id); odyssey_claims has no
--    task_id column, so no claim linkage is possible. Exact-ID
--    hard-delete is safe. No other rows exist on either date. History
--    of all other dates is untouched.
-- 2. Inserts 6 fresh tasks (3 per date, existing architecture and
--    config shapes only, no new framework, no new tables, no new
--    engines, no new upload system):
--    2026-09-20 (Decision making / Root-cause analysis / Observation):
--    - Step 1: MINI_GAME (Kalau Ini Terjadi, Kamu Ngapain?) — 5
--      realistic everyday situations (spilled water near power, long
--      queue, forgotten wallet, unresponsive groupmate, heavy rain),
--      each with 3 actions + deterministic delta + hint. DECISION_
--      PRIORITY + POINTS, initial 0, target 100; best path = 5 x 20 =
--      100. AUTO via the existing ValidateDecisionChoices server
--      recompute (client score never trusted; any complete choice set
--      accepted per engine contract). Reward 50c/100xp.
--    - Step 2: TEXT_RESPONSE (Temukan Penyebabnya) — member picks ONE
--      real everyday problem, lists possible causes, explains the
--      reasoning behind each, and proposes a solution. min 150 chars,
--      max 2000 (server-enforced by odyssey_submit_manual_task from
--      config); ADMIN_REVIEW. Reward 40c/100xp.
--    - Step 3: PHOTO_UPLOAD (Cari Hal yang Sering Diabaikan) — member
--      finds ONE small overlooked thing, photographs it, and explains
--      via payload.note what it is, why it deserves attention, and what
--      fix they propose. max_files 1, camera_only = false (gallery/
--      file-picker allowed via existing CameraCaptureModal path, same
--      as tasks 628/632); ADMIN_REVIEW via the existing approve/reject
--      flow. Reward 40c/100xp.
--    2026-09-21 (Practical reasoning / Process improvement / Teaching):
--    - Step 1: QUIZ (Mana yang Lebih Masuk Akal?) — 4 practical
--      reasoning questions (unit price, time buffer, slip risk,
--      long-run efficiency) with deterministic letter answers. All
--      numbers were independently calculated (see EDUCATIONAL
--      VALIDATION note below). AUTO via the existing
--      odyssey_submit_auto_task (all-correct-required) grading;
--      explanations render from config. Reward 50c/100xp.
--    - Step 2: TEXT_RESPONSE (Bikin Cara yang Lebih Praktis) — member
--      picks ONE routine activity that feels slow/complicated, describes
--      the current way step by step, proposes a more practical way, and
--      explains why (time/effort/cost). min 150, max 2000;
--      ADMIN_REVIEW. Reward 40c/100xp.
--    - Step 3: VIDEO (Jelaskan Tanpa Membaca) — member picks ONE simple
--      piece of knowledge/skill and teaches it from memory WITHOUT
--      reading (intro -> core explanation with example -> closing).
--      recording.enabled, max_duration_seconds = 60, camera user.
--      Resolves to ADMIN_REVIEW via existing
--      ResolveEvaluationTypeForConfig (same shape as tasks 616, 629,
--      631, 641, 647). Uses the existing VIDEO submission
--      infrastructure only. Reward 40c/100xp.
--    EDUCATIONAL VALIDATION (independent checks, 2026-09-19):
--    - Galon math: A = 20000/19 = 1052.63/L; B = 17000/15 = 1133.33/L
--      -> A cheaper per liter. Correct A.
--    - Time buffer: lesson 15:00, travel 45 min -> exact departure
--      14:15 (no buffer); 14:00 gives a 15-min buffer; 14:30 arrives
--      15:15 (late). Correct A.
--    - Slip risk: asking for a warning sign + waiting/using the dry
--      side is the only option that removes the hazard before moving.
--      Correct A.
--    - Lamp efficiency: LED 10W Rp35.000 / 15000h vs pijar 60W
--      Rp10.000 / 1000h at 5h/day -> LED lasts 3000 days (~8 years);
--      pijar needs 15 replacements (Rp150.000) plus ~6x electricity.
--      Correct A (LED clearly cheaper long-run).
--    - Scenario E: every event has 3 valid options; any complete choice
--      set is accepted (engine contract) with delta feedback; best-path
--      total = 100 points, worst = 0. No opinion grading.
-- 3. Idempotency: the pre-clean DELETEs below match ONLY the 6 exact
--    new titles on their exact dates (narrow exact-title guard for dev
--    re-runs, same spirit as migrations 052/083/084/085/086/087/088). No
--    broad date/title matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), 2026-09-13
--    (630/631/632), 2026-09-15 (633/634/635), 2026-09-16 (636/637/638),
--    2026-09-17 (639/640/641), 2026-09-18 (642/643/644), 2026-09-19
--    (645/646/647), tasks 608-620 (610 camera_only stays true),
--    payout/economy config, announcements, RLS, RPCs, triggers, claims,
--    shop, or any other date.
-- ============================================================

-- 1. Remove exact legacy tasks for 2026-09-20 / 2026-09-21 (proven dependency-free).
DELETE FROM odyssey_tasks
WHERE id IN (550, 551)
  AND active_date = '2026-09-20';

DELETE FROM odyssey_tasks
WHERE id IN (552, 553)
  AND active_date = '2026-09-21';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact dates). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-20'
  AND title IN (
    'Kalau Ini Terjadi, Kamu Ngapain?',
    'Temukan Penyebabnya',
    'Cari Hal yang Sering Diabaikan'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-21'
  AND title IN (
    'Mana yang Lebih Masuk Akal?',
    'Bikin Cara yang Lebih Praktis',
    'Jelaskan Tanpa Membaca'
  );

-- 3. Insert 3 fresh tasks for 2026-09-20.
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
    'Kalau Ini Terjadi, Kamu Ngapain?',
    'Hidup penuh kejadian tak terduga — yang membedakan adalah keputusanmu! Ada 5 situasi sehari-hari yang realistis: air tumpah dekat listrik, antrean panjang, dompet ketinggalan, teman kelompok yang hilang kabar, dan hujan deras. Pilih tindakan yang paling tepat untuk setiap situasi. Setiap pilihan ada konsekuensinya — bijaklah seperti dalam kehidupan nyata!',
    'MINI_GAME',
    'AUTO',
    1,
    '2026-09-20',
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
            "title": "Situasi 1 — Air Tumpah Dekat Stopkontak",
            "description": "Kamu tidak sengaja menumpahkan segelas air di meja, dan genangannya mengalir ke arah stopkontak dan kabel charger yang masih tercolok. Apa tindakan pertamamu?",
            "options": [
              {
                "id": "a",
                "label": "Cabut charger/matikan listrik di bagian itu dulu, baru lap genangan air sampai kering",
                "delta": 20,
                "hint": "Listrik + air itu berbahaya — amankan sumber bahayanya dulu sebelum membersihkan"
              },
              {
                "id": "b",
                "label": "Langsung lap airnya dengan kain tanpa mencabut charger",
                "delta": 5,
                "hint": "Membersihkan itu benar, tetapi tangan basah di dekat listrik yang masih menyala berisiko tersengat"
              },
              {
                "id": "c",
                "label": "Biarkan saja, nanti juga kering sendiri",
                "delta": 0,
                "hint": "Air yang dibiarkan dekat listrik bisa menyebabkan korsleting sebelum sempat kering"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 — Antrean Panjang",
            "description": "Kamu harus mengurus dokumen di loket. Antrean panjang dan nomor antreanmu baru dipanggil sekitar 15 menit lagi. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Siapkan semua dokumen dan syarat selagi antre, catat nomormu, dan perhatikan panggilan",
                "delta": 20,
                "hint": "Waktu menunggu jadi produktif dan kamu siap saat giliran tiba"
              },
              {
                "id": "b",
                "label": "Main HP untuk mengisi waktu tanpa menyiapkan apa pun",
                "delta": 5,
                "hint": "Waktu terbuang dan kamu bisa kelabakan mencari dokumen saat dipanggil"
              },
              {
                "id": "c",
                "label": "Tinggalkan antrean untuk jajan tanpa titip ke siapa pun",
                "delta": 0,
                "hint": "Nomormu bisa terlewat dan kamu harus antre ulang dari awal"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 — Dompet Ketinggalan",
            "description": "Di pintu keluar parkir kamu baru sadar dompet ketinggalan di rumah, sedangkan kamu harus membayar parkir sekarang. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Jelaskan dengan jujur ke petugas, hubungi keluarga/teman terdekat untuk bantuan, dan catat sebagai pengingat",
                "delta": 20,
                "hint": "Kejujuran + meminta bantuan dengan cara yang benar menyelesaikan masalah tanpa merugikan siapa pun"
              },
              {
                "id": "b",
                "label": "Berdebat dengan petugas agar digratiskan",
                "delta": 5,
                "hint": "Petugas hanya menjalankan tugasnya — berdebat membuang waktu dan tidak menyelesaikan apa pun"
              },
              {
                "id": "c",
                "label": "Kabur melewati palang tanpa membayar",
                "delta": 0,
                "hint": "Itu merugikan pengelola dan bisa berujung masalah yang jauh lebih besar"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 — Teman Kelompok Hilang Kabar",
            "description": "Tugas kelompok dikumpulkan besok. Satu anggota belum mengirim bagiannya dan tidak membalas chat sejak kemarin. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Hubungi dia baik-baik untuk memastikan kabarnya, tawarkan bantuan, siapkan rencana cadangan, dan kabari ketua/guru",
                "delta": 20,
                "hint": "Komunikasi yang baik + rencana cadangan menyelamatkan tugas tanpa menyalahkan siapa pun"
              },
              {
                "id": "b",
                "label": "Diam-diam kerjakan bagiannya sendiri tanpa memberi tahu siapa pun",
                "delta": 10,
                "hint": "Tugas selesai, tetapi tanpa komunikasi masalah yang sama akan terulang dan pembagian jadi tidak adil"
              },
              {
                "id": "c",
                "label": "Marah-marah di grup tanpa memberi solusi",
                "delta": 0,
                "hint": "Emosi tanpa solusi merusak kerja sama dan tidak mendekatkan tugas ke selesai"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 — Hujan Deras Tanpa Jas Hujan",
            "description": "Sepulang sekolah hujan turun sangat deras, kamu tidak membawa jas hujan maupun payung, dan keluarga belum bisa menjemput. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Berteduh di tempat yang aman, kabari keluarga tentang posisimu, dan tunggu hujan reda",
                "delta": 20,
                "hint": "Keselamatan dulu — memberi kabar membuat keluarga tenang dan tahu di mana mencarimu"
              },
              {
                "id": "b",
                "label": "Terobos hujan deras berjalan kaki sampai rumah",
                "delta": 5,
                "hint": "Bisa sampai lebih cepat, tetapi berisiko sakit, terpeleset, dan barang-barangmu rusak"
              },
              {
                "id": "c",
                "label": "Berteduh di bawah pohon besar atau baliho saat ada petir",
                "delta": 0,
                "hint": "Pohon dan baliho justru berbahaya saat petir dan angin kencang — pilih bangunan yang kokoh"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Temukan Penyebabnya',
    'Kenapa kran menetes terus padahal sudah diputar kencang? Kenapa nilai turun padahal belajar lebih lama? Setiap masalah punya penyebab — dan orang yang bisa menemukan penyebabnya bisa memperbaikinya. Pilih SATU masalah nyata di sekitarmu, selidiki kemungkinan penyebabnya, jelaskan alasanmu, dan tawarkan solusinya.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    2,
    '2026-09-20',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis hasil penyelidikanmu dengan format:\n1) Masalah yang kamu amati (sebutkan masalah nyata yang bisa diamati, bukan khayalan):\n2) Kemungkinan penyebab — tulis minimal 2 dugaan penyebab yang berbeda:\n3) Alasan — untuk setiap penyebab, jelaskan mengapa kamu menduga itu (fakta/pengamatan apa yang mendukung dugaanmu?):\n4) Solusi — untuk setiap penyebab, tulis solusi yang masuk akal:\n5) Solusi utama pilihanmu — penyebab mana yang paling mungkin dan solusi mana yang akan kamu jalankan duluan, beserta alasannya:\n\nWajib masalah nyata yang kamu amati sendiri. Jangan menyalin dari internet — tulis dengan bahasamu. Hindari masalah medis/kesehatan, dan jangan menyarankan tindakan berbahaya. Admin akan menolak jawaban tanpa masalah yang jelas, tanpa minimal 2 penyebab, tanpa alasan yang masuk akal, atau tanpa solusi.",
      "minimum_characters": 150,
      "maximum_characters": 2000
    }'::jsonb
),
(
    'demo-crew-1',
    'Cari Hal yang Sering Diabaikan',
    'Setiap hari kita melewati hal-hal kecil yang sebenarnya butuh perhatian: keran yang menetes, lampu yang kedip, sampah yang menumpuk di sudut, cat yang mengelupas, atau kabel yang berserakan. Temukan SATU hal kecil yang sering diabaikan di sekitarmu, foto, lalu jelaskan kenapa hal itu layak diperhatikan dan apa usulan perbaikanmu.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-09-20',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Cari SATU hal kecil yang sering diabaikan di rumah, sekolah, atau lingkungan sekitarmu (contoh: keran menetes, lampu kedip, sampah menumpuk, cat mengelupas, kabel berserakan). Jangan memotret orang lain tanpa izin dan jangan memotret hal yang memalukan/merugikan pihak tertentu. 2) Ambil/unggah 1 foto hal tersebut (boleh kamera langsung atau galeri). 3) Wajib tulis catatan dengan 4 bagian: (a) Temuan — apa yang kamu temukan dan di mana, (b) Penjelasan — kenapa hal ini bisa terjadi menurut pengamatanmu, (c) Kenapa penting — apa akibatnya jika terus diabaikan, (d) Usulan — perbaikan sederhana apa yang kamu sarankan dan siapa yang bisa melakukannya. Admin akan menolak submission tanpa foto, tanpa 4 bagian catatan, atau temuan yang tidak pantas."
    }'::jsonb
);

-- 4. Insert 3 fresh tasks for 2026-09-21.
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
    'Mana yang Lebih Masuk Akal?',
    'Jawaban yang terlihat murah, cepat, atau mudah belum tentu paling masuk akal! Ada 4 situasi sehari-hari soal harga, waktu, risiko, dan efisiensi. Hitung dan pikirkan baik-baik sebelum memilih — semua soal harus benar agar lolos. Bukan soal hafalan, tapi soal nalar!',
    'QUIZ',
    'AUTO',
    1,
    '2026-09-21',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "questions": [
        {
          "id": "q1",
          "question": "Air galon A: isi 19 liter seharga Rp20.000. Air galon B: isi 15 liter seharga Rp17.000. Kualitas air setara. Mana yang lebih hemat per liter?",
          "options": [
            "A. Galon A (sekitar Rp1.053 per liter)",
            "B. Galon B (sekitar Rp1.133 per liter)",
            "C. Sama hematnya",
            "D. Tidak bisa dibandingkan"
          ],
          "correct_answer": "A",
          "explanation": "Hitung harga per liter: A = 20.000 / 19 = sekitar Rp1.053 per liter. B = 17.000 / 15 = sekitar Rp1.133 per liter. Galon A lebih murah per liter, jadi lebih hemat meski harga labelnya lebih mahal."
        },
        {
          "id": "q2",
          "question": "Les dimulai pukul 15:00 dan perjalanan dari rumah memakan waktu 45 menit. Kapan waktu berangkat yang paling masuk akal?",
          "options": [
            "A. Pukul 14:00 (45 menit perjalanan + 15 menit cadangan)",
            "B. Pukul 14:15 (tiba tepat 15:00, tanpa cadangan)",
            "C. Pukul 14:30 (tiba 15:15, terlambat 15 menit)",
            "D. Pukul 13:00 (tiba 1 jam lebih awal)"
          ],
          "correct_answer": "A",
          "explanation": "Berangkat 14:15 memang tiba tepat 15:00 di atas kertas, tetapi tanpa cadangan sedikit pun — macet sebentar saja langsung terlambat. Berangkat 14:00 memberi 15 menit cadangan menghadapi hal tak terduga. Berangkat 14:30 jelas terlambat, dan 13:00 membuang satu jam tanpa perlu."
        },
        {
          "id": "q3",
          "question": "Lantai koridor baru saja dipel dan masih licin, tetapi kamu harus lewat untuk ke kelas. Tindakan mana yang paling masuk akal?",
          "options": [
            "A. Minta dipasangi tanda licin / tunggu sebentar sampai kering, lalu lewat sisi yang kering dengan pelan",
            "B. Berlari kecil agar cepat sampai ke sisi lain",
            "C. Melompat jauh melewati area yang basah",
            "D. Menyuruh teman lewat duluan sebagai percobaan"
          ],
          "correct_answer": "A",
          "explanation": "Satu-satunya pilihan yang menghilangkan bahayanya dulu sebelum bergerak. Berlari dan melompat di lantai licin justru memperbesar risiko terpeleset, dan menjadikan teman sebagai percobaan membahayakan orang lain."
        },
        {
          "id": "q4",
          "question": "Lampu LED: 10 watt, harga Rp35.000, tahan 15.000 jam. Lampu pijar: 60 watt, harga Rp10.000, tahan 1.000 jam. Dipakai 5 jam sehari. Mana yang lebih hemat untuk jangka panjang?",
          "options": [
            "A. Lampu LED (lebih awet dan jauh lebih hemat listrik)",
            "B. Lampu pijar (harga belinya lebih murah)",
            "C. Sama hematnya",
            "D. Tidak bisa dibandingkan"
          ],
          "correct_answer": "A",
          "explanation": "Dipakai 5 jam sehari, LED tahan 15.000 / 5 = 3.000 hari (sekitar 8 tahun). Lampu pijar hanya tahan 1.000 / 5 = 200 hari, sehingga butuh 15 kali ganti (Rp150.000) plus listrik sekitar 6 kali lipat (60 watt vs 10 watt). Harga beli yang murah tidak berarti hemat jangka panjang."
        }
      ]
    }'::jsonb
),
(
    'demo-crew-1',
    'Bikin Cara yang Lebih Praktis',
    'Setiap hari kita melakukan hal yang sama berulang-ulang: menyiapkan bekal, merapikan kamar, mencatat tugas, mencuci piring, berangkat sekolah. Pasti ada yang terasa ribet atau lambat! Pilih SATU aktivitas rutinmu, jelaskan caramu sekarang, lalu rancang cara yang lebih praktis beserta alasannya.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    2,
    '2026-09-21',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis usulan perbaikanmu dengan format:\n1) Aktivitas rutin pilihanmu (kegiatan nyata yang sering kamu lakukan):\n2) Cara sekarang — jelaskan langkah per langkah bagaimana kamu melakukannya selama ini:\n3) Masalahnya — bagian mana yang terasa ribet, lambat, atau boros (waktu/tenaga/biaya)?\n4) Cara baru usulmu — jelaskan langkah per langkah cara yang lebih praktis:\n5) Alasan — jelaskan kenapa cara barumu lebih praktis (hemat waktu berapa, hemat tenaga/biaya apa, atau lebih rapi/aman bagaimana):\n\nWajib aktivitas nyata yang kamu lakukan sendiri. Jangan menyalin dari internet — tulis dengan bahasamu. Usulan harus aman dan bisa benar-benar dilakukan. Admin akan menolak jawaban tanpa aktivitas yang jelas, tanpa cara sekarang, tanpa cara baru yang konkret, atau tanpa alasan yang masuk akal.",
      "minimum_characters": 150,
      "maximum_characters": 2000
    }'::jsonb
),
(
    'demo-crew-1',
    'Jelaskan Tanpa Membaca',
    'Orang yang benar-benar paham bisa menjelaskan tanpa mencontek teks! Pilih SATU pengetahuan atau hal sederhana yang kamu kuasai — misalnya kenapa langit biru, cara menghemat baterai HP, cara mencuci tangan yang benar, atau pengetahuan praktis lainnya. Rekam video maksimal 60 detik yang MENGAJARKANNYA dari ingatanmu, seolah kamu mengajari seorang teman. Dilarang membaca teks/catatan selama merekam!',
    'VIDEO',
    'ADMIN_REVIEW',
    3,
    '2026-09-21',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik yang mengajarkan SATU pengetahuan sederhana SEPENUHNYA DARI INGATAN (tanpa membaca teks, catatan, atau layar):\n1. Pembuka — sebutkan apa yang akan kamu jelaskan dan kenapa itu berguna\n2. Penjelasan inti — jelaskan dengan kata-katamu sendiri + beri SATU contoh nyata agar mudah dipahami\n3. Penutup — rangkum dalam satu kalimat kunci yang mudah diingat\n\nDILARANG membaca teks/catatan/layar selama merekam — admin akan menolak video yang terlihat membaca. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau gambarmu sambil menjelaskan dengan suaramu. Pilih topik yang aman dan pantas. Jangan menyebut data pribadi sensitif. Bicaralah senatural mungkin seperti mengajari teman!"
      }
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '090_tasks_replace_2026_09_20_21')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
