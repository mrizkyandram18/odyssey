# ADR-002: Authentication

- **Date:** 2026-08-02
- **Status:** Accepted
- **Deciders:** Lead Architect / Product Architect

## Context

Odyssey needs an authentication and authorization system. A production-ready
authentication device-trust provider already exists: **Gatekeeper**. Gatekeeper
is an Android app that writes device status to a Firestore database, which
Family Reward uses to implement a **BOTH login mode** (device compliance +
password credential).

Building a new auth system is explicitly out of scope (see
[Non-Goals](../non-goals.md)). The question is how Odyssey reuses Gatekeeper's
BOTH login flow while remaining fully independent.

Critically, the Odyssey **domain layer** (quest, progression, economy logic)
must not know about Firestore, Firebase, or any auth implementation detail.
Identity is verified through an adapter that is swappable and isolated.

## Decision

Odyssey **reuses Gatekeeper's BOTH login mode** as its sole authentication
path, implemented through a **port-and-adapter** pattern.

### 1. The Authentication Provider Port

The domain layer defines an `Authenticator` interface:

```go
type Authenticator interface {
    // Verify checks Gatekeeper device trust + credential for the given UID.
    Verify(ctx context.Context, uid, credential string, device DevicePayload) error
}
```

`pkg/game` depends only on this interface. It never imports `firestore`,
`firebase`, or any external auth package.

### 2. The Firestore Adapter

`pkg/auth/firestore.go` implements `Authenticator` by reading the Gatekeeper
device document from Firestore at:

```
users/{PARENT_ID}/children/{uid}
```

It verifies:

- The document exists and `isOnline` is `true`.
- `details.appBuildNumber` is >= the minimum (configurable, default "49").
- `details.permissions` contains all required permissions.

This mirrors Family Reward's `ValidateGts26GatekeeperCompliance` but is
implemented independently in Odyssey's own `pkg/auth` package. **Gatekeeper's
code and Firestore documents are never modified by Odyssey.**

If Gatekeeper ever migrates off Firestore (e.g., to a REST API), only the
adapter changes — the port and all domain logic remain untouched.

### 3. Device Binding

On first successful compliance check, the adapter records the device ID in
Odyssey's `odyssey_user_profiles` table. Subsequent logins verify the device
ID matches the bound one (defending against credential theft from a
non-registered device).

### 4. Credential Verification

After Gatekeeper compliance passes, the user supplies a password (or PIN)
verified against `odyssey_user_profiles`. The password is hashed (bcrypt or
argon2). This is **credential storage and verification**, not a new
authentication system — the trust anchor remains Gatekeeper's device
compliance.

### 5. Session Issuance

After both compliance and credential checks pass, the backend issues an
**HMAC-signed session token** (not JWT). The format mirrors Family Reward's
session approach:

```
base64(JSON claims) + "." + base64(HMAC-SHA256(payload, secret))
```

- **Session kinds:** `user` (8 h TTL). A `setup` kind (30 min) is used for
  first-credential creation.
- **Signing secret:** `SESSION_SIGNING_SECRET` env var (falls back to
  `ADMIN_SECRET` only for bootstrap, same as Family Reward's pattern).
- **Storage:** `localStorage` on the PWA, sent as
  `Authorization: Bearer` and `X-User-Session` headers (matching Family
  Reward's dual-header convention for consistency).

### 6. UID as Shared Identity

The same UID used in Gatekeeper and Family Reward is used in Odyssey. This
enables future cross-system achievement signaling (see
[ADR-004](ADR-004-reward-integration.md)) without requiring Odyssey to read
or write Family Reward's tables.

## Alternatives Considered

### A1: Use Supabase Auth (new auth system)

**Rejected.** This would create a parallel authentication system, violating
the "no new auth system" constraint and the reuse principle. It would also
mean family members manage a separate Supabase Auth account, fragmenting
identity.

### A2: Skip Gatekeeper, use only a password

**Rejected.** This abandons device-trust authentication entirely, defeating
the purpose of reusing Gatekeeper and weakening the security posture for a
family-focused product.

### A3: Reuse Family Reward's session tokens directly

**Rejected.** Family Reward's sessions are signed with a secret known only to
its deployment. Sharing that secret would create tight coupling and a single
compromise affecting both systems. Sessions are intentionally per-system.

### A4: OAuth / OIDC bridge to Gatekeeper

**Rejected.** Gatekeeper is a device-trust app, not an OAuth provider. There
is no OAuth interface to consume. Introducing one would require modifying
Gatekeeper, which is forbidden.

### A5: Direct Firestore coupling in the domain

**Rejected.** Having `pkg/game` import Firestore directly would couple game
logic to an implementation detail, make unit testing harder, and violate the
modularity principle. The port-and-adapter pattern keeps the boundary clean.

## Consequences

### Positive

- Single, strong authentication path leverages a production-proven device-trust
  system.
- The UID is consistent across all three systems, enabling future integration
  without identity mapping.
- Port-and-adapter design isolates Firestore to a single, replaceable module.
- HMAC-signed sessions are simple, stateless, and familiar to anyone who
  worked on Family Reward.
- No new auth framework is built — implementation effort stays focused on the
  game.

### Negative

- The family must have the Gatekeeper app installed and online to log in to
  Odyssey. This is a deliberate trade-off (device trust over convenience).
- If Family Reward ever changes its Firestore document shape, the adapter in
  `pkg/auth` may need an update. This risk is mitigated by documenting the
  expected Firestore contract in [Integrations](../integrations.md).

### Migration Paths

None required. If Gatekeeper evolves or changes its backing store, the
Firestore adapter is the single place that needs updating. The
`Authenticator` port remains stable.

## References

- [Integrations](../integrations.md)
- [Architecture](../architecture.md) — Module Boundaries
- [ADR-004: Reward Integration](ADR-004-reward-integration.md)
- Family Reward reference: `pkg/shared/gatekeeper.go`, `pkg/shared/gts26_login.go`,
  `pkg/shared/session.go`, `pkg/shared/firebase.go`
