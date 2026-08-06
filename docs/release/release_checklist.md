# Operational Release Checklist

Use this checklist to perform pre-release validation before tagging any Phase 1 (or future) release.

---

## 📁 Repository Baseline
- [ ] Working tree is 100% clean (`git status`).
- [ ] No uncommitted changes or untracked temporary debug files.
- [ ] Commit history is clean with clear commit messages.
- [ ] Scope Freeze is active (no un-audited Phase 2 features included).

---

## ⚙️ Backend Quality Gate
- [ ] `go fmt ./...` passes with zero modified files.
- [ ] `go vet ./...` passes with 0 warnings/errors.
- [ ] `go test ./...` passes 100% of test packages.
- [ ] Race detector executed (`go test -race ./...`) OR CGO/OS limitation explicitly documented in Release Report.

---

## 🎨 Frontend Quality Gate
- [ ] `npx tsc --noEmit` passes with 0 TypeScript compilation errors.
- [ ] `npm run test` (Vitest) passes 100% of unit tests.
- [ ] `npm run build` succeeds, generating production bundle in `dist/`.

---

## 📱 PWA Verification
- [ ] App manifest (`manifest.webmanifest`) validates without errors in DevTools.
- [ ] Service Worker registers cleanly on initial page load.
- [ ] App shell renders correctly when offline (DevTools Network -> Offline).
- [ ] Service worker updates and cache invalidation prompt user reload when new bundle is deployed.

---

## 🎮 Golden Path (Live E2E Verification)
Perform manual validation or run automated E2E against a running server (`go run api/dev/main.go`):
- [ ] **Step 1: Login** — Enter Gatekeeper credentials, verify session token issued.
- [ ] **Step 2: Home Mount** — Verify Explorer level, streak badge, and realm progress load.
- [ ] **Step 3: Daily Turn** — Consume daily turn, verify XP increase and streak increment.
- [ ] **Step 4: Start Quest** — Select pending quest, click "Start Quest", verify status changes to ACTIVE.
- [ ] **Step 5: Challenge List** — Inspect active challenges inside Quest Detail view.
- [ ] **Step 6: Complete Challenge** — Click "Complete Challenge", verify XP award modal/banner and quest completion status.
- [ ] **Step 7: Progression** — Verify Explorer level up trigger when XP threshold is crossed.
- [ ] **Step 8: Realm Progress** — Verify crew realm progress bar advances after quest completion.
- [ ] **Step 9: Journal** — Open Family Journal, verify new Milestone & Lore entries are unlocked.
- [ ] **Step 10: Chest & Relic** — Open earned chest from Home, verify relic awarded to inventory.
- [ ] **Step 11: Logout** — Click "Sign Out" in Profile, verify redirect to login and cookie cleanup.
- [ ] **Step 12: Persistence** — Sign in again, verify explorer XP, level, and realm progress persist intact.

---

## 🏷️ Release Sign-Off
- [ ] All checklist items above are checked **PASS**.
- [ ] `release_report.md` updated to **GO**.
- [ ] Create Git release tag: `git tag -a v0.1.0-mvp -m "Phase 1 MVP Production Release"`.
- [ ] Push tag: `git push origin v0.1.0-mvp`.
