-- Local Authentication Schema for Prototype/Development
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
