# Architecture

## Goals

- **Modular monolith backend** (Go, mostly standard library) that can deploy as
  serverless functions or a single service.
- **PWA frontend** (React 19, Vite, TypeScript, Tailwind CSS) that works on
  phones first and is installable to the home screen.
- **Clear module boundaries** so each feature owns exactly one concern.
- **Port-and-adapter auth.** The domain defines an `Authenticator` interface;
  an adapter implements it by reading Gatekeeper's Firestore documents. The
  domain never references Firestore directly.
- **Stable, narrow integrations** with Gatekeeper and Family Reward.

## High-Level View

```
┌──────────────────┐        ┌────────────────────────┐        ┌──────────────────┐
│  Family (Browser) │◀──────│  Odyssey Frontend       │◀──────│  Odyssey Backend  │
│  PWA (React 19)   │  HTTP  │  (Vite, TS, Tailwind)   │  HTTP  │  (Go, stdlib)     │
└──────────────────┘        └────────────────────────┘        └────────┬────────┘
                                                                      │
                              ┌────────────────────────────────────────┼─────────────────────────────────┐
                              ▼                                        ▼                                  ▼
                    ┌──────────────────┐                    ┌──────────────────┐            ┌──────────────────────┐
                    │  Supabase        │                    │  Gatekeeper      │            │  Family Reward       │
                    │  (PostgreSQL)    │                    │  (Firestore +     │            │  (Future signal     │
                    │  odyssey_* tables│                    │   Android app)   │            │   consumer)          │
                    └──────────────────┘                    └──────────────────┘            └──────────────────────┘
```

### Interaction Flow

1. **Login.** The PWA sends credentials + device payload to the backend. The
   backend's `Authenticator` adapter reads Gatekeeper's Firestore device
   document (read-only), verifies compliance, checks the credential, and
   returns an Odyssey session token.
2. **Play.** The PWA calls the Odyssey backend for game state reads/writes.
   The backend persists to `odyssey_*` tables in Supabase.
3. **Sync.** The PWA fetches fresh world state on foreground and after network
   reconnect. No local write queue in the MVP — contributions are sent
   immediately on submission.
4. **(Future) Reward signals.** In Phase 5, the backend may emit achievement
   signals for Family Reward consumption. See [ADR-004](decisions/ADR-004-reward-integration.md).

## Component Map

```
odyssey/
├── api/                          # Go backend handlers (serverless-compatible)
│   └── <resource>/index.go       # One Handler(w, r) per resource
├── internal/
│   └── api/
│       └── story_fragments/      # Story fragment discovery/replay handlers
├── pkg/                          # Shared, reusable Go packages
│   ├── auth/                     # Authenticator port + Firestore adapter
│   ├── content/                  # ContentService (DB-backed definitions & caching)
│   ├── db/                       # Supabase REST client & store implementations
│   ├── game/                     # Game logic (quests, progression, economy, chapters, lore, achievements)
│   │   ├── quest/                # Quest & challenge domain
│   │   ├── progression/          # XP, levels, milestones, relics
│   │   ├── creative/             # Creative-space, story submissions
│   │   ├── fragment/             # Story fragment discovery, replay & gallery domain
│   │   └── world/                # Realm progress, world state
│   ├── observability/            # Structured logging, metrics, health, profiler
│   └── shared/                   # Config, security middleware, CORS, rate limiting
├── web/                          # React 19 PWA frontend
│   ├── src/
│   │   ├── app/                  # App shell, routing, session provider
│   │   ├── features/<feature>/   # Feature modules (co-located components)
│   │   ├── shared/               # UI primitives, hooks, utils
│   │   └── assets/               # Illustrations, icons
│   └── vite.config.ts
├── docs/                         # This documentation
│   └── decisions/                # ADRs
├── scripts/                      # Local dev / deploy scripts
└── CLAUDE.md
```

## Module Boundaries

### `pkg/auth` — Authentication Provider (Port + Adapter)

**Port (domain-facing):** an `Authenticator` interface with a single method:

```go
type Authenticator interface {
    // Verify checks device trust + credential. Returns uid, device bound, and error.
    Verify(ctx context.Context, uid, credential string, device DevicePayload) (bool, error)
}
```

The domain layer (`pkg/game`) depends only on this interface. It has no
knowledge of Firestore, Gatekeeper, or Firebase.

**Adapter (implementation):** `pkg/auth/firestore.go` implements
`Authenticator` by reading the Gatekeeper device document from Firestore
(read-only). This is the single place that touches Firestore.

`pkg/auth` also owns:

- `session.go` — HMAC-signed Odyssey session tokens (user kind, ~8 h TTL).
- Credential verification against `odyssey_user_profiles`.
- Device binding and validation logic.

**Does not own:** password hashing beyond a standard lib (bcrypt/argon2). Does
not know about quests or progression.

### `pkg/db`

Owns:

- A thin Supabase REST wrapper (mirrors Family Reward's `supabase_get` /
  `supabase_mutate` pattern but under the `odyssey_` namespace).
- Typed result structs for each `odyssey_*` table.

**Does not own:** game logic. Only transport and serialization.

### `pkg/game` (and sub-packages)

Owns:

- Quest, challenge, world-state, and progression logic.
- Story fragment discovery, catalog listing, and realm replay logic (`pkg/game/fragment`).
- Reward computation (XP, Relic, Chest grants).
- All game-domain types and rules.

**Does not reach into** `auth` or `db` directly. It receives a `Store`
interface (dependency injection) and pure data. This makes game logic unit-
testable without Supabase or Gatekeeper.

**Does not own:** HTTP, networking, or persistence mechanics.

#### Quest Pacing & Reward Idempotency (Phase 3)
- **Quest Start Pacing:** The system enforces a daily limit on starting new quests (`max_new_quests_per_day = 1`). This is validated using the `started_at` and `started_by` tracking fields on `QuestInstance`.
- **Idempotent Completion:** Quest completion uses a Compare-And-Swap (CAS) mechanism (`UpdateUserIfMatch` for XP and Coin mutations). This ensures that concurrent or retry requests do not duplicate rewards.
- **Explicit Reward Mapping:** Chests are deterministic. They map explicitly from a `quest_slug` → `ChestDefinition.reward_relic` → `Relic`. No RNG or loot-box mechanics are used.

### `pkg/content` — Content Engine

Owns:

- **ContentService** — the central service for all game content definitions.
  Loads definitions from the database, caches them in memory, and provides
  lookup methods by slug, realm, chapter, and season.
- **In-memory cache** — a lightweight RWMutex-based cache with configurable
  TTL. Supports manual refresh via the admin API.
- **Fallback chain** — when the database is empty, the ContentService falls
  back to hardcoded defaults embedded in the existing catalog packages
  (`pkg/game/chest`, `pkg/game/quest`, `pkg/game/relic`, `pkg/game/world`, `pkg/game/fragment`).

The ContentService is consumed by gameplay services (QuestService,
RewardEngine, RelicService, CreativeService, FragmentService) instead of hardcoded catalogs
wherever possible. The fallback chain ensures the application never fails
because content is missing.

### `pkg/db`

Owns:

- A thin Supabase REST wrapper (mirrors Family Reward's `supabase_get` /
  `supabase_mutate` pattern but under the `odyssey_` namespace).
- Typed result structs for each `odyssey_*` table.
- **Content store implementations** — Supabase-backed implementations of
  the content store interfaces defined in `pkg/game/content`.

**Does not own:** game logic. Only transport and serialization.

### `api/<resource>/index.go`

Each resource has exactly one handler file with a single `Handler` function.
Handlers are thin: parse request → call `pkg/game` or `pkg/content` or
`pkg/auth` → write JSON response. Business logic lives in `pkg/`, never in
`api/`.

### `internal/api/story_fragments/index.go` — Story Fragments API

Provides endpoints for collectible story fragments and realm replays:

- `GET /api/story_fragments` — List all story fragments with caller's discovery status
- `POST /api/story_fragments/discover` — Discover/collect a fragment by slug (+20 XP)
- `POST /api/story_fragments/replay` — Replay a completed realm for bonus dialogue & secret fragments

#### Story Fragments Flow (`list` → `discover` → `replay`)
1. **List (`GET /api/story_fragments`):** Fetches the full story fragment catalog merged with the player's discovery history (`discovered_at`).
2. **Discover (`POST /api/story_fragments/discover`):** Player collects a fragment slug. Grants +20 XP if newly discovered.
3. **Replay (`POST /api/story_fragments/replay`):** Verifies that the crew has completed the specified realm (`status = COMPLETE` or `progress >= 100`). Reveals secret hidden fragments (`is_hidden = true`) and returns narrative bonus dialogue.

#### Authentication & Idempotency
- **Auth & Authorization:** All endpoints require a valid session HMAC token validated via `auth.ClaimsFromRequest`. Queries and mutations are strictly scoped by `claims.UID` and `claims.CrewID`.
- **Idempotency:** Fragment discovery awards +20 XP only on the initial discovery (`newlyDiscovered == true`). Subsequent calls for an already-discovered fragment return `newlyDiscovered = false` and `XPGranted = 0`, ensuring exactly-once XP grants.

### `api/admin/` — Admin CMS API

Authenticated admin-only read API for content management. Provides:

- `GET /api/admin/content/status` — content load status (counts per type)
- `POST /api/admin/content/reload` — manual cache refresh
- `GET /api/admin/content/realms` — list all realm definitions
- `GET /api/admin/content/quests` — list all quest definitions
- `GET /api/admin/content/relics` — list all relic definitions
- `GET /api/admin/content/chests` — list all chest definitions
- `GET /api/admin/content/prompts` — list all creative prompt definitions

Only READ operations in the MVP. No editing UI yet.

## Frontend Architecture

```
src/
├── app/                  # Shell: routing, session provider, PWA bootstrap
├── features/             # Feature modules (each self-contained)
│   ├── login/            # Login flow (Gatekeeper BOTH integration)
│   ├── home/             # Crew dashboard
│   ├── quest/            # Quest view, challenge submission
│   ├── creative/         # Creative-space canvas
│   ├── journal/          # Achievements, story log
│   └── profile/          # Explorer, level, role
├── shared/               # Reusable UI + logic
│   ├── components/       # atoms, molecules, organisms (design system)
│   ├── hooks/            # Data fetching + state hooks
│   ├── lib/              # API client, session storage
│   └── types/            # Shared TS types
└── assets/
```

- **Routing:** Client-side hash routing (PWA-installable, no server route
  configuration required beyond SPA fallback).
- **State:** Local component state for ephemeral UI; a thin data layer
  (SWR/informal cache) for server data.
- **Session:** Stored in `localStorage`; sent as `Authorization: Bearer`
  and `X-User-Session` (mirroring Family Reward's dual-header convention for
  consistency).
- **Network:** Contributions are sent immediately on submission. The UI shows
  an optimistic result and rolls back on error. Full offline-first
  persistence is deferred to a future phase (see [docs/future.md](future.md)).

## Backend Architecture

- **Language:** Go (matching Family Reward's Go 1.25 target).
- **Framework:** Go standard library `net/http` only (no Echo/Fiber/Gin).
  Each `api/<resource>/index.go` exposes `Handler(w http.ResponseWriter, r *http.Request)`.
- **Routing:** Resource-per-folder convention; SPA fallback for static assets.
  On serverless (Vercel), `vercel.json` rewrites (if adopted) map
  `/api/<resource>` → `/api/<resource>/index`.
- **Sessions:** HMAC-signed tokens (base64 payload + base64 HMAC-SHA256
  signature), matching Family Reward's session approach. Not JWT. This keeps
  the auth model familiar to anyone who worked on Family Reward.
- **Auth adapter:** `pkg/auth/firestore.go` reads Gatekeeper Firestore data
  via the Firebase Admin SDK (using `google.golang.org/api` +
  `cloud.google.com/go/firestore`), same dependency set as Family Reward.
  This is the **only** Firestore access in the backend.
- **Supabase:** Direct REST (PostgREST) calls with the service-role key,
  same pattern as Family Reward.

## Environment & Configuration

| Variable | Purpose |
|---|---|
| `SUPABASE_URL`, `SUPABASE_SERVICE_KEY` | Supabase REST access |
| `FIREBASE_CREDENTIALS` / `GOOGLE_APPLICATION_CREDENTIALS_JSON` | Firestore (Gatekeeper adapter, read-only) |
| `PARENT_ID` | Gatekeeper parent UID |
| `SESSION_SIGNING_SECRET` | HMAC key for Odyssey sessions |
| `GATEKEEPER_MIN_BUILD_NUMBER` | Min Gatekeeper build (default "49") |

## Deployment Model

TBD in implementation. Options (evaluated later, documented as a decision):

- **Serverless functions** (Vercel) — matches Family Reward. Good for low
  traffic, zero-ops.
- **Single Go service** (container) — simpler local dev, full control.

Either way, the backend code is structured so handlers are pure functions
over `http.ResponseWriter`/`*http.Request`. The deployment wrapper is an
adapter, not embedded in business logic.

## Non-Functional Requirements

| Concern | Target |
|---|---|
| Time to first playable quest | < 3 seconds on 3G |
| Session lifetime | 8 hours (user), 30 min (setup) |
| PWA installable | Yes (manifest + service worker) |
| Mobile screen support | iPhone SE (375px) and up |
| API response time | < 500 ms p95 |
| Offline | Planned for a future phase (not MVP) |
