-- Migration 024: Slice 2.3 — Shared crew text board reaction target
-- Scope: append-only multi-entry text board (NOT real-time collaborative editing).
-- Posts live in existing odyssey_creative_items (kind = SHARED_TEXT).
-- Reactions may target TEXT_BOARD (creative_items.id).

-- Expand reaction target_type check to allow TEXT_BOARD.
ALTER TABLE odyssey_reactions DROP CONSTRAINT IF EXISTS odyssey_reactions_target_type_check;
ALTER TABLE odyssey_reactions
  ADD CONSTRAINT odyssey_reactions_target_type_check
  CHECK (target_type IN ('JOURNAL', 'QUEST', 'TEXT_BOARD'));

-- Helpful index for listing board posts by crew + kind
CREATE INDEX IF NOT EXISTS idx_odyssey_creative_items_crew_kind
  ON odyssey_creative_items (crew_id, kind, created_at DESC);

INSERT INTO odyssey_schema_version (key, value)
VALUES ('schema_version', '024_shared_text_board')
ON CONFLICT (key) DO UPDATE SET
  value = EXCLUDED.value,
  updated_at = timezone('utc'::text, now());
