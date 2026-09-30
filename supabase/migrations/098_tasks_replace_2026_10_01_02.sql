-- ============================================================
-- Migration 098: Fresh tasks for 1 + 2 October 2026
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction/question/scenario text may be
-- hardcoded in frontend/backend source. Admin can edit every value below
-- through the existing Admin UI (Tasks tab + taskConfigBuilder shapes).
-- 1. No legacy rows exist on 2026-10-01 / 2026-10-02 (verified: zero
--    grep hits for 2026-10 in supabase/migrations; pre-apply script
--    re-audits live and ABORTS if any row carries submissions/ledger).
--    odyssey_claims has no task_id column, so no claim linkage is
--    possible. Fresh INSERT only; no delete/deactivate of other dates.
-- 2. Inserts 6 fresh tasks (3 per date, existing architecture and
--    config shapes only, no new framework, no new tables, no new
--    engines, no new upload system, no economy change):
--    2026-10-01 (memulai bulan baru):
--    - Step 1: MINI_GAME (Kalau Mau Mulai Lagi, Pilih yang Mana?) -
--      6 awal-bulan situations (uang menipis, jadwal berantakan,
--      perlengkapan belum siap, target bulan lalu belum selesai,
--      pekerjaan tertunda, ruang kerja berantakan), each with 3
--      actions + deterministic delta + hint. DECISION_PRIORITY +
--      POINTS, initial 0, target 100. AUTO via the existing
--      ValidateDecisionChoices server recompute (client score never
--      trusted). Reward 50c/100xp.
--    - Step 2: PHOTO_UPLOAD (Sebelum & Sesudah Awal Bulan) - member
--      does one real small change (meja/tas/lemari/area belajar/
--      tempat perlengkapan), submits exactly 1 final-result photo
--      plus a note with 3 parts (apa/kenapa/manfaat). before ->
--      action -> after narrative. max_files 1, camera_only = false
--      (gallery/file-picker allowed via existing CameraCaptureModal
--      path, same as tasks 628/632/643/650); ADMIN_REVIEW via the
--      existing approve/reject flow. Reward 40c/100xp.
--    - Step 3: TEXT_RESPONSE (Satu Hal yang Mau Kamu Benerin Bulan
--      Ini) - member picks one practical thing to fix during
--      October: kondisi sekarang / perubahan Oktober / langkah
--      pertama realistis. min 200 chars, max 1000 (server-enforced
--      by odyssey_submit_manual_task P0008 from config);
--      ADMIN_REVIEW. Reward 40c/100xp.
--    2026-10-02 (cek dan siapkan):
--    - Step 1: VIDEO (Tunjukkan Satu Trik yang Sering Kamu Pakai) -
--      member records max 60 seconds DEMONSTRATING one specific
--      practical trick they actually use often (shortcut HP,
--      merapikan kabel, packing, menyimpan file, trik kerja);
--      must show demo + object/device where relevant, must be
--      imitable by others. Talking-head without demo does not meet
--      intent (admin-reviewable). recording.enabled,
--      max_duration_seconds = 60, camera user; ADMIN_REVIEW.
--      Reward 40c/100xp.
--    - Step 2: MINI_GAME (Sebelum Dipakai, Cek Dulu) - 6 pre-action
--      readiness/safety situations (sebelum pakai alat, kirim
--      dokumen, tinggalkan rumah, simpan makanan, pakai perangkat,
--      selesaikan pekerjaan), each with 3 priority actions +
--      deterministic delta + hint. DECISION_PRIORITY + POINTS,
--      initial 0, target 100. Explicitly NOT DECISION_ANOMALY (no
--      spot-the-odd pattern of task 635). AUTO via existing server
--      recompute. Reward 50c/100xp.
--    - Step 3: DOCUMENT_UPLOAD (Bikin Checklist yang Bisa Langsung
--      Dipakai) - member uploads a practical CHECKLIST (max 10
--      action/check items, directly usable by others), not a long
--      tutorial/guide prose. allowed_extensions + accepted_extensions
--      (dual-key intentional: DocUploadModal reads accepted_
--      extensions while admin builder writes allowed_extensions),
--      max_file_size_mb = 4; ADMIN_REVIEW. Reward 40c/100xp.
-- 3. Idempotency: the pre-clean DELETEs below match ONLY the 6 exact
--    new titles on their exact dates (narrow exact-title guard for dev
--    re-runs, same spirit as migrations 083/084/085/086/087/088/090/
--    091/093). No broad date/title matching.
-- 4. Does NOT touch tasks 608-665 (610 camera_only stays true;
--    628/632/643/650 stay false), payout/economy config,
--    announcements, RLS, RPCs, triggers, claims, shop, or any other
--    date. Economy (cap 3320 / target 0 / rate 100) unchanged.
-- ============================================================

-- 1. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact dates). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-01'
  AND title IN (
    'Kalau Mau Mulai Lagi, Pilih yang Mana?',
    'Sebelum & Sesudah Awal Bulan',
    'Satu Hal yang Mau Kamu Benerin Bulan Ini'
  );

DELETE FROM odyssey_tasks
WHERE active_date = '2026-10-02'
  AND title IN (
    'Tunjukkan Satu Trik yang Sering Kamu Pakai',
    'Sebelum Dipakai, Cek Dulu',
    'Bikin Checklist yang Bisa Langsung Dipakai'
  );

-- 2. Insert 3 fresh tasks for 2026-10-01.
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
    'Kalau Mau Mulai Lagi, Pilih yang Mana?',
    'Bulan baru, awal baru! Oktober baru saja dimulai dan ada 6 situasi khas awal bulan yang harus kamu hadapi: uang mulai menipis, jadwal berantakan, perlengkapan belum siap, target bulan lalu belum selesai, pekerjaan tertunda, dan ruang kerja yang berantakan. Pada setiap situasi, pilih tindakan yang paling masuk akal untuk memulai bulan baru dengan benar. Setiap pilihan ada konsekuensinya!',
    'MINI_GAME',
    'AUTO',
    1,
    '2026-10-01',
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
            "title": "Situasi 1 - Uang Mulai Menipis",
            "description": "Baru tanggal 1, kamu sadar sisa uang tinggal sedikit sementara kebutuhan sebulan masih panjang. Apa langkah pertamamu?",
            "options": [
              {
                "id": "a",
                "label": "Catat sisa uang dan susun anggaran prioritas: kebutuhan wajib dulu, keinginan ditunda",
                "delta": 20,
                "hint": "Awal bulan yang baik dimulai dari tahu persis posisi keuanganmu sebelum uang habis tanpa jejak"
              },
              {
                "id": "b",
                "label": "Langsung berhemat tanpa mencatat, kurangi saja semua pengeluaran secara merata",
                "delta": 10,
                "hint": "Berhemat itu bagus, tetapi tanpa catatan kamu tidak tahu pos mana yang sebenarnya bocor"
              },
              {
                "id": "c",
                "label": "Belanja dulu mumpung awal bulan, urusan kehabisan dipikir nanti saja",
                "delta": 0,
                "hint": "Mengulang pola bulan lalu yang membuat uang menipis justru menjamin masalah yang sama terulang"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 - Jadwal Berantakan",
            "description": "Jadwal minggu pertamamu bertabrakan di mana-mana: janji, pekerjaan, dan urusan keluarga saling tumpuk. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Tulis semua jadwal di satu tempat, tentukan prioritas, lalu atur ulang yang bisa digeser",
                "delta": 20,
                "hint": "Jadwal yang terlihat semua dalam satu daftar bisa diatur; yang hanya di kepala pasti ada yang terlewat"
              },
              {
                "id": "b",
                "label": "Kerjakan yang paling mendesak saja hari ini, sisanya dipikir besok",
                "delta": 5,
                "hint": "Reaktif tanpa peta jadwal membuat tabrakan berikutnya tidak terhindarkan"
              },
              {
                "id": "c",
                "label": "Batalkan semuanya sekaligus agar tenang, nanti disusun ulang kalau sempat",
                "delta": 0,
                "hint": "Membatalkan semua tanpa memilah merusak komitmen yang sebenarnya masih bisa dipenuhi"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 - Perlengkapan Belum Siap",
            "description": "Barang dan perlengkapan yang kamu butuhkan untuk beraktivitas bulan ini belum lengkap dan berserakan. Apa langkahmu?",
            "options": [
              {
                "id": "a",
                "label": "Daftar dulu apa yang benar-benar dibutuhkan, kumpulkan yang sudah ada, baru lengkapi yang kurang",
                "delta": 20,
                "hint": "Daftar kebutuhan mencegah beli ganda dan memastikan yang paling penting siap duluan"
              },
              {
                "id": "b",
                "label": "Beli semua yang terasa kurang sekaligus agar langsung lengkap",
                "delta": 5,
                "hint": "Belanja tanpa daftar di awal bulan berisiko menguras anggaran untuk barang yang ternyata sudah ada"
              },
              {
                "id": "c",
                "label": "Jalani saja apa adanya tanpa menyiapkan apa-apa",
                "delta": 0,
                "hint": "Bulan baru tanpa persiapan perlengkapan hanya memindahkan kerepotan ke setiap harinya"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 - Target Bulan Lalu Belum Selesai",
            "description": "Ada target bulan September yang belum selesai dan terbawa ke Oktober. Bagaimana kamu menyikapinya?",
            "options": [
              {
                "id": "a",
                "label": "Pilih satu sisa target yang paling penting, beri tenggat baru yang realistis, dan mulai dari langkah terkecil",
                "delta": 20,
                "hint": "Satu sisa target yang diselesaikan tuntas lebih berharga daripada semua sisa target yang terus dibawa-bawa"
              },
              {
                "id": "b",
                "label": "Lanjutkan semua sisa target sekaligus bersamaan dengan target baru Oktober",
                "delta": 5,
                "hint": "Menumpuk semua sisa target dengan target baru tanpa memilah membuat tidak ada yang selesai"
              },
              {
                "id": "c",
                "label": "Lupakan semua sisa target bulan lalu dan anggap Oktober mulai dari nol",
                "delta": 0,
                "hint": "Membuang sisa target tanpa evaluasi menjamin pelajaran yang sama hilang begitu saja"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 - Pekerjaan Tertunda Menumpuk",
            "description": "Beberapa pekerjaan tertunda dari minggu-minggu sebelumnya menumpuk dan semuanya terasa mendesak. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Urutkan pekerjaan tertunda berdasarkan dampak dan tenggat, selesaikan satu per satu mulai dari yang paling penting",
                "delta": 20,
                "hint": "Tumpukan yang diurutkan berubah menjadi antrean yang bisa diselesaikan; yang tidak diurutkan tetap menjadi beban"
              },
              {
                "id": "b",
                "label": "Kerjakan yang paling mudah dulu semuanya agar tumpukan cepat berkurang",
                "delta": 10,
                "hint": "Yang mudah selesai duluan memang melegakan, tetapi yang penting dan mendesak bisa kedaluwarsa menunggu"
              },
              {
                "id": "c",
                "label": "Tunggu sampai benar-benar mepet baru kerjakan semuanya sekaligus",
                "delta": 0,
                "hint": "Menunda yang tertunda hanya melipatgandakan tekanan tanpa menambah waktu"
              }
            ]
          },
          {
            "id": "ev_6",
            "title": "Situasi 6 - Ruang Kerja Berantakan",
            "description": "Meja dan area kerjamu berantakan setelah sebulan penuh dipakai. Awal bulan rasanya ingin yang segar. Apa langkahmu?",
            "options": [
              {
                "id": "a",
                "label": "Luangkan waktu khusus merapikan: buang yang tidak perlu, kembalikan tiap barang ke tempatnya",
                "delta": 20,
                "hint": "Ruang kerja yang rapi di awal bulan menghemat waktu mencari barang selama sebulan penuh"
              },
              {
                "id": "b",
                "label": "Geser barang-barang ke satu sudut agar meja terlihat rapi sekilas",
                "delta": 5,
                "hint": "Rapi sekilas tanpa memilah hanya memindahkan kekacauan ke sudut lain yang akan meledak lagi"
              },
              {
                "id": "c",
                "label": "Biarkan saja, yang penting langsung mulai bekerja",
                "delta": 0,
                "hint": "Memulai bulan baru di atas kekacauan lama membuat setiap pekerjaan kecil terasa lebih berat"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Sebelum & Sesudah Awal Bulan',
    'Buktikan awal bulanmu dimulai dengan perubahan nyata! Pilih satu tempat di sekitarmu (meja, tas kerja, lemari, area belajar, atau tempat menyimpan perlengkapan), lakukan satu perubahan kecil yang benar-benar nyata, lalu foto HASIL AKHIRNYA dan ceritakan perubahanmu.',
    'PHOTO_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-10-01',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "max_files": 1,
      "camera_only": false,
      "instruction": "Lakukan alur before -> action -> after: 1) BEFORE - amati dan pilih SATU tempat yang perlu diubah (contoh: meja, tas kerja, lemari, area belajar, atau tempat menyimpan perlengkapan). 2) ACTION - lakukan satu perubahan kecil yang benar-benar nyata di tempat itu (merapikan, menyusun ulang, membersihkan, atau menata ulang). 3) AFTER - ambil atau unggah tepat 1 foto HASIL AKHIR perubahanmu (boleh kamera langsung atau galeri, maksimal 1 file). 4) Wajib tulis catatan dengan 3 bagian: (a) Apa - jelaskan apa yang kamu ubah dan bagaimana kondisi sebelumnya, (b) Kenapa - jelaskan kenapa tempat itu perlu diubah, (c) Manfaat - jelaskan apa manfaat perubahan itu untuk aktivitasmu. Admin akan menolak submission tanpa foto hasil akhir, tanpa 3 bagian catatan (apa/kenapa/manfaat), atau perubahan yang tidak terlihat nyata."
    }'::jsonb
),
(
    'demo-crew-1',
    'Satu Hal yang Mau Kamu Benerin Bulan Ini',
    'Oktober punya 31 hari - cukup untuk memperbaiki satu hal dengan sungguh-sungguh! Pilih SATU kebiasaan atau hal praktis yang ingin kamu perbaiki selama bulan Oktober, lalu tulis kondisimu sekarang, perubahan yang kamu inginkan, dan langkah pertama yang realistis.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-10-01',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis jawabanmu dengan format 3 bagian:\n1) Kondisi sekarang - jelaskan kondisi saat ini dari SATU hal yang ingin kamu perbaiki (contoh: kebiasaan bangun siang, meja yang selalu berantakan, keuangan yang bocor, atau cara mengatur waktu). Ceritakan sejujurnya seperti apa kondisinya sekarang dan apa dampaknya buatmu:\n2) Perubahan yang ingin dilakukan selama Oktober - jelaskan perubahan seperti apa yang kamu targetkan sampai akhir Oktober. Buat spesifik dan terukur, bukan sekadar ingin lebih baik:\n3) Langkah pertama yang realistis - tulis SATU langkah pertama yang kecil, konkret, dan bisa kamu mulai dalam 24 jam ke depan:\n\nFokus pada satu hal saja, jangan membahas banyak hal sekaligus. Tulis dari pengalamanmu sendiri. Admin akan menolak jawaban di bawah 200 karakter, tanpa 3 bagian di atas, membahas terlalu banyak hal sekaligus, atau langkah pertama yang tidak realistis.",
      "minimum_characters": 200,
      "maximum_characters": 1000
    }'::jsonb
)
;

-- 3. Insert 3 fresh tasks for 2026-10-02.
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
    'Tunjukkan Satu Trik yang Sering Kamu Pakai',
    'Punya trik andalan yang hampir setiap hari kamu pakai? Saatnya menunjukkannya! Rekam video maksimal 60 detik yang MENDEMONSTRASIKAN satu trik praktis yang benar-benar sering kamu gunakan - misalnya shortcut HP, cara merapikan kabel, trik packing, cara menyimpan file, atau trik sederhana saat bekerja. Orang lain harus bisa menirunya setelah menonton videomu!',
    'VIDEO',
    'ADMIN_REVIEW',
    1,
    '2026-10-02',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "recording": {
        "enabled": true,
        "max_duration_seconds": 60,
        "camera_facing": "user",
        "instruction": "Rekam video maksimal 60 detik dengan struktur:\n1. Sebutkan SATU trik spesifik yang benar-benar sering kamu pakai (contoh: shortcut HP, cara merapikan kabel, trik packing, cara menyimpan file, atau trik sederhana saat bekerja) - fokus pada satu trik saja, jangan banyak-banyak\n2. DEMONSTRASIKAN trik itu secara langsung di depan kamera (wajib ada demonstrasi nyata, bukan hanya bicara atau bercerita) - tunjukkan objek atau perangkat yang kamu gunakan jika relevan (HP, kabel, tas, aplikasi, peralatan kerja)\n3. Jelaskan kenapa trik itu berguna dan dalam situasi apa orang lain bisa menirunya\n\nCatatan: submission yang hanya berbicara atau bercerita tanpa demonstrasi TIDAK memenuhi task ini. Admin akan menolak video tanpa demonstrasi nyata, tanpa satu trik yang spesifik, tanpa objek atau perangkat yang ditunjukkan padahal relevan, durasi melebihi 60 detik, atau trik yang tidak bisa ditiru orang lain."
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Sebelum Dipakai, Cek Dulu',
    'Orang yang teliti selalu mengecek dulu sebelum bertindak! Hadapi 6 situasi sehari-hari di mana kamu harus memastikan sesuatu aman dan siap dipakai: sebelum menggunakan alat, mengirim dokumen, meninggalkan rumah, menyimpan makanan, menggunakan perangkat, dan menyelesaikan pekerjaan. Pada setiap situasi, tentukan tindakan pengecekan yang paling tepat. Ini soal urutan dan prioritas tindakan - bukan mencari yang janggal!',
    'MINI_GAME',
    'AUTO',
    2,
    '2026-10-02',
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
            "title": "Situasi 1 - Sebelum Menggunakan Alat",
            "description": "Kamu akan memakai sebuah alat (misalnya setrika, blender, atau perkakas) yang sudah lama tidak dipakai. Apa urutan tindakan yang paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Periksa dulu kondisi alat dan kabelnya, pastikan aman, baru gunakan sesuai fungsinya",
                "delta": 20,
                "hint": "Pengecekan kondisi sebelum memakai alat mencegah kecelakaan yang sebenarnya bisa dihindari dalam satu menit"
              },
              {
                "id": "b",
                "label": "Langsung pakai saja, kalau rusak baru berhenti",
                "delta": 0,
                "hint": "Menunggu sampai rusak atau celaka untuk berhenti justru adalah cara paling mahal untuk mengecek"
              },
              {
                "id": "c",
                "label": "Minta orang lain mencoba dulu, kalau aman baru kamu pakai",
                "delta": 5,
                "hint": "Memindahkan risiko ke orang lain bukan pengecekan - tanggung jawab memastikan keamanan ada padamu"
              }
            ]
          },
          {
            "id": "ev_2",
            "title": "Situasi 2 - Sebelum Mengirim Dokumen",
            "description": "Kamu akan mengirim dokumen penting (lamaran, tugas, atau berkas). Tenggatnya mepet. Apa yang kamu lakukan?",
            "options": [
              {
                "id": "a",
                "label": "Cek ulang isi, nama file, dan alamat tujuan, baru kirim walau butuh beberapa menit ekstra",
                "delta": 20,
                "hint": "Dokumen penting yang salah kirim tidak bisa ditarik kembali - beberapa menit pengecekan jauh lebih murah daripada mengulang"
              },
              {
                "id": "b",
                "label": "Kirim secepatnya agar tidak telat, cek belakangan kalau sempat",
                "delta": 0,
                "hint": "Kecepatan tanpa pengecekan pada dokumen penting berisiko mengirim kesalahan yang permanen"
              },
              {
                "id": "c",
                "label": "Kirim dulu versi seadanya, nanti kirim revisi kalau ada yang salah",
                "delta": 5,
                "hint": "Mengirim versi seadanya dengan rencana revisi menunjukkan kamu tidak memeriksa sebelum bertindak"
              }
            ]
          },
          {
            "id": "ev_3",
            "title": "Situasi 3 - Sebelum Meninggalkan Rumah",
            "description": "Kamu akan pergi dan rumah kosong seharian. Apa urutan pengecekan yang paling tepat sebelum berangkat?",
            "options": [
              {
                "id": "a",
                "label": "Cek kompor dan listrik sudah mati atau aman, kunci pintu dan jendela, bawa kunci serta barang penting, baru berangkat",
                "delta": 20,
                "hint": "Urutan cek sebelum meninggalkan rumah menyelamatkan dari bahaya dan dari keharusan pulang di tengah jalan"
              },
              {
                "id": "b",
                "label": "Langsung berangkat agar tidak telat, rumah aman-aman saja biasanya",
                "delta": 0,
                "hint": "Mengandalkan kebiasaan aman tanpa mengecek adalah cara paling umum musibah rumah kosong terjadi"
              },
              {
                "id": "c",
                "label": "Kunci pintu saja, urusan kompor dan listrik dipikir nanti lewat telepon ke keluarga",
                "delta": 5,
                "hint": "Mengunci saja tanpa mengecek sumber bahaya meninggalkan risiko terbesar tanpa pengawasan"
              }
            ]
          },
          {
            "id": "ev_4",
            "title": "Situasi 4 - Sebelum Menyimpan Makanan",
            "description": "Ada makanan sisa yang ingin kamu simpan untuk besok. Apa tindakan yang paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Pastikan makanan masih layak, simpan dalam wadah bersih tertutup, dan beri tanda kapan menyimpannya",
                "delta": 20,
                "hint": "Makanan yang dicek kelayakannya dan disimpan tertutup dengan penanda waktu aman dikonsumsi besok"
              },
              {
                "id": "b",
                "label": "Langsung masukkan ke kulkas apa adanya dalam piring terbuka",
                "delta": 5,
                "hint": "Menyimpan tanpa wadah tertutup dan tanpa mengecek kelayakan mengundang kontaminasi dan kebusukan"
              },
              {
                "id": "c",
                "label": "Biarkan saja di meja tertutup serbet, besok masih bisa dimakan",
                "delta": 0,
                "hint": "Makanan yang dibiarkan di suhu ruang semalaman berisiko basi dan berbahaya bagi kesehatan"
              }
            ]
          },
          {
            "id": "ev_5",
            "title": "Situasi 5 - Sebelum Menggunakan Perangkat",
            "description": "Kamu akan memakai HP atau laptop untuk pekerjaan penting yang lama. Apa persiapan yang paling tepat?",
            "options": [
              {
                "id": "a",
                "label": "Cek daya baterai atau colokan, pastikan data penting tersimpan, siapkan koneksi, baru mulai bekerja",
                "delta": 20,
                "hint": "Pekerjaan penting yang disiapkan daya, data, dan koneksinya tidak akan terhenti di tengah jalan"
              },
              {
                "id": "b",
                "label": "Langsung mulai bekerja, urusan baterai dan data dipikir kalau ada masalah",
                "delta": 0,
                "hint": "Bekerja tanpa memastikan daya dan simpanan data mempertaruhkan seluruh hasil pekerjaanmu"
              },
              {
                "id": "c",
                "label": "Pinjam perangkat orang lain saja agar tidak repot menyiapkan perangkat sendiri",
                "delta": 5,
                "hint": "Bergantung pada perangkat pinjaman tanpa persiapan sendiri menambah risiko yang tidak perlu"
              }
            ]
          },
          {
            "id": "ev_6",
            "title": "Situasi 6 - Sebelum Menyelesaikan Pekerjaan",
            "description": "Pekerjaanmu sudah hampir selesai dan tinggal dinyatakan beres. Apa langkah terakhirmu?",
            "options": [
              {
                "id": "a",
                "label": "Cek ulang hasilnya satu per satu, perbaiki yang kurang, rapikan, baru nyatakan selesai",
                "delta": 20,
                "hint": "Pengecekan akhir sebelum menyatakan selesai membedakan pekerjaan yang rapi dari yang asal beres"
              },
              {
                "id": "b",
                "label": "Langsung nyatakan selesai agar cepat, toh sudah dikerjakan",
                "delta": 0,
                "hint": "Menyatakan selesai tanpa cek ulang menyerahkan penilaian kualitas pada keberuntungan"
              },
              {
                "id": "c",
                "label": "Minta orang lain yang mengecek semuanya, kamu langsung lanjut ke hal lain",
                "delta": 5,
                "hint": "Melempar seluruh pengecekan akhir ke orang lain berarti kamu tidak memastikan sendiri hasil kerjamu"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Bikin Checklist yang Bisa Langsung Dipakai',
    'Tutorial panjang jarang dibaca sampai habis - checklist praktis selalu dipakai! Buat SATU dokumen checklist (maksimal 10 item) yang bisa langsung digunakan orang lain tanpa membaca tutorial panjang. Pilih topik praktis: checklist sebelum berangkat kerja, sebelum packing, sebelum mengirim dokumen, sebelum membersihkan sesuatu, atau sebelum melakukan pekerjaan tertentu.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    3,
    '2026-10-02',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "accepted_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4,
      "instruction": "Buat SATU dokumen checklist dengan aturan: 1) Maksimal 10 checklist items - tidak boleh lebih dari 10. 2) Setiap item harus berupa tindakan atau cek yang konkret (contoh format: [ ] Cek ..., [ ] Pastikan ..., [ ] Bawa ...) - bukan paragraf penjelasan panjang. 3) Dokumen harus bisa LANGSUNG DIGUNAKAN orang lain sebagai quick checklist tanpa perlu membaca tutorial (orang lain cukup mencentang satu per satu). 4) Pilih SATU topik praktis (contoh: checklist sebelum berangkat kerja, sebelum packing, sebelum mengirim dokumen, sebelum membersihkan sesuatu, atau sebelum melakukan pekerjaan tertentu). 5) Unggah dokumen (pdf/docx/txt/jpg/jpeg/png, maksimal 4MB) dan tulis catatan singkat: topik checklistmu dan untuk siapa checklist itu berguna. Admin akan menolak dokumen berupa tutorial atau guide panjang tanpa format checklist, dokumen dengan lebih dari 10 item, item yang bukan tindakan atau cek konkret, atau dokumen yang tidak bisa langsung dipakai."
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '098_tasks_replace_2026_10_01_02')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
