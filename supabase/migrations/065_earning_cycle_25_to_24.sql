-- ============================================================
-- Migration 065: Rolling earning cycle 25th -> 24th next month
-- Replaces fixed Day 1-24 cycle with rolling [25th, next 25th)
-- Anchor day 25; handles Feb, leap year, Dec->Jan, variable 28-31 days
-- Additive, CREATE OR REPLACE, no historical data mutation
-- Redemption window 21-26 remains independent
-- Target 3200 remains unchanged (decoupled)
-- ============================================================

-- 1. Introduce explicit anchor day for rolling cycle (KISS: explicit > ambiguous start=end=25)
-- DEFERRED CUTOVER: anchor is prepared but function remains 1-24 until 2026-10-25
-- Do NOT activate 25->24 in September; keep Sep 1-24 behavior until cutover
INSERT INTO odyssey_system_config(key, value) VALUES ('earning_cycle_anchor_day','25') ON CONFLICT(key) DO UPDATE SET value='25', updated_at=timezone('utc'::text, now());
INSERT INTO odyssey_system_config(key, value) VALUES ('earning_cycle_cutover_date','2026-10-25') ON CONFLICT(key) DO UPDATE SET value='2026-10-25', updated_at=timezone('utc'::text, now());

-- Keep legacy keys for rollback visibility and for pre-cutover 1-24 behavior
-- target_earning_start_day / target_earning_end_day remain 1/24 until cutover; do NOT update to 25 yet
-- They will be updated to 25 at cutover for documentation, but function gates on cutover date

-- 2. Rolling period bounds: [25th, next 25th) after cutover, [1st, 25th) before cutover
-- Preserves September 1-24 behavior until 2026-10-25 00:00 Asia/Jakarta
CREATE OR REPLACE FUNCTION odyssey_target_period_bounds(p_tz TEXT DEFAULT NULL)
RETURNS TABLE (period_start DATE, period_end DATE, period_start_ts TIMESTAMPTZ, period_end_ts TIMESTAMPTZ)
LANGUAGE plpgsql STABLE AS $$
DECLARE v_tz TEXT; v_anchor INT; v_now_tz TIMESTAMPTZ; v_y INT; v_m INT; v_d INT; v_cutover DATE; v_now_date DATE;
BEGIN
    -- Resolve timezone (same as 061)
    v_tz := COALESCE(NULLIF(trim(p_tz), ''), COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'), 'Asia/Jakarta'));
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;

    v_now_tz := timezone(v_tz, now());
    v_now_date := (v_now_tz)::date;

    -- Cutover gate: before 2026-10-25 use legacy 1-24, after use rolling 25
    v_cutover := COALESCE((SELECT value::DATE FROM odyssey_system_config WHERE key='earning_cycle_cutover_date'), '2026-10-25'::DATE);
    IF v_now_date < v_cutover THEN
        -- Legacy 1-24 behavior (preserves September)
        DECLARE v_start_day INT; v_end_day INT;
        BEGIN
            v_start_day := COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='target_earning_start_day'),1);
            v_end_day := COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='target_earning_end_day'),24);
            IF v_start_day<1 OR v_start_day>31 THEN v_start_day:=1; END IF;
            IF v_end_day<1 OR v_end_day>31 OR v_end_day < v_start_day THEN v_end_day:=24; END IF;
            v_y := EXTRACT(YEAR FROM v_now_tz)::INT;
            v_m := EXTRACT(MONTH FROM v_now_tz)::INT;
            period_start := make_date(v_y, v_m, v_start_day);
            period_end := make_date(v_y, v_m, v_end_day) + interval '1 day';
            period_start_ts := (period_start::text || ' 00:00:00 ' || v_tz)::timestamptz;
            period_end_ts := (period_end::text || ' 00:00:00 ' || v_tz)::timestamptz;
            RETURN NEXT;
            RETURN;
        END;
    END IF;

    -- Post-cutover rolling 25->24
    v_anchor := COALESCE((SELECT value::INT FROM odyssey_system_config WHERE key='earning_cycle_anchor_day'), 25);
    IF v_anchor IS NULL OR v_anchor < 1 OR v_anchor > 31 THEN v_anchor := 25; END IF;

    v_y := EXTRACT(YEAR FROM v_now_tz)::INT;
    v_m := EXTRACT(MONTH FROM v_now_tz)::INT;
    v_d := EXTRACT(DAY FROM v_now_tz)::INT;

    IF v_d >= v_anchor THEN
        period_start := make_date(v_y, v_m, v_anchor);
        IF v_m = 12 THEN
            period_end := make_date(v_y + 1, 1, v_anchor);
        ELSE
            period_end := make_date(v_y, v_m + 1, v_anchor);
        END IF;
    ELSE
        IF v_m = 1 THEN
            period_start := make_date(v_y - 1, 12, v_anchor);
        ELSE
            period_start := make_date(v_y, v_m - 1, v_anchor);
        END IF;
        period_end := make_date(v_y, v_m, v_anchor);
    END IF;

    period_start_ts := (period_start::text || ' 00:00:00 ' || v_tz)::timestamptz;
    period_end_ts := (period_end::text || ' 00:00:00 ' || v_tz)::timestamptz;
    RETURN NEXT;
END; $$;

-- No change to odyssey_calc_target_reward: it already SELECTs from odyssey_target_period_bounds()
-- Join-date logic, USER scope fix (064), floor+remainder remain intact via that call.

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','065_earning_cycle_25_to_24')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
