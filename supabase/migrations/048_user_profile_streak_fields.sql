-- ============================================================
-- Migration 048: User Profile Streak Columns and RPC Fix
-- Adds streak_days and last_active_date to odyssey_user_profiles
-- ensuring odyssey_update_user_streak and dependent task submissions succeed.
-- ============================================================

DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_user_profiles' AND column_name = 'streak_days') THEN
        ALTER TABLE odyssey_user_profiles ADD COLUMN streak_days INT NOT NULL DEFAULT 0;
    END IF;

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_user_profiles' AND column_name = 'last_active_date') THEN
        ALTER TABLE odyssey_user_profiles ADD COLUMN last_active_date DATE;
    END IF;
END $$;

-- Recompile odyssey_update_user_streak to ensure clean dependency binding
CREATE OR REPLACE FUNCTION odyssey_update_user_streak(p_user_uid TEXT)
RETURNS INT
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
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

REVOKE ALL ON FUNCTION odyssey_update_user_streak(TEXT) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION odyssey_update_user_streak(TEXT) TO service_role;

-- Record migration version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '048_user_profile_streak_fields')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
