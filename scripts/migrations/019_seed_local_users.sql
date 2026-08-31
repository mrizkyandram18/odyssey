-- Migration 019: Seed Local Users for QA
-- This table maps local usernames and bcrypt hashes to the actual profile UIDs.
-- This keeps the domain schema (odyssey_user_profiles) completely free of passwords.

CREATE TABLE IF NOT EXISTS odyssey_local_users (
    id            TEXT PRIMARY KEY,
    username      TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    profile_uid   TEXT NOT NULL REFERENCES odyssey_user_profiles(uid),
    created_at    TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at    TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

ALTER TABLE odyssey_local_users ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on local_users" ON odyssey_local_users;
CREATE POLICY "Allow service_role full access on local_users" ON odyssey_local_users FOR ALL TO service_role USING (true);

-- Seed demo users for QA
INSERT INTO odyssey_local_users (id, username, password_hash, profile_uid)
VALUES 
  ('local-demo-1', 'user_testing', '$2a$10$NmX5Vh66fmfaHOOnp0pnV.NjpCtAkylXUFKlki/dfcxA41Fgy.9Ay', 'demo-uid-1'),
  ('local-demo-2', 'admin', '$2a$10$36SADVo2qG2FhrpvTXMsaO.wGgERYfygx.Y2vDhaNmnlRANAgDRe6', 'demo-uid-2')
ON CONFLICT (username) DO NOTHING;
