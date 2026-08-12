-- Migration 030: Add learn_text and result_text to missions for Phase 1

ALTER TABLE odyssey_quest_definitions
ADD COLUMN IF NOT EXISTS learn_text TEXT NULL,
ADD COLUMN IF NOT EXISTS result_text TEXT NULL;
