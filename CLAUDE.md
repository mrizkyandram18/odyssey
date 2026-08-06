# Odyssey

Cooperative adventure platform for private family groups.

## Tech Stack

- Frontend: React 19, Vite, TypeScript, Tailwind CSS
- Backend: Go 1.25 (mostly standard library)
- Database: Supabase (PostgreSQL) — shared project, `odyssey_*` tables only
- Auth: Gatekeeper (external, immutable) via Authentication Provider adapter

## Documentation

| File | Purpose |
|---|---|
| `CLAUDE.md` | This file. Project overview for the assistant. |
| `docs/vision.md` | Product vision, audience, success criteria. |
| `docs/principles.md` | 11 design principles (priority-ordered). |
| `docs/non-goals.md` | Explicit boundaries — what Odyssey will never become. |
| `docs/domain-model.md` | Core domain entities, layers, and table mapping. |
| `docs/architecture.md` | High-level architecture, module map, deployment. |
| `docs/gameplay.md` | Core gameplay, quests, roles, daily turn, creative quests. |
| `docs/game-loop.md` | Moment-to-moment, daily, quest, progression loops. |
| `docs/progression.md` | Explorer Level, crew level, Realm Progress, milestones. |
| `docs/economy.md` | XP / Relics / Chests / Inspiration design. |
| `docs/integrations.md` | Gatekeeper (adapter), Family Reward (deferred), Supabase. |
| `docs/roadmap.md` | Phase 0–5 roadmap, MVP scope. |
| `docs/glossary.md` | Domain terminology. |
| `docs/coding-standards.md` | Go + TypeScript conventions. |
| `docs/ui-guidelines.md` | Mobile-first, PWA, component, motion, voice. |
| `docs/ai.md` | Future AI integration (story, quest, comic, quiz). |
| `docs/future.md` | Future expansion ideas backlog. |
| `docs/decisions/` | Architecture Decision Records (ADRs). |

**Start here:** `vision.md` → `principles.md` → `non-goals.md` →
`domain-model.md` → `architecture.md` →
`decisions/ADR-001` through `ADR-004`.

## Key Constraints

- ONLY new tables prefixed `odyssey_` in shared Supabase (never modify existing business tables).
- Real rewards via Family Reward integration only — deferred to Phase 5 (approved by ADR-004).
- NO real money inside Odyssey — XP, Relics, Chests, and Inspiration are fictional only.
- MVP resources: XP, Relics, Chests, Story Fragments, Inspiration. No Adventure Tokens, no Coins (Coin currency is Phase 2).
- Mobile-first, PWA-first (offline-first deferred to a future phase).
- Gatekeeper BOTH login mode (device trust + credential) is the ONLY authentication path.
  Verified through an Authentication Provider adapter — the domain never touches Firestore directly.
- Gatekeeper and Family Reward must NEVER be modified.

## Reference Systems (READ ONLY — do not modify)

| System | Path | Notes |
|---|---|---|
| Family Reward | `D:\Personal\Projects\kuota - Copy` | Auth flow, API style, DB conventions. |
| Gatekeeper | Android app (Firestore-backed) | Device trust, `users/{PARENT_ID}/children/{uid}`. |

## Development & Project Structure

### Build & Run
- **Backend:** `go run api/dev/main.go`
- **Frontend:** `cd web && npm run dev`
- **Build (CI):** `go build ./...` for backend, `npm run build` for frontend.

### Testing
- **Backend Tests:** `go test -v -race -count=1 ./...`
- **Frontend Tests:** `npm run test` (Vitest)
- **Linting:** `go vet ./...` and `go fmt ./...` for Go, `npm run lint` for frontend.

### Directories
- `api/` - HTTP entry points
- `cmd/` - CLI utilities
- `docs/` - Architecture & ADRs
- `pkg/` - Core domain, database, auth
- `web/` - React frontend
