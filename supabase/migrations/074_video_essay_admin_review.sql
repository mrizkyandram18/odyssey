-- ============================================================
-- Migration 074: VIDEO + Essay Requires Admin Review (forward-only)
-- VIDEO tasks carrying an essay prompt (config.prompt, i.e. frontend
-- video_answer_mode='essay') must go through MANUAL_VERIFY / PENDING
-- instead of instant AUTO approval. Mirrors TEXT_RESPONSE behavior.
-- Safe: CREATE OR REPLACE only, signature preserved, no schema change.
-- 1. Extends odyssey_submit_manual_task text-length validation to
--    VIDEO+prompt (pure watch-only VIDEO and VIDEO+QUIZ unaffected).
-- 2. Backfills stored evaluation_type AUTO -> ADMIN_REVIEW for existing
--    VIDEO+prompt rows (backend resolver also overrides at read time).
-- Base: odyssey_submit_manual_task body copied verbatim from 065
-- (latest version incl. P0021 block check); only the essay-validation
-- condition is widened.
-- ============================================================

CREATE OR REPLACE FUNCTION odyssey_submit_manual_task(
    p_task_id BIGINT,
    p_user_uid TEXT,
    p_payload JSONB
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_task RECORD;
    v_profile RECORD;
    v_submission_id BIGINT;
    v_min_chars INT;
    v_max_chars INT;
    v_text_len INT;
    v_text_content TEXT;
BEGIN
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;
    IF NOT v_profile.is_active THEN
        RAISE EXCEPTION 'Akun Anda diblokir, tidak dapat mengumpulkan tugas' USING ERRCODE='P0021';
    END IF;
    SELECT * INTO v_task FROM odyssey_tasks WHERE id = p_task_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Task tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;
    IF NOT v_task.is_active THEN
        RAISE EXCEPTION 'Task sedang tidak aktif' USING ERRCODE = 'P0001';
    END IF;
    IF v_task.family_id IS NOT NULL AND v_profile.family_id IS NOT NULL AND v_task.family_id != v_profile.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: Task bukan milik keluarga Anda' USING ERRCODE = 'P0003';
    END IF;
    IF EXISTS (
        SELECT 1 FROM odyssey_task_submissions
        WHERE task_id = p_task_id AND user_uid = p_user_uid AND status = 'APPROVED'
    ) THEN
        RAISE EXCEPTION 'Tugas ini sudah disetujui sebelumnya' USING ERRCODE = 'P0004';
    END IF;
    -- Essay-length validation: TEXT_RESPONSE always; VIDEO only when it
    -- carries an essay prompt (VIDEO + Esai Teks Panjang). Pure watch-only
    -- VIDEO and VIDEO+QUIZ skip text validation entirely.
    IF v_task.task_type = 'TEXT_RESPONSE'
       OR (v_task.task_type = 'VIDEO' AND trim(COALESCE(v_task.config->>'prompt', '')) <> '')
    THEN
        v_text_content := trim(COALESCE(p_payload->>'text', ''));
        v_text_len := length(v_text_content);
        v_min_chars := COALESCE((v_task.config->>'minimum_characters')::INT, 1);
        v_max_chars := COALESCE((v_task.config->>'maximum_characters')::INT, 5000);
        IF v_text_len < v_min_chars THEN
            RAISE EXCEPTION 'Panjang teks minimal % karakter (saat ini: %)', v_min_chars, v_text_len USING ERRCODE = 'P0008';
        END IF;
        IF v_text_len > v_max_chars THEN
            RAISE EXCEPTION 'Panjang teks maksimal % karakter', v_max_chars USING ERRCODE = 'P0008';
        END IF;
    END IF;
    INSERT INTO odyssey_task_submissions (
        task_id, user_uid, submission_type, status, payload, created_at, admin_notes
    ) VALUES (
        p_task_id, p_user_uid, 'MANUAL_VERIFY', 'PENDING', p_payload, timezone('utc'::text, now()), NULL
    )
    ON CONFLICT (task_id, user_uid) DO UPDATE SET
        payload = p_payload,
        status = 'PENDING',
        created_at = timezone('utc'::text, now()),
        admin_notes = NULL
    RETURNING id INTO v_submission_id;
    RETURN jsonb_build_object(
        'success', true,
        'submission_id', v_submission_id,
        'status', 'PENDING'
    );
END;
$$;
REVOKE ALL ON FUNCTION odyssey_submit_manual_task(BIGINT,TEXT,JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_submit_manual_task(BIGINT,TEXT,JSONB) TO service_role;

-- Backfill: existing VIDEO+prompt rows stored as AUTO must be ADMIN_REVIEW.
-- (Backend resolver overrides at read/submit time regardless; this keeps
-- the stored column truthful for admin list display and direct DB reads.)
UPDATE odyssey_tasks
SET evaluation_type = 'ADMIN_REVIEW'
WHERE task_type = 'VIDEO'
  AND trim(COALESCE(config->>'prompt', '')) <> ''
  AND evaluation_type IS DISTINCT FROM 'ADMIN_REVIEW';

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','074_video_essay_admin_review')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
