-- ============================================================
-- Migration 088: Replace tasks for 18 + 19 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction/question/scenario text may be
-- hardcoded in frontend/backend source. Admin can edit every value below
-- through the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. DELETES exact tasks 546/547 on 2026-09-18 and 548/549 on
--    2026-09-19. Production evidence (audit at apply time): all four
--    carry 0 submissions in odyssey_task_submissions and 0 linked rows
--    in odyssey_coin_transactions (reference_id); odyssey_claims has no
--    task_id column, so no claim linkage is possible. Exact-ID
--    hard-delete is safe. No other rows exist on either date. History
--    of all other dates is untouched.
-- 2. Inserts 6 fresh tasks (3 per date, existing architecture and
--    config shapes only, no new framework, no new tables, no new
--    engines, no new upload system):
--    2026-09-18 (Research / Experiment / Critical thinking):
--    - Step 1: TEXT_RESPONSE (Kenapa Barang Ini Bisa Begitu?) — member
--      picks ONE ordinary object, investigates why it works / is
--      designed that way, cites where they learned it, explains in
--      their own words + what they learned. min 150 chars, max 2000
--      (server-enforced by odyssey_submit_manual_task from config);
--      ADMIN_REVIEW. No purchase, no sensitive data, no medical
--      claims, no dangerous experiments.
--    - Step 2: PHOTO_UPLOAD (Uji Coba Sederhana) — member runs ONE
--      safe household micro-experiment and reports Prediksi → Percobaan
--      → Hasil → Kesimpulan via the existing payload.note field (same
--      convention as DOCUMENT_UPLOAD and task 628), plus 1 photo of
--      the actual setup/result. max_files 1, camera_only = false
--      (gallery/file-picker allowed via existing CameraCaptureModal
--      path); ADMIN_REVIEW via the existing approve/reject flow.
--      Safety rules (no fire/chemicals/electricity/danger) live in
--      config.instruction — enforced by admin review, not new code.
--    - Step 3: QUIZ (Mitos atau Fakta?) — 5 everyday claims, each
--      MITOS vs FAKTA with correct_answer + explanation in config.
--      AUTO evaluation via the existing odyssey_submit_auto_task
--      (all-correct-required) grading; explanations render in the
--      existing VideoQuizModal from config. Every claim was selected
--      for objective verifiability (see EDUCATIONAL VALIDATION note
--      below). Reward 50c/100xp.
--    2026-09-19 (Financial literacy / Problem solving / Teaching):
--    - Step 1: QUIZ (Hitung Mana yang Lebih Hemat?) — 4 unit-pricing /
--      discount / bundle scenarios with deterministic letter answers.
--      All numbers were independently calculated (see EDUCATIONAL
--      VALIDATION note below). AUTO via the same existing quiz path.
--      Reward 50c/100xp.
--    - Step 2: MINI_GAME (Cari Jalan Keluar dari Masalah) — decision
--      scenario (DECISION_PRIORITY + POINTS, initial 0, 4 events) about
--      handling low phone battery + unfamiliar location + one important
--      message under time pressure. Reuses the existing
--      DecisionGameModal + server-authoritative balance recompute
--      (ValidateDecisionChoices; client score never trusted). Reward is
--      completion-based per the existing engine contract; deltas give
--      per-choice feedback. Reward 50c/100xp.
--    - Step 3: VIDEO (Ajarkan Satu Skill) — member records ~60 seconds
--      TEACHING one real skill with a visible demonstration
--      (Masalah/Tujuan → Cara → Demonstrasi → Hasil). recording.enabled,
--      max_duration_seconds = 60, camera user. Resolves to ADMIN_REVIEW
--      via existing ResolveEvaluationTypeForConfig (same shape as tasks
--      616, 629, 631, 641). Uses the existing VIDEO submission
--      infrastructure only. Reward 40c/100xp.
--    EDUCATIONAL VALIDATION (independent checks, 2026-09-17):
--    - Mitos/Fakta answers: petir repeat strikes (MITOS — documented
--      repeat strikes on tall structures); water 100C (MITOS — only at
--      1 atm sea level, lower at altitude); heavier falls faster
--      (MITOS — Galileo/Apollo hammer-feather, air drag is shape not
--      weight); 20s soap handwash (FAKTA — WHO/CDC guidance); must
--      drain Li-ion to 0% (MITOS — partial charges extend Li-ion life).
--    - Hemat math: 18000/250=72 vs 25000/400=62.5 → B; 22000/L vs
--      40000/2=20000/L → B; 100000*0.8=80000 vs 90000*0.9=81000 → A;
--      65000/5=13000/kg vs 120000/10=12000/kg → B.
--    - Scenario E: every event has >=2 valid options; any complete
--      choice set is accepted (engine contract) with delta feedback;
--      best-path total = 80 points, worst = 0. No opinion grading.
-- 3. Idempotency: the pre-clean DELETEs below match ONLY the 6 exact
--    new titles on their exact dates (narrow exact-title guard for dev
--    re-runs, same spirit as migrations 052/083/084/085/086/087). No
--    broad date/title matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), 2026-09-13
--    (630/631/632), 2026-09-15 (633/634/635), 2026-09-16 (636/637/638),
--    2026-09-17, tasks 608-620 (610 camera_only stays true),
--    payout/economy config, announcements, RLS, RPCs, triggers, claims,
--    shop, or any other date.
-- ============================================================

-- 1. Remove exact legacy tasks for 2026-09-18 / 2026-09-19 (proven dependency-free).
DELETE FROM odyssey_tasks
WHERE id IN (546, 547)
  AND active_date = '2026-09-18';

DELETE FROM odyssey_tasks
WHERE id IN (548, 549)
  AND active_date = '2026-09-19';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact dates). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-18'
  AND title IN (
    'Kenapa Barang Ini Bisa Begitu?',
    'Uji Coba Sederhana',
    'Mitos atau Fakta?'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-19'
  AND title IN (
    'Hitung Mana yang Lebih Hemat?',
    'Cari Jalan Keluar dari Masalah',
    'Ajarkan Satu Skill'
  );

-- 3. Insert 3 fresh tasks for 2026-09-18.
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
    'Kenapa Barang Ini Bisa Begitu?',
    'Pilih SATU benda biasa yang ada di sekitarmu — misalnya charger yang menghangat, botol dengan bentuk tertentu, ban dengan pola kembang, remote yang harus diarahkan, atau benda sehari-hari lainnya. Cari tahu kenapa benda itu bekerja atau didesain seperti itu, lalu jelaskan temuanmu dengan kata-katamu sendiri. Tidak perlu ilmu tingkat lanjut dan tidak perlu membeli apa pun — cukup rasa ingin tahu dan kemauan mencari tahu.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    1,
    '2026-09-18',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis hasil penyelidikanmu dengan format:\n1) Benda yang kamu pilih (sebutkan nama bendanya):\n2) Pertanyaanmu tentang benda itu (contoh: kenapa charger jadi hangat? kenapa ban ada pola kembangnya?):\n3) Penjelasan yang kamu temukan (dengan kata-katamu sendiri, minimal 3 kalimat):\n4) Dari mana kamu mendapatkan penjelasan itu (misalnya: label kemasan, buku pelajaran, video edukasi, bertanya ke orang yang paham, artikel — tulis sejelas mungkin):\n5) Satu hal baru yang kamu pelajari:\n\nWajib memilih benda nyata yang bisa diamati. Jangan menyalin mentah-mentah dari internet — tulis ulang dengan bahasamu. Hindari klaim kesehatan/medis dan hindari percobaan berbahaya (tidak perlu membongkar atau mengubah apa pun). Admin akan menolak jawaban tanpa benda yang jelas, tanpa penjelasan dengan kata sendiri, tanpa sumber belajar, atau tanpa hal yang dipelajari.",
      "minimum_characters": 150,
      "maximum_characters": 2000
    }'::jsonb
),
(
    'demo-crew-1',
    'Uji Coba Sederhana',
    'Lakukan SATU percobaan kecil yang aman dengan benda-benda yang sudah ada di rumah — misalnya menguji bahan mana yang paling cepat menyerap air, di mana es batu paling cepat mencair, atau susunan mana yang membuat benda lebih stabil. Tulis prediksimu SEBELUM mencoba, lakukan percobaannya, foto kondisi atau hasilnya, lalu tulis kesimpulanmu. Tidak boleh ada api, bahan kimia, listrik, benda tajam berbahaya, atau apa pun yang berisiko melukai.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-09-18',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "1) Pilih SATU uji coba sederhana yang AMAN dari contoh di deskripsi, atau idemu sendiri yang setara amannya. DILARANG: api, bahan kimia/pembersih, mengutak-atik listrik/stopkontak, benda tajam, memakan/mencicipi bahan percobaan, atau apa pun yang bisa melukai dirimu, orang lain, atau merusak barang. 2) Tulis dulu PREDIKSImu (tebakan hasil sebelum mencoba). 3) Lakukan percobaannya dan ambil/unggah 1 foto kondisi percobaan atau hasilnya (boleh kamera langsung atau galeri). 4) Wajib tulis catatan dengan 4 bagian: (a) Prediksi, (b) Percobaan — apa yang kamu lakukan langkah per langkah, (c) Hasil — apa yang benar-benar terjadi, (d) Kesimpulan — apakah prediksimu benar dan apa yang kamu pelajari. Admin akan menolak submission tanpa foto, tanpa 4 bagian catatan, atau percobaan yang tidak aman."
    }'::jsonb
),
(
    'demo-crew-1',
    'Mitos atau Fakta?',
    'Uji ketajaman berpikirmu: ada 5 pernyataan sehari-hari yang sering beredar. Tentukan masing-masing MITOS atau FAKTA, dan baca penjelasannya setelah menjawab. Semua harus benar agar lolos — pikirkan baik-baik sebelum mengunci jawaban!',
    'QUIZ',
    'AUTO',
    3,
    '2026-09-18',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "questions": [
        {
          "id": "q1",
          "question": "Pernyataan: Petir tidak pernah menyambar tempat yang sama dua kali. Ini MITOS atau FAKTA?",
          "options": [
            "A. MITOS",
            "B. FAKTA"
          ],
          "correct_answer": "A",
          "explanation": "MITOS. Petir justru sering menyambar tempat yang sama berulang kali — gedung-gedung tinggi dan menara disambar petir berkali-kali setiap tahun. Tidak ada aturan alam yang melarang sambaran ulang."
        },
        {
          "id": "q2",
          "question": "Pernyataan: Air selalu mendidih pada suhu 100 derajat Celsius, di mana pun. Ini MITOS atau FAKTA?",
          "options": [
            "A. MITOS",
            "B. FAKTA"
          ],
          "correct_answer": "A",
          "explanation": "MITOS. Air mendidih pada 100 derajat hanya pada tekanan udara normal di permukaan laut. Di dataran tinggi (misalnya daerah pegunungan) tekanan udara lebih rendah sehingga air mendidih di bawah 100 derajat."
        },
        {
          "id": "q3",
          "question": "Pernyataan: Benda yang lebih berat selalu jatuh lebih cepat daripada benda yang ringan. Ini MITOS atau FAKTA?",
          "options": [
            "A. MITOS",
            "B. FAKTA"
          ],
          "correct_answer": "A",
          "explanation": "MITOS. Tanpa hambatan udara, semua benda jatuh sama cepatnya apa pun beratnya — sudah dibuktikan sejak percobaan Galileo dan eksperimen palu-bulu di Bulan. Di udara, yang berpengaruh adalah bentuk dan hambatan udara, bukan sekadar berat."
        },
        {
          "id": "q4",
          "question": "Pernyataan: Mencuci tangan dengan sabun sambil menggosok minimal 20 detik membersihkan kuman lebih efektif daripada sekadar membilas dengan air. Ini MITOS atau FAKTA?",
          "options": [
            "A. MITOS",
            "B. FAKTA"
          ],
          "correct_answer": "B",
          "explanation": "FAKTA. Sabun melarutkan minyak dan kotoran tempat kuman menempel, dan gosokan selama sekitar 20 detik memberi waktu untuk mengangkatnya. Inilah anjuran standar organisasi kesehatan dunia."
        },
        {
          "id": "q5",
          "question": "Pernyataan: Baterai HP harus selalu dikosongkan sampai 0% dulu sebelum diisi ulang agar awet. Ini MITOS atau FAKTA?",
          "options": [
            "A. MITOS",
            "B. FAKTA"
          ],
          "correct_answer": "A",
          "explanation": "MITOS. Aturan itu berlaku untuk baterai lama jenis nikel. Baterai lithium-ion modern justru lebih awet bila diisi sebelum habis total — pengosongan sampai 0% yang berulang malah mempercepat penurunan kapasitasnya."
        }
      ]
    }'::jsonb
);

-- 4. Insert 3 fresh tasks for 2026-09-19.
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
    'Hitung Mana yang Lebih Hemat?',
    'Harga termurah di label belum tentu paling hemat! Ada 4 perbandingan belanja sehari-hari: kemasan beda ukuran, diskon beda persen, dan paket beda isi. Hitung harga per satuannya (per ml, per liter, per kg, atau harga akhir setelah diskon) lalu pilih jawaban yang benar. Semua soal harus benar agar lolos.',
    'QUIZ',
    'AUTO',
    1,
    '2026-09-19',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "questions": [
        {
          "id": "q1",
          "question": "Sabun cair A: Rp18.000 untuk 250 ml. Sabun cair B: Rp25.000 untuk 400 ml. Isinya sama bagusnya. Mana yang lebih hemat per ml?",
          "options": [
            "A. Produk A (Rp72 per ml)",
            "B. Produk B (Rp62,50 per ml)",
            "C. Sama hematnya",
            "D. Tidak bisa dibandingkan"
          ],
          "correct_answer": "B",
          "explanation": "Hitung harga per ml: A = 18.000 / 250 = Rp72 per ml. B = 25.000 / 400 = Rp62,50 per ml. Produk B lebih murah per ml, jadi lebih hemat meski harga labelnya lebih mahal."
        },
        {
          "id": "q2",
          "question": "Minyak goreng A: 1 liter seharga Rp22.000. Minyak goreng B: 2 liter seharga Rp40.000. Mana yang lebih hemat per liter?",
          "options": [
            "A. Minyak A (Rp22.000 per liter)",
            "B. Minyak B (Rp20.000 per liter)",
            "C. Sama hematnya",
            "D. Tidak bisa dibandingkan"
          ],
          "correct_answer": "B",
          "explanation": "Harga per liter: A = Rp22.000 per liter. B = 40.000 / 2 = Rp20.000 per liter. Kemasan besar B lebih hemat Rp2.000 per liter."
        },
        {
          "id": "q3",
          "question": "Tas A: Rp100.000 diskon 20%. Tas B: Rp90.000 diskon 10%. Kualitas setara. Mana harga akhir yang lebih murah?",
          "options": [
            "A. Tas A (Rp80.000)",
            "B. Tas B (Rp81.000)",
            "C. Sama murahnya",
            "D. Tidak bisa dibandingkan"
          ],
          "correct_answer": "A",
          "explanation": "Harga akhir: A = 100.000 x 0,8 = Rp80.000. B = 90.000 x 0,9 = Rp81.000. Diskon persen lebih besar tidak selalu berarti harga akhir lebih murah — selalu hitung angka akhirnya."
        },
        {
          "id": "q4",
          "question": "Beras A: kemasan 5 kg seharga Rp65.000. Beras B: kemasan 10 kg seharga Rp120.000. Mana yang lebih hemat per kg?",
          "options": [
            "A. Beras A (Rp13.000 per kg)",
            "B. Beras B (Rp12.000 per kg)",
            "C. Sama hematnya",
            "D. Tidak bisa dibandingkan"
          ],
          "correct_answer": "B",
          "explanation": "Harga per kg: A = 65.000 / 5 = Rp13.000 per kg. B = 120.000 / 10 = Rp12.000 per kg. Beras B lebih hemat Rp1.000 per kg, total hemat Rp10.000 untuk 10 kg."
        }
      ]
    }'::jsonb
),
(
    'demo-crew-1',
    'Cari Jalan Keluar dari Masalah',
    'Simulasi keadaan nyata: malam hari kamu berada di tempat yang belum kamu kenal. Baterai HP tinggal 15%, kamu harus mengirim satu kabar penting ke keluarga, dan sampai di rumah dengan selamat dalam sekitar 30 menit. Hadapi 4 situasi berurutan dan pilih tindakan yang paling tepat untuk kondisimu. Setiap pilihan ada konsekuensinya — bijaklah seperti dalam kehidupan nyata!',
    'MINI_GAME',
    'AUTO',
    2,
    '2026-09-19',
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
            "title": "Situasi 1 — Baterai Tinggal 15%",
            "description": "Kamu baru sadar baterai HP tinggal 15% di tempat asing, malam hari. Satu pesan penting belum terkirim. Apa langkah pertamamu?",
            "options": [
              {
                "id": "a",
                "label": "Aktifkan mode hemat daya, kecilkan kecerahan layar, dan tutup aplikasi yang tidak perlu",
                "delta": 20,
                "hint": "Menghemat setiap persen baterai untuk hal yang benar-benar penting"
              },
              {
                "id": "b",
                "label": "Biarkan semua pengaturan seperti biasa dan langsung buka peta online dengan kecerahan penuh",
                "delta": 5,
                "hint": "Peta memang berguna, tetapi layar penuh dan data menyedot baterai dengan cepat"
              },
              {
                "id": "c",
                "label": "Buka media sosial dulu untuk mengisi waktu sambil berpikir",
                "delta": 0,
                "hint": "Hiburan menghabiskan baterai tanpa mendekatkanmu ke solusi"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 — Mengirim Kabar Penting",
            "description": "Keluarga menunggumu di rumah dan mulai khawatir. Baterai makin menipis. Bagaimana kamu mengabari mereka?",
            "options": [
              {
                "id": "a",
                "label": "Kirim satu pesan teks singkat dan padat (lokasi, kondisimu, rencanamu), lalu matikan data seluler",
                "delta": 20,
                "hint": "Satu pesan lengkap menenangkan keluarga sekaligus menghemat baterai"
              },
              {
                "id": "b",
                "label": "Telepon video panjang agar keluarga bisa melihat keadaanmu secara detail",
                "delta": 5,
                "hint": "Video call menenangkan sesaat tetapi menguras baterai paling besar"
              },
              {
                "id": "c",
                "label": "Tidak mengabari sama sekali dan berharap sampai rumah sebelum mereka panik",
                "delta": 0,
                "hint": "Tanpa kabar, keluarga khawatir dan tidak bisa membantumu jika terjadi sesuatu"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 — Butuh Arah Jalan Pulang",
            "description": "Kamu tidak hafal jalan di daerah ini. Baterai tinggal sedikit. Bagaimana cara mendapatkan arah yang benar?",
            "options": [
              {
                "id": "a",
                "label": "Lihat rute di peta selagi masih ada baterai, simpan tangkapan layarnya, lalu matikan data dan ikuti petunjuk itu",
                "delta": 20,
                "hint": "Satu kali akses peta yang direncanakan memberi panduan tanpa baterai terus menyala"
              },
              {
                "id": "b",
                "label": "Bertanya arah ke petugas, penjaga, atau penjual di sekitar dan mencatat patokan jalannya",
                "delta": 15,
                "hint": "Bertanya ke orang setempat itu cerdas — catat patokannya agar tidak lupa"
              },
              {
                "id": "c",
                "label": "Jalan mengikuti firasat tanpa panduan apa pun",
                "delta": 0,
                "hint": "Di tempat asing, firasat tanpa informasi justru berisiko tersesat lebih jauh"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 — Baterai Tinggal 8%",
            "description": "Peringatan baterai lemah muncul: tinggal 8%, perjalanan masih sekitar 15 menit. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Kunci layar, simpan HP di saku, ikuti patokan yang sudah dicatat, nyalakan HP hanya bila benar-benar perlu",
                "delta": 20,
                "hint": "Menjaga sisa baterai sebagai cadangan darurat sampai kamu tiba"
              },
              {
                "id": "b",
                "label": "Nyalakan senter dan putar musik agar perjalanan terasa lebih tenang",
                "delta": 5,
                "hint": "Nyaman sesaat, tetapi senter dan musik menghabiskan sisa baterai dengan cepat"
              },
              {
                "id": "c",
                "label": "Panik dan terus-menerus membuka tutup peta online untuk memastikan posisi",
                "delta": 0,
                "hint": "Kepanikan dan refresh berulang menghabiskan baterai tanpa menambah kepastian"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Ajarkan Satu Skill',
    'Pilih SATU keahlian nyata yang benar-benar kamu kuasai — misalnya fitur HP yang jarang orang tahu, cara memasak sederhana, cara memperbaiki hal kecil, shortcut kerja, atau teknik praktis apa pun yang aman. Rekam video sekitar 60 detik yang MENGAJARKANNYA sambil MEMPRAKTIKKANNYA langsung: tunjukkan masalahnya, tunjukkan caranya langkah per langkah, dan tunjukkan hasilnya. Wajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya sambil menjelaskan dengan suaramu.',
    'VIDEO',
    'ADMIN_REVIEW',
    3,
    '2026-09-19',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik yang mengajarkan SATU skill nyata milikmu DENGAN demonstrasi langsung (bukan sekadar bercerita):\n1. Masalah/Tujuan — sebutkan skill apa yang kamu ajarkan dan untuk apa gunanya\n2. Cara — jelaskan langkahnya satu per satu dengan bahasa sederhana\n3. Demonstrasi — PRAKTIKKAN langsung di depan kamera selagi menjelaskan\n4. Hasil — tunjukkan hasil akhirnya agar penonton yakin caranya berhasil\n\nWajah tidak wajib terlihat — boleh merekam tangan, benda, atau prosesnya. Pilih hal yang aman (tanpa api, bahan kimia, listrik, atau benda berbahaya). Jangan menyebut data pribadi sensitif. Bicaralah senatural mungkin!"
      }
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '088_tasks_replace_2026_09_18_19')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
