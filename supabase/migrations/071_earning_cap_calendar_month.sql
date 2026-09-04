-- ============================================================
-- Migration 071: Earning cap calendar-month window
-- Migrates earning-cap period from target_earning 1-24 to
-- true calendar month (1 → last day, Asia/Jakarta).
-- Keeps odyssey_target_period_bounds for reward distribution (1-24)
-- and odyssey_member_monthly_targets history as-is.
-- Source of Truth: odyssey_earned_this_period / odyssey_is_earning_halted
-- ============================================================

-- Helper: calendar-month bounds (reuse for cap)
CREATE OR REPLACE FUNCTION odyssey_earning_cap_period_bounds(p_tz TEXT DEFAULT NULL)
RETURNS TABLE (period_start_ts TIMESTAMPTZ, period_end_ts TIMESTAMPTZ)
LANGUAGE plpgsql STABLE AS $$
DECLARE v_tz TEXT; v_now_tz TIMESTAMPTZ; v_ps DATE; v_pe DATE;
BEGIN
    v_tz := COALESCE(NULLIF(trim(p_tz), ''), COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'), 'Asia/Jakarta'));
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    v_now_tz := timezone(v_tz, now());
    v_ps := date_trunc('month', v_now_tz)::date;
    v_pe := (date_trunc('month', v_now_tz) + interval '1 month')::date;
    period_start_ts := (v_ps::text || ' 00:00:00 ' || v_tz)::timestamptz;
    period_end_ts := (v_pe::text || ' 00:00:00 ' || v_tz)::timestamptz;
    RETURN NEXT;
END; $$;

REVOKE ALL ON FUNCTION odyssey_earning_cap_period_bounds(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_earning_cap_period_bounds(TEXT) TO service_role;

-- Replace earned_this_period to use calendar month (not 1-24)
CREATE OR REPLACE FUNCTION odyssey_earned_this_period(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_ps TIMESTAMPTZ; v_pe TIMESTAMPTZ; v_sum INT;
BEGIN
    SELECT period_start_ts, period_end_ts INTO v_ps, v_pe FROM odyssey_earning_cap_period_bounds();
    SELECT COALESCE(SUM(amount),0)::INT INTO v_sum
    FROM odyssey_coin_transactions
    WHERE user_uid = p_user_uid
      AND type = 'TASK_REWARD'
      AND created_at >= v_ps
      AND created_at < v_pe;
    RETURN COALESCE(v_sum,0);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_earned_this_period(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_earned_this_period(TEXT) TO service_role;

CREATE OR REPLACE FUNCTION odyssey_earned_this_period(p_user_uid TEXT, p_tz TEXT)
RETURNS INT
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_ps TIMESTAMPTZ; v_pe TIMESTAMPTZ; v_sum INT;
BEGIN
    SELECT period_start_ts, period_end_ts INTO v_ps, v_pe FROM odyssey_earning_cap_period_bounds(p_tz);
    SELECT COALESCE(SUM(amount),0)::INT INTO v_sum
    FROM odyssey_coin_transactions
    WHERE user_uid = p_user_uid
      AND type = 'TASK_REWARD'
      AND created_at >= v_ps
      AND created_at < v_pe;
    RETURN COALESCE(v_sum,0);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_earned_this_period(TEXT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_earned_this_period(TEXT,TEXT) TO service_role;

-- is_earning_halted unchanged logic but now uses calendar-month earned
CREATE OR REPLACE FUNCTION odyssey_is_earning_halted(p_user_uid TEXT)
RETURNS BOOLEAN
LANGUAGE plpgsql STABLE SECURITY DEFINER SET search_path = public AS $$
DECLARE v_cap INT; v_earned INT;
BEGIN
    v_cap := odyssey_get_effective_earning_cap(p_user_uid);
    IF v_cap IS NULL OR v_cap <= 0 THEN
        RETURN false;
    END IF;
    v_earned := odyssey_earned_this_period(p_user_uid);
    RETURN v_earned >= v_cap;
END;
$$;
REVOKE ALL ON FUNCTION odyssey_is_earning_halted(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_is_earning_halted(TEXT) TO service_role;

-- Re-apply earning-cap enforcement in submit/verify RPCs to ensure they use new helper (idempotent)
-- odyssey_submit_auto_task and odyssey_verify_submission already call odyssey_earned_this_period / odyssey_get_effective_earning_cap
-- which are now calendar-month, so no DDL needed beyond helpers above. Re-create is not required but keep version.

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','071_earning_cap_calendar_month')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
