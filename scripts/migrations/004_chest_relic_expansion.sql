-- Chest & Relic System expansion (Milestone: Chest & Relic)
-- No breaking migrations: only ADD COLUMN and CREATE TABLE

-- Expand odyssey_chests with type metadata
ALTER TABLE odyssey_chests
  ADD COLUMN IF NOT EXISTS chest_slug TEXT,
  ADD COLUMN IF NOT EXISTS rarity TEXT,
  ADD COLUMN IF NOT EXISTS icon TEXT,
  ADD COLUMN IF NOT EXISTS description TEXT,
  ADD COLUMN IF NOT EXISTS drop_table TEXT;

-- Expand odyssey_relics with template metadata and owned_count
ALTER TABLE odyssey_relics
  ADD COLUMN IF NOT EXISTS slug TEXT,
  ADD COLUMN IF NOT EXISTS name TEXT,
  ADD COLUMN IF NOT EXISTS description TEXT,
  ADD COLUMN IF NOT EXISTS realm TEXT,
  ADD COLUMN IF NOT EXISTS rarity TEXT,
  ADD COLUMN IF NOT EXISTS image TEXT,
  ADD COLUMN IF NOT EXISTS lore TEXT,
  ADD COLUMN IF NOT EXISTS owned_count INTEGER DEFAULT 1;

-- odyssey_player_relics tracks per-player relic collection state
CREATE TABLE IF NOT EXISTS odyssey_player_relics (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  uid TEXT NOT NULL,
  relic_slug TEXT NOT NULL,
  relic_id BIGINT,
  owned_count INTEGER DEFAULT 1,
  discovered_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()),
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now())
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_player_relics_uid_slug
  ON odyssey_player_relics (uid, relic_slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_player_relics_uid
  ON odyssey_player_relics (uid);

-- Enable RLS
ALTER TABLE odyssey_chests ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_relics ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_player_relics ENABLE ROW LEVEL SECURITY;

-- Service role policy (mirrors existing convention)
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_chests' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_chests FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_relics' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_relics FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_player_relics' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_player_relics FOR ALL TO service_role USING (true)';
  END IF;
END $$;
