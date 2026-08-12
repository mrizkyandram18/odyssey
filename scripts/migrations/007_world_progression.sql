-- Migration 007: World Progression System
-- Adds tables and columns for course progression, concept unlocks,
-- seasonal content filtering, quest prerequisites, and achievement triggers.
-- All new columns are additive (nullable with defaults) — no breaking changes.
-- Follows ADR-003 conventions: odyssey_ prefix, snake_case, timestamps, RLS,
-- service_role policy.

-- ============================================================
-- ALTER: odyssey_quest_definitions — add prerequisite & seasonality columns
-- ============================================================
ALTER TABLE odyssey_quest_definitions
    ADD COLUMN IF NOT EXISTS is_mandatory      BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS required_mission_slug TEXT,
    ADD COLUMN IF NOT EXISTS required_course     TEXT,
    ADD COLUMN IF NOT EXISTS required_journey       TEXT,
    ADD COLUMN IF NOT EXISTS required_level       INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS season_slug          TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_season
    ON odyssey_quest_definitions (season_slug);
CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_required_course
    ON odyssey_quest_definitions (required_course);
CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_course
    ON odyssey_quest_definitions (course);

-- ============================================================
-- ALTER: odyssey_concept_definitions — link concept to chapters & seasons
-- ============================================================
ALTER TABLE odyssey_concept_definitions
    ADD COLUMN IF NOT EXISTS course     TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS season_slug TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_concept_definitions_course
    ON odyssey_concept_definitions (course);
CREATE INDEX IF NOT EXISTS idx_odyssey_concept_definitions_season
    ON odyssey_concept_definitions (season_slug);

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
CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_achievements_family_id_code
    ON odyssey_achievements (family_id, code)
    WHERE family_id IS NOT NULL AND uid IS NULL;

-- ============================================================
-- CREATE: odyssey_course_progress
-- Tracks a crew's progress through each course.
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_course_progress (
    family_id      TEXT    NOT NULL,
    course      TEXT    NOT NULL,
    journey        TEXT    NOT NULL,
    status       TEXT    NOT NULL DEFAULT 'LOCKED',
    completed_at TIMESTAMPTZ,
    created_at   TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at   TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    PRIMARY KEY (family_id, course)
);

CREATE INDEX IF NOT EXISTS idx_odyssey_course_progress_family_id
    ON odyssey_course_progress (family_id);
CREATE INDEX IF NOT EXISTS idx_odyssey_course_progress_journey
    ON odyssey_course_progress (journey);
CREATE INDEX IF NOT EXISTS idx_odyssey_course_progress_status
    ON odyssey_course_progress (status);

-- ============================================================
-- CREATE: odyssey_concept_unlocks
-- Tracks which concept entries a crew has unlocked.
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_concept_unlocks (
    family_id     TEXT NOT NULL,
    concept_slug   TEXT NOT NULL,
    journey       TEXT NOT NULL,
    course     TEXT NOT NULL DEFAULT '',
    unlocked_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    PRIMARY KEY (family_id, concept_slug)
);

CREATE INDEX IF NOT EXISTS idx_odyssey_concept_unlocks_family_id
    ON odyssey_concept_unlocks (family_id);
CREATE INDEX IF NOT EXISTS idx_odyssey_concept_unlocks_concept_slug
    ON odyssey_concept_unlocks (concept_slug);

-- ============================================================
-- ALTER: odyssey_missions — add course column for instance tracking
-- ============================================================
ALTER TABLE odyssey_missions
    ADD COLUMN IF NOT EXISTS course TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_odyssey_missions_course
    ON odyssey_missions (course);
CREATE INDEX IF NOT EXISTS idx_odyssey_missions_status
    ON odyssey_missions (status);

-- ============================================================
-- Row-Level Security for new tables
-- ============================================================
ALTER TABLE odyssey_course_progress ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_concept_unlocks ENABLE ROW LEVEL SECURITY;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'odyssey_course_progress'
        AND policyname = 'Allow service_role full access'
    ) THEN
        EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_course_progress FOR ALL TO service_role USING (true)';
    END IF;
END $$;

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'odyssey_concept_unlocks'
        AND policyname = 'Allow service_role full access'
    ) THEN
        EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_concept_unlocks FOR ALL TO service_role USING (true)';
    END IF;
END $$;
