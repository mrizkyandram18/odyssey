# Gameplay

## Core Concept

Odyssey presents a shared, illustrated world that the family explores together.
Each session is a small turn: family members contribute actions, the group
advances the story, and everyone sees the result. The world is persistent — it
remembers where the family left off and what they've accomplished.

## The Shared World

The world is divided into **realms** — themed areas such as "The Whispering
Woods," "The Clockwork City," or "The Starlit Library." Journeys are unlocked
sequentially through story progression; the family starts in a single tutorial
journey and expands outward.

Each journey contains:

- **Story scenes** — illustrated narrative moments that respond to the
  family's choices.
- **Missions** — the primary cooperative activity.
- **A Creative Space** — a sandbox where the family can leave marks on the
  world (plant a tree, paint a mural, add a story paragraph).

## Missions

A **quest** is the primary unit of play. It weaves a short story and offers
1–4 **exercises** that must be completed to finish the quest.

### Mission Types

| Type | Description | Example |
|---|---|---|
| **Solo** | One family member takes the lead; others can assist. | "Decode the riddle of the old lighthouse." |
| **Relay** | Each family member completes one leg in sequence. | Member A observes, Member B records, Member C combines. |
| **Group** | The whole family collaborates on shared exercises. | "Assemble the constellation map together." |
| **Creative** | A creative-space quest where the family builds, draws, or tells stories together. | "Design a heraldry for your crew" or "Write a course of your family legend." |

### Creative Mission Submissions

Creative Missions accept four first-class submission types. These are the
pebble in the pond where AI amplification (see [docs/ai.md](ai.md)) can enhance
family creativity without replacing it:

| Kind | Description |
|---|---|
| **Story** | A written narrative snippet. |
| **Comic** | A multi-panel illustrated submission. |
| **Photo** | A real-world photograph tied to a challenge. |
| **Video** | A short (≤ 30 s) video contribution. |

Submissions are attributed to the contributing Explorer and visible to all
family members in the quest log and creative-space gallery.

### Exercise Examples

- **Observation:** "Find something red in your house and describe it."
- **Research:** "Look up why leaves change color and share one fact."
- **Movement:** "Take a photo of your shadow at a specific angle."
- **Creative:** "Draw a quick sketch of the creature you imagine."
- **Puzzle:** "Solve this logic riddle together."
- **Story:** "Each person adds one sentence to our crew's legend."

### Completion

A quest is complete when all its exercises are resolved. The family receives
a quest reward (XP, a Gift, a story unlock). The outcome may depend on *how*
exercises were completed — different choices open different branches, but all
valid choices advance the story.

## Roles

Each family member has an **Explorer** who takes on a **role**. Roles are
lightweight narrative framing — they are **not** "classes" with stat bonuses,
resource multipliers, or mechanical advantages. They help the story comment on
each member's contributions.

| Role | Flavor |
|---|---|
| **Seeker** | Finds things, observes details, does research. |
| **Builder** | Builds, draws, creates, composes. |
| **Guide** | Leads, makes decisions, narrates. |

Roles are chosen or rotate freely. A family member might be the Seeker for one
quest and the Guide for the next. Mastery (how often you've played each role)
is a quiet metric that unlocks narrative flavor, not mechanical advantage.

## Daily Turn

Each day, the family receives one or more **daily turns** — small, optional
invitations to engage:

- A micro-quest (2–3 minutes).
- A creative-space prompt.
- A story continuation choice.

Daily turns are designed to be completable in short bursts but contribute to
larger missions. Streaks and gentle reminders encourage consistency, never
pressure.

## Journey Progress (Community Progress)

Each journey has a shared **Journey Progress** bar. It fills as the family
completes missions and exercises within that journey. Milestones along the way
unlock new missions in the journey, a shared Journey Gift, and narrative concept.
When the bar is full, the next journey becomes available.

## Social Fabric

- **Shared view:** Every family member sees the same world state.
- **Attribution:** Each action is attributed to an Explorer, building a
  visible history of who contributed what.
- **Comments & reactions:** Family members can react to each other's
  contributions (stickers, quick replies) within creative spaces and quest
  logs.
- **Handoffs:** Relay missions surface a clear "your turn" notification to the
  next family member.

## Offline Considerations

Observation, creative, and puzzle exercises are designed so a family member
can think through an answer and submit it when convenient. A fuller
offline-first experience (local state persistence, sync-on-reconnect) is
planned for a future phase — see [docs/future.md](future.md).
