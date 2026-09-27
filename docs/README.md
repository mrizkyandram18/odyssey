# Odyssey Documentation

Handover documentation for a developer who has never seen Odyssey before.
It describes the **actual current implementation** (Go stdlib monolith +
React PWA + Supabase Postgres), not historical designs.

## What Odyssey is

A private family daily-task → coin → redemption platform. Admins publish a
sequenced task list per day; members complete tasks in the PWA; submissions
are graded automatically in Postgres or queued for admin review; approved
work earns coins + XP + streaks + levels; coins are redeemed from a reward
catalog via admin-approved claims (money transfer is manual); a parallel
loop grants daily tickets redeemable for avatar cosmetics.

## Map

| File | Covers | Read when you need to… |
|---|---|---|
| `architecture.md` | Purpose, actors, context, building blocks, runtime flows, rule ownership, security, deployment, risks, glossary | …understand how the system fits together and where any rule lives |
| `database.md` | Tables, relationships, state machines, RPCs, ledger model, invariants, migration conventions | …change schema, write an RPC call, or debug data |
| `api.md` | HTTP surface grouped by domain, with auth, behavior, and error notes | …implement or debug a frontend feature |
| `operations.md` | Prerequisites, env setup, local dev, tests, DB workflow, deploy, CI, health, troubleshooting, manual payout | …run, configure, deploy, or fix the system |
| `decisions.md` | Architecture-significant decisions as Context / Decision / Consequences / Evidence | …understand why the system is shaped this way |

## Recommended reading order

```text
1. README.md          (root — what it is, quick start)
2. CLAUDE.md          (root — stack, enforced rules, commands)
3. docs/README.md     (this index)
4. docs/architecture.md
5. docs/database.md
6. docs/api.md
7. docs/operations.md
8. docs/decisions.md  (on demand)
```

New backend developer: 1 → 4 → 5 → 6 (tasks + shop sections).
New frontend developer: 1 → 4 (§4–§6) → 6 → `web/src/shared/lib/api.ts`.
On-call / ops: `operations.md` troubleshooting + `architecture.md` §9.

## Source-of-truth hierarchy

```text
1. Actual source code (pkg/, internal/api/, web/src/)
2. Database migrations / RPC / constraints (supabase/migrations/)
3. Tests (*_test.go, web/src/**/*.test.*, scripts/verify-*.cjs)
4. Current configuration (.env.example, production.env.example)
5. CI/CD and deployment (Dockerfile, docker-compose.yml, vercel.json, .github/)
6. These docs
7. README / historical documentation
```

If documentation conflicts with code, **code wins** — fix the docs.
Claims that could not be confirmed read-only are marked `UNVERIFIED` /
`UNDEFINED` together with the evidence still needed, instead of guessing.

## Terminology conventions

Canonical terms are defined in `architecture.md` §12 (Glossary) and used
consistently across all files:

- **Member / Admin** — roles (`ADMIN` vs `MEMBER`; legacy `GUIDE`/`BUILDER`
  normalize to admin).
- **Task** — a daily definition in `odyssey_tasks` (`AUTO` vs
  `ADMIN_REVIEW` evaluation).
- **Submission** — one attempt (`PENDING → APPROVED | REJECTED`).
- **Ledger** — `odyssey_coin_transactions`, append-only.
- **Claim** — a redemption request (`PENDING → APPROVED | REJECTED`).
- **Ticket / cosmetic** — daily reward ticket (`GRANTED → SPENT`) and avatar
  items (frame/effect, tier 1–3).
- **Earning cap** — monthly coin ceiling; `0` = unlimited.

Obsolete adventure-era words (adventure, RPG, journey, mission, quest,
gatekeeper, family reward, old economy/auth models) must not be used for
current behavior. The root `README.md` previously described a cooperative
adventure game — that description is superseded and has been rewritten.

## How to update these docs

- Docs-only changes must not touch code, migrations, or config behavior.
- Verify every claim against code/schema/tests before writing it; cite the
  source file where the behavior is non-obvious.
- Keep the surface lean: 6 files. Add a file only for a clear
  onboarding/maintenance need — never to satisfy a template checklist.
- When architecture changes: update `architecture.md` (flows/ownership) +
  whichever of `database.md` / `api.md` / `operations.md` is affected, and
  add a `decisions.md` entry only if the change is architecture-significant.
- No commit/push/PR unless explicitly requested.
