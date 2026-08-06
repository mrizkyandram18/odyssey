# Economy

## Guiding Rule

> **There is no money in Odyssey.**

The Odyssey economy is entirely fictional. Experience (XP), Relics, and Chests
are earned through play and spent on game progression. They can never be
exchanged for real currency, real goods, or any external value. This is a hard
constraint, not a design choice that can be revisited.

## Resources (MVP)

| Resource | Type | Source | Spend On |
|---|---|---|---|
| **Experience (XP)** | Soft currency | Completing quests, daily turns, creative contributions | Explorer Level, Crew Level |
| **Relic** | Collectible | Quest completion, milestones, story fragments | Crew gallery, trading (Phase 2) |
| **Chest** | Container | Quest completion, Explorer Level-ups | Reveals known Relics / cosmetics |
| **Story Fragment** | Collectible | Found during quests | Reassembling full stories, unlocking lore |
| **Inspiration** | Soft currency | Creative-space contributions, peer reactions | Creative tools, cosmetics |

A **Coin** soft currency is introduced in Phase 2 for family trading and
premium unlocks. Coins, like all Odyssey resources, are never exchangeable
for real money.

There is **no** exchange rate between any resource and real money. There is
**no** shop that accepts real payment. There is **no** paywall.

## Earning

Resources are earned through the core loop:

- **Daily turn completion:** A small, fixed XP reward. Enough to feel
  meaningful but not enough to skip other activities.
- **Quest completion:** XP + a **Chest** (with known contents) scaled by quest
  difficulty and number of participants. Group quests reward slightly more
  per person to incentivize cooperation.
- **Creative contributions:** Inspiration is awarded when a family member
  contributes to a creative space and receives peer reactions.
- **First Discovery:** The first family member to complete a challenge type or
  find a hidden element gets a bonus Relic.
- **Streak bonus:** Consecutive daily participation yields a modest multiplier
  (capped) on daily-turn XP.

## Spending

- **Chests:** Contain known, fixed contents (Relics, cosmetics, creative
  tool unlocks). Contents are revealed before opening; there is no
  randomization that simulates gambling.
- **Creative tools:** Basic creative tools are free. Advanced tools (new brush
  types, sticker packs, custom colors) cost Inspiration.
- **Story reveals:** Occasionally a story scene is unlocked through a
  milestone rather than a cost, creating a sense of earned discovery.
- **Cosmetics:** Outfits, effects, and realm themes are unlocked through
  Explorer Levels and milestones. Rare cosmetics are high-XP goals, never
  purchases.

## Balance Philosophy

- Resources are **generous by default.** A family that plays regularly should
  never feel starved for the content they want.
- Scarcity is used **deliberately** to create meaningful choice, not
  frustration. When a choice requires resources, it should be because the
  choice between options is interesting.
- All progression is **reversible in spirit**: if a family spends Inspiration
  on a tool and doesn't use it, future Inspiration can always be earned. We
  never put essential content behind a high cost.

## Group vs. Individual Balance

- Individual resources feed Explorer Level and personal cosmetic unlocks.
- Shared (crew) resources feed realm unlocks and collective story branches.
- Some creative-space tools are **shared**: once one family member unlocks a
  tool, it is available to all. This reinforces cooperation over competition.

## Anti-Exploitation

Because the family is a trusted, closed circle:

- There is no energy system that blocks play.
- There are no randomized drops that could be "farmed."
- There is no player-vs-player competition for resources.
- Daily turns reset at a family-local midnight to respect different time zones
  within the family.

## Reward Integration

XP, Relics, and achievements may, at the Odyssey team's discretion, be
surfaced to the **Family Reward** system as signals. Family Reward decides
independently whether to issue a real reward. See
[Integrations](integrations.md) and [ADR-004](decisions/ADR-004-reward-integration.md).
