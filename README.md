# Odyssey

Cooperative adventure platform for private family groups.

## Overview

Odyssey is a lightweight, persistent, cooperative adventure game designed for a
small, closed family circle. Every mechanic is built around bringing families
together through shared quests, collaborative challenges, and collective
storytelling — purely for fun, learning, and engagement.

**No real money. No gambling. No microtransactions.** Resources inside Odyssey
(XP, Relics, Chests, Story Fragments, Inspiration) are entirely fictional.

## Tech Stack

| Layer    | Technology                          |
|----------|-------------------------------------|
| Frontend | React 19, Vite, TypeScript, Tailwind CSS |
| Backend  | Go 1.25 (mostly standard library)   |
| Database | Supabase (PostgreSQL) — `odyssey_*` tables |
| Auth     | Local Authentication (prototype)    |

## Project Structure

- `api/` - HTTP handlers and entry points
- `cmd/` - CLI tools and utilities
- `docs/` - Project documentation
- `pkg/` - Core domain logic and integrations
- `scripts/` - Local dev/deploy scripts and database migrations (`scripts/migrations/`)
- `supabase/` - Supabase project configuration (`config.toml`)
- `web/` - React frontend application

## Getting Started

1. Copy `.env.example` to `.env` and fill in the values.
2. **Backend:** `cd api/dev && go run main.go`
3. **Frontend:** `cd web && npm install && npm run dev`

### Prototype Login

The current prototype uses Local Authentication for demonstration purposes. Use any of the following accounts:

- **demo1** / `odyssey123` (Leo - Seeker)
- **demo2** / `odyssey123` (Maya - Guide)
- **demo3** / `odyssey123` (Sam - Builder)

## CI Status

GitHub Actions CI is currently green (Docker build & tests).

## Documentation

Start at `CLAUDE.md`, then read `docs/vision.md` → `docs/principles.md` →
`docs/domain-model.md` → `docs/architecture.md` → `docs/decisions/`.

## Phases

| Phase | Scope | Status |
|-------|-------|--------|
| 0 — Foundation | Vision, ADRs, architecture, coding standards | Complete |
| 1 — MVP | One realm, 6 quests, daily turns, basic progression | Release Candidate (Conditional GO) |
| 2 — Social Fabric | Peer reactions, coin currency, expanded creative tools | Future |
| 3 — Deeper Storytelling | Branching quests, second realm, achievement log | Future |
| 4 — Richer Creation | Comic mode, expanded creative tools, trading | Future |
| 5 — Long Arc & Integration | Seasonal realms, Family Reward signal integration, AI | Future |
