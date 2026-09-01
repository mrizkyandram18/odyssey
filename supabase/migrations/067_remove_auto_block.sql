-- ============================================================
-- Migration 067: Remove auto-block execution (keep manual block)
-- Drops automatic blocking RPCs; preserves manual block fields
-- and blocked guards on submit/create_claim.
-- 065/066 already applied in production — this is safe additive DROP.
-- ============================================================

-- Drop auto-block RPCs (idempotent)
DROP FUNCTION IF EXISTS odyssey_auto_block_inactive_users();
DROP FUNCTION IF EXISTS odyssey_auto_block_inactive_users_with_threshold(INT, TEXT);
DROP FUNCTION IF EXISTS odyssey_auto_block_inactive_users_with_threshold(INT);

-- Keep manual block RPCs: odyssey_block_user, odyssey_unblock_user
-- Keep blocked guards on odyssey_submit_auto_task, odyssey_submit_manual_task, odyssey_create_claim
-- Keep columns blocked_at, blocked_by, block_reason (manual blocking)

-- Optionally keep config key auto_block_inactivity_days for display threshold
-- No deletion of historical data

INSERT INTO odyssey_schema_version(key,value) VALUES('schema_version','067_remove_auto_block')
ON CONFLICT(key) DO UPDATE SET value=EXCLUDED.value, updated_at=timezone('utc'::text, now());
