package chest

import (
	"context"
	"errors"
	"testing"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
)

type mockChestStore struct {
	gifts []game.Gift
	err    error
}

func (m *mockChestStore) CreateChest(ctx context.Context, ch *game.Gift) (*game.Gift, error) {
	if m.err != nil {
		return nil, m.err
	}
	ch.ID = int64(len(m.gifts) + 1)
	m.gifts = append(m.gifts, *ch)
	return &m.gifts[len(m.gifts)-1], nil
}

func (m *mockChestStore) GetChest(ctx context.Context, chestID int64) (*game.Gift, error) {
	return nil, game.ErrNotFound
}

func (m *mockChestStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	return nil
}
func (m *mockChestStore) UpdateChestIfMatch(ctx context.Context, chestID int64, oldOpened bool, patch map[string]any) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	for i := range m.gifts {
		if m.gifts[i].ID == chestID {
			if m.gifts[i].Opened != oldOpened {
				return false, nil
			}
			if opened, ok := patch["opened"].(bool); ok {
				m.gifts[i].Opened = opened
			}
			return true, nil
		}
	}
	return false, game.ErrNotFound
}

func (m *mockChestStore) ListChestsByUser(ctx context.Context, uid string) ([]game.Gift, error) {
	return nil, nil
}

type mockRelicStore struct{}

func (m *mockRelicStore) CreateRelic(ctx context.Context, r *game.Collection) (*game.Collection, error) {
	return r, nil
}

func (m *mockRelicStore) GetRelic(ctx context.Context, relicID int64) (*game.Collection, error) {
	return nil, game.ErrNotFound
}

func (m *mockRelicStore) ListRelics(ctx context.Context) ([]game.Collection, error) {
	return nil, nil
}

func (m *mockRelicStore) CountRelics(ctx context.Context, uid string) (int, error) {
	return 0, nil
}

type mockPlayerRelicStore struct{}

func (m *mockPlayerRelicStore) GetPlayerRelic(ctx context.Context, uid, relicSlug string) (*game.PlayerRelic, error) {
	return nil, game.ErrNotFound
}

func (m *mockPlayerRelicStore) CreatePlayerRelic(ctx context.Context, pr *game.PlayerRelic) (*game.PlayerRelic, error) {
	return pr, nil
}

func (m *mockPlayerRelicStore) UpdatePlayerRelic(ctx context.Context, uid, relicSlug string, patch map[string]any) error {
	return nil
}

func (m *mockPlayerRelicStore) ListPlayerRelics(ctx context.Context, uid string) ([]game.PlayerRelic, error) {
	return nil, nil
}

func (m *mockPlayerRelicStore) CountUniqueRelics(ctx context.Context, uid string) (int, error) {
	return 0, nil
}

type mockUserStore struct{}

func (m *mockUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	return nil, game.ErrNotFound
}

func (m *mockUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	return nil
}

func (m *mockUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return nil
}
func (m *mockUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	return true, nil
}

type mockContentGateway struct {
	quest *gamecontent.QuestDefinition
	err   error
}

func (m *mockContentGateway) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quest, nil
}

func TestCreateChest_Success(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	ch, err := svc.CreateChest(context.Background(), "uid1", "wooden-chest", "QUEST", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch == nil {
		t.Fatal("expected chest to be created")
	}
	if ch.UID != "uid1" {
		t.Errorf("expected UID uid1, got %s", ch.UID)
	}
	if ch.GiftSlug != "wooden-chest" {
		t.Errorf("expected chest slug wooden-chest, got %s", ch.GiftSlug)
	}
	if ch.Source != "QUEST" {
		t.Errorf("expected source QUEST, got %s", ch.Source)
	}
	if ch.Opened {
		t.Error("expected chest to be unopened")
	}
	if len(store.gifts) != 1 {
		t.Errorf("expected 1 chest in store, got %d", len(store.gifts))
	}
}

func TestCreateChest_DefinitionNotFound(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	_, err := svc.CreateChest(context.Background(), "uid1", "nonexistent-chest", "QUEST", "")
	if err == nil {
		t.Fatal("expected error for nonexistent chest definition")
	}
}

func TestCreateChest_StoreError(t *testing.T) {
	store := &mockChestStore{err: errors.New("db error")}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	_, err := svc.CreateChest(context.Background(), "uid1", "wooden-chest", "QUEST", "")
	if err == nil {
		t.Fatal("expected error from store")
	}
}

func TestQuestCompletedHandler_CreatesChest(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	content := &mockContentGateway{
		quest: &gamecontent.QuestDefinition{
			Slug:        "morning-light",
			RewardChest: "wooden-chest",
		},
	}
	handler := NewQuestCompletedHandler(svc, content)

	dispatcher := events.NewDispatcher()
	dispatcher.Subscribe(events.EventTypeQuestCompleted, handler)

	dispatcher.Publish(context.Background(), events.QuestCompletedEvent{
		MissionID:      1,
		FamilyID:       "c1",
		TemplateSlug: "morning-light",
		PlayerUID:    "uid1",
	})

	if len(store.gifts) != 1 {
		t.Errorf("expected 1 chest created, got %d", len(store.gifts))
	}
	if store.gifts[0].GiftSlug != "wooden-chest" {
		t.Errorf("expected chest slug wooden-chest, got %s", store.gifts[0].GiftSlug)
	}
	if store.gifts[0].UID != "uid1" {
		t.Errorf("expected UID uid1, got %s", store.gifts[0].UID)
	}
}

func TestQuestCompletedHandler_NoRewardChest(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	content := &mockContentGateway{
		quest: &gamecontent.QuestDefinition{
			Slug:        "morning-light",
			RewardChest: "",
		},
	}
	handler := NewQuestCompletedHandler(svc, content)

	dispatcher := events.NewDispatcher()
	dispatcher.Subscribe(events.EventTypeQuestCompleted, handler)

	dispatcher.Publish(context.Background(), events.QuestCompletedEvent{
		MissionID:      1,
		FamilyID:       "c1",
		TemplateSlug: "morning-light",
		PlayerUID:    "uid1",
	})

	if len(store.gifts) != 0 {
		t.Errorf("expected 0 gifts created, got %d", len(store.gifts))
	}
}

func TestQuestCompletedHandler_ContentError(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	content := &mockContentGateway{err: errors.New("db error")}
	handler := NewQuestCompletedHandler(svc, content)

	dispatcher := events.NewDispatcher()
	dispatcher.Subscribe(events.EventTypeQuestCompleted, handler)

	dispatcher.Publish(context.Background(), events.QuestCompletedEvent{
		MissionID:      1,
		FamilyID:       "c1",
		TemplateSlug: "morning-light",
		PlayerUID:    "uid1",
	})

	if len(store.gifts) != 0 {
		t.Errorf("expected 0 gifts created on content error, got %d", len(store.gifts))
	}
}

func TestQuestCompletedHandler_NilContent(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	handler := NewQuestCompletedHandler(svc, nil)

	dispatcher := events.NewDispatcher()
	dispatcher.Subscribe(events.EventTypeQuestCompleted, handler)

	dispatcher.Publish(context.Background(), events.QuestCompletedEvent{
		MissionID:      1,
		FamilyID:       "c1",
		TemplateSlug: "morning-light",
		PlayerUID:    "uid1",
	})

	if len(store.gifts) != 0 {
		t.Errorf("expected 0 gifts created with nil content, got %d", len(store.gifts))
	}
}

func TestQuestCompletedHandler_NonQuestEvent(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	content := &mockContentGateway{}
	handler := NewQuestCompletedHandler(svc, content)

	dispatcher := events.NewDispatcher()
	dispatcher.Subscribe(events.EventTypeQuestCompleted, handler)

	dispatcher.Publish(context.Background(), events.ChapterCompletedEvent{
		FamilyID:  "c1",
		Course: "ch1",
	})

	if len(store.gifts) != 0 {
		t.Errorf("expected 0 gifts for non-quest event, got %d", len(store.gifts))
	}
}

type mockBalanceStoreForRewardCount struct {
	overrides map[string]int64
}

func (m *mockBalanceStoreForRewardCount) GetOverride(ctx context.Context, key string) (*balance.Override, error) {
	if v, ok := m.overrides[key]; ok {
		return &balance.Override{Key: key, Value: v}, nil
	}
	return nil, balance.ErrConfigNotFound
}

func (m *mockBalanceStoreForRewardCount) ListOverrides(ctx context.Context) ([]balance.Override, error) {
	result := make([]balance.Override, 0, len(m.overrides))
	for k, v := range m.overrides {
		result = append(result, balance.Override{Key: k, Value: v})
	}
	return result, nil
}

func TestRewardCountForRarity_WithBalanceOverrides(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	balStore := &mockBalanceStoreForRewardCount{
		overrides: map[string]int64{
			"chest_reward_count_common":    5,
			"chest_reward_count_legendary": 10,
		},
	}
	bal := balance.NewService(balStore)
	if err := bal.Load(context.Background()); err != nil {
		t.Fatalf("failed to load balance: %v", err)
	}
	svc.SetBalance(bal)

	tests := []struct {
		name     string
		rarity   game.Rarity
		expected int
	}{
		{"Common overridden", game.RarityCommon, 5},
		{"Uncommon default", game.RarityUncommon, 2},
		{"Rare default", game.RarityRare, 2},
		{"Epic default", game.RarityEpic, 3},
		{"Legendary overridden", game.RarityLegendary, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.rewardCountForRarity(tt.rarity)
			if got != tt.expected {
				t.Errorf("rewardCountForRarity(%s) = %d, want %d", tt.rarity, got, tt.expected)
			}
		})
	}
}

func TestRewardCountForRarity_NilBalance(t *testing.T) {
	store := &mockChestStore{}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, &mockPlayerRelicStore{}, &mockRelicStore{}, &mockUserStore{}, engine)

	tests := []struct {
		rarity   game.Rarity
		expected int
	}{
		{game.RarityCommon, 1},
		{game.RarityUncommon, 2},
		{game.RarityRare, 2},
		{game.RarityEpic, 3},
		{game.RarityLegendary, 4},
	}

	for _, tt := range tests {
		got := svc.rewardCountForRarity(tt.rarity)
		if got != tt.expected {
			t.Errorf("rewardCountForRarity(%s) = %d, want %d (no balance)", tt.rarity, got, tt.expected)
		}
	}
}
func (m *mockUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}
