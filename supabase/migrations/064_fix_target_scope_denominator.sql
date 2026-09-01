-- Migration 064: Fix target denominator to respect USER scope
-- Tasks with target_scope=USER for another user must not affect current user's target calculation

CREATE OR REPLACE FUNCTION odyssey_calc_target_reward(
    p_target INT,
    p_task_id BIGINT,
    p_family_id TEXT,
    p_user_uid TEXT DEFAULT NULL
) RETURNS INT
LANGUAGE plpgsql STABLE
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE v_total_weight INT; v_weight INT; v_period_start DATE; v_period_end DATE; v_actual INT; v_user_join_date DATE;
BEGIN
    IF p_target IS NULL OR p_target<1 OR p_target>10000 THEN p_target:=3200; END IF;
    SELECT period_start, period_end INTO v_period_start, v_period_end FROM odyssey_target_period_bounds();
    IF p_user_uid IS NOT NULL THEN
        SELECT (created_at AT TIME ZONE COALESCE((SELECT value FROM odyssey_system_config WHERE key='timezone'),'Asia/Jakarta'))::date
          INTO v_user_join_date FROM odyssey_user_profiles WHERE uid=p_user_uid;
        IF v_user_join_date IS NOT NULL AND v_user_join_date > v_period_start THEN
            v_period_start := v_user_join_date;
        END IF;
    END IF;
    IF v_period_start >= v_period_end THEN
        RETURN 0;
    END IF;
    SELECT COALESCE(SUM(reward_coins),0)::INT INTO v_total_weight FROM odyssey_tasks
      WHERE family_id = COALESCE(p_family_id, family_id)
        AND is_active = true AND reward_coins > 0
        AND active_date >= v_period_start AND active_date < v_period_end
        AND (p_family_id IS NULL OR family_id = p_family_id)
        AND (target_scope IS NULL OR target_scope <> 'USER' OR target_user_uid = p_user_uid);
    IF v_total_weight IS NULL OR v_total_weight = 0 THEN
        RETURN 0;
    END IF;
    SELECT COALESCE(reward_coins,50) INTO v_weight FROM odyssey_tasks WHERE id=p_task_id;
    IF NOT EXISTS (
        SELECT 1 FROM odyssey_tasks WHERE id=p_task_id
          AND is_active = true AND reward_coins > 0
          AND active_date >= v_period_start AND active_date < v_period_end
          AND (p_family_id IS NULL OR family_id = p_family_id)
          AND (target_scope IS NULL OR target_scope <> 'USER' OR target_user_uid = p_user_uid)
    ) THEN
        RETURN 0;
    END IF;
    WITH eligible AS (
        SELECT id, reward_coins, active_date, step_order,
               floor(p_target::numeric * reward_coins::numeric / v_total_weight::numeric)::int AS base,
               (p_target::numeric * reward_coins::numeric / v_total_weight::numeric) - floor(p_target::numeric * reward_coins::numeric / v_total_weight::numeric) AS frac
        FROM odyssey_tasks
        WHERE family_id = COALESCE(p_family_id, family_id)
          AND is_active = true AND reward_coins > 0
          AND active_date >= v_period_start AND active_date < v_period_end
          AND (p_family_id IS NULL OR family_id = p_family_id)
          AND (target_scope IS NULL OR target_scope <> 'USER' OR target_user_uid = p_user_uid)
    ), sum_base AS (SELECT COALESCE(SUM(base),0)::int AS s FROM eligible),
    ranked AS (
        SELECT e.id, e.base, e.frac, ROW_NUMBER() OVER (ORDER BY e.frac DESC, e.active_date ASC, e.step_order ASC, e.id ASC) AS rnk
        FROM eligible e
    ), remainder AS (SELECT (p_target - (SELECT s FROM sum_base))::int AS r)
    SELECT base + CASE WHEN rnk <= (SELECT r FROM remainder) THEN 1 ELSE 0 END INTO v_actual
    FROM ranked WHERE id = p_task_id;
    IF v_actual IS NULL THEN RETURN 0; END IF;
    RETURN v_actual;
END; $$;
REVOKE ALL ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT,TEXT) TO service_role;

-- Keep wrapper for 3-arg calls
CREATE OR REPLACE FUNCTION odyssey_calc_target_reward(
    p_target INT,
    p_task_id BIGINT,
    p_family_id TEXT
) RETURNS INT
LANGUAGE sql STABLE SECURITY DEFINER SET search_path = public AS $$
    SELECT odyssey_calc_target_reward(p_target, p_task_id, p_family_id, NULL);
$$;
REVOKE ALL ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_calc_target_reward(INT,BIGINT,TEXT) TO service_role;

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','064_fix_target_scope_denominator')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
