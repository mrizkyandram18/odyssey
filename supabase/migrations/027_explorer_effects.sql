-- Migration 027: Animated Explorer Effects (Slice 4.5)
-- Adds equipped_explorer_effect to user profiles and reward_cosmetic_id
-- to achievement definitions for achievement-gated cosmetic unlocks.

ALTER TABLE odyssey_user_profiles ADD COLUMN IF NOT EXISTS equipped_explorer_effect TEXT NOT NULL DEFAULT 'none';
ALTER TABLE odyssey_achievement_definitions ADD COLUMN IF NOT EXISTS reward_cosmetic_id TEXT NOT NULL DEFAULT '';

INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '027_explorer_effects')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
