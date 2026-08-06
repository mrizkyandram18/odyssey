# ADR-001: Project Scope

- **Date:** 2026-08-02
- **Status:** Accepted
- **Deciders:** Lead Architect / Product Architect

## Context

A greenfield project, **Odyssey**, is being established. Two production systems
already exist and must never be modified:

1. **Gatekeeper** — an Android device-trust application and authentication
   provider.
2. **Family Reward** (`kuota`) — a production reward/quota platform using
   Supabase (PostgreSQL) and a Go + Vite stack.

Odyssey shares the same Supabase project and the same Gatekeeper identity
system, but its purpose, data, users, and code must be **fully independent**.

## Decision

Odyssey is a **new, independent cooperative-adventure platform** for a private
family group of ~8 users. It is not a reward application, not a work tracker,
and not a public product.

Specifically, we decide:

1. **Separate codebase.** Odyssey lives in its own repository
   (`D:/Personal/Projects/Odyssey`). It has no shared Go modules, no shared
   frontend packages, and no shared CI configuration with Family Reward.
2. **Read-only reference.** Family Reward may be read as a *reference* to
   understand auth flow, API style, session handling, and database naming
   conventions. Its business logic is never copied.
3. **No tight coupling.** Odyssey integrates with Gatekeeper and Family Reward
   only through stable, narrow interfaces (documented in
   [ADR-002](ADR-002-authentication.md), [ADR-003](ADR-003-database.md), and
   [ADR-004](ADR-004-reward-integration.md)).
4. **Scope is a hard constraint.** The product vision is cooperative play, not
   monetization, not scale, not gamified productivity. See
   [Non-Goals](../non-goals.md) and [Vision](../vision.md).
5. **MVP is intentionally small.** The first phase ships one realm, 3 quests,
   daily turns, and basic progression. Everything beyond that is explicitly
   future work.

## Consequences

### Positive

- Clear boundaries reduce the risk of accidental coupling or shared bugs.
- Read-only reference to Family Reward lets us borrow proven patterns (HMAC
  sessions, Supabase REST, serverless handler layout) without inheriting
  technical debt.
- A small MVP ensures a complete, polished loop before expansion.

### Negative

- Some patterns are duplicated (e.g., session management, Supabase REST
  helpers). This is accepted: duplication across independent systems is
  cheaper than coupling.
- Family members maintain a separate credential for Odyssey (in BOTH login
  mode). This is a minor UX cost, justified by system independence.

### Neutral

- Future convergence is possible but never automatic. Any shared interface
  must be formalized through a new ADR.

## References

- [Vision](../vision.md)
- [Non-Goals](../non-goals.md)
- [ADR-002: Authentication](ADR-002-authentication.md)
- [ADR-003: Database](ADR-003-database.md)
- [ADR-004: Reward Integration](ADR-004-reward-integration.md)
