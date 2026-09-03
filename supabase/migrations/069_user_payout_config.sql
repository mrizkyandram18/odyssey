-- ============================================================
-- Migration 069: Per-User Configurable Payout & Withdrawal Policy
-- Adds odyssey_user_payout_config with per-user frequency and
-- minimum withdrawal, plus system defaults. No hardcoding of
-- 500, 21-26, weekly/monthly intervals.
-- ============================================================

-- 1. Config table per user
CREATE TABLE IF NOT EXISTS odyssey_user_payout_config (
    user_uid TEXT PRIMARY KEY REFERENCES odyssey_user_profiles(uid) ON DELETE CASCADE,
    payout_frequency TEXT NOT NULL CHECK (payout_frequency IN ('THRESHOLD','WEEKLY','MONTHLY')),
    minimum_withdrawal_coins INT NOT NULL CHECK (minimum_withdrawal_coins > 0 AND minimum_withdrawal_coins <= 100000),
    payout_weekday INT CHECK (payout_weekday IS NULL OR (payout_weekday >= 0 AND payout_weekday <= 6)),
    payout_month_start_day INT CHECK (payout_month_start_day IS NULL OR (payout_month_start_day >= 1 AND payout_month_start_day <= 31)),
    payout_month_end_day INT CHECK (payout_month_end_day IS NULL OR (payout_month_end_day >= 1 AND payout_month_end_day <= 31)),
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    CONSTRAINT chk_payout_weekday_required CHECK (
        payout_frequency != 'WEEKLY' OR payout_weekday IS NOT NULL
    ),
    CONSTRAINT chk_payout_month_required CHECK (
        payout_frequency != 'MONTHLY' OR (payout_month_start_day IS NOT NULL AND payout_month_end_day IS NOT NULL)
    ),
    CONSTRAINT chk_month_window_valid CHECK (
        payout_month_start_day IS NULL OR payout_month_end_day IS NULL OR payout_month_start_day <= payout_month_end_day
    )
);

CREATE INDEX IF NOT EXISTS idx_user_payout_config_frequency ON odyssey_user_payout_config(payout_frequency);

ALTER TABLE odyssey_user_payout_config ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on user_payout_config" ON odyssey_user_payout_config;
CREATE POLICY "Allow service_role full access on user_payout_config" ON odyssey_user_payout_config FOR ALL USING (true);
REVOKE ALL ON odyssey_user_payout_config FROM anon, authenticated;

-- 2. System defaults (data-driven, not hardcoded)
INSERT INTO odyssey_system_config(key, value) VALUES ('default_payout_frequency','MONTHLY') ON CONFLICT(key) DO NOTHING;
INSERT INTO odyssey_system_config(key, value) VALUES ('default_minimum_withdrawal_coins','500') ON CONFLICT(key) DO NOTHING;
INSERT INTO odyssey_system_config(key, value) VALUES ('default_payout_weekday','1') ON CONFLICT(key) DO NOTHING;
INSERT INTO odyssey_system_config(key, value) VALUES ('default_payout_month_start_day','24') ON CONFLICT(key) DO NOTHING;
INSERT INTO odyssey_system_config(key, value) VALUES ('default_payout_month_end_day','26') ON CONFLICT(key) DO NOTHING;

-- Ensure redemption window keys exist as fallback for monthly schedule
INSERT INTO odyssey_system_config(key, value) VALUES ('redemption_start_day','24') ON CONFLICT(key) DO NOTHING;
INSERT INTO odyssey_system_config(key, value) VALUES ('redemption_end_day','26') ON CONFLICT(key) DO NOTHING;

-- 3. Helper: resolve effective payout config for a user
CREATE OR REPLACE FUNCTION odyssey_get_effective_payout_config(p_user_uid TEXT)
RETURNS TABLE (
    payout_frequency TEXT,
    minimum_withdrawal_coins INT,
    payout_weekday INT,
    payout_month_start_day INT,
    payout_month_end_day INT,
    source TEXT
)
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_freq TEXT;
    v_min INT;
    v_wd INT;
    v_ms INT;
    v_me INT;
    v_src TEXT;
    v_row RECORD;
BEGIN
    -- System defaults
    SELECT COALESCE((SELECT value FROM odyssey_system_config WHERE key='default_payout_frequency'), 'MONTHLY') INTO v_freq;
    SELECT COALESCE(NULLIF((SELECT value FROM odyssey_system_config WHERE key='default_minimum_withdrawal_coins'),'')::INT, 500) INTO v_min;
    SELECT COALESCE(NULLIF((SELECT value FROM odyssey_system_config WHERE key='default_payout_weekday'),'')::INT, 1) INTO v_wd;
    -- Monthly window: prefer dedicated defaults, fallback to redemption window
    SELECT COALESCE(
        NULLIF((SELECT value FROM odyssey_system_config WHERE key='default_payout_month_start_day'),'')::INT,
        NULLIF((SELECT value FROM odyssey_system_config WHERE key='redemption_start_day'),'')::INT,
        24) INTO v_ms;
    SELECT COALESCE(
        NULLIF((SELECT value FROM odyssey_system_config WHERE key='default_payout_month_end_day'),'')::INT,
        NULLIF((SELECT value FROM odyssey_system_config WHERE key='redemption_end_day'),'')::INT,
        26) INTO v_me;
    v_src := 'system';

    -- Per-user override if exists and enabled
    SELECT * INTO v_row FROM odyssey_user_payout_config WHERE user_uid = p_user_uid AND enabled = true;
    IF FOUND THEN
        v_freq := v_row.payout_frequency;
        v_min := v_row.minimum_withdrawal_coins;
        IF v_row.payout_weekday IS NOT NULL THEN v_wd := v_row.payout_weekday; END IF;
        IF v_row.payout_month_start_day IS NOT NULL THEN v_ms := v_row.payout_month_start_day; END IF;
        IF v_row.payout_month_end_day IS NOT NULL THEN v_me := v_row.payout_month_end_day; END IF;
        v_src := 'user';
    END IF;

    RETURN QUERY SELECT v_freq, v_min, v_wd, v_ms, v_me, v_src;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_get_effective_payout_config(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_get_effective_payout_config(TEXT) TO service_role;

-- 4. Harden odyssey_create_claim: enforce per-user minimum withdrawal + frequency schedule (threshold check included)
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

    -- Resolve max payout dynamically
    SELECT COALESCE(NULLIF(value, '')::INT, 3200) INTO v_max_payout FROM odyssey_system_config WHERE key = 'max_payout_coins';
    IF v_max_payout IS NULL OR v_max_payout <= 0 THEN v_max_payout := 3200; END IF;
    SELECT COALESCE(SUM(coins_redeemed), 0) INTO v_existing_payout FROM odyssey_claims WHERE user_uid = p_user_uid AND status IN ('PENDING','APPROVED');
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

    UPDATE odyssey_user_profiles SET coins = coins - p_coins WHERE uid = p_user_uid RETURNING coins INTO v_new_balance;

    RETURN jsonb_build_object('success', true, 'claim_id', v_claim_id, 'new_balance', v_new_balance, 'frequency', v_freq, 'minimum_withdrawal', v_min_withdrawal);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT, BIGINT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT, BIGINT) TO service_role;
-- Keep 4-arg overload for backward compat
DROP FUNCTION IF EXISTS odyssey_create_claim(TEXT, INT, TEXT, TEXT);
CREATE OR REPLACE FUNCTION odyssey_create_claim(
    p_user_uid TEXT,
    p_coins INT,
    p_target_type TEXT,
    p_target_value TEXT
) RETURNS JSONB LANGUAGE sql SECURITY DEFINER SET search_path = public AS $$
    SELECT odyssey_create_claim(p_user_uid, p_coins, p_target_type, p_target_value, NULL);
$$;
REVOKE ALL ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_create_claim(TEXT, INT, TEXT, TEXT) TO service_role;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','069_user_payout_config')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
