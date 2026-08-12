-- Migration 002: Indexes for Odyssey MVP
-- All indexes use the odyssey_ prefix and follow the naming convention
-- idx_<table>_<column> as documented in docs/decisions/ADR-003-database.md

-- odyssey_user_profiles
CREATE INDEX IF NOT EXISTS idx_odyssey_user_profiles_family_id ON odyssey_user_profiles(family_id);

-- odyssey_missions
CREATE INDEX IF NOT EXISTS idx_odyssey_missions_family_id ON odyssey_missions(family_id);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_missions_family_id_slug
    ON odyssey_missions(family_id, template_slug);

-- odyssey_exercises
CREATE INDEX IF NOT EXISTS idx_odyssey_exercises_mission_id ON odyssey_exercises(mission_id);

-- odyssey_journey_progress
-- family_id is already the primary key; no additional index needed.

-- odyssey_creative_items
CREATE INDEX IF NOT EXISTS idx_odyssey_creative_items_family_id ON odyssey_creative_items(family_id);
CREATE INDEX IF NOT EXISTS idx_odyssey_creative_items_author_uid ON odyssey_creative_items(author_uid);

-- odyssey_daily_missions
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_daily_missions_uid_date
    ON odyssey_daily_missions(uid, date);

-- odyssey_achievements
CREATE INDEX IF NOT EXISTS idx_odyssey_achievements_uid ON odyssey_achievements(uid);
CREATE INDEX IF NOT EXISTS idx_odyssey_achievements_family_id ON odyssey_achievements(family_id);

-- odyssey_collections
CREATE INDEX IF NOT EXISTS idx_odyssey_collections_uid ON odyssey_collections(uid);

-- odyssey_gifts
CREATE INDEX IF NOT EXISTS idx_odyssey_gifts_uid ON odyssey_gifts(uid);
