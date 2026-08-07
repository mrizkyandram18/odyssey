-- Migration 016: Seed Reactions
-- Seed reaction data using the correct schema mapped to demo-uid target users

INSERT INTO odyssey_reactions (id, creator_id, target_user_id, quest_id, emoji_code, created_at)
VALUES
  ('00000000-0000-0000-0000-000000000001', 'demo-uid-2', 'demo-uid-1', NULL, '✨', timezone('utc'::text, now() - interval '2 days')),
  ('00000000-0000-0000-0000-000000000002', 'demo-uid-3', 'demo-uid-1', NULL, '❤️', timezone('utc'::text, now() - interval '2 days')),
  ('00000000-0000-0000-0000-000000000003', 'demo-uid-1', 'demo-uid-2', NULL, '👏', timezone('utc'::text, now() - interval '2 days')),
  ('00000000-0000-0000-0000-000000000004', 'demo-uid-2', 'demo-uid-3', NULL, '👍', timezone('utc'::text, now() - interval '1 day')),
  ('00000000-0000-0000-0000-000000000005', 'demo-uid-3', 'demo-uid-1', NULL, '🔥', timezone('utc'::text, now() - interval '2 hours'))
ON CONFLICT (id) DO NOTHING;
