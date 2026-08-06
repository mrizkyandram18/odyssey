# Codebase Audit: Hardcoded Constants & Architectural Coupling

## Overview
Audit of the Odyssey game server codebase to identify hardcoded values, missing abstraction layers, data-driven design gaps, and schema consistency issues that block content-driven configuration.

## Scope
- `pkg/game/quest/` - Quest and realm catalog systems
- `pkg/game/world/` - Realm/zone definitions
- `pkg/game/balance/` - Balance override system
- `pkg/game/chest/` - Chest generation and reward engine
- `pkg/game/relic/` - Relic catalog
- `pkg/game/progression/` - Level progression system
- `pkg/game/content/` - Content type definitions
- `pkg/db/migrations/` - Database schema migrations
- `pkg/db/content_store_impl.go` - Content store DB mapping layer

---

## Finding 1: Realm Progress & Threshold - OVERRIDABLE via balance service

### File: `pkg/game/quest/catalog.go:45-52`
```go
const RealmProgressPerQuest = 25
const RealmCompletionThreshold = 100
```

### File: `pkg/game/quest/handler.go:175-187`
```go
progressPerQuest := RealmProgressPerQuest    // fallback: 25
completionThreshold := RealmCompletionThreshold  // fallback: 100
if h.realmCfg != nil {
    if def, ok := h.realmCfg.Get(realm); ok {
        if def.MaxProgress > 0 {
            completionThreshold = def.MaxProgress  // catalog value takes precedence
        }
    }
}
if h.balance != nil {
    progressPerQuest = h.balance.OverrideRealmProgressPerQuest(progressPerQuest)
    completionThreshold = h.balance.OverrideRealmCompletionThreshold(completionThreshold)
}
```

**Status:** WORKING - The handler uses a 3-tier fallback:
1. Catalog `MaxProgress` (per-realm, from `world/catalog.go`)
2. `RealmCompletionThreshold` constant (global fallback: 100)
3. Balance override (runtime override via `game_config` table)

Config keys: `balance/store.go:19-20`
- `KeyRealmProgressPerQuest = "realm_progress_per_quest"`
- `KeyRealmCompletionThreshold = "realm_completion_threshold"`

**Tests:** `quest/handler_test.go:477-493` confirms both overrides work with mock balance store.

---

## Finding 2: Realm MaxProgress - Override method WAS UNWIRED (now FIXED)

### File: `pkg/game/world/catalog.go:89-109`
```go
func (c *RealmCatalog) Override(slug, field, value string) bool {
    // handles "name" and "max_progress" fields with validation
}
```

**Previous issue:** The `Override()` method correctly handled per-realm `max_progress` overrides, but was never called. The `odyssey_system_config` table (key format: `realm:<slug>:<field>`) existed and was readable via `ConfigStore.GetSystemConfig()`, but no code connected them.

### Applied Fix

Added `ApplyOverrides(ctx context.Context, loader ConfigLoader)` method to `RealmCatalog` (`world/catalog.go:115-145`) that reads config rows and applies them. Wired in `api/dev/main.go:75-77`:

```go
realmCfg := world.DefaultRealmCatalog
if err := realmCfg.ApplyOverrides(ctx, repo.Config); err != nil {
    log.Printf("Warning: realm config overrides failed: %v", err)
}
```

**Status:** FIXED - Realm `MaxProgress` and `Name` are now overridable at runtime via `odyssey_system_config` table rows. 6 test cases added in `world/catalog_test.go`.

---

## Finding 3: Chest Reward Count - NOW OVERRIDABLE via balance service

### Files Modified

**`pkg/game/balance/store.go`** - Added 5 config keys:
```go
KeyChestRewardCountCommon    ConfigKey = "chest_reward_count_common"
KeyChestRewardCountUncommon  ConfigKey = "chest_reward_count_uncommon"
KeyChestRewardCountRare      ConfigKey = "chest_reward_count_rare"
KeyChestRewardCountEpic      ConfigKey = "chest_reward_count_epic"
KeyChestRewardCountLegendary ConfigKey = "chest_reward_count_legendary"
```

**`pkg/game/balance/service.go`** - Added `OverrideChestRewardCount(rarity, def)`:
```go
func (s *Service) OverrideChestRewardCount(rarity game.Rarity, def int) int {
    var key ConfigKey
    switch rarity {
    case game.RarityCommon:     key = KeyChestRewardCountCommon
    case game.RarityUncommon:   key = KeyChestRewardCountUncommon
    case game.RarityRare:       key = KeyChestRewardCountRare
    case game.RarityEpic:       key = KeyChestRewardCountEpic
    case game.RarityLegendary:  key = KeyChestRewardCountLegendary
    default:                    return def
    }
    return int(s.GetOverride(key, int64(def)))
}
```

**`pkg/game/chest/service.go`** - Added `rewardCountForRarity(rarity)` method on `ChestService` that applies balance overrides, and updated `OpenChest` to use it:
```go
func (s *ChestService) rewardCountForRarity(r game.Rarity) int {
    def := rewardCountByRarity(r)
    if s.balance != nil {
        return s.balance.OverrideChestRewardCount(r, def)
    }
    return def
}
```

**`pkg/game/chest/service_test.go`** - Added 2 test functions with 6 test cases covering overridden values and nil balance fallback.

**Status:** FIXED - Chest reward counts are now overridable per rarity via `game_config` table.

---

## Finding 4: Relics & Chests - Hardcoded in code, no DB-backed content

### File: `pkg/game/relic/catalog.go:9`
```go
var DefaultRelicCatalog = []RelicDefinition{
    // ~18 relic definitions with hardcoded slugs, names, rarities, stats
}
```

### File: `pkg/game/chest/catalog.go:9`
```go
var DefaultChestCatalog = []ChestType{
    // 4 chest types with hardcoded names, rarities, gold ranges
}
```

**Status:** NOT ADDRESSED - Both catalogs are hardcoded Go slices. While content types exist (`content/types.go`), there's no DB-backed loading path for relic or chest definitions. The balance service's `OverrideDropRateMultiplier` only adjusts overall drop rate, not per-relic drop rates.

---

## Finding 5: Progression - XP IS overridable

### File: `pkg/game/progression/service.go:23`
```go
return int64(level-1) * XPPerLevel
```

### File: `pkg/game/progression/service.go:111-133`
```go
XPPerLevel: s.balance.OverrideXPForLevel(def.XPPerLevel),
```

**Status:** WORKING - XP-per-level is a simple linear formula, but the `XPPerLevel` value itself is overridable via the balance service (`balance.KeyXPPerLevel = "xp_per_level"` config key). The `ProgressionService.applyBalanceOverrides()` method applies this at startup.

**Limitation:** The formula itself is linear (`level * XPPerLevel`). A per-level XP curve table would allow non-linear progression, but the current system supports at least global XP adjustment.

---

## Finding 6: Content Store Type Mapping

### File: `pkg/game/content/types.go` (QuestDefinition)
```go
type QuestDefinition struct {
    Slug                  string   `json:"slug"`
    Level                 int      `json:"level"`
    RealmSlug             string   `json:"realm_slug"`
    RequiredQuestSlug     string   `json:"required_quest_slug"`      // single
    RequiredQuestSlugs    []string `json:"required_quest_slugs"`     // plural
    MaxProgress           int      `json:"max_progress"`
    Rewards               []QuestReward `json:"rewards"`
}
```

### File: `pkg/db/content_store_impl.go`
```go
requiredQuestSlugs := []string{}
if d.RequiredQuestSlugs != nil {
    requiredQuestSlugs = d.RequiredQuestSlugs
} else if d.RequiredQuestSlug != "" {
    requiredQuestSlugs = []string{d.RequiredQuestSlug}
}
```

**Status:** WORKING - The content store correctly handles the dual `RequiredQuestSlug`/`RequiredQuestSlugs` backward-compatible fields. The `MaxProgress` from content types maps through to `RealmDefinition.MaxProgress` in the content store.

---

## Finding 7: DB Migrations Schema Evolution

### File: `pkg/db/migrations/`
- Timestamped SQL migration files with `up`/`down` functions
- `20250807_000000_create_realm_definitions.go` - Adds `max_progress` column to realm definitions table
- `odyssey_system_config` table for key-value config rows (referenced in `world/catalog.go:15` comment)
- `game_config` table for balance overrides (used by `balance.Store`)

**Status:** WORKING - Migration system is in place. Two config tables exist:
- `game_config` - Used by balance service for numeric overrides (`balance/store.go`, `db/config.go`)
- `odyssey_system_config` - Available via `ConfigStore.GetSystemConfig()` - **now wired** to `RealmCatalog.ApplyOverrides()` at startup (`world/catalog.go:115`, `api/dev/main.go:75`)

---

## Summary Table

| # | Issue | Files Affected | Status | Severity |
|---|-------|----------------|--------|----------|
| 1 | Realm progress/threshold constants | `quest/catalog.go:45-52`, `quest/handler.go:175-187` | Working (balance overrides applied) | N/A |
| 2 | Realm MaxProgress override was un-wired | `world/catalog.go:58-81` | **FIXED** - `ApplyOverrides()` added and wired in `main.go:75` | N/A |
| 3 | Chest reward count by rarity | `chest/service.go:137-152` | **FIXED** - Balance overrides added (`balance/store.go`, `service.go`) | N/A |
| 4 | Relic/chest hardcoded in Go source | `relic/catalog.go:9`, `chest/catalog.go:9` | Not addressed | High |
| 5 | XP-per-level formula | `progression/service.go:23,111` | Working (global override exists) | Low |
| 6 | Content store type mapping | `content/types.go`, `db/content_store_impl.go` | Working | N/A |
| 7 | DB migrations & dual config tables | `pkg/db/migrations/` | Working | N/A |

## Recommended Next Steps (Priority Order)

1. **Done:** Wired realm catalog overrides at startup - `ApplyOverrides()` reads `odyssey_system_config` rows and applies to `RealmCatalog` (`world/catalog.go:115`, `api/dev/main.go:75`)
2. **Done:** Added balance overrides for chest reward counts - 5 new config keys + `OverrideChestRewardCount` method, wired into `ChestService.OpenChest` (`balance/store.go:23-27`, `service.go:84-103`, `chest/service.go:161-169`)
3. **Migrate relic/chest definitions to DB-backed content** - Add content types and load path for `DefaultRelicCatalog` and `DefaultChestCatalog` (high effort, high value)