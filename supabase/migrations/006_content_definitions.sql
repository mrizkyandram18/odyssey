-- Migration 006: Content Definitions Tables
-- All tables use the odyssey_ prefix and follow ADR-003 conventions.
-- RLS is enabled with the service_role full-access policy.

-- ============================================================
-- odyssey_journey_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_journey_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  "order" INTEGER NOT NULL DEFAULT 0,
  max_progress INTEGER NOT NULL DEFAULT 100,
  icon TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_journey_definitions_slug
  ON odyssey_journey_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_journey_definitions_order
  ON odyssey_journey_definitions ("order");

-- ============================================================
-- odyssey_course_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_course_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  journey TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  "order" INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_course_definitions_slug
  ON odyssey_course_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_course_definitions_journey
  ON odyssey_course_definitions (journey);

-- ============================================================
-- odyssey_quest_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_quest_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  journey TEXT NOT NULL,
  course TEXT NOT NULL DEFAULT '',
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  quest_type TEXT NOT NULL DEFAULT 'SOLO',
  challenge_defs JSONB NOT NULL DEFAULT '[]',
  reward_xp BIGINT NOT NULL DEFAULT 0,
  reward_chest TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_slug
  ON odyssey_quest_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_journey
  ON odyssey_quest_definitions (journey);

-- ============================================================
-- odyssey_creative_prompt_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_creative_prompt_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  journey TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  prompt TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL DEFAULT 'STORY',
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_creative_prompt_definitions_slug
  ON odyssey_creative_prompt_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_creative_prompt_definitions_journey
  ON odyssey_creative_prompt_definitions (journey);

-- ============================================================
-- odyssey_achievement_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_achievement_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  code TEXT NOT NULL UNIQUE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL DEFAULT 'PERSONAL',
  threshold INTEGER NOT NULL DEFAULT 1,
  reward_xp BIGINT NOT NULL DEFAULT 0,
  reward_relic TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_achievement_definitions_code
  ON odyssey_achievement_definitions (code);

-- ============================================================
-- odyssey_season_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_season_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  start_at TIMESTAMPTZ NOT NULL,
  end_at TIMESTAMPTZ NOT NULL,
  journey TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_season_definitions_slug
  ON odyssey_season_definitions (slug);

-- ============================================================
-- odyssey_concept_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_concept_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  journey TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '',
  "order" INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_concept_definitions_slug
  ON odyssey_concept_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_concept_definitions_journey
  ON odyssey_concept_definitions (journey);

-- ============================================================
-- Row-Level Security
-- All odyssey_* tables use service-role RLS policy.
-- ============================================================

ALTER TABLE odyssey_journey_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_course_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_quest_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_creative_prompt_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_achievement_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_season_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_concept_definitions ENABLE ROW LEVEL SECURITY;

-- Service role policies
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_journey_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_journey_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_course_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_course_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_quest_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_quest_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_creative_prompt_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_creative_prompt_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_achievement_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_achievement_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_season_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_season_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_concept_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_concept_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;