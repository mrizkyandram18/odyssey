-- ============================================================
-- Migration 094: One-time historical compensation (Sep 25-26 zero-coin bug)
--
-- CONTEXT: before 092, odyssey_target_period_bounds() returned [day 1,
-- day 24+1), so odyssey_calc_target_reward() returned 0 for tasks dated
-- the 25th+. Four Selvi submissions (193,194,195,196) were APPROVED with
-- coins_earned=0 and no TASK_REWARD row (XP/streak still granted).
-- 092 fixed the window prospectively with no backfill. This migration
-- corrects exactly those four rows and nothing else.
--
-- COMPENSATION AMOUNT (verified live via odyssey_calc_target_reward with
-- Selvi's target 3320 under the fixed window):
--   sub 193 (task 560) -> 83
--   sub 194 (task 561) -> 74
--   sub 195 (task 562) -> 74
--   sub 196 (task 563) -> 28
--   TOTAL = 259.
-- NOTE on cap counterfactual: Sept earned before these approvals was 3390
-- vs cap 3600 (headroom 210), so a literal replay would have P0016-blocked
-- the last approval. This mechanism deliberately restores the full
-- corrected entitlement as a historical correction instead of replaying the
-- rate limit: the new HISTORICAL_COMPENSATION type is NOT counted by
-- odyssey_earned_this_period() (TASK_REWARD-only sum), so normal cap math,
-- TASK_REWARD distribution, claims, and future rewards are undistorted.
-- No XP/level/streak changes (already granted historically).
--
-- SAFETY: append-only INSERT + atomic balance update in one RPC body;
-- per-submission idempotency via EXISTS check + partial UNIQUE index
-- backstop (safe under retry/concurrency); fixed grant allowlist (no
-- generic give-coins path); submission rows untouched (status,
-- coins_earned, timestamps preserved); ADMIN/GUIDE-gated.
-- Forward-only: CREATE/ADD only, no data rewrite, no backfill of other rows.
-- ============================================================

-- 1. Extend ledger type check with HISTORICAL_COMPENSATION (060 pattern)
ALTER TABLE odyssey_coin_transactions DROP CONSTRAINT IF EXISTS odyssey_coin_transactions_type_check;
ALTER TABLE odyssey_coin_transactions ADD CONSTRAINT odyssey_coin_transactions_type_check
    CHECK (type IN ('TASK_REWARD', 'CLAIM_REDEEM', 'CLAIM_REFUND', 'TASK_PENALTY', 'HISTORICAL_COMPENSATION'));

-- 2. Fixed one-time grant allowlist (PK = original submission_id)
CREATE TABLE IF NOT EXISTS odyssey_historical_compensation_grants (
    submission_id BIGINT PRIMARY KEY,
    user_uid TEXT NOT NULL REFERENCES odyssey_user_profiles(uid) ON DELETE CASCADE,
    task_id BIGINT NOT NULL REFERENCES odyssey_tasks(id) ON DELETE CASCADE,
    amount INT NOT NULL CHECK (amount > 0),
    reason TEXT NOT NULL
);
ALTER TABLE odyssey_historical_compensation_grants ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on historical_compensation_grants" ON odyssey_historical_compensation_grants;
CREATE POLICY "Allow service_role full access on historical_compensation_grants" ON odyssey_historical_compensation_grants FOR ALL USING (true);
REVOKE ALL ON odyssey_historical_compensation_grants FROM anon, authenticated;

INSERT INTO odyssey_historical_compensation_grants (submission_id, user_uid, task_id, amount, reason) VALUES
    (193, 'usr_1788196798_d46e4385da497091', 560, 83, 'HOTFIX-092-WINDOW'),
    (194, 'usr_1788196798_d46e4385da497091', 561, 74, 'HOTFIX-092-WINDOW'),
    (195, 'usr_1788196798_d46e4385da497091', 562, 74, 'HOTFIX-092-WINDOW'),
    (196, 'usr_1788196798_d46e4385da497091', 563, 28, 'HOTFIX-092-WINDOW')
ON CONFLICT (submission_id) DO NOTHING;

-- 3. DB-level idempotency backstop: at most one compensation row per submission
CREATE UNIQUE INDEX IF NOT EXISTS uq_histcomp_ledger_ref
    ON odyssey_coin_transactions (reference_id)
    WHERE type = 'HISTORICAL_COMPENSATION';

-- 4. Narrowly scoped compensation RPC (no generic give-coins path)
CREATE OR REPLACE FUNCTION odyssey_apply_historical_compensation(
    p_submission_id BIGINT,
    p_user_uid TEXT,
    p_amount INT,
    p_admin_uid TEXT,
    p_reference TEXT DEFAULT 'HOTFIX-092-WINDOW'
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_grant RECORD; v_sub RECORD; v_admin RECORD; v_member RECORD;
    v_task_title TEXT; v_new_coins INT; v_ref TEXT; v_rows INT;
BEGIN
    -- Admin gate (same convention as odyssey_verify_submission)
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid = p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN', 'GUIDE') THEN
        RAISE EXCEPTION 'Hanya admin keluarga yang dapat menerapkan kompensasi historis' USING ERRCODE = 'P0003';
    END IF;

    -- Fixed grant must exist; caller inputs must match it exactly
    SELECT * INTO v_grant FROM odyssey_historical_compensation_grants WHERE submission_id = p_submission_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Submission % tidak termasuk dalam kompensasi historis yang disetujui', p_submission_id USING ERRCODE = 'P0002';
    END IF;
    IF v_grant.user_uid != p_user_uid THEN
        RAISE EXCEPTION 'Kompensasi submission % hanya untuk user yang tercatat', p_submission_id USING ERRCODE = 'P0003';
    END IF;
    IF p_amount IS NULL OR p_amount <= 0 OR p_amount != v_grant.amount THEN
        RAISE EXCEPTION 'Nominal kompensasi tidak sesuai dengan data hotfix yang disetujui' USING ERRCODE = 'P0005';
    END IF;

    -- Submission identity under lock: must be APPROVED, zero-coin, same user+task
    SELECT * INTO v_sub FROM odyssey_task_submissions WHERE id = p_submission_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Submission tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;
    IF v_sub.user_uid != p_user_uid OR v_sub.task_id != v_grant.task_id THEN
        RAISE EXCEPTION 'Identitas submission tidak cocok dengan data hotfix' USING ERRCODE = 'P0003';
    END IF;
    IF v_sub.status != 'APPROVED' THEN
        RAISE EXCEPTION 'Kompensasi hanya untuk submission APPROVED (status saat ini: %)', v_sub.status USING ERRCODE = 'P0004';
    END IF;
    IF COALESCE(v_sub.coins_earned, 0) != 0 THEN
        RAISE EXCEPTION 'Submission ini sudah menerima reward, tidak perlu kompensasi' USING ERRCODE = 'P0005';
    END IF;

    -- Member lock (serializes concurrent compensation for the same user)
    SELECT * INTO v_member FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;
    IF v_admin.family_id IS NOT NULL AND v_member.family_id IS NOT NULL AND v_admin.family_id != v_member.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: user bukan anggota keluarga Anda' USING ERRCODE = 'P0003';
    END IF;

    -- Idempotent fast path: already compensated -> credit 0, report balance
    v_ref := p_submission_id::TEXT;
    IF EXISTS (SELECT 1 FROM odyssey_coin_transactions WHERE type = 'HISTORICAL_COMPENSATION' AND reference_id = v_ref) THEN
        RETURN jsonb_build_object('success', true, 'already_applied', true, 'credited', 0, 'new_balance', v_member.coins, 'submission_id', p_submission_id);
    END IF;

    SELECT title INTO v_task_title FROM odyssey_tasks WHERE id = v_sub.task_id;
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, v_grant.amount, 'HISTORICAL_COMPENSATION', v_ref,
            'Kompensasi historis [' || p_reference || ']: Reward: ' || COALESCE(v_task_title, 'task') || ' (sub ' || v_ref || ')')
    ON CONFLICT (reference_id) WHERE type = 'HISTORICAL_COMPENSATION' DO NOTHING;
    GET DIAGNOSTICS v_rows = ROW_COUNT;
    IF v_rows = 0 THEN
        -- Lost a concurrent race: the winner's row is now visible; credit nothing
        SELECT coins INTO v_new_coins FROM odyssey_user_profiles WHERE uid = p_user_uid;
        RETURN jsonb_build_object('success', true, 'already_applied', true, 'credited', 0, 'new_balance', v_new_coins, 'submission_id', p_submission_id);
    END IF;

    UPDATE odyssey_user_profiles SET coins = coins + v_grant.amount WHERE uid = p_user_uid RETURNING coins INTO v_new_coins;
    RETURN jsonb_build_object('success', true, 'already_applied', false, 'credited', v_grant.amount, 'new_balance', v_new_coins, 'submission_id', p_submission_id);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_apply_historical_compensation(BIGINT, TEXT, INT, TEXT, TEXT) FROM PUBLIC;
REVOKE ALL ON FUNCTION odyssey_apply_historical_compensation(BIGINT, TEXT, INT, TEXT, TEXT) FROM anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_apply_historical_compensation(BIGINT, TEXT, INT, TEXT, TEXT) TO service_role;

INSERT INTO odyssey_schema_version(key, value) VALUES ('schema_version', '094_historical_compensation_sep25_26')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
