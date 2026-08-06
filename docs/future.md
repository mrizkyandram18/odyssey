# Future Expansions

> **Status:** Brainstorm / backlog. None of these are committed to the MVP or
> any specific phase. They exist to capture ideas so they are not forgotten
> and to inform the domain model's extensibility. See [ADR-001](decisions/ADR-001-project-scope.md).

## Social & Group Features

| Idea | Description | Phase |
|---|---|---|
| **Village** | A shared, persistent settlement that the family builds together over time. Furniture, decorations, and NPCs are unlocked through crew milestones. | Phase 4+ |
| **Guild** | Extended family circles or friend groups beyond the core 8. A player can belong to one guild; guilds share a larger world map. | Phase 5+ |
| **Boss** | A family-wide raid encounter. Requires coordination across all members to tackle in sequence. Single-player attempts are possible but harder. | Phase 4+ |
| **Dungeon** | Procedurally-generated short adventures. Each run is different; the family chooses entry modifiers. | Phase 3+ |
| **Trading** | Players trade Relics with each other. Requires mutual agreement; no auction house. Peer-to-peer only. | Phase 2 |
| **Mentorship** | Older family members can leave "hints" or "gifts" for younger ones at specific creative spaces. | Phase 2 |

## Real-Time & Multiplayer

| Idea | Description | Phase |
|---|---|---|
| **Real-time Sync** | Live updates via WebSocket/Supabase Realtime so the family sees contributions appear immediately. | Phase 2+ (PWA capability check) |
| **Shared Screen** | A TV-friendly mode for family gatherings where everyone sees the world on a big screen while contributing from phones. | Phase 4 |
| **Voice Chat** | In-quest voice chat for relay quests. Optional and always opt-in. | Phase 3 |

## AI Integration

| Idea | Description | Phase |
|---|---|---|
| **AI Dungeon Master** | Generative narrative engine that adapts to family choices. | Phase 5 |
| **AI Quest Generator** | Procedurally generates daily turns tailored to interests. | Phase 4 |
| **AI Comic** | Converts story submissions into illustrated comics. | Phase 4 |
| **AI Quiz** | Generates educational puzzle challenges on demand. | Phase 3 |
| **AI Judge** | Provides constructive feedback on creative submissions. | Phase 3 |

See [docs/ai.md](ai.md) for the full AI readiness rationale.

## Content Depth

| Idea | Description | Phase |
|---|---|---|
| **Season** | Time-bounded progression arcs with a thematic story (e.g., "Winter Festival"). Resets with cosmetic rewards. | Phase 3 |
| **Comic Mode** | A full comic-reader view where the family's story submissions are assembled into illustrated chapters. | Phase 4 |
| **Photo Album** | A gallery view of all `PHOTO`-kind submissions, organized by realm and date. | Phase 3 |
| **Video Reel** | A gallery of `VIDEO`-kind submissions, compiled into a family story reel. | Phase 4 |
| **Mini-Games** | Simple, single-tap mini-games (e.g., a reaction test, a quick puzzle) embedded in quests. | Phase 2 |
| **Music / Soundtrack** | Procedural or community-composed background music that evolves with the story. | Phase 5 |

## Mechanics Extensions

| Idea | Description | Phase |
|---|---|---|
| **Inventory** | A catalog of collected items (Relics, tools, cosmetics). Phase 2 adds trading. | Phase 2 |
| **Companion** | A persistent virtual companion that follows the party and evolves based on family choices. | Phase 4 |
| **Time Capsules** | Entries the family writes that unlock for future family members on a future date. | Phase 5 |
| **Branching Campaign** | A season-long story with meaningful branches where the family's cumulative choices reshape the world. | Phase 5 |
| **Achievement Showcase** | A dedicated screen where earned Relics and Achievements are displayed in an interactive gallery. | Phase 3 |
| **Custom Quests** | Family members author and share custom quests with each other (peer-to-peer). | Phase 3 |
