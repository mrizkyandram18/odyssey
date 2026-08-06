# Roadmap

## Philosophy

The roadmap is intentionally small and milestone-driven. Each phase ships a
complete, playable slice. Scope is **time-boxed**, not feature-maximized. We
follow the principle: **ship the smallest delightful loop first, then expand.**

Phases are additive. Later phases never break or replace earlier ones; they
extend them.

---

## Phase 0 — Foundation

**Status:** Complete

**Goal:** Establish the long-term foundation: vision, architecture decisions,
domain model, and coding standards. No application code is written.

**Deliverables:**

- All docs in `docs/` and `docs/decisions/`.
- Updated `CLAUDE.md`.
- Approved ADRs.

**Exit criteria:** Team sign-off on scope, auth adapter approach, database
strategy, and reward integration deferral.

---

## Phase 1 — MVP: The First Adventure (Weeks 1—4)

**Status:** Release Candidate (Conditional GO / Pending Live Verification)

**Goal:** Ship the smallest complete cooperative loop that delivers a real
"aha" moment.

**What ships:**

- Gatekeeper BOTH login (device trust via auth adapter + credential) → Odyssey session.
- A single **home screen** showing the family's crew level, daily turn, and active quests.
- **One realm** ("The Whispering Woods") with a short introductory story.
- **3 quests** in that realm, each with 2–3 challenges:
  - One solo quest (observation/research).
  - One relay quest (sequential legs).
  - One **Creative Quest** (a group Story submission).
- **Daily turn** system (one per calendar day): a micro-quest and a creative
  prompt.
- **Basic progression:** Explorer Level (XP), Chests with fixed contents,
  Relics, Realm Progress bar, crew level.
- **Realm Progress** — a shared bar that unlocks new quests and the next realm.
- **Minimal admin:** a single system config to toggle maintenance mode and set
  the gatekeeper minimum build number.

**Resources in MVP:** XP, Relics, Chests, Story Fragments, Inspiration.
**No** Adventure Tokens, **no** Coins. (Coin currency is Phase 2.)

**What does NOT ship:**

- Multiple realms (beyond the tutorial).
- Peer reactions / comments (sticker reactions deferred).
- Rich creative tools (only basic text + a simple draw submission in Creative Quests).
- Photo / Video / Comic submission kinds in the MVP (Story only; others added in Phase 2).
- Offline-first / offline sync.
- Family Reward integration or reward signals.
- Role mastery mechanics.

**Exit criteria:** A family of 3–4 can play through all 3 quests, complete
daily turns for a week, and see meaningful progression — entirely on a phone,
PWA installed.

---

## Phase 2 — Social Fabric (Weeks 5–8)

**Goal:** Add the connective tissue that makes the family feel like a crew.

**What ships:**

- **Peer reactions** (stickers) on creative contributions and quest logs.
- **Quest handoff notifications** for relay quests ("your turn").
- **Crew streak** tracking and Realm Progress milestones.
- **Expanded creative space:** drawing tools, color palettes, a shared text canvas.
- **Role rotation** UI and role-mastery flavor text (no stat bonuses).
- **Coin currency** — a soft currency for family trading (gifting Relics between
  Explorers) and premium unlocks.
- **Photo / Comic / Video** submission kinds for Creative Quests.
- **Push notifications** (PWA) for daily turns and handoffs.

**Exit criteria:** The family regularly coordinates through the app to complete
relay quests, and creative spaces feel like a shared living space.

---

## Phase 3 — Deeper Storytelling (Weeks 9–14)

**Goal:** Introduce narrative depth and branching.

**What ships:**

- **Branch choices** in quests — different story outcomes based on family
  decisions (no currency cost; choices are narrative forks).
- **Story Fragment** collectible mechanic with a completion reward (a Relic or
  Chest).
- **Second realm** unlocked through crew + Realm Progress.
- **Quest variety:** puzzle, movement-based, and research challenges.
- **Achievement log** (personal + group milestones) accessible from a
  journal screen.
- **Realm replay:** returning to completed realms reveals new dialogue and
  hidden story fragments.

**Exit criteria:** The family debates branching choices together and revisits
realms to find new content.

---

## Phase 4 — Richer Creation (Months 4–6)

**Goal:** Make creative spaces feel expressive and alive.

**What ships:**

- **Expanded creative tools:** stamp libraries, layering, export to story pages.
- **Comic Mode** — assemble Story and Comic submissions into illustrated
  chapters viewable in a comic-reader UI.
- **Custom crew banner** and shared realm themes.
- **Animated Explorer effects** unlocked through milestones.
- **Family gallery** — a scrollable showcase of all creative contributions
  (Story, Comic, Photo, Video).
- **Trading** — family members can gift Relics to each other (Phase 2 coins
  make this meaningful).

**Exit criteria:** Creative spaces are the most-visited area; family members
spend time there even without an active quest.

---

## Phase 5 — Long Arc & Integration (Months 6–?)

**Goal:** Multi-season storytelling, AI amplification, and the Family Reward
bridge.

**What ships:**

- **Seasonal realms** — a new realm theme every quarter, with a
  self-contained story arc.
- **Family Reward integration:** if deemed worthwhile, a minimal
  `odyssey_reward_signals` table is introduced and achievement signals are
  emitted. Family Reward consumes these via a stable interface to decide on
  real rewards. (See [ADR-004](decisions/ADR-004-reward-integration.md).)
- **AI-assisted story generation** (see [docs/ai.md](ai.md)) — optional, opt-in,
  always reviewed by the family.
- **Long-term replay:** quest lines that span months become replayable with
  new branches.

**What does NOT change:**

- No real money enters Odyssey.
- Gatekeeper remains the sole auth path.
- Database tables remain `odyssey_` prefixed.

---

## What Gets Canceled

A phase is **canceled** (not just deferred) if:

- Family engagement drops below the Phase 1 baseline.
- A mechanic causes friction or pressure within the family.
- The feature cannot be polished to a high standard within its timebox.

Cancellation is a feature, not a failure. We prefer fewer excellent features
to many mediocre ones.
