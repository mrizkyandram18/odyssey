-- Chest & Relic definitions expansion (configurable catalogs)
-- No breaking migrations: only CREATE TABLE

-- odyssey_chest_definitions stores chest templates (admin-managed)
CREATE TABLE IF NOT EXISTS odyssey_chest_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  rarity TEXT NOT NULL,
  icon TEXT NOT NULL,
  description TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()),
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_odyssey_chest_definitions_slug
  ON odyssey_chest_definitions (slug);

-- odyssey_drop_tables stores rarity weights per chest definition
CREATE TABLE IF NOT EXISTS odyssey_drop_tables (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  chest_slug TEXT NOT NULL REFERENCES odyssey_chest_definitions(slug) ON DELETE CASCADE,
  rarity TEXT NOT NULL,
  relic_id BIGINT,
  weight DOUBLE PRECISION NOT NULL,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now())
);

CREATE UNIQUE INDEX IF NOT EXISTS uniq_odyssey_drop_tables_chest_rarity
  ON odyssey_drop_tables (chest_slug, rarity);

-- odyssey_relic_definitions stores relic templates (admin-managed)
CREATE TABLE IF NOT EXISTS odyssey_relic_definitions (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  description TEXT NOT NULL,
  realm TEXT NOT NULL,
  rarity TEXT NOT NULL,
  image TEXT NOT NULL,
  lore TEXT NOT NULL,
  created_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now()),
  updated_at TIMESTAMPTZ DEFAULT timezone('utc'::text, now())
);

CREATE INDEX IF NOT EXISTS idx_odyssey_relic_definitions_slug
  ON odyssey_relic_definitions (slug);

CREATE INDEX IF NOT EXISTS idx_odyssey_relic_definitions_realm
  ON odyssey_relic_definitions (realm);

-- Enable RLS
ALTER TABLE odyssey_chest_definitions ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_drop_tables ENABLE ROW LEVEL SECURITY;
ALTER TABLE odyssey_relic_definitions ENABLE ROW LEVEL SECURITY;

-- Service role policies
DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_chest_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_chest_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_drop_tables' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_drop_tables FOR ALL TO service_role USING (true)';
  END IF;
END $$;

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_policies
    WHERE schemaname = 'public' AND tablename = 'odyssey_relic_definitions' AND policyname = 'Allow service_role full access'
  ) THEN
    EXECUTE 'CREATE POLICY "Allow service_role full access" ON odyssey_relic_definitions FOR ALL TO service_role USING (true)';
  END IF;
END $$;
