-- Migration 012: Reliability & idempotency schema changes
-- 1. Drop the overly restrictive unique index on odyssey_daily_missions(uid, date)
--    The application supports MaxTurnsPerDay > 1, so multiple rows per user per day
--    are valid. The unique index prevented legitimate multiple daily turns.
-- 2. Add a version column to odyssey_user_profiles for optimistic concurrency
--    control on XP/level updates to prevent lost-update races.

-- ============================================================
-- odyssey_daily_missions: remove restrictive unique index
-- ============================================================
DROP INDEX IF EXISTS uniq_odyssey_daily_missions_uid_date;

-- ============================================================
-- odyssey_user_profiles: add version column for optimistic locking
-- ============================================================
ALTER TABLE odyssey_user_profiles
    ADD COLUMN IF NOT EXISTS version INTEGER NOT NULL DEFAULT 1;

-- ============================================================
-- Update schema_version to reflect migration 012
-- ============================================================
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '12')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
