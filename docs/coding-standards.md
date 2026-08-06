# Coding Standards

This document defines conventions for the Odyssey codebase. It is divided by
language / layer. When in doubt, match the existing code in the repository;
this document codifies what "existing" will look like once development begins.

---

## Go (Backend)

### Module & Tooling

- **Module path:** `odyssey` (matches the repo name).
- **Go version:** 1.25+ (matching Family Reward).
- **Dependencies:** Mostly standard library for the core. Allow only the
  minimum third-party dependencies required for:
  - Supabase REST (direct `net/http`, no SDK — we use raw HTTP like Family Reward).
  - Firestore Admin (for the Gatekeeper auth adapter, read-only).
  - UUID generation (`github.com/google/uuid`).
  - Env loading in dev (`github.com/joho/godotenv`).
- **No web frameworks** (no Gin, Echo, Fiber). Handlers implement
  `http.HandlerFunc`.

### File & Package Organization

- One package per directory.
- Each `api/<resource>/index.go` contains exactly one `Handler` function.
- Shared logic lives in `pkg/<domain>/`, never in `api/`.
- The **auth adapter** (`pkg/auth/firestore.go`) is the single module that
  touches Firestore. The domain never imports Firebase packages.
- Test files live next to their source (`foo.go` + `foo_test.go`).
- No `internal` keyword enforcement yet (single repo); prefer explicit
  package namespacing via `pkg/`.

### Naming

- **Packages:** short, lowercase, singular (`auth`, `game`, `db`, `shared`).
- **Functions:** `CamelCase`, exported = `VerbNoun` (`Verify`, `IssueSession`,
  `GetUser`). Unexported = `camelCase`.
- **Types:** `Structs` for domain objects; `Interfaces` describe capabilities
  (`Authenticator`, `Store`). Avoid single-letter type params.
- **Constants:** `PascalCase` if exported (`SessionKindUser`); the codebase
  favors `UPPER_SNAKE` only for compile-time immutable config constants.
- **Errors:** Prefer typed sentinel errors (`var ErrSessionExpired = errors.New(...)`)
  and `errors.Is`/`errors.As` checks. Wrap with `fmt.Errorf("context: %w", err)`.

### Error Handling

- Always return errors up the call stack; do not `panic` in request paths.
- Log at the boundary (handler), not in `pkg/`. `pkg/` functions return errors;
  the handler decides logging level and HTTP status.
- Map domain errors to HTTP status codes in a single helper (`WriteJSONError`).

### HTTP Handlers

```go
// api/<resource>/index.go
func Handler(w http.ResponseWriter, r *http.Request) {
    // 1. CORS
    // 2. Method check
    // 3. Auth (authorize)
    // 4. Parse body
    // 5. Call interactor in pkg/
    // 6. Write JSON
}
```

- Set `Content-Type: application/json` on every response.
- Use a shared `WriteJSON(w, code, data)` helper (no inline `json.NewEncoder`
  outside that helper).
- CORS: match the resource being developed; default to allowing the family's
  deployed PWA origin.

### Database Access (Supabase REST)

- All table names use the `odyssey_` prefix.
- Column names are `snake_case`.
- Timestamps: `created_at` and `updated_at`, defaulting to
  `timezone('utc'::text, now())` in Supabase (matches Family Reward convention).
- Use PostgREST filters in URL paths, not raw SQL where avoidable.
- Return `return=representation` on POST/PATCH to read back computed IDs.
- Optimistic concurrency: read-then-patch with a filter on the previous value
  (same pattern as Family Reward's `UpdateStatus`).

### Logging

- Use the standard `log` package (or a structured logger TBD).
- Log at request boundaries, not inside `pkg/` domain logic.
- Never log secrets, tokens, or full PII.

### Testing

- Use the standard `testing` package only (no test framework dependencies).
- Table-driven tests for pure functions.
- Mocks for `Store` / `Authenticator` interfaces (inject, never global).
- Integration tests against a local Supabase are optional and separate.

## TypeScript / React (Frontend)

### Tooling

- **Runtime:** React 19.
- **Build:** Vite.
- **Language:** TypeScript (strict mode).
- **Styling:** Tailwind CSS. No CSS-in-JS.
- **Lint/Format:** ESLint + Prettier (configs to be added at scaffold).

### File Organization

```
src/
├── app/              # Shell, routing, global providers
├── features/<name>/  # Co-located feature: component + hooks + types
├── shared/
│   ├── components/   # Design-system primitives (atoms, molecules, organisms)
│   ├── hooks/        # Data fetching + state hooks
│   ├── lib/          # API client, session storage
│   └── types/        # .ts shared types
└── assets/
```

- **Feature slices** are self-contained: a feature owns its UI, hooks, and
  types. Shared cross-feature code goes in `shared/`.
- **No barrel files** beyond `shared/`. Import by explicit path.

### Naming

- **Components:** `PascalCase`, file name matches export (`Button.tsx`).
- **Hooks:** `use`-prefixed (`useQuests`).
- **Types:** `PascalCase` for interfaces/types (`Quest`, `UserSession`).
- **Variables/functions:** `camelCase`.
- **Constants:** `UPPER_SNAKE_CASE` for module-level constants.

### Typing Discipline

- `strict: true` in `tsconfig.json`.
- No `any`. Use `unknown` for untyped external data, then narrow.
- Define explicit request/response types for every API call (generated or
  hand-written in `shared/types`).
- Prefer `interface` for domain shapes; `type` for unions and aliases.

### React Conventions

- Function components with hooks only. No class components.
- Keep components small (aim for < 100 lines). Extract logic into hooks.
- Props interfaces are co-located in the component file.
- Use `type` for props to allow extension via intersection.

### Data Fetching

- A single API client module (`shared/lib/api.ts`) wraps `fetch`.
- All requests go through the client (uniform headers, error handling,
  serialization).
- The client automatically attaches the session token from
  `localStorage` (via `Authorization: Bearer` + `X-User-Session`).
- Contributions are submitted immediately on action; the UI shows an
  optimistic result and rolls back on failure.

### Styling

- Tailwind utility classes only.
- Design tokens via CSS custom properties (e.g., `--color-primary`).
- Mobile-first: write base styles for mobile, use `sm:`/`md:` breakpoints
  only for enhancement, never regression.
- No inline styles (except dynamic values that can't be classes).

## Git & Workflow

- **Branch strategy:** `main` is always deployable. Features branch from `main`.
- **Commits:** Conventional, imperative mood (`docs: add vision`,
  `feat(game): add quest resolver`).
- **PRs:** Required for all merges. Squash on merge.

## Cross-Language Consistency

| Concept | Go | TypeScript |
|---|---|---|
| Session | HMAC-signed token in `Authorization` header | Read/written from `localStorage` |
| Errors | `JSONError(w, msg, code)` | `apiError.message` field |
| Timestamps | `time.Time` UTC | `Date` (UTC) |
| IDs | `int64` or `string` | `string` (JSON-safe) |
