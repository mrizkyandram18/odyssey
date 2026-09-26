-- ============================================================
-- Migration 092: Full-calendar-month target earning window (hotfix)
-- BEFORE: odyssey_target_period_bounds() returned
--   [day 1 00:00, day target_earning_end_day+1 00:00), default end 24,
--   so tasks dated the 25th+ earned v_actual = 0 (APPROVED, 0 coins).
-- AFTER: [first of month 00:00, first of next month 00:00) in the
--   system timezone, so tasks earn throughout the whole calendar month.
-- Contract preserved: same name, signature, return columns, volatility.
-- Sole live caller is odyssey_calc_target_reward(); its signature and
-- largest-remainder logic are untouched. Monthly target stays the pool,
-- monthly cap (P0016), redeem minimum, idempotency guards, and all
-- historical rows are untouched. No backfill. Prospectively effective.
-- NOTE: full-month is NOT implemented via target_earning_end_day = 31;
-- make_date() raises 22008 for invalid dates (e.g. 2026-09-31), which
-- would break every reward calculation in 30-day months and February.
-- ============================================================

CREATE OR REPLACE FUNCTION odyssey_target_period_bounds(p_tz TEXT DEFAULT NULL)
RETURNS TABLE (period_start DATE, period_end DATE, period_start_ts TIMESTAMPTZ, period_end_ts TIMESTAMPTZ)
LANGUAGE plpgsql STABLE AS $$
DECLARE v_tz TEXT; v_now_tz TIMESTAMPTZ;
BEGIN
    v_tz := COALESCE(NULLIF(trim(p_tz), ''), COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'), 'Asia/Jakarta'));
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    v_now_tz := timezone(v_tz, now());
    period_start := date_trunc('month', v_now_tz)::date;
    period_end := (date_trunc('month', v_now_tz) + interval '1 month')::date;
    period_start_ts := (period_start::text || ' 00:00:00 ' || v_tz)::timestamptz;
    period_end_ts := (period_end::text || ' 00:00:00 ' || v_tz)::timestamptz;
    RETURN NEXT;
END; $$;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','092_full_month_earning_window')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
