-- Migration 060: Admin Dashboard Performance & Deterministic Pagination Indexes

-- 1. Composite index for tenant-scoped user profiles with deterministic sort
CREATE INDEX IF NOT EXISTS idx_user_profiles_family_created 
ON odyssey_user_profiles (family_id, created_at DESC, uid DESC);

-- 2. Foreign reference index on local user credentials for fast profile lookup
CREATE INDEX IF NOT EXISTS idx_local_users_profile_uid 
ON odyssey_local_users (profile_uid);

-- 3. Composite index for claims filtered by user and status with deterministic sort
CREATE INDEX IF NOT EXISTS idx_claims_user_status_created 
ON odyssey_claims (user_uid, status, created_at DESC, id DESC);

-- 4. Composite index for task submissions filtered by user and status with deterministic sort
CREATE INDEX IF NOT EXISTS idx_submissions_user_status_created 
ON odyssey_task_submissions (user_uid, status, created_at DESC, id DESC);

-- 5. Record schema version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '060_admin_dashboard_perf_indexes')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
