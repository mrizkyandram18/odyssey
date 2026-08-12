-- Migration 022: Phase 1 MVP — 6 Whispering Woods demo missions
-- --------------------------------------------------------------------------
-- Ensures demo-crew-1 has exactly 6 playable missions in whispering-woods with
-- SOLO + RELAY + CREATIVE variety. Non-MVP demo quest instances are removed.
-- Definitions for future realms (010) are left intact for later phases.
--
-- Safe to re-run: writes are idempotent (ON CONFLICT / IF NOT EXISTS).
-- Prerequisites: migrations 001-021 must already be applied.

-- ============================================================
-- 1. Unique index on exercises(mission_id, slug)
-- ============================================================
CREATE UNIQUE INDEX IF NOT EXISTS uq_odyssey_exercises_mission_slug
    ON odyssey_exercises (mission_id, slug);

-- ============================================================
-- 2. Mission type variety for the MVP six (definitions from 010)
-- ============================================================
UPDATE odyssey_quest_definitions
SET quest_type = 'SOLO'
WHERE slug IN (
    'morning-light',
    'gather-herbs',
    'riddle-of-the-stones',
    'forest-riddle'
);

UPDATE odyssey_quest_definitions
SET quest_type = 'RELAY'
WHERE slug = 'shadow-trail';

UPDATE odyssey_quest_definitions
SET quest_type = 'CREATIVE'
WHERE slug = 'the-old-growth';

-- ============================================================
-- 3. Drop non-MVP demo quest instances (other realms / ad-hoc seed)
--    Exercises cascade via FK ON DELETE CASCADE.
-- ============================================================
DELETE FROM odyssey_missions
WHERE family_id = 'demo-crew-1'
  AND template_slug NOT IN (
       'morning-light',
       'gather-herbs',
       'riddle-of-the-stones',
       'shadow-trail',
       'the-old-growth',
       'forest-riddle'
  );

-- ============================================================
-- 4. Drop stray exercises that do not belong to the MVP six
-- ============================================================
DELETE FROM odyssey_exercises
WHERE mission_id IN (
        SELECT id FROM odyssey_missions
        WHERE family_id = 'demo-crew-1'
          AND template_slug NOT IN (
              'morning-light',
              'gather-herbs',
              'riddle-of-the-stones',
              'shadow-trail',
              'the-old-growth',
              'forest-riddle'
          )
      )
   OR (
        mission_id IN (SELECT id FROM odyssey_missions WHERE family_id = 'demo-crew-1')
        AND slug NOT IN (
           'find-the-dew', 'morning-fact',
           'spot-the-green', 'herb-concept',
           'stone-shape', 'solve-riddle',
           'trace-shadow', 'shadow-story',
           'draw-tree', 'tree-history',
           'riddle-solve', 'find-marker'
        )
   );

-- ============================================================
-- 5. Exactly 6 Whispering Woods quest instances for demo-crew-1
--    All PENDING so a family can play through every quest.
-- ============================================================
INSERT INTO odyssey_missions
    (id, family_id, template_slug, title, course, status, started_at, completed_at)
VALUES
    (101, 'demo-crew-1', 'morning-light',        'Morning Light',        'the-awakening',  'PENDING', NULL, NULL),
    (102, 'demo-crew-1', 'gather-herbs',         'Gather Herbs',         'the-awakening',  'PENDING', NULL, NULL),
    (103, 'demo-crew-1', 'riddle-of-the-stones', 'Riddle of the Stones', 'the-awakening',  'PENDING', NULL, NULL),
    (104, 'demo-crew-1', 'shadow-trail',         'Shadow Trail',         'the-deep-woods', 'PENDING', NULL, NULL),
    (105, 'demo-crew-1', 'the-old-growth',       'The Old Growth',       'the-deep-woods', 'PENDING', NULL, NULL),
    (106, 'demo-crew-1', 'forest-riddle',        'Forest Riddle',        'the-deep-woods', 'PENDING', NULL, NULL)
ON CONFLICT (id) DO UPDATE SET
    family_id       = EXCLUDED.family_id,
    template_slug = EXCLUDED.template_slug,
    title         = EXCLUDED.title,
    course       = EXCLUDED.course,
    status        = EXCLUDED.status,
    started_at    = EXCLUDED.started_at,
    completed_at  = EXCLUDED.completed_at;

-- ============================================================
-- 6. Canonical exercises (2 each) — all PENDING for full playthrough
-- ============================================================
INSERT INTO odyssey_exercises (mission_id, slug, description, status)
VALUES
    (101, 'find-the-dew',   'Find something glistening outside your door and describe it.',            'PENDING'),
    (101, 'morning-fact',   'Look up one fact about morning sunlight and share it.',                     'PENDING'),
    (102, 'spot-the-green', 'Point out three shades of green you can see right now.',                    'PENDING'),
    (102, 'herb-concept',      'Name one use for a common houseplant.',                                     'PENDING'),
    (103, 'stone-shape',    'Find a stone or brick and describe its shape.',                             'PENDING'),
    (103, 'solve-riddle',   'Solve: I have no voice, yet I answer every question. What am I?',         'PENDING'),
    (104, 'trace-shadow',   'Trace the shape of a shadow on the ground and describe it.',                'PENDING'),
    (104, 'shadow-story',   'Invent a short story about where the shadow leads.',                        'PENDING'),
    (105, 'draw-tree',      'Draw or describe the oldest tree you can see.',                             'PENDING'),
    (105, 'tree-history',   'Write a short paragraph about what this tree has witnessed.',               'PENDING'),
    (106, 'riddle-solve',   'Solve the riddle: I am always hungry, I must always be fed. The finger I touch will soon turn red. What am I?', 'PENDING'),
    (106, 'find-marker',    'Find a natural marker (stone, stick, leaf) that matches the riddle answer.', 'PENDING')
ON CONFLICT (mission_id, slug) DO UPDATE SET
    description = EXCLUDED.description,
    status      = EXCLUDED.status,
    completed_by = NULL,
    completed_at = NULL;

-- ============================================================
-- 7. Journey progress: only Whispering Woods is playable for MVP demo
-- ============================================================
INSERT INTO odyssey_journey_progress (family_id, journey, status, progress, last_unlocked_at)
VALUES
    ('demo-crew-1', 'whispering-woods', 'ACTIVE', 0, timezone('utc'::text, now())),
    ('demo-crew-1', 'clockwork-city',   'LOCKED', 0, NULL),
    ('demo-crew-1', 'starlit-library',  'LOCKED', 0, NULL)
ON CONFLICT (family_id, journey) DO UPDATE SET
    status           = EXCLUDED.status,
    progress         = EXCLUDED.progress,
    last_unlocked_at = EXCLUDED.last_unlocked_at
WHERE odyssey_journey_progress.journey IN ('clockwork-city', 'starlit-library')
   OR (odyssey_journey_progress.journey = 'whispering-woods'
       AND odyssey_journey_progress.status = 'LOCKED');

-- Heal whispering-woods to ACTIVE if somehow locked; do not clobber QA progress
-- above 0 unless status was LOCKED (handled above). Ensure ACTIVE at minimum:
UPDATE odyssey_journey_progress
SET status = 'ACTIVE',
    last_unlocked_at = COALESCE(last_unlocked_at, timezone('utc'::text, now()))
WHERE family_id = 'demo-crew-1'
  AND journey = 'whispering-woods'
  AND status <> 'COMPLETE';

-- ============================================================
-- 8. Schema version
-- ============================================================
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '022_mvp_six_missions')
ON CONFLICT (key) DO UPDATE SET
  value      = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
