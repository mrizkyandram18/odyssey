-- Migration 002: Indexes for Odyssey MVP
-- All indexes use the odyssey_ prefix and follow the naming convention
-- idx_<table>_<column> as documented in docs/decisions/ADR-003-database.md

-- odyssey_user_profiles
CREATE INDEX IF NOT EXISTS idx_odyssey_user_profiles_crew_id ON odyssey_user_profiles(crew_id);

-- odyssey_quests
CREATE INDEX IF NOT EXISTS idx_odyssey_quests_crew_id ON odyssey_quests(crew_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_quests_crew_id_slug
    ON odyssey_quests(crew_id, template_slug);

-- odyssey_challenges
CREATE INDEX IF NOT EXISTS idx_odyssey_challenges_quest_id ON odyssey_challenges(quest_id);

-- odyssey_realm_progress
-- crew_id is already the primary key; no additional index needed.

-- odyssey_creative_items
CREATE INDEX IF NOT EXISTS idx_odyssey_creative_items_crew_id ON odyssey_creative_items(crew_id);
CREATE INDEX IF NOT EXISTS idx_odyssey_creative_items_author_uid ON odyssey_creative_items(author_uid);

-- odyssey_daily_turns
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_daily_turns_uid_date
    ON odyssey_daily_turns(uid, date);

-- odyssey_achievements
CREATE INDEX IF NOT EXISTS idx_odyssey_achievements_uid ON odyssey_achievements(uid);
CREATE INDEX IF NOT EXISTS idx_odyssey_achievements_crew_id ON odyssey_achievements(crew_id);

-- odyssey_relics
CREATE INDEX IF NOT EXISTS idx_odyssey_relics_uid ON odyssey_relics(uid);

-- odyssey_chests
CREATE INDEX IF NOT EXISTS idx_odyssey_chests_uid ON odyssey_chests(uid);
