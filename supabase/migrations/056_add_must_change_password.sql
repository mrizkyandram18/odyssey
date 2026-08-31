-- Migration 056: Add must_change_password column
ALTER TABLE odyssey_user_profiles 
ADD COLUMN IF NOT EXISTS must_change_password BOOLEAN NOT NULL DEFAULT FALSE;

-- Ensure admin accounts never have must_change_password set to true
UPDATE odyssey_user_profiles 
SET must_change_password = FALSE 
WHERE role = 'ADMIN' OR uid = 'demo-uid-2';
