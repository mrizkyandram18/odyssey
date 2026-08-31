-- ============================================================
-- Migration 044: Cleanup Legacy RPG Tables, Harden Family Isolation & Anti-Double-Claim
-- ============================================================

-- 1. Add family_id to odyssey_tasks and backfill
DO $$ 
DECLARE
    v_default_family_id TEXT;
BEGIN
    -- Check for existing family id or create fallback
    SELECT id INTO v_default_family_id FROM odyssey_families LIMIT 1;
    IF v_default_family_id IS NULL THEN
        v_default_family_id := 'crew_default';
        INSERT INTO odyssey_families (id, name) 
        VALUES ('crew_default', 'Keluarga Utama')
        ON CONFLICT (id) DO NOTHING;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'family_id') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN family_id TEXT;
    END IF;

    -- Backfill family_id from task creator profile or default family
    UPDATE odyssey_tasks t
    SET family_id = COALESCE(p.family_id, v_default_family_id)
    FROM odyssey_user_profiles p
    WHERE t.created_by = p.uid AND t.family_id IS NULL;

    UPDATE odyssey_tasks
    SET family_id = v_default_family_id
    WHERE family_id IS NULL;

    ALTER TABLE odyssey_tasks ALTER COLUMN family_id SET NOT NULL;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'fk_odyssey_tasks_family'
    ) THEN
        ALTER TABLE odyssey_tasks 
        ADD CONSTRAINT fk_odyssey_tasks_family FOREIGN KEY (family_id) REFERENCES odyssey_families(id) ON DELETE CASCADE;
    END IF;
END $$;

CREATE INDEX IF NOT EXISTS idx_odyssey_tasks_family_date 
ON odyssey_tasks (family_id, active_date, step_order) 
WHERE is_active = TRUE;

-- 2. HARDEN RPC: odyssey_submit_auto_task (Strict Anti-Double-Claim & Family Isolation)
CREATE OR REPLACE FUNCTION odyssey_submit_auto_task(
    p_task_id BIGINT,
    p_user_uid TEXT,
    p_answers JSONB
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_task RECORD;
    v_profile RECORD;
    v_reward_coins INT;
    v_reward_xp INT;
    v_submission_id BIGINT;
    v_new_coins INT;
    v_new_xp INT;
    v_questions JSONB;
    v_q JSONB;
    v_q_id TEXT;
    v_correct TEXT;
    v_user_ans TEXT;
    v_new_streak INT;
BEGIN
    -- 1. Verify User Profile & Lock Row
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    -- 2. Validate Task & Family Scoping
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

    -- 3. CRITICAL: Anti-Double-Claim Check (Invariant: 1 User + 1 Task = Max 1 Approved Reward)
    IF EXISTS (
        SELECT 1 FROM odyssey_task_submissions 
        WHERE task_id = p_task_id AND user_uid = p_user_uid AND status = 'APPROVED'
    ) THEN
        RAISE EXCEPTION 'Tugas ini sudah diselesaikan dan reward sudah diterima' USING ERRCODE = 'P0004';
    END IF;

    -- 4. Extract questions from config or legacy questions column
    v_questions := COALESCE(v_task.config->'questions', v_task.questions, '[]'::jsonb);

    -- 5. Strict Quiz Validation
    IF jsonb_array_length(v_questions) > 0 THEN
        FOR v_q IN SELECT * FROM jsonb_array_elements(v_questions)
        LOOP
            v_q_id := v_q->>'id';
            v_correct := trim(COALESCE(v_q->>'correct_answer', ''));
            
            IF v_correct = '' THEN
                RAISE EXCEPTION 'Soal kuis tidak memiliki kunci jawaban' USING ERRCODE = 'P0009';
            END IF;

            v_user_ans := trim(COALESCE(p_answers->>v_q_id, ''));
            IF v_user_ans = '' OR lower(v_user_ans) != lower(v_correct) THEN
                RAISE EXCEPTION 'Jawaban kuis belum tepat, silakan periksa kembali' USING ERRCODE = 'P0008';
            END IF;
        END LOOP;
    END IF;

    v_reward_coins := COALESCE(v_task.reward_coins, 50);
    v_reward_xp := COALESCE(v_task.reward_xp, 100);

    -- 6. Upsert Submission Record
    INSERT INTO odyssey_task_submissions (
        task_id, user_uid, submission_type, status, payload, coins_earned, xp_earned, reviewed_at
    ) VALUES (
        p_task_id, p_user_uid, 'AUTO_QUIZ', 'APPROVED', p_answers, v_reward_coins, v_reward_xp, timezone('utc'::text, now())
    )
    ON CONFLICT (task_id, user_uid) DO UPDATE SET
        submission_type = 'AUTO_QUIZ',
        status = 'APPROVED',
        payload = p_answers,
        coins_earned = v_reward_coins,
        xp_earned = v_reward_xp,
        reviewed_at = timezone('utc'::text, now())
    RETURNING id INTO v_submission_id;

    -- 7. Insert Immutable Ledger Transaction (+Coins)
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, v_reward_coins, 'TASK_REWARD', v_submission_id::TEXT, 'Reward: ' || v_task.title);

    -- 8. Update User Profile Projection (Coins & XP)
    UPDATE odyssey_user_profiles
    SET coins = coins + v_reward_coins,
        xp = COALESCE(xp, 0) + v_reward_xp,
        level = floor(sqrt((COALESCE(xp, 0) + v_reward_xp) / 100)) + 1
    WHERE uid = p_user_uid
    RETURNING coins, xp INTO v_new_coins, v_new_xp;

    -- 9. Update Streak
    v_new_streak := odyssey_update_user_streak(p_user_uid);

    RETURN jsonb_build_object(
        'success', true,
        'submission_id', v_submission_id,
        'coins_earned', v_reward_coins,
        'xp_earned', v_reward_xp,
        'new_balance', v_new_coins,
        'new_xp', v_new_xp,
        'streak', v_new_streak
    );
END;
$$;

-- 3. HARDEN RPC: odyssey_submit_manual_task (Family Isolation & Pending Guard)
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
BEGIN
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid = p_user_uid;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
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

    -- Cannot re-submit if already approved
    IF EXISTS (
        SELECT 1 FROM odyssey_task_submissions 
        WHERE task_id = p_task_id AND user_uid = p_user_uid AND status = 'APPROVED'
    ) THEN
        RAISE EXCEPTION 'Tugas ini sudah disetujui sebelumnya' USING ERRCODE = 'P0004';
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

-- 4. HARDEN RPC: odyssey_verify_submission (Admin Family Isolation & Atomic Grant)
CREATE OR REPLACE FUNCTION odyssey_verify_submission(
    p_submission_id BIGINT,
    p_admin_uid TEXT,
    p_status TEXT,
    p_admin_notes TEXT DEFAULT NULL
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_sub RECORD;
    v_task RECORD;
    v_admin RECORD;
    v_member RECORD;
    v_reward_coins INT;
    v_reward_xp INT;
    v_new_coins INT;
    v_new_xp INT;
    v_new_streak INT;
BEGIN
    -- Verify Admin Profile
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid = p_admin_uid;
    IF NOT FOUND OR v_admin.role != 'GUIDE' THEN
        RAISE EXCEPTION 'Hanya admin keluarga yang dapat memverifikasi submission' USING ERRCODE = 'P0003';
    END IF;

    SELECT * INTO v_sub FROM odyssey_task_submissions WHERE id = p_submission_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Submission tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    IF v_sub.status = 'APPROVED' THEN
        RAISE EXCEPTION 'Submission ini sudah disetujui sebelumnya' USING ERRCODE = 'P0004';
    END IF;

    SELECT * INTO v_task FROM odyssey_tasks WHERE id = v_sub.task_id;
    SELECT * INTO v_member FROM odyssey_user_profiles WHERE uid = v_sub.user_uid;

    -- Family Isolation Check
    IF v_admin.family_id IS NOT NULL AND v_member.family_id IS NOT NULL AND v_admin.family_id != v_member.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: Submission bukan milik anggota keluarga Anda' USING ERRCODE = 'P0003';
    END IF;

    v_reward_coins := COALESCE(v_task.reward_coins, 50);
    v_reward_xp := COALESCE(v_task.reward_xp, 100);

    IF p_status = 'APPROVED' THEN
        UPDATE odyssey_task_submissions
        SET status = 'APPROVED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now()),
            coins_earned = v_reward_coins,
            xp_earned = v_reward_xp
        WHERE id = p_submission_id;

        -- Record Immutable Coin Ledger
        INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
        VALUES (v_sub.user_uid, v_reward_coins, 'TASK_REWARD', p_submission_id::TEXT, 'Reward: ' || v_task.title);

        -- Award Coins & EXP
        UPDATE odyssey_user_profiles
        SET coins = coins + v_reward_coins,
            xp = COALESCE(xp, 0) + v_reward_xp,
            level = floor(sqrt((COALESCE(xp, 0) + v_reward_xp) / 100)) + 1
        WHERE uid = v_sub.user_uid
        RETURNING coins, xp INTO v_new_coins, v_new_xp;

        -- Update Streak
        v_new_streak := odyssey_update_user_streak(v_sub.user_uid);

        RETURN jsonb_build_object(
            'success', true,
            'status', 'APPROVED',
            'coins_earned', v_reward_coins,
            'xp_earned', v_reward_xp,
            'new_balance', v_new_coins,
            'new_xp', v_new_xp,
            'streak', v_new_streak
        );
    ELSIF p_status = 'REJECTED' THEN
        UPDATE odyssey_task_submissions
        SET status = 'REJECTED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now())
        WHERE id = p_submission_id;

        RETURN jsonb_build_object('success', true, 'status', 'REJECTED');
    ELSE
        RAISE EXCEPTION 'Status tidak valid' USING ERRCODE = 'P0005';
    END IF;
END;
$$;

-- 5. SAFELY DROP VERIFIED DEAD LEGACY RPG TABLES & DUPLICATE SUBMISSION TABLES
-- Drops only tables that have zero runtime participation in the Family Task & Reward app.
DROP TABLE IF EXISTS odyssey_task_completions CASCADE;
DROP TABLE IF EXISTS odyssey_reactions_legacy CASCADE;
DROP TABLE IF EXISTS odyssey_player_story_fragments CASCADE;
DROP TABLE IF EXISTS odyssey_story_fragments CASCADE;
DROP TABLE IF EXISTS odyssey_lore_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_creative_submissions CASCADE;
DROP TABLE IF EXISTS odyssey_creative_items CASCADE;
DROP TABLE IF EXISTS odyssey_creative_prompt_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_drop_tables CASCADE;
DROP TABLE IF EXISTS odyssey_gift_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_gifts CASCADE;
DROP TABLE IF EXISTS odyssey_player_collections CASCADE;
DROP TABLE IF EXISTS odyssey_collection_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_collections CASCADE;
DROP TABLE IF EXISTS odyssey_exercises CASCADE;
DROP TABLE IF EXISTS odyssey_mission_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_missions CASCADE;
DROP TABLE IF EXISTS odyssey_course_progress CASCADE;
DROP TABLE IF EXISTS odyssey_course_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_journey_progress CASCADE;
DROP TABLE IF EXISTS odyssey_journey_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_learning_concepts CASCADE;
DROP TABLE IF EXISTS odyssey_daily_activity_completions CASCADE;
DROP TABLE IF EXISTS odyssey_daily_activities CASCADE;
DROP TABLE IF EXISTS odyssey_daily_activity CASCADE;
DROP TABLE IF EXISTS odyssey_daily_missions CASCADE;
DROP TABLE IF EXISTS odyssey_achievement_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_achievements CASCADE;
DROP TABLE IF EXISTS odyssey_reactions CASCADE;
DROP TABLE IF EXISTS odyssey_reward_signals CASCADE;
DROP TABLE IF EXISTS odyssey_season_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_balance_configs CASCADE;
DROP TABLE IF EXISTS odyssey_cosmetic_unlocks CASCADE;
DROP TABLE IF EXISTS odyssey_reward_ledgers CASCADE;
