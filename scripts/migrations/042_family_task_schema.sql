-- ============================================================
-- Migration 042: Family Task & Reward App Schema and RPC Functions
-- ============================================================

-- 1. Ensure user_profiles has non-negative coins check
DO $$ BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint WHERE conname = 'chk_coins_non_negative'
    ) THEN
        ALTER TABLE odyssey_user_profiles
        ADD CONSTRAINT chk_coins_non_negative CHECK (coins >= 0);
    END IF;
END $$;

-- 2. Create odyssey_tasks table
CREATE TABLE IF NOT EXISTS odyssey_tasks (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title         TEXT NOT NULL,
    description   TEXT,
    task_type     TEXT NOT NULL DEFAULT 'YOUTUBE_VIDEO' CHECK (task_type IN ('YOUTUBE_VIDEO', 'GENERAL')),
    youtube_url   TEXT,
    questions     JSONB DEFAULT '[]'::jsonb,
    reward_coins  INT NOT NULL DEFAULT 50 CHECK (reward_coins > 0),
    active_from   DATE NOT NULL DEFAULT CURRENT_DATE,
    active_until  DATE NOT NULL DEFAULT CURRENT_DATE,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_by    TEXT REFERENCES odyssey_user_profiles(uid),
    created_at    TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tasks_active_dates ON odyssey_tasks (is_active, active_from, active_until);

-- 3. Create odyssey_task_completions table (Idempotency Enforced)
CREATE TABLE IF NOT EXISTS odyssey_task_completions (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id       BIGINT NOT NULL REFERENCES odyssey_tasks(id) ON DELETE CASCADE,
    user_uid      TEXT NOT NULL REFERENCES odyssey_user_profiles(uid),
    answers       JSONB NOT NULL DEFAULT '{}'::jsonb,
    coins_earned  INT NOT NULL CHECK (coins_earned > 0),
    completed_at  TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    CONSTRAINT uq_task_user UNIQUE (task_id, user_uid)
);

-- 4. Create odyssey_coin_transactions table (Immutable Ledger)
CREATE TABLE IF NOT EXISTS odyssey_coin_transactions (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_uid      TEXT NOT NULL REFERENCES odyssey_user_profiles(uid),
    amount        INT NOT NULL,
    type          TEXT NOT NULL CHECK (type IN ('TASK_REWARD', 'CLAIM_REDEEM', 'CLAIM_REFUND')),
    reference_id  TEXT,
    description   TEXT,
    created_at    TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_transactions_user ON odyssey_coin_transactions (user_uid, created_at DESC);

-- Trigger to prevent UPDATE and DELETE on ledger
CREATE OR REPLACE FUNCTION odyssey_prevent_ledger_mutation()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'odyssey_coin_transactions adalah immutable ledger; UPDATE dan DELETE dilarang' USING ERRCODE = 'P0012';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_odyssey_coin_transactions_immutable ON odyssey_coin_transactions;
CREATE TRIGGER trg_odyssey_coin_transactions_immutable
BEFORE UPDATE OR DELETE ON odyssey_coin_transactions
FOR EACH ROW EXECUTE FUNCTION odyssey_prevent_ledger_mutation();

-- 5. Create odyssey_claims table
CREATE TABLE IF NOT EXISTS odyssey_claims (
    id             BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_uid       TEXT NOT NULL REFERENCES odyssey_user_profiles(uid),
    coins_redeemed INT NOT NULL CHECK (coins_redeemed > 0),
    target_type    TEXT NOT NULL CHECK (target_type IN ('EWALLET', 'BANK', 'PHONE', 'KUOTA')),
    target_value   TEXT NOT NULL,
    status         TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    admin_notes    TEXT,
    created_at     TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    processed_at   TIMESTAMPTZ
);

CREATE UNIQUE INDEX IF NOT EXISTS uq_one_pending_claim_per_user
ON odyssey_claims (user_uid)
WHERE status = 'PENDING';

-- ============================================================
-- Row-Level Security (RLS) & Role Permission Hardening
-- ============================================================
ALTER TABLE odyssey_tasks ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on tasks" ON odyssey_tasks;
CREATE POLICY "Allow service_role full access on tasks" ON odyssey_tasks FOR ALL USING (true);

ALTER TABLE odyssey_task_completions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on task_completions" ON odyssey_task_completions;
CREATE POLICY "Allow service_role full access on task_completions" ON odyssey_task_completions FOR ALL USING (true);

ALTER TABLE odyssey_coin_transactions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on coin_transactions" ON odyssey_coin_transactions;
CREATE POLICY "Allow service_role full access on coin_transactions" ON odyssey_coin_transactions FOR ALL USING (true);

ALTER TABLE odyssey_claims ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on claims" ON odyssey_claims;
CREATE POLICY "Allow service_role full access on claims" ON odyssey_claims FOR ALL USING (true);

-- Revoke direct client mutation permissions from public/anon/authenticated
REVOKE ALL ON odyssey_tasks FROM anon, authenticated;
REVOKE ALL ON odyssey_task_completions FROM anon, authenticated;
REVOKE ALL ON odyssey_coin_transactions FROM anon, authenticated;
REVOKE ALL ON odyssey_claims FROM anon, authenticated;

-- ============================================================
-- RPC Functions (Single Authority & Atomic Context)
-- ============================================================

-- 6. RPC: odyssey_complete_task
CREATE OR REPLACE FUNCTION odyssey_complete_task(
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
    v_reward INT;
    v_completion_id BIGINT;
    v_new_balance INT;
    v_q JSONB;
    v_q_id TEXT;
    v_correct TEXT;
    v_user_ans TEXT;
BEGIN
    -- 1. Validate Task
    SELECT * INTO v_task FROM odyssey_tasks WHERE id = p_task_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Task tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    IF NOT v_task.is_active OR CURRENT_DATE < v_task.active_from OR CURRENT_DATE > v_task.active_until THEN
        RAISE EXCEPTION 'Task sedang tidak aktif atau di luar periode' USING ERRCODE = 'P0001';
    END IF;

    -- 2. Strict Quiz Validation
    IF v_task.questions IS NOT NULL AND jsonb_array_length(v_task.questions) > 0 THEN
        FOR v_q IN SELECT * FROM jsonb_array_elements(v_task.questions)
        LOOP
            v_q_id := v_q->>'id';
            v_correct := trim(COALESCE(v_q->>'correct_answer', ''));
            
            IF v_correct = '' THEN
                RAISE EXCEPTION 'Soal kuis tidak memiliki kunci jawaban yang valid' USING ERRCODE = 'P0009';
            END IF;

            v_user_ans := trim(COALESCE(p_answers->>v_q_id, ''));
            IF v_user_ans = '' OR lower(v_user_ans) != lower(v_correct) THEN
                RAISE EXCEPTION 'Jawaban kuis belum tepat, silakan coba lagi' USING ERRCODE = 'P0008';
            END IF;
        END LOOP;
    END IF;

    v_reward := v_task.reward_coins;

    -- 3. Insert Completion Record (Idempotency authority: UNIQUE constraint catches races)
    INSERT INTO odyssey_task_completions (task_id, user_uid, answers, coins_earned)
    VALUES (p_task_id, p_user_uid, p_answers, v_reward)
    RETURNING id INTO v_completion_id;

    -- 4. Insert Ledger Transaction (+Reward)
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, v_reward, 'TASK_REWARD', v_completion_id::TEXT, 'Reward: ' || v_task.title);

    -- 5. Update Balance Projection & Verify Profile Exists
    UPDATE odyssey_user_profiles
    SET coins = coins + v_reward
    WHERE uid = p_user_uid
    RETURNING coins INTO v_new_balance;

    IF v_new_balance IS NULL THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    RETURN jsonb_build_object(
        'success', true,
        'completion_id', v_completion_id,
        'coins_earned', v_reward,
        'new_balance', v_new_balance
    );
END;
$$;

REVOKE ALL ON FUNCTION odyssey_complete_task(BIGINT, TEXT, JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_complete_task(BIGINT, TEXT, JSONB) TO service_role;

-- 7. RPC: odyssey_create_claim
CREATE OR REPLACE FUNCTION odyssey_create_claim(
    p_user_uid TEXT,
    p_coins INT,
    p_target_type TEXT,
    p_target_value TEXT
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_current_balance INT;
    v_claim_id BIGINT;
    v_new_balance INT;
BEGIN
    -- 1. Input Validation
    IF p_coins <= 0 THEN
        RAISE EXCEPTION 'Jumlah koin harus lebih besar dari 0' USING ERRCODE = 'P0010';
    END IF;

    IF trim(COALESCE(p_target_type, '')) = '' OR trim(COALESCE(p_target_value, '')) = '' THEN
        RAISE EXCEPTION 'Target klaim tidak boleh kosong' USING ERRCODE = 'P0011';
    END IF;

    -- 2. Validate single pending claim per user
    IF EXISTS (SELECT 1 FROM odyssey_claims WHERE user_uid = p_user_uid AND status = 'PENDING') THEN
        RAISE EXCEPTION 'Anda masih memiliki klaim pending yang belum diproses' USING ERRCODE = 'P0006';
    END IF;

    -- 3. Lock profile row & ensure sufficient balance
    SELECT coins INTO v_current_balance FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF v_current_balance IS NULL THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    IF v_current_balance < p_coins THEN
        RAISE EXCEPTION 'Saldo koin tidak mencukupi' USING ERRCODE = 'P0003';
    END IF;

    -- 4. Insert Claim PENDING
    INSERT INTO odyssey_claims (user_uid, coins_redeemed, target_type, target_value, status)
    VALUES (p_user_uid, p_coins, p_target_type, p_target_value, 'PENDING')
    RETURNING id INTO v_claim_id;

    -- 5. Insert Ledger Deduction (-Claim)
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, -p_coins, 'CLAIM_REDEEM', v_claim_id::TEXT, 'Pengajuan klaim: ' || p_target_type);

    -- 6. Deduct Balance Projection
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

REVOKE ALL ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT) TO service_role;

-- 8. RPC: odyssey_process_claim
CREATE OR REPLACE FUNCTION odyssey_process_claim(
    p_claim_id BIGINT,
    p_status TEXT,
    p_admin_notes TEXT DEFAULT NULL
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_claim RECORD;
BEGIN
    SELECT * INTO v_claim FROM odyssey_claims WHERE id = p_claim_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Klaim tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    IF v_claim.status != 'PENDING' THEN
        RAISE EXCEPTION 'Klaim sudah diproses sebelumnya' USING ERRCODE = 'P0004';
    END IF;

    IF p_status = 'APPROVED' THEN
        UPDATE odyssey_claims
        SET status = 'APPROVED', admin_notes = p_admin_notes, processed_at = now()
        WHERE id = p_claim_id;

    ELSIF p_status = 'REJECTED' THEN
        UPDATE odyssey_claims
        SET status = 'REJECTED', admin_notes = p_admin_notes, processed_at = now()
        WHERE id = p_claim_id;

        -- 1. Refund to ledger (+amount)
        INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
        VALUES (v_claim.user_uid, v_claim.coins_redeemed, 'CLAIM_REFUND', p_claim_id::TEXT, 'Pengembalian koin: Klaim ditolak');

        -- 2. Refund balance projection
        UPDATE odyssey_user_profiles
        SET coins = coins + v_claim.coins_redeemed
        WHERE uid = v_claim.user_uid;
    ELSE
        RAISE EXCEPTION 'Status tidak valid' USING ERRCODE = 'P0005';
    END IF;

    RETURN jsonb_build_object('success', true, 'status', p_status);
END;
$$;

REVOKE ALL ON FUNCTION odyssey_process_claim(BIGINT, TEXT, TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_process_claim(BIGINT, TEXT, TEXT) TO service_role;
