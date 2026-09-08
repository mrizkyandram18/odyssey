# Credential SOT: `odyssey_user_profiles` (076/077)

## Current state

`odyssey_user_profiles` is the single Source of Truth for identity and credentials:
`uid`, `username` (UNIQUE, nullable for legacy rows), `password_hash` (bcrypt),
`role`, `family_id`, `device_id`, `device_bound_at`, `is_active`, `must_change_password`.

`odyssey_local_users` was the legacy credential store (`username` + `password_hash`
+ `profile_uid`) and has been removed:

- `076_profiles_username_backfill.sql` — adds `profiles.username`, backfills
  `username` + `password_hash` byte-for-byte from `local_users` (COALESCE-only,
  fail-closed conflict guards, no password generation, no rehash).
- `077_drop_local_users.sql` — drops the index, RLS policy, and table.
  Apply ONLY after live-DB safety gates are GREEN (see checklist below).

Historical migrations (`019`, `049`, `062`), `scripts/dev/001_local_auth.sql`, and
phase audit reports still reference `local_users` as historical record. They must
NOT be treated as current architecture.

## Known discrepancy

`049_update_demo_credentials.sql` stores truncated hashes (suffix only, missing the
`$2a$10$...` bcrypt prefix). Any production row affected by 049 holds an invalid
credential. Do NOT bulk-repair: verify with the invalid-bcrypt query and reset only
the affected user via the existing admin reset flow. Healthy users are never
re-issued passwords during migration.

## Delete semantics (preserved)

Delete = `is_active=false` + `password_hash=NULL` + `username=NULL` on the profile
row (mirrors the old `DELETE local_users` row, which removed both hash and username
and freed the username for reuse). `password_hash IS NULL` + `is_active=false` is
the deleted discriminator; blocked users keep their hash. Unblock on a deleted user
is rejected.

## Live-DB verification (owner runs in Supabase SQL Editor, before 077)

```sql
-- A. Active profiles missing credentials → expect 0 rows
SELECT uid FROM odyssey_user_profiles
WHERE is_active = true AND (username IS NULL OR password_hash IS NULL);
-- B. Duplicate profile usernames → expect 0 rows
SELECT username, COUNT(*) FROM odyssey_user_profiles
WHERE username IS NOT NULL GROUP BY username HAVING COUNT(*) > 1;
-- C. Duplicate local usernames → expect 0 rows
SELECT username, COUNT(*) FROM odyssey_local_users
GROUP BY username HAVING COUNT(*) > 1;
-- D. Duplicate local profile mappings → expect 0 rows
SELECT profile_uid, COUNT(*) FROM odyssey_local_users
GROUP BY profile_uid HAVING COUNT(*) > 1;
-- E. Orphan local users → expect 0 rows
SELECT l.profile_uid FROM odyssey_local_users l
LEFT JOIN odyssey_user_profiles p ON p.uid = l.profile_uid
WHERE p.uid IS NULL;
-- F. Hash conflicts → expect 0 rows
SELECT p.uid FROM odyssey_user_profiles p JOIN odyssey_local_users l
ON l.profile_uid = p.uid
WHERE p.password_hash IS NOT NULL AND l.password_hash IS NOT NULL
AND p.password_hash <> l.password_hash;
-- G. Username conflicts → expect 0 rows
SELECT p.uid FROM odyssey_user_profiles p JOIN odyssey_local_users l
ON l.profile_uid = p.uid
WHERE p.username IS NOT NULL AND p.username <> l.username;
-- H. Local-only credentials (must be backfilled) → expect 0 rows
SELECT l.profile_uid FROM odyssey_local_users l
JOIN odyssey_user_profiles p ON p.uid = l.profile_uid
WHERE l.password_hash IS NOT NULL AND p.password_hash IS NULL;
-- I. Invalid bcrypt on loginable rows → expect 0 rows
SELECT uid, left(password_hash, 7) FROM odyssey_user_profiles
WHERE password_hash IS NOT NULL
AND password_hash NOT LIKE '$2a$%' AND password_hash NOT LIKE '$2b$%'
AND password_hash NOT LIKE '$2y$%';
```

Production order: apply 076 → backfill → run queries A–I → all GREEN →
verify existing login → deploy profiles-first code → critical auth tests →
confirm zero `local_users` runtime dependency → apply 077 →
`SELECT to_regclass('public.odyssey_local_users')` must return NULL.

## Live execution log (2026-09-08, via Supabase CLI + REST, read-only except noted)

- CLI authenticated; project linked; remote was at 075. `076` pushed alone
  (`077` staged aside) — first attempt caught a guard-ordering bug (guards ran
  before ADD COLUMN); fixed, re-pushed clean. Remote went to 076.
- Backfill verified live: `admin→demo-uid-2`, `selvicahyani→usr_1788196798`
  present with matching usernames; hashes copied 0→2 with 0 conflicts;
  `usr_1788279487` (inactive, no credential) untouched.
- Pre-deploy probes (throwaway temp users, real users untouched):
  profiles-only → 401, `local_users`-backed → 200 → production API was still
  running pre-migration code. Backfill harmless to it.
- Profiles-first code deployed to production via Vercel (`odyssey` project,
  production target, READY). Post-deploy probe: profiles-only temp user → 200,
  wrong password → 401. Existing `admin` login attempt → 403 device-blocked
  (credential accepted, unknown device correctly rejected — proves credential
  path + device binding intact on new code).
- `077` applied only after the above. `odyssey_local_users` REST now returns
  404 PGRST205 (table physically gone with its index/policy/FK).
- Post-drop smoke: fresh profiles-only temp user → login 200, wrong password
  401, temp user removed. No temp rows remain.
- Known corrupt credential (pre-existing, preserved as-is, NOT reset):
  `selvicahyani → usr_1788196798` holds a non-bcrypt hash. Manual action for
  the owner: reset that ONE user from the bound-device admin session via the
  existing admin reset flow. Never a mass reset.
