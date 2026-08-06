# Principles

These principles govern every design and implementation decision in Odyssey.
They are listed in priority order: when a trade-off arises, the principle higher
on the list wins.

## P1: Play First, Optimization Second

The product exists to be played. If a decision improves the game feel,
engagement, or fun at the expense of implementation simplicity, the game wins.
Conversely, if an optimization cannot be felt in play, it is deferred.

## P2: Family Trust, Not Adversarial Defenses

Odyssey serves a closed family circle. Security is important, but the
threat model is "lost phone," not "malicious insider." We prefer clear error
messages, easy recovery, and graceful degradation over punitive blocks and
complex anti-abuse systems.

## P3: No Real Stakes

There is no money, gambling, or real-world consequence for losing. Losing a
challenge feels like losing a board game — momentarily disappointing, never
harmful. This shapes every balance decision: risk and reward are entirely
fictional and reversible.

## P4: Cooperation Over Competition

Every mechanic must create an opportunity for family members to work toward a
shared goal. Pure head-to-head competition between family members is avoided
unless it directly enables a cooperative outcome (e.g., a relay where each
person's leg advances the group).

## P5: Mobile-First, Offline-Tolerable

The primary device is a smartphone. The UI must work flawlessly on a 5-inch
screen with a thumb. Network availability inside the family home should not
block the core loop: local progress must be possible offline and synced when
connectivity returns.

## P6: PWA-First, Web-Native

Odyssey is delivered as a web application (PWA). No app-store gatekeeping, no
native build complexity. The web is the platform. Installation to the home
screen is optional and encouraged.

## P7: Reuse, Do Not Rebuild

Authentication is Gatekeeper. Data storage is Supabase. We do not invent new
auth systems, new ORMs, or new protocols. We compose existing, proven
building blocks and integrate through stable, narrow interfaces.

## P8: One Job Per Module

The codebase is a collection of small, single-purpose modules. Each module
has one reason to change. Cross-cutting concerns (auth, persistence, logging)
are handled by thin adapters in the shared layer, never scattered across
feature code.

## P9: Simplicity Survives Reviews

When in doubt, choose the simpler design. A system that is easy to understand
today will be easy to extend tomorrow. Complexity must be invited, not
defaulted into.

## P10: Ship the Smallest Delight

The MVP must contain a complete, tiny loop that feels good to complete. A
small, polished experience that works end-to-end beats a large, half-built
feature set. Expansion is always additive.

## P11: Intentional Growth

Feature scope is a constraint, not an ambition. Every new feature must pass a
three-question test:

1. Does it reinforce the cooperative adventure vision?
2. Can it be built and polished before the next family session?
3. Does it compose cleanly with existing modules?

If any answer is "no," the feature is deferred or rejected.
