# Phase 4 Walkthrough — Configurable Family Task & Reward Platform

Odyssey has been elevated into a configurable private family daily task & reward engine supporting **Video**, **Quiz**, **Photo Upload**, **Document Upload**, **Text Response**, and **Mini Game** task types.

---

## 1. End-to-End User Experience

### 1.1 Member Daily Stepper (`web/src/features/stepper/LinearPath.tsx`)
- Displays daily tasks in linear order with status-aware nodes:
  - 🎥 **Video Task** (`VideoQuizModal.tsx`): Watch YouTube video and complete questions.
  - 🧠 **Quiz Task** (`VideoQuizModal.tsx`): Multiple choice questions with zero-leak answer keys.
  - 📸 **Photo Upload Task** (`CameraCaptureModal.tsx`): Native camera capture, client-side compression, timestamp watermark, and upload.
  - 📄 **Document Upload Task** (`DocUploadModal.tsx`): Direct download of admin template/attachment, external editing, and upload of completed document.
  - ✍️ **Text Response Task** (`TextResponseModal.tsx`): Prompt instruction with real-time character counter and server-enforced length bounds.
  - 🎮 **Mini Game Task** (`MiniGameModal.tsx`): Interactive Memory Card Challenge with difficulty levels, move counter, timer, target score validation, and celebration.

---

## 2. Admin Task Builder & Review Queue (`web/src/features/admin/AdminPage.tsx`)

### 2.1 Dynamic Task Builder
- Admin selects task type from dropdown (`Video`, `Quiz`, `Photo Upload`, `Document Upload`, `Text Response`, `Mini Game`).
- Form dynamically presents the appropriate configuration fields:
  - **Video**: YouTube URL, minimum duration
  - **Quiz**: Questions, multiple choice options (A, B, C, D), server-side correct answer key
  - **Photo**: Instructions, max files
  - **Document**: Document template URL, attachment name, accepted extensions
  - **Text**: Writing prompt, minimum characters, maximum characters
  - **Mini Game**: Game type, difficulty level (Easy, Medium, Hard), target score

### 2.2 Submissions Verification Queue
- Rich media preview:
  - High-res photo zoom modal
  - Document download button with file size metadata
  - Formatted text responses with character counts
  - Mini-game completion scores and step counts
- One-click Approve (+Coins & EXP) and Reject with optional notes.

---

## 3. Security & Validation Verification

- **Family Tenant Isolation**: 100% enforced across all endpoints.
- **Double Reward Protection**: Tested with 100 concurrent requests; exactly 1 reward granted.
- **Upload Safety**: Disallows executable extensions (`.exe`, `.sh`, `.bat`, `.php`), enforces 10MB limits, sanitizes path traversal.
- **Quiz Zero-Leakage**: Deep recursive sanitization removes all answer keys before client transmission.
- **Authoritative Server Scoring**: Mini game scores validated against server-configured target scores.
