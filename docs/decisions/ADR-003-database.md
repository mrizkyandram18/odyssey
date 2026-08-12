# ADR-003: Database

- **Date:** 2026-08-02
- **Status:** Accepted
- **Deciders:** Lead Architect / Product Architect

## Context

Odyssey and Family Reward share the **same Supabase (PostgreSQL) project**.
Family Reward uses tables such as `user_profiles`, `user_kuota_request`,
`work_logs`, `system_configs`, `gts26_events`, `gts26_rewards`, and others.

Odyssey needs persistent storage for game state (missions, progression,
creative contributions, world state) but must:

1. Never modify existing business tables.
2. Avoid name collisions with existing tables.
3. Follow conventions that make the schema predictable and maintainable.

## Decision

Odyssey creates **only new tables prefixed with `odyssey_`**. No existing table
is altered. The naming convention, column conventions, and RLS policy pattern
mirror Family Reward's established practices for consistency, but the
implementation is fully independent.

### Naming Convention

| Element | Convention | Example |
|---|---|---|
| Tables | `odyssey_` + `snake_case` | `odyssey_user_profiles` |
| Columns | `snake_case` | `family_id`, `completed_at` |
| Primary key | `id` (BIGINT, identity) or `uid` (TEXT, natural) | `id`, `uid` |
| Timestamps | `created_at`, `updated_at` | `TIMESTAMP WITH TIME ZONE DEFAULT timezone('utc'::text, now())` |
| Indexes | `idx_<table>_<column>` | `idx_odyssey_missions_family_id` |
| Unique indexes | `uniq_<table>_<column>` | `uniq_odyssey_missions_family_id_slug` |

### Tables (MVP)

| Table | Columns | Notes |
|---|---|---|
| `odyssey_user_profiles` | `uid TEXT PK`, `family_id TEXT`, `explorer_name TEXT`, `role TEXT`, `level INTEGER`, `xp BIGINT`, `created_at`, `updated_at` | UID is shared with Gatekeeper. One user → one crew. Stores hashed password for credential verification. |
| `odyssey_families` | `id TEXT PK` (UUID), `name TEXT`, `created_at`, `updated_at` | A family group. Seeded manually per family. |
| `odyssey_missions` | `id BIGINT PK`, `family_id TEXT`, `template_slug TEXT`, `title TEXT`, `status TEXT` (`ACTIVE`/`DONE`), `started_at`, `completed_at`, `created_at` | One row per quest instance per crew. |
| `odyssey_exercises` | `id BIGINT PK`, `mission_id BIGINT FK`, `slug TEXT`, `description TEXT`, `status TEXT` (`PENDING`/`DONE`), `completed_by TEXT` (uid), `completed_at`, `created_at` | |
| `odyssey_journey_progress` | `family_id TEXT PK`, `journey TEXT`, `status TEXT` (`LOCKED`/`ACTIVE`/`COMPLETE`), `story_branch TEXT`, `progress INTEGER` (0–100), `last_unlocked_at`, `updated_at` | The family's shared journey progress. Renamed from `odyssey_world_state` for clarity. |
| `odyssey_creative_items` | `id BIGINT PK`, `family_id TEXT`, `journey TEXT`, `author_uid TEXT`, `kind TEXT` (`STORY`/`COMIC`/`PHOTO`/`VIDEO`), `payload JSONB`, `created_at` | Append-only: contributions to creative spaces. |
| `odyssey_daily_missions` | `id BIGINT PK`, `uid TEXT`, `date DATE`, `mission_slug TEXT`, `completed BOOLEAN`, `created_at` | One per user per calendar day. |
| `odyssey_achievements` | `id BIGINT PK`, `uid TEXT`, `code TEXT`, `kind TEXT` (`PERSONAL`/`GROUP`), `awarded_at`, `created_at` | Personal and group milestones. |
| `odyssey_collections` | `id BIGINT PK`, `uid TEXT`, `code TEXT`, `awarded_at`, `created_at` | Collected Collections, personal. |
| `odyssey_gifts` | `id BIGINT PK`, `uid TEXT`, `source TEXT`, `opened BOOLEAN`, `opened_at`, `created_at` | Reward containers with known, fixed contents. |

### Tables Deferred to Phase 5

The `odyssey_reward_signals` table is **not** created in the MVP. It will be
introduced only when the Family Reward integration is activated. See
[ADR-004](ADR-004-reward-integration.md).

### RLS & Access Policy

- **All** `odyssey_*` tables enable Row-Level Security.
- A single policy `Allow service_role full access` is applied (matching
  Family Reward's convention), since access control is enforced at the
  application layer via session UIDs.
- The `odyssey_user_profiles` table's `uid` is the natural primary key
  (matching Family Reward's `user_profiles` which is keyed by `uid`).

### Feature Flags

- Feature flags are stored in the **shared** `system_configs` table
  (key/value: `key TEXT PK`, `value TEXT`). This table already exists in the
  shared Supabase project.
- Odyssey reads only keys prefixed with `odyssey_` (e.g.,
  `odyssey_maintenance_mode`, `odyssey_min_build_number`).
- This avoids table duplication while preventing key collisions with
  Family Reward's config keys.

### Concurrency & Consistency

- **Optimistic concurrency:** Read-then-patch with a filter on the previous
  value, the same pattern Family Reward uses in `UpdateStatus`. This prevents
  lost updates on concurrent writes (e.g., two family members completing a
  relay quest leg simultaneously).
- **Append-only logs** for creative contributions and milestone events —
  never overwrite, only insert.
- **Atomic guards:** Unique indexes enforce invariants (e.g., one active
  quest per template per crew, one daily turn per user per day).

### Access Pattern

- Backend uses **Supabase REST API (PostgREST)** via direct HTTP calls with
  the service-role key — the same pattern Family Reward uses in
  `supabase_http.go`. No Supabase client SDK in the MVP.
- Responses are JSON arrays for list endpoints and single objects for
  point lookups (PostgREST convention).
- `return=representation` is used on POST/PATCH to read back computed fields.

### Migrations

- Migrations are managed via SQL scripts in `scripts/migrations/`
  (`001_initial_schema.sql`, `002_indexes.sql`, etc.).
- Each migration is idempotent (`CREATE TABLE IF NOT EXISTS`,
  `ADD COLUMN IF NOT EXISTS`).
- Migrations are applied manually or via a future deploy hook. No migration
  framework is introduced in the MVP scope.

## Alternatives Considered

### D1: Separate Supabase project for Odyssey

**Rejected.** The project brief explicitly states: "Same Supabase project."
A separate project would add operational overhead and break the shared-identity
model (UID consistency across Gatekeeper and Family Reward).

### D2: Use the same tables as Family Reward

**Rejected.** Family Reward's tables are domain-specific (quota requests,
gts26 rewards). Sharing them would tightly couple the two applications and
risk corrupting Family Reward's data. The `odyssey_` prefix provides a hard
boundary.

### D3: Use Supabase real-time subscriptions (WebSocket)

**Rejected for MVP.** The core loop is turn-based with eventual consistency.
Polling on foreground and immediate-on-submit is sufficient.
Real-time subscriptions are considered for Phase 2 (handoff notifications).

### D4: Use an ORM (GORM, etc.)

**Rejected.** An ORM adds complexity and opinion for a handful of tables.
Direct PostgREST calls (as Family Reward does) are transparent, debuggable,
and require no dependency. If the schema grows significantly, an ORM can be
introduced later behind the `pkg/db` adapter.

## Consequences

### Positive

- Hard namespace boundary prevents collisions and accidental coupling.
- Mirroring Family Reward's conventions (snake_case, timestamps, RLS policy
  names, PostgREST access) reduces cognitive load for developers familiar with
  the existing system.
- A clear migration path exists if the schema outgrows the manual approach.

### Negative

- Some duplication of helper code (Supabase REST client). Accepted cost of
  system independence (see [ADR-001](ADR-001-project-scope.md)).
- The shared `system_configs` table creates a soft coupling via key naming
  convention only. If collisions arise, a dedicated `odyssey_configs` table
  can be introduced.

## References

- [Integrations](../integrations.md)
- [ADR-001: Project Scope](ADR-001-project-scope.md)
- [ADR-004: Reward Integration](ADR-004-reward-integration.md)
- Family Reward reference: `supabase_schema.sql`, `pkg/shared/supabase.go`,
  `pkg/shared/supabase_http.go`
