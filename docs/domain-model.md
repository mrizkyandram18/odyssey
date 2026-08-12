# Domain Model

This document defines the core domain entities of Odyssey and their
relationships. It is the single source of truth for the game's data model.
Entities are grouped by layer: **Identity**, **Progression**, **Content**,
**Creative**, and **Integration**.

## Entity Map

```
Player ──► Family ──► WorldState ──► Journey ──► Course ──► Mission ──► QuestInstance
  │                    │               │         │         │         │
  │                    │               │         │         │         │
  │              JourneyProgress         │         │         │         │
  │                    │               │         │         │         │
  └─► ExplorerLevel     │               │         │         │         │
  └─► Collection ◄───Gift──┘               │         │         │         │
  └─► Achievement                     │         │         │         │
  └─► Story ──► Prompt ──► Submission  │         │         │         │
                                       │         │         │
                        Exercise ──────► QuestInstance
                             │
                      Submission (per Exercise)
```

---

## Identity Layer

### Player

A single family member. Identified by the **shared UID** (same UID used by
Gatekeeper and Family Reward).

| Field | Type | Notes |
|---|---|---|
| `uid` | TEXT (PK) | Shared identity across all three systems |
| `family_id` | TEXT | FK → Family |
| `explorer_name` | TEXT | Chosen display name |
| `explorer_level` | INTEGER | Starts at 1; derived from `xp` |
| `xp` | BIGINT | Accumulated experience |
| `role` | TEXT | `SEEKER` / `BUILDER` / `GUIDE` — narrative framing, **no stat bonuses** |
| `created_at` / `updated_at` | TIMESTAMPTZ | |

### Family

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
pool of rewards (Collections, Gifts, or Creative-tool unlocks).

There is also a **Family Level** (see Journey Progress) derived from the sum of
quest completions.

### Collection

A collectible item tied to achievement or story completion. Collections are personal
but displayed on a shared crew gallery. They are **not currency** — they are
collected, admired, and (Phase 2) traded within the family.

### Gift

A reward container earned through quest completion or Explorer Level-ups.
MVP Gifts have **known, fixed contents** (not randomized in a way that
simulates gambling — see Principles P3 in [principles](principles.md)).

| Field | Type | Notes |
|---|---|---|
| `id` | TEXT (PK, UUID) | |
| `uid` | TEXT | FK → Player |
| `gift_slug` | TEXT | FK → ChestDefinition |
| `reward_relic` | TEXT | (Phase 3) Explicit mapped relic slug inside the chest |

### Balance Configuration

System-wide economy configuration (injected at runtime).

| Field | Type | Notes |
|---|---|---|
| `xp_per_level` | INTEGER | Default 500 (Phase 3 progression pacing) |
| `max_new_missions_per_day` | INTEGER | Default 1 (Phase 3 quest pacing) |

### Achievement

A milestone (personal or group). Achievements award XP, a Collection, or unlock
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

### Journey

A themed area of the world. Journeys are unlocked sequentially through crew
progress.

### JourneyProgress (Community Progress)

Shared progress state for a crew within a Journey. Tracks story branch
selection, unlocked chapters, and journey-specific group milestones.

| Field | Type | Notes |
|---|---|---|
| `family_id` | TEXT | FK → Family |
| `journey` | TEXT | |
| `status` | TEXT | `LOCKED` / `ACTIVE` / `COMPLETE` |
| `story_branch` | TEXT | Which narrative branch was chosen |
| `unlocked_at` | TIMESTAMPTZ | |

### Course

A story course within a Journey. Courses contain Missions.

### Mission (Template)

A reusable quest definition: exercises, story beats, and rewards.

| Field | Type | Notes |
|---|---|---|
| `slug` | TEXT (PK) | |
| `reward_relic` | TEXT | (Phase 3) Explicit relic slug granted upon completion (deterministic) |

### QuestInstance

A quest taken on by a specific Family. Tracks status (`PENDING` / `ACTIVE` /
`DONE`) and completion timestamps. Multiple families can run the same Mission
template independently.

| Field | Type | Notes |
|---|---|---|
| `id` | TEXT (PK, UUID) | |
| `family_id` | TEXT | FK → Family |
| `mission_slug` | TEXT | FK → QuestDefinition |
| `status` | TEXT | `PENDING` / `ACTIVE` / `DONE` |
| `started_at` | TIMESTAMPTZ | Time when the quest transitioned to ACTIVE |
| `started_by` | TEXT | UID of the explorer who started the quest |

### Exercise

A single task within a QuestInstance. Each challenge has a type
(`OBSERVATION`, `RESEARCH`, `PUZZLE`, `MOVEMENT`, `DRAW`, `WRITE`) and a
target state.

### Submission

A player's response to a Exercise. Content may be text, image, or voice note.
### LearningConcept

A collectible narrative fragment tied to a specific journey. Secret/hidden fragments (`is_hidden = true`) are revealed when returning to complete a **Journey Replay**.

| Field | Type | Notes |
|---|---|---|
| `slug` | TEXT (PK) | Unique fragment identifier |
| `journey` | TEXT | FK → Journey |
| `title` | TEXT | Display title |
| `content` | TEXT | Story text snippet |
| `set_name` | TEXT | Grouping set name |
| `is_hidden` | BOOLEAN | `true` if unlocked only via journey replay |

### PlayerStoryFragment

Tracks a player's discovery of a specific `LearningConcept`.

| Field | Type | Notes |
|---|---|---|
| `id` | BIGINT (PK) | Auto-increment primary key |
| `uid` | TEXT | FK → Player |
| `family_id` | TEXT | FK → Family |
| `fragment_slug` | TEXT | FK → LearningConcept |
| `discovered_at` | TIMESTAMPTZ | Timestamp of discovery |

---

## Creative Layer

### Story

A player-authored narrative piece produced in response to a **Creative Mission**.
Story is the umbrella entity; the actual output is a Submission of a specific
kind (see below).

### Prompt

A creative writing prompt or theme that spawns Story Submissions.

### Submission Kinds (First-Class Creative Outputs)

Creative Missions accept these submission types as first-class content:

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

### Definition Layer (`pkg/content` and `pkg/game/content`)

Owns ALL game content definitions. These are data-only types with no
runtime behavior. Each definition is independent and maps to a database table.

| Definition | DB Table | Description |
|---|---|---|
| `RealmDefinition` | `odyssey_journey_definitions` | Themed area of the world |
| `ChapterDefinition` | `odyssey_course_definitions` | Story course within a Journey |
| `QuestDefinition` | `odyssey_quest_definitions` | Reusable quest template |
| `CreativePromptDefinition` | `odyssey_creative_prompt_definitions` | Creative prompt for missions |
| `ChestDefinition` | `odyssey_chest_definitions` | Gift template (admin-managed) |
| `DropTableEntry` | `odyssey_drop_tables` | Rarity-weight pairs for gifts |
| `RelicDefinition` | `odyssey_relic_definitions` | Collection template (admin-managed) |
| `AchievementDefinition` | `odyssey_achievement_definitions` | Milestone definition |
| `SeasonDefinition` | `odyssey_season_definitions` | Time-bounded progression arc |
| `LoreDefinition` | `odyssey_concept_definitions` | Narrative concept entry |
| `LearningConcept` | `odyssey_story_fragments` | Collectible story fragment definition |

### Runtime Layer (`pkg/game`, `pkg/content`)

Owns game logic and the ContentService. Runtime entities are created
from definitions but carry player-specific state.

| Runtime Entity | Source | Description |
|---|---|---|
| `QuestInstance` | QuestDefinition + crew | A quest taken on by a crew |
| `QuestCompletion` | QuestInstance + exercises | Result of completing a quest |
| `PlayerRelic` | RelicDefinition + player | Player's owned relic instance |
| `PlayerAchievement` | AchievementDefinition + player | Player's earned achievement |
| `Gift` | ChestDefinition + player | Player's owned chest instance |

### Player Layer (`pkg/game` domain entities)

Owns player-specific state. These entities are per-player and are NOT
content definitions.

| Entity | Table | Description |
|---|---|---|
| `Player` | `odyssey_user_profiles` | Individual family member |
| `Family` | `odyssey_families` | Family group |
| `JourneyProgress` | `odyssey_journey_progress` | Family's progress in a journey |
| `Collection` | `odyssey_collections` | Awarded relic instance |
| `Gift` | `odyssey_gifts` | Awarded chest instance |
| `Achievement` | `odyssey_achievements` | Earned achievement |
| `PlayerStoryFragment` | `odyssey_player_story_fragments` | Player's discovered story fragment |

---

## Content Entity → Table Mapping

| Entity | DB Table | Layer |
|---|---|---|
| RealmDefinition | `odyssey_journey_definitions` | Definition |
| ChapterDefinition | `odyssey_course_definitions` | Definition |
| QuestDefinition | `odyssey_quest_definitions` | Definition |
| CreativePromptDefinition | `odyssey_creative_prompt_definitions` | Definition |
| ChestDefinition | `odyssey_chest_definitions` | Definition |
| DropTableEntry | `odyssey_drop_tables` | Definition |
| RelicDefinition | `odyssey_relic_definitions` | Definition |
| AchievementDefinition | `odyssey_achievement_definitions` | Definition |
| SeasonDefinition | `odyssey_season_definitions` | Definition |
| LoreDefinition | `odyssey_concept_definitions` | Definition |
| LearningConcept | `odyssey_story_fragments` | Definition |
| QuestInstance | `odyssey_missions` | Runtime |
| PlayerRelic | `odyssey_player_collections` | Player |
| Achievement | `odyssey_achievements` | Player |
| Gift | `odyssey_gifts` | Player |
| PlayerStoryFragment | `odyssey_player_story_fragments` | Player |
