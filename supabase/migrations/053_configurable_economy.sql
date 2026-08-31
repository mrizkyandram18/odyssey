-- ============================================================
-- Migration 053: Configurable Economy Source of Truth & Dynamic RPC Cap
-- Default config: 1 Koin = Rp100, Target Rp320.000 (derived 3200 koin), Max 3200 koin
-- Payday: 24, Redemption: 24–26, Earning Period: 30 days, Timezone: Asia/Jakarta
-- ============================================================

-- 1. Seed configurable economy keys in odyssey_system_config
INSERT INTO odyssey_system_config (key, value)
VALUES
    ('coin_conversion_rate', '100'),
    ('payout_target_rupiah', '320000'),
    ('max_payout_coins', '3200'),
    ('earning_period_days', '30'),
    ('payout_day', '24'),
    ('redemption_start_day', '24'),
    ('redemption_end_day', '26'),
    ('timezone', 'Asia/Jakarta')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());

-- Clean up any legacy standalone payout_target_coins key to prevent source-of-truth drift
DELETE FROM odyssey_system_config WHERE key = 'payout_target_coins';

-- 2. Harden odyssey_create_claim with dynamic max_payout_coins cap from odyssey_system_config
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

    -- Resolve max payout dynamically from system config
    SELECT COALESCE(NULLIF(value, '')::INT, 3200) INTO v_max_payout
    FROM odyssey_system_config WHERE key = 'max_payout_coins';
    IF v_max_payout IS NULL OR v_max_payout <= 0 THEN
        v_max_payout := 3200;
    END IF;

    -- Sum of PENDING + APPROVED (REJECTED excluded — refunded)
    SELECT COALESCE(SUM(coins_redeemed), 0) INTO v_existing_payout
    FROM odyssey_claims
    WHERE user_uid = p_user_uid AND status IN ('PENDING','APPROVED');

    IF v_existing_payout + p_coins > v_max_payout THEN
        RAISE EXCEPTION 'Pencairan melebihi batas maksimum periode.' USING ERRCODE = 'P0013';
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

-- Record migration version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '053_configurable_economy')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
