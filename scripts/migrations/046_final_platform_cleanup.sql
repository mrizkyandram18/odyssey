-- ============================================================
-- Migration 046: Final Platform Cleanup
-- Drops obsolete RPCs and legacy RPG tables with dependency safety.
-- ============================================================

-- 1. Drop Obsolete RPCs
DROP FUNCTION IF EXISTS odyssey_complete_task(BIGINT, TEXT, JSONB);

-- 2. Drop all obsolete RPG tables with CASCADE safety
DROP TABLE IF EXISTS odyssey_task_completions CASCADE;
DROP TABLE IF EXISTS odyssey_reactions_legacy CASCADE;
DROP TABLE IF EXISTS odyssey_player_story_fragments CASCADE;
DROP TABLE IF EXISTS odyssey_story_fragments CASCADE;
DROP TABLE IF EXISTS odyssey_lore_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_creative_submissions CASCADE;
DROP TABLE IF EXISTS odyssey_creative_items CASCADE;
DROP TABLE IF EXISTS odyssey_creative_prompt_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_drop_tables CASCADE;
DROP TABLE IF EXISTS odyssey_gift_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_gifts CASCADE;
DROP TABLE IF EXISTS odyssey_player_collections CASCADE;
DROP TABLE IF EXISTS odyssey_collection_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_collections CASCADE;
DROP TABLE IF EXISTS odyssey_exercises CASCADE;
DROP TABLE IF EXISTS odyssey_mission_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_missions CASCADE;
DROP TABLE IF EXISTS odyssey_course_progress CASCADE;
DROP TABLE IF EXISTS odyssey_course_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_journey_progress CASCADE;
DROP TABLE IF EXISTS odyssey_journey_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_learning_concepts CASCADE;
DROP TABLE IF EXISTS odyssey_daily_activity_completions CASCADE;
DROP TABLE IF EXISTS odyssey_daily_activities CASCADE;
DROP TABLE IF EXISTS odyssey_daily_activity CASCADE;
DROP TABLE IF EXISTS odyssey_daily_missions CASCADE;
DROP TABLE IF EXISTS odyssey_achievement_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_achievements CASCADE;
DROP TABLE IF EXISTS odyssey_reactions CASCADE;
DROP TABLE IF EXISTS odyssey_reward_signals CASCADE;
DROP TABLE IF EXISTS odyssey_season_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_balance_configs CASCADE;
DROP TABLE IF EXISTS odyssey_cosmetic_unlocks CASCADE;
DROP TABLE IF EXISTS odyssey_reward_ledgers CASCADE;
DROP TABLE IF EXISTS odyssey_system_config CASCADE;
DROP TABLE IF EXISTS odyssey_audit_logs CASCADE;
DROP TABLE IF EXISTS odyssey_chapter_progress CASCADE;
DROP TABLE IF EXISTS odyssey_lore_unlocks CASCADE;
DROP TABLE IF EXISTS odyssey_relic_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_relics CASCADE;
DROP TABLE IF EXISTS odyssey_player_relics CASCADE;
DROP TABLE IF EXISTS odyssey_chest_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_chests CASCADE;
DROP TABLE IF EXISTS odyssey_realm_progress CASCADE;
DROP TABLE IF EXISTS odyssey_realm_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_chapter_definitions CASCADE;
DROP TABLE IF EXISTS odyssey_quests CASCADE;
DROP TABLE IF EXISTS odyssey_challenges CASCADE;
DROP TABLE IF EXISTS odyssey_daily_turns CASCADE;

-- 3. Register Migration Version in odyssey_schema_version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '046_final_platform_cleanup')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
