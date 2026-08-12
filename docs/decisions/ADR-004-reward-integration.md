# ADR-004: Reward Integration

- **Date:** 2026-08-02
- **Status:** Accepted (integration deferred to Phase 5)
- **Deciders:** Lead Architect / Product Architect

## Context

Family Reward (`kuota`) is a production reward platform. It manages real-world
rewards (quota, cash) and has a complex business workflow around claims,
payouts, and eligibility.

Odyssey is **not** a reward application. Inside Odyssey there is no money and
no real-world payout. However, Odyssey tracks achievements and milestones that
*could* be meaningful signals for Family Reward to act upon — for example,
"family completed 50 missions" or "every member reached Explorer Level 10."

The question is: how, if at all, does Odyssey communicate these achievements
to Family Reward without tight coupling?

## Decision

Odyssey **does not issue or track real rewards.** It stores all achievement
data in its own `odyssey_*` tables. If and when a real-world reward is
warranted, Odyssey emits a **minimal achievement signal** that Family Reward
consumes through a **stable, narrow interface**.

### No Signal Table in the MVP

**The `odyssey_reward_signals` table is not created in the MVP.** It will be
introduced as a new table only when the Phase 5 integration is activated. Until
then:

- Achievement data is stored in `odyssey_achievements`, `odyssey_collections`, and
  `odyssey_missions` — all fully queryable for retrospective signal generation.
- No webhooks, no polling, no signal emission code exists.
- The integration contract is documented here as a decision record and
  revisited in Phase 5.

### Future Integration Design

When activating Phase 5, the following design will be used:

1. **Signal table.** A new `odyssey_reward_signals` table is created with a
   minimal schema:
   - `uid TEXT` — the family member's shared UID.
   - `achievement_code TEXT` — an opaque code defined by Odyssey.
   - `issued_at TIMESTAMPTZ` — when the signal was created.
   - `consumed BOOLEAN` — whether Family Reward has acknowledged it.

2. **Signal semantics.** A signal is a *suggestion*, not a command. Family
   Reward independently:
   - Validates the signal's authenticity (shared secret or signed payload).
   - Applies its own rules (eligibility, caps, deduplication).
   - Decides whether to issue a real reward.

3. **No business-table sharing.** Family Reward reads only the
   `odyssey_reward_signals` table — never Odyssey's game-state tables.
   The signal table is the **only** cross-system data surface, designed as an
   explicit interface.

4. **No tight coupling.** The signal table name and schema are the only
   shared contract. Neither system imports the other's code.

5. **Notification mechanism.** Family Reward will be notified via either:
   - A webhook (Odyssey POSTs to a Family Reward endpoint with a shared
     secret), **or**
   - Family Reward polls the `odyssey_reward_signals` table directly via
     Supabase REST.

   The specific mechanism will be decided collaboratively when activation is
   imminent.

## Alternatives Considered

### R1: Emit a reward signal for every milestone

**Rejected.** This would couple Odyssey's game design to Family Reward's
reward cadence, creating noise and potential abuse. Signals should be rare
and meaningful, not a firehose of events.

### R2: Let Family Reward read Odyssey's game tables directly

**Rejected.** This tightly couples Family Reward to Odyssey's internal schema.
Any schema migration in Odyssey would risk breaking Family Reward. The
signal table (when introduced) is a deliberate, minimal interface that
insulates both systems.

### R3: Make Family Reward the reward engine for Odyssey

**Rejected.** This inverts the relationship — Odyssey would become dependent
on Family Reward's business logic, payout flows, and admin tooling. Family
Reward is for real rewards; Odyssey is for play. They collaborate, not merge.

### R4: Real-time event streaming (Kafka / WebSocket mesh)

**Rejected for MVP.** Over-engineered for a ~8-user family group. A table
(with optional webhook) is sufficient, simple, and debuggable.

### R5: Ship the signal table in the MVP but leave it empty

**Rejected.** Shipping an unused table adds schema surface area and invites
premature decisions about signal semantics. The table is introduced only when
a real consumer (Family Reward) is ready to read it.

## Consequences

### Positive

- Odyssey ships its full game experience without any reward-integration
  surface area.
- When integration activates, it is a single, well-defined interface — easy
  to build, test, and tear down if it proves unnecessary.
- Achievement data is never lost: it is stored in `odyssey_` tables from day
  one, ready for retrospective signal generation.
- Family Reward retains full autonomy over whether, when, and how to issue
  real rewards.

### Negative

- If the family expects Odyssey achievements to translate to rewards, the
  integration gap may be disappointing. This is managed by clear messaging:
  Odyssey is a game; rewards are Family Reward's domain.
- No signal table means the integration cannot be wired up without a schema
  change. This is a deliberate trade-off for MVP simplicity.

## References

- [Integrations](../integrations.md)
- [Non-Goals](../non-goals.md)
- [ADR-001: Project Scope](ADR-001-project-scope.md)
- [ADR-003: Database](ADR-003-database.md)
- Family Reward reference: `pkg/shared/gts26_reward.go`,
  `pkg/shared/gts26_claim_audit.go`
