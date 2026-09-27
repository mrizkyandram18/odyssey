-- ============================================================
-- Migration 086: Replace tasks for 16 September 2026 (fresh set)
-- NOTE: All rows below are INITIAL ADMIN CONFIGURATION DATA (seed).
-- Runtime behavior must always read these values from DB/task config;
-- no title/reward/limit/prompt/instruction text may be hardcoded
-- in frontend/backend source. Admin can edit every value below through
-- the existing Admin UI (Tasks tab).
-- 1. DELETES exact tasks 542/543 on 2026-09-16. Production evidence
--    (audit at apply time): both carry 0 submissions in
--    odyssey_task_submissions and 0 linked rows in
--    odyssey_coin_transactions (reference_id), so exact-ID hard-delete
--    is safe. No other rows exist on 2026-09-16. History of other
--    dates is untouched.
-- 2. Inserts 3 fresh canonical tasks for 2026-09-16 forming one mini
--    job-search journey (existing architecture and config shapes only,
--    no new framework, no new tables, no email service):
--    - Step 1: DOCUMENT_UPLOAD (Bikin CV Sederhana) — member creates a
--      simple usable CV on their phone (nama, kontak, ringkasan,
--      pendidikan, pengalaman/organisasi, skill) and uploads it
--      (mobile-friendly extensions, max 4 MB to respect the Vercel
--      payload limit enforced by DocUploadModal). ADMIN_REVIEW via the
--      existing approve/reject flow. SUPPORT: reuses existing document
--      capability; no CV-specific subsystem.
--    - Step 2: TEXT_RESPONSE (Cari Lowongan yang Cocok) — member finds
--      ONE real job listing and reports position + company + link +
--      reason it fits (minimum 120 chars so a bare job title cannot
--      pass; admin verifies the link). ADMIN_REVIEW. SUPPORT: reuses
--      existing text capability; URL is plain text evidence, no new
--      link-storage table.
--    - Step 3: TEXT_RESPONSE (Kirim CV untuk Melamar) — SUPPORT VERDICT:
--      the app has NO email-sending integration (no SMTP/Gmail/Outlook/
--      mail API anywhere; only a VAPID mailto subject string). The task
--      therefore teaches correct application practice (clear subject,
--      polite body, attach CV, verify recipient, re-check) and the
--      member sends the email from their OWN mail client, then reports
--      structured evidence (recipient, subject, body excerpt, sent
--      timestamp, CV-attached statement) for ADMIN_REVIEW. No automated
--      email is ever sent by the app; no third party is contacted by
--      smoke tests.
-- 3. Idempotency: the pre-clean DELETE below matches ONLY the 3 exact
--    new titles on 2026-09-16 (narrow exact-title guard for dev re-runs,
--    same spirit as migrations 052/083/084/085). No broad date/title
--    matching.
-- 4. Does NOT touch tasks of 2026-09-12 (627/628/629), 2026-09-13
--    (630/631/632), 2026-09-15 (633/634/635), tasks 608-620 (610
--    camera_only stays true), payout/economy config, announcements,
--    RLS, RPCs, triggers, or any other date.
-- ============================================================

-- 1. Remove exact legacy tasks for 2026-09-16 (proven dependency-free).
DELETE FROM odyssey_tasks
WHERE id IN (542, 543)
  AND active_date = '2026-09-16';

-- 2. Idempotent pre-clean: remove only a previous partial run of THIS
--    migration (exact new titles, exact date). New titles have never
--    existed before, so this cannot touch any other task.
DELETE FROM odyssey_tasks
WHERE active_date = '2026-09-16'
  AND title IN (
    'Bikin CV Sederhana',
    'Cari Lowongan yang Cocok',
    'Kirim CV untuk Melamar'
  );

-- 3. Insert 3 fresh canonical tasks for 2026-09-16.
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
    'Bikin CV Sederhana',
    'Buat satu CV sederhana 1 halaman yang layak dipakai melamar pekerjaan. WAJIB memuat: 1) nama lengkap, 2) kontak yang bisa dihubungi (nomor HP dan/atau email), 3) ringkasan singkat 2-3 kalimat tentang dirimu, 4) pendidikan terakhir, 5) pengalaman kerja ATAU pengalaman organisasi/kegiatan jika belum pernah bekerja, 6) skill yang kamu kuasai. Boleh diketik di HP lalu disimpan sebagai PDF/DOCX/TXT, atau ditulis tangan yang rapi lalu difoto dengan jelas (JPG/PNG). Tidak perlu memaksa field yang tidak relevan dengan kondisimu — tulis yang jujur dan apa adanya.',
    'DOCUMENT_UPLOAD',
    'ADMIN_REVIEW',
    1,
    '2026-09-16',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "allowed_extensions": ["pdf", "docx", "txt", "jpg", "jpeg", "png"],
      "max_file_size_mb": 4,
      "instruction": "Buat 1 CV sederhana yang siap dilampirkan ke lamaran. Isi WAJIB: nama lengkap; kontak (HP dan/atau email); ringkasan singkat 2-3 kalimat; pendidikan terakhir; pengalaman kerja ATAU pengalaman organisasi/kegiatan; skill. Boleh diketik di HP (simpan PDF/DOCX/TXT) atau ditulis tangan rapi lalu difoto jelas (JPG/PNG). Maksimal 4 MB. Admin menolak file yang tidak bisa dibuka, CV tanpa nama/kontak, atau file yang bukan CV."
    }'::jsonb
),
(
    'demo-crew-1',
    'Cari Lowongan yang Cocok',
    'Temukan SATU lowongan pekerjaan NYATA yang sesuai dengan kemampuan atau minatmu (boleh dari Jobstreet, Glints, JobStreet, LinkedIn, Instagram perusahaan, situs resmi perusahaan, atau papan info lowongan lainnya). Bukan sekadar menyebut pekerjaan impian — harus ada lowongan yang benar-benar kamu temukan. Catat posisinya, perusahaannya, simpan link lowongannya, dan jelaskan singkat kenapa lowongan itu cocok untukmu.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    2,
    '2026-09-16',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis hasil pencarian lowonganmu dengan format:\n1) Posisi yang dilamar:\n2) Nama perusahaan:\n3) Link lowongan (tempel URL lengkapnya):\n4) Sumber lowongan (misalnya Jobstreet / Glints / Instagram / situs perusahaan):\n5) Kenapa lowongan ini cocok untukmu (2-3 kalimat: hubungkan dengan kemampuan, pendidikan, atau minatmu):\n\nWajib mencantumkan link yang bisa dibuka. Admin akan menolak jawaban tanpa link, tanpa nama perusahaan, atau tanpa alasan yang jelas.",
      "minimum_characters": 120,
      "maximum_characters": 2000
    }'::jsonb
),
(
    'demo-crew-1',
    'Kirim CV untuk Melamar',
    'Pilih SATU lowongan dari tugas sebelumnya, lalu kirim CV-mu untuk melamar lewat EMAIL dari aplikasi email milikmu sendiri (Gmail, Yahoo, Outlook, atau email HP lainnya) — Odyssey tidak mengirim email untukmu. Praktik yang benar: 1) gunakan subject yang jelas, contoh: Lamaran — [Posisi] — [Nama Lengkap]; 2) tulis body email yang sopan (salam pembuka, perkenalan 1-2 kalimat, posisi yang dilamar, penutup dan terima kasih); 3) lampirkan file CV dari tugas pertama; 4) pastikan alamat email penerima benar; 5) baca ulang semuanya sebelum menekan kirim. Setelah terkirim, laporkan buktinya di sini. JANGAN mengirim ke alamat sembarangan untuk coba-coba — kirim hanya ke alamat perekrut resmi dari lowongan yang kamu temukan, atau ke alamat emailmu sendiri sebagai latihan jika belum siap melamar sungguhan.',
    'TEXT_RESPONSE',
    'ADMIN_REVIEW',
    3,
    '2026-09-16',
    40,
    100,
    'ALL',
    TRUE,
    '{
      "prompt": "Tulis bukti pengiriman lamaranmu dengan format:\n1) Email penerima (tujuan lamaran):\n2) Subject email yang kamu pakai:\n3) Isi/body email yang kamu kirim (salin minimal 3 kalimat pembuka dan penutup):\n4) Nama file CV yang dilampirkan:\n5) Tanggal dan jam pengiriman:\n6) Pernyataan cek ulang (contoh: Saya sudah memastikan penerima benar, subject jelas, body sopan, dan CV terlampir sebelum mengirim):\n\nContoh subject yang baik: Lamaran — Asisten Toko — Budi Santoso. Admin akan menolak laporan tanpa penerima, tanpa subject, tanpa isi email, atau tanpa pernyataan file CV dilampirkan.",
      "minimum_characters": 120,
      "maximum_characters": 2000
    }'::jsonb
)
;

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '086_tasks_replace_2026_09_16')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
