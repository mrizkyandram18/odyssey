-- ============================================================
-- Migration 066: Make auto-block cycle-aware (earning cycle reset)
-- Inactivity is counted only within current earning cycle via
-- odyssey_target_period_bounds(). Cross-cycle accumulation is NOT allowed.
-- Never-completed in current cycle => NOT blocked.
-- ============================================================

-- Replace auto-block RPC with cycle-aware version
CREATE OR REPLACE FUNCTION odyssey_auto_block_inactive_users()
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
    v_threshold INT;
    v_tz TEXT;
    v_today DATE;
    v_period_start DATE;
    v_period_end DATE;
    v_blocked_count INT := 0;
BEGIN
    v_threshold := odyssey_resolve_auto_block_days();
    IF v_threshold IS NULL OR v_threshold <=0 THEN
        RETURN jsonb_build_object('success',true,'blocked_count',0,'reason','threshold disabled or invalid','threshold',v_threshold);
    END IF;
    SELECT COALESCE(NULLIF(trim(value),''), 'Asia/Jakarta') INTO v_tz FROM odyssey_system_config WHERE key='timezone';
    IF v_tz IS NULL OR trim(v_tz)='' THEN v_tz:='Asia/Jakarta'; END IF;
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    BEGIN v_today := (timezone(v_tz, now()))::date; EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; v_today := (timezone(v_tz, now()))::date; END;

    -- Resolve current earning cycle bounds using canonical helper
    BEGIN
        SELECT period_start, period_end INTO v_period_start, v_period_end FROM odyssey_target_period_bounds(v_tz);
    EXCEPTION WHEN OTHERS THEN
        v_period_start := date_trunc('month', timezone(v_tz, now()))::date;
        v_period_end := v_period_start + interval '1 month';
        v_period_end := v_period_end::date;
    END;
    IF v_period_start IS NULL OR v_period_end IS NULL THEN
        v_period_start := date_trunc('month', timezone(v_tz, now()))::date;
        v_period_end := (v_period_start + interval '1 month')::date;
    END IF;

    -- Cycle-aware last success: only submissions within current cycle period
    WITH last_success AS (
        SELECT user_uid, max((COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date) AS last_date
        FROM odyssey_task_submissions
        WHERE status='APPROVED'
          AND (COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date >= v_period_start
          AND (COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date < v_period_end
        GROUP BY user_uid
    )
    UPDATE odyssey_user_profiles p
    SET is_active=false,
        blocked_at=timezone('utc'::text, now()),
        blocked_by=NULL,
        block_reason='auto-block: inactivity >= ' || v_threshold || ' days (cycle ' || v_period_start::text || ' to ' || (v_period_end - interval '1 day')::date::text || ')',
        updated_at=timezone('utc'::text, now())
    WHERE p.is_active = true
      AND p.role IN ('MEMBER','SEEKER')
      AND p.blocked_at IS NULL
      AND EXISTS (SELECT 1 FROM last_success ls WHERE ls.user_uid=p.uid AND ls.last_date IS NOT NULL AND (v_today - ls.last_date) >= v_threshold)
      -- Never-completed in current cycle (no row in last_success) => NOT blocked, handled by EXISTS false
    ;
    GET DIAGNOSTICS v_blocked_count = ROW_COUNT;

    RETURN jsonb_build_object('success',true,'blocked_count',v_blocked_count,'threshold',v_threshold,'today',v_today::text,'timezone',v_tz,'period_start',v_period_start::text,'period_end',v_period_end::text,'cycle_aware',true);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_auto_block_inactive_users() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_auto_block_inactive_users() TO service_role;

CREATE OR REPLACE FUNCTION odyssey_auto_block_inactive_users_with_threshold(p_threshold INT, p_tz TEXT DEFAULT NULL)
RETURNS JSONB
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_threshold INT; v_tz TEXT; v_today DATE; v_period_start DATE; v_period_end DATE; v_blocked_count INT :=0;
BEGIN
    v_threshold := p_threshold;
    IF v_threshold IS NULL OR v_threshold <=0 OR v_threshold >365 THEN
        v_threshold := odyssey_resolve_auto_block_days();
    END IF;
    IF v_threshold IS NULL OR v_threshold <=0 THEN
        RETURN jsonb_build_object('success',true,'blocked_count',0,'reason','threshold disabled','threshold',v_threshold);
    END IF;
    v_tz := COALESCE(NULLIF(trim(p_tz),''), COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'),'Asia/Jakarta'));
    IF v_tz IS NULL OR trim(v_tz)='' THEN v_tz:='Asia/Jakarta'; END IF;
    BEGIN PERFORM (SELECT 1 FROM pg_timezone_names WHERE name=v_tz); EXCEPTION WHEN OTHERS THEN v_tz:='Asia/Jakarta'; END;
    v_today := (timezone(v_tz, now()))::date;
    BEGIN SELECT period_start, period_end INTO v_period_start, v_period_end FROM odyssey_target_period_bounds(v_tz); EXCEPTION WHEN OTHERS THEN v_period_start := date_trunc('month', timezone(v_tz, now()))::date; v_period_end := (v_period_start + interval '1 month')::date; END;
    WITH last_success AS (
        SELECT user_uid, max((COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date) AS last_date
        FROM odyssey_task_submissions WHERE status='APPROVED'
          AND (COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date >= v_period_start
          AND (COALESCE(reviewed_at, created_at) AT TIME ZONE v_tz)::date < v_period_end
        GROUP BY user_uid
    )
    UPDATE odyssey_user_profiles p
    SET is_active=false, blocked_at=timezone('utc'::text, now()), blocked_by=NULL, block_reason='auto-block: inactivity >= '||v_threshold||' days (cycle '||v_period_start::text||' to '||(v_period_end - interval '1 day')::date::text||')', updated_at=timezone('utc'::text, now())
    WHERE p.is_active=true AND p.role IN ('MEMBER','SEEKER') AND p.blocked_at IS NULL
      AND EXISTS (SELECT 1 FROM last_success ls WHERE ls.user_uid=p.uid AND ls.last_date IS NOT NULL AND (v_today - ls.last_date) >= v_threshold);
    GET DIAGNOSTICS v_blocked_count = ROW_COUNT;
    RETURN jsonb_build_object('success',true,'blocked_count',v_blocked_count,'threshold',v_threshold,'today',v_today::text,'timezone',v_tz,'period_start',v_period_start::text,'period_end',v_period_end::text,'cycle_aware',true);
END;
$$;
REVOKE ALL ON FUNCTION odyssey_auto_block_inactive_users_with_threshold(INT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_auto_block_inactive_users_with_threshold(INT,TEXT) TO service_role;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','066_cycle_aware_auto_block')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
