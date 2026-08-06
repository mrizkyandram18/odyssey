package chest

import (
	"context"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/relic"
)

// RewardEngine generates relic rewards when a chest is opened.
type RewardEngine struct {
	chestCatalog ChestCatalog
	relCatalog   relic.RelicCatalog
	rs           RandomSource
	balance      *balance.Service
}

// NewRewardEngine constructs a RewardEngine with optional stores.
// When stores are nil, only the hardcoded fallback catalog is used.
func NewRewardEngine(defStore game.ChestDefinitionStore, relDefStore game.RelicDefinitionStore) *RewardEngine {
	return &RewardEngine{
		chestCatalog: NewContentChestCatalog(defStore),
		relCatalog:   relic.NewContentRelicCatalog(relDefStore),
		rs:           newDefaultRandomSource(),
	}
}

// NewRewardEngineWithRandomSource constructs a RewardEngine with a custom RandomSource.
func NewRewardEngineWithRandomSource(defStore game.ChestDefinitionStore, relDefStore game.RelicDefinitionStore, rs RandomSource) *RewardEngine {
	return &RewardEngine{
		chestCatalog: NewContentChestCatalog(defStore),
		relCatalog:   relic.NewContentRelicCatalog(relDefStore),
		rs:           rs,
	}
}

// SetBalance attaches a balance service for drop rate overrides.
func (e *RewardEngine) SetBalance(b *balance.Service) {
	e.balance = b
}

// GetChestType returns a chest definition, trying the catalog first then falling back.
func (e *RewardEngine) GetChestType(ctx context.Context, slug string) *ChestType {
	return e.chestCatalog.Get(ctx, slug)
}

// GenerateRewardsForChest produces a list of relic rewards for the given chest slug.
func (e *RewardEngine) GenerateRewardsForChest(ctx context.Context, chestSlug string, count int) []RewardItem {
	ct := e.GetChestType(ctx, chestSlug)
	if ct == nil {
		return nil
	}
	if e.balance != nil {
		ct = e.withDropRateMultiplier(ct)
	}
	return e.GenerateRewards(ct, count)
}

// withDropRateMultiplier returns a cloned ChestType with adjusted drop table weights.
func (e *RewardEngine) withDropRateMultiplier(ct *ChestType) *ChestType {
	multiplier := e.balance.OverrideDropRateMultiplier(1.0)
	if multiplier == 1.0 {
		return ct
	}
	clone := *ct
	clone.DropTable = make(map[game.Rarity]float64, len(ct.DropTable))
	for r, w := range ct.DropTable {
		clone.DropTable[r] = w * multiplier
	}
	return &clone
}

// GenerateRewards produces rewards from a specific ChestType using the engine's relic catalog.
func (e *RewardEngine) GenerateRewards(ct *ChestType, count int) []RewardItem {
	return GenerateRewards(ct, count, e.relCatalog)
}

// RollRarity rolls a rarity from the chest's drop table.
func (e *RewardEngine) RollRarity(ct *ChestType) game.Rarity {
	lt := &LootTable{Weights: ct.DropTable}
	return lt.RollRarity(e.rs)
}
