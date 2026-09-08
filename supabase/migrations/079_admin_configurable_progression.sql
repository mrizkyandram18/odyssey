-- ============================================================
-- Migration 079: Admin-Configurable Progression & Elimination of Hardcoded Business Config
-- Principle: DB is Source of Truth (SOT).
-- Business configuration must never be hardcoded in Go, TS, or SQL runtime.
-- ============================================================

-- 1. Ensure initial SOT rows exist in odyssey_system_config
-- Using ON CONFLICT DO NOTHING so existing production configurations are preserved.
INSERT INTO odyssey_system_config (key, value)
VALUES
    ('default_monthly_earning_cap', '3320'),
    ('max_monthly_earning_cap_ceiling', '10000'),
    ('timezone', 'Asia/Jakarta'),
    ('level_cap_bonus', '{"5":300,"10":500,"20":1000}')
ON CONFLICT (key) DO NOTHING;

-- 2. Optional friendly display name for cosmetic items
ALTER TABLE cosmetic_items ADD COLUMN IF NOT EXISTS name TEXT;

-- 3. Replace odyssey_get_effective_earning_cap:
-- Strictly enforces:
--   - default_monthly_earning_cap from DB SOT. Missing/invalid raises P0026 (never defaults to 0).
--   - 0 = explicit unlimited (preserved).
--   - max_monthly_earning_cap_ceiling from DB SOT (no hardcoded 10000).
--   - level_cap_bonus JSON from DB SOT.
CREATE OR REPLACE FUNCTION odyssey_get_effective_earning_cap(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE
    v_cap INT;
    v_global INT;
    v_ceiling INT;
    v_level INT;
    v_bonus_cfg TEXT;
    v_bonus INT := 0;
    v_effective INT;
BEGIN
    -- 1. Fetch default_monthly_earning_cap from DB SOT
    SELECT NULLIF(trim(value), '')::INT INTO v_global
    FROM odyssey_system_config
    WHERE key = 'default_monthly_earning_cap';

    IF NOT FOUND OR v_global IS NULL OR v_global < 0 THEN
        RAISE EXCEPTION 'default_monthly_earning_cap configuration is missing or invalid' USING ERRCODE = 'P0026';
    END IF;

    -- 2. Fetch max_monthly_earning_cap_ceiling from DB SOT (dynamic ceiling)
    BEGIN
        SELECT NULLIF(trim(value), '')::INT INTO v_ceiling
        FROM odyssey_system_config
        WHERE key = 'max_monthly_earning_cap_ceiling';
    EXCEPTION WHEN OTHERS THEN
        v_ceiling := NULL;
    END;

    -- 3. Fetch user custom cap and level
    SELECT monthly_earning_cap, COALESCE(level, 1)
    INTO v_cap, v_level
    FROM odyssey_user_profiles
    WHERE uid = p_user_uid;

    IF v_cap IS NULL THEN
        v_cap := v_global;
    END IF;

    -- 0 = unlimited (bonus must not cap the uncapped)
    IF v_cap = 0 THEN
        RETURN 0;
    END IF;

    -- 4. Calculate level bonus from DB SOT (level_cap_bonus)
    BEGIN
        SELECT value INTO v_bonus_cfg FROM odyssey_system_config WHERE key = 'level_cap_bonus';
        IF v_bonus_cfg IS NOT NULL AND trim(v_bonus_cfg) <> '' THEN
            SELECT (kv.value)::INT INTO v_bonus
            FROM jsonb_each_text(v_bonus_cfg::jsonb) AS kv(key, value)
            WHERE kv.key ~ '^[0-9]+$' AND kv.key::INT <= v_level
              AND kv.value ~ '^[0-9]+$' AND kv.value::INT >= 0
            ORDER BY kv.key::INT DESC
            LIMIT 1;
        END IF;
    EXCEPTION WHEN OTHERS THEN
        v_bonus := 0;
    END;
    IF v_bonus IS NULL OR v_bonus < 0 THEN v_bonus := 0; END IF;

    v_effective := v_cap + v_bonus;

    -- 5. Clamp to ceiling if configured and > 0
    IF v_ceiling IS NOT NULL AND v_ceiling > 0 AND v_effective > v_ceiling THEN
        v_effective := v_ceiling;
    END IF;

    RETURN v_effective;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_get_effective_earning_cap(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_get_effective_earning_cap(TEXT) TO service_role;

-- 4. Replace odyssey_claim_daily_ticket:
-- Strictly enforces timezone from DB SOT (no hardcoded 'Asia/Jakarta' string literal).
CREATE OR REPLACE FUNCTION odyssey_claim_daily_ticket(p_user_uid TEXT)
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_today DATE;
    v_active BOOLEAN;
    v_granted BOOLEAN := FALSE;
    v_balance INT;
    v_timezone TEXT;
BEGIN
    SELECT trim(value) INTO v_timezone FROM odyssey_system_config WHERE key = 'timezone';
    IF NOT FOUND OR v_timezone IS NULL OR v_timezone = '' THEN
        RAISE EXCEPTION 'timezone configuration is missing' USING ERRCODE = 'P0027';
    END IF;

    BEGIN
        v_today := (timezone(v_timezone, now()))::DATE;
    EXCEPTION WHEN OTHERS THEN
        RAISE EXCEPTION 'timezone configuration is invalid' USING ERRCODE = 'P0027';
    END;

    SELECT is_active INTO v_active FROM odyssey_user_profiles WHERE uid = p_user_uid;
    IF NOT FOUND OR v_active IS FALSE THEN
        RAISE EXCEPTION 'Akun tidak aktif' USING ERRCODE = 'P0021';
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM odyssey_task_submissions s
        JOIN odyssey_tasks t ON t.id = s.task_id
        WHERE s.user_uid = p_user_uid AND s.status = 'APPROVED'
          AND (timezone(v_timezone, COALESCE(s.reviewed_at, s.created_at)))::DATE = v_today
    ) THEN
        RAISE EXCEPTION 'Selesaikan minimal 1 tugas hari ini untuk mendapatkan Tiket Hadiah' USING ERRCODE = 'P0023';
    END IF;

    INSERT INTO reward_tickets (user_uid, ticket_date, status)
    VALUES (p_user_uid, v_today, 'GRANTED')
    ON CONFLICT (user_uid, ticket_date) DO NOTHING
    RETURNING TRUE INTO v_granted;

    IF v_granted IS NULL THEN v_granted := FALSE; END IF;

    SELECT COUNT(*)::INT INTO v_balance FROM reward_tickets WHERE user_uid = p_user_uid AND status = 'GRANTED';

    RETURN jsonb_build_object('granted', v_granted, 'tickets', v_balance);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_claim_daily_ticket(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_claim_daily_ticket(TEXT) TO service_role;

-- 5. Register migration version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '079_admin_configurable_progression')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
