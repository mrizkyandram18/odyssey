-- Migration 037: Seed Realistic Family Learning Content
-- Cleans up fantasy wrappers and introduces practical family learning modules.

-- First, we ensure the new journeys and courses exist
INSERT INTO odyssey_journey_definitions (slug, name, description, "order")
VALUES
  ('literasi-keluarga', 'Literasi Keluarga', 'Belajar keterampilan dasar komunikasi dan kebiasaan baik di rumah.', 1),
  ('literasi-finansial', 'Literasi Finansial', 'Manajemen uang, menabung, dan perencanaan keuangan.', 2),
  ('persiapan-karier', 'Persiapan Karier', 'Mengenal dunia kerja, teknologi, dan keamanan digital.', 3)
ON CONFLICT (slug) DO UPDATE SET name = EXCLUDED.name, description = EXCLUDED.description;

INSERT INTO odyssey_course_definitions (slug, journey, title, description, "order")
VALUES
  ('komunikasi-dasar', 'literasi-keluarga', 'Komunikasi Dasar', 'Belajar menyampaikan pesan dengan baik dan sopan.', 1),
  ('kebiasaan-baik', 'literasi-keluarga', 'Kebiasaan Baik', 'Mengatur jadwal dan rutinitas harian.', 2),
  ('matematika-praktis', 'literasi-finansial', 'Matematika Praktis', 'Hitung-hitungan dasar untuk belanja dan diskon.', 1),
  ('manajemen-uang', 'literasi-finansial', 'Manajemen Uang', 'Cara menghemat dan membedakan kebutuhan vs keinginan.', 2),
  ('dunia-digital', 'persiapan-karier', 'Dunia Digital', 'Memahami teknologi dan menghindari penipuan online.', 1),
  ('dunia-kerja', 'persiapan-karier', 'Dunia Kerja', 'Persiapan melamar pekerjaan dan etika bekerja.', 2)
ON CONFLICT (slug) DO UPDATE SET title = EXCLUDED.title, description = EXCLUDED.description;

-- Update existing missions to map to the new realistic journeys and courses
UPDATE odyssey_mission_definitions SET journey = 'literasi-keluarga', course = 'komunikasi-dasar' WHERE slug IN ('ww-message', 'ww-polite', 'ww-announcement');
UPDATE odyssey_mission_definitions SET journey = 'literasi-keluarga', course = 'kebiasaan-baik' WHERE slug IN ('ww-schedule', 'ww-clean', 'ww-instructions');

UPDATE odyssey_mission_definitions SET journey = 'literasi-finansial', course = 'matematika-praktis' WHERE slug IN ('cc-discount', 'cc-change', 'cc-fraction', 'cc-distance', 'cc-ratio');
UPDATE odyssey_mission_definitions SET journey = 'literasi-finansial', course = 'manajemen-uang' WHERE slug IN ('cc-budget', 'cc-needs', 'cc-installments', 'cc-expense');

UPDATE odyssey_mission_definitions SET journey = 'persiapan-karier', course = 'dunia-digital' WHERE slug IN ('sl-phishing', 'sl-password', 'sl-hoax', 'sl-otp', 'sl-whatsapp', 'cc-scam');
UPDATE odyssey_mission_definitions SET journey = 'persiapan-karier', course = 'dunia-kerja' WHERE slug IN ('sl-cv', 'sl-fake-job', 'sl-interview', 'sl-outfit', 'sl-work-ethics', 'sl-typing', 'sl-note');

-- Hide old fantasy journeys by changing their order to 99
UPDATE odyssey_journey_definitions SET "order" = 99 WHERE slug IN ('whispering-woods', 'clockwork-city', 'starlit-library');

-- Note: Because odyssey_mission_definitions uses the JSON challenge_defs, the actual task content 
-- (like 'Tuliskan daftar kegiatanmu...') is already perfectly realistic and doesn't need to change.
-- The wrap around them is now grounded in reality (e.g. 'Literasi Finansial' instead of 'Clockwork City').
