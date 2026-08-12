package chest

import (
	"context"
	"testing"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/collection"
)

type mockRelicCatalog struct{}

func (m *mockRelicCatalog) Get(ctx context.Context, slug string) *relic.RelicDefinition {
	for i := range relic.DefaultRelicCatalog {
		if relic.DefaultRelicCatalog[i].Slug == slug {
			return &relic.DefaultRelicCatalog[i]
		}
	}
	return nil
}

func (m *mockRelicCatalog) ListAll(ctx context.Context) []relic.RelicDefinition {
	result := make([]relic.RelicDefinition, len(relic.DefaultRelicCatalog))
	copy(result, relic.DefaultRelicCatalog)
	return result
}

func TestWeightedRarity(t *testing.T) {
	dropTable := map[game.Rarity]float64{
		game.RarityCommon:    0.60,
		game.RarityUncommon:  0.25,
		game.RarityRare:      0.10,
		game.RarityEpic:      0.04,
		game.RarityLegendary: 0.01,
	}

	seen := make(map[game.Rarity]int)
	for i := 0; i < 10000; i++ {
		r := WeightedRarity(dropTable)
		seen[r]++
	}

	if seen[game.RarityCommon] == 0 {
		t.Error("expected Common to appear at least once")
	}
	if seen[game.RarityLegendary] == 0 {
		t.Error("expected Legendary to appear at least once")
	}
}

func TestGenerateRewards(t *testing.T) {
	ct := &ChestType{
		Slug: "wooden-chest",
		DropTable: map[game.Rarity]float64{
			game.RarityCommon: 1.0,
		},
	}
	catalog := &mockRelicCatalog{}
	rewards := GenerateRewards(ct, 3, catalog)
	if len(rewards) != 3 {
		t.Errorf("expected 3 rewards, got %d", len(rewards))
	}
	for _, r := range rewards {
		if r.Rarity != game.RarityCommon {
			t.Errorf("expected Common rarity, got %s", r.Rarity)
		}
	}
}

func TestPickRelicOfRarity(t *testing.T) {
	catalog := &mockRelicCatalog{}
	r := PickRelicOfRarity(game.RarityCommon, catalog)
	if r.Rarity != game.RarityCommon {
		t.Errorf("expected Common rarity, got %s", r.Rarity)
	}
	if r.Slug == "" {
		t.Error("expected non-empty slug")
	}
}

func TestGetChestType(t *testing.T) {
	ct := GetChestType("wooden-chest")
	if ct == nil {
		t.Fatal("expected wooden-chest to exist")
	}
	if ct.Name != "Wooden Gift" {
		t.Errorf("expected Wooden Gift, got %s", ct.Name)
	}

	missing := GetChestType("nonexistent")
	if missing != nil {
		t.Error("expected nil for nonexistent chest type")
	}
}

func TestLootTable_RollRarity(t *testing.T) {
	lt := &LootTable{
		GiftSlug: "test",
		Weights: map[game.Rarity]float64{
			game.RarityCommon: 1.0,
		},
	}
	rs := &fixedRandom{value: 0.5}
	r := lt.RollRarity(rs)
	if r != game.RarityCommon {
		t.Errorf("expected Common, got %s", r)
	}
}

func TestRewardEngine_GenerateRewards(t *testing.T) {
	ct := &ChestType{
		Slug: "wooden-chest",
		DropTable: map[game.Rarity]float64{
			game.RarityCommon: 1.0,
		},
	}
	engine := NewRewardEngine(nil, nil)
	rewards := engine.GenerateRewards(ct, 2)
	if len(rewards) != 2 {
		t.Errorf("expected 2 rewards, got %d", len(rewards))
	}
}

func TestGenerateRewards_WithRelicWeights(t *testing.T) {
	ct := &ChestType{
		Slug: "weighted-chest",
		RelicWeights: []RelicWeight{
			{CollectionID: 1, Weight: 0.8},
			{CollectionID: 2, Weight: 0.2},
		},
	}
	catalog := &mockRelicCatalog{}
	rewards := GenerateRewards(ct, 10, catalog)
	if len(rewards) != 10 {
		t.Errorf("expected 10 rewards, got %d", len(rewards))
	}
}

func TestPickRelicByWeight(t *testing.T) {
	weights := []RelicWeight{
		{CollectionID: 1, Weight: 1.0},
	}
	catalog := &mockRelicCatalog{}
	relic := PickRelicByWeight(weights, catalog)
	if relic.ID != 1 {
		t.Errorf("expected relic ID 1, got %d", relic.ID)
	}
}

func TestPickRelicByWeight_ZeroTotal(t *testing.T) {
	weights := []RelicWeight{
		{CollectionID: 1, Weight: 0},
	}
	catalog := &mockRelicCatalog{}
	relic := PickRelicByWeight(weights, catalog)
	if relic.ID == 0 {
		t.Error("expected a valid relic from fallback")
	}
}

func TestPickRelicByWeight_Empty(t *testing.T) {
	catalog := &mockRelicCatalog{}
	relic := PickRelicByWeight(nil, catalog)
	if relic.ID == 0 {
		t.Error("expected a valid relic from fallback")
	}
}

func TestRewardEngine_GenerateRewardsForChest_WithBalanceMultiplier(t *testing.T) {
	engine := NewRewardEngine(nil, nil)
	engine.SetBalance(balance.NewService(&mockBalanceStoreForChest{
		multiplier: 200,
	}))

	ct := engine.GetChestType(context.Background(), "wooden-chest")
	if ct == nil {
		t.Fatal("expected wooden-chest to exist")
	}

	rewards := engine.GenerateRewardsForChest(context.Background(), "wooden-chest", 100)
	commonCount := 0
	for _, r := range rewards {
		if r.Rarity == game.RarityCommon {
			commonCount++
		}
	}
	if commonCount == 0 {
		t.Error("expected some Common rewards with 2.0x multiplier")
	}
	if commonCount == 100 {
		t.Error("expected not all rewards to be Common with 2.0x multiplier")
	}
}

type mockBalanceStoreForChest struct {
	multiplier int64
}

func (m *mockBalanceStoreForChest) GetOverride(ctx context.Context, key string) (*balance.Override, error) {
	if key == "drop_rate_multiplier" {
		return &balance.Override{Key: key, Value: m.multiplier}, nil
	}
	return nil, balance.ErrConfigNotFound
}

func (m *mockBalanceStoreForChest) ListOverrides(ctx context.Context) ([]balance.Override, error) {
	return nil, nil
}

type fixedRandom struct {
	value float64
	idx   int
}

func (f *fixedRandom) Float64() float64 {
	return f.value
}

func (f *fixedRandom) Intn(n int) int {
	f.idx++
	return f.idx % n
}
