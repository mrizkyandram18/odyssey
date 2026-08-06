-- Migration 008: Live Operations & Production Readiness
-- Adds: content versioning (published/draft/version/updated_by/published_at),
--       soft delete (deleted_at), audit log table, balance config table,
--       indexes, and published_at backfill for existing rows.
-- No breaking changes: all new columns are additive with safe defaults.

-- ============================================================
-- Versioning model:
--   * published BOOLEAN  — true = live row visible to players
--   * draft JSONB        — working copy snapshot (null = no draft)
--   * version INTEGER    — incremented on every draft save / publish
--   * updated_by TEXT    — admin UID who last touched the draft
--   * published_at       — timestamp of last publish
--   * deleted_at         — NULL = active; non-NULL = soft-deleted (hidden from players)
--
-- Workflow:
--   SaveDraft  → set draft JSONB, version++, updated_by = admin
--   Publish    → copy draft JSON to live columns, published=true, published_at=now, version++, clear draft
--   Preview    → read draft JSON (or live columns if no draft)
--   SoftDelete → deleted_at=now(), published=false (hidden from players)
--   Restore    → deleted_at=NULL (and published=true if it was published before)
-- ============================================================

-- ============================================================
-- Versioning + Soft Delete columns for definition tables
-- ============================================================

-- odyssey_realm_definitions
ALTER TABLE odyssey_realm_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft           JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_chapter_definitions
ALTER TABLE odyssey_chapter_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_quest_definitions
ALTER TABLE odyssey_quest_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_creative_prompt_definitions
ALTER TABLE odyssey_creative_prompt_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_achievement_definitions
ALTER TABLE odyssey_achievement_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_season_definitions
ALTER TABLE odyssey_season_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_lore_definitions
ALTER TABLE odyssey_lore_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_chest_definitions
ALTER TABLE odyssey_chest_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- odyssey_relic_definitions
ALTER TABLE odyssey_relic_definitions
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- ============================================================
-- Versioning + Soft Delete for odyssey_drop_tables
-- ============================================================
ALTER TABLE odyssey_drop_tables
    ADD COLUMN IF NOT EXISTS published     BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN IF NOT EXISTS draft         JSONB,
    ADD COLUMN IF NOT EXISTS version       INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN IF NOT EXISTS updated_by    TEXT,
    ADD COLUMN IF NOT EXISTS published_at  TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS deleted_at    TIMESTAMPTZ;

-- ============================================================
-- Backfill published_at for existing rows (use created_at)
-- Must run after ALTER TABLE adds the published_at column.
-- ============================================================
UPDATE odyssey_realm_definitions      SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_chapter_definitions    SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_quest_definitions      SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_creative_prompt_definitions SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_achievement_definitions SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_season_definitions     SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_lore_definitions       SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_chest_definitions      SET published_at = created_at WHERE published_at IS NULL;
UPDATE odyssey_relic_definitions      SET published_at = created_at WHERE published_at IS NULL;

-- ============================================================
-- Indexes for versioning + soft delete filters
-- ============================================================
-- ContentService queries filter on (published = true, deleted_at IS NULL)
-- Add partial indexes for performance.

CREATE INDEX IF NOT EXISTS idx_odyssey_realm_definitions_published  ON odyssey_realm_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_realm_definitions_deleted  ON odyssey_realm_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_chapter_definitions_published  ON odyssey_chapter_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_chapter_definitions_deleted  ON odyssey_chapter_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_published  ON odyssey_quest_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_quest_definitions_deleted  ON odyssey_quest_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_creative_prompt_definitions_published  ON odyssey_creative_prompt_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_creative_prompt_definitions_deleted  ON odyssey_creative_prompt_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_achievement_definitions_published  ON odyssey_achievement_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_achievement_definitions_deleted  ON odyssey_achievement_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_season_definitions_published  ON odyssey_season_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_season_definitions_deleted  ON odyssey_season_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_lore_definitions_published  ON odyssey_lore_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_lore_definitions_deleted  ON odyssey_lore_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_chest_definitions_published  ON odyssey_chest_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_chest_definitions_deleted  ON odyssey_chest_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_relic_definitions_published  ON odyssey_relic_definitions (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_relic_definitions_deleted  ON odyssey_relic_definitions (deleted_at) WHERE deleted_at IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_odyssey_drop_tables_published ON odyssey_drop_tables (published) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_odyssey_drop_tables_deleted ON odyssey_drop_tables (deleted_at) WHERE deleted_at IS NOT NULL;

-- ============================================================
-- odyssey_balance_configs — runtime balancing overrides
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_balance_configs (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    key        TEXT NOT NULL UNIQUE,
    value      JSONB NOT NULL,
    updated_by TEXT,
    created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_balance_configs_key ON odyssey_balance_configs (key);

-- RLS + service role policy
ALTER TABLE odyssey_balance_configs ENABLE ROW LEVEL SECURITY;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'odyssey_balance_configs'
        AND policyname = 'Allow service_role full access'
    ) THEN
        EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_balance_configs FOR ALL TO service_role USING (true)';
    END IF;
END $$;

-- ============================================================
-- odyssey_audit_logs — admin action audit trail
-- ============================================================
CREATE TABLE IF NOT EXISTS odyssey_audit_logs (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    resource    TEXT NOT NULL,
    resource_id TEXT,
    operation   TEXT NOT NULL,
    admin_uid   TEXT NOT NULL,
    old_value   JSONB,
    new_value   JSONB,
    created_at  TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_odyssey_audit_logs_resource   ON odyssey_audit_logs (resource);
CREATE INDEX IF NOT EXISTS idx_odyssey_audit_logs_admin_uid  ON odyssey_audit_logs (admin_uid);
CREATE INDEX IF NOT EXISTS idx_odyssey_audit_logs_created_at ON odyssey_audit_logs (created_at);
CREATE INDEX IF NOT EXISTS idx_odyssey_audit_logs_operation  ON odyssey_audit_logs (operation);

-- RLS + service role policy
ALTER TABLE odyssey_audit_logs ENABLE ROW LEVEL SECURITY;
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_policies
        WHERE schemaname = 'public' AND tablename = 'odyssey_audit_logs'
        AND policyname = 'Allow service_role full access'
    ) THEN
        EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_audit_logs FOR ALL TO service_role USING (true)';
    END IF;
END $$;

-- ============================================================
-- odyssey_system_config — existing table; add build_version row
-- (seeded by deployment, not a definition table — no versioning needed)
-- ============================================================
-- No schema changes needed; the ConfigStore already reads from odyssey_system_config.
