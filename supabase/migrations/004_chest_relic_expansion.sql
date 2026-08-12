-- Gift & Collection System expansion (Milestone: Gift & Collection)
-- No breaking migrations: only ADD COLUMN and CREATE TABLE

-- Expand odyssey_gifts with type metadata
ALTER TABLE odyssey_gifts
  ADD COLUMN IF NOT EXISTS gift_slug TEXT,
  ADD COLUMN IF NOT EXISTS rarity TEXT,
  ADD COLUMN IF NOT EXISTS icon TEXT,
  ADD COLUMN IF NOT EXISTS description TEXT,
  ADD COLUMN IF NOT EXISTS drop_table TEXT;

-- Expand odyssey_collections with template metadata and owned_count
ALTER TABLE odyssey_collections
  ADD COLUMN IF NOT EXISTS slug TEXT,
  ADD COLUMN IF NOT EXISTS name TEXT,
  ADD COLUMN IF NOT EXISTS description TEXT,
  ADD COLUMN IF NOT EXISTS journey TEXT,
  ADD COLUMN IF NOT EXISTS rarity TEXT,
  ADD COLUMN IF NOT EXISTS image TEXT,
  ADD COLUMN IF NOT EXISTS concept TEXT,
  ADD COLUMN IF NOT EXISTS owned_count INTEGER DEFAULT 1;

-- odyssey_player_collections tracks per-player relic collection state
CREATE TABLE IF NOT EXISTS odyssey_player_collections (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  uid TEXT NOT NULL,
  collection_slug TEXT NOT NULL,
  collection_id BIGINT,
  owned_count INTEGER DEFAULT 1,
  discovered_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()),
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now())
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_player_collections_uid_slug
  ON odyssey_player_collections (uid, collection_slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_player_collections_uid
  ON odyssey_player_collections (uid);

-- Enable RLS
ALTER TABLE odyssey_gifts ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_collections ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_player_collections ENABLE ROW LEVEL SECURITY;

-- Service role policy (mirrors existing convention)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_gifts' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_gifts FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_collections' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_collections FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_player_collections' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_player_collections FOR ALL TO service_role USING (true)';
  END IF;
END $$;
