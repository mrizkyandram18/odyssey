-- ============================================================
-- Migration 051: Admin Control Plane - User Active Status & Task Targeting
-- Adds is_active to odyssey_user_profiles
-- Adds target_scope ('ALL', 'FAMILY', 'USER') and target_user_uid to odyssey_tasks
-- ============================================================

-- 1. Upgrade odyssey_user_profiles with is_active
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_user_profiles' AND column_name = 'is_active') THEN
        ALTER TABLE odyssey_user_profiles ADD COLUMN is_active BOOLEAN NOT NULL DEFAULT TRUE;
    END IF;
END $$;

-- 2. Upgrade odyssey_tasks with target_scope & target_user_uid
DO $$ BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'target_scope') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN target_scope TEXT NOT NULL DEFAULT 'ALL';
    END IF;

    -- Ensure check constraint for target_scope
    ALTER TABLE odyssey_tasks DROP CONSTRAINT IF EXISTS odyssey_tasks_target_scope_check;
    ALTER TABLE odyssey_tasks ADD CONSTRAINT odyssey_tasks_target_scope_check 
        CHECK (target_scope IN ('ALL', 'FAMILY', 'USER'));

    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'odyssey_tasks' AND column_name = 'target_user_uid') THEN
        ALTER TABLE odyssey_tasks ADD COLUMN target_user_uid TEXT REFERENCES odyssey_user_profiles(uid) ON DELETE SET NULL;
    END IF;
END $$;

-- Index for targeted task lookups
CREATE INDEX IF NOT EXISTS idx_tasks_target ON odyssey_tasks (family_id, target_scope, target_user_uid, active_date) WHERE is_active = TRUE;

-- 3. Register migration version in odyssey_schema_version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '051_admin_user_task_control')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
