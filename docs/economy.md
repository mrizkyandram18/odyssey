# Economy

## Guiding Rule

> **There is no money in Odyssey.**

The Odyssey economy is entirely fictional. Experience (XP), Collections, and Gifts
are earned through play and spent on game progression. They can never be
exchanged for real currency, real goods, or any external value. This is a hard
constraint, not a design choice that can be revisited.

## Resources (MVP)

| Resource | Type | Source | Spend On |
|---|---|---|---|
| **Experience (XP)** | Soft currency | Completing missions, daily turns, creative contributions | Explorer Level, Family Level |
| **Collection** | Collectible | Mission completion, milestones, story fragments | Family gallery, trading (Phase 2) |
| **Gift** | Container | Mission completion, Explorer Level-ups | Reveals known Collections / cosmetics |
| **Story Fragment** | Collectible | Found during missions | Reassembling full stories, unlocking concept |
| **Inspiration** | Soft currency | Creative-space contributions, peer reactions | Creative tools, cosmetics |

A **Coin** soft currency is introduced in Phase 2 for family trading and
premium unlocks. Coins, like all Odyssey resources, are never exchangeable
for real money.

There is **no** exchange rate between any resource and real money. There is
**no** shop that accepts real payment. There is **no** paywall.

## Coins (Slice 2.1 earn + Slice 2.2 spend)

### Earn (unchanged)

| Event | Amount | Ledger source | Notes |
|---|---:|---|---|
| Mission completed | **+5** | `QUEST_COMPLETED` | Once per quest instance (`mission_id` in metadata) |
| Daily turn completed | **+1** | `DAILY_STREAK` | Once per successful daily consume |

### Spend (Slice 2.2 — one cosmetic)

| Item | Price | Ledger source | Notes |
|---|---:|---|---|
| **Gold Avatar Frame** (`avatar_frame_gold`) | **−3** | `COSMETIC_PURCHASE` | Once per explorer; unique ownership |

- Balance lives on `odyssey_user_profiles.coins`.
- History lives on `odyssey_reward_ledgers` (earns positive, spends negative).
- Ownership: `odyssey_cosmetic_unlocks` with `UNIQUE (uid, cosmetic_id)` (retry-safe).
- Equipped frame: `odyssey_user_profiles.avatar_frame` (`none` \| `gold`).
- **No trade, gift, marketplace, or premium real-money path.**
- Coins are fictional only — never real money.

## Earning

Resources are earned through the core loop:

- **Daily turn completion:** A small, fixed XP reward. Enough to feel
  meaningful but not enough to skip other activities. Also grants **+1 Coin**
  (Slice 2.1).
- **Mission completion:** XP + a **Gift** (with known contents) scaled by quest
  difficulty and number of participants. Group missions reward slightly more
  per person to incentivize cooperation. Also grants **+5 Coins** (Slice 2.1).
- **Creative contributions:** Inspiration is awarded when a family member
  contributes to a creative space and receives peer reactions.
- **First Discovery:** The first family member to complete a challenge type or
  find a hidden element gets a bonus Collection.
- **Streak bonus:** Consecutive daily participation yields a modest multiplier
  (capped) on daily-turn XP.

## Spending

- **Slice 2.2 Gold Avatar Frame:** costs **3 Coins** once; unlocks a gold
  portrait ring. Idempotent (cannot double-charge).
- **Gifts:** Contain known, fixed contents (Collections, cosmetics, creative
  tool unlocks). Contents are revealed before opening; there is no
  randomization that simulates gambling.
- **Creative tools:** Basic creative tools are free. Advanced tools (new brush
  types, sticker packs, custom colors) cost Inspiration.
- **Story reveals:** Occasionally a story scene is unlocked through a
  milestone rather than a cost, creating a sense of earned discovery.
- **Other cosmetics:** Further outfits/effects remain progression or future
  slices — not a large shop in 2.2.

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
- Shared (crew) resources feed journey unlocks and collective story branches.
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

XP, Collections, and achievements may, at the Odyssey team's discretion, be
surfaced to the **Family Reward** system as signals. Family Reward decides
independently whether to issue a real reward. See
[Integrations](integrations.md) and [ADR-004](decisions/ADR-004-reward-integration.md).
