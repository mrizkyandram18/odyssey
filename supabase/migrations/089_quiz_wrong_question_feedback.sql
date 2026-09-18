-- 089_quiz_wrong_question_feedback.sql
-- Quiz auto-grading now reports WHICH question numbers were wrong instead of a
-- generic rejection. Error code stays P0008; the correct answers are still
-- never revealed (only 1-based question positions). Frontend displays the
-- message so members can retry exactly the wrong items.
-- Full redefinition of odyssey_submit_auto_task (last defined in 073);
-- all other behavior (guards, cap, ledger, streak) is unchanged.

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
    v_task RECORD; v_profile RECORD; v_reward_coins INT; v_reward_xp INT; v_submission_id BIGINT;
    v_new_coins INT; v_new_xp INT; v_questions JSONB; v_q JSONB; v_q_id TEXT; v_correct TEXT; v_user_ans TEXT;
    v_new_streak INT; v_game_target INT; v_game_score INT;
    v_target INT; v_actual INT;
    v_cap INT; v_earned INT;
    v_q_idx INT; v_wrong INT[];
BEGIN
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid=p_user_uid FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE='P0007'; END IF;
    SELECT * INTO v_task FROM odyssey_tasks WHERE id=p_task_id;
    IF NOT FOUND THEN RAISE EXCEPTION 'Task tidak ditemukan' USING ERRCODE='P0002'; END IF;
    IF NOT v_task.is_active THEN RAISE EXCEPTION 'Task sedang tidak aktif' USING ERRCODE='P0001'; END IF;
    IF v_task.family_id IS NOT NULL AND v_profile.family_id IS NOT NULL AND v_task.family_id != v_profile.family_id THEN RAISE EXCEPTION 'Akses ditolak: Task bukan milik keluarga Anda' USING ERRCODE='P0003'; END IF;
    IF EXISTS (SELECT 1 FROM odyssey_task_submissions WHERE task_id=p_task_id AND user_uid=p_user_uid AND status='APPROVED') THEN RAISE EXCEPTION 'Tugas ini sudah diselesaikan dan reward sudah diterima' USING ERRCODE='P0004'; END IF;
    v_questions := COALESCE(v_task.config->'questions', v_task.questions, '[]'::jsonb);
    IF jsonb_array_length(v_questions) > 0 THEN
        v_q_idx := 0; v_wrong := '{}'::INT[];
        FOR v_q IN SELECT * FROM jsonb_array_elements(v_questions) LOOP
            v_q_idx := v_q_idx + 1;
            v_q_id := v_q->>'id'; v_correct := trim(COALESCE(v_q->>'correct_answer',''));
            IF v_correct='' THEN RAISE EXCEPTION 'Soal kuis tidak memiliki kunci jawaban' USING ERRCODE='P0009'; END IF;
            v_user_ans := trim(COALESCE(p_answers->>v_q_id,''));
            IF v_user_ans='' OR (lower(v_user_ans)!=lower(v_correct) AND lower(v_user_ans) NOT LIKE lower(v_correct)||'.%' AND lower(v_user_ans) NOT LIKE lower(v_correct)||')%' AND lower(v_correct) NOT LIKE lower(v_user_ans)||'.%' AND lower(v_correct) NOT LIKE lower(v_user_ans)||')%') THEN v_wrong := v_wrong || v_q_idx; END IF;
        END LOOP;
        IF array_length(v_wrong, 1) > 0 THEN RAISE EXCEPTION 'Jawaban kuis belum tepat pada soal nomor %, silakan periksa kembali', array_to_string(v_wrong, ', ') USING ERRCODE='P0008'; END IF;
    END IF;
    IF v_task.task_type='MINI_GAME' THEN
        v_game_target:=COALESCE((v_task.config->>'target_score')::INT,0); v_game_score:=COALESCE((p_answers->>'score')::INT,0);
        IF v_game_score<0 OR v_game_score>1000000 THEN RAISE EXCEPTION 'Skor permainan tidak valid' USING ERRCODE='P0008'; END IF;
        IF v_game_target>0 AND v_game_score < v_game_target THEN RAISE EXCEPTION 'Skor permainan belum mencapai target minimum (% vs target %)', v_game_score, v_game_target USING ERRCODE='P0008'; END IF;
    END IF;
    -- Resolve target (member-specific, global default 0; legacy 3200 retired) and compute scaled reward (user-aware window)
    v_target := COALESCE(v_profile.monthly_coin_target, COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='default_monthly_coin_target'),0));
    IF v_target IS NULL OR v_target<1 OR v_target>10000 THEN v_target:=0; END IF;
    v_actual := odyssey_calc_target_reward(v_target, p_task_id, v_profile.family_id, p_user_uid);
    v_reward_xp := COALESCE(v_task.reward_xp,100);

    -- === EARNING CAP ENFORCEMENT (authoritative) ===
    v_cap := odyssey_get_effective_earning_cap(p_user_uid);
    IF v_cap IS NOT NULL AND v_cap > 0 THEN
        -- Must recalculate earned AFTER profile FOR UPDATE to ensure serialization
        v_earned := odyssey_earned_this_period(p_user_uid);
        IF v_earned >= v_cap THEN
            RAISE EXCEPTION 'Batas earning bulanan tercapai (%/%). Tidak dapat memperoleh reward lagi sampai periode berikutnya.', v_earned, v_cap USING ERRCODE='P0016';
        END IF;
        IF v_actual > 0 AND v_earned + v_actual > v_cap THEN
            RAISE EXCEPTION 'Reward (%) akan melebihi batas earning bulanan (%/%, sisa %). Tugas tidak dapat diberi reward pada periode ini.', v_actual, v_earned, v_cap, (v_cap - v_earned) USING ERRCODE='P0016';
        END IF;
    END IF;

    INSERT INTO odyssey_task_submissions (task_id, user_uid, submission_type, status, payload, coins_earned, xp_earned, reviewed_at)
    VALUES (p_task_id, p_user_uid, 'AUTO_QUIZ', 'APPROVED', p_answers, v_actual, v_reward_xp, timezone('utc'::text, now()))
    ON CONFLICT (task_id, user_uid) DO UPDATE SET submission_type='AUTO_QUIZ', status='APPROVED', payload=p_answers, coins_earned=v_actual, xp_earned=v_reward_xp, reviewed_at=timezone('utc'::text, now())
    RETURNING id INTO v_submission_id;
    IF v_actual > 0 THEN
        INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description) VALUES (p_user_uid, v_actual, 'TASK_REWARD', v_submission_id::TEXT, 'Reward: ' || v_task.title);
    END IF;
    UPDATE odyssey_user_profiles SET coins=coins+v_actual, xp=COALESCE(xp,0)+v_reward_xp, level=floor(sqrt((COALESCE(xp,0)+v_reward_xp)/100))+1 WHERE uid=p_user_uid RETURNING coins, xp INTO v_new_coins, v_new_xp;
    v_new_streak := odyssey_update_user_streak(p_user_uid);
    RETURN jsonb_build_object('success',true,'submission_id',v_submission_id,'coins_earned',v_actual,'xp_earned',v_reward_xp,'new_balance',v_new_coins,'new_xp',v_new_xp,'streak',v_new_streak,'base_reward',v_task.reward_coins,'target',v_target);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_submit_auto_task(BIGINT,TEXT,JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_submit_auto_task(BIGINT,TEXT,JSONB) TO service_role;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','089_quiz_wrong_question_feedback')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
