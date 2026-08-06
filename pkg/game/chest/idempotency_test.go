package chest

import (
	"context"
	"errors"
	"sync"
	"testing"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
)

type concurrentChestStore struct {
	mu     sync.Mutex
	chests map[int64]*game.Chest
	nextID int64
}

func newConcurrentChestStore() *concurrentChestStore {
	return &concurrentChestStore{
		chests: make(map[int64]*game.Chest),
	}
}

func (m *concurrentChestStore) CreateChest(ctx context.Context, ch *game.Chest) (*game.Chest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	ch.ID = m.nextID
	m.chests[ch.ID] = ch
	return ch, nil
}

func (m *concurrentChestStore) GetChest(ctx context.Context, chestID int64) (*game.Chest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.chests[chestID]
	if !ok {
		return nil, game.ErrNotFound
	}
	p := *ch
	return &p, nil
}

func (m *concurrentChestStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	return errors.New("use UpdateChestIfMatch")
}

func (m *concurrentChestStore) UpdateChestIfMatch(ctx context.Context, chestID int64, oldOpened bool, patch map[string]any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ch, ok := m.chests[chestID]
	if !ok {
		return false, game.ErrNotFound
	}
	if ch.Opened != oldOpened {
		return false, nil
	}
	if opened, ok := patch["opened"].(bool); ok {
		ch.Opened = opened
	}
	return true, nil
}

func (m *concurrentChestStore) ListChestsByUser(ctx context.Context, uid string) ([]game.Chest, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]game.Chest, 0, len(m.chests))
	for _, ch := range m.chests {
		if ch.UID == uid {
			result = append(result, *ch)
		}
	}
	return result, nil
}

type concurrentPlayerRelicStore struct {
	mu     sync.Mutex
	relics map[string]*game.PlayerRelic
}

func newConcurrentPlayerRelicStore() *concurrentPlayerRelicStore {
	return &concurrentPlayerRelicStore{relics: make(map[string]*game.PlayerRelic)}
}

func (m *concurrentPlayerRelicStore) GetPlayerRelic(ctx context.Context, uid, relicSlug string) (*game.PlayerRelic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	r, ok := m.relics[uid+"|"+relicSlug]
	if !ok {
		return nil, game.ErrNotFound
	}
	p := *r
	return &p, nil
}

func (m *concurrentPlayerRelicStore) CreatePlayerRelic(ctx context.Context, pr *game.PlayerRelic) (*game.PlayerRelic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := pr.UID + "|" + pr.RelicSlug
	if _, exists := m.relics[key]; exists {
		return nil, errors.New("duplicate relic")
	}
	m.relics[key] = pr
	return pr, nil
}

func (m *concurrentPlayerRelicStore) UpdatePlayerRelic(ctx context.Context, uid, relicSlug string, patch map[string]any) error {
	return nil
}

func (m *concurrentPlayerRelicStore) ListPlayerRelics(ctx context.Context, uid string) ([]game.PlayerRelic, error) {
	return nil, nil
}

func (m *concurrentPlayerRelicStore) CountUniqueRelics(ctx context.Context, uid string) (int, error) {
	return 0, nil
}

type concurrentRelicStore struct {
	mu     sync.Mutex
	relics []game.Relic
}

func (m *concurrentRelicStore) CreateRelic(ctx context.Context, r *game.Relic) (*game.Relic, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.relics = append(m.relics, *r)
	return r, nil
}

func (m *concurrentRelicStore) GetRelic(ctx context.Context, relicID int64) (*game.Relic, error) {
	return nil, game.ErrNotFound
}

func (m *concurrentRelicStore) ListRelics(ctx context.Context) ([]game.Relic, error) {
	return nil, nil
}

func (m *concurrentRelicStore) CountRelics(ctx context.Context, uid string) (int, error) {
	return 0, nil
}

type concurrentUserStore struct {
	mu     sync.Mutex
	player *game.Player
}

func (m *concurrentUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := *m.player
	return &p, nil
}

func (m *concurrentUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	return nil
}

func (m *concurrentUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return errors.New("use UpdateUserIfMatch")
}

func (m *concurrentUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.player.Version != version {
		return false, nil
	}
	if v, ok := patch["xp"].(int64); ok {
		m.player.XP = v
	}
	if v, ok := patch["level"].(int); ok {
		m.player.Level = v
	}
	if v, ok := patch["version"].(int); ok {
		m.player.Version = v
	}
	return true, nil
}

func TestOpenChest_ConcurrentRequests_OnlyOneSucceeds(t *testing.T) {
	chestStore := newConcurrentChestStore()
	relicStore := &concurrentRelicStore{}
	playerRelicStore := newConcurrentPlayerRelicStore()
	userStore := &concurrentUserStore{player: &game.Player{UID: "uid1", CrewID: "crew1", Level: 1, XP: 0, Version: 1}}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(chestStore, playerRelicStore, relicStore, userStore, engine)

	ch, err := svc.CreateChest(context.Background(), "uid1", "wooden-chest", "QUEST")
	if err != nil {
		t.Fatalf("create chest: %v", err)
	}

	var wg sync.WaitGroup
	results := make([]*OpenResult, 2)
	errs := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.OpenChest(context.Background(), ch.ID, "uid1")
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		}
	}

	if successCount != 1 {
		t.Errorf("expected exactly 1 successful open, got %d", successCount)
	}

	failCount := 0
	for _, err := range errs {
		if err != nil {
			failCount++
		}
	}
	if failCount != 1 {
		t.Errorf("expected exactly 1 failed open, got %d (errors: %v)", failCount, errs)
	}
}

func TestOpenChest_CompareAndSet_PreventsDoubleOpen(t *testing.T) {
	store := newConcurrentChestStore()
	relicStore := &concurrentRelicStore{}
	playerRelicStore := newConcurrentPlayerRelicStore()
	userStore := &concurrentUserStore{player: &game.Player{UID: "uid1", CrewID: "crew1", Level: 1, XP: 0, Version: 1}}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, playerRelicStore, relicStore, userStore, engine)

	ch, _ := svc.CreateChest(context.Background(), "uid1", "wooden-chest", "QUEST")

	result, err := svc.OpenChest(context.Background(), ch.ID, "uid1")
	if err != nil {
		t.Fatalf("first open: %v", err)
	}
	if result == nil || !result.Chest.Opened {
		t.Error("expected chest to be opened")
	}

	_, err = svc.OpenChest(context.Background(), ch.ID, "uid1")
	if err == nil {
		t.Fatal("expected error on second open")
	}
	if err.Error() != "chest already opened" {
		t.Errorf("expected 'chest already opened', got %v", err)
	}
}

func TestQuestCompletedHandler_NoDuplicateChestsOnReplay(t *testing.T) {
	store := newConcurrentChestStore()
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, newConcurrentPlayerRelicStore(), &concurrentRelicStore{}, &concurrentUserStore{}, engine)

	content := &mockContentGateway{
		quest: &gamecontent.QuestDefinition{
			Slug:        "morning-light",
			RewardChest: "wooden-chest",
		},
	}
	handler := NewQuestCompletedHandler(svc, content)

	dispatcher := events.NewDispatcher()
	dispatcher.Subscribe(events.EventTypeQuestCompleted, handler)

	for i := 0; i < 3; i++ {
		dispatcher.Publish(context.Background(), events.QuestCompletedEvent{
			QuestID:      1,
			CrewID:       "c1",
			TemplateSlug: "morning-light",
			PlayerUID:    "uid1",
		})
	}

	chests := store.chests
	questChests := 0
	for _, ch := range chests {
		if ch.Source == "QUEST" && ch.ChestSlug == "wooden-chest" {
			questChests++
		}
	}
	if questChests != 1 {
		t.Errorf("expected 1 quest chest, got %d", questChests)
	}
}
func (m *concurrentUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) { return nil, nil }
