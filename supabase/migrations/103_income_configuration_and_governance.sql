-- ============================================================
-- Migration 103: Income Budget Governance, Historical Adjustments, and Safe Account Deactivation
--
-- 1. odyssey_income_status enum & income_status column on odyssey_user_profiles:
--    - 'NOT_CONFIGURED': Newly registered member, effective budget is strictly 0 coins.
--    - 'CONFIGURED': Explicit admin-configured cap, effective budget = base cap - monthly adjustment.
--    - 'DISABLED': Deactivated or tombstoned member, effective budget is strictly 0 coins.
--    - Non-negotiable: 0 or NULL cap NEVER grants unlimited earnings.
--
-- 2. odyssey_member_monthly_adjustments table:
--    - Dedicated table for monthly earning deductions (e.g., historical excess allocation).
--    - Idempotently seeded for Selvi (1,437 coins Nov 2026, 1,437 coins Dec 2026).
--
-- 3. odyssey_cap_audit_logs table:
--    - Tracks all changes to member income status and base cap with admin UID and reason.
--
-- 4. Authoritative odyssey_get_effective_earning_cap:
--    - STABLE function returning 0 for NOT_CONFIGURED, DISABLED, non-members, or cap <= 0.
--    - Clamps base cap to odyssey_system_config ceiling (10,000 coins).
--    - Deducts active monthly adjustment for current calendar month in Asia/Jakarta.
--
-- 5. Hardened reward RPCs (odyssey_submit_auto_task, odyssey_verify_submission):
--    - Serializes via profile SELECT FOR UPDATE.
--    - Clamps reward strictly to remaining allowance (v_actual := 0 when v_cap <= 0).
--    - Never writes 0-amount transactions into odyssey_coin_transactions.
--
-- 6. Canonical account preservation & tombstoning:
--    - Admin ('demo-uid-2') and Selvi ('usr_1788196798_d46e4385da497091') preserved.
--    - Zero-tx smoke accounts physically deleted.
--    - Accounts with immutable ledger txs tombstoned (is_active=false, credentials revoked).
-- ============================================================

-- 1. Create Income Status Type
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'odyssey_income_status') THEN
        CREATE TYPE odyssey_income_status AS ENUM ('NOT_CONFIGURED', 'CONFIGURED', 'DISABLED');
    END IF;
END $$;

-- 2. Add income_status column to odyssey_user_profiles
ALTER TABLE odyssey_user_profiles
ADD COLUMN IF NOT EXISTS income_status odyssey_income_status NOT NULL DEFAULT 'NOT_CONFIGURED';

-- 3. Dedicated Monthly Adjustments Table
CREATE TABLE IF NOT EXISTS odyssey_member_monthly_adjustments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uid TEXT NOT NULL REFERENCES odyssey_user_profiles(uid) ON DELETE CASCADE,
    period_year INT NOT NULL CHECK (period_year >= 2024 AND period_year <= 2100),
    period_month INT NOT NULL CHECK (period_month >= 1 AND period_month <= 12),
    deduction_amount INT NOT NULL CHECK (deduction_amount >= 0),
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now()),
    created_by TEXT NOT NULL,
    CONSTRAINT uq_member_monthly_adjustment UNIQUE (user_uid, period_year, period_month)
);

CREATE INDEX IF NOT EXISTS idx_member_monthly_adjustments_user_period
    ON odyssey_member_monthly_adjustments (user_uid, period_year, period_month);

GRANT SELECT, INSERT, UPDATE, DELETE ON odyssey_member_monthly_adjustments TO service_role;

-- 4. Audit Log Table for Cap and Income Status Changes
CREATE TABLE IF NOT EXISTS odyssey_cap_audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_uid TEXT NOT NULL REFERENCES odyssey_user_profiles(uid) ON DELETE CASCADE,
    changed_by TEXT NOT NULL,
    old_income_status odyssey_income_status,
    new_income_status odyssey_income_status,
    old_cap INT,
    new_cap INT,
    reason TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_cap_audit_logs_user
    ON odyssey_cap_audit_logs (user_uid, created_at DESC);

GRANT SELECT, INSERT, UPDATE ON odyssey_cap_audit_logs TO service_role;

-- 5. Canonical Backfill & Safe Cleanup
-- 5a. Configure Selvi Cahyani (Canonical Active Member)
UPDATE odyssey_user_profiles
SET income_status = 'CONFIGURED',
    monthly_earning_cap = 3320
WHERE uid = 'usr_1788196798_d46e4385da497091';

-- 5b. Configure Admin (Canonical Admin)
UPDATE odyssey_user_profiles
SET income_status = 'DISABLED',
    monthly_earning_cap = NULL
WHERE uid = 'demo-uid-2';

-- 5c. Physically delete smoke accounts that have ZERO transactions in immutable ledger
DELETE FROM odyssey_user_payout_config WHERE user_uid IN ('usr_1788874517_908f79b2b9552f0e', 'smoke-uid-1790453034689');
DELETE FROM odyssey_member_monthly_targets WHERE user_uid IN ('usr_1788874517_908f79b2b9552f0e', 'smoke-uid-1790453034689');
DELETE FROM odyssey_user_profiles WHERE uid IN ('usr_1788874517_908f79b2b9552f0e', 'smoke-uid-1790453034689');

-- 5d. Tombstone Selva and remaining smoke accounts with historical ledger transactions
-- Preserves immutable ledger integrity while permanently revoking access and hiding from all active member queries
UPDATE odyssey_user_profiles
SET is_active = false,
    income_status = 'DISABLED',
    monthly_earning_cap = 0,
    password_hash = NULL,
    username = NULL,
    device_id = NULL,
    device_bound_at = NULL,
    blocked_by = 'demo-uid-2',
    blocked_at = timezone('utc'::text, now()),
    block_reason = 'HISTORICAL_TOMBSTONE_DELETED',
    explorer_name = '[DELETED]'
WHERE uid NOT IN ('demo-uid-2', 'usr_1788196798_d46e4385da497091');

-- 5e. Clean orphaned non-ledger data for tombstoned accounts
DELETE FROM odyssey_push_subscriptions WHERE uid NOT IN ('demo-uid-2', 'usr_1788196798_d46e4385da497091');
DELETE FROM odyssey_user_payout_config WHERE user_uid NOT IN ('demo-uid-2', 'usr_1788196798_d46e4385da497091');
DELETE FROM odyssey_member_monthly_targets WHERE user_uid NOT IN ('demo-uid-2', 'usr_1788196798_d46e4385da497091');

-- 6. Idempotently Seed Selvi Historical Excess Adjustments (Nov/Dec 2026: 1,437 each)
DO $$
DECLARE
    v_conflict RECORD;
BEGIN
    FOR v_conflict IN
        SELECT period_year, period_month, deduction_amount
        FROM odyssey_member_monthly_adjustments
        WHERE user_uid = 'usr_1788196798_d46e4385da497091'
          AND period_year = 2026 AND period_month IN (11, 12)
    LOOP
        IF v_conflict.deduction_amount != 1437 THEN
            RAISE EXCEPTION 'Historical adjustment conflict for %-%: existing % != expected 1437',
                v_conflict.period_year, v_conflict.period_month, v_conflict.deduction_amount;
        END IF;
    END LOOP;

    INSERT INTO odyssey_member_monthly_adjustments (user_uid, period_year, period_month, deduction_amount, reason, created_by)
    VALUES
        ('usr_1788196798_d46e4385da497091', 2026, 11, 1437, 'Oktober 2026 historical excess allocation (1/2)', 'system_policy'),
        ('usr_1788196798_d46e4385da497091', 2026, 12, 1437, 'Oktober 2026 historical excess allocation (2/2)', 'system_policy')
    ON CONFLICT (user_uid, period_year, period_month) DO NOTHING;
END $$;

-- 7. Authoritative Per-Member Effective Cap Resolver
CREATE OR REPLACE FUNCTION odyssey_get_effective_earning_cap(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_role TEXT;
    v_income_status odyssey_income_status;
    v_base_cap INT;
    v_deduction INT := 0;
    v_year INT;
    v_month INT;
    v_ceiling INT := 10000;
    v_effective INT;
BEGIN
    SELECT role, income_status, monthly_earning_cap
    INTO v_role, v_income_status, v_base_cap
    FROM odyssey_user_profiles
    WHERE uid = p_user_uid;

    -- If user does not exist or is not a MEMBER, effective cap is 0
    IF NOT FOUND OR v_role IS NULL OR v_role != 'MEMBER' THEN
        RETURN 0;
    END IF;

    -- Strict business rule: NOT_CONFIGURED or DISABLED members have strictly 0 earning budget
    IF v_income_status IS NULL OR v_income_status != 'CONFIGURED' THEN
        RETURN 0;
    END IF;

    -- If CONFIGURED, but base cap is NULL or <= 0, budget is 0 (0 NEVER means unlimited!)
    IF v_base_cap IS NULL OR v_base_cap <= 0 THEN
        RETURN 0;
    END IF;

    -- Enforce system ceiling
    SELECT COALESCE(NULLIF(trim(value), '')::INT, 10000) INTO v_ceiling
    FROM odyssey_system_config
    WHERE key = 'max_monthly_earning_cap_ceiling';

    IF v_ceiling IS NULL OR v_ceiling <= 0 THEN
        v_ceiling := 10000;
    END IF;

    IF v_base_cap > v_ceiling THEN
        v_base_cap := v_ceiling;
    END IF;

    -- Current Jakarta calendar year and month
    SELECT EXTRACT(YEAR FROM (now() AT TIME ZONE 'Asia/Jakarta'))::INT,
           EXTRACT(MONTH FROM (now() AT TIME ZONE 'Asia/Jakarta'))::INT
    INTO v_year, v_month;

    -- Check for scheduled historical excess adjustments
    v_deduction := 0;
    SELECT COALESCE(deduction_amount, 0) INTO v_deduction
    FROM odyssey_member_monthly_adjustments
    WHERE user_uid = p_user_uid AND period_year = v_year AND period_month = v_month;

    IF v_deduction IS NULL THEN
        v_deduction := 0;
    END IF;

    v_effective := v_base_cap - v_deduction;
    IF v_effective < 0 THEN
        v_effective := 0;
    END IF;

    RETURN v_effective;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_get_effective_earning_cap(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_get_effective_earning_cap(TEXT) TO service_role;

-- 8. Harden odyssey_submit_auto_task with Strict Cap Enforcement & Idempotency
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
    IF v_cap <= 0 THEN
        -- If effective budget is 0 (NOT_CONFIGURED, DISABLED, or cap <= 0), no coin reward is earned
        v_actual := 0;
    ELSE
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

    -- Idempotent ledger insertion (only when v_actual > 0)
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
        'cap_reached', (v_cap > 0 AND (COALESCE(v_earned, 0) + v_actual) >= v_cap)
    );
END;
$$;
REVOKE ALL ON FUNCTION odyssey_submit_auto_task(BIGINT,TEXT,JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_submit_auto_task(BIGINT,TEXT,JSONB) TO service_role;

-- 9. Harden odyssey_verify_submission with Strict Cap Enforcement
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
        IF v_cap <= 0 THEN
            v_actual := 0;
        ELSE
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
            'cap_reached', (v_cap > 0 AND (COALESCE(v_earned, 0) + v_actual) >= v_cap)
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

-- 10. Record migration version
INSERT INTO odyssey_schema_version(key, value)
VALUES ('schema_version', '103_income_configuration_and_governance')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
