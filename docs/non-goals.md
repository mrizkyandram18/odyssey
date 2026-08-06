# Non-Goals

This document defines what Odyssey will **never** become. It exists to prevent
feature creep and to keep the team focused on the cooperative adventure core.

## Business & Monetization

- **NOT a reward application.** Real-world rewards are the sole domain of
  Family Reward. Odyssey may *signal* achievements to Family Reward through a
  stable interface, but it never issues, manages, or tracks real currency.
- **NO money.** No in-app purchases, no subscriptions, no microtransactions,
  no virtual currency that can be exchanged for real goods.
- **NO pay-to-win.** Progression is earned through play, never through payment
  of any kind.
- **NO gambling or loot boxes.** No randomized reward mechanics that simulate
  gambling. Any surprise elements are deterministic or cosmetic-only.

## Product Shape

- **NOT an MMORPG.** No massive open world, no thousands of concurrent users,
  no persistent realm servers. The scale is a handful of family members.
- **NOT a social network.** No public leaderboards, no friend-finding, no
  user-generated content from strangers, no viral loops designed for growth.
- **NOT an employee management tool.** No time tracking, no productivity
  scoring, no work-hour logging, no performance metrics.
- **NOT a work tracker.** No deadlines, no obligations, no "must complete"
  pressure. Quests are invitations, never requirements.
- **NOT a gamified productivity app.** We do not gamify real-life chores,
  habits, or obligations. The adventure is its own thing.
- **NO single-player mode.** Odyssey is cooperative by design. A solo mode
  would undermine the family-circle premise.

## Technical Direction

- **NOT a native mobile app.** No iOS/Android app stores. The delivery
  vehicle is a web PWA.
- **NO new authentication system.** We reuse Gatekeeper's BOTH login flow.
  We do not implement OAuth, JWT-based auth, or custom password hashing
  frameworks.
- **NO database migration from existing systems.** We create only new tables
  prefixed `odyssey_`. We never modify Family Reward or Gatekeeper tables.
- **NO real-time multiplayer infrastructure.** No WebSocket mesh, no WebRTC,
  no server-authoritative game state synchronization in real time. State is
  turn-based or near-real-time via standard API reads.
- **NOT platform-agnostic at the cost of clarity.** We optimize for the
  web/PWA experience. Native-only features are rejected even if technically
  possible.

## Growth & Scale

- **NO viral growth mechanics.** No referral programs, no share-to-unlock, no
  invite rewards designed to expand the user base. New family members are
  added deliberately by an existing member.
- **NO public deployment.** The system is for a private family group. There is
  no public signup page.
- **NO analytics harvesting.** Minimal, privacy-respecting usage data only.
  We do not build surveillance features or sell data.
