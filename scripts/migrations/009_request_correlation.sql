-- Migration 009: Request correlation & observability
-- Adds: request_id column to audit logs, indexes for correlation.
-- Schema version tracking table was created in migration 003.

-- ============================================================
-- Add request_id to audit logs for request correlation
-- ============================================================
ALTER TABLE odyssey_audit_logs ADD COLUMN IF NOT EXISTS request_id TEXT;

CREATE INDEX IF NOT EXISTS idx_odyssey_audit_logs_request_id
    ON odyssey_audit_logs (request_id)
    WHERE request_id IS NOT NULL;

-- ============================================================
-- Schema version tracking
-- ============================================================
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '9')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());

-- RLS + service role policy
ALTER TABLE odyssey_schema_version ENABLE ROW LEVEL SECURITY;
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_schema_version'
    AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_schema_version FOR ALL TO service_role USING (true)';
  END IF;
END $$;
