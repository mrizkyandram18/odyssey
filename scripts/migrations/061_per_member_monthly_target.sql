-- ============================================================
-- Migration 061: Per-Member Monthly Coin Target
-- Adds monthly_coin_target (TARGET, not quota), global defaults,
-- period config 1-24, history table, and target-aware reward calc.
-- ============================================================

-- 1. Per-member target column (NULL = use global default)
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name='odyssey_user_profiles' AND column_name='monthly_coin_target'
    ) THEN
        ALTER TABLE odyssey_user_profiles
        ADD COLUMN monthly_coin_target INT CHECK (monthly_coin_target IS NULL OR (monthly_coin_target BETWEEN 1 AND 10000));
    END IF;
END $$;

-- 2. Global defaults (single source of truth)
INSERT INTO odyssey_system_config(key, value) VALUES ('default_monthly_coin_target','3200') ON CONFLICT(key) DO NOTHING;
INSERT INTO odyssey_system_config(key, value) VALUES ('target_earning_start_day','1') ON CONFLICT(key) DO NOTHING;
INSERT INTO odyssey_system_config(key, value) VALUES ('target_earning_end_day','24') ON CONFLICT(key) DO NOTHING;

-- Backfill existing earning users to 3200 (preserve current 3200 behavior)
UPDATE odyssey_user_profiles SET monthly_coin_target = 3200
WHERE monthly_coin_target IS NULL AND role IN ('MEMBER','SEEKER');

-- 3. History table for audit (what was target in Aug 2026?)
CREATE TABLE IF NOT EXISTS odyssey_member_monthly_targets (
    user_uid TEXT NOT NULL REFERENCES odyssey_user_profiles(uid) ON DELETE CASCADE,
    period_start DATE NOT NULL,
    period_end DATE NOT NULL,
    target INT NOT NULL CHECK (target BETWEEN 1 AND 10000),
    assigned_by TEXT REFERENCES odyssey_user_profiles(uid),
    created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    PRIMARY KEY (user_uid, period_start)
);
CREATE INDEX IF NOT EXISTS idx_member_monthly_targets_period ON odyssey_member_monthly_targets(period_start);
ALTER TABLE odyssey_member_monthly_targets ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on member_monthly_targets" ON odyssey_member_monthly_targets;
CREATE POLICY "Allow service_role full access on member_monthly_targets" ON odyssey_member_monthly_targets FOR ALL USING (true);
REVOKE ALL ON odyssey_member_monthly_targets FROM anon, authenticated;

-- Seed current 1-24 period history for backfilled members
DO $$ DECLARE v_tz TEXT; v_ps DATE; v_pe DATE; BEGIN
    SELECT COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'), 'Asia/Jakarta') INTO v_tz;
    BEGIN PERFORM v_tz::text; PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    v_ps := (date_trunc('month', timezone(v_tz, now())) + (COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='target_earning_start_day'),1)-1) * interval '1 day')::date;
    v_pe := (date_trunc('month', timezone(v_tz, now())) + COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='target_earning_end_day'),24) * interval '1 day')::date;
    INSERT INTO odyssey_member_monthly_targets (user_uid, period_start, period_end, target)
    SELECT uid, v_ps, v_pe, COALESCE(monthly_coin_target,3200) FROM odyssey_user_profiles WHERE role IN ('MEMBER','SEEKER')
    ON CONFLICT (user_uid, period_start) DO NOTHING;
END $$;

-- 4. Helper: target earning period bounds (day 1 through 24 inclusive, timezone-aware)
CREATE OR REPLACE FUNCTION odyssey_target_period_bounds(p_tz TEXT DEFAULT NULL)
RETURNS TABLE (period_start DATE, period_end DATE, period_start_ts TIMESTAMPTZ, period_end_ts TIMESTAMPTZ)
LANGUAGE plpgsql STABLE AS $$
DECLARE v_tz TEXT; v_start_day INT; v_end_day INT; v_now_tz TIMESTAMPTZ; v_y INT; v_m INT;
BEGIN
    v_tz := COALESCE(NULLIF(trim(p_tz), ''), COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'), 'Asia/Jakarta'));
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    v_start_day := COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='target_earning_start_day'),1);
    v_end_day := COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='target_earning_end_day'),24);
    IF v_start_day<1 OR v_start_day>31 THEN v_start_day:=1; END IF;
    IF v_end_day<1 OR v_end_day>31 OR v_end_day < v_start_day THEN v_end_day:=24; END IF;
    v_now_tz := timezone(v_tz, now());
    v_y := EXTRACT(YEAR FROM v_now_tz)::INT;
    v_m := EXTRACT(MONTH FROM v_now_tz)::INT;
    period_start := make_date(v_y, v_m, v_start_day);
    period_end := make_date(v_y, v_m, v_end_day) + interval '1 day';
    period_start_ts := (period_start::text || ' 00:00:00 ' || v_tz)::timestamptz;
    period_end_ts := (period_end::text || ' 00:00:00 ' || v_tz)::timestamptz;
    RETURN NEXT;
END; $$;

-- 5. Helper: deterministic target reward for a task (floor + remainder) — user-aware for mid-month join
CREATE OR REPLACE FUNCTION odyssey_calc_target_reward(
    p_target INT,
    p_task_id BIGINT,
    p_family_id TEXT,
    p_user_uid TEXT DEFAULT NULL
) RETURNS INT
LANGUAGE plpgsql STABLE
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_total_weight INT; v_weight INT; v_period_start DATE; v_period_end DATE; v_actual INT; v_user_join_date DATE;
BEGIN
    IF p_target IS NULL OR p_target<1 OR p_target>10000 THEN p_target:=3200; END IF;
    SELECT period_start, period_end INTO v_period_start, v_period_end FROM odyssey_target_period_bounds();
    -- User-aware start: user cannot earn for tasks before they joined
    IF p_user_uid IS NOT NULL THEN
        SELECT (created_at AT TIME ZONE COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'),'Asia/Jakarta'))::date
          INTO v_user_join_date FROM odyssey_user_profiles WHERE uid=p_user_uid;
        IF v_user_join_date IS NOT NULL AND v_user_join_date > v_period_start THEN
            v_period_start := v_user_join_date;
        END IF;
    END IF;
    -- If join date is after period end, no earning in this period
    IF v_period_start >= v_period_end THEN
        RETURN 0;
    END IF;
    SELECT COALESCE(SUM(reward_coins),0)::INT INTO v_total_weight FROM odyssey_tasks
      WHERE family_id = COALESCE(p_family_id, family_id)
        AND is_active = true AND reward_coins > 0
        AND active_date >= v_period_start AND active_date < v_period_end
        AND (p_family_id IS NULL OR family_id = p_family_id);
    -- No eligible tasks in this user's window => monthly-target reward 0 (do not fallback to base)
    IF v_total_weight IS NULL OR v_total_weight = 0 THEN
        RETURN 0;
    END IF;
    -- Current task weight (for fallback check)
    SELECT COALESCE(reward_coins,50) INTO v_weight FROM odyssey_tasks WHERE id=p_task_id;
    -- If task is outside this user's eligible window or inactive, monthly-target reward is 0
    IF NOT EXISTS (
        SELECT 1 FROM odyssey_tasks WHERE id=p_task_id
          AND is_active = true AND reward_coins > 0
          AND active_date >= v_period_start AND active_date < v_period_end
          AND (p_family_id IS NULL OR family_id = p_family_id)
    ) THEN
        RETURN 0;
    END IF;
    -- Compute deterministic reward with remainder distribution within this user's window
    WITH eligible AS (
        SELECT id, reward_coins, active_date, step_order,
               floor(p_target::numeric * reward_coins::numeric / v_total_weight::numeric)::int AS base,
               (p_target::numeric * reward_coins::numeric / v_total_weight::numeric) - floor(p_target::numeric * reward_coins::numeric / v_total_weight::numeric) AS frac
        FROM odyssey_tasks
        WHERE family_id = COALESCE(p_family_id, family_id)
          AND is_active = true AND reward_coins > 0
          AND active_date >= v_period_start AND active_date < v_period_end
          AND (p_family_id IS NULL OR family_id = p_family_id)
    ), sum_base AS (SELECT COALESCE(SUM(base),0)::int AS s FROM eligible),
    ranked AS (
        SELECT e.id, e.base, e.frac, ROW_NUMBER() OVER (ORDER BY e.frac DESC, e.active_date ASC, e.step_order ASC, e.id ASC) AS rnk
        FROM eligible e
    ), remainder AS (SELECT (p_target - (SELECT s FROM sum_base))::int AS r)
    SELECT base + CASE WHEN rnk <= (SELECT r FROM remainder) THEN 1 ELSE 0 END INTO v_actual
    FROM ranked WHERE id = p_task_id;
    IF v_actual IS NULL THEN RETURN 0; END IF;
    RETURN v_actual;
END; $$;
REVOKE ALL ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT,TEXT) TO service_role;
-- Keep backward-compatible 3-arg overload for any legacy caller (returns 0 for outside-period tasks now)
CREATE OR REPLACE FUNCTION odyssey_calc_target_reward(
    p_target INT,
    p_task_id BIGINT,
    p_family_id TEXT
) RETURNS INT
LANGUAGE sql STABLE SECURITY DEFINER SET search_path = public AS $$
    SELECT odyssey_calc_target_reward(p_target, p_task_id, p_family_id, NULL);
$$;
REVOKE ALL ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT) TO service_role;

-- 6. Upgrade AUTO RPC to use target-scaled reward
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

-- 7. Upgrade MANUAL verify RPC (APPROVED path) to use same target calc
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
DECLARE v_sub RECORD; v_task RECORD; v_admin RECORD; v_member RECORD; v_reward_coins INT; v_reward_xp INT; v_new_coins INT; v_new_xp INT; v_new_streak INT; v_penalty INT; v_actual_penalty INT; v_target INT; v_actual INT;
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

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','061_per_member_monthly_target')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
