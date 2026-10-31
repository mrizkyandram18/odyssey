-- ============================================================
-- Migration 102: Per-User Monthly Earning Cap & Calendar-Month Invariant
-- 
-- 1. odyssey_get_effective_earning_cap:
--    Authoritative per-member cap: validated administrator-authorized limit
--    stored on odyssey_user_profiles.monthly_earning_cap.
--    If NULL, falls back to odyssey_system_config.default_monthly_earning_cap (3320).
--    0 = explicitly unlimited.
--    NON-NEGOTIABLE FINANCIAL RULE: Level progression, RPG bonuses, and
--    derived formulas must NEVER increase the cap above the admin-authorized
--    budget. The effective cap equals exactly the authorized cap.
--
-- 2. odyssey_submit_auto_task:
--    - Enforces graceful capping: if remaining allowance > 0, credits exactly
--      min(v_actual, v_remaining) so total never exceeds authorized cap.
--    - If remaining allowance is 0, task is APPROVED with coins_earned = 0,
--      XP and streak are granted, and NO transaction is inserted into ledger.
--    - Idempotency guard: if a TASK_REWARD transaction for this submission
--      already exists, returns current balance without re-crediting.
--
-- 3. odyssey_verify_submission:
--    - Enforces graceful capping: if remaining allowance > 0, credits exactly
--      min(v_actual, v_remaining).
--    - If remaining allowance is 0, task is APPROVED with coins_earned = 0,
--      XP and streak are granted, and approval succeeds without P0016 rejection.
--
-- 4. Unique index on odyssey_coin_transactions (reference_id) WHERE type = 'TASK_REWARD':
--    Guarantees ledger atomicity and prevents duplicate reward entries at DB level.
-- ============================================================

-- 1. Authoritative Per-Member Cap Resolver
CREATE OR REPLACE FUNCTION odyssey_get_effective_earning_cap(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_cap INT;
    v_global INT;
BEGIN
    -- 1. Fetch default_monthly_earning_cap from DB SOT
    SELECT NULLIF(trim(value), '')::INT INTO v_global
    FROM odyssey_system_config
    WHERE key = 'default_monthly_earning_cap';

    IF NOT FOUND OR v_global IS NULL OR v_global < 0 THEN
        v_global := 3320;
    END IF;

    -- 2. Fetch user custom cap
    SELECT monthly_earning_cap
    INTO v_cap
    FROM odyssey_user_profiles
    WHERE uid = p_user_uid;

    -- If NULL, inherit system default
    IF v_cap IS NULL THEN
        v_cap := v_global;
    END IF;

    -- 0 = unlimited
    IF v_cap = 0 THEN
        RETURN 0;
    END IF;

    -- Non-negotiable financial invariant:
    -- Effective cap is strictly the validated administrator-authorized cap.
    -- Level progression / derived formulas must never inflate this budget.
    RETURN v_cap;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_get_effective_earning_cap(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_get_effective_earning_cap(TEXT) TO service_role;

-- 2. Ledger Idempotency Unique Index (Pre-verified 0 duplicates)
CREATE UNIQUE INDEX IF NOT EXISTS uq_task_reward_ledger_ref
    ON odyssey_coin_transactions (reference_id)
    WHERE type = 'TASK_REWARD';

-- 3. Harden odyssey_submit_auto_task with Graceful Capping & Idempotency
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
    v_cap INT; v_earned INT; v_remaining INT;
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

    -- Resolve target and compute scaled reward
    v_target := COALESCE(v_profile.monthly_coin_target, COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='default_monthly_coin_target'),0));
    IF v_target IS NULL OR v_target<1 OR v_target>10000 THEN v_target:=0; END IF;
    v_actual := odyssey_calc_target_reward(v_target, p_task_id, v_profile.family_id, p_user_uid);
    v_reward_xp := COALESCE(v_task.reward_xp,100);

    -- === AUTHORITATIVE PER-MEMBER EARNING CAP ENFORCEMENT ===
    v_cap := odyssey_get_effective_earning_cap(p_user_uid);
    IF v_cap IS NOT NULL AND v_cap > 0 THEN
        v_earned := odyssey_earned_this_period(p_user_uid);
        v_remaining := GREATEST(0, v_cap - v_earned);
        IF v_actual > v_remaining THEN
            v_actual := v_remaining; -- Gracefully clamp reward to remaining allowance
        END IF;
    END IF;

    INSERT INTO odyssey_task_submissions (task_id, user_uid, submission_type, status, payload, coins_earned, xp_earned, reviewed_at)
    VALUES (p_task_id, p_user_uid, 'AUTO_QUIZ', 'APPROVED', p_answers, v_actual, v_reward_xp, timezone('utc'::text, now()))
    ON CONFLICT (task_id, user_uid) DO UPDATE SET submission_type='AUTO_QUIZ', status='APPROVED', payload=p_answers, coins_earned=v_actual, xp_earned=v_reward_xp, reviewed_at=timezone('utc'::text, now())
    RETURNING id INTO v_submission_id;

    -- Idempotent ledger insertion
    IF v_actual > 0 THEN
        IF NOT EXISTS (
            SELECT 1 FROM odyssey_coin_transactions
            WHERE type = 'TASK_REWARD' AND reference_id = v_submission_id::TEXT
        ) THEN
            INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
            VALUES (p_user_uid, v_actual, 'TASK_REWARD', v_submission_id::TEXT, 'Reward: ' || v_task.title);
        ELSE
            -- Already credited: do not re-add to profile balance
            v_actual := 0;
        END IF;
    END IF;

    UPDATE odyssey_user_profiles
    SET coins = coins + v_actual,
        xp = COALESCE(xp,0) + v_reward_xp,
        level = floor(sqrt((COALESCE(xp,0) + v_reward_xp)/100)) + 1
    WHERE uid = p_user_uid
    RETURNING coins, xp INTO v_new_coins, v_new_xp;

    v_new_streak := odyssey_update_user_streak(p_user_uid);
    RETURN jsonb_build_object(
        'success', true,
        'submission_id', v_submission_id,
        'coins_earned', v_actual,
        'xp_earned', v_reward_xp,
        'new_balance', v_new_coins,
        'new_xp', v_new_xp,
        'streak', v_new_streak,
        'base_reward', v_task.reward_coins,
        'target', v_target,
        'cap_reached', (v_cap > 0 AND (v_earned + v_actual) >= v_cap)
    );
END;
$$;
REVOKE ALL ON FUNCTION odyssey_submit_auto_task(BIGINT,TEXT,JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_submit_auto_task(BIGINT,TEXT,JSONB) TO service_role;

-- 4. Harden odyssey_verify_submission with Graceful Capping
CREATE OR REPLACE FUNCTION odyssey_verify_submission(
    p_submission_id BIGINT,
    p_admin_uid TEXT,
    p_status TEXT,
    p_admin_notes TEXT DEFAULT NULL,
    p_penalty_coins INT DEFAULT 0
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_sub RECORD; v_task RECORD; v_admin RECORD; v_member RECORD;
    v_reward_coins INT; v_reward_xp INT; v_new_coins INT; v_new_xp INT; v_new_streak INT;
    v_penalty INT; v_actual_penalty INT; v_target INT; v_actual INT;
    v_cap INT; v_earned INT; v_remaining INT;
BEGIN
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid=p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN','GUIDE') THEN RAISE EXCEPTION 'Hanya admin keluarga yang dapat memverifikasi submission' USING ERRCODE='P0003'; END IF;
    SELECT * INTO v_sub FROM odyssey_task_submissions WHERE id=p_submission_id FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'Submission tidak ditemukan' USING ERRCODE='P0002'; END IF;
    IF v_sub.status != 'PENDING' THEN RAISE EXCEPTION 'Submission sudah diproses sebelumnya (status saat ini: %)', v_sub.status USING ERRCODE='P0004'; END IF;
    SELECT * INTO v_member FROM odyssey_user_profiles WHERE uid=v_sub.user_uid FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE='P0007'; END IF;
    IF v_admin.family_id IS NOT NULL AND v_member.family_id IS NOT NULL AND v_admin.family_id != v_member.family_id THEN RAISE EXCEPTION 'Akses ditolak: Submission bukan milik anggota keluarga Anda' USING ERRCODE='P0003'; END IF;
    SELECT * INTO v_task FROM odyssey_tasks WHERE id=v_sub.task_id;
    v_penalty := GREATEST(COALESCE(p_penalty_coins,0),0);

    IF p_status='APPROVED' THEN
        IF v_penalty>0 THEN RAISE EXCEPTION 'Penalti poin tidak dapat diterapkan pada submission yang disetujui' USING ERRCODE='P0005'; END IF;
        IF EXISTS (SELECT 1 FROM odyssey_coin_transactions WHERE type IN ('TASK_REWARD','TASK_PENALTY') AND reference_id=p_submission_id::TEXT) THEN
            RAISE EXCEPTION 'Transaksi untuk submission ini sudah tercatat di ledger' USING ERRCODE='P0004';
        END IF;

        v_target := COALESCE(v_member.monthly_coin_target, COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='default_monthly_coin_target'),0));
        IF v_target IS NULL OR v_target<1 OR v_target>10000 THEN v_target:=0; END IF;
        v_actual := odyssey_calc_target_reward(v_target, v_sub.task_id, v_member.family_id, v_member.uid);
        v_reward_xp := COALESCE(v_task.reward_xp,100);

        -- === AUTHORITATIVE PER-MEMBER EARNING CAP ENFORCEMENT ===
        v_cap := odyssey_get_effective_earning_cap(v_member.uid);
        IF v_cap IS NOT NULL AND v_cap > 0 THEN
            v_earned := odyssey_earned_this_period(v_member.uid);
            v_remaining := GREATEST(0, v_cap - v_earned);
            IF v_actual > v_remaining THEN
                v_actual := v_remaining; -- Gracefully clamp reward to remaining allowance
            END IF;
        END IF;

        UPDATE odyssey_task_submissions
        SET status='APPROVED', admin_notes=p_admin_notes, reviewed_by=p_admin_uid,
            reviewed_at=timezone('utc'::text, now()), coins_earned=v_actual, xp_earned=v_reward_xp
        WHERE id=p_submission_id;

        IF v_actual > 0 THEN
            INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
            VALUES (v_sub.user_uid, v_actual, 'TASK_REWARD', p_submission_id::TEXT, 'Reward: ' || v_task.title);
        END IF;

        UPDATE odyssey_user_profiles
        SET coins=coins+v_actual, xp=COALESCE(xp,0)+v_reward_xp, level=floor(sqrt((COALESCE(xp,0)+v_reward_xp)/100))+1
        WHERE uid=v_sub.user_uid RETURNING coins, xp INTO v_new_coins, v_new_xp;

        v_new_streak := odyssey_update_user_streak(v_sub.user_uid);
        RETURN jsonb_build_object(
            'success', true,
            'status', 'APPROVED',
            'coins_earned', v_actual,
            'xp_earned', v_reward_xp,
            'new_balance', v_new_coins,
            'new_xp', v_new_xp,
            'streak', v_new_streak,
            'base_reward', v_task.reward_coins,
            'target', v_target,
            'cap_reached', (v_cap > 0 AND (v_earned + v_actual) >= v_cap)
        );
    ELSIF p_status='REJECTED' THEN
        v_actual_penalty:=0;
        IF v_penalty>0 THEN
            v_actual_penalty:=LEAST(COALESCE(v_member.coins,0), v_penalty);
            IF v_actual_penalty>0 THEN
                INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
                VALUES (v_sub.user_uid, -v_actual_penalty, 'TASK_PENALTY', p_submission_id::TEXT, 'Penalti tugas: ' || v_task.title);
                UPDATE odyssey_user_profiles SET coins=coins - v_actual_penalty WHERE uid=v_sub.user_uid RETURNING coins INTO v_new_coins;
            ELSE
                v_new_coins:=COALESCE(v_member.coins,0);
            END IF;
        ELSE
            v_new_coins:=COALESCE(v_member.coins,0);
        END IF;

        UPDATE odyssey_task_submissions
        SET status='REJECTED', admin_notes=p_admin_notes, reviewed_by=p_admin_uid,
            reviewed_at=timezone('utc'::text, now()), coins_earned=-v_actual_penalty, xp_earned=0
        WHERE id=p_submission_id;

        RETURN jsonb_build_object('success',true,'status','REJECTED','coins_deducted',v_actual_penalty,'new_balance',v_new_coins);
    ELSE
        RAISE EXCEPTION 'Status tidak valid' USING ERRCODE='P0005';
    END IF;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_verify_submission(BIGINT,TEXT,TEXT,TEXT,INT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_verify_submission(BIGINT,TEXT,TEXT,TEXT,INT) TO service_role;

-- 5. Record migration version
INSERT INTO odyssey_schema_version(key, value)
VALUES ('schema_version', '102_per_user_monthly_earning_cap')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
