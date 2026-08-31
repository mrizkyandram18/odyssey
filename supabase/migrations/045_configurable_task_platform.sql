-- ============================================================
-- Migration 045: Configurable Family Task Platform Schema & RPC Upgrades
-- ============================================================

-- 1. Upgrade odyssey_tasks to support generic task types & evaluation_type
DO $$ BEGIN
    -- Add evaluation_type column if not present
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'evaluation_type') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN evaluation_type TEXT NOT NULL DEFAULT 'AUTO';
    END IF;

    -- Drop old task_type constraint and allow expanded generic types + backward compatibility aliases
    ALTER TABLE odyssey_tasks DROP CONSTRAINT IF EXISTS odyssey_tasks_task_type_check;
    ALTER TABLE odyssey_tasks ADD CONSTRAINT odyssey_tasks_task_type_check 
        CHECK (task_type IN (
            'VIDEO', 
            'QUIZ', 
            'PHOTO_UPLOAD', 
            'DOCUMENT_UPLOAD', 
            'TEXT_RESPONSE', 
            'MINI_GAME', 
            'VIDEO_QUIZ', 
            'PHOTO_PROOF', 
            'GENERAL', 
            'YOUTUBE_VIDEO'
        ));

    -- Ensure evaluation_type constraint
    ALTER TABLE odyssey_tasks DROP CONSTRAINT IF EXISTS odyssey_tasks_eval_type_check;
    ALTER TABLE odyssey_tasks ADD CONSTRAINT odyssey_tasks_eval_type_check
        CHECK (evaluation_type IN ('AUTO', 'ADMIN_REVIEW'));
END $$;

-- 2. HARDEN RPC: odyssey_submit_auto_task
-- Handles auto-graded tasks: QUIZ, VIDEO, MINI_GAME, VIDEO_QUIZ, YOUTUBE_VIDEO, GENERAL
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
    v_game_target INT;
    v_game_score INT;
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

    -- 4. Server-Side Validation based on Task Type
    -- 4a. QUIZ / VIDEO_QUIZ Question validation
    v_questions := COALESCE(v_task.config->'questions', v_task.questions, '[]'::jsonb);
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

    -- 4b. MINI_GAME validation (bounds & target score check)
    IF v_task.task_type = 'MINI_GAME' THEN
        v_game_target := COALESCE((v_task.config->>'target_score')::INT, 0);
        v_game_score := COALESCE((p_answers->>'score')::INT, 0);
        
        -- Score cannot be negative or absurdly infinite (> 1,000,000)
        IF v_game_score < 0 OR v_game_score > 1000000 THEN
            RAISE EXCEPTION 'Skor permainan tidak valid' USING ERRCODE = 'P0008';
        END IF;

        IF v_game_target > 0 AND v_game_score < v_game_target THEN
            RAISE EXCEPTION 'Skor permainan belum mencapai target minimum (% vs target %)', v_game_score, v_game_target USING ERRCODE = 'P0008';
        END IF;
    END IF;

    v_reward_coins := COALESCE(v_task.reward_coins, 50);
    v_reward_xp := COALESCE(v_task.reward_xp, 100);

    -- 5. Upsert Submission Record
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

    -- 6. Insert Immutable Ledger Transaction (+Coins)
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, v_reward_coins, 'TASK_REWARD', v_submission_id::TEXT, 'Reward: ' || v_task.title);

    -- 7. Update User Profile Projection (Coins & XP)
    UPDATE odyssey_user_profiles
    SET coins = coins + v_reward_coins,
        xp = COALESCE(xp, 0) + v_reward_xp,
        level = floor(sqrt((COALESCE(xp, 0) + v_reward_xp) / 100)) + 1
    WHERE uid = p_user_uid
    RETURNING coins, xp INTO v_new_coins, v_new_xp;

    -- 8. Update Streak
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

-- 3. HARDEN RPC: odyssey_submit_manual_task
-- Handles tasks requiring ADMIN_REVIEW: PHOTO_UPLOAD, DOCUMENT_UPLOAD, TEXT_RESPONSE, PHOTO_PROOF
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

    -- Validation for TEXT_RESPONSE task type
    IF v_task.task_type = 'TEXT_RESPONSE' THEN
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
