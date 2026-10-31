-- ============================================================
-- Migration 105: Harden Reward RPC Privileges
--
-- Security Hardening:
-- Defends against direct client execution of reward-issuing RPCs.
-- 1. Function odyssey_submit_auto_task(BIGINT, TEXT, JSONB):
--    - REVOKE EXECUTE from PUBLIC, anon, authenticated.
--    - GRANT EXECUTE to service_role.
-- 2. Function odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT):
--    - REVOKE EXECUTE from PUBLIC, anon, authenticated.
--    - GRANT EXECUTE to service_role.
-- 3. Version tracking:
--    - Update odyssey_schema_version to '105_harden_reward_rpc_privileges'.
--
-- ROLLBACK INSTRUCTIONS:
-- To rollback Migration 105 and restore initial PostgreSQL catalog proacl:
--
-- REVOKE ALL ON FUNCTION public.odyssey_submit_auto_task(BIGINT, TEXT, JSONB) FROM PUBLIC, anon, authenticated, service_role;
-- GRANT EXECUTE ON FUNCTION public.odyssey_submit_auto_task(BIGINT, TEXT, JSONB) TO anon, authenticated, service_role;
-- REVOKE ALL ON FUNCTION public.odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) FROM PUBLIC, anon, authenticated, service_role;
-- GRANT EXECUTE ON FUNCTION public.odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) TO anon, authenticated, service_role;
--
-- UPDATE public.odyssey_schema_version
-- SET value = '104_atomic_admin_income_configuration',
--     updated_at = timezone('utc'::text, now())
-- WHERE key = 'schema_version' AND value = '105_harden_reward_rpc_privileges';
-- ============================================================

-- 1. Restrict odyssey_submit_auto_task
REVOKE ALL ON FUNCTION public.odyssey_submit_auto_task(BIGINT, TEXT, JSONB) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION public.odyssey_submit_auto_task(BIGINT, TEXT, JSONB) TO service_role;

-- 2. Restrict odyssey_verify_submission
REVOKE ALL ON FUNCTION public.odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) FROM PUBLIC, anon, authenticated;
GRANT EXECUTE ON FUNCTION public.odyssey_verify_submission(BIGINT, TEXT, TEXT, TEXT, INT) TO service_role;

-- 3. Update schema version
INSERT INTO public.odyssey_schema_version (key, value)
VALUES ('schema_version', '105_harden_reward_rpc_privileges')
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
