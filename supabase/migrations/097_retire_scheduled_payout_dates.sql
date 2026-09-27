-- ============================================================
-- Migration 097: Retire scheduled monthly payout dates in favor of flexible on-demand withdrawals (min 500 coins)
--
-- CONTEXT: Scheduled monthly payout dates (e.g. payout_day 24, window 21-26)
-- are legacy and confusing to users and admins. Withdrawals are now flexible
-- and on-demand whenever the user has reached at least 500 coins (in multiples of 500).
-- ============================================================

-- 1. Ensure system default payout frequency is THRESHOLD (flexible on-demand)
INSERT INTO odyssey_system_config(key, value)
VALUES ('default_payout_frequency', 'THRESHOLD')
ON CONFLICT (key) DO UPDATE SET value = 'THRESHOLD';

-- 2. Ensure minimum withdrawal is 500 coins
INSERT INTO odyssey_system_config(key, value)
VALUES ('default_minimum_withdrawal_coins', '500')
ON CONFLICT (key) DO UPDATE SET value = '500';

-- 3. Update any user payout configs that were still on MONTHLY to THRESHOLD
UPDATE odyssey_user_payout_config
SET payout_frequency = 'THRESHOLD'
WHERE payout_frequency = 'MONTHLY';

-- 4. Bump schema version
INSERT INTO odyssey_schema_version(key, value)
VALUES ('schema_version', '097_retire_scheduled_payout_dates')
ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
