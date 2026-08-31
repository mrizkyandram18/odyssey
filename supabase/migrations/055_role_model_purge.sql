-- Migration 055: Purge RPG Legacy Roles into Clean Product Roles (ADMIN / MEMBER)

UPDATE odyssey_user_profiles SET role = 'ADMIN' WHERE role IN ('GUIDE', 'BUILDER');
UPDATE odyssey_user_profiles SET role = 'MEMBER' WHERE role IN ('SEEKER');

ALTER TABLE odyssey_user_profiles ALTER COLUMN role SET DEFAULT 'MEMBER';

INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '055_role_model_purge')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
