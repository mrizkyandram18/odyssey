-- ============================================================
-- Migration 107: Atomic Reward Reservation, Review-Cycle Versioning,
--                Cap Continuity Governance, and Accounting Period Attribution
--
-- Features:
-- 1. Dual-Layer Period & Review Version Schema Expansion:
--    - odyssey_task_submissions:
--        reserved_coins INT NOT NULL DEFAULT 0 CHECK (reserved_coins >= 0)
--        reserved_period_year INT CHECK (reserved_period_year IS NULL OR (reserved_period_year >= 2024 AND reserved_period_year <= 2100))
--        reserved_period_month INT CHECK (reserved_period_month IS NULL OR (reserved_period_month >= 1 AND reserved_period_month <= 12))
--        submission_version INT NOT NULL DEFAULT 1 CHECK (submission_version >= 1)
--    - odyssey_coin_transactions:
--        period_year INT CHECK (period_year IS NULL OR (period_year >= 2024 AND period_year <= 2100))
--        period_month INT CHECK (period_month IS NULL OR (period_month >= 1 AND period_month <= 12))
--
-- 2. Enterprise Invariant CHECK Constraints (NOT VALID -> VALIDATE):
--    - chk_coin_tx_period_pair: period_year and period_month both NULL or both populated.
--    - chk_coin_tx_penalty_ref_format: validates numeric legacy reference or structured ':penalty:v<N>'.
--    - chk_sub_reserve_coherence: positive reservation requires valid period; zero reservation requires NULL period.
--    - chk_sub_status_reserve: non-pending submissions cannot retain active reservations.
--
-- 3. Atomic odyssey_submit_manual_task with Strict PENDING No-Op:
--    - Canonical profile lock order (odyssey_user_profiles FOR UPDATE -> submission FOR UPDATE).
--    - PENDING status retry is a STRICT NO-OP: preserves existing reservation, period, version, and created_at.
--    - Version increment occurs strictly on resubmission after REJECTED.
--    - Full Reservation Policy: reserves task.reward_coins if allowance >= reward, else 0 (Hanya EXP).
--
-- 4. Atomic odyssey_verify_submission with Isolated Ledger Idempotency:
--    - Canonical lock order: profile FOR UPDATE -> submission FOR UPDATE.
--    - Approval idempotency checks strictly for duplicate TASK_REWARD.
--    - Rejection penalty binds deterministically to submission_id and submission_version.
--    - Strict Assertion (P0028): positive settlement without valid reservation period raises exception and rolls back.
--    - Cross-month invariance: settlement preserves exact reservation period, independent of review timestamp.
--
-- 5. Hardened Atomic odyssey_submit_auto_task:
--    - Canonical profile lock order. Clamps to available allowance; records period_year/period_month.
--
-- 6. Privilege Hardening:
--    - Strict REVOKE from PUBLIC, anon, authenticated; GRANT to service_role.
-- ============================================================

-- 1. Schema Expansion
ALTER TABLE odyssey_task_submissions
ADD COLUMN IF NOT EXISTS reserved_coins INT NOT NULL DEFAULT 0 CHECK (reserved_coins >= 0),
ADD COLUMN IF NOT EXISTS reserved_period_year INT CHECK (reserved_period_year IS NULL OR (reserved_period_year >= 2024 AND reserved_period_year <= 2100)),
ADD COLUMN IF NOT EXISTS reserved_period_month INT CHECK (reserved_period_month IS NULL OR (reserved_period_month >= 1 AND reserved_period_month <= 12)),
ADD COLUMN IF NOT EXISTS submission_version INT NOT NULL DEFAULT 1 CHECK (submission_version >= 1);

CREATE INDEX IF NOT EXISTS idx_task_submissions_pending_reserve
ON odyssey_task_submissions (user_uid, status, reserved_period_year, reserved_period_month)
WHERE status = 'PENDING' AND reserved_coins > 0;

ALTER TABLE odyssey_coin_transactions
ADD COLUMN IF NOT EXISTS period_year INT CHECK (period_year IS NULL OR (period_year >= 2024 AND period_year <= 2100)),
ADD COLUMN IF NOT EXISTS period_month INT CHECK (period_month IS NULL OR (period_month >= 1 AND period_month <= 12));

CREATE INDEX IF NOT EXISTS idx_coin_transactions_user_period
ON odyssey_coin_transactions (user_uid, period_year, period_month)
WHERE type = 'TASK_REWARD';

-- Unique Index for Task Reward Idempotency (1 reward per submission ID)
CREATE UNIQUE INDEX IF NOT EXISTS uq_task_reward_ledger_ref
ON odyssey_coin_transactions (reference_id)
WHERE type = 'TASK_REWARD';

-- Unique Index for Versioned Task Penalty Idempotency (1 penalty per submission review cycle)
CREATE UNIQUE INDEX IF NOT EXISTS uq_task_penalty_ledger_ref
ON odyssey_coin_transactions (reference_id)
WHERE type = 'TASK_PENALTY' AND reference_id LIKE '%:penalty:v%';

-- 2. Enterprise Table Constraints (NOT VALID -> VALIDATE CONSTRAINT)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_coin_tx_period_pair'
    ) THEN
        ALTER TABLE odyssey_coin_transactions
        ADD CONSTRAINT chk_coin_tx_period_pair
        CHECK (
            (period_year IS NULL AND period_month IS NULL)
            OR
            (period_year IS NOT NULL AND period_month IS NOT NULL)
        ) NOT VALID;
        ALTER TABLE odyssey_coin_transactions VALIDATE CONSTRAINT chk_coin_tx_period_pair;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_coin_tx_penalty_ref_format'
    ) THEN
        ALTER TABLE odyssey_coin_transactions
        ADD CONSTRAINT chk_coin_tx_penalty_ref_format
        CHECK (
            type != 'TASK_PENALTY'
            OR
            reference_id ~ '^[0-9]+(:penalty:v[0-9]+)?$'
        ) NOT VALID;
        ALTER TABLE odyssey_coin_transactions VALIDATE CONSTRAINT chk_coin_tx_penalty_ref_format;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_sub_reserve_coherence'
    ) THEN
        ALTER TABLE odyssey_task_submissions
        ADD CONSTRAINT chk_sub_reserve_coherence
        CHECK (
            (reserved_coins = 0 AND reserved_period_year IS NULL AND reserved_period_month IS NULL)
            OR
            (reserved_coins > 0 AND reserved_period_year IS NOT NULL AND reserved_period_month IS NOT NULL)
        ) NOT VALID;
        ALTER TABLE odyssey_task_submissions VALIDATE CONSTRAINT chk_sub_reserve_coherence;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_sub_status_reserve'
    ) THEN
        ALTER TABLE odyssey_task_submissions
        ADD CONSTRAINT chk_sub_status_reserve
        CHECK (
            status = 'PENDING'
            OR
            (reserved_coins = 0 AND reserved_period_year IS NULL AND reserved_period_month IS NULL)
        ) NOT VALID;
        ALTER TABLE odyssey_task_submissions VALIDATE CONSTRAINT chk_sub_status_reserve;
    END IF;
END $$;

-- 3. Authoritative Period Active Reservation Resolver
CREATE OR REPLACE FUNCTION odyssey_reserved_this_period(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_ps TIMESTAMPTZ;
    v_pe TIMESTAMPTZ;
    v_year INT;
    v_month INT;
    v_sum INT;
BEGIN
    SELECT period_start_ts, period_end_ts INTO v_ps, v_pe FROM odyssey_earning_cap_period_bounds();
    v_year := EXTRACT(YEAR FROM (v_ps AT TIME ZONE 'Asia/Jakarta'))::INT;
    v_month := EXTRACT(MONTH FROM (v_ps AT TIME ZONE 'Asia/Jakarta'))::INT;

    SELECT COALESCE(SUM(reserved_coins), 0)::INT INTO v_sum
    FROM odyssey_task_submissions
    WHERE user_uid = p_user_uid
      AND status = 'PENDING'
      AND (
          (reserved_period_year IS NOT NULL AND reserved_period_year = v_year AND reserved_period_month = v_month)
          OR
          (reserved_period_year IS NULL AND created_at >= v_ps AND created_at < v_pe)
      );
    RETURN COALESCE(v_sum, 0);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_reserved_this_period(TEXT) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_reserved_this_period(TEXT) TO service_role;

-- 4. Authoritative Period Settled Earning Resolvers
CREATE OR REPLACE FUNCTION odyssey_earned_this_period(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_ps TIMESTAMPTZ;
    v_pe TIMESTAMPTZ;
    v_year INT;
    v_month INT;
    v_sum INT;
BEGIN
    SELECT period_start_ts, period_end_ts INTO v_ps, v_pe FROM odyssey_earning_cap_period_bounds();
    v_year := EXTRACT(YEAR FROM (v_ps AT TIME ZONE 'Asia/Jakarta'))::INT;
    v_month := EXTRACT(MONTH FROM (v_ps AT TIME ZONE 'Asia/Jakarta'))::INT;

    SELECT COALESCE(SUM(amount), 0)::INT INTO v_sum
    FROM odyssey_coin_transactions
    WHERE user_uid = p_user_uid
      AND type = 'TASK_REWARD'
      AND (
          (period_year IS NOT NULL AND period_year = v_year AND period_month = v_month)
          OR
          (period_year IS NULL AND created_at >= v_ps AND created_at < v_pe)
      );
    RETURN COALESCE(v_sum, 0);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_earned_this_period(TEXT) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_earned_this_period(TEXT) TO service_role;

CREATE OR REPLACE FUNCTION odyssey_earned_this_period(p_user_uid TEXT, p_tz TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_ps TIMESTAMPTZ;
    v_pe TIMESTAMPTZ;
    v_year INT;
    v_month INT;
    v_sum INT;
    v_tz TEXT;
BEGIN
    SELECT period_start_ts, period_end_ts INTO v_ps, v_pe FROM odyssey_earning_cap_period_bounds(p_tz);
    v_tz := COALESCE(NULLIF(trim(p_tz), ''), 'Asia/Jakarta');
    v_year := EXTRACT(YEAR FROM (v_ps AT TIME ZONE v_tz))::INT;
    v_month := EXTRACT(MONTH FROM (v_ps AT TIME ZONE v_tz))::INT;

    SELECT COALESCE(SUM(amount), 0)::INT INTO v_sum
    FROM odyssey_coin_transactions
    WHERE user_uid = p_user_uid
      AND type = 'TASK_REWARD'
      AND (
          (period_year IS NOT NULL AND period_year = v_year AND period_month = v_month)
          OR
          (period_year IS NULL AND created_at >= v_ps AND created_at < v_pe)
      );
    RETURN COALESCE(v_sum, 0);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_earned_this_period(TEXT, TEXT) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_earned_this_period(TEXT, TEXT) TO service_role;

-- 5. Authoritative Available Earning Allowance Resolver
CREATE OR REPLACE FUNCTION odyssey_get_available_allowance(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_cap INT;
    v_settled INT;
    v_reserved INT;
BEGIN
    v_cap := odyssey_get_effective_earning_cap(p_user_uid);
    IF v_cap IS NULL OR v_cap <= 0 THEN
        RETURN 0;
    END IF;
    v_settled := odyssey_earned_this_period(p_user_uid);
    v_reserved := odyssey_reserved_this_period(p_user_uid);
    RETURN GREATEST(0, v_cap - v_settled - v_reserved);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_get_available_allowance(TEXT) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_get_available_allowance(TEXT) TO service_role;

-- 6. Hardened Atomic odyssey_submit_manual_task with PENDING Strict No-Op
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
    v_submission RECORD;
    v_submission_id BIGINT;
    v_min_chars INT;
    v_max_chars INT;
    v_text_len INT;
    v_text_content TEXT;
    v_cap INT;
    v_available INT;
    v_reserve_amount INT := 0;
    v_cur_year INT;
    v_cur_month INT;
    v_res_year INT := NULL;
    v_res_month INT := NULL;
    v_next_version INT := 1;
BEGIN
    -- 1. Canonical Root Lock: Lock user profile first
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;
    IF NOT v_profile.is_active THEN
        RAISE EXCEPTION 'Akun Anda diblokir, tidak dapat mengumpulkan tugas' USING ERRCODE = 'P0021';
    END IF;

    -- 2. Verify Task existence & family isolation
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

    -- 3. Lock submission row second
    SELECT * INTO v_submission
    FROM odyssey_task_submissions
    WHERE task_id = p_task_id AND user_uid = p_user_uid
    FOR UPDATE;

    -- Invariant: Resubmission of already approved task is strictly forbidden
    IF v_submission.id IS NOT NULL AND v_submission.status = 'APPROVED' THEN
        RAISE EXCEPTION 'Tugas ini sudah disetujui sebelumnya' USING ERRCODE = 'P0004';
    END IF;

    -- Invariant (P0): PENDING Retry Strict No-Op Guard
    -- If already PENDING, return existing locked reservation without modifying budget or created_at
    IF v_submission.id IS NOT NULL AND v_submission.status = 'PENDING' THEN
        RETURN jsonb_build_object(
            'success', true,
            'submission_id', v_submission.id,
            'status', 'PENDING',
            'reserved_coins', v_submission.reserved_coins,
            'cap_reached', (v_submission.reserved_coins = 0)
        );
    END IF;

    -- 4. Essay validation (TEXT_RESPONSE or VIDEO with essay prompt)
    IF v_task.task_type = 'TEXT_RESPONSE'
       OR (v_task.task_type = 'VIDEO' AND trim(COALESCE(v_task.config->>'prompt', '')) <> '')
    THEN
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

    -- 5. Calculate Review Cycle Version: increment strictly on resubmission after REJECTED
    IF v_submission.id IS NOT NULL AND v_submission.status = 'REJECTED' THEN
        v_next_version := COALESCE(v_submission.submission_version, 1) + 1;
    ELSE
        v_next_version := 1;
    END IF;

    -- 6. Current accounting period in Asia/Jakarta & Available Allowance Calculation
    v_cur_year := EXTRACT(YEAR FROM (timezone('Asia/Jakarta', now())))::INT;
    v_cur_month := EXTRACT(MONTH FROM (timezone('Asia/Jakarta', now())))::INT;
    v_cap := odyssey_get_effective_earning_cap(p_user_uid);
    v_available := odyssey_get_available_allowance(p_user_uid);

    -- Full Reservation Policy
    IF v_available >= COALESCE(v_task.reward_coins, 0) THEN
        v_reserve_amount := COALESCE(v_task.reward_coins, 0);
        v_res_year := v_cur_year;
        v_res_month := v_cur_month;
    ELSE
        v_reserve_amount := 0;
        v_res_year := NULL;
        v_res_month := NULL;
    END IF;

    -- 7. Upsert Submission with locked reservation, period identity, and review version
    INSERT INTO odyssey_task_submissions (
        task_id, user_uid, submission_type, status, payload, created_at, admin_notes,
        reserved_coins, reserved_period_year, reserved_period_month, submission_version
    ) VALUES (
        p_task_id, p_user_uid, 'MANUAL_VERIFY', 'PENDING', p_payload, timezone('utc'::text, now()), NULL,
        v_reserve_amount, v_res_year, v_res_month, v_next_version
    )
    ON CONFLICT (task_id, user_uid) DO UPDATE SET
        payload = p_payload,
        status = 'PENDING',
        created_at = timezone('utc'::text, now()),
        admin_notes = NULL,
        reserved_coins = v_reserve_amount,
        reserved_period_year = v_res_year,
        reserved_period_month = v_res_month,
        submission_version = v_next_version
    RETURNING id INTO v_submission_id;

    RETURN jsonb_build_object(
        'success', true,
        'submission_id', v_submission_id,
        'status', 'PENDING',
        'reserved_coins', v_reserve_amount,
        'cap_reached', (v_cap > 0 AND v_available = 0)
    );
END;
$$;
REVOKE ALL ON FUNCTION odyssey_submit_manual_task(BIGINT, TEXT, JSONB) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_submit_manual_task(BIGINT, TEXT, JSONB) TO service_role;

-- 7. Hardened Atomic odyssey_verify_submission with Strict Settlement & Review Version Penalty
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
    v_sub_uid TEXT;
    v_task RECORD;
    v_admin RECORD;
    v_member RECORD;
    v_reward_xp INT;
    v_new_coins INT;
    v_new_xp INT;
    v_new_streak INT;
    v_penalty INT;
    v_actual_penalty INT;
    v_settle_coins INT;
    v_res_year INT;
    v_res_month INT;
    v_penalty_ref TEXT;
    v_cur_year INT;
    v_cur_month INT;
BEGIN
    -- 1. Verify Admin Profile & Role
    SELECT * INTO v_admin FROM odyssey_user_profiles WHERE uid = p_admin_uid;
    IF NOT FOUND OR v_admin.role NOT IN ('ADMIN', 'GUIDE') THEN
        RAISE EXCEPTION 'Hanya admin keluarga yang dapat memverifikasi submission' USING ERRCODE = 'P0003';
    END IF;

    -- 2. Lookup submission to identify target member UID
    SELECT user_uid INTO v_sub_uid FROM odyssey_task_submissions WHERE id = p_submission_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Submission tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    -- 3. Canonical Root Lock: Lock Member Profile FIRST
    SELECT * INTO v_member FROM odyssey_user_profiles WHERE uid = v_sub_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    -- 4. Lock Submission row SECOND
    SELECT * INTO v_sub FROM odyssey_task_submissions WHERE id = p_submission_id FOR UPDATE;
    IF v_sub.status != 'PENDING' THEN
        RAISE EXCEPTION 'Submission sudah diproses sebelumnya (status saat ini: %)', v_sub.status USING ERRCODE = 'P0004';
    END IF;

    -- 5. Enforce Family Tenant Isolation
    IF v_admin.family_id IS NOT NULL AND v_member.family_id IS NOT NULL AND v_admin.family_id != v_member.family_id THEN
        RAISE EXCEPTION 'Akses ditolak: Submission bukan milik anggota keluarga Anda' USING ERRCODE = 'P0003';
    END IF;

    SELECT * INTO v_task FROM odyssey_tasks WHERE id = v_sub.task_id;
    v_penalty := GREATEST(COALESCE(p_penalty_coins, 0), 0);

    -- ========================================================
    -- CABANG APPROVAL
    -- ========================================================
    IF p_status = 'APPROVED' THEN
        IF v_penalty > 0 THEN
            RAISE EXCEPTION 'Penalti poin tidak dapat diterapkan pada submission yang disetujui' USING ERRCODE = 'P0005';
        END IF;

        -- Idempotency check: strictly check for duplicate TASK_REWARD (unblocks resubmission after penalty)
        IF EXISTS (
            SELECT 1 FROM odyssey_coin_transactions
            WHERE type = 'TASK_REWARD' AND reference_id = p_submission_id::TEXT
        ) THEN
            RAISE EXCEPTION 'Reward untuk submission ini sudah tercatat di ledger' USING ERRCODE = 'P0004';
        END IF;

        v_settle_coins := COALESCE(v_sub.reserved_coins, 0);
        v_reward_xp := COALESCE(v_task.reward_xp, 100);
        v_res_year := v_sub.reserved_period_year;
        v_res_month := v_sub.reserved_period_month;

        -- Strict Assertion (P0028): No silent substitution of timestamps
        IF v_settle_coins > 0 THEN
            IF v_res_year IS NULL OR v_res_month IS NULL
               OR v_res_year < 2024 OR v_res_year > 2100
               OR v_res_month < 1 OR v_res_month > 12
            THEN
                RAISE EXCEPTION 'Integritas periode reservasi tidak valid untuk settlement reward' USING ERRCODE = 'P0028';
            END IF;
        END IF;

        -- Atomically update submission status and clear active reservation
        UPDATE odyssey_task_submissions
        SET status = 'APPROVED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now()),
            coins_earned = v_settle_coins,
            xp_earned = v_reward_xp,
            reserved_coins = 0,
            reserved_period_year = NULL,
            reserved_period_month = NULL
        WHERE id = p_submission_id;

        -- Atomically insert settled ledger entry with bound accounting period
        IF v_settle_coins > 0 THEN
            INSERT INTO odyssey_coin_transactions (
                user_uid, amount, type, reference_id, description, period_year, period_month
            ) VALUES (
                v_sub.user_uid, v_settle_coins, 'TASK_REWARD', p_submission_id::TEXT,
                'Reward: ' || v_task.title, v_res_year, v_res_month
            );
        END IF;

        -- Update member balance and stats
        UPDATE odyssey_user_profiles
        SET coins = coins + v_settle_coins,
            xp = COALESCE(xp, 0) + v_reward_xp,
            level = floor(sqrt((COALESCE(xp, 0) + v_reward_xp) / 100)) + 1
        WHERE uid = v_sub.user_uid
        RETURNING coins, xp INTO v_new_coins, v_new_xp;

        v_new_streak := odyssey_update_user_streak(v_sub.user_uid);

        RETURN jsonb_build_object(
            'success', true,
            'status', 'APPROVED',
            'coins_earned', v_settle_coins,
            'xp_earned', v_reward_xp,
            'new_balance', v_new_coins,
            'new_xp', v_new_xp,
            'streak', v_new_streak,
            'base_reward', v_task.reward_coins,
            'period_year', v_res_year,
            'period_month', v_res_month,
            'cap_reached', (odyssey_get_available_allowance(v_sub.user_uid) = 0)
        );

    -- ========================================================
    -- CABANG REJECTION
    -- ========================================================
    ELSIF p_status = 'REJECTED' THEN
        v_cur_year := EXTRACT(YEAR FROM (timezone('Asia/Jakarta', now())))::INT;
        v_cur_month := EXTRACT(MONTH FROM (timezone('Asia/Jakarta', now())))::INT;
        v_penalty_ref := p_submission_id::TEXT || ':penalty:v' || COALESCE(v_sub.submission_version, 1)::TEXT;
        v_actual_penalty := 0;

        IF v_penalty > 0 THEN
            v_actual_penalty := LEAST(COALESCE(v_member.coins, 0), v_penalty);
            IF v_actual_penalty > 0 THEN
                -- Idempotency check for versioned penalty
                IF EXISTS (
                    SELECT 1 FROM odyssey_coin_transactions
                    WHERE type = 'TASK_PENALTY' AND reference_id = v_penalty_ref
                ) THEN
                    RAISE EXCEPTION 'Penalti untuk siklus review ini sudah tercatat di ledger' USING ERRCODE = 'P0004';
                END IF;

                INSERT INTO odyssey_coin_transactions (
                    user_uid, amount, type, reference_id, description, period_year, period_month
                ) VALUES (
                    v_sub.user_uid, -v_actual_penalty, 'TASK_PENALTY', v_penalty_ref,
                    'Penalti tugas: ' || v_task.title, v_cur_year, v_cur_month
                );

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

        -- Atomically release reservation on rejection
        UPDATE odyssey_task_submissions
        SET status = 'REJECTED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now()),
            coins_earned = -v_actual_penalty,
            xp_earned = 0,
            reserved_coins = 0,
            reserved_period_year = NULL,
            reserved_period_month = NULL
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
REVOKE ALL ON FUNCTION odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) TO service_role;

-- 8. Hardened Atomic odyssey_submit_auto_task with Period Identity
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
    v_cap INT;
    v_available INT;
    v_actual INT;
    v_q_idx INT;
    v_wrong INT[];
    v_cur_year INT;
    v_cur_month INT;
BEGIN
    -- 1. Canonical Root Lock: Lock user profile first
    SELECT * INTO v_profile FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;
    IF NOT v_profile.is_active THEN
        RAISE EXCEPTION 'Akun Anda diblokir, tidak dapat mengumpulkan tugas' USING ERRCODE = 'P0021';
    END IF;

    -- 2. Verify Task existence & family isolation
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
        RAISE EXCEPTION 'Tugas ini sudah diselesaikan dan reward sudah diterima' USING ERRCODE = 'P0004';
    END IF;

    -- 3. Validate quiz questions if present
    v_questions := COALESCE(v_task.config->'questions', v_task.questions, '[]'::jsonb);
    IF jsonb_array_length(v_questions) > 0 THEN
        v_q_idx := 0;
        v_wrong := '{}'::INT[];
        FOR v_q IN SELECT * FROM jsonb_array_elements(v_questions) LOOP
            v_q_idx := v_q_idx + 1;
            v_q_id := v_q->>'id';
            v_correct := trim(COALESCE(v_q->>'correct_answer', ''));
            IF v_correct = '' THEN
                RAISE EXCEPTION 'Soal kuis tidak memiliki kunci jawaban' USING ERRCODE = 'P0009';
            END IF;
            v_user_ans := trim(COALESCE(p_answers->>v_q_id, ''));
            IF v_user_ans = '' OR (
                lower(v_user_ans) != lower(v_correct)
                AND lower(v_user_ans) NOT LIKE lower(v_correct) || '.%'
                AND lower(v_user_ans) NOT LIKE lower(v_correct) || ')%'
                AND lower(v_correct) NOT LIKE lower(v_user_ans) || '.%'
                AND lower(v_correct) NOT LIKE lower(v_user_ans) || ')%'
            ) THEN
                v_wrong := v_wrong || v_q_idx;
            END IF;
        END LOOP;
        IF array_length(v_wrong, 1) > 0 THEN
            RAISE EXCEPTION 'Jawaban kuis belum tepat pada soal nomor %, silakan periksa kembali', array_to_string(v_wrong, ', ') USING ERRCODE = 'P0008';
        END IF;
    END IF;

    -- 4. Validate Mini Game score if present
    IF v_task.task_type = 'MINI_GAME' THEN
        v_game_target := COALESCE((v_task.config->>'target_score')::INT, 0);
        v_game_score := COALESCE((p_answers->>'score')::INT, 0);
        IF v_game_score < 0 OR v_game_score > 1000000 THEN
            RAISE EXCEPTION 'Skor permainan tidak valid' USING ERRCODE = 'P0008';
        END IF;
        IF v_game_target > 0 AND v_game_score < v_game_target THEN
            RAISE EXCEPTION 'Skor permainan belum mencapai target minimum (% vs target %)', v_game_score, v_game_target USING ERRCODE = 'P0008';
        END IF;
    END IF;

    -- 5. Current accounting period in Asia/Jakarta & Available Allowance Clamping
    v_cur_year := EXTRACT(YEAR FROM (timezone('Asia/Jakarta', now())))::INT;
    v_cur_month := EXTRACT(MONTH FROM (timezone('Asia/Jakarta', now())))::INT;
    v_cap := odyssey_get_effective_earning_cap(p_user_uid);
    v_available := odyssey_get_available_allowance(p_user_uid);
    v_reward_xp := COALESCE(v_task.reward_xp, 100);

    IF v_available >= COALESCE(v_task.reward_coins, 0) THEN
        v_actual := COALESCE(v_task.reward_coins, 0);
    ELSE
        -- Gracefully clamp or set to 0 if no allowance left (Task continuity for EXP)
        v_actual := v_available;
    END IF;

    -- 6. Upsert submission
    INSERT INTO odyssey_task_submissions (
        task_id, user_uid, submission_type, status, payload, coins_earned, xp_earned,
        reviewed_at, reserved_coins, reserved_period_year, reserved_period_month, submission_version
    ) VALUES (
        p_task_id, p_user_uid, 'AUTO_QUIZ', 'APPROVED', p_answers, v_actual, v_reward_xp,
        timezone('utc'::text, now()), 0, NULL, NULL, 1
    )
    ON CONFLICT (task_id, user_uid) DO UPDATE SET
        submission_type = 'AUTO_QUIZ',
        status = 'APPROVED',
        payload = p_answers,
        coins_earned = v_actual,
        xp_earned = v_reward_xp,
        reviewed_at = timezone('utc'::text, now()),
        reserved_coins = 0,
        reserved_period_year = NULL,
        reserved_period_month = NULL
    RETURNING id INTO v_submission_id;

    -- 7. Idempotent Ledger Entry with explicit period attribution (Only when v_actual > 0)
    IF v_actual > 0 THEN
        IF NOT EXISTS (
            SELECT 1 FROM odyssey_coin_transactions
            WHERE type = 'TASK_REWARD' AND reference_id = v_submission_id::TEXT
        ) THEN
            INSERT INTO odyssey_coin_transactions (
                user_uid, amount, type, reference_id, description, period_year, period_month
            ) VALUES (
                p_user_uid, v_actual, 'TASK_REWARD', v_submission_id::TEXT,
                'Reward: ' || v_task.title, v_cur_year, v_cur_month
            );
        ELSE
            v_actual := 0;
        END IF;
    END IF;

    -- 8. Update Profile Balance & XP
    UPDATE odyssey_user_profiles
    SET coins = coins + v_actual,
        xp = COALESCE(xp, 0) + v_reward_xp,
        level = floor(sqrt((COALESCE(xp, 0) + v_reward_xp) / 100)) + 1
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
        'period_year', v_cur_year,
        'period_month', v_cur_month,
        'cap_reached', (v_cap > 0 AND odyssey_get_available_allowance(p_user_uid) = 0)
    );
END;
$$;
REVOKE ALL ON FUNCTION odyssey_submit_auto_task(BIGINT, TEXT, JSONB) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION odyssey_submit_auto_task(BIGINT, TEXT, JSONB) TO service_role;

-- 9. Record migration version
INSERT INTO odyssey_schema_version(key, value)
VALUES ('schema_version', '107_atomic_reward_reservation_and_cap_continuity')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
