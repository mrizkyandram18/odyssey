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
| Backend  | Go (standard library only)          |
| Database | Supabase (PostgreSQL) — `odyssey_*` tables |
| Auth     | Gatekeeper (BOTH login mode) via adapter |

## Getting Started

1. Copy `.env.example` to `.env` and fill in the values.
2. **Backend:** `go run api/dev/main.go`
3. **Frontend:** `cd web && npm install && npm run dev`

## Documentation

Start at `CLAUDE.md`, then read `docs/vision.md` → `docs/principles.md` →
`docs/domain-model.md` → `docs/architecture.md` → `docs/decisions/`.

## Phases

| Phase | Scope | Status |
|-------|-------|--------|
| 0 — Foundation | Vision, ADRs, architecture, coding standards | Complete |
| 1 — MVP | One realm, 3 quests, daily turns, basic progression | Next |
| 2 — Social Fabric | Peer reactions, coin currency, expanded creative tools | Future |
| 3 — Deeper Storytelling | Branching quests, second realm, achievement log | Future |
| 4 — Richer Creation | Comic mode, expanded creative tools, trading | Future |
| 5 — Long Arc & Integration | Seasonal realms, Family Reward signal integration, AI | Future |
