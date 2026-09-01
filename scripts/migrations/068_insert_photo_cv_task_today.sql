-- ============================================================
-- Migration 068: Insert Foto Langsung Kesiapan Profil CV as Task #2 for today (Asia/Jakarta)
-- Idempotent, shifts existing step_order >=2, inserts camera-only photo task
-- Only active for today (active_date = today in Asia/Jakarta), not recurring
-- ============================================================

DO $$
DECLARE
  v_today DATE := (timezone('Asia/Jakarta', now()))::date;
  v_family RECORD;
  v_exists BOOLEAN;
  v_title TEXT := 'Foto Langsung Kesiapan Profil CV (Rapi & Profesional)';
  v_description TEXT := 'Latih standar foto lamaran kerja profesional dengan mengambil foto langsung dari kamera HP menggunakan pakaian rapi dan tanda kesiapan kerja.';
  v_config JSONB := jsonb_build_object(
    'instruction', '1. Gunakan kemeja/pakaian berkerah rapi seperti standar foto CV.' || E'\n' ||
                   '2. Pegang kertas kecil di depan dada bertuliskan: [Nama Lengkap] - Siap Kerja [Tanggal Hari Ini]' || E'\n' ||
                   '3. Wajah tegak menghadap kamera, tersenyum ramah, dengan latar belakang dinding polos.' || E'\n' ||
                   '4. Ambil foto langsung menggunakan kamera HP lalu kirim.',
    'camera_only', true,
    'require_live_capture', true,
    'max_files', 1,
    'accepted_mime_types', jsonb_build_array('image/jpeg','image/png','image/webp')
  );
BEGIN
  -- Ensure at least one family exists (for local dev)
  -- Shifts and inserts per family that has tasks today or all families
  FOR v_family IN SELECT id FROM odyssey_families LOOP
    -- Check if task already exists for this family today (idempotent)
    SELECT EXISTS (
      SELECT 1 FROM odyssey_tasks
      WHERE family_id = v_family.id
        AND active_date = v_today
        AND title = v_title
    ) INTO v_exists;
    IF v_exists THEN
      CONTINUE;
    END IF;

    -- Shift existing tasks today step_order >=2 by +1 (2→3, 3→4...)
    -- Do in descending order to avoid unique violation if constraint exists
    UPDATE odyssey_tasks
    SET step_order = step_order + 1
    WHERE family_id = v_family.id
      AND active_date = v_today
      AND step_order >= 2
      AND is_active = true;

    -- Insert new camera-only task as step_order 2
    INSERT INTO odyssey_tasks (
      family_id, title, description, task_type, evaluation_type, step_order, active_date,
      reward_coins, reward_xp, config, is_active
    ) VALUES (
      v_family.id,
      v_title,
      v_description,
      'PHOTO_UPLOAD',
      'ADMIN_REVIEW',
      2,
      v_today,
      50,
      50,
      v_config,
      true
    );
  END LOOP;

  -- Fallback: if no families exist (should not happen), still ensure demo-crew-1 exists and has task
  IF NOT EXISTS (SELECT 1 FROM odyssey_families) THEN
    INSERT INTO odyssey_families (id, name) VALUES ('demo-crew-1', 'Keluarga Demo') ON CONFLICT (id) DO NOTHING;
    -- Retry single family
    SELECT EXISTS (
      SELECT 1 FROM odyssey_tasks WHERE family_id='demo-crew-1' AND active_date=v_today AND title=v_title
    ) INTO v_exists;
    IF NOT v_exists THEN
      UPDATE odyssey_tasks SET step_order = step_order + 1
      WHERE family_id='demo-crew-1' AND active_date=v_today AND step_order >=2 AND is_active=true;
      INSERT INTO odyssey_tasks (family_id, title, description, task_type, evaluation_type, step_order, active_date, reward_coins, reward_xp, config, is_active)
      VALUES ('demo-crew-1', v_title, v_description, 'PHOTO_UPLOAD', 'ADMIN_REVIEW', 2, v_today, 50, 50, v_config, true);
    END IF;
  END IF;
END $$;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','068_insert_photo_cv_task_today')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
