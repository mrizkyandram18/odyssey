-- ============================================================
-- Migration 085: Replace tasks for 15 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/question/scenario/instruction text may be hardcoded
-- in frontend/backend source. Admin can edit every value below through
-- the existing Admin UI (Tasks tab).
-- 1. DEACTIVATES (never deletes) exact tasks 540/541 on 2026-09-15.
--    Production evidence (2026-09-15 audit at apply time):
--    - 540 (Bandingkan 2 Lowongan Sejenis, TEXT_RESPONSE) carries
--      1 PENDING submission (id 128) from a real user. DELETE would
--      orphan history, so deactivation is the only safe semantics.
--    - 541 (Latihan Kasir, QUIZ) carries 1 APPROVED submission
--      (id 129) + 1 immutable odyssey_coin_transactions row (id 98).
--      DELETE would violate FK history and the immutable ledger
--      trigger, so deactivation is the only safe semantics.
--    Deactivation is naturally idempotent.
-- 2. Inserts 3 fresh canonical tasks for 2026-09-15 (existing
--    architecture and config shapes only, no new framework):
--    - Step 1: MINI_GAME (Susun Sampai Beres) — 6 sequential
--      morning-routine prioritization decisions, POINTS currency,
--      target_score 100. SUPPORT NOTE: no dedicated drag-and-drop
--      ordering capability exists in the app (no sortable UI, no
--      array-order validator). Closest existing server-validated
--      capability is the MINI_GAME decision scenario
--      (config.scenario.events + ValidateDecisionChoices +
--      DecisionGameModal), so sequencing/planning/prioritization is
--      taught as ordered priority decisions over one timeline.
--    - Step 2: DOCUMENT_UPLOAD (Bikin Panduan yang Bisa Dipakai
--      Orang Lain) — member creates a real 1-page usable guide on
--      their phone and uploads it (mobile-friendly extensions,
--      max 4 MB to respect the Vercel payload limit enforced by
--      DocUploadModal). ADMIN_REVIEW via existing approve/reject flow.
--    - Step 3: MINI_GAME (Mana yang Janggal?) — 6 anomaly-detection
--      scenarios (scam message, contradictory info, illogical work
--      instruction, suspicious price, risky digital situation, simple
--      logic error), POINTS currency, target_score 100. Same existing
--      decision-scenario shape; scoring deterministic via server
--      recompute (score forced to 100 after ValidateDecisionChoices).
-- 3. Idempotency: the pre-clean DELETE below matches ONLY the 3 exact
--    new titles on 2026-09-15 (narrow exact-title guard for dev re-runs,
--    same spirit as migrations 052/083/084). No broad date/title matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), 2026-09-13
--    (630/631/632), tasks 608-620, payout/economy config,
--    announcements, RLS, RPCs, triggers, or any other date.
-- ============================================================

-- 1. Deactivate exact legacy tasks for 2026-09-15 (history-preserving).
UPDATE odyssey_tasks
SET is_active = FALSE
WHERE id IN (540, 541)
  AND active_date = '2026-09-15';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact date). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-15'
  AND title IN (
    'Susun Sampai Beres',
    'Bikin Panduan yang Bisa Dipakai Orang Lain',
    'Mana yang Janggal?'
  );

-- 3. Insert 3 fresh canonical tasks for 2026-09-15.
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
    'Susun Sampai Beres',
    'Pagi ini waktumu mepet: HP tinggal 10 persen, baju belum siap, hujan mulai turun, kendaraan belum dicek, dan jam berangkat sudah dekat. Lewati 6 titik keputusan berurutan dan pilih langkah yang paling masuk akal di setiap titik. Melatih sequencing, planning, dan prioritization dalam satu timeline pagi yang realistis.',
    'MINI_GAME',
    'AUTO',
    1,
    '2026-09-15',
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
            "id": "urut_1",
            "title": "Titik 1 — 20 Menit Sebelum Berangkat, HP 10 Persen",
            "description": "Kamu baru bangun. Baterai HP tinggal 10 persen, dan banyak hal harus disiapkan. Apa langkah pembuka yang paling masuk akal?",
            "options": [
              {
                "id": "a",
                "label": "Colok HP ke charger dulu, lalu kerjakan persiapan lain selagi HP mengisi daya",
                "delta": 20,
                "hint": "Tugas pasif (mengisi daya) dijalankan paralel dengan tugas aktif — prinsip sequencing yang benar"
              },
              {
                "id": "b",
                "label": "Main HP dulu sebentar untuk cek pesan sambil menunggu mood siap",
                "delta": 5,
                "hint": "Menghabiskan baterai yang sudah kritis untuk hal yang tidak mendesak"
              },
              {
                "id": "c",
                "label": "Abaikan baterai sepenuhnya dan berharap HP bertahan sampai tempat tujuan",
                "delta": 0,
                "hint": "Mengabaikan risiko yang sudah terlihat jelas di depan mata"
              }
            ]
          },
          {
            "id": "urut_2",
            "title": "Titik 2 — Baju Belum Siap",
            "description": "HP sudah mengisi daya. Lemari terbuka dan baju untuk hari ini belum dipilih. Di luar, langit mulai mendung. Apa urutan yang tepat?",
            "options": [
              {
                "id": "a",
                "label": "Siapkan baju lengkap dari atas sampai bawah sekarang, termasuk antisipasi hujan (jaket atau payung)",
                "delta": 20,
                "hint": "Menyiapkan kebutuhan diri sebelum keluar rumah, sekalian membaca sinyal cuaca"
              },
              {
                "id": "b",
                "label": "Cek kendaraan dulu ke garasi dengan masih memakai baju tidur",
                "delta": 10,
                "hint": "Bolak-balik tidak efisien: berpakaian dulu baru keluar memeriksa kendaraan"
              },
              {
                "id": "c",
                "label": "Pakai baju apa saja seadanya tanpa memikirkan hujan yang mulai turun",
                "delta": 0,
                "hint": "Mengabaikan informasi cuaca yang sudah tersedia saat menyusun persiapan"
              }
            ]
          },
          {
            "id": "urut_3",
            "title": "Titik 3 — Hujan Mulai Turun",
            "description": "Kamu sudah berpakaian rapi. Hujan rintik mulai turun di luar. Kendaraan belum dicek sama sekali. Apa prioritas berikutnya?",
            "options": [
              {
                "id": "a",
                "label": "Amankan barang bawaan terhadap hujan (tutup tas, siapkan jas hujan atau payung), baru cek kendaraan",
                "delta": 20,
                "hint": "Melindungi barang bawaan dulu sebelum beraktivitas di luar saat hujan"
              },
              {
                "id": "b",
                "label": "Langsung terobos hujan untuk cek kendaraan tanpa melindungi barang bawaan",
                "delta": 5,
                "hint": "Tindakan cepat tetapi membuat dokumen dan barang elektronik berisiko basah"
              },
              {
                "id": "c",
                "label": "Menunggu hujan reda sambil rebahan tanpa menyiapkan apa pun",
                "delta": 0,
                "hint": "Waktu berangkat makin dekat sementara tidak ada persiapan yang maju"
              }
            ]
          },
          {
            "id": "urut_4",
            "title": "Titik 4 — Kendaraan Belum Dicek",
            "description": "Tinggal sekitar 10 menit. Kendaraan harus dipastikan siap: bensin, ban, lampu, dan rem. Mana pemeriksaan yang paling masuk akal?",
            "options": [
              {
                "id": "a",
                "label": "Cek hal yang menggagalkan perjalanan dulu (bensin cukup, ban tidak kempes, rem dan lampu berfungsi), lalu rapikan sisanya",
                "delta": 20,
                "hint": "Prinsip critical-path: pastikan dulu hal yang bisa membatalkan perjalanan"
              },
              {
                "id": "b",
                "label": "Cuci dan poles kendaraan sampai mengilap agar terlihat bagus di jalan",
                "delta": 5,
                "hint": "Penampilan diutamakan di atas fungsi saat waktu sudah kritis"
              },
              {
                "id": "c",
                "label": "Langsung berangkat tanpa pengecekan apa pun karena merasa kemarin baik-baik saja",
                "delta": 0,
                "hint": "Asumsi kemarin baik berarti hari ini aman adalah tebakan, bukan perencanaan"
              }
            ]
          },
          {
            "id": "urut_5",
            "title": "Titik 5 — Barang Bawaan Tercecer",
            "description": "Kendaraan siap. Dompet, kunci cadangan, dan bekal masih tercecer di tiga tempat berbeda. Waktu tinggal 5 menit. Bagaimana merapikannya?",
            "options": [
              {
                "id": "a",
                "label": "Kumpulkan semua barang penting ke satu titik dekat pintu, ceklis satu per satu, baru keluar",
                "delta": 20,
                "hint": "Single staging point plus ceklis mencegah barang tertinggal saat terburu-buru"
              },
              {
                "id": "b",
                "label": "Bawa yang terlihat saja dan berharap sisanya tidak dibutuhkan hari ini",
                "delta": 5,
                "hint": "Berharap-harap tanpa verifikasi adalah sumber barang tertinggal"
              },
              {
                "id": "c",
                "label": "Bolak-balik mengambil barang satu per satu tanpa daftar sampai waktu habis",
                "delta": 0,
                "hint": "Gerak tanpa sistem menghabiskan sisa waktu yang sudah sangat sempit"
              }
            ]
          },
          {
            "id": "urut_6",
            "title": "Titik 6 — Detik-Detik Berangkat",
            "description": "Semua siap: HP cukup daya, baju rapi, barang aman dari hujan, kendaraan oke, bawaan lengkap. Satu menit terakhir sebelum kunci pintu. Apa penutup yang tepat?",
            "options": [
              {
                "id": "a",
                "label": "Verifikasi akhir 30 detik (kompor dan lampu mati, pintu terkunci, HP dan dompet di saku), lalu berangkat tenang",
                "delta": 20,
                "hint": "Closing checklist singkat mencegah bencana kecil yang merusak seluruh persiapan"
              },
              {
                "id": "b",
                "label": "Langsung pergi secepatnya tanpa verifikasi karena semua terasa sudah siap",
                "delta": 10,
                "hint": "Perasaan siap tanpa verifikasi sering melewatkan satu hal kecil yang penting"
              },
              {
                "id": "c",
                "label": "Mengulang seluruh persiapan dari awal karena tiba-tiba tidak yakin",
                "delta": 0,
                "hint": "Mengulang tanpa alasan membuang hasil urutan yang sudah benar"
              }
            ]
          }
        ]
      }
    }'::jsonb
),
(
    'demo-crew-1',
    'Bikin Panduan yang Bisa Dipakai Orang Lain',
    'Buat satu panduan praktis 1 halaman yang benar-benar bisa dipakai orang lain: misalnya cara menyiapkan barang sebelum kerja, cara mengatur uang belanja mingguan, checklist sebelum berangkat, cara merapikan file di HP, atau cara melakukan aktivitas sederhana dengan lebih teratur. Boleh dibuat dan difoto langsung dari HP. Unggah hasilnya sebagai dokumen — bukan sekadar essay atau curhat.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    2,
    '2026-09-15',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4,
      "instruction": "Buat 1 panduan praktis yang bisa langsung dipakai orang lain. Pilih SATU topik: 1) menyiapkan barang sebelum kerja, 2) mengatur uang belanja mingguan, 3) checklist sebelum berangkat, 4) merapikan file di HP, atau 5) aktivitas sederhana lain yang dibuat lebih teratur. Isi panduan WAJIB memuat: judul yang jelas, daftar bahan atau syarat (jika ada), langkah bernomor yang berurutan, dan 1 tips agar berhasil. Boleh diketik di HP lalu disimpan sebagai PDF/DOCX/TXT, atau ditulis tangan rapi lalu difoto jelas (JPG/PNG). Maksimal 4 MB. Admin menolak file yang tidak bisa dibuka, bukan panduan (misalnya curhat atau essay tanpa langkah), atau tanpa langkah yang bisa diikuti."
    }'::jsonb
),
(
    'demo-crew-1',
    'Mana yang Janggal?',
    'Ada yang tidak beres di 6 situasi ini: pesan berhadiah, harga yang terlalu manis, instruksi kerja yang bertentangan, email bank yang aneh, kabar yang simpang siur, dan kode QR tempelan. Temukan bagian yang janggal di setiap situasi. Ini permainan penalaran, bukan kuis hafalan — baca dengan teliti sebelum memilih.',
    'MINI_GAME',
    'AUTO',
    3,
    '2026-09-15',
    50,
    100,
    'ALL',
    TRUE,
    '{
      "game": "DECISION_ANOMALY",
      "target_score": 100,
      "scenario": {
        "currency": "POINTS",
        "initial_balance": 0,
        "events": [
          {
            "id": "janggal_1",
            "title": "Kasus 1 — Pesan Menang Undian",
            "description": "Kamu dapat pesan: SELAMAT! Nomor kamu menang undian Rp10 juta dari operator. Untuk mencairkan, balas dengan KODE OTP yang baru saja dikirim ke HP kamu, plus foto KTP. Mana bagian yang paling janggal?",
            "options": [
              {
                "id": "a",
                "label": "Permintaan kode OTP dan foto KTP — itu data rahasia yang tidak pernah diminta oleh penyelenggara resmi",
                "delta": 20,
                "hint": "Ciri scam klasik: hadiah dipakai sebagai umpan untuk mencuri kode verifikasi dan identitas"
              },
              {
                "id": "b",
                "label": "Nominal hadiahnya yang terlalu kecil sehingga tidak menarik",
                "delta": 5,
                "hint": "Besar-kecil hadiah bukan inti masalah; mekanisme pencairannya yang berbahaya"
              },
              {
                "id": "c",
                "label": "Pesan dikirim pada jam kerja sehingga terlihat resmi",
                "delta": 0,
                "hint": "Jam pengiriman tidak menentukan keaslian sebuah pesan"
              }
            ]
          },
          {
            "id": "janggal_2",
            "title": "Kasus 2 — Harga yang Terlalu Manis",
            "description": "Sebuah toko online menjual HP keluaran terbaru seharga 10 persen dari harga pasaran. Syaratnya: transfer lunas ke rekening pribadi sebelum barang dikirim, dan tidak bisa COD atau marketplace. Mana yang paling mencurigakan?",
            "options": [
              {
                "id": "a",
                "label": "Kombinasi diskon tidak masuk akal dengan pembayaran transfer ke rekening pribadi tanpa perlindungan pembeli",
                "delta": 20,
                "hint": "Dua anomali sekaligus: harga mustahil plus jalur pembayaran yang menghilangkan jejak perlindungan"
              },
              {
                "id": "b",
                "label": "Toko online-nya memiliki banyak foto produk yang bagus",
                "delta": 5,
                "hint": "Foto produk mudah disalin; bukan bukti kepercayaan"
              },
              {
                "id": "c",
                "label": "HP tersebut adalah keluaran terbaru sehingga wajar langka",
                "delta": 0,
                "hint": "Kelangkaan justru membuat harga normalnya tinggi, bukan murah ekstrem"
              }
            ]
          },
          {
            "id": "janggal_3",
            "title": "Kasus 3 — Instruksi Kerja yang Bertentangan",
            "description": "Atasan lewat chat grup menulis: Hapus SEMUA file laporan bulan ini dari folder bersama SEKARANG, jangan backup, jangan tanya-tanya, dan jangan lapor ke siapa pun. Mana yang paling janggal?",
            "options": [
              {
                "id": "a",
                "label": "Perintah menghapus data penting tanpa backup sambil melarang konfirmasi dan pelaporan — melanggar prosedur wajar",
                "delta": 20,
                "hint": "Instruksi sah selalu bisa diverifikasi lewat jalur resmi; perintah gelap yang melarang verifikasi adalah anomali"
              },
              {
                "id": "b",
                "label": "Instruksi dikirim lewat chat grup sehingga banyak yang baca",
                "delta": 5,
                "hint": "Saluran pengiriman bukan masalah utama; isi perintahnya yang berbahaya"
              },
              {
                "id": "c",
                "label": "Laporan bulan ini jumlahnya banyak sehingga menghapusnya melelahkan",
                "delta": 0,
                "hint": "Beban kerja bukan anomali; penghapusan tanpa jejak itulah masalahnya"
              }
            ]
          },
          {
            "id": "janggal_4",
            "title": "Kasus 4 — Email dari Bank",
            "description": "Ada email mengatasnamakan bank: akunmu akan diblokir dalam 2 jam kecuali kamu klik tautan dan mengisi ulang kata sandi, nomor kartu, dan kode OTP di halaman tersebut. Alamat pengirimnya bank-promo-secure-random123.net. Mana yang paling janggal?",
            "options": [
              {
                "id": "a",
                "label": "Bank asli tidak pernah meminta kata sandi dan OTP lewat tautan email, apalagi dari domain aneh dengan ancaman waktu",
                "delta": 20,
                "hint": "Tiga ciri phishing sekaligus: permintaan kredensial, domain palsu, dan tekanan urgensi"
              },
              {
                "id": "b",
                "label": "Email-nya menggunakan bahasa yang terlalu sopan",
                "delta": 5,
                "hint": "Kesopanan bahasa tidak relevan dengan keaslian pengirim"
              },
              {
                "id": "c",
                "label": "Ancaman blokir 2 jam menunjukkan bank sigap melayani",
                "delta": 0,
                "hint": "Tekanan waktu justru teknik manipulasi agar korban panik dan tidak berpikir"
              }
            ]
          },
          {
            "id": "janggal_5",
            "title": "Kasus 5 — Kabar Simpang Siur",
            "description": "Di grup chat beredar info: besok libur total, semua kegiatan dibatalkan. Pesan itu tanpa tanggal, tanpa kop, tanpa nama penanggung jawab, dan bertentangan dengan pengumuman resmi minggu lalu yang masih berlaku. Mana yang paling janggal?",
            "options": [
              {
                "id": "a",
                "label": "Info tanpa sumber, tanpa tanggal, dan bertentangan dengan pengumuman resmi yang masih berlaku",
                "delta": 20,
                "hint": "Prinsip verifikasi: kabar tanpa identitas sumber dan bertentangan dengan kanal resmi wajib dicurigai"
              },
              {
                "id": "b",
                "label": "Info disebarkan di grup chat yang ramai anggota",
                "delta": 5,
                "hint": "Ramainya grup tidak mengubah status kebenaran sebuah kabar"
              },
              {
                "id": "c",
                "label": "Isi kabar tentang libur sehingga menyenangkan banyak orang",
                "delta": 0,
                "hint": "Kabar yang menyenangkan justru paling sering diteruskan tanpa dicek"
              }
            ]
          },
          {
            "id": "janggal_6",
            "title": "Kasus 6 — Kode QR Tempelan",
            "description": "Di mesin parkir ada stiker kode QR baru yang ditempel menutupi kode asli. Tulisannya: SCAN DI SINI — PROMO PARKIR 90 PERSEN, bayar via transfer ke nomor pribadi, jangan pakai aplikasi resmi. Mana yang paling janggal?",
            "options": [
              {
                "id": "a",
                "label": "QR tempelan yang menutupi kode asli plus arahan menghindari aplikasi resmi dan membayar ke nomor pribadi",
                "delta": 20,
                "hint": "Quishing: stiker palsu mengalihkan pembayaran ke penipu dengan iming-iming diskon"
              },
              {
                "id": "b",
                "label": "Promo 90 persen menunjukkan pengelola sedang berbaik hati",
                "delta": 5,
                "hint": "Diskon ekstrem pada layanan publik yang tarifnya tetap adalah sinyal bahaya"
              },
              {
                "id": "c",
                "label": "Stiker terlihat baru dan bersih sehingga meyakinkan",
                "delta": 0,
                "hint": "Kondisi fisik stiker tidak membuktikan keasliannya"
              }
            ]
          }
        ]
      }
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '085_tasks_replace_2026_09_15')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
