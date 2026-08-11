-- Migration 026: Custom Crew Banner & Shared Realm Themes (Slice 4.4)
-- Adds banner_url and theme columns to odyssey_crews for crew customization.

ALTER TABLE odyssey_crews ADD COLUMN IF NOT EXISTS banner_url TEXT;
ALTER TABLE odyssey_crews ADD COLUMN IF NOT EXISTS theme TEXT DEFAULT 'default';

INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '026_crew_banner_and_theme')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
