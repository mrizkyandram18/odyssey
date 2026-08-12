-- Migration 036: Refactor Domain to Family Learning
-- This migration renames tables and columns from the old RPG terminology to the new Family Learning terminology.

-- 1. Rename Crews to Families
ALTER TABLE odyssey_crews RENAME TO odyssey_families;
ALTER TABLE odyssey_users RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_realm_progress RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_chapter_progress RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_lore_unlocks RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_achievements RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_reactions RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_creative_items RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_creative_submissions RENAME COLUMN crew_id TO family_id;

-- 2. Rename Quests to Missions
ALTER TABLE odyssey_quest_definitions RENAME TO odyssey_mission_definitions;
ALTER TABLE odyssey_mission_definitions RENAME COLUMN quest_type TO mission_type;
ALTER TABLE odyssey_mission_definitions RENAME COLUMN challenge_defs TO exercise_defs;
ALTER TABLE odyssey_quests RENAME TO odyssey_missions;
ALTER TABLE odyssey_missions RENAME COLUMN crew_id TO family_id;
ALTER TABLE odyssey_missions RENAME COLUMN quest_slug TO mission_slug;

-- 3. Rename Challenges to Exercises
ALTER TABLE odyssey_challenges RENAME TO odyssey_exercises;
ALTER TABLE odyssey_exercises RENAME COLUMN quest_id TO mission_id;
ALTER TABLE odyssey_exercises RENAME COLUMN challenge_slug TO exercise_slug;

-- 4. Rename Daily Turns to Daily Missions
ALTER TABLE odyssey_daily_turns RENAME TO odyssey_daily_missions;
ALTER TABLE odyssey_daily_missions RENAME COLUMN quest_slug TO mission_slug;
ALTER TABLE odyssey_daily_missions RENAME COLUMN daily_turn_id TO daily_mission_id;

-- 5. Rename Realms & Chapters to Journeys & Courses
ALTER TABLE odyssey_realm_definitions RENAME TO odyssey_journey_definitions;
ALTER TABLE odyssey_chapter_definitions RENAME TO odyssey_course_definitions;
ALTER TABLE odyssey_realm_progress RENAME TO odyssey_journey_progress;
ALTER TABLE odyssey_journey_progress RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_course_progress RENAME TO odyssey_course_progress;
ALTER TABLE odyssey_course_progress RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_course_progress RENAME COLUMN chapter TO course;
ALTER TABLE odyssey_creative_items RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_relic_definitions RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_relics RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_lore_unlocks RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_lore_unlocks RENAME COLUMN chapter TO course;
ALTER TABLE odyssey_journey_definitions RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_course_definitions RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_course_definitions RENAME COLUMN chapter TO course;
ALTER TABLE odyssey_mission_definitions RENAME COLUMN realm TO journey;
ALTER TABLE odyssey_mission_definitions RENAME COLUMN chapter TO course;

-- 6. Rename Lore to Learning Concepts
ALTER TABLE odyssey_lore_unlocks RENAME TO odyssey_learning_concepts;
ALTER TABLE odyssey_learning_concepts RENAME COLUMN lore_slug TO concept_slug;

-- 7. Rename Relics to Collections
ALTER TABLE odyssey_relic_definitions RENAME TO odyssey_collection_definitions;
ALTER TABLE odyssey_relics RENAME TO odyssey_collections;
ALTER TABLE odyssey_player_relics RENAME TO odyssey_player_collections;
ALTER TABLE odyssey_player_collections RENAME COLUMN relic_slug TO collection_slug;
ALTER TABLE odyssey_player_collections RENAME COLUMN relic_id TO collection_id;
ALTER TABLE odyssey_drop_tables RENAME COLUMN relic_id TO collection_id;
ALTER TABLE odyssey_chests RENAME COLUMN reward_relic TO reward_collection;

-- 8. Rename Chests to Gifts
ALTER TABLE odyssey_chest_definitions RENAME TO odyssey_gift_definitions;
ALTER TABLE odyssey_chests RENAME TO odyssey_gifts;
ALTER TABLE odyssey_gifts RENAME COLUMN chest_slug TO gift_slug;
ALTER TABLE odyssey_drop_tables RENAME COLUMN chest_slug TO gift_slug;

-- 9. Update Creative Submissions
ALTER TABLE odyssey_creative_submissions RENAME COLUMN quest_id TO mission_id;
ALTER TABLE odyssey_creative_submissions RENAME COLUMN challenge_id TO exercise_id;
