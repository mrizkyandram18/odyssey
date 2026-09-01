-- Fix overload ambiguity: keep only 4-arg version with default
DROP FUNCTION IF EXISTS odyssey_calc_target_reward(INT,BIGINT,TEXT);
-- Ensure 4-arg version exists (already created in 061, re-create to be safe)
-- No-op, already exists as odyssey_calc_target_reward(INT,BIGINT,TEXT,TEXT) with default NULL
SELECT 1;
