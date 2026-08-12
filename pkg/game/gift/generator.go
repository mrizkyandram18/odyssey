package chest

import (
	"context"
	"math/rand"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/collection"
)

// WeightedRarity picks a rarity based on cumulative weights.
func WeightedRarity(dropTable map[game.Rarity]float64) game.Rarity {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	total := 0.0
	for _, w := range dropTable {
		total += w
	}
	roll := r.Float64() * total
	cumulative := 0.0
	for rarity, weight := range dropTable {
		cumulative += weight
		if roll <= cumulative {
			return rarity
		}
	}
	return game.RarityCommon
}

// RelicWeight associates a relic ID with a selection weight.
type RelicWeight struct {
	CollectionID int64
	Weight  float64
}

// GenerateRewards produces a list of relic rewards for a given chest type.
// When the chest type has RelicWeights, selection is proportional to weight.
// Otherwise, falls back to rarity-based weighted selection with uniform relic pick.
func GenerateRewards(ct *ChestType, count int, relCatalog relic.RelicCatalog) []RewardItem {
	rewards := make([]RewardItem, count)
	for i := 0; i < count; i++ {
		if len(ct.RelicWeights) > 0 {
			relicDef := PickRelicByWeight(ct.RelicWeights, relCatalog)
			rewards[i] = RewardItem{
				CollectionSlug: relicDef.Slug,
				Name:      relicDef.Name,
				Rarity:    relicDef.Rarity,
				Journey:     relicDef.Journey,
			}
		} else {
			rarity := WeightedRarity(ct.DropTable)
			relicDef := PickRelicOfRarity(rarity, relCatalog)
			rewards[i] = RewardItem{
				CollectionSlug: relicDef.Slug,
				Name:      relicDef.Name,
				Rarity:    rarity,
				Journey:     relicDef.Journey,
			}
		}
	}
	return rewards
}

// PickRelicByWeight selects a relic proportional to its weight from the provided weights.
func PickRelicByWeight(weights []RelicWeight, relCatalog relic.RelicCatalog) relic.RelicDefinition {
	total := 0.0
	for _, w := range weights {
		total += w.Weight
	}
	if total == 0 {
		defs := relCatalog.ListAll(context.Background())
		if len(defs) > 0 {
			return defs[0]
		}
		return relic.DefaultRelicCatalog[0]
	}
	roll := rand.New(rand.NewSource(time.Now().UnixNano())).Float64() * total
	cumulative := 0.0
	for _, w := range weights {
		cumulative += w.Weight
		if roll <= cumulative {
			defs := relCatalog.ListAll(context.Background())
			for _, def := range defs {
				if def.ID == w.CollectionID {
					return def
				}
			}
		}
	}
	defs := relCatalog.ListAll(context.Background())
	if len(defs) == 0 {
		return relic.DefaultRelicCatalog[0]
	}
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	return defs[src.Intn(len(defs))]
}

// PickRelicOfRarity returns a random relic definition of the given rarity from the provided catalog.
func PickRelicOfRarity(r game.Rarity, relCatalog relic.RelicCatalog) relic.RelicDefinition {
	defs := relCatalog.ListAll(context.Background())
	candidates := make([]relic.RelicDefinition, 0)
	for _, def := range defs {
		if def.Rarity == r {
			candidates = append(candidates, def)
		}
	}
	if len(candidates) == 0 {
		return relic.DefaultRelicCatalog[0]
	}
	src := rand.New(rand.NewSource(time.Now().UnixNano()))
	return candidates[src.Intn(len(candidates))]
}

// MakeRelicCatalog builds a relic catalog map from definitions.
func MakeRelicCatalog(defs []relic.RelicDefinition) map[string]relic.RelicDefinition {
	m := make(map[string]relic.RelicDefinition, len(defs))
	for _, def := range defs {
		m[def.Slug] = def
	}
	if len(m) == 0 {
		for _, def := range relic.DefaultRelicCatalog {
			m[def.Slug] = def
		}
	}
	return m
}
