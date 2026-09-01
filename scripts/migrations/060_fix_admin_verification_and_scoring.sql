-- ============================================================
-- Migration 060: Fix Admin Verification Bug & Scoring/Edit Submission
-- 1. Expands odyssey_coin_transactions type check to include TASK_PENALTY
-- 2. Hardens odyssey_verify_submission with ADMIN/GUIDE role check,
--    strict PENDING-only transition guard, and atomic TASK_PENALTY deduction.
-- 3. Adds odyssey_admin_edit_submission for editing PENDING/REJECTED submission payloads.
-- ============================================================

-- 1. Expand ledger type check to include TASK_PENALTY
ALTER TABLE odyssey_coin_transactions DROP CONSTRAINT IF EXISTS odyssey_coin_transactions_type_check;
ALTER TABLE odyssey_coin_transactions ADD CONSTRAINT odyssey_coin_transactions_type_check 
    CHECK (type IN ('TASK_REWARD', 'CLAIM_REDEEM', 'CLAIM_REFUND', 'TASK_PENALTY'));

-- 2. Harden RPC: odyssey_verify_submission
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
    v_sub RECORD;
    v_task RECORD;
    v_admin RECORD;
    v_member RECORD;
    v_reward_coins INT;
    v_reward_xp INT;
    v_new_coins INT;
    v_new_xp INT;
    v_new_streak INT;
    v_penalty INT;
    v_actual_penalty INT;
BEGIN
    -- 1. Verify Admin Profile & Role (supports ADMIN and legacy GUIDE)
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid = p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN', 'GUIDE') THEN
        RAISE EXCEPTION 'Hanya admin keluarga yang dapat memverifikasi submission' USING ERRCODE = 'P0003';
    END IF;

    -- 2. Lock & Validate Submission
    SELECT * INTO v_sub FROM odyssey_task_submissions WHERE id = p_submission_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Submission tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    -- Invariant: Verification is ONLY permitted from PENDING status
    IF v_sub.status != 'PENDING' THEN
        RAISE EXCEPTION 'Submission sudah diproses sebelumnya (status saat ini: %)', v_sub.status USING ERRCODE = 'P0004';
    END IF;

    -- 3. Lock & Validate Member Profile
    SELECT * INTO v_member FROM odyssey_user_profiles WHERE uid = v_sub.user_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    -- 4. Family Tenant Isolation Check
    IF v_admin.family_id IS NOT NULL AND v_member.family_id IS NOT NULL AND v_admin.family_id != v_member.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: Submission bukan milik anggota keluarga Anda' USING ERRCODE = 'P0003';
    END IF;

    -- 5. Fetch Task definition
    SELECT * INTO v_task FROM odyssey_tasks WHERE id = v_sub.task_id;

    v_penalty := GREATEST(COALESCE(p_penalty_coins, 0), 0);

    -- 6. Process State Transition
    IF p_status = 'APPROVED' THEN
        IF v_penalty > 0 THEN
            RAISE EXCEPTION 'Penalti poin tidak dapat diterapkan pada submission yang disetujui' USING ERRCODE = 'P0005';
        END IF;

        v_reward_coins := COALESCE(v_task.reward_coins, 50);
        v_reward_xp := COALESCE(v_task.reward_xp, 100);

        -- Double-ledger protection check
        IF EXISTS (
            SELECT 1 FROM odyssey_coin_transactions 
            WHERE type IN ('TASK_REWARD', 'TASK_PENALTY') AND reference_id = p_submission_id::TEXT
        ) THEN
            RAISE EXCEPTION 'Transaksi untuk submission ini sudah tercatat di ledger' USING ERRCODE = 'P0004';
        END IF;

        UPDATE odyssey_task_submissions
        SET status = 'APPROVED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now()),
            coins_earned = v_reward_coins,
            xp_earned = v_reward_xp
        WHERE id = p_submission_id;

        -- Record Immutable Coin Ledger (Credit)
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
        -- Point deduction if penalty > 0 (strictly clamped to prevent negative balance)
        v_actual_penalty := 0;
        IF v_penalty > 0 THEN
            v_actual_penalty := LEAST(COALESCE(v_member.coins, 0), v_penalty);
            
            IF v_actual_penalty > 0 THEN
                -- Record Immutable Coin Ledger (Debit)
                INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
                VALUES (v_sub.user_uid, -v_actual_penalty, 'TASK_PENALTY', p_submission_id::TEXT, 'Penalti tugas: ' || v_task.title);

                -- Deduct coins from profile projection
                UPDATE odyssey_user_profiles
                SET coins = coins - v_actual_penalty
                WHERE uid = v_sub.user_uid
                RETURNING coins INTO v_new_coins;
            ELSE
                v_new_coins := COALESCE(v_member.coins, 0);
            END IF;
        ELSE
            v_new_coins := COALESCE(v_member.coins, 0);
        END IF;

        UPDATE odyssey_task_submissions
        SET status = 'REJECTED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now()),
            coins_earned = -v_actual_penalty,
            xp_earned = 0
        WHERE id = p_submission_id;

        RETURN jsonb_build_object(
            'success', true,
            'status', 'REJECTED',
            'coins_deducted', v_actual_penalty,
            'new_balance', v_new_coins
        );

    ELSE
        RAISE EXCEPTION 'Status tidak valid' USING ERRCODE = 'P0005';
    END IF;
END;
$$;

REVOKE ALL ON FUNCTION odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) TO service_role;

-- 3. Create RPC: odyssey_admin_edit_submission
-- Allows admin to edit response payload and notes for PENDING or REJECTED submissions
-- Does NOT alter lifecycle status (PENDING stays PENDING, REJECTED stays REJECTED)
-- Rejects editing on APPROVED submissions
CREATE OR REPLACE FUNCTION odyssey_admin_edit_submission(
    p_submission_id BIGINT,
    p_admin_uid TEXT,
    p_payload JSONB,
    p_admin_notes TEXT DEFAULT NULL
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_admin RECORD;
    v_sub RECORD;
    v_member RECORD;
BEGIN
    -- 1. Verify Admin Profile & Role
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid = p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN', 'GUIDE') THEN
        RAISE EXCEPTION 'Hanya admin keluarga yang dapat mengedit submission' USING ERRCODE = 'P0003';
    END IF;

    -- 2. Lock & Validate Submission
    SELECT * INTO v_sub FROM odyssey_task_submissions WHERE id = p_submission_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Submission tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    -- Invariant: APPROVED submissions cannot be edited
    IF v_sub.status = 'APPROVED' THEN
        RAISE EXCEPTION 'Submission yang sudah disetujui tidak dapat diedit' USING ERRCODE = 'P0004';
    END IF;

    -- 3. Lock & Validate Member Profile
    SELECT * INTO v_member FROM odyssey_user_profiles WHERE uid = v_sub.user_uid;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    -- 4. Family Isolation Check
    IF v_admin.family_id IS NOT NULL AND v_member.family_id IS NOT NULL AND v_admin.family_id != v_member.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: Submission bukan milik anggota keluarga Anda' USING ERRCODE = 'P0003';
    END IF;

    -- 5. Update submission payload and notes (preserving current status)
    UPDATE odyssey_task_submissions
    SET payload = p_payload,
        admin_notes = COALESCE(p_admin_notes, admin_notes)
    WHERE id = p_submission_id;

    RETURN jsonb_build_object(
        'success', true,
        'submission_id', p_submission_id,
        'status', v_sub.status,
        'payload', p_payload,
        'admin_notes', COALESCE(p_admin_notes, v_sub.admin_notes)
    );
END;
$$;

REVOKE ALL ON FUNCTION odyssey_admin_edit_submission(BIGINT, TEXT, JSONB, TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_admin_edit_submission(BIGINT, TEXT, JSONB, TEXT) TO service_role;

-- 4. Record migration version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '060_fix_admin_verification_and_scoring')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
