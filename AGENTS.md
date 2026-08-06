# Schema ↔ Go Type Consistency Audit

## Scope
- Migrations: `scripts/migrations/001_initial_schema.sql` … `012_reliability_idempotency.sql`
- DB layer: `pkg/db/` (structs + store implementations)
- Domain models: `pkg/game/domain.go`, `pkg/game/content/types.go`
- Service layer: `pkg/game/{balance,chest,quest,catalog,reward_engine,progression,creative}/*.go`
- Entry point: `api/dev/main.go`

## PostgREST JSON Semantics (key to every judgment)
PostgREST encodes JSONB columns as **native JSON values** (not stringified). TEXT columns are encoded as JSON strings. BYTEA would be base64. All unmarshaling into Go structs happens via `encoding/json` on the full HTTP response body. This means:
- JSONB column → Go `json.RawMessage`, `[]T`, `map[string]T`, or any `json.Unmarshal` into a struct field that matches the JSON shape
- JSONB column → Go `string` — **FAILS at unmarshal** (can't unmarshal JSON object/array into string)
- TEXT column → Go `string` — works (JSON string unmarshals cleanly into Go string)

---

## Findings

### 1. `challenge_defs` — JSONB column vs `string` Go field
- **Schema** (migration 006): `challenge_defs JSONB NOT NULL DEFAULT '[]'`
- **DB struct** (`pkg/db/content.go:51`): `ChallengeDefs string`
- **Consumer** (`pkg/db/content_store_impl.go:247-248`):
  ```go
  if d.ChallengeDefs != "" {
      if err := json.Unmarshal([]byte(d.ChallengeDefs), &challengeDefs); err != nil {
  ```
- **Seed** (`scripts/migrations/010_seed_definitions.sql:35`): populates with JSON array literal
- **Analysis**: PostgREST returns JSONB as native JSON. A row like `{"challenge_defs":[{"slug":"find-the-dew",...}]}` arrives. `json.Unmarshal` into `struct{ ChallengeDefs string }` **FAILS**: `cannot unmarshal array into Go struct field`. The `if d.ChallengeDefs != ""` guard doesn't help — unmarshal already errored. `mapQuestDefinition` returns an error, which `ListQuests`/`GetQuest` silently swallows via `continue` (`content_store_impl.go:194`). Result: quests silently have empty `ChallengeDefs`, breaking quest creation (no challenges generated).
- **Confidence**: HIGH (reproduced failure mode, build passes because it's a runtime data-flow bug)
- **Status**: INCORRECT — confirmed mismatch

### 2. `odyssey_balance_configs.value` — JSONB column vs JSON value insertion
- **Schema** (migration 008): `value JSONB NOT NULL`
- **DB struct** (`pkg/db/balance.go:15`): `Value json.RawMessage`
- **Consumer** (`pkg/db/balance.go:62-66`): `json.Unmarshal(bc.Value, &val)` where `val` is `int64`
- **Seed** (`scripts/migrations/010_seed_definitions.sql:176`): `'100'::jsonb` — stores JSON number `100` (not string `"100"`)
- **Analysis**: `json.RawMessage` captures raw JSON bytes (`100`). `json.Unmarshal` of `100` into `int64` succeeds. Domain `Override.Value` is `int64`; service methods (`OverrideDropRateMultiplier`, `OverrideQuestRewardXP`, `OverrideAchievementThresholdMultiplier`) divide by 100. Seed values like `100` → multiplier `1.0`, `20` → `0.2`. **Correct for current seed data.**
- **Robustness note**: If a value is inserted as a JSON string `"100"` (e.g., via admin API without `::jsonb` cast), `json.Unmarshal` into `int64` fails, `val` defaults to `0`, and the override silently breaks (e.g., `0/100 = 0.0` multiplier). The seed bypasses this by using `::jsonb`.
- **Confidence**: HIGH on current correctness; MEDIUM on robustness
- **Status**: CORRECT for current seed data

### 3. `odyssey_system_config.value` — TEXT column vs return type
- **Schema** (migration 003): `value TEXT NOT NULL`
- **DB struct**: No dedicated struct; `GetSystemConfig` returns `string(raw)` (entire response body)
- **Consumer** (`pkg/game/world/catalog.go:123-133`): expects `raw` to be parseable JSON, unmarshals into `[]sysConfigRow` with `Value json.RawMessage`
- **Analysis**: `string(raw)` returns the **entire PostgREST JSON response** (e.g., `[{"key":"...","value":"...","created_at":"...","updated_at":"..."}]`), not just the `value` column. The consumer then unmarshals the whole array and extracts `rows[0].Value`. This works because PostgREST returns TEXT values as JSON strings, and `json.RawMessage` captures the raw JSON string bytes. However, if multiple rows match (shouldn't happen with PK lookup but theoretically), only `rows[0]` is used. If the key doesn't exist, `raw` is `null` or `[]`, and the guard `raw == "" || raw == "[]"` handles it.
- **Confidence**: HIGH
- **Status**: CORRECT (works, but fragile design — `string(raw)` is misleading; could be tightened)

### 4. `draft` columns — JSONB vs `json.RawMessage`
- **Schema** (migration 008): `draft JSONB` (nullable) on all 8 definition tables
- **DB structs** (`pkg/db/content.go` + `pkg/db/types.go:157,185`): `Draft json.RawMessage`
- **Analysis**: `json.RawMessage` correctly captures native JSON from PostgREST. On read it preserves raw bytes; on write via admin store (`content_admin_store.go`), `map[string]any` patches with `json.Marshal` serialize correctly. No `json.RawMessage`-into-`string` unmarshal path exists for `draft`.
- **Confidence**: HIGH
- **Status**: CORRECT

### 5. `required_quest_slugs` — JSONB vs `[]string`
- **Schema** (migration 011): `required_quest_slugs JSONB NOT NULL DEFAULT '[]'`
- **DB struct** (`pkg/db/content.go:56`): `RequiredQuestSlugs []string`
- **Consumer** (`pkg/db/content_store_impl.go:256-261`): `mapQuestDefinition` copies `d.RequiredQuestSlugs` directly
- **Analysis**: PostgREST returns JSONB array as native JSON array `[...]`. `json.Unmarshal` into `[]string` works for an array of string values. The seed data doesn't populate this column (uses default `'[]'`), so it's an empty array — unmarshals fine.
- **Confidence**: HIGH
- **Status**: CORRECT

### 6. `odyssey_creative_submissions.content` — TEXT vs `string`
- **Schema** (migration 003): `content TEXT NOT NULL`
- **DB struct** (`pkg/db/types.go:78`): `Content string`
- **Analysis**: PostgREST returns TEXT as JSON string. `json.Unmarshal` into `string` works cleanly. The migration comment says "base64-encoded" but actual column is TEXT and Go uses `string` — no encoding mismatch for the data path shown.
- **Confidence**: HIGH
- **Status**: CORRECT

### 7. `odyssey_chests.drop_table` — TEXT vs `string` (unused at runtime)
- **Schema** (migration 004): `drop_table TEXT`
- **DB struct** (`pkg/db/types.go:130`): `DropTable string`
- **Consumer**: `mapChest` (`pkg/db/chest.go:131`) passes it through. **`ChestService.OpenChest` does NOT read `ch.DropTable`** — it delegates to `RewardEngine.GenerateRewardsForChest` which reads from `ChestType.DropTable` (a `map[game.Rarity]float64`), populated from `odyssey_drop_tables` by `ContentChestCatalog` (`pkg/game/chest/catalog.go`), NOT from the per-chest instance's `drop_table` column.
- **Analysis**: The `drop_table` column on `odyssey_chests` is a leftover from migration 001's design ("Reward containers with known, fixed contents"). The current code path ignores it entirely — rewards are generated from the drop_tables definition table keyed by chest slug. `CreateChest` sets `DropTable: ""`. There's no type mismatch (TEXT↔string is valid), but the column appears dead/legacy.
- **Confidence**: HIGH on "unused"; the column is structurally valid but semantically vestigial
- **Status**: CORRECT mapping, but column is UNUSED (design debt)

### 8. `odyssey_audit_logs.old_value / new_value` — JSONB vs no Go struct
- **Schema** (migration 008): `old_value JSONB, new_value JSONB`
- **DB structs**: No audit log struct exists in `pkg/db/audit.go`. Audit is write-only (no read-back structs).
- **Analysis**: No Go struct attempts to read these JSONB columns. No type mismatch possible for current code. If audit reading is added later, use `json.RawMessage` or `map[string]any`.
- **Confidence**: HIGH
- **Status**: N/A (no read path; no mismatch)

---

## Summary Table

| # | Column | Table | Go Type | Schema Type | Status | Confidence |
|---|--------|-------|---------|-------------|--------|------------|
| 1 | `challenge_defs` | `odyssey_quest_definitions` | `string` | `JSONB` | **INCORRECT** — unmarshal fails; quests lose challenges | HIGH |
| 2 | `value` | `odyssey_balance_configs` | `json.RawMessage` → `int64` | `JSONB` | CORRECT (fragile: string values break) | HIGH / MED (robustness) |
| 3 | `value` | `odyssey_system_config` | `string(raw)` response body | `TEXT` | CORRECT (whole-response pattern, works) | HIGH |
| 4 | `draft` | all 8 definition tables | `json.RawMessage` | `JSONB` | CORRECT | HIGH |
| 5 | `required_quest_slugs` | `odyssey_quest_definitions` | `[]string` | `JSONB` | CORRECT | HIGH |
| 6 | `content` | `odyssey_creative_submissions` | `string` | `TEXT` | CORRECT | HIGH |
| 7 | `drop_table` | `odyssey_chests` | `string` | `TEXT` | CORRECT but UNUSED — vestigial column | HIGH |
| 8 | `old_value`, `new_value` | `odyssey_audit_logs` | (no struct) | `JSONB` | N/A — no read path | HIGH |

---

## Recommended Fixes

### Fix 1 (CRITICAL): `challenge_defs` — migrate `string` → `json.RawMessage`
**File**: `pkg/db/content.go:51`
```go
// Before:
ChallengeDefs string          `json:"challenge_defs"`
// After:
ChallengeDefs json.RawMessage `json:"challenge_defs"`
```
And update the consumer in `pkg/db/content_store_impl.go:245-251`:
```go
// Before:
if d.ChallengeDefs != "" {
    if err := json.Unmarshal([]byte(d.ChallengeDefs), &challengeDefs); err != nil {
        return nil, fmt.Errorf("parse challenge defs: %w", err)
    }
}
// After:
if len(d.ChallengeDefs) > 0 {
    if err := json.Unmarshal(d.ChallengeDefs, &challengeDefs); err != nil {
        return nil, fmt.Errorf("parse challenge defs: %w", err)
    }
}
```
This mirrors the pattern already used for `draft` columns (which correctly use `json.RawMessage`).

### Fix 2 (ROBUSTNESS): `balance_configs.value` — defensive parsing
The `mapBalanceOverride` function at `pkg/db/balance.go:62-66` silently defaults to `0` on unmarshal failure. Consider logging the error or supporting string-encoded numbers:
```go
// At minimum, surface the error for observability
if err := json.Unmarshal(bc.Value, &val); err != nil {
    // Could also try: json.Unmarshal(bc.Value, &json.Number) then parse
    val = 0 // current behavior
}
```

### Fix 3 (DESIGN): `odyssey_chests.drop_table` — remove or document
The `DropTable string` field and `drop_table TEXT` column are not consumed by any code path. Either:
- Remove the column via migration + remove the Go field, OR
- Document it as intentionally deprecated (chests now use `odyssey_drop_tables` definition table keyed by `chest_slug`)

---

## Validation
- `go build ./...` → BUILD OK (all findings are runtime data-flow issues, not compile errors)
- Tests pass with mocked/stubbed JSON (they use hand-crafted JSON strings, not PostgREST-native JSONB shape — this is why the `challenge_defs` bug isn't caught by existing tests)
