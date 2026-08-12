-- Migration 011: Story Fragments & Player Story Fragments
-- Defines collectible story fragments and player discovery tracking.

CREATE TABLE IF NOT EXISTS odyssey_story_fragments (
    slug TEXT PRIMARY KEY,
    journey TEXT NOT NULL,
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    set_name TEXT NOT NULL DEFAULT 'general',
    is_hidden BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS odyssey_player_story_fragments (
    id BIGSERIAL PRIMARY KEY,
    uid TEXT NOT NULL,
    family_id TEXT NOT NULL,
    fragment_slug TEXT NOT NULL REFERENCES odyssey_story_fragments(slug) ON DELETE CASCADE,
    discovered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_player_fragment UNIQUE (uid, fragment_slug)
);

CREATE INDEX IF NOT EXISTS idx_player_story_fragments_uid ON odyssey_player_story_fragments(uid);
CREATE INDEX IF NOT EXISTS idx_player_story_fragments_crew ON odyssey_player_story_fragments(family_id);

-- Seed minimal playable story fragments for Whispering Woods & Clockwork City
INSERT INTO odyssey_story_fragments (slug, journey, title, content, set_name, is_hidden)
VALUES
  ('ancient-bark-whisper', 'whispering-woods', 'Bisikan Pepohonan Tua', 'Pohon-pohon raksasa di Hutan Berbisik menyimpan gema langkah penjelajah pertama.', 'whispering-set', false),
  ('echo-of-the-first-explorer', 'whispering-woods', 'Gema Penjelajah Perdana', 'Rahasia Replay: Di balik lumut tua, terukir ukiran kompas kuno yang ditinggalkan ribuan purnama lalu.', 'whispering-set', true),
  ('copper-cog-diagram', 'clockwork-city', 'Bagan Roda Gigi Tembaga', 'Diagram kuno yang menunjukkan susunan roda gigi raksasa di pusat Kota Jam.', 'clockwork-set', false),
  ('secret-steam-valve', 'clockwork-city', 'Katup Uap Rahasia', 'Rahasia Replay: Katup kuningan yang menyembunyikan lorong ruang uap tak tersentuh.', 'clockwork-set', true)
ON CONFLICT (slug) DO NOTHING;
