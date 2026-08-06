# Domain Model

This document defines the core domain entities of Odyssey and their
relationships. It is the single source of truth for the game's data model.
Entities are grouped by layer: **Identity**, **Progression**, **Content**,
**Creative**, and **Integration**.

## Entity Map

```
Player ──► Crew ──► WorldState ──► Realm ──► Chapter ──► Quest ──► QuestInstance
  │                    │               │         │         │         │
  │                    │               │         │         │         │
  │              RealmProgress         │         │         │         │
  │                    │               │         │         │         │
  └─► ExplorerLevel     │               │         │         │         │
  └─► Relic ◄───Chest──┘               │         │         │         │
  └─► Achievement                     │         │         │         │
  └─► Story ──► Prompt ──► Submission  │         │         │         │
                                       │         │         │
                        Challenge ──────► QuestInstance
                             │
                      Submission (per Challenge)
```

---

## Identity Layer

### Player

A single family member. Identified by the **shared UID** (same UID used by
Gatekeeper and Family Reward).

| Field | Type | Notes |
|---|---|---|
| `uid` | TEXT (PK) | Shared identity across all three systems |
| `crew_id` | TEXT | FK → Crew |
| `explorer_name` | TEXT | Chosen display name |
| `explorer_level` | INTEGER | Starts at 1; derived from `xp` |
| `xp` | BIGINT | Accumulated experience |
| `role` | TEXT | `SEEKER` / `BUILDER` / `GUIDE` — narrative framing, **no stat bonuses** |
| `created_at` / `updated_at` | TIMESTAMPTZ | |

### Crew

The family group (~8 players). Progression and unlocks are tracked at the crew
level.

| Field | Type | Notes |
|---|---|---|
| `id` | TEXT (PK, UUID) | |
| `name` | TEXT | Optional display name |
| `created_at` | TIMESTAMPTZ | |

---

## Progression Layer

### Explorer Level

Each player has an **Explorer Level** (not "Avatar Level" — see
[ADR-001](decisions/ADR-001-project-scope.md)). Leveling is smooth: the XP
curve is gentle so casual play steadily advances. Each level grants a small
pool of rewards (Relics, Chests, or Creative-tool unlocks).

There is also a **Crew Level** (see Realm Progress) derived from the sum of
quest completions.

### Relic

A collectible item tied to achievement or story completion. Relics are personal
but displayed on a shared crew gallery. They are **not currency** — they are
collected, admired, and (Phase 2) traded within the family.

### Chest

A reward container earned through quest completion or Explorer Level-ups.
MVP Chests have **known, fixed contents** (not randomized in a way that
simulates gambling — see Principles P3 in [principles](principles.md)).

### Achievement

A milestone (personal or group). Achievements award XP, a Relic, or unlock
content.

| Field | Type | Notes |
|---|---|---|
| `code` | TEXT (PK) | Opaque code, e.g. `FIRST_QUEST`, `CREW_LEVEL_5` |
| `title` | TEXT | |
| `description` | TEXT | |
| `kind` | TEXT | `PERSONAL` / `GROUP` |
| `threshold` | INTEGER | What triggers it |

---

## Content Layer

### Realm

A themed area of the world. Realms are unlocked sequentially through crew
progress.

### RealmProgress (Community Progress)

Shared progress state for a crew within a Realm. Tracks story branch
selection, unlocked chapters, and realm-specific group milestones.

| Field | Type | Notes |
|---|---|---|
| `crew_id` | TEXT | FK → Crew |
| `realm` | TEXT | |
| `status` | TEXT | `LOCKED` / `ACTIVE` / `COMPLETE` |
| `story_branch` | TEXT | Which narrative branch was chosen |
| `unlocked_at` | TIMESTAMPTZ | |

### Chapter

A story chapter within a Realm. Chapters contain Quests.

### Quest (Template)

A reusable quest definition: challenges, story beats, and rewards.

### QuestInstance

A quest taken on by a specific Crew. Tracks status (`PENDING` / `ACTIVE` /
`DONE`) and completion timestamps. Multiple crews can run the same Quest
template independently.

### Challenge

A single task within a QuestInstance. Each challenge has a type
(`OBSERVATION`, `RESEARCH`, `PUZZLE`, `MOVEMENT`, `DRAW`, `WRITE`) and a
target state.

### Submission

A player's response to a Challenge. Content may be text, image, or voice note.
See the **Creative Layer** for Story-submission specializations.

---

## Creative Layer

### Story

A player-authored narrative piece produced in response to a **Creative Quest**.
Story is the umbrella entity; the actual output is a Submission of a specific
kind (see below).

### Prompt

A creative writing prompt or theme that spawns Story Submissions.

### Submission Kinds (First-Class Creative Outputs)

Creative Quests accept these submission types as first-class content:

| Kind | Description |
|---|---|
| `STORY` | A written narrative snippet. |
| `COMIC` | A multi-panel illustrated submission. |
| `PHOTO` | A real-world photograph tied to a challenge. |
| `VIDEO` | A short (≤ 30 s) video contribution. |

These are stored as `odyssey_creative_items` rows with a `kind` discriminator
and a `payload` JSONB column (base64 for small MVP payloads).

---

## Integration Layer

### Authentication Provider (Port)

Odyssey authenticates exclusively through the **Gatekeeper BOTH login flow**.
The domain layer never references Firestore or any auth implementation detail
directly. Instead, an `Authenticator` port is defined in `pkg/auth`, and an
adapter (e.g., `FirestoreAuthenticator`) implements it.

The adapter:
- Reads the Gatekeeper device document from Firestore (read-only).
- Verifies device online status, build number, and permissions.
- Returns a boolean + error to the domain.

This keeps the domain decoupled: if Gatekeeper's backing store ever changes
(e.g., to a REST endpoint), only the adapter changes.

See [ADR-002](decisions/ADR-002-authentication.md).

### Reward (Deferred — Phase 5)

A real-world reward issued by Family Reward. Odyssey never creates or tracks
rewards itself. It may emit a minimal achievement signal to Family Reward
through a future integration (see [ADR-004](decisions/ADR-004-reward-integration.md)).

| Field | Type | Notes |
|---|---|---|
| `uid` | TEXT | Player UID |
| `achievement_code` | TEXT | Opaque Odyssey code |
| `issued_at` | TIMESTAMPTZ | |

### Season (Future)

A time-bounded progression arc. Seasons are planned for Phase 3 and are not
part of the MVP domain.

---

## Layer Architecture

The domain is organized into three layers:

### Definition Layer (`pkg/game/content`)

Owns ALL game content definitions. These are data-only types with no
runtime behavior. Each definition is independent and maps to a database table.

| Definition | DB Table | Description |
|---|---|---|
| `RealmDefinition` | `odyssey_realm_definitions` | Themed area of the world |
| `ChapterDefinition` | `odyssey_chapter_definitions` | Story chapter within a Realm |
| `QuestDefinition` | `odyssey_quest_definitions` | Reusable quest template |
| `CreativePromptDefinition` | `odyssey_creative_prompt_definitions` | Creative prompt for quests |
| `ChestDefinition` | `odyssey_chest_definitions` | Chest template (admin-managed) |
| `DropTableEntry` | `odyssey_drop_tables` | Rarity-weight pairs for chests |
| `RelicDefinition` | `odyssey_relic_definitions` | Relic template (admin-managed) |
| `AchievementDefinition` | `odyssey_achievement_definitions` | Milestone definition |
| `SeasonDefinition` | `odyssey_season_definitions` | Time-bounded progression arc |
| `LoreDefinition` | `odyssey_lore_definitions` | Narrative lore entry |

### Runtime Layer (`pkg/game`, `pkg/content`)

Owns game logic and the ContentService. Runtime entities are created
from definitions but carry player-specific state.

| Runtime Entity | Source | Description |
|---|---|---|
| `QuestInstance` | QuestDefinition + crew | A quest taken on by a crew |
| `QuestCompletion` | QuestInstance + challenges | Result of completing a quest |
| `PlayerRelic` | RelicDefinition + player | Player's owned relic instance |
| `PlayerAchievement` | AchievementDefinition + player | Player's earned achievement |
| `Chest` | ChestDefinition + player | Player's owned chest instance |

### Player Layer (`pkg/game` domain entities)

Owns player-specific state. These entities are per-player and are NOT
content definitions.

| Entity | Table | Description |
|---|---|---|
| `Player` | `odyssey_user_profiles` | Individual family member |
| `Crew` | `odyssey_crews` | Family group |
| `RealmProgress` | `odyssey_realm_progress` | Crew's progress in a realm |
| `Relic` | `odyssey_relics` | Awarded relic instance |
| `Chest` | `odyssey_chests` | Awarded chest instance |
| `Achievement` | `odyssey_achievements` | Earned achievement |

---

## Content Entity → Table Mapping

| Entity | DB Table | Layer |
|---|---|---|
| RealmDefinition | `odyssey_realm_definitions` | Definition |
| ChapterDefinition | `odyssey_chapter_definitions` | Definition |
| QuestDefinition | `odyssey_quest_definitions` | Definition |
| CreativePromptDefinition | `odyssey_creative_prompt_definitions` | Definition |
| ChestDefinition | `odyssey_chest_definitions` | Definition |
| DropTableEntry | `odyssey_drop_tables` | Definition |
| RelicDefinition | `odyssey_relic_definitions` | Definition |
| AchievementDefinition | `odyssey_achievement_definitions` | Definition |
| SeasonDefinition | `odyssey_season_definitions` | Definition |
| LoreDefinition | `odyssey_lore_definitions` | Definition |
| QuestInstance | `odyssey_quests` | Runtime |
| PlayerRelic | `odyssey_player_relics` | Player |
| Achievement | `odyssey_achievements` | Player |
| Chest | `odyssey_chests` | Player |
