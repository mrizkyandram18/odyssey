-- Migration 011: Add required_quest_slugs for multi-prerequisite quest chains
-- Adds a JSONB column to odyssey_quest_definitions to support multiple
-- quest prerequisites. Existing single prerequisite is preserved in
-- required_quest_slug for backward compatibility.
--
-- Prerequisites: migrations 001-010 must be applied first.

ALTER TABLE odyssey_quest_definitions
  ADD COLUMN IF NOT EXISTS required_quest_slugs JSONB NOT NULL DEFAULT '[]'::jsonb;

-- Update schema version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '11')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
