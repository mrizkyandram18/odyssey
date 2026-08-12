-- Migration 011: Demo Data Seed
-- Seeds a realistic family crew to make the prototype feel alive immediately.

-- 1. Demo Family: The Starseekers
INSERT INTO odyssey_families (id, name)
VALUES ('demo-crew-1', 'The Starseekers')
ON CONFLICT (id) DO NOTHING;

-- 2. Demo Profiles
INSERT INTO odyssey_user_profiles (uid, family_id, explorer_name, role, level, xp)
VALUES
  ('demo-uid-1', 'demo-crew-1', 'Leo', 'SEEKER', 2, 150),
  ('demo-uid-2', 'demo-crew-1', 'Maya', 'GUIDE', 2, 175),
  ('demo-uid-3', 'demo-crew-1', 'Sam', 'BUILDER', 1, 90)
ON CONFLICT (uid) DO NOTHING;


-- 4. Journey Progress
INSERT INTO odyssey_journey_progress (family_id, journey, status, progress, last_unlocked_at)
VALUES ('demo-crew-1', 'whispering-woods', 'ACTIVE', 50, timezone('utc'::text, now()))
ON CONFLICT (family_id, journey) DO NOTHING;

-- 5. Missions (Realistic progression)
-- Mission 1: Completed (Morning Light)
INSERT INTO odyssey_missions (id, family_id, template_slug, title, status, started_at, completed_at)
VALUES (101, 'demo-crew-1', 'morning-light', 'Morning Light', 'DONE', timezone('utc'::text, now() - interval '2 days'), timezone('utc'::text, now() - interval '2 days'))
ON CONFLICT (id) DO NOTHING;

INSERT INTO odyssey_exercises (mission_id, slug, description, status, completed_by, completed_at)
VALUES
  (101, 'find-the-dew', 'Find something glistening outside your door and describe it.', 'DONE', 'demo-uid-1', timezone('utc'::text, now() - interval '2 days')),
  (101, 'morning-fact', 'Look up one fact about morning sunlight and share it.', 'DONE', 'demo-uid-2', timezone('utc'::text, now() - interval '2 days'))
ON CONFLICT (id) DO NOTHING;

-- Mission 2: Active (Gather Herbs)
INSERT INTO odyssey_missions (id, family_id, template_slug, title, status, started_at)
VALUES (102, 'demo-crew-1', 'gather-herbs', 'Gather Herbs', 'ACTIVE', timezone('utc'::text, now() - interval '1 day'))
ON CONFLICT (id) DO NOTHING;

INSERT INTO odyssey_exercises (mission_id, slug, description, status, completed_by, completed_at)
VALUES
  (102, 'spot-the-green', 'Point out three shades of green you can see right now.', 'DONE', 'demo-uid-3', timezone('utc'::text, now() - interval '1 day')),
  (102, 'herb-concept', 'Name one use for a common houseplant.', 'PENDING', NULL, NULL)
ON CONFLICT (id) DO NOTHING;

-- Mission 3: Active (Riddle of the Stones)
INSERT INTO odyssey_missions (id, family_id, template_slug, title, status, started_at)
VALUES (103, 'demo-crew-1', 'riddle-of-the-stones', 'Riddle of the Stones', 'ACTIVE', timezone('utc'::text, now()))
ON CONFLICT (id) DO NOTHING;

INSERT INTO odyssey_exercises (mission_id, slug, description, status)
VALUES
  (103, 'stone-shape', 'Find a stone or brick and describe its shape.', 'PENDING'),
  (103, 'solve-riddle', 'Solve: I have no voice, yet I answer every question. What am I?', 'PENDING')
ON CONFLICT (id) DO NOTHING;

-- 6. Creative Memories (Journal/Gallery)
INSERT INTO odyssey_creative_items (family_id, journey, author_uid, kind, payload, created_at)
VALUES
  ('demo-crew-1', 'whispering-woods', 'demo-uid-1', 'STORY', 'I found a dewdrop on a spiderweb that looked exactly like a tiny diamond. It disappeared when the sun hit it fully.', timezone('utc'::text, now() - interval '2 days')),
  ('demo-crew-1', 'whispering-woods', 'demo-uid-2', 'STORY', 'Morning sunlight takes 8 minutes and 20 seconds to travel from the sun to the earth. I thought it was instant!', timezone('utc'::text, now() - interval '2 days')),
  ('demo-crew-1', 'whispering-woods', 'demo-uid-3', 'STORY', 'The basil plant in our kitchen smells like summer. I think it counts as a forest herb.', timezone('utc'::text, now() - interval '1 day')),
  ('demo-crew-1', 'whispering-woods', 'demo-uid-1', 'STORY', 'I saw a shadow stretching across the living room carpet that looked like a long, thin dragon.', timezone('utc'::text, now() - interval '4 hours')),
  ('demo-crew-1', 'whispering-woods', 'demo-uid-2', 'STORY', 'If trees could walk, I bet they would move very slowly, like giant wooden tortoises.', timezone('utc'::text, now() - interval '1 hour'))
ON CONFLICT (id) DO NOTHING;


-- 8. Daily Turns (Streak Logic)
INSERT INTO odyssey_daily_missions (uid, date, mission_slug, completed, created_at)
VALUES
  ('demo-uid-1', current_date - interval '2 days', 'morning-light', true, timezone('utc'::text, now() - interval '2 days')),
  ('demo-uid-1', current_date - interval '1 day', 'gather-herbs', true, timezone('utc'::text, now() - interval '1 day')),
  ('demo-uid-2', current_date - interval '2 days', 'morning-light', true, timezone('utc'::text, now() - interval '2 days')),
  ('demo-uid-2', current_date - interval '1 day', 'gather-herbs', true, timezone('utc'::text, now() - interval '1 day')),
  ('demo-uid-3', current_date - interval '2 days', 'morning-light', true, timezone('utc'::text, now() - interval '2 days')),
  ('demo-uid-3', current_date - interval '1 day', 'gather-herbs', false, timezone('utc'::text, now() - interval '1 day'))
ON CONFLICT (id) DO NOTHING;

-- 9. Rewards (Collections & Gifts)
INSERT INTO odyssey_collections (uid, code, awarded_at)
VALUES
  ('demo-uid-1', 'acorn-shard', timezone('utc'::text, now() - interval '2 days')),
  ('demo-uid-2', 'whispering-leaf', timezone('utc'::text, now() - interval '1 day'))
ON CONFLICT (id) DO NOTHING;

INSERT INTO odyssey_gifts (uid, source, opened, opened_at)
VALUES
  ('demo-uid-1', 'LEVEL_UP', true, timezone('utc'::text, now() - interval '1 day')),
  ('demo-uid-2', 'QUEST', false, NULL),
  ('demo-uid-3', 'LEVEL_UP', true, timezone('utc'::text, now() - interval '12 hours'))
ON CONFLICT (id) DO NOTHING;

-- Update schema version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '11')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
