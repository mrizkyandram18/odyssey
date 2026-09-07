-- ============================================================
-- Migration 075: Backfill camera_only = true on existing Bukti Diskusi tasks
-- Ensures existing tasks titled 'Bukti Diskusi' have camera_only set to true
-- Idempotent, scoped strictly to title = 'Bukti Diskusi'.
-- Generic PHOTO_UPLOAD tasks without camera_only remain untouched.
-- ============================================================

UPDATE odyssey_tasks
SET config = jsonb_set(
    jsonb_set(
        COALESCE(config, '{}'::jsonb),
        '{camera_only}',
        'true'::jsonb,
        true
    ),
    '{camera_facing}',
    to_jsonb(COALESCE(config->>'camera_facing', 'environment')),
    true
)
WHERE title = 'Bukti Diskusi'
  AND (config IS NULL OR config->>'camera_only' IS NULL OR config->>'camera_only' = 'false');

INSERT INTO odyssey_schema_version(key, value)
VALUES('schema_version', '075_bukti_diskusi_camera_only')
ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value, updated_at = timezone('utc'::text, now());
