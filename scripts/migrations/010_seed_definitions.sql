-- Migration 010: Seed Playable Content Definitions
-- Seeds every definition table with minimum playable content.
-- All content is published (published=true) and non-deleted.
-- Uses ON CONFLICT DO NOTHING for idempotent execution.
--
-- Prerequisites: migrations 001-009 must be applied first.

-- ============================================================
-- odyssey_journey_definitions
-- ============================================================
INSERT INTO odyssey_journey_definitions (slug, name, description, "order", max_progress, icon, published, version)
VALUES
  ('whispering-woods', 'Whispering Woods', 'A gentle forest journey of moss, light, and quiet secrets.', 1, 100, '🌲', true, 1),
  ('clockwork-city', 'Clockwork City', 'A bustling city of gears, steam, and copper curiosities.', 2, 100, '⚙️', true, 1),
  ('starlit-library', 'Starlit Library', 'A vast library floating among the stars, filled with ancient knowledge.', 3, 100, '📚', true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_course_definitions
-- ============================================================
INSERT INTO odyssey_course_definitions (slug, journey, title, description, "order", published, version)
VALUES
  ('the-awakening', 'whispering-woods', 'The Awakening', 'The forest wakes around you. Learn to see what others miss.', 1, true, 1),
  ('the-deep-woods', 'whispering-woods', 'The Deep Woods', 'Venture deeper into the forest where shadows hold stories.', 2, true, 1),
  ('gears-and-gold', 'clockwork-city', 'Gears and Gold', 'Discover the clockwork heart of the city and its hidden treasures.', 1, true, 1),
  ('first-stars', 'starlit-library', 'First Stars', 'The library opens its doors to those who look upward.', 1, true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_quest_definitions
-- ============================================================
INSERT INTO odyssey_quest_definitions (slug, journey, course, title, description, quest_type, challenge_defs, reward_xp, reward_chest, is_mandatory, required_level, published, version)
VALUES
  -- Journey 1, Course 1: The Awakening
  ('morning-light', 'whispering-woods', 'the-awakening', 'Morning Light', 'Find something glistening outside your door and share what you discovered.', 'SOLO', '[{"slug":"find-the-dew","description":"Find something glistening outside your door and describe it.","type":"OBSERVATION"},{"slug":"morning-fact","description":"Look up one fact about morning sunlight and share it.","type":"RESEARCH"}]', 80, 'wooden-chest', true, 0, true, 1),
  ('gather-herbs', 'whispering-woods', 'the-awakening', 'Gather Herbs', 'Point out three shades of green and name one use for a common houseplant.', 'SOLO', '[{"slug":"spot-the-green","description":"Point out three shades of green you can see right now.","type":"OBSERVATION"},{"slug":"herb-concept","description":"Name one use for a common houseplant.","type":"RESEARCH"}]', 80, 'wooden-chest', true, 0, true, 1),
  ('riddle-of-the-stones', 'whispering-woods', 'the-awakening', 'Riddle of the Stones', 'Find a stone or brick and solve the riddle it poses.', 'SOLO', '[{"slug":"stone-shape","description":"Find a stone or brick and describe its shape.","type":"OBSERVATION"},{"slug":"solve-riddle","description":"Solve: I have no voice, yet I answer every question. What am I?","type":"PUZZLE"}]', 100, 'bronze-chest', true, 0, true, 1),

  -- Journey 1, Course 2: The Deep Woods
  ('shadow-trail', 'whispering-woods', 'the-deep-woods', 'Shadow Trail', 'Follow the shadow that appears when the light shifts.', 'SOLO', '[{"slug":"trace-shadow","description":"Trace the shape of a shadow on the ground and describe it.","type":"OBSERVATION"},{"slug":"shadow-story","description":"Invent a short story about where the shadow leads.","type":"WRITE"}]', 100, 'bronze-chest', true, 0, true, 1),
  ('the-old-growth', 'whispering-woods', 'the-deep-woods', 'The Old Growth', 'Draw the oldest tree you can find and write about its history.', 'SOLO', '[{"slug":"draw-tree","description":"Draw or describe the oldest tree you can see.","type":"DRAW"},{"slug":"tree-history","description":"Write a short paragraph about what this tree has witnessed.","type":"WRITE"}]', 120, 'silver-chest', true, 0, true, 1),
  ('forest-riddle', 'whispering-woods', 'the-deep-woods', 'Forest Riddle', 'Solve the forest riddle and find the hidden marker.', 'SOLO', '[{"slug":"riddle-solve","description":"Solve the riddle: I am always hungry, I must always be fed. The finger I touch will soon turn red. What am I?","type":"PUZZLE"},{"slug":"find-marker","description":"Find a natural marker (stone, stick, leaf) that matches the riddle answer.","type":"OBSERVATION"}]', 120, 'silver-chest', true, 0, true, 1),

  -- Journey 2, Course 1: Gears and Gold
  ('clockwork-intro', 'clockwork-city', 'gears-and-gold', 'Clockwork Introduction', 'Observe the mechanical world around you and note three gears or springs.', 'SOLO', '[{"slug":"find-gears","description":"Find three mechanical objects with gears or springs.","type":"OBSERVATION"},{"slug":"gear-fact","description":"Research one fact about how clocks or gears work.","type":"RESEARCH"}]', 100, 'bronze-chest', true, 0, true, 1),
  ('gear-hunt', 'clockwork-city', 'gears-and-gold', 'Gear Hunt', 'Search for a copper or brass object and describe its shape.', 'SOLO', '[{"slug":"copper-find","description":"Find a copper or brass object and describe its shape and texture.","type":"OBSERVATION"},{"slug":"gear-puzzle","description":"Solve a simple mechanical puzzle or riddle about gears.","type":"PUZZLE"}]', 120, 'silver-chest', true, 0, true, 1),
  ('the-copper-key', 'clockwork-city', 'gears-and-gold', 'The Copper Key', 'Write a story about finding a key that opens a clockwork door.', 'SOLO', '[{"slug":"key-story","description":"Write a short story about finding a mysterious copper key.","type":"WRITE"},{"slug":"key-research","description":"Research one fact about locks or keys in history.","type":"RESEARCH"}]', 140, 'silver-chest', true, 0, true, 1),
  ('clockwork-expedition', 'clockwork-city', 'gears-and-gold', 'Clockwork Expedition', 'Walk 100 steps or observe a sundial shadow angle.', 'SOLO', '[{"slug":"step-count","description":"Walk 100 steps or observe the shadow angle of a clock tower.","type":"MOVEMENT"},{"slug":"gear-observation","description":"Note three rotating mechanical objects you found.","type":"OBSERVATION"}]', 110, 'bronze-chest', true, 0, true, 1),

  -- Journey 3, Course 1: First Stars
  ('star-observation', 'starlit-library', 'first-stars', 'Star Observation', 'Look at the sky and identify one constellation or bright star.', 'SOLO', '[{"slug":"find-constellation","description":"Find and name one constellation or bright star in the night sky.","type":"OBSERVATION"},{"slug":"star-fact","description":"Research one fact about the star or constellation you found.","type":"RESEARCH"}]', 100, 'bronze-chest', true, 0, true, 1),
  ('constellation-map', 'starlit-library', 'first-stars', 'Constellation Map', 'Draw your own constellation map and write the story behind it.', 'SOLO', '[{"slug":"draw-map","description":"Draw a map of the stars you can see, connecting them into a pattern.","type":"DRAW"},{"slug":"map-story","description":"Write the story of the constellation you invented.","type":"WRITE"}]', 120, 'silver-chest', true, 0, true, 1),
  ('library-concept', 'starlit-library', 'first-stars', 'Library Concept', 'Research a famous library from history and share what you learned.', 'SOLO', '[{"slug":"library-research","description":"Research one famous library from history (Alexandria, etc.).","type":"RESEARCH"},{"slug":"library-reflection","description":"Write a short reflection on why libraries matter.","type":"WRITE"}]', 140, 'golden-chest', true, 0, true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_creative_prompt_definitions
-- ============================================================
INSERT INTO odyssey_creative_prompt_definitions (slug, journey, title, description, prompt, kind, published, version)
VALUES
  ('whisper-story', 'whispering-woods', 'Whisper Story', 'Write a short story about a secret the forest told you.', 'The trees are whispering. What secret do they share with you? Write a short story (50-200 words) about what the forest reveals.', 'STORY', true, 1),
  ('forest-drawing', 'whispering-woods', 'Forest Drawing', 'Draw what you imagine lives between the trees.', 'Close your eyes and imagine what lives between the trees in the Whispering Woods. Draw it.', 'DRAW', true, 1),
  ('clockwork-story', 'clockwork-city', 'Clockwork Story', 'Write a story about a machine that thinks for itself.', 'In Clockwork City, a machine wakes up and starts thinking. What does it think about? Write a short story.', 'STORY', true, 1),
  ('gear-drawing', 'clockwork-city', 'Gear Drawing', 'Draw your ideal clockwork invention.', 'If you could build one clockwork invention, what would it do? Draw it and label the parts.', 'DRAW', true, 1),
  ('starlit-story', 'starlit-library', 'Starlit Story', 'Write about a book that writes itself.', 'In the Starlit Library, books write themselves overnight. What story appears on the page when you open one?', 'STORY', true, 1),
  ('constellation-drawing', 'starlit-library', 'Constellation Drawing', 'Draw a new constellation and give it a name.', 'The stars are waiting for someone to connect them into a new shape. Draw your constellation and name it.', 'DRAW', true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_achievement_definitions
-- ============================================================
INSERT INTO odyssey_achievement_definitions (code, title, description, kind, trigger, threshold, reward_xp, reward_relic, published, version)
VALUES
  ('FIRST_QUEST', 'First Mission', 'Complete your first quest.', 'PERSONAL', 'QUEST_COMPLETED', 1, 50, '', true, 1),
  ('CHAPTER_COMPLETE', 'Course Complete', 'Finish all missions in a course.', 'GROUP', 'CHAPTER_COMPLETED', 1, 100, '', true, 1),
  ('REALM_COMPLETE', 'Journey Complete', 'Complete all chapters in a journey.', 'GROUP', 'REALM_COMPLETED', 1, 200, '', true, 1),
  ('DAILY_STREAK_3', 'Three-Day Streak', 'Complete daily turns for 3 consecutive days.', 'PERSONAL', 'DAILY_STREAK', 3, 30, '', true, 1),
  ('DAILY_STREAK_7', 'Seven-Day Streak', 'Complete daily turns for 7 consecutive days.', 'PERSONAL', 'DAILY_STREAK', 7, 70, '', true, 1),
  ('FIRST_CHEST', 'First Gift', 'Open your first chest.', 'PERSONAL', 'CHEST_OPENED', 1, 25, '', true, 1),
  ('FIRST_RELIC', 'First Collection', 'Collect your first relic.', 'PERSONAL', 'RELIC_COLLECTED', 1, 25, '', true, 1),
  ('EXPLORER_LEVEL_3', 'Explorer Level 3', 'Reach Explorer Level 3.', 'PERSONAL', 'LEVEL_REACHED', 3, 50, '', true, 1),
  ('CREATIVE_FIRST', 'Creative Spark', 'Submit your first creative work.', 'PERSONAL', 'CREATIVE_SUBMISSION', 1, 30, '', true, 1),
  ('QUEST_MASTER', 'Mission Master', 'Complete 10 missions.', 'PERSONAL', 'QUEST_COMPLETED', 10, 100, '', true, 1)
ON CONFLICT (code) DO NOTHING;

-- ============================================================
-- odyssey_concept_definitions
-- ============================================================
INSERT INTO odyssey_concept_definitions (slug, journey, course, title, content, "order", published, version)
VALUES
  ('forest-origin', 'whispering-woods', 'the-awakening', 'The Origin of the Whispering Woods', 'Long ago, the first trees grew from seeds carried on the wind. Each tree holds a memory, and when the wind blows through the leaves, it whispers those memories to anyone who listens.', 1, true, 1),
  ('forest-creatures', 'whispering-woods', 'the-awakening', 'Creatures of the Forest', 'The Whispering Woods is home to gentle creatures: moss-foxes who glow in the dark, song-birds that remember every melody ever sung, and stone-turtles that carry entire ecosystems on their backs.', 2, true, 1),
  ('forest-deep', 'whispering-woods', 'the-deep-woods', 'Into the Deep Woods', 'Beyond the clearing, the forest grows older. The trees here are so tall their tops disappear into the clouds. The shadows here are not dark — they are full of stories waiting to be told.', 1, true, 1),
  ('forest-ancient', 'whispering-woods', 'the-deep-woods', 'The Ancient Oak', 'At the heart of the Deep Woods stands an oak tree so old it remembers the world before names. Its roots drink from underground rivers that flow with starlight.', 2, true, 1),
  ('city-origin', 'clockwork-city', 'gears-and-gold', 'The Birth of Clockwork City', 'Clockwork City was built by inventors who believed that every problem could be solved with a gear, a spring, or a clever idea. The first gear they ever made is still turning in the city center.', 1, true, 1),
  ('city-gold', 'clockwork-city', 'gears-and-gold', 'The Gold Beneath the Gears', 'Beneath the cobblestone streets, veins of copper and brass run through the earth. The city glows at dusk because the metal absorbs the last light of the sun and releases it slowly through the night.', 2, true, 1),
  ('library-origin', 'starlit-library', 'first-stars', 'The Founding of the Starlit Library', 'The Starlit Library was built by the first astronomers, who believed that knowledge should be as vast as the sky. Every book in the library contains a star-map that leads to a new discovery.', 1, true, 1),
  ('library-secrets', 'starlit-library', 'first-stars', 'Secrets of the Shelves', 'The highest shelves in the Starlit Library hold books that have not yet been written. They wait for someone curious enough to climb up and read the blank pages — which fill themselves when touched.', 2, true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_chest_definitions
-- ============================================================
INSERT INTO odyssey_chest_definitions (slug, name, rarity, icon, description, published, version)
VALUES
  ('wooden-chest', 'Wooden Gift', 'COMMON', '📦', 'A simple wooden chest, worn by time but still holding surprises.', true, 1),
  ('bronze-chest', 'Bronze Gift', 'UNCOMMON', '🟤', 'A sturdy bronze chest with ornate fittings and a satisfying click.', true, 1),
  ('silver-chest', 'Silver Gift', 'RARE', '⚪', 'A polished silver chest that gleams in the dark. Rare and rewarding.', true, 1),
  ('golden-chest', 'Golden Gift', 'EPIC', '🟡', 'A magnificent golden chest, warm to the touch and heavy with reward.', true, 1),
  ('mystic-chest', 'Mystic Gift', 'LEGENDARY', '🔮', 'An otherworldly chest that hums with arcane energy. The rarest find.', true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_drop_tables
-- ============================================================
INSERT INTO odyssey_drop_tables (gift_slug, rarity, weight)
VALUES
  ('wooden-chest', 'COMMON', 0.70),
  ('wooden-chest', 'UNCOMMON', 0.25),
  ('wooden-chest', 'RARE', 0.05),
  ('bronze-chest', 'COMMON', 0.50),
  ('bronze-chest', 'UNCOMMON', 0.35),
  ('bronze-chest', 'RARE', 0.12),
  ('bronze-chest', 'EPIC', 0.03),
  ('silver-chest', 'COMMON', 0.30),
  ('silver-chest', 'UNCOMMON', 0.35),
  ('silver-chest', 'RARE', 0.25),
  ('silver-chest', 'EPIC', 0.08),
  ('silver-chest', 'LEGENDARY', 0.02),
  ('golden-chest', 'COMMON', 0.15),
  ('golden-chest', 'UNCOMMON', 0.25),
  ('golden-chest', 'RARE', 0.30),
  ('golden-chest', 'EPIC', 0.22),
  ('golden-chest', 'LEGENDARY', 0.08),
  ('mystic-chest', 'COMMON', 0.05),
  ('mystic-chest', 'UNCOMMON', 0.15),
  ('mystic-chest', 'RARE', 0.25),
  ('mystic-chest', 'EPIC', 0.30),
  ('mystic-chest', 'LEGENDARY', 0.25)
ON CONFLICT (gift_slug, rarity) DO NOTHING;

-- ============================================================
-- odyssey_relic_definitions
-- ============================================================
INSERT INTO odyssey_relic_definitions (slug, name, description, journey, rarity, image, concept, published, version)
VALUES
  ('acorn-shard', 'Acorn Shard', 'A tiny piece of an acorn that still glows with warmth.', 'whispering-woods', 'COMMON', '🌰', 'Found in the Whispering Woods. A reminder that great things grow from small beginnings.', true, 1),
  ('copper-gear', 'Copper Gear', 'A small copper gear that still turns when held.', 'clockwork-city', 'COMMON', '⚙️', 'From the workshops of Clockwork City. It turns in your palm like a tiny sun.', true, 1),
  ('star-dust', 'Star Dust', 'A pinch of shimmering dust that fell from the night sky.', 'starlit-library', 'COMMON', '✨', 'Collected from the library rooftop on a clear night. It sparkles faintly in moonlight.', true, 1),
  ('whispering-leaf', 'Whispering Leaf', 'A leaf that seems to murmur when you hold it close.', 'whispering-woods', 'UNCOMMON', '🍃', 'The trees of the Whispering Woods share their secrets through this leaf.', true, 1),
  ('clock-spring', 'Clock Spring', 'A tightly wound spring from an ancient clockwork mechanism.', 'clockwork-city', 'UNCOMMON', '🔩', 'Still wound after centuries. It remembers the time it was made.', true, 1),
  ('moon-page', 'Moon Page', 'A page from a book that only appears under moonlight.', 'starlit-library', 'UNCOMMON', '🌙', 'The ink on this page shifts with the phases of the moon.', true, 1),
  ('ancient-oak', 'Ancient Oak', 'A splinter of the oldest tree in the Whispering Woods.', 'whispering-woods', 'RARE', '🌳', 'This splinter has witnessed a thousand seasons. It hums with quiet power.', true, 1),
  ('brass-compass', 'Brass Compass', 'A compass that always points toward the nearest wonder.', 'clockwork-city', 'RARE', '🧭', 'Its needle spins until it finds something worth discovering.', true, 1),
  ('celestial-map', 'Celestial Map', 'A hand-drawn map of stars that no astronomer has charted.', 'starlit-library', 'RARE', '🗺️', 'The map reveals constellations that exist only in imagination.', true, 1),
  ('journey-totem', 'Journey Totem', 'A carved totem that represents the heart of a journey.', 'whispering-woods', 'EPIC', '🏺', 'Carved from the heartwood of the world-tree. It resonates with the spirit of the forest.', true, 1),
  ('time-gear', 'Time Gear', 'A gear that seems to tick in reverse.', 'clockwork-city', 'EPIC', '⏳', 'When held, it makes the holder feel the passage of time in their fingertips.', true, 1),
  ('star-compass', 'Star Compass', 'A compass that points not north, but toward the nearest star.', 'starlit-library', 'EPIC', '⭐', 'Its needle always points to the brightest star in the sky.', true, 1),
  ('world-tree-seed', 'World Tree Seed', 'A seed from the tree at the center of all realms.', 'whispering-woods', 'LEGENDARY', '🌰', 'Legend says that planting this seed will grow a tree that connects all worlds.', true, 1),
  ('cosmic-clock', 'Cosmic Clock', 'A clock that measures not hours, but stories.', 'clockwork-city', 'LEGENDARY', '🕰️', 'Its hands move whenever a new story begins somewhere in the world.', true, 1),
  ('infinity-library', 'Infinity Library', 'A tiny book that contains an infinite number of pages.', 'starlit-library', 'LEGENDARY', '📖', 'Every time you turn a page, a new story appears. No two readers see the same tale.', true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_season_definitions
-- ============================================================
INSERT INTO odyssey_season_definitions (slug, name, description, start_at, end_at, journey, published, version)
VALUES
  ('season-spring-2026', 'Spring 2026', 'The first season of Odyssey. Explore the Whispering Woods and discover the magic of the forest.', '2026-01-01T00:00:00Z', '2026-12-31T23:59:59Z', 'whispering-woods', true, 1),
  ('season-autumn-2026', 'Autumn 2026', 'Autumn arrives in Clockwork City. Unravel the ticking mysteries of gears and steam.', '2026-09-01T00:00:00Z', '2026-11-30T23:59:59Z', 'clockwork-city', true, 1)
ON CONFLICT (slug) DO NOTHING;

-- ============================================================
-- odyssey_balance_configs
-- ============================================================
INSERT INTO odyssey_balance_configs (key, value, updated_by)
VALUES
  ('xp_per_level', '100'::jsonb, 'system'),
  ('challenge_xp', '20'::jsonb, 'system'),
  ('completion_bonus_xp', '60'::jsonb, 'system'),
  ('drop_rate_multiplier', '100'::jsonb, 'system'),
  ('daily_mission_xp', '10'::jsonb, 'system'),
  ('journey_progress_per_quest', '25'::jsonb, 'system'),
  ('journey_completion_threshold', '100'::jsonb, 'system'),
  ('achievement_threshold_multiplier', '100'::jsonb, 'system'),
  ('quest_reward_xp_multiplier', '100'::jsonb, 'system')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_by = 'system',
  updated_at = timezone('utc'::text, now());

-- ============================================================
-- Row-Level Security for seed data
-- ============================================================
ALTER TABLE odyssey_journey_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_course_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_quest_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_creative_prompt_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_achievement_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_concept_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_chest_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_drop_tables ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_relic_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_season_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_balance_configs ENABLE ROW LEVEL SECURITY;

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
    WHERE schemaname = 'public' AND tablename = 'odyssey_concept_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_concept_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_chest_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_chest_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_drop_tables' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_drop_tables FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_relic_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_relic_definitions FOR ALL TO service_role USING (true)';
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
    WHERE schemaname = 'public' AND tablename = 'odyssey_balance_configs' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_balance_configs FOR ALL TO service_role USING (true)';
  END IF;
END $$;

-- Update schema version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '10')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());