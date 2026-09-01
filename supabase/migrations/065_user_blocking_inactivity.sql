-- ============================================================
-- Migration 065: User Blocking & Auto-Block Inactivity (Production-Safe)
-- Adds blocking audit fields, configurable threshold, manual block/unblock RPCs,
-- auto-block RPC with calendar-day semantics in Asia/Jakarta, and
-- blocked-user guards in task/reward flows.
-- ============================================================

-- 1. Blocking audit columns (extend existing is_active model)
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='odyssey_user_profiles' AND column_name='blocked_at') THEN
        ALTER TABLE odyssey_user_profiles ADD COLUMN blocked_at TIMESTAMPTZ;
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='odyssey_user_profiles' AND column_name='blocked_by') THEN
        ALTER TABLE odyssey_user_profiles ADD COLUMN blocked_by TEXT REFERENCES odyssey_user_profiles(uid);
    END IF;
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='odyssey_user_profiles' AND column_name='block_reason') THEN
        ALTER TABLE odyssey_user_profiles ADD COLUMN block_reason TEXT;
    END IF;
END $$;

-- Index for scheduler / admin filtering
CREATE INDEX IF NOT EXISTS idx_odyssey_user_profiles_blocked ON odyssey_user_profiles(is_active, blocked_at);
CREATE INDEX IF NOT EXISTS idx_odyssey_user_profiles_role_active ON odyssey_user_profiles(role, is_active);

-- 2. Configurable inactivity threshold
-- Use lower_snake convention (existing keys are lower_snake). Also insert upper variant for spec compatibility.
INSERT INTO odyssey_system_config(key, value) VALUES ('auto_block_inactivity_days','5') ON CONFLICT(key) DO NOTHING;
-- Support reading via either key: keep both in sync (do not overwrite existing custom value)
INSERT INTO odyssey_system_config(key, value) VALUES ('AUTO_BLOCK_INACTIVITY_DAYS','5') ON CONFLICT(key) DO NOTHING;

-- 3. Manual block RPC (idempotent, admin-only, tenant-isolated, audit)
CREATE OR REPLACE FUNCTION odyssey_block_user(
    p_target_uid TEXT,
    p_admin_uid TEXT,
    p_reason TEXT DEFAULT NULL
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_admin RECORD; v_target RECORD;
BEGIN
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid=p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN','GUIDE') THEN
        RAISE EXCEPTION 'Hanya admin dapat memblokir pengguna' USING ERRCODE='P0003';
    END IF;
    SELECT * INTO v_target FROM odyssey_user_profiles WHERE uid=p_target_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Pengguna tidak ditemukan' USING ERRCODE='P0007';
    END IF;
    IF v_admin.family_id IS NOT NULL AND v_target.family_id IS NOT NULL AND v_admin.family_id != v_target.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: bukan anggota keluarga Anda' USING ERRCODE='P0003';
    END IF;
    -- Do not block admin/service accounts
    IF v_target.role IN ('ADMIN','GUIDE','BUILDER') THEN
        RAISE EXCEPTION 'Akun admin tidak dapat diblokir' USING ERRCODE='P0005';
    END IF;
    IF NOT v_target.is_active THEN
        -- Already blocked: idempotent, no duplicate audit event, return current state
        RETURN jsonb_build_object('success',true,'uid',p_target_uid,'already_blocked',true,'is_active',false);
    END IF;
    UPDATE odyssey_user_profiles
    SET is_active=false,
        blocked_at=timezone('utc'::text, now()),
        blocked_by=p_admin_uid,
        block_reason=p_reason,
        updated_at=timezone('utc'::text, now())
    WHERE uid=p_target_uid;
    RETURN jsonb_build_object('success',true,'uid',p_target_uid,'is_active',false,'blocked_at',timezone('utc'::text, now()));
END;
$$;
REVOKE ALL ON FUNCTION odyssey_block_user(TEXT,TEXT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_block_user(TEXT,TEXT,TEXT) TO service_role;

CREATE OR REPLACE FUNCTION odyssey_unblock_user(
    p_target_uid TEXT,
    p_admin_uid TEXT
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_admin RECORD; v_target RECORD;
BEGIN
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid=p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN','GUIDE') THEN
        RAISE EXCEPTION 'Hanya admin dapat membuka blokir' USING ERRCODE='P0003';
    END IF;
    SELECT * INTO v_target FROM odyssey_user_profiles WHERE uid=p_target_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Pengguna tidak ditemukan' USING ERRCODE='P0007';
    END IF;
    IF v_admin.family_id IS NOT NULL AND v_target.family_id IS NOT NULL AND v_admin.family_id != v_target.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: bukan anggota keluarga Anda' USING ERRCODE='P0003';
    END IF;
    IF v_target.is_active THEN
        RETURN jsonb_build_object('success',true,'uid',p_target_uid,'already_active',true,'is_active',true);
    END IF;
    UPDATE odyssey_user_profiles
    SET is_active=true,
        blocked_at=NULL,
        blocked_by=NULL,
        block_reason=NULL,
        updated_at=timezone('utc'::text, now())
    WHERE uid=p_target_uid;
    RETURN jsonb_build_object('success',true,'uid',p_target_uid,'is_active',true);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_unblock_user(TEXT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_unblock_user(TEXT,TEXT) TO service_role;

-- 4. Helper: resolve auto_block threshold (reads both key variants, safe defaults)
CREATE OR REPLACE FUNCTION odyssey_resolve_auto_block_days()
RETURNS INT
LANGUAGE plpgsql STABLE AS $$
DECLARE v_val TEXT; v_n INT;
BEGIN
    SELECT value INTO v_val FROM odyssey_system_config WHERE key='auto_block_inactivity_days';
    IF v_val IS NULL OR trim(v_val)='' THEN
        SELECT value INTO v_val FROM odyssey_system_config WHERE key='AUTO_BLOCK_INACTIVITY_DAYS';
    END IF;
    IF v_val IS NULL OR trim(v_val)='' THEN
        RETURN 5;
    END IF;
    BEGIN
        v_n := trim(v_val)::INT;
    EXCEPTION WHEN OTHERS THEN
        RETURN 5;
    END;
    IF v_n IS NULL OR v_n <=0 OR v_n > 365 THEN
        -- 0 or negative means disabled; out-of-range fallback to default
        IF v_n <=0 THEN RETURN 0; END IF;
        RETURN 5;
    END IF;
    RETURN v_n;
END;
$$;

-- 5. Auto-block RPC: calendar-day semantics in configured timezone
-- Definition: inactive when (today_date - last_success_date) >= threshold
-- last_success_date = max date of APPROVED task submission (COALESCE(reviewed_at, created_at)) converted to tz
-- never completed => NOT blocked, already blocked => skipped, ADMIN => skipped
CREATE OR REPLACE FUNCTION odyssey_auto_block_inactive_users()
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_threshold INT;
    v_tz TEXT;
    v_today DATE;
    v_blocked_count INT := 0;
    v_blocked_uids TEXT[];
BEGIN
    v_threshold := odyssey_resolve_auto_block_days();
    IF v_threshold IS NULL OR v_threshold <=0 THEN
        RETURN jsonb_build_object('success',true,'blocked_count',0,'reason','threshold disabled or invalid','threshold',v_threshold);
    END IF;
    SELECT COALESCE(NULLIF(trim(value),''), 'Asia/Jakarta') INTO v_tz FROM odyssey_system_config WHERE key='timezone';
    IF v_tz IS NULL OR trim(v_tz)='' THEN v_tz:='Asia/Jakarta'; END IF;
    BEGIN
        PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz);
    EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    -- Safe fallback if invalid tz name
    BEGIN
        v_today := (timezone(v_tz, now()))::date;
    EXCEPTION WHEN OTHERS THEN
        v_tz:='Asia/Jakarta';
        v_today := (timezone(v_tz, now()))::date;
    END;

    -- Atomic conditional update: only active MEMBER/SEEKER, not yet blocked, inactive >= threshold
    -- Uses subquery for last_success_date; never-completed users have NULL last date and are NOT blocked.
    WITH last_success AS (
        SELECT user_uid, max((COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date) AS last_date
        FROM odyssey_task_submissions
        WHERE status='APPROVED'
        GROUP BY user_uid
    )
    UPDATE odyssey_user_profiles p
    SET is_active=false,
        blocked_at=timezone('utc'::text, now()),
        blocked_by=NULL,
        block_reason='auto-block: inactivity >= ' || v_threshold || ' days',
        updated_at=timezone('utc'::text, now())
    WHERE p.is_active = true
      AND p.role IN ('MEMBER','SEEKER')
      AND p.blocked_at IS NULL
      AND EXISTS (SELECT 1 FROM last_success ls WHERE ls.user_uid=p.uid AND ls.last_date IS NOT NULL AND (v_today - ls.last_date) >= v_threshold)
    RETURNING p.uid INTO v_blocked_uids;

    -- Count blocked
    -- Re-query because RETURNING into array may not populate correctly with multiple rows in plpgsql; use GET DIAGNOSTICS
    GET DIAGNOSTICS v_blocked_count = ROW_COUNT;

    RETURN jsonb_build_object('success',true,'blocked_count',v_blocked_count,'threshold',v_threshold,'today',v_today::text,'timezone',v_tz);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_auto_block_inactive_users() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_auto_block_inactive_users() TO service_role;

-- Also provide parametrized variant for testing (explicit threshold override)
CREATE OR REPLACE FUNCTION odyssey_auto_block_inactive_users_with_threshold(p_threshold INT, p_tz TEXT DEFAULT NULL)
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_threshold INT; v_tz TEXT; v_today DATE; v_blocked_count INT :=0;
BEGIN
    v_threshold := p_threshold;
    IF v_threshold IS NULL OR v_threshold <=0 OR v_threshold >365 THEN
        v_threshold := odyssey_resolve_auto_block_days();
    END IF;
    IF v_threshold IS NULL OR v_threshold <=0 THEN
        RETURN jsonb_build_object('success',true,'blocked_count',0,'reason','threshold disabled','threshold',v_threshold);
    END IF;
    v_tz := COALESCE(NULLIF(trim(p_tz),''), COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'),'Asia/Jakarta'));
    IF v_tz IS NULL OR trim(v_tz)='' THEN v_tz:='Asia/Jakarta'; END IF;
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    v_today := (timezone(v_tz, now()))::date;
    WITH last_success AS (
        SELECT user_uid, max((COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date) AS last_date
        FROM odyssey_task_submissions WHERE status='APPROVED' GROUP BY user_uid
    )
    UPDATE odyssey_user_profiles p
    SET is_active=false, blocked_at=timezone('utc'::text, now()), blocked_by=NULL, block_reason='auto-block: inactivity >= '||v_threshold||' days', updated_at=timezone('utc'::text, now())
    WHERE p.is_active=true AND p.role IN ('MEMBER','SEEKER') AND p.blocked_at IS NULL
      AND EXISTS (SELECT 1 FROM last_success ls WHERE ls.user_uid=p.uid AND ls.last_date IS NOT NULL AND (v_today - ls.last_date) >= v_threshold);
    GET DIAGNOSTICS v_blocked_count = ROW_COUNT;
    RETURN jsonb_build_object('success',true,'blocked_count',v_blocked_count,'threshold',v_threshold,'today',v_today::text,'timezone',v_tz);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_auto_block_inactive_users_with_threshold(INT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_auto_block_inactive_users_with_threshold(INT,TEXT) TO service_role;

-- 6. Guard existing earning RPCs: blocked user must not generate new rewards
-- Patch odyssey_submit_auto_task (preserve target logic from 061)
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
BEGIN
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid=p_user_uid FOR UPDATE;
    IF NOT FOUND THEN RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE='P0007'; END IF;
    IF NOT v_profile.is_active THEN
        RAISE EXCEPTION 'Akun Anda diblokir, tidak dapat mengerjakan tugas' USING ERRCODE='P0021';
    END IF;
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
    v_target := COALESCE(v_profile.monthly_coin_target, COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='default_monthly_coin_target'),3200));
    IF v_target IS NULL OR v_target<1 OR v_target>10000 THEN v_target:=3200; END IF;
    v_actual := odyssey_calc_target_reward(v_target, p_task_id, v_profile.family_id, p_user_uid);
    v_reward_xp := COALESCE(v_task.reward_xp,100);
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
REVOKE ALL ON FUNCTION odyssey_submit_manual_task(BIGINT,TEXT,JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_submit_manual_task(BIGINT,TEXT,JSONB) TO service_role;

-- Patch odyssey_create_claim to block inactive users
CREATE OR REPLACE FUNCTION odyssey_create_claim(
    p_user_uid TEXT,
    p_coins INT,
    p_target_type TEXT,
    p_target_value TEXT,
    p_reward_id BIGINT DEFAULT NULL
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_current_balance INT;
    v_claim_id BIGINT;
    v_new_balance INT;
    v_max_payout INT;
    v_existing_payout INT;
    v_profile RECORD;
BEGIN
    IF p_coins <= 0 THEN
        RAISE EXCEPTION 'Jumlah koin harus lebih besar dari 0' USING ERRCODE = 'P0010';
    END IF;
    IF trim(COALESCE(p_target_type, '')) = '' OR trim(COALESCE(p_target_value, '')) = '' THEN
        RAISE EXCEPTION 'Target penukaran tidak boleh kosong' USING ERRCODE = 'P0011';
    END IF;
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid=p_user_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE='P0007';
    END IF;
    IF NOT v_profile.is_active THEN
        RAISE EXCEPTION 'Akun Anda diblokir, tidak dapat melakukan penukaran' USING ERRCODE='P0021';
    END IF;
    IF EXISTS (SELECT 1 FROM odyssey_claims WHERE user_uid = p_user_uid AND status = 'PENDING') THEN
        RAISE EXCEPTION 'Anda masih memiliki klaim pending yang belum diproses' USING ERRCODE = 'P0006';
    END IF;
    SELECT COALESCE(NULLIF(value, '')::INT, 3200) INTO v_max_payout
    FROM odyssey_system_config WHERE key = 'max_payout_coins';
    IF v_max_payout IS NULL OR v_max_payout <= 0 THEN
        v_max_payout := 3200;
    END IF;
    SELECT COALESCE(SUM(coins_redeemed), 0) INTO v_existing_payout
    FROM odyssey_claims
    WHERE user_uid = p_user_uid AND status IN ('PENDING','APPROVED');
    IF v_existing_payout + p_coins > v_max_payout THEN
        RAISE EXCEPTION 'Pencairan melebihi batas maksimum periode.' USING ERRCODE = 'P0013';
    END IF;
    SELECT coins INTO v_current_balance FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF v_current_balance IS NULL THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;
    IF v_current_balance < p_coins THEN
        RAISE EXCEPTION 'Saldo koin tidak mencukupi' USING ERRCODE = 'P0003';
    END IF;
    INSERT INTO odyssey_claims (user_uid, coins_redeemed, target_type, target_value, status, reward_id)
    VALUES (p_user_uid, p_coins, p_target_type, p_target_value, 'PENDING', p_reward_id)
    RETURNING id INTO v_claim_id;
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, -p_coins, 'CLAIM_REDEEM', v_claim_id::TEXT, 'Pengajuan penukaran: ' || p_target_type);
    UPDATE odyssey_user_profiles
    SET coins = coins - p_coins
    WHERE uid = p_user_uid
    RETURNING coins INTO v_new_balance;
    RETURN jsonb_build_object(
        'success', true,
        'claim_id', v_claim_id,
        'new_balance', v_new_balance
    );
END;
$$;
REVOKE ALL ON FUNCTION odyssey_create_claim(TEXT,INT,TEXT,TEXT,BIGINT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_create_claim(TEXT,INT,TEXT,TEXT,BIGINT) TO service_role;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','065_user_blocking_inactivity')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
