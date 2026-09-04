-- ============================================================
-- Migration 070: Per-User Monthly Earning Cap & Automatic HALT
-- Hard limit per user per earning period (1→24 Asia/Jakarta)
-- Distinct from monthly_coin_target (distribution) and max_payout_coins (withdrawal)
-- Uses odyssey_target_period_bounds() as canonical period source (lazy reset)
-- Source-of-truth: odyssey_coin_transactions.type=TASK_REWARD
-- ============================================================

-- 1. Per-user hard earning cap column (NULL = use global default, 0 = unlimited)
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='odyssey_user_profiles' AND column_name='monthly_earning_cap'
    ) THEN
        ALTER TABLE odyssey_user_profiles
        ADD COLUMN monthly_earning_cap INT CHECK (monthly_earning_cap IS NULL OR monthly_earning_cap >= 0);
    END IF;
END $$;

-- Optional explicit check: 0..10000 when set (matches API validation)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname='chk_monthly_earning_cap_range') THEN
        ALTER TABLE odyssey_user_profiles
        ADD CONSTRAINT chk_monthly_earning_cap_range CHECK (monthly_earning_cap IS NULL OR (monthly_earning_cap BETWEEN 0 AND 10000));
    END IF;
END $$;

-- 2. Global default (separate from monthly_coin_target and max_payout_coins)
INSERT INTO odyssey_system_config(key, value) VALUES ('default_monthly_earning_cap','3320') ON CONFLICT(key) DO NOTHING;

-- 3. Helper: resolve effective earning cap per user
CREATE OR REPLACE FUNCTION odyssey_get_effective_earning_cap(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_cap INT; v_global INT;
BEGIN
    SELECT COALESCE(NULLIF(trim(value),''), '')::INT INTO v_global FROM odyssey_system_config WHERE key='default_monthly_earning_cap';
    IF v_global IS NULL OR v_global < 0 THEN
        v_global := 3320;
    END IF;
    IF v_global > 10000 THEN v_global := 3320; END IF;
    SELECT monthly_earning_cap INTO v_cap FROM odyssey_user_profiles WHERE uid = p_user_uid;
    IF v_cap IS NULL THEN
        RETURN v_global;
    END IF;
    RETURN v_cap; -- 0 = unlimited, >0 = hard cap
END;
$$;
REVOKE ALL ON FUNCTION odyssey_get_effective_earning_cap(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_get_effective_earning_cap(TEXT) TO service_role;

-- 4. Helper: coins earned in current earning period (canonical via odyssey_target_period_bounds)
CREATE OR REPLACE FUNCTION odyssey_earned_this_period(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_ps TIMESTAMPTZ; v_pe TIMESTAMPTZ; v_sum INT;
BEGIN
    SELECT period_start_ts, period_end_ts INTO v_ps, v_pe FROM odyssey_target_period_bounds();
    SELECT COALESCE(SUM(amount),0)::INT INTO v_sum
    FROM odyssey_coin_transactions
    WHERE user_uid = p_user_uid
      AND type = 'TASK_REWARD'
      AND created_at >= v_ps
      AND created_at < v_pe;
    RETURN COALESCE(v_sum,0);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_earned_this_period(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_earned_this_period(TEXT) TO service_role;

-- Optional overload with explicit tz (for future testing)
CREATE OR REPLACE FUNCTION odyssey_earned_this_period(p_user_uid TEXT, p_tz TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_ps TIMESTAMPTZ; v_pe TIMESTAMPTZ; v_sum INT;
BEGIN
    SELECT period_start_ts, period_end_ts INTO v_ps, v_pe FROM odyssey_target_period_bounds(p_tz);
    SELECT COALESCE(SUM(amount),0)::INT INTO v_sum
    FROM odyssey_coin_transactions
    WHERE user_uid = p_user_uid
      AND type = 'TASK_REWARD'
      AND created_at >= v_ps
      AND created_at < v_pe;
    RETURN COALESCE(v_sum,0);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_earned_this_period(TEXT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_earned_this_period(TEXT,TEXT) TO service_role;

-- 5. Helper: is user earning-halted?
CREATE OR REPLACE FUNCTION odyssey_is_earning_halted(p_user_uid TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_cap INT; v_earned INT;
BEGIN
    v_cap := odyssey_get_effective_earning_cap(p_user_uid);
    IF v_cap IS NULL OR v_cap <= 0 THEN
        RETURN false;
    END IF;
    v_earned := odyssey_earned_this_period(p_user_uid);
    RETURN v_earned >= v_cap;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_is_earning_halted(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_is_earning_halted(TEXT) TO service_role;

-- 6. Harden odyssey_submit_auto_task with earning-cap enforcement
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
        FOR v_q IN SELECT * FROM jsonb_array_elements(v_questions) LOOP
            v_q_id := v_q->>'id'; v_correct := trim(COALESCE(v_q->>'correct_answer',''));
            IF v_correct='' THEN RAISE EXCEPTION 'Soal kuis tidak memiliki kunci jawaban' USING ERRCODE='P0009'; END IF;
            v_user_ans := trim(COALESCE(p_answers->>v_q_id,''));
            IF v_user_ans='' OR (lower(v_user_ans)!=lower(v_correct) AND lower(v_user_ans) NOT LIKE lower(v_correct)||'.%' AND lower(v_user_ans) NOT LIKE lower(v_correct)||')%' AND lower(v_correct) NOT LIKE lower(v_user_ans)||'.%' AND lower(v_correct) NOT LIKE lower(v_user_ans)||')%') THEN RAISE EXCEPTION 'Jawaban kuis belum tepat, silakan periksa kembali' USING ERRCODE='P0008'; END IF;
        END LOOP;
    END IF;
    IF v_task.task_type='MINI_GAME' THEN
        v_game_target:=COALESCE((v_task.config->>'target_score')::INT,0); v_game_score:=COALESCE((p_answers->>'score')::INT,0);
        IF v_game_score<0 OR v_game_score>1000000 THEN RAISE EXCEPTION 'Skor permainan tidak valid' USING ERRCODE='P0008'; END IF;
        IF v_game_target>0 AND v_game_score < v_game_target THEN RAISE EXCEPTION 'Skor permainan belum mencapai target minimum (% vs target %)', v_game_score, v_game_target USING ERRCODE='P0008'; END IF;
    END IF;
    -- Resolve target (member-specific, default 3200) and compute scaled reward (user-aware window)
    v_target := COALESCE(v_profile.monthly_coin_target, COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='default_monthly_coin_target'),3200));
    IF v_target IS NULL OR v_target<1 OR v_target>10000 THEN v_target:=3200; END IF;
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

-- 7. Harden odyssey_verify_submission (ADMIN_REVIEW approval) with earning-cap enforcement
DROP FUNCTION IF EXISTS odyssey_verify_submission(BIGINT,TEXT,TEXT,TEXT);
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
DECLARE v_sub RECORD; v_task RECORD; v_admin RECORD; v_member RECORD; v_reward_coins INT; v_reward_xp INT; v_new_coins INT; v_new_xp INT; v_new_streak INT; v_penalty INT; v_actual_penalty INT; v_target INT; v_actual INT; v_cap INT; v_earned INT;
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
        IF EXISTS (SELECT 1 FROM odyssey_coin_transactions WHERE type IN ('TASK_REWARD','TASK_PENALTY') AND reference_id=p_submission_id::TEXT) THEN RAISE EXCEPTION 'Transaksi untuk submission ini sudah tercatat di ledger' USING ERRCODE='P0004'; END IF;
        v_target := COALESCE(v_member.monthly_coin_target, COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='default_monthly_coin_target'),3200));
        IF v_target IS NULL OR v_target<1 OR v_target>10000 THEN v_target:=3200; END IF;
        v_actual := odyssey_calc_target_reward(v_target, v_sub.task_id, v_member.family_id, v_member.uid);
        v_reward_xp := COALESCE(v_task.reward_xp,100);

        -- === EARNING CAP ENFORCEMENT (authoritative, admin path) ===
        v_cap := odyssey_get_effective_earning_cap(v_member.uid);
        IF v_cap IS NOT NULL AND v_cap > 0 THEN
            v_earned := odyssey_earned_this_period(v_member.uid);
            IF v_earned >= v_cap THEN
                RAISE EXCEPTION 'Batas earning bulanan tercapai (%/%). Tidak dapat memberi reward lagi sampai periode berikutnya.', v_earned, v_cap USING ERRCODE='P0016';
            END IF;
            IF v_actual > 0 AND v_earned + v_actual > v_cap THEN
                RAISE EXCEPTION 'Reward (%) akan melebihi batas earning bulanan (%/%, sisa %). Approval ditolak.', v_actual, v_earned, v_cap, (v_cap - v_earned) USING ERRCODE='P0016';
            END IF;
        END IF;

        UPDATE odyssey_task_submissions SET status='APPROVED', admin_notes=p_admin_notes, reviewed_by=p_admin_uid, reviewed_at=timezone('utc'::text, now()), coins_earned=v_actual, xp_earned=v_reward_xp WHERE id=p_submission_id;
        IF v_actual>0 THEN INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description) VALUES (v_sub.user_uid, v_actual, 'TASK_REWARD', p_submission_id::TEXT, 'Reward: ' || v_task.title); END IF;
        UPDATE odyssey_user_profiles SET coins=coins+v_actual, xp=COALESCE(xp,0)+v_reward_xp, level=floor(sqrt((COALESCE(xp,0)+v_reward_xp)/100))+1 WHERE uid=v_sub.user_uid RETURNING coins, xp INTO v_new_coins, v_new_xp;
        v_new_streak := odyssey_update_user_streak(v_sub.user_uid);
        RETURN jsonb_build_object('success',true,'status','APPROVED','coins_earned',v_actual,'xp_earned',v_reward_xp,'new_balance',v_new_coins,'new_xp',v_new_xp,'streak',v_new_streak,'base_reward',v_task.reward_coins,'target',v_target);
    ELSIF p_status='REJECTED' THEN
        v_actual_penalty:=0;
        IF v_penalty>0 THEN
            v_actual_penalty:=LEAST(COALESCE(v_member.coins,0), v_penalty);
            IF v_actual_penalty>0 THEN INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description) VALUES (v_sub.user_uid, -v_actual_penalty, 'TASK_PENALTY', p_submission_id::TEXT, 'Penalti tugas: ' || v_task.title); UPDATE odyssey_user_profiles SET coins=coins - v_actual_penalty WHERE uid=v_sub.user_uid RETURNING coins INTO v_new_coins; ELSE v_new_coins:=COALESCE(v_member.coins,0); END IF;
        ELSE v_new_coins:=COALESCE(v_member.coins,0); END IF;
        UPDATE odyssey_task_submissions SET status='REJECTED', admin_notes=p_admin_notes, reviewed_by=p_admin_uid, reviewed_at=timezone('utc'::text, now()), coins_earned=-v_actual_penalty, xp_earned=0 WHERE id=p_submission_id;
        RETURN jsonb_build_object('success',true,'status','REJECTED','coins_deducted',v_actual_penalty,'new_balance',v_new_coins);
    ELSE RAISE EXCEPTION 'Status tidak valid' USING ERRCODE='P0005'; END IF;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_verify_submission(BIGINT,TEXT,TEXT,TEXT,INT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_verify_submission(BIGINT,TEXT,TEXT,TEXT,INT) TO service_role;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','070_monthly_earning_cap')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
