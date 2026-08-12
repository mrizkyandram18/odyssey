-- Migration 021: Phase 2A Idempotent Narrative Seed
-- Seeds demo data for Phase 2A Social Fabric: avatars, creative submissions, and reactions.
--
-- Prerequisites: migrations 001–020 must already be applied.
-- Safe to re-run: all inserts use ON CONFLICT ... DO UPDATE / DO NOTHING.
--
-- Narrative: The Starseekers crew (demo-crew-1) has been adventuring for a few days.
-- Each member has a distinct avatar, has written journal memories after missions,
-- and reacted warmly to each other's entries.

-- ============================================================
-- 1. Avatar Seeds for Demo Profiles
-- Each explorer gets a unique, distinctive avatar seed.
-- Idempotent: updates avatar fields if profiles already exist.
-- ============================================================

UPDATE odyssey_user_profiles
SET avatar_style = 'adventurer', avatar_seed = 'leo-starseeker'
WHERE uid = 'demo-uid-1' AND avatar_seed = 'default-seed';

UPDATE odyssey_user_profiles
SET avatar_style = 'adventurer', avatar_seed = 'maya-guide'
WHERE uid = 'demo-uid-2' AND avatar_seed = 'default-seed';

UPDATE odyssey_user_profiles
SET avatar_style = 'adventurer', avatar_seed = 'sam-builder'
WHERE uid = 'demo-uid-3' AND avatar_seed = 'default-seed';

-- ============================================================
-- 2. Creative Submissions (Journal Entries) — Phase 2A targets
-- These are the entries that reactions will target (target_type='JOURNAL').
-- Uses explicit IDs starting at 1001 to avoid collision with test fixtures.
-- ON CONFLICT (id) DO UPDATE to make re-runs safe.
-- ============================================================

INSERT INTO odyssey_creative_submissions
  (id, mission_id, exercise_id, family_id, author_uid, kind, content, status, created_at, updated_at)
VALUES
  -- Mission 101 (Morning Light) journal entries
  (1001, 101, (SELECT id FROM odyssey_exercises WHERE mission_id=101 AND slug='find-the-dew' LIMIT 1),
   'demo-crew-1', 'demo-uid-1', 'STORY',
   'I found a dewdrop on a spiderweb outside our front door. It looked exactly like a tiny diamond — perfectly round and trembling in the breeze. When the first ray of sunlight hit it, it vanished.',
   'APPROVED',
   timezone('utc'::text, now() - interval '2 days'),
   timezone('utc'::text, now() - interval '2 days')),

  (1002, 101, (SELECT id FROM odyssey_exercises WHERE mission_id=101 AND slug='morning-fact' LIMIT 1),
   'demo-crew-1', 'demo-uid-2', 'STORY',
   'Morning sunlight takes 8 minutes and 20 seconds to travel from the sun to Earth. I always thought it was instant! If the sun vanished, we would not even know for over eight minutes.',
   'APPROVED',
   timezone('utc'::text, now() - interval '2 days'),
   timezone('utc'::text, now() - interval '2 days')),

  -- Mission 102 (Gather Herbs) journal entry
  (1003, 102, (SELECT id FROM odyssey_exercises WHERE mission_id=102 AND slug='spot-the-green' LIMIT 1),
   'demo-crew-1', 'demo-uid-3', 'STORY',
   'I spotted three shades of green: the deep forest green of our basil plant, the bright lime of new grass, and the silvery-grey-green of our olive tree. The basil smells like summer.',
   'APPROVED',
   timezone('utc'::text, now() - interval '1 day'),
   timezone('utc'::text, now() - interval '1 day')),

  -- Extra memory entries to populate the Family Journal timeline
  (1004, 101, (SELECT id FROM odyssey_exercises WHERE mission_id=101 AND slug='find-the-dew' LIMIT 1),
   'demo-crew-1', 'demo-uid-1', 'STORY',
   'I also noticed the shadow of our door stretching across the carpet like a long, thin dragon. The morning has so many secrets.',
   'APPROVED',
   timezone('utc'::text, now() - interval '4 hours'),
   timezone('utc'::text, now() - interval '4 hours')),

  (1005, 101, (SELECT id FROM odyssey_exercises WHERE mission_id=101 AND slug='morning-fact' LIMIT 1),
   'demo-crew-1', 'demo-uid-2', 'STORY',
   'If trees could walk, I bet they would move very slowly — like giant wooden tortoises. They would probably head toward the light, just like we do on a good morning.',
   'APPROVED',
   timezone('utc'::text, now() - interval '1 hour'),
   timezone('utc'::text, now() - interval '1 hour'))
ON CONFLICT (id) DO UPDATE SET
  content    = EXCLUDED.content,
  status     = EXCLUDED.status,
  updated_at = timezone('utc'::text, now());

-- ============================================================
-- 3. Reactions on Journal entries (new schema: family_id, target_type, target_id, actor_uid)
-- Narrative: crew members react warmly and specifically to each other's memories.
-- Idempotent via unique constraint (family_id, target_type, target_id, actor_uid).
-- ============================================================

INSERT INTO odyssey_reactions (family_id, target_type, target_id, actor_uid, reaction_type, created_at)
VALUES
  -- Maya and Sam heart Leo's dewdrop memory
  ('demo-crew-1', 'JOURNAL', 1001, 'demo-uid-2', 'HEART', timezone('utc'::text, now() - interval '47 hours')),
  ('demo-crew-1', 'JOURNAL', 1001, 'demo-uid-3', 'CLAP',  timezone('utc'::text, now() - interval '46 hours')),

  -- Leo and Sam react to Maya's sunlight fact
  ('demo-crew-1', 'JOURNAL', 1002, 'demo-uid-1', 'STAR',  timezone('utc'::text, now() - interval '45 hours')),
  ('demo-crew-1', 'JOURNAL', 1002, 'demo-uid-3', 'HEART', timezone('utc'::text, now() - interval '44 hours')),

  -- Leo and Maya react to Sam's herb colours memory
  ('demo-crew-1', 'JOURNAL', 1003, 'demo-uid-1', 'HEART', timezone('utc'::text, now() - interval '22 hours')),
  ('demo-crew-1', 'JOURNAL', 1003, 'demo-uid-2', 'STAR',  timezone('utc'::text, now() - interval '21 hours')),

  -- Reactions on the two newer memories
  ('demo-crew-1', 'JOURNAL', 1004, 'demo-uid-2', 'CLAP',  timezone('utc'::text, now() - interval '3 hours')),
  ('demo-crew-1', 'JOURNAL', 1005, 'demo-uid-1', 'HEART', timezone('utc'::text, now() - interval '30 minutes')),
  ('demo-crew-1', 'JOURNAL', 1005, 'demo-uid-3', 'CLAP',  timezone('utc'::text, now() - interval '25 minutes'))
ON CONFLICT (family_id, target_type, target_id, actor_uid)
DO UPDATE SET
  reaction_type = EXCLUDED.reaction_type;

-- ============================================================
-- 4. Update schema version
-- ============================================================
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '021_seed_phase2a')
ON CONFLICT (key) DO UPDATE SET
  value      = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
