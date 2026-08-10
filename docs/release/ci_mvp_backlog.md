# CI MVP Backlog — Deferred Checks

**Date:** 2026-08-10  
**Mode:** MVP speed — delivery over release-hardening  
**Status:** Deferred ≠ done. Re-enable when infra is stable.

## Blocking CI (retained)

Workflow: `.github/workflows/ci.yml` (on push/PR)

| Job | Purpose |
| --- | --- |
| Lint | `go fmt` / `go vet`, `tsc --noEmit`, ESLint |
| Unit Tests | `go test ./...`, Vitest |
| Build Artifacts | `go build` backend, `npm run build` frontend |

## Deferred (manual only)

Workflow: `.github/workflows/ci-heavy.yml` (`workflow_dispatch`)

| Job | Why deferred | Re-enable criteria |
| --- | --- | --- |
| Race Detector | Slow; not required for MVP gameplay loop | Green on CI Linux with CGO; no new data races |
| Integration (Playwright) | Failing on Postgres role/`root` + incomplete backend seed in job | Fix service env (user/password/URL), real backend start + migrations, stable E2E |
| Docker Build | Infra-heavy; not needed for local/Vercel MVP path | Image builds cleanly; optional deploy uses it |
| Check Release | Echo-only gate; no real release automation | Replace with real tag/`v0.1.0` process after MVP content complete |

## Not deferred as "complete"

These remain **open backlog**, not finished work:

1. Fix Playwright integration job (Postgres role, backend process, migrations, seed).
2. Re-attach race detector to default PR path once green.
3. Re-attach Docker build if container deploy becomes required.
4. Formal release tag gate (`v0.1.0`) after 6 playable quests verified.
5. Production E2E / deploy verification (manual for now).

## What was not deleted

- All Playwright specs under `web/tests/e2e/`
- Unit tests, Dockerfile, prior release docs
- Heavy job definitions (moved to manual workflow)

## Branch protection

Repo `main` had **no** branch protection / required checks at time of change.  
If required checks are added later, require only: **Lint**, **Unit Tests**, **Build Artifacts**.
