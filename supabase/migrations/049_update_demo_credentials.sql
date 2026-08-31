-- Migration 049: Update Demo Credentials
-- Admin: admin / admin123
-- Normal User: user_testing / odyssey123

DELETE FROM odyssey_local_users WHERE username IN ('demo1', 'demo2', 'demo3');

INSERT INTO odyssey_local_users (id, username, password_hash, profile_uid)
VALUES 
  ('local-demo-1', 'user_testing', '.NjpCtAkylXUFKlki/dfcxA41Fgy.9Ay', 'demo-uid-1'),
  ('local-demo-2', 'admin', '.wGgERYfygx.Y2vDhaNmnlRANAgDRe6', 'demo-uid-2')
ON CONFLICT (id) DO UPDATE SET
  username = EXCLUDED.username,
  password_hash = EXCLUDED.password_hash,
  profile_uid = EXCLUDED.profile_uid,
  updated_at = timezone('utc'::text, now());
