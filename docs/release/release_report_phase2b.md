# Release Report — Phase 2B: Drawing Canvas

**Date:** 2026-08-09
**Milestone:** Phase 2B — Creative Quest Drawing Canvas
**Basis Commit (main):** `997a10c` (Merge `feature/phase-2b-creative-canvas`)
**Process Standard:** Release Engineering Pipeline (Evidence First Principle)

---

## 📊 Executive Summary

Phase 2B menambahkan **mode menggambar (drawing canvas)** pada Creative
Quests: Explorer dapat memilih antara menulis Story atau menggambar langsung di
canvas, lalu mengirimnya sebagai submission `DRAWING` yang divalidasi aman di
backend sebelum disimpan dan dirender ulang sebagai gambar.

Semua gate CI lulus hijau pada commit merge ke `main`, termasuk Playwright
Integration (E2E), Race Detector, Build Artifacts, dan Docker Build.

**Rekomendasi:** **GO** untuk konten Phase 2B (semua gate CI PASS).

---

## 🚦 CI Status — Merge ke main (`997a10c`)

| Job | Status | Catatan |
| :--- | :---: | :--- |
| **Lint** | **PASS** | `go vet` + TypeScript type check + ESLint. |
| **Unit Tests** | **PASS** | Go unit tests + Vitest (24 tests, termasuk `svg.test.ts` & `svg_test.go`). |
| **Race Detector** | **PASS** | `go test -race` hijau di CI. |
| **Integration (Playwright)** | **PASS** | Golden path E2E terhadap backend + DB sungguhan. |
| **Build Artifacts** | **PASS** | Backend binary + frontend `dist/` ter-build. |
| **Docker Build** | **PASS** | Image container berhasil di-build (no push). |
| **Check Release** | **PASS** | Release tag belum dibutuhkan (menunggu verifikasi manual). |

---

## 🛠️ Scope & Changes

### Backend (Go)

| File | Perubahan |
| :--- | :--- |
| `pkg/game/domain.go` | Konstant `SubmissionDrawing = "DRAWING"` ditambahkan. |
| `pkg/game/creative/svg.go` | **Baru** — `ValidateSVG`: allowlist/denylist SVG (max 250 KiB, well-formed XML, satu root `<svg>`, larang `<script>`/`<foreignObject>`, larang atribut `on*`, href lokal `#` saja). |
| `pkg/game/creative/svg_test.go` | **Baru** — 99 baris test validasi SVG (valid, kosong, terlalu besar, malformed, script injection, event handler, external URI). |
| `pkg/game/creative/service.go` | Validasi `ValidateSVG` dipanggil saat `Kind == DRAWING`; `isValidKind` menerima `DRAWING`. |
| `internal/api/creative/index.go` | Semua error `ErrSVG*` dipetakan ke `400 Bad Request` (bukan 500). |

### Frontend (React/TypeScript)

| File | Perubahan |
| :--- | :--- |
| `web/src/features/quest/CreativeCanvas.tsx` | **Baru** — kanvas menggambar (`react-sketch-canvas`) dengan 5 warna, eraser, undo, clear, dan export SVG. |
| `web/src/features/creative/SubmissionForm.tsx` | Toggle **Write Story / Draw Canvas**; submit `DRAWING` via SVG. |
| `web/src/shared/utils/svg.ts` + `svg.test.ts` | **Baru** — `toSvgDataUri` untuk render SVG aman sebagai `<img>`. |
| `web/src/shared/components/molecules/CreativeCard.tsx` | Submission `DRAWING` dirender sebagai gambar, bukan teks mentah. |
| `web/src/shared/types/index.ts` | `SubmissionKind` kini menyertakan `'DRAWING'`. |

### Fix CI Build (`0cc1730`)

- `tsc -b` menemukan 4 error yang tak terlihat oleh `tsc --noEmit` (yang ternyata
  no-op karena `tsconfig.json` berisi `files:[]` + references).
- Fix: hapus import `React` tak terpakai & `type`-only import
  (`CreativeCanvas.tsx`), tambah `DRAWING` ke `SubmissionKind` frontend,
  label `DRAWING` di `CreativePage.tsx`, dan perbaiki variant `outline` →
  `ghost` di `ProfilePage.tsx` (bug pre-existing Phase 2A).
- `seed-live.js` ditambahkan sebagai script seeding live DB (membaca key dari `.env`).

### E2E (`6d92e21`)

- Selector quest E2E diselaraskan dengan live seed data (`golden-path`,
  `home`, `journal`, `quest`, `persistence`, `regression`).

---

## 🔒 Keamanan: SVG Sanitization

Fokus utama Phase 2B adalah **keamanan payload SVG** (mitigasi SVG-based XSS):

- ✅ Maksimum **250 KiB** (hindari abuse storage/render).
- ✅ **Well-formed XML**, tepat satu root `<svg>`.
- ✅ **Denylist tag**: `<script>` dan `<foreignObject>` ditolak.
- ✅ **Denylist atribut**: semua `on*` (event handler) ditolak.
- ✅ **Href eksternal** ditolak; hanya href lokal (`#id`) yang diperbolehkan.
- ✅ Diuji: `pkg/game/creative/svg_test.go` mencakup kasus injeksi.

---

## 🏁 Status & Rekomendasi

| Item | Status |
| :--- | :---: |
| Teknis (lint, unit, race, build, docker) | **PASS** |
| Fungsional (Playwright E2E) | **PASS** |
| Rekomendasi rilis | **GO** (Phase 2B content) |

### Catatan Lanjutan (di luar scope 2B)
- Node 20 pada GitHub Actions mulai deprecated → pertimbangkan upgrade action
  versions ke runner Node 24.
- Dependabot: 3 vulnerability terdeteksi di default branch (1 critical, 2 high)
  — belum ditangani.
- Verifikasi manual UI canvas (warna, eraser, undo) pada perangkat touch belum
  dieksekusi langsung; direkomendasikan smoke test manual sebelum tag rilis.
