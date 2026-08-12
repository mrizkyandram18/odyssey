-- Migration 033: Phase 3 Economy Pacing & Deterministic Rewards
-- 1. Updates balance_configs for xp_per_level and max_new_quests_per_day
-- 2. Adds started_by to odyssey_quests to track user quest pacing
-- 3. Adds reward_relic mapping to quest definitions and chests

-- 1. Balance Configs Update
INSERT INTO odyssey_balance_configs (key, value, updated_by)
VALUES
  ('xp_per_level', '500'::jsonb, 'system'),
  ('max_new_quests_per_day', '1'::jsonb, 'system')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_by = 'system',
  updated_at = timezone('utc'::text, now());

-- 2. Track Started By for pacing limit per-user
ALTER TABLE odyssey_quests
  ADD COLUMN IF NOT EXISTS started_by TEXT REFERENCES odyssey_user_profiles(uid);

-- 3. Deterministic Reward Mappings (No RNG)
ALTER TABLE odyssey_quest_definitions
  ADD COLUMN IF NOT EXISTS reward_relic TEXT;

ALTER TABLE odyssey_chests
  ADD COLUMN IF NOT EXISTS reward_relic TEXT;

-- Seed explicit reward_relic for existing quests
UPDATE odyssey_quest_definitions SET reward_relic = 'acorn-shard' WHERE slug = 'morning-light';
UPDATE odyssey_quest_definitions SET reward_relic = 'acorn-shard' WHERE slug = 'gather-herbs';
UPDATE odyssey_quest_definitions SET reward_relic = 'whispering-leaf' WHERE slug = 'riddle-of-the-stones';
UPDATE odyssey_quest_definitions SET reward_relic = 'whispering-leaf' WHERE slug = 'shadow-trail';
UPDATE odyssey_quest_definitions SET reward_relic = 'ancient-oak' WHERE slug = 'the-old-growth';
UPDATE odyssey_quest_definitions SET reward_relic = 'ancient-oak' WHERE slug = 'forest-riddle';

UPDATE odyssey_quest_definitions SET reward_relic = 'copper-gear' WHERE slug = 'clockwork-intro';
UPDATE odyssey_quest_definitions SET reward_relic = 'copper-gear' WHERE slug = 'clockwork-expedition';
UPDATE odyssey_quest_definitions SET reward_relic = 'clock-spring' WHERE slug = 'gear-hunt';
UPDATE odyssey_quest_definitions SET reward_relic = 'brass-compass' WHERE slug = 'the-copper-key';

UPDATE odyssey_quest_definitions SET reward_relic = 'star-dust' WHERE slug = 'star-observation';
UPDATE odyssey_quest_definitions SET reward_relic = 'moon-page' WHERE slug = 'constellation-map';
UPDATE odyssey_quest_definitions SET reward_relic = 'star-compass' WHERE slug = 'library-lore';

-- Update schema version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '33')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
