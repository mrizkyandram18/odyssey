-- Migration 077: Drop obsolete odyssey_local_users (final step)
-- ONLY run after: 076 backfill GREEN + runtime reads/writes profiles SOT + safety gates A-G GREEN.
-- After this migration, rollback to pre-076 application code is NOT safe.

DROP INDEX IF EXISTS idx_local_users_profile_uid;

DROP POLICY IF EXISTS
  "Allow service_role full access on local_users"
  ON odyssey_local_users;

DROP TABLE IF EXISTS odyssey_local_users;

-- Record schema version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '077_drop_local_users')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
