-- ============================================================
-- Migration 096: Enforce withdrawal amount in multiples of 500
--
-- CONTEXT: Withdrawal amounts must be a multiple of 500 coins
-- (e.g. 500, 1000, 1500, etc.). If a user has 720 coins, they
-- can withdraw 500 coins. If they have 1250 coins, they can
-- withdraw 500 or 1000 coins.
-- ============================================================

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
    v_min_withdrawal INT;
    v_freq TEXT;
    v_wd INT;
    v_ms INT;
    v_me INT;
    v_src TEXT;
    v_tz TEXT;
    v_now TIMESTAMPTZ;
    v_cur_day INT;
    v_cur_wd INT;
BEGIN
    IF p_coins <= 0 THEN
        RAISE EXCEPTION 'Jumlah koin harus lebih besar dari 0' USING ERRCODE = 'P0010';
    END IF;
    IF trim(COALESCE(p_target_type, '')) = '' OR trim(COALESCE(p_target_value, '')) = '' THEN
        RAISE EXCEPTION 'Target penukaran tidak boleh kosong' USING ERRCODE = 'P0011';
    END IF;
    IF EXISTS (SELECT 1 FROM odyssey_claims WHERE user_uid = p_user_uid AND status = 'PENDING') THEN
        RAISE EXCEPTION 'Anda masih memiliki klaim pending yang belum diproses' USING ERRCODE = 'P0006';
    END IF;

    -- Resolve effective payout config
    SELECT payout_frequency, minimum_withdrawal_coins, payout_weekday, payout_month_start_day, payout_month_end_day, source
      INTO v_freq, v_min_withdrawal, v_wd, v_ms, v_me, v_src
      FROM odyssey_get_effective_payout_config(p_user_uid);

    IF v_min_withdrawal IS NULL OR v_min_withdrawal <= 0 THEN v_min_withdrawal := 500; END IF;

    -- Enforce minimum withdrawal (configurable, not hardcoded)
    IF p_coins < v_min_withdrawal THEN
        RAISE EXCEPTION 'Koin yang diajukan (% ) di bawah minimum penarikan (% ) untuk konfigurasi Anda', p_coins, v_min_withdrawal USING ERRCODE = 'P0014';
    END IF;

    -- Enforce multiple of 500 coins
    IF (p_coins % 500) != 0 THEN
        RAISE EXCEPTION 'Jumlah koin yang diajukan (% ) harus kelipatan 500', p_coins USING ERRCODE = 'P0017';
    END IF;

    -- Enforce schedule based on frequency
    SELECT COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'), 'Asia/Jakarta') INTO v_tz;
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    v_now := timezone(v_tz, now());
    v_cur_day := EXTRACT(DAY FROM v_now)::INT;
    v_cur_wd := EXTRACT(DOW FROM v_now)::INT; -- 0 Sunday

    IF v_freq = 'WEEKLY' THEN
        IF v_wd IS NULL THEN v_wd := 1; END IF;
        IF v_cur_wd != v_wd THEN
            RAISE EXCEPTION 'Penarikan hanya dapat dilakukan pada hari payout mingguan yang dikonfigurasi' USING ERRCODE = 'P0015';
        END IF;
    ELSIF v_freq = 'MONTHLY' THEN
        IF v_ms IS NULL THEN v_ms := 24; END IF;
        IF v_me IS NULL THEN v_me := 26; END IF;
        IF v_cur_day < v_ms OR v_cur_day > v_me THEN
            RAISE EXCEPTION 'Penarikan hanya dapat dilakukan pada periode payout bulanan %–%', v_ms, v_me USING ERRCODE = 'P0015';
        END IF;
    ELSIF v_freq = 'THRESHOLD' THEN
        -- No date restriction; only threshold matters (already checked)
        NULL;
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

    UPDATE odyssey_user_profiles SET coins = coins - p_coins WHERE uid = p_user_uid RETURNING coins INTO v_new_balance;

    RETURN jsonb_build_object('success', true, 'claim_id', v_claim_id, 'new_balance', v_new_balance, 'frequency', v_freq, 'minimum_withdrawal', v_min_withdrawal);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT, BIGINT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT, BIGINT) TO service_role;
DROP FUNCTION IF EXISTS odyssey_create_claim(TEXT, INT, TEXT, TEXT);

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','096_withdrawal_multiple_of_500')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
