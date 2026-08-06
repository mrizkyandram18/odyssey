# UI Guidelines

## Core Tenet: Mobile-First, PWA-First

Odyssey's primary form factor is a smartphone held in one hand. Every design
decision starts there. Desktop is secondary — a wider canvas for the same
content, never a separate layout with more features.

## 1. Screen & Interaction

### Touch Targets

- **Minimum 48 × 48 px** for all interactive elements (taps, buttons, links).
- **8 px minimum** spacing between adjacent touch targets.
- Avoid hover-only affordances. Every interactive element must have a clear
  tap target.

### Viewport

- **Mobile base:** 375 px (iPhone SE-class) and up.
- **Breakpoints:** `sm` (640 px) for small tablets, `md` (768 px) for larger
  tablets / small laptops. No `lg`/`xl` breakpoints in the MVP — the layout
  does not structurally change beyond `md`.

### Navigation

- A **bottom tab bar** (3–4 tabs) for the primary sections: Home, World Map,
  Creative, Journal. On `md+`, this may collapse into a sidebar (same links,
  same order).
- A **back button** is provided by the browser/PWA context; do not override
  history navigation unless within a multi-step form.
- **Modal dialogs** use `aria-modal` and are dismissible via ESC / backdrop tap.
- **Overflow:** Prefer vertical scrolling pages over horizontal carousels. When
  a carousel is necessary, make swipe gestures the primary navigation (not
  dot indicators alone).

## 2. Typography

- **Primary font:** A clean, highly legible system font stack. No custom font
  loading in the MVP (to keep PWA fast).
- **Scale:** 4-step system. `text-xs` (captions), `text-sm` (body), `text-lg`
  (subhead), `text-xl` (title). Headings are `font-semibold` only — no bold
  weights beyond 600.
- **Line height:** 1.5 for body, 1.3 for headings.
- **Color:** Text is always high-contrast against its background. Body text
  meets WCAG AA. No decorative text in low-contrast colors.

## 3. Color & Theme

- **Palette:** A single, cohesive 5-color system: primary, secondary,
  background, surface, and one accent. All derived from CSS custom properties
  so a dark mode is a single class toggle.
- **Dark mode:** Default. The world is a "night adventure" — dark backgrounds
  with luminous accents fit the fantasy tone and save OLED battery. A light
  mode is available via system preference but is not the focus.
- **Action colors:** Primary = adventure progress (blue/cyan). Secondary =
  creative / social (purple). Accent = rewards / relics (amber/gold).
- **Status:** Success (green), error (red). Neutral for everything else.

## 4. PWA Behavior

### Installability

- A web app manifest is present (`manifest.webmanifest`).
- A service worker caches the shell and core assets.
- "Add to Home Screen" is prompted at an opportune moment (after the user
  completes their first quest, not on page load).

### Future: Offline

A full offline-first experience (local state persistence, sync-on-reconnect)
is deferred to a future phase — see [docs/future.md](future.md). For the MVP,
the PWA caches the shell and core assets for installability and fast loading,
but game-state contributions are sent immediately on submission.

### Performance

- **First paint:** < 1.5 s on a mid-tier phone over 3G.
- **Hero assets:** Illustrated world map and quest art are lazy-loaded; the
  critical path renders the home screen immediately.
- **No runtime JS bundles** larger than 50 KB gzipped for the initial route.

## 5. Component Design

### Structure

- **Atoms:** Button, Input, Badge, ExplorerIcon, Icon.
- **Molecules:** QuestCard, RelicDisplay, DailyTurnBanner, StreakBadge.
- **Organisms:** QuestDetail, CreativeCanvas, WorldMap, CrewDashboard.

### Principles

- Components are **self-documenting.** A component's props should make its
  behavior obvious without reading its internals.
- **No presentational/container split** in the React sense. Instead: data
  fetching lives in custom hooks (`useQuest`), components consume the return.
- **Composition over configuration.** Components accept children or slot
  props rather than dozens of boolean flags.
- **Accessibility:** Every component uses semantic HTML. Images have alt text.
  Interactive elements have accessible labels. Colors are never the only
  signal.

## 6. Motion & Feedback

### Animations

- **Duration:** 150 ms for micro-interactions (button press), 300 ms for
  transitions (screen change), 500 ms for story reveals.
- **Easing:** `cubic-bezier(0.4, 0, 0.2, 1)` for entrance, `ease-out` for
  dismissal.
- **Preference:** Respect `prefers-reduced-motion`. When set, skip all non-
  essential animations.
- **No auto-playing videos or GIFs.** All motion is triggered by user action
  or a brief, dismissible story reveal.

### Feedback

- **Tap:** Instant visual state change (opacity or scale 0.96).
- **Success:** A brief, celebratory animation (Relic collection, Chest open).
  Keep it short (500 ms) and non-blocking.
- **Error:** Inline, adjacent to the field. No modal alerts for input errors.
- **Progress:** Always visible. XP bars, Relic indicators, and streak badges
  are persistent in the top bar.

## 7. Content & Voice

- **Tone:** Warm, encouraging, slightly whimsical — like a storybook narrator
  who is also a cheerleader.
- **Microcopy:** Action buttons use verbs ("Explore," "Contribute," "Unlock").
  Status messages are positive-framed ("1 more to go!" not "1 remaining").
- **Localization:** All strings are externalized from day one. The MVP ships
  in Indonesian and English. A third language is possible if the family is
  multilingual — but do not add a fourth without clear demand.
- **Images:** Illustrated, not photographic. Consistent art style throughout.
  Alt text describes the story beat, not "image of a button."

## 8. Forms & Data Entry

- **Minimize typing.** On mobile, prefer selection, drawing, or photo
  capture over text input.
- **Single-column layout** for all forms. Labels above fields.
- **Inline validation** — errors appear as the user types or on blur, not on
  submit.
- **Optimistic updates** — the UI reflects the user's action immediately; the
  backend sync happens in the background with a rollback on failure.

## 9. Dark Patterns (Anti-Guide)

- No urgency timers for non-time-sensitive actions.
- No "you must log in daily or lose progress" messaging (streaks are rewards,
  not punishments).
- No hidden costs or ambiguous resource sinks.
- No notifications asking for reviews or ratings.
