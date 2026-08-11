-- ADR-004: Family Reward Integration Signals
-- This table acts as a decoupled interface for the Family Reward system.

CREATE TABLE IF NOT EXISTS odyssey_reward_signals (
    uid TEXT NOT NULL,
    achievement_code TEXT NOT NULL,
    issued_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    consumed BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (uid, achievement_code)
);

-- Enable RLS
ALTER TABLE odyssey_reward_signals ENABLE ROW LEVEL SECURITY;

-- Service role access only for MVP.
-- The backend uses the service role key, so it bypasses RLS, but we explicitly
-- grant access and prevent public access.
CREATE POLICY "Service role can manage reward signals" 
    ON odyssey_reward_signals
    FOR ALL 
    TO service_role 
    USING (true)
    WITH CHECK (true);
