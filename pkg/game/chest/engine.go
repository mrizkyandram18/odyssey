package chest

import (
	"context"
	"math/rand"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/relic"
)

// RandomSource abstracts randomness for testability and configurability.
type RandomSource interface {
	Float64() float64
	Intn(n int) int
}

// defaultRandomSource uses math/rand with a time-based seed.
type defaultRandomSource struct {
	r *rand.Rand
}

func newDefaultRandomSource() RandomSource {
	return &defaultRandomSource{r: rand.New(rand.NewSource(time.Now().UnixNano()))}
}

func (d *defaultRandomSource) Float64() float64 {
	return d.r.Float64()
}

func (d *defaultRandomSource) Intn(n int) int {
	return d.r.Intn(n)
}

// LootTable represents a weighted rarity table for a chest type.
type LootTable struct {
	ChestSlug string
	Weights   map[game.Rarity]float64
	Relics    []RelicWeight
}

// NewLootTable constructs a LootTable from drop table entries.
func NewLootTable(chestSlug string, entries []game.DropTableEntry) *LootTable {
	weights := make(map[game.Rarity]float64, len(entries))
	relics := make([]RelicWeight, 0)
	for _, e := range entries {
		if e.RelicID != 0 {
			relics = append(relics, RelicWeight{RelicID: e.RelicID, Weight: e.Weight})
		} else {
			weights[game.Rarity(e.Rarity)] = e.Weight
		}
	}
	return &LootTable{
		ChestSlug: chestSlug,
		Weights:   weights,
		Relics:    relics,
	}
}

// RollRarity picks a rarity based on cumulative weights using the provided RandomSource.
func (lt *LootTable) RollRarity(rs RandomSource) game.Rarity {
	total := 0.0
	for _, w := range lt.Weights {
		total += w
	}
	if total == 0 {
		return game.RarityCommon
	}
	roll := rs.Float64() * total
	cumulative := 0.0
	for rarity, weight := range lt.Weights {
		cumulative += weight
		if roll <= cumulative {
			return rarity
		}
	}
	return game.RarityCommon
}

// RollRelic picks a relic by weight using the provided RandomSource and catalog.
// Falls back to rarity-based selection if no relic weights are defined.
func (lt *LootTable) RollRelic(rs RandomSource, relCatalog relic.RelicCatalog) relic.RelicDefinition {
	if len(lt.Relics) > 0 {
		total := 0.0
		for _, w := range lt.Relics {
			total += w.Weight
		}
		if total == 0 {
			defs := relCatalog.ListAll(context.Background())
			if len(defs) > 0 {
				return defs[0]
			}
			return relic.DefaultRelicCatalog[0]
		}
		roll := rs.Float64() * total
		cumulative := 0.0
		for _, w := range lt.Relics {
			cumulative += w.Weight
			if roll <= cumulative {
				defs := relCatalog.ListAll(context.Background())
				for _, def := range defs {
					if def.ID == w.RelicID {
						return def
					}
				}
			}
		}
	}
	rarity := lt.RollRarity(rs)
	return PickRelicOfRarity(rarity, relCatalog)
}
