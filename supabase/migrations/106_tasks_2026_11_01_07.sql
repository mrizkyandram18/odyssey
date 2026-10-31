-- ============================================================
-- Migration 106: Canonical Task Batch 2026-11-01 .. 2026-11-07
--
-- Inserts 21 canonical tasks (3 per day) for November 1–7, 2026.
-- Features:
-- 1. Table-level EXCLUSIVE lock on odyssey_tasks with 5s lock_timeout.
-- 2. Fully bidirectional, NULL-safe IS NOT DISTINCT FROM validation against canonical manifest.
-- 3. Unmarked collision detection across target date range.
-- 4. Fail-closed state machine (0 -> insert, 21 exact -> no-op, 1..20 or >21 or mismatch -> abort).
-- 5. Updates odyssey_schema_version to '106_tasks_2026_11_01_07'.
--
-- ROLLBACK INSTRUCTIONS:
-- As agreed in Gate 4.4 and 4.6, once tasks are released and members begin submissions,
-- destructive DELETE is strictly prohibited to preserve foreign key references and submission history.
-- 1. Non-destructive deactivation (safe post-release):
--    UPDATE public.odyssey_tasks SET is_active = false
--    WHERE config->>'migration' = '106_tasks_2026_11_01_07';
-- 2. Clean rollback (if no submissions exist and clean rollback is explicitly verified):
--    DELETE FROM public.odyssey_tasks
--    WHERE config->>'migration' = '106_tasks_2026_11_01_07';
--    UPDATE public.odyssey_schema_version
--    SET value = '105_harden_reward_rpc_privileges', updated_at = timezone('utc'::text, now())
--    WHERE key = 'schema_version' AND value = '106_tasks_2026_11_01_07';
-- ============================================================

SET LOCAL lock_timeout = '5s';
LOCK TABLE public.odyssey_tasks IN EXCLUSIVE MODE;

DO $$
DECLARE
    v_marker TEXT := '106_tasks_2026_11_01_07';
    v_marker_count INT;
    v_unmarked_collision INT;
    v_missing_from_db INT;
    v_unexpected_in_db INT;
BEGIN
    -- 1. Buat tabel sementara canonical manifest (21 baris)
    CREATE TEMP TABLE temp_canonical_manifest (
        family_id TEXT NOT NULL,
        title TEXT NOT NULL,
        description TEXT NOT NULL,
        task_type TEXT NOT NULL,
        evaluation_type TEXT NOT NULL,
        step_order INT NOT NULL,
        active_date DATE NOT NULL,
        reward_coins INT NOT NULL,
        reward_xp INT NOT NULL,
        target_scope TEXT NOT NULL,
        is_active BOOLEAN NOT NULL,
        config JSONB NOT NULL
    ) ON COMMIT DROP;

    -- 2. Sisipkan 21 task canonical
    INSERT INTO temp_canonical_manifest (
        family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, target_scope, is_active, config
    ) VALUES
    ('demo-crew-1', 'Daster Biasa, Gaya Luar Biasa!', 'Buat tiga gaya berbeda dari satu daster rumahan.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, '2026-11-01', 20, 100, 'ALL', true, '{"min_photos":3,"camera_only":true,"camera_facing":"environment","migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Peragaan Busana Paling Nyeleneh', 'Rekam runway mini menggunakan outfit rumahan.', 'VIDEO', 'ADMIN_REVIEW', 2, '2026-11-01', 20, 100, 'ALL', true, '{"recording":{"enabled":true,"max_duration_seconds":60,"min_duration_seconds":15},"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Kenapa Warna Ini Cocok?', 'Jelaskan kombinasi warna outfit pilihanmu.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 3, '2026-11-01', 20, 100, 'ALL', true, '{"minimum_characters":50,"maximum_characters":500,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Tebak Harga Lima Barang di Rumah', 'Tebak harga 5 barang di rumah, lalu cek harga aslinya.', 'DOCUMENT_UPLOAD', 'ADMIN_REVIEW', 1, '2026-11-02', 20, 100, 'ALL', true, '{"accepted_extensions":[".pdf",".jpg",".png"],"max_file_size_mb":5,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Misi Belanja Rp50 Ribu', 'Simulasi kelola anggaran belanja Rp50.000 untuk 6 kebutuhan dapur.', 'MINI_GAME', 'AUTO', 2, '2026-11-02', 25, 100, 'ALL', true, '{"game":"DECISION_PRIORITY","target_score":100,"scenario":{"initial_balance":0,"currency":"POINTS","events":[{"id":"ev_1","title":"Kebutuhan Pokok - Beras","options":[{"id":"a","label":"Beli beras kiloan kualitas medium Rp15.000","delta":20,"hint":"Pilihan hemat dan mencukupi kebutuhan pokok"},{"id":"b","label":"Beli beras premium mini pack Rp25.000","delta":5,"hint":"Terlalu menguras anggaran awal"},{"id":"c","label":"Tidak beli beras, ganti mie instan saja Rp10.000","delta":0,"hint":"Gizi tidak seimbang untuk makanan pokok"}]},{"id":"ev_2","title":"Lauk Pauk - Protein","options":[{"id":"a","label":"Beli telur 1/2 kg Rp14.000 dan tempe Rp5.000","delta":20,"hint":"Kombinasi protein hewani dan nabati yang sangat ekonomis"},{"id":"b","label":"Beli daging ayam 2 potong Rp22.000","delta":5,"hint":"Cukup mahal, sisa anggaran untuk sayur menipis"},{"id":"c","label":"Beli sosis beku curah Rp8.000","delta":0,"hint":"Kurang bergizi dan berpengawet"}]},{"id":"ev_3","title":"Sayuran Segar","options":[{"id":"a","label":"Beli bayam dan wortel Rp6.000","delta":20,"hint":"Serat dan vitamin lengkap dengan harga terjangkau"},{"id":"b","label":"Beli sayur kalengan Rp12.000","delta":5,"hint":"Boros untuk porsi sedikit"},{"id":"c","label":"Lewatkan sayur hari ini Rp0","delta":0,"hint":"Asupan gizi keluarga tidak lengkap"}]},{"id":"ev_4","title":"Bumbu Dapur","options":[{"id":"a","label":"Beli paket bumbu sop dan bawang Rp5.000","delta":20,"hint":"Cukup untuk memasak seluruh bahan di atas"},{"id":"b","label":"Beli bumbu instan kemasan botol Rp10.000","delta":5,"hint":"Melebihi alokasi bumbu dapur"},{"id":"c","label":"Pakai garam saja Rp0","delta":0,"hint":"Masakan kurang selera dan nutrisi bumbu minim"}]},{"id":"ev_5","title":"Minyak Goreng","options":[{"id":"a","label":"Beli minyak goreng pouch 250ml Rp4.000","delta":20,"hint":"Tepat sesuai kebutuhan memasak hari ini"},{"id":"b","label":"Beli minyak jagung 500ml Rp18.000","delta":0,"hint":"Anggaran belanja langsung defisit/habis"},{"id":"c","label":"Minta tetangga Rp0","delta":5,"hint":"Tidak mandiri dalam perencanaan belanja"}]},{"id":"ev_6","title":"Sisa Anggaran & Evaluasi","options":[{"id":"a","label":"Simpan sisa kembalian Rp1.000 untuk tabungan celengan","delta":20,"hint":"Pengelolaan anggaran sempurna dan tersisa tabungan"},{"id":"b","label":"Belikan permen Rp1.000","delta":5,"hint":"Pengeluaran impulsif meski bernilai kecil"},{"id":"c","label":"Hutang warung karena uang kurang Rp0","delta":0,"hint":"Perencanaan gagal dan menimbulkan beban baru"}]}]},"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Detektif Kemasan', 'Bandingkan 2 produk serupa dari isi netto vs harga.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 3, '2026-11-02', 20, 100, 'ALL', true, '{"min_photos":1,"camera_only":true,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Jembatan Kertas', 'Buat jembatan kertas penahan beban benda kecil.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, '2026-11-03', 20, 100, 'ALL', true, '{"min_photos":1,"camera_only":true,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Pesawat Kertas Terbang', 'Buat, uji, dan rekam terbang pesawat kertas.', 'VIDEO', 'ADMIN_REVIEW', 2, '2026-11-03', 20, 100, 'ALL', true, '{"recording":{"enabled":true,"max_duration_seconds":30,"min_duration_seconds":10},"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Kenapa Hasilnya Berbeda?', 'Ceritakan hasil observasi eksperimen kertas.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 3, '2026-11-03', 20, 100, 'ALL', true, '{"minimum_characters":50,"maximum_characters":500,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Plating Anti-Biasa', 'Sajikan camilan sederhana secara estetik.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, '2026-11-04', 20, 100, 'ALL', true, '{"min_photos":1,"camera_only":true,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Chef Dadakan', 'Rekam proses membuat camilan sederhana dan higienis.', 'VIDEO', 'ADMIN_REVIEW', 2, '2026-11-04', 20, 100, 'ALL', true, '{"recording":{"enabled":true,"max_duration_seconds":60,"min_duration_seconds":15},"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Rahasia Resep', 'Tulis resep beserta kalkulasi estimasi biaya per porsi.', 'DOCUMENT_UPLOAD', 'ADMIN_REVIEW', 3, '2026-11-04', 20, 100, 'ALL', true, '{"accepted_extensions":[".pdf",".jpg",".png"],"max_file_size_mb":5,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Barang Bekas Naik Kelas', 'Ubah kardus/kemasan menjadi barang berguna.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, '2026-11-05', 20, 100, 'ALL', true, '{"min_photos":1,"camera_only":true,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Iklan Produk Buatan Sendiri', 'Rekam video promosi kreasi daur ulang.', 'VIDEO', 'ADMIN_REVIEW', 2, '2026-11-05', 20, 100, 'ALL', true, '{"recording":{"enabled":true,"max_duration_seconds":60,"min_duration_seconds":15},"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Ceritakan Ide di Balik Karyamu', 'Jelaskan konsep dan pemikiran desain karya.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 3, '2026-11-05', 20, 100, 'ALL', true, '{"minimum_characters":50,"maximum_characters":500,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Tebak Gaya', 'Peragakan sebuah profesi dengan pantomim.', 'VIDEO', 'ADMIN_REVIEW', 1, '2026-11-06', 20, 100, 'ALL', true, '{"recording":{"enabled":true,"max_duration_seconds":45,"min_duration_seconds":15},"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Cerita dari Tiga Benda', 'Tulis narasi imajinatif menghubungkan tiga benda acak.', 'TEXT_RESPONSE', 'ADMIN_REVIEW', 2, '2026-11-06', 20, 100, 'ALL', true, '{"minimum_characters":75,"maximum_characters":600,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Buat Kartu Tebak-Tebakan Keluarga', 'Buat lembar flashcard tebakan untuk keluarga.', 'DOCUMENT_UPLOAD', 'ADMIN_REVIEW', 3, '2026-11-06', 20, 100, 'ALL', true, '{"accepted_extensions":[".pdf",".jpg",".png"],"max_file_size_mb":5,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Pameran Mini', 'Tata sudut galeri tiga hasil karya terbaik.', 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 1, '2026-11-07', 20, 100, 'ALL', true, '{"min_photos":1,"camera_only":true,"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Satu Menit Jadi Presenter', 'Presentasikan karya favorit secara verbal.', 'VIDEO', 'ADMIN_REVIEW', 2, '2026-11-07', 20, 100, 'ALL', true, '{"recording":{"enabled":true,"max_duration_seconds":60,"min_duration_seconds":30},"migration":"106_tasks_2026_11_01_07"}'::jsonb),
    ('demo-crew-1', 'Katalog Karya', 'Buat ringkasan katalog karya mingguan.', 'DOCUMENT_UPLOAD', 'ADMIN_REVIEW', 3, '2026-11-07', 20, 100, 'ALL', true, '{"accepted_extensions":[".pdf",".jpg",".png"],"max_file_size_mb":5,"migration":"106_tasks_2026_11_01_07"}'::jsonb);

    -- 3. Assert integritas internal manifest
    IF (SELECT COUNT(*) FROM temp_canonical_manifest) != 21 THEN
        RAISE EXCEPTION 'Internal error: Manifest harus berisi tepat 21 baris' USING ERRCODE = 'P0106';
    END IF;

    IF (SELECT COUNT(DISTINCT (family_id, active_date, step_order)) FROM temp_canonical_manifest) != 21 THEN
        RAISE EXCEPTION 'Internal error: Duplikasi business key terdeteksi pada manifest' USING ERRCODE = 'P0106';
    END IF;

    -- 4. Pengecekan Tabrakan Task Asing (Unmarked Tasks) pada Rentang Target 1-7 Nov
    SELECT COUNT(*) INTO v_unmarked_collision
    FROM public.odyssey_tasks 
    WHERE active_date >= '2026-11-01' AND active_date <= '2026-11-07'
      AND COALESCE(config->>'migration', '') != v_marker;

    IF v_unmarked_collision > 0 THEN
        RAISE EXCEPTION 'Precondition Failed: Terdeteksi % task asing tanpa marker % pada 1-7 Nov 2026!', v_unmarked_collision, v_marker USING ERRCODE = 'P0106';
    END IF;

    -- 5. Evaluasi Baris Ber-Marker Eksisting
    SELECT COUNT(*) INTO v_marker_count
    FROM public.odyssey_tasks 
    WHERE config->>'migration' = v_marker;

    -- State 1-20 atau > 21: STOP MUTLAK
    IF v_marker_count > 0 AND v_marker_count != 21 THEN
        RAISE EXCEPTION 'Precondition Failed: Terdeteksi % baris marker % (Wajib 0 atau 21)', v_marker_count, v_marker USING ERRCODE = 'P0106';
    END IF;

    -- State 21 Baris: Evaluasi Bidirectional & Null-Safe
    IF v_marker_count = 21 THEN
        -- Arah 1: Manifest yang hilang di database
        SELECT COUNT(*) INTO v_missing_from_db
        FROM temp_canonical_manifest m
        WHERE NOT EXISTS (
            SELECT 1 FROM public.odyssey_tasks t
            WHERE t.config->>'migration' = v_marker
              AND t.family_id IS NOT DISTINCT FROM m.family_id
              AND t.active_date IS NOT DISTINCT FROM m.active_date
              AND t.step_order IS NOT DISTINCT FROM m.step_order
              AND t.title IS NOT DISTINCT FROM m.title
              AND t.description IS NOT DISTINCT FROM m.description
              AND t.task_type IS NOT DISTINCT FROM m.task_type
              AND t.evaluation_type IS NOT DISTINCT FROM m.evaluation_type
              AND t.reward_coins IS NOT DISTINCT FROM m.reward_coins
              AND t.reward_xp IS NOT DISTINCT FROM m.reward_xp
              AND t.target_scope IS NOT DISTINCT FROM m.target_scope
              AND t.is_active IS NOT DISTINCT FROM m.is_active
              AND t.config IS NOT DISTINCT FROM m.config
        );

        -- Arah 2: Baris database ber-marker yang tidak cocok ke manifest
        SELECT COUNT(*) INTO v_unexpected_in_db
        FROM public.odyssey_tasks t
        WHERE t.config->>'migration' = v_marker
          AND NOT EXISTS (
            SELECT 1 FROM temp_canonical_manifest m
            WHERE t.family_id IS NOT DISTINCT FROM m.family_id
              AND t.active_date IS NOT DISTINCT FROM m.active_date
              AND t.step_order IS NOT DISTINCT FROM m.step_order
              AND t.title IS NOT DISTINCT FROM m.title
              AND t.description IS NOT DISTINCT FROM m.description
              AND t.task_type IS NOT DISTINCT FROM m.task_type
              AND t.evaluation_type IS NOT DISTINCT FROM m.evaluation_type
              AND t.reward_coins IS NOT DISTINCT FROM m.reward_coins
              AND t.reward_xp IS NOT DISTINCT FROM m.reward_xp
              AND t.target_scope IS NOT DISTINCT FROM m.target_scope
              AND t.is_active IS NOT DISTINCT FROM m.is_active
              AND t.config IS NOT DISTINCT FROM m.config
        );

        IF v_missing_from_db = 0 AND v_unexpected_in_db = 0 THEN
            -- 100% Identik: No-op exit cleanly
            RAISE NOTICE 'Migration 106 idempotent no-op: exactly 21 canonical tasks already present.';
            RETURN;
        ELSE
            RAISE EXCEPTION 'Precondition Failed: 21 task ada tetapi tidak cocok (Missing: %, Unexpected: %)', v_missing_from_db, v_unexpected_in_db USING ERRCODE = 'P0106';
        END IF;
    END IF;

    -- State 0 Baris: Insert seluruh 21 task
    INSERT INTO public.odyssey_tasks (
        family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, target_scope, is_active, config
    )
    SELECT family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, target_scope, is_active, config
    FROM temp_canonical_manifest;

    -- 6. Update Schema Version secara atomik
    INSERT INTO public.odyssey_schema_version (key, value)
    VALUES ('schema_version', v_marker)
    ON CONFLICT (key) DO UPDATE
    SET value = EXCLUDED.value,
        updated_at = timezone('utc'::text, now());

    RAISE NOTICE 'Migration 106 applied successfully: 21 canonical tasks inserted.';
END $$;
