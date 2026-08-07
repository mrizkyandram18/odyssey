-- Migration 007: World Progression System
-- Adds tables and columns for chapter progression, lore unlocks,
-- seasonal content filtering, quest prerequisites, and achievement triggers.
-- All new columns are additive (nullable with defaults) — no breaking changes.
-- Follows ADR-003 conventions: odyssey_ prefix, snake_case, timestamps, RLS,
-- service_role policy.

-- ============================================================
-- ALTER: odyssey_quest_definitions — add prerequisite & seasonality columns
-- ============================================================
ALTER TABLE odyssey_quest_definitions
    ADD COLUMN IF NOT EXISTS is_mandatory      BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS required_quest_slug TEXT,
    ADD COLUMN IF NOT EXISTS required_chapter     TEXT,
    ADD COLUMN IF NOT EXISTS required_realm       TEXT,
    ADD COLUMN IF NOT EXISTS required_level       INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS season_slug          TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_season
    ON odyssey_quest_definitions (season_slug);
CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_required_chapter
    ON odyssey_quest_definitions (required_chapter);
CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_chapter
    ON odyssey_quest_definitions (chapter);

-- ============================================================
-- ALTER: odyssey_lore_definitions — link lore to chapters & seasons
-- ============================================================
ALTER TABLE odyssey_lore_definitions
    ADD COLUMN IF NOT EXISTS chapter     TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS season_slug TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_lore_definitions_chapter
    ON odyssey_lore_definitions (chapter);
CREATE INDEX IF NOT EXISTS idx_odyssey_lore_definitions_season
    ON odyssey_lore_definitions (season_slug);

-- ============================================================
-- ALTER: odyssey_achievement_definitions — add trigger & seasonality
-- The `trigger` column defines what event kinds activate this achievement.
-- Values are free-text to stay definition-driven (no hardcoded enums in DB).
-- ============================================================
ALTER TABLE odyssey_achievement_definitions
    ADD COLUMN IF NOT EXISTS trigger     TEXT NOT NULL DEFAULT 'QUEST_COMPLETED',
    ADD COLUMN IF NOT EXISTS season_slug TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_achievement_definitions_trigger
    ON odyssey_achievement_definitions (trigger);
CREATE INDEX IF NOT EXISTS idx_odyssey_achievement_definitions_season
    ON odyssey_achievement_definitions (season_slug);

-- ============================================================
-- ALTER: odyssey_creative_prompt_definitions — add seasonality
-- ============================================================
ALTER TABLE odyssey_creative_prompt_definitions
    ADD COLUMN IF NOT EXISTS season_slug TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_creative_prompt_definitions_season
    ON odyssey_creative_prompt_definitions (season_slug);

-- ============================================================
-- ALTER: odyssey_chest_definitions — add seasonality
-- ============================================================
ALTER TABLE odyssey_chest_definitions
    ADD COLUMN IF NOT EXISTS season_slug TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_chest_definitions_season
    ON odyssey_chest_definitions (season_slug);

-- ============================================================
-- ALTER: odyssey_achievements (player achievements) — add trigger & count
-- ============================================================
ALTER TABLE odyssey_achievements
    ADD COLUMN IF NOT EXISTS trigger       TEXT,
    ADD COLUMN IF NOT EXISTS completion_count INTEGER NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_odyssey_achievements_trigger
    ON odyssey_achievements (trigger);
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_achievements_uid_code
    ON odyssey_achievements (uid, code)
    WHERE uid IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_achievements_crew_id_code
    ON odyssey_achievements (crew_id, code)
    WHERE crew_id IS NOT NULL AND uid IS NULL;

-- ============================================================
-- CREATE: odyssey_chapter_progress
-- Tracks a crew's progress through each chapter.
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_chapter_progress (
    crew_id      TEXT    NOT NULL,
    chapter      TEXT    NOT NULL,
    realm        TEXT    NOT NULL,
    status       TEXT    NOT NULL DEFAULT 'LOCKED',
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at   TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    PRIMARY KEY (crew_id, chapter)
);

CREATE INDEX IF NOT EXISTS idx_odyssey_chapter_progress_crew_id
    ON odyssey_chapter_progress (crew_id);
CREATE INDEX IF NOT EXISTS idx_odyssey_chapter_progress_realm
    ON odyssey_chapter_progress (realm);
CREATE INDEX IF NOT EXISTS idx_odyssey_chapter_progress_status
    ON odyssey_chapter_progress (status);

-- ============================================================
-- CREATE: odyssey_lore_unlocks
-- Tracks which lore entries a crew has unlocked.
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_lore_unlocks (
    crew_id     TEXT NOT NULL,
    lore_slug   TEXT NOT NULL,
    realm       TEXT NOT NULL,
    chapter     TEXT NOT NULL DEFAULT '',
    unlocked_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    PRIMARY KEY (crew_id, lore_slug)
);

CREATE INDEX IF NOT EXISTS idx_odyssey_lore_unlocks_crew_id
    ON odyssey_lore_unlocks (crew_id);
CREATE INDEX IF NOT EXISTS idx_odyssey_lore_unlocks_lore_slug
    ON odyssey_lore_unlocks (lore_slug);

-- ============================================================
-- ALTER: odyssey_quests — add chapter column for instance tracking
-- ============================================================
ALTER TABLE odyssey_quests
    ADD COLUMN IF NOT EXISTS chapter TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_odyssey_quests_chapter
    ON odyssey_quests (chapter);
CREATE INDEX IF NOT EXISTS idx_odyssey_quests_status
    ON odyssey_quests (status);

-- ============================================================
-- Row-Level Security for new tables
-- ============================================================
ALTER TABLE odyssey_chapter_progress ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_lore_unlocks ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'odyssey_chapter_progress'
        AND policyname = 'Allow service_role full access'
    ) THEN
        EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_chapter_progress FOR ALL TO service_role USING (true)';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'odyssey_lore_unlocks'
        AND policyname = 'Allow service_role full access'
    ) THEN
        EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_lore_unlocks FOR ALL TO service_role USING (true)';
    END IF;
END $$;
