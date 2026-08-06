# ADR-005: Authentication Compatibility Assessment

- **Date:** 2026-08-06
- **Status:** Accepted
- **Deciders:** Lead Architect / Product Architect

## Context

Odyssey needs an authentication and authorization system. A production-ready device-trust authentication provider already exists: **Gatekeeper**. Gatekeeper is an Android app that writes device status to a Firestore database. Family Reward currently uses this to implement a **BOTH login mode** (device compliance + password credential). Building a new authentication system from scratch is explicitly out of scope for this project.

## Problem Statement

How can Odyssey reuse Gatekeeper's BOTH login flow while remaining fully independent of its implementation details (Firestore, Firebase)? The Odyssey domain layer must remain pristine and completely unaware of external dependencies.

## Assessment Scope

The assessment evaluates the feasibility of integrating Odyssey with the existing Gatekeeper Firestore document structure (`users/{PARENT_ID}/children/{uid}`) without compromising Odyssey's clean architecture or requiring modifications to the external Gatekeeper or Family Reward systems.

## Technical Comparison

- **Gatekeeper**: Android-based device trust app. Writes to Firestore.
- **Family Reward**: Reads Gatekeeper Firestore documents. Uses proprietary session secrets.
- **Odyssey**: Go-based backend. Needs to verify device compliance and issue its own secure sessions without sharing secrets or tightly coupling its domain to Firestore.

## Component Reusability Matrix

| Component | Reusability | Notes |
|---|---|---|
| Gatekeeper Device Trust | **High** | Core trust anchor. Can be read from Firestore without modification. |
| Family Reward Session Token | **Low** | Sessions are signed with deployment-specific secrets. Cannot be reused safely. |
| Firestore Database | **Medium** | Reusable as a read-only source of truth via an isolated adapter. |
| User Identity (UID) | **High** | Shared UID enables future cross-system integration. |

## Risks

1. **External Schema Changes:** If Family Reward or Gatekeeper changes the shape of the Firestore document, the Odyssey adapter will break.
2. **Availability Dependency:** The family must have the Gatekeeper app installed and online to log into Odyssey.
3. **First-login Device Binding:** Client-supplied device IDs must be trusted on first login (mitigated by password requirement).

## Decision

Odyssey will **reuse Gatekeeper's BOTH login mode** as its sole authentication path, implemented through a strict **port-and-adapter** pattern.

1. The domain layer will depend only on a generic `Authenticator` interface.
2. A `pkg/auth/firestore.go` adapter will implement this interface by reading from Gatekeeper's Firestore without importing Firestore into the domain layer.
3. Odyssey will issue its own HMAC-signed session tokens (not JWT), avoiding the sharing of secrets with Family Reward.

## Consequences

### Positive
- A strong, production-proven authentication path is reused without building a new framework.
- The UID is consistent across all systems (Gatekeeper, Family Reward, Odyssey).
- The domain layer remains fully isolated from Firestore/Firebase logic.

### Negative
- Users must use the Gatekeeper app to log in.
- The Firestore adapter requires updates if Gatekeeper changes its data model.

## Alternatives Considered

- **Supabase Auth:** Rejected. Creates a parallel authentication system, violating the "no new auth system" constraint and fragmenting family identity.
- **Skip Gatekeeper (Password Only):** Rejected. Weakens security posture and abandons the device-trust model.
- **Reuse Family Reward Session Tokens:** Rejected. Requires sharing signing secrets between independent deployments, violating security boundaries.
- **OAuth / OIDC Bridge:** Rejected. Gatekeeper is not an OAuth provider, and adding this would require modifying Gatekeeper, which is forbidden.

## Future Work

- Implement cross-system achievement signaling using the shared UID (see ADR-004) without direct database coupling.
- Add robust retry/fallback mechanisms in the Firestore adapter if read latency spikes.
