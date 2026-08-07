-- Epic 4: Reward Foundation

-- Add coins to user profiles
ALTER TABLE odyssey_user_profiles ADD COLUMN IF NOT EXISTS coins BIGINT NOT NULL DEFAULT 0;

-- Reward Ledgers Table
CREATE TABLE IF NOT EXISTS odyssey_reward_ledgers (
    id            TEXT PRIMARY KEY, -- uuid or ulid
    user_id       TEXT NOT NULL REFERENCES odyssey_user_profiles(uid),
    source        TEXT NOT NULL,    -- e.g. 'QUEST_COMPLETED', 'DAILY_STREAK'
    amount        BIGINT NOT NULL DEFAULT 0,
    reward_type   TEXT NOT NULL,    -- 'COINS', 'XP', 'RELIC_ITEM'
    metadata      JSONB,            -- e.g. {"quest_id": 123}
    created_at    TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_odyssey_reward_ledgers_user_id ON odyssey_reward_ledgers(user_id);
CREATE INDEX IF NOT EXISTS idx_odyssey_reward_ledgers_created_at ON odyssey_reward_ledgers(created_at);

-- RLS
ALTER TABLE odyssey_reward_ledgers ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies 
        WHERE schemaname = 'public' AND tablename = 'odyssey_reward_ledgers' AND policyname = 'Allow service_role full access'
    ) THEN
        EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_reward_ledgers FOR ALL TO service_role USING (true)';
    END IF;
END $$;
