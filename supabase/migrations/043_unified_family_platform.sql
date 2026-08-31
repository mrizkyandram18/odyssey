-- ============================================================
-- Migration 043: Unified Family Platform Schema & Atomic RPCs
-- ============================================================

-- 1. Upgrade odyssey_tasks to support dynamic steps & polymorphism
DO $$ BEGIN
    -- Add step_order column if not present
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'step_order') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN step_order INT NOT NULL DEFAULT 1;
    END IF;

    -- Add active_date column if not present
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'active_date') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN active_date DATE NOT NULL DEFAULT CURRENT_DATE;
    END IF;

    -- Add reward_xp column if not present
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'reward_xp') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN reward_xp INT NOT NULL DEFAULT 100 CHECK (reward_xp >= 0);
    END IF;

    -- Add config JSONB column if not present
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'config') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN config JSONB NOT NULL DEFAULT '{}'::jsonb;
    END IF;

    -- Drop old task_type constraint and allow expanded types
    ALTER TABLE odyssey_tasks DROP CONSTRAINT IF EXISTS odyssey_tasks_task_type_check;
    ALTER TABLE odyssey_tasks ADD CONSTRAINT odyssey_tasks_task_type_check 
        CHECK (task_type IN ('VIDEO_QUIZ', 'DOCUMENT_UPLOAD', 'PHOTO_PROOF', 'GENERAL', 'YOUTUBE_VIDEO'));
END $$;

CREATE INDEX IF NOT EXISTS idx_odyssey_tasks_active_date ON odyssey_tasks (active_date, step_order) WHERE is_active = TRUE;

-- 2. Create Unified Task Submissions Table (for both auto-quiz and manual verification)
CREATE TABLE IF NOT EXISTS odyssey_task_submissions (
    id              BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    task_id         BIGINT NOT NULL REFERENCES odyssey_tasks(id) ON DELETE CASCADE,
    user_uid        TEXT NOT NULL REFERENCES odyssey_user_profiles(uid),
    submission_type TEXT NOT NULL CHECK (submission_type IN ('AUTO_QUIZ', 'MANUAL_VERIFY')),
    status          TEXT NOT NULL DEFAULT 'PENDING' CHECK (status IN ('PENDING', 'APPROVED', 'REJECTED')),
    payload         JSONB NOT NULL DEFAULT '{}'::jsonb,
    admin_notes     TEXT,
    coins_earned    INT NOT NULL DEFAULT 0,
    xp_earned       INT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    reviewed_at     TIMESTAMPTZ,
    reviewed_by     TEXT REFERENCES odyssey_user_profiles(uid),
    CONSTRAINT uq_user_task_submission UNIQUE (task_id, user_uid)
);

CREATE INDEX IF NOT EXISTS idx_submissions_user_task ON odyssey_task_submissions (user_uid, task_id);
CREATE INDEX IF NOT EXISTS idx_submissions_status ON odyssey_task_submissions (status, created_at DESC);

-- 3. Create Reward Catalog Table
CREATE TABLE IF NOT EXISTS odyssey_reward_catalog (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title        TEXT NOT NULL,
    description  TEXT,
    category     TEXT NOT NULL CHECK (category IN ('PULSA', 'EWALLET', 'CASH', 'SPECIAL')),
    cost_coins   INT NOT NULL CHECK (cost_coins > 0),
    icon_name    TEXT DEFAULT 'gift',
    is_available BOOLEAN NOT NULL DEFAULT TRUE,
    created_at   TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Seed default reward catalog items if empty
INSERT INTO odyssey_reward_catalog (title, description, category, cost_coins, icon_name)
SELECT * FROM (VALUES
    ('Pulsa Rp 10.000', 'Pulsa reguler all operator (Telkomsel, Indosat, XL, Tri)', 'PULSA', 1000, 'smartphone'),
    ('Pulsa Rp 25.000', 'Pulsa reguler all operator', 'PULSA', 2500, 'smartphone'),
    ('Saldo GoPay Rp 20.000', 'Top-up saldo GoPay instan ke nomor terdaftar', 'EWALLET', 2000, 'wallet'),
    ('Saldo DANA Rp 25.000', 'Top-up saldo DANA instan', 'EWALLET', 2500, 'wallet'),
    ('Saldo OVO Rp 50.000', 'Top-up saldo OVO Cash', 'EWALLET', 5000, 'credit-card'),
    ('Uang Saku Tunai Rp 50.000', 'Pencairan uang tunai / transfer bank langsung', 'CASH', 5000, 'banknote')
) AS v(title, description, category, cost_coins, icon_name)
WHERE NOT EXISTS (SELECT 1 FROM odyssey_reward_catalog LIMIT 1);

-- 4. Adjust Claims Table to support reward catalog reference
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_claims' AND column_name = 'reward_id') THEN
        ALTER TABLE odyssey_claims ADD COLUMN reward_id BIGINT REFERENCES odyssey_reward_catalog(id);
    END IF;
END $$;

-- 5. Configure Storage Bucket for Task Proofs (Public for ease of admin review)
INSERT INTO storage.buckets (id, name, public) 
VALUES ('task-proofs', 'task-proofs', true)
ON CONFLICT (id) DO UPDATE SET public = true;

-- 6. Row-Level Security (RLS) Configuration
ALTER TABLE odyssey_task_submissions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on task_submissions" ON odyssey_task_submissions;
CREATE POLICY "Allow service_role full access on task_submissions" ON odyssey_task_submissions FOR ALL USING (true);

ALTER TABLE odyssey_reward_catalog ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on reward_catalog" ON odyssey_reward_catalog;
CREATE POLICY "Allow service_role full access on reward_catalog" ON odyssey_reward_catalog FOR ALL USING (true);

REVOKE ALL ON odyssey_task_submissions FROM anon, authenticated;
REVOKE ALL ON odyssey_reward_catalog FROM anon, authenticated;

-- ============================================================
-- RPC Functions
-- ============================================================

-- Helper: Update User Streak in profile
CREATE OR REPLACE FUNCTION odyssey_update_user_streak(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql
SECURITY DEFINER
AS $$
DECLARE
    v_last_date DATE;
    v_streak INT := 1;
    v_profile RECORD;
BEGIN
    SELECT streak_days, last_active_date INTO v_profile
    FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    
    IF NOT FOUND THEN
        RETURN 1;
    END IF;

    v_last_date := v_profile.last_active_date;
    v_streak := COALESCE(v_profile.streak_days, 0);

    IF v_last_date IS NULL THEN
        v_streak := 1;
    ELSIF v_last_date = CURRENT_DATE THEN
        -- Already active today, streak unchanged
        RETURN v_streak;
    ELSIF v_last_date = CURRENT_DATE - INTERVAL '1 day' THEN
        -- Streak continues
        v_streak := v_streak + 1;
    ELSE
        -- Streak broken
        v_streak := 1;
    END IF;

    UPDATE odyssey_user_profiles
    SET streak_days = v_streak,
        last_active_date = CURRENT_DATE,
        updated_at = timezone('utc'::text, now())
    WHERE uid = p_user_uid;

    RETURN v_streak;
END;
$$;

-- 7. RPC: odyssey_submit_auto_task (Auto-grading for Video & Quiz)
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
    v_reward_coins INT;
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
BEGIN
    -- 1. Validate Task
    SELECT * INTO v_task FROM odyssey_tasks WHERE id = p_task_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Task tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    IF NOT v_task.is_active THEN
        RAISE EXCEPTION 'Task sedang tidak aktif' USING ERRCODE = 'P0001';
    END IF;

    -- Extract questions from config or legacy questions column
    v_questions := COALESCE(v_task.config->'questions', v_task.questions, '[]'::jsonb);

    -- 2. Strict Quiz Validation
    IF jsonb_array_length(v_questions) > 0 THEN
        FOR v_q IN SELECT * FROM jsonb_array_elements(v_questions)
        LOOP
            v_q_id := v_q->>'id';
            v_correct := trim(COALESCE(v_q->>'correct_answer', ''));
            
            IF v_correct = '' THEN
                RAISE EXCEPTION 'Soal kuis tidak memiliki kunci jawaban' USING ERRCODE = 'P0009';
            END IF;

            v_user_ans := trim(COALESCE(p_answers->>v_q_id, ''));
            IF v_user_ans = '' OR lower(v_user_ans) != lower(v_correct) THEN
                RAISE EXCEPTION 'Jawaban kuis belum tepat, silakan coba lagi' USING ERRCODE = 'P0008';
            END IF;
        END LOOP;
    END IF;

    v_reward_coins := COALESCE(v_task.reward_coins, 50);
    v_reward_xp := COALESCE(v_task.reward_xp, 100);

    -- 3. Upsert Submission Record
    INSERT INTO odyssey_task_submissions (
        task_id, user_uid, submission_type, status, payload, coins_earned, xp_earned, reviewed_at
    ) VALUES (
        p_task_id, p_user_uid, 'AUTO_QUIZ', 'APPROVED', p_answers, v_reward_coins, v_reward_xp, timezone('utc'::text, now())
    )
    ON CONFLICT (task_id, user_uid) DO UPDATE SET
        submission_type = 'AUTO_QUIZ',
        status = 'APPROVED',
        payload = p_answers,
        coins_earned = v_reward_coins,
        xp_earned = v_reward_xp,
        reviewed_at = timezone('utc'::text, now())
    RETURNING id INTO v_submission_id;

    -- Also record to legacy completions table for backward compatibility
    INSERT INTO odyssey_task_completions (task_id, user_uid, answers, coins_earned)
    VALUES (p_task_id, p_user_uid, p_answers, v_reward_coins)
    ON CONFLICT (task_id, user_uid) DO NOTHING;

    -- 4. Insert Ledger Transaction (+Coins)
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, v_reward_coins, 'TASK_REWARD', v_submission_id::TEXT, 'Reward: ' || v_task.title);

    -- 5. Update User Profile (Coins & XP)
    UPDATE odyssey_user_profiles
    SET coins = coins + v_reward_coins,
        xp = COALESCE(xp, 0) + v_reward_xp,
        level = floor(sqrt((COALESCE(xp, 0) + v_reward_xp) / 100)) + 1
    WHERE uid = p_user_uid
    RETURNING coins, xp INTO v_new_coins, v_new_xp;

    IF v_new_coins IS NULL THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    -- 6. Update Streak
    v_new_streak := odyssey_update_user_streak(p_user_uid);

    RETURN jsonb_build_object(
        'success', true,
        'submission_id', v_submission_id,
        'coins_earned', v_reward_coins,
        'xp_earned', v_reward_xp,
        'new_balance', v_new_coins,
        'new_xp', v_new_xp,
        'streak', v_new_streak
    );
END;
$$;

REVOKE ALL ON FUNCTION odyssey_submit_auto_task(BIGINT, TEXT, JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_submit_auto_task(BIGINT, TEXT, JSONB) TO service_role;

-- 8. RPC: odyssey_submit_manual_task (Document Upload & Photo Proof)
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
    v_submission_id BIGINT;
BEGIN
    SELECT * INTO v_task FROM odyssey_tasks WHERE id = p_task_id;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Task tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    IF NOT v_task.is_active THEN
        RAISE EXCEPTION 'Task sedang tidak aktif' USING ERRCODE = 'P0001';
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

REVOKE ALL ON FUNCTION odyssey_submit_manual_task(BIGINT, TEXT, JSONB) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_submit_manual_task(BIGINT, TEXT, JSONB) TO service_role;

-- 9. RPC: odyssey_verify_submission (Admin One-Click Review)
CREATE OR REPLACE FUNCTION odyssey_verify_submission(
    p_submission_id BIGINT,
    p_admin_uid TEXT,
    p_status TEXT,
    p_admin_notes TEXT DEFAULT NULL
) RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_sub RECORD;
    v_task RECORD;
    v_reward_coins INT;
    v_reward_xp INT;
    v_new_coins INT;
    v_new_xp INT;
    v_new_streak INT;
BEGIN
    SELECT * INTO v_sub FROM odyssey_task_submissions WHERE id = p_submission_id FOR UPDATE;
    IF NOT FOUND THEN
        RAISE EXCEPTION 'Submission tidak ditemukan' USING ERRCODE = 'P0002';
    END IF;

    IF v_sub.status = 'APPROVED' THEN
        RAISE EXCEPTION 'Submission ini sudah disetujui sebelumnya' USING ERRCODE = 'P0004';
    END IF;

    SELECT * INTO v_task FROM odyssey_tasks WHERE id = v_sub.task_id;
    v_reward_coins := COALESCE(v_task.reward_coins, 50);
    v_reward_xp := COALESCE(v_task.reward_xp, 100);

    IF p_status = 'APPROVED' THEN
        UPDATE odyssey_task_submissions
        SET status = 'APPROVED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now()),
            coins_earned = v_reward_coins,
            xp_earned = v_reward_xp
        WHERE id = p_submission_id;

        -- Record coin ledger
        INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
        VALUES (v_sub.user_uid, v_reward_coins, 'TASK_REWARD', p_submission_id::TEXT, 'Reward: ' || v_task.title);

        -- Award Coins & EXP to profile
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
        UPDATE odyssey_task_submissions
        SET status = 'REJECTED',
            admin_notes = p_admin_notes,
            reviewed_by = p_admin_uid,
            reviewed_at = timezone('utc'::text, now())
        WHERE id = p_submission_id;

        RETURN jsonb_build_object('success', true, 'status', 'REJECTED');
    ELSE
        RAISE EXCEPTION 'Status tidak valid' USING ERRCODE = 'P0005';
    END IF;
END;
$$;

REVOKE ALL ON FUNCTION odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT) TO service_role;

-- 10. RPC: odyssey_create_claim (Updated with reward_id support)
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
BEGIN
    IF p_coins <= 0 THEN
        RAISE EXCEPTION 'Jumlah koin harus lebih besar dari 0' USING ERRCODE = 'P0010';
    END IF;

    IF trim(COALESCE(p_target_type, '')) = '' OR trim(COALESCE(p_target_value, '')) = '' THEN
        RAISE EXCEPTION 'Target penukaran tidak boleh kosong' USING ERRCODE = 'P0011';
    END IF;

    -- Enforce single pending claim per user
    IF EXISTS (SELECT 1 FROM odyssey_claims WHERE user_uid = p_user_uid AND status = 'PENDING') THEN
        RAISE EXCEPTION 'Anda masih memiliki klaim pending yang belum diproses' USING ERRCODE = 'P0006';
    END IF;

    -- Lock profile row & ensure sufficient balance
    SELECT coins INTO v_current_balance FROM odyssey_user_profiles WHERE uid = p_user_uid FOR UPDATE;
    IF v_current_balance IS NULL THEN
        RAISE EXCEPTION 'User profile tidak ditemukan' USING ERRCODE = 'P0007';
    END IF;

    IF v_current_balance < p_coins THEN
        RAISE EXCEPTION 'Saldo koin tidak mencukupi' USING ERRCODE = 'P0003';
    END IF;

    -- Insert Claim
    INSERT INTO odyssey_claims (user_uid, coins_redeemed, target_type, target_value, status, reward_id)
    VALUES (p_user_uid, p_coins, p_target_type, p_target_value, 'PENDING', p_reward_id)
    RETURNING id INTO v_claim_id;

    -- Record deduction in ledger
    INSERT INTO odyssey_coin_transactions (user_uid, amount, type, reference_id, description)
    VALUES (p_user_uid, -p_coins, 'CLAIM_REDEEM', v_claim_id::TEXT, 'Pengajuan penukaran: ' || p_target_type);

    -- Deduct balance
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

REVOKE ALL ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT, BIGINT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT, BIGINT) TO service_role;
