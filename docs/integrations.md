# Integrations

Odyssey is an independent system. It integrates with existing production
systems through narrow, stable interfaces — never by sharing tables, code, or
internal implementation details.

## 1. Gatekeeper (Authentication Provider)

### What It Is

Gatekeeper is an Android device-trust application. It is the **sole identity
provider** for Odyssey. **Gatekeeper must never be modified.**

### How Odyssey Uses It (Through an Adapter)

Odyssey reuses the **BOTH login mode** — the only supported authentication
path. Crucially, the Odyssey **domain layer never references Firestore
directly**. Instead:

1. The domain defines an **`Authenticator` port** (an interface) with a single
   method: verify device trust + credential for a given UID.
2. An **adapter** (`pkg/auth/firestore.go`) implements this port by reading
   the Gatekeeper device document from Firestore at:
   `users/{PARENT_ID}/children/{uid}`. It verifies:
   - The document exists and `isOnline` is `true`.
   - `details.appBuildNumber` is >= the minimum (configurable, defaults to "49").
   - `details.permissions` includes all required permissions.

   See [ADR-002](decisions/ADR-002-authentication.md).

3. **Device Binding.** On first successful compliance check, the adapter
   records the device ID in Odyssey's `odyssey_user_profiles` table.
   Subsequent logins verify the device ID matches the bound one.

4. **Credential.** After Gatekeeper compliance passes, the user supplies a
   password (or PIN) verified against `odyssey_user_profiles`. This is
   **credential storage and verification**, not a new authentication system —
   the trust anchor remains Gatekeeper's device compliance.

### What Odyssey Does NOT Do

- Does not modify Gatekeeper or its Firestore documents.
- Does not call Gatekeeper's internal APIs (if any).
- Does not replicate Gatekeeper's business logic.
- Does not share its session tokens with Family Reward or vice versa.
- The domain layer does not import `cloud.google.com/go/firestore`. Only
  the adapter does.

### Environment Dependencies

| Variable | Purpose | Source |
|---|---|---|
| `FIREBASE_CREDENTIALS` / `GOOGLE_APPLICATION_CREDENTIALS_JSON` | Service account to read Firestore | Same Firebase project as Gatekeeper |
| `PARENT_ID` | The Gatekeeper parent user whose children are family members | Configured per family group |
| `GATEKEEPER_MIN_BUILD_NUMBER` | Minimum Gatekeeper app build (default "49") | System config |

### Identity

The **UID** is the shared identity. The same UID that a family member uses to
sign into Gatekeeper and Family Reward is used in Odyssey. This enables
cross-system achievement signaling without tight coupling.

---

## 2. Family Reward (Reward Provider)

### What It Is

Family Reward (codebase: `kuota`) is a production-ready quota/reward platform.
**Family Reward must never be modified.**

### How Odyssey Integrates

Odyssey **does not** issue or track real rewards. Achievement data is stored
in `odyssey_*` tables only.

A **signal** interface (for Phase 5) is planned but not implemented in the MVP.
When activated, Odyssey would write a minimal achievement signal that Family
Reward consumes and independently decides upon. See
[ADR-004](decisions/ADR-004-reward-integration.md) for the full decision.

### What This Means Today

For the MVP, there is **no Family Reward integration**. No signal table, no
webhooks, no polling. The Family Reward boundary is documented only as a
future option.

---

## 3. Supabase (Data Store)

### What It Is

A single Supabase (PostgreSQL) project shared with Family Reward.

### How Odyssey Uses It

- Odyssey creates **only new tables** prefixed with `odyssey_`.
- Odyssey uses the Supabase REST API (PostgREST) via the service-role key,
  following the same pattern Family Reward uses.
- Row-level security (RLS) is enabled on all `odyssey_*` tables with a
  service-role policy (identical to Family Reward's convention), since
  per-row access control is enforced at the application layer.
- Feature flags share the existing `system_configs` table (read-only) using
  `odyssey_`-prefixed keys to avoid collisions.

### Tables (Planned MVP)

| Table | Purpose |
|---|---|
| `odyssey_user_profiles` | Explorer data, UID, crew, role, level, XP. |
| `odyssey_crews` | Family-group metadata. |
| `odyssey_quests` | Quest instances (per crew). |
| `odyssey_challenges` | Individual challenge instances. |
| `odyssey_realm_progress` | Shared realm progress and story branch. |
| `odyssey_creative_items` | Creative-space contributions (Story/Comic/Photo/Video). |
| `odyssey_daily_turns` | One-per-user-per-day turn tracking. |
| `odyssey_achievements` | Personal and group milestones. |

The `odyssey_reward_signals` table from earlier drafts is **removed** from
the MVP. It will be introduced as part of the Phase 5 integration, as
documented in [ADR-004](decisions/ADR-004-reward-integration.md).

See [ADR-003](decisions/ADR-003-database.md).

---

## 4. Deployment (Reference Pattern)

Odyssey may follow the same deployment model as Family Reward (Go serverless
functions behind a frontend served as static assets), but this is a
**reference pattern**, not a coupling. The architecture should remain
flexible enough to switch to a single Go service or containerized deployment
if that better fits the family's needs.

See [ADR-001](decisions/ADR-001-project-scope.md) and
[ADR-003](decisions/ADR-003-database.md).

---

## Integration Principles

- **Stable interfaces only.** Integrations cross system boundaries through
  documented, versioned interfaces (Firestore document shape, REST endpoints,
  shared env config, signal tables). We never reach into another system's
  internal code.
- **Adapter isolation.** External system details (Firestore, REST API shapes,
  env var names) live in adapter layers, never in the game domain.
- **Read-only where possible.** Odyssey reads Gatekeeper Firestore data; it
  never writes to Gatekeeper's collections.
- **Fail open/closed by design.** If an integration is temporarily
  unavailable, Odyssey degrades gracefully (see Principles P2).
- **No shared deployments.** Odyssey is independently deployable from both
  Gatekeeper and Family Reward.
