-- Migration 006: Content Definitions Tables
-- All tables use the odyssey_ prefix and follow ADR-003 conventions.
-- RLS is enabled with the service_role full-access policy.

-- ============================================================
-- odyssey_realm_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_realm_definitions (
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

CREATE INDEX IF NOT EXISTS idx_odyssey_realm_definitions_slug
  ON odyssey_realm_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_realm_definitions_order
  ON odyssey_realm_definitions ("order");

-- ============================================================
-- odyssey_chapter_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_chapter_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  realm TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  "order" INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_chapter_definitions_slug
  ON odyssey_chapter_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_chapter_definitions_realm
  ON odyssey_chapter_definitions (realm);

-- ============================================================
-- odyssey_quest_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_quest_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  realm TEXT NOT NULL,
  chapter TEXT NOT NULL DEFAULT '',
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

CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_realm
  ON odyssey_quest_definitions (realm);

-- ============================================================
-- odyssey_creative_prompt_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_creative_prompt_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  realm TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  prompt TEXT NOT NULL DEFAULT '',
  kind TEXT NOT NULL DEFAULT 'STORY',
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_creative_prompt_definitions_slug
  ON odyssey_creative_prompt_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_creative_prompt_definitions_realm
  ON odyssey_creative_prompt_definitions (realm);

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
  realm TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_season_definitions_slug
  ON odyssey_season_definitions (slug);

-- ============================================================
-- odyssey_lore_definitions
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_lore_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  realm TEXT NOT NULL,
  title TEXT NOT NULL,
  content TEXT NOT NULL DEFAULT '',
  "order" INTEGER NOT NULL DEFAULT 0,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_lore_definitions_slug
  ON odyssey_lore_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_lore_definitions_realm
  ON odyssey_lore_definitions (realm);

-- ============================================================
-- Row-Level Security
-- All odyssey_* tables use service-role RLS policy.
-- ============================================================

ALTER TABLE odyssey_realm_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_chapter_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_quest_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_creative_prompt_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_achievement_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_season_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_lore_definitions ENABLE ROW LEVEL SECURITY;

-- Service role policies
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_realm_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_realm_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_chapter_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_chapter_definitions FOR ALL TO service_role USING (true)';
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
    WHERE schemaname = 'public' AND tablename = 'odyssey_lore_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_lore_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;