-- ============================================================
-- Migration 047: Bulletproof Task Engine RPC Hardening
-- Hardens auto-grading evaluation with deterministic quiz answer matching
-- and strict range-checked mini-game bounds.
-- ============================================================

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

    -- 3. Anti-Double-Claim Check (Invariant: 1 User + 1 Task = Max 1 Approved Reward)
    IF EXISTS (
        SELECT 1 FROM odyssey_task_submissions 
        WHERE task_id = p_task_id AND user_uid = p_user_uid AND status = 'APPROVED'
    ) THEN
        RAISE EXCEPTION 'Tugas ini sudah diselesaikan dan reward sudah diterima' USING ERRCODE = 'P0004';
    END IF;

    -- 4. Server-Side Validation based on Task Type
    -- 4a. QUIZ / VIDEO_QUIZ Question validation (Deterministic letter code & text matching)
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
            IF v_user_ans = '' OR (
                lower(v_user_ans) != lower(v_correct) AND
                lower(v_user_ans) NOT LIKE lower(v_correct) || '.%' AND
                lower(v_user_ans) NOT LIKE lower(v_correct) || ')%' AND
                lower(v_correct) NOT LIKE lower(v_user_ans) || '.%' AND
                lower(v_correct) NOT LIKE lower(v_user_ans) || ')%'
            ) THEN
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

-- 2. Record migration version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '047_bulletproof_task_engine')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
