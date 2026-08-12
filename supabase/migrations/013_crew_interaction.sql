-- Migration 013: Family Interaction (Reactions & Activity Tracking)

-- Reactions table (Prototype schema)
CREATE TABLE IF NOT EXISTS odyssey_reactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id TEXT NOT NULL,
    target_user_id TEXT NOT NULL,
    mission_id UUID, -- nullable, for when reaction is linked to a specific quest
    emoji_code TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_reactions_target_user ON odyssey_reactions(target_user_id);

-- Daily Activity tracking (for streaks and history)
CREATE TABLE IF NOT EXISTS odyssey_daily_activity (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id TEXT NOT NULL,
    activity_date DATE NOT NULL,
    activity_type TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    -- Prevent duplicate activity types per user per day
    UNIQUE(user_id, activity_date, activity_type)
);

CREATE INDEX IF NOT EXISTS idx_daily_activity_user_date ON odyssey_daily_activity(user_id, activity_date);
ALTER TABLE odyssey_reactions ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on reactions" ON odyssey_reactions;
CREATE POLICY "Allow service_role full access on reactions" ON odyssey_reactions FOR ALL TO service_role USING (true);

ALTER TABLE odyssey_daily_activity ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on daily_activity" ON odyssey_daily_activity;
CREATE POLICY "Allow service_role full access on daily_activity" ON odyssey_daily_activity FOR ALL TO service_role USING (true);
