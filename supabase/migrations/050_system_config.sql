-- ============================================================
-- Migration 050: System Configuration Table
-- Supports configurable redemption window (redemption_start_day, redemption_end_day)
-- ============================================================

CREATE TABLE IF NOT EXISTS odyssey_system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()) NOT NULL
);

-- Row-Level Security
ALTER TABLE odyssey_system_config ENABLE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS "Allow service_role full access on system_config" ON odyssey_system_config;
CREATE POLICY "Allow service_role full access on system_config" ON odyssey_system_config FOR ALL USING (true);

-- Revoke direct mutation from anon & authenticated
REVOKE ALL ON odyssey_system_config FROM anon, authenticated;

-- Seed default redemption period configuration (21 to 26)
INSERT INTO odyssey_system_config (key, value)
VALUES 
    ('redemption_start_day', '21'),
    ('redemption_end_day', '26')
ON CONFLICT (key) DO NOTHING;

-- Register migration version in odyssey_schema_version
INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '050_system_config')
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    updated_at = timezone('utc'::text, now());
