-- Migration 076: Move username + credential SOT to odyssey_user_profiles (additive, idempotent)
-- Prerequisite for removing odyssey_local_users. Does NOT drop anything.
-- Safe to run multiple times. Preserves bcrypt hashes byte-for-byte.
-- Does NOT overwrite existing valid profile values. Does NOT generate passwords.

-- 1. Add username column FIRST (guards below reference it).
-- Nullable first; NOT NULL enforced only after backfill verification.
ALTER TABLE odyssey_user_profiles ADD COLUMN IF NOT EXISTS username TEXT;

-- 2. Unique index for nullable existing data (partial index avoids blocking NULL rows)
CREATE UNIQUE INDEX IF NOT EXISTS uq_profiles_username
ON odyssey_user_profiles(username)
WHERE username IS NOT NULL;

-- 3. Conflict guards (fail closed). If any of these raise, STOP and resolve
-- manually. Do NOT silently overwrite conflicting data.
DO $$ BEGIN
  IF EXISTS (
    SELECT username FROM odyssey_user_profiles WHERE username IS NOT NULL
    GROUP BY username HAVING count(*) > 1
  ) THEN
    RAISE EXCEPTION 'migration 076 blocked: duplicate username already present in odyssey_user_profiles';
  END IF;
  IF EXISTS (
    SELECT username FROM odyssey_local_users GROUP BY username HAVING count(*) > 1
  ) THEN
    RAISE EXCEPTION 'migration 076 blocked: duplicate username in odyssey_local_users';
  END IF;
  IF EXISTS (
    SELECT profile_uid FROM odyssey_local_users GROUP BY profile_uid HAVING count(*) > 1
  ) THEN
    RAISE EXCEPTION 'migration 076 blocked: duplicate profile_uid in odyssey_local_users';
  END IF;
  IF EXISTS (
    SELECT p.uid FROM odyssey_user_profiles p JOIN odyssey_local_users l
      ON l.profile_uid = p.uid
    WHERE p.password_hash IS NOT NULL AND l.password_hash IS NOT NULL
      AND p.password_hash <> l.password_hash
  ) THEN
    RAISE EXCEPTION 'migration 076 blocked: conflicting password_hash between profiles and local_users';
  END IF;
  IF EXISTS (
    SELECT p.uid FROM odyssey_user_profiles p JOIN odyssey_local_users l
      ON l.profile_uid = p.uid
    WHERE p.username IS NOT NULL AND p.username <> l.username
  ) THEN
    RAISE EXCEPTION 'migration 076 blocked: conflicting username between profiles and local_users';
  END IF;
END $$;

-- 4. Backfill username + password_hash from odyssey_local_users, ONLY where profile is NULL.
-- Idempotent and rerunnable. No decrypt / rehash / password generation.
UPDATE odyssey_user_profiles p SET
  username = COALESCE(p.username, l.username),
  password_hash = COALESCE(p.password_hash, l.password_hash)
FROM odyssey_local_users l
WHERE p.uid = l.profile_uid
  AND (p.username IS NULL OR p.password_hash IS NULL);

-- 5. Record schema version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '076_profiles_username_backfill')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
