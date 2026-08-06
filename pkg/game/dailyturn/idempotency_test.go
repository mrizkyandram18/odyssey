package dailyturn

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type concurrentDailyTurnStore struct {
	mu     sync.Mutex
	turns  []game.DailyTurn
	nextID int64
}

func newConcurrentDailyTurnStore() *concurrentDailyTurnStore {
	return &concurrentDailyTurnStore{}
}

func (m *concurrentDailyTurnStore) CreateDailyTurn(ctx context.Context, dt *game.DailyTurn) (*game.DailyTurn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nextID++
	dt.ID = m.nextID
	dt.CreatedAt = time.Now().UTC()
	m.turns = append(m.turns, *dt)
	return dt, nil
}

func (m *concurrentDailyTurnStore) UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error {
	return nil
}

func (m *concurrentDailyTurnStore) ListDailyTurns(ctx context.Context, uid string) ([]game.DailyTurn, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]game.DailyTurn, len(m.turns))
	copy(result, m.turns)
	return result, nil
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

func TestConsumeDailyTurn_IdempotentWithSameQuest(t *testing.T) {
	store := newConcurrentDailyTurnStore()
	svc := NewDailyTurnService(store, &DailyTurnConfig{XP: 10, MaxTurnsPerDay: 3, Timezone: "UTC"})

	turn1, err := svc.ConsumeDailyTurn(context.Background(), "user-1", "2026-08-05", "morning-light")
	if err != nil {
		t.Fatalf("first consume: %v", err)
	}
	if !turn1.Completed {
		t.Error("expected turn to be completed")
	}

	turn2, err := svc.ConsumeDailyTurn(context.Background(), "user-1", "2026-08-05", "morning-light")
	if err != ErrNoTurnsRemaining {
		t.Errorf("expected ErrNoTurnsRemaining on duplicate, got %v", err)
	}
	if turn2 != nil {
		t.Error("expected nil turn on duplicate")
	}
}

func TestConsumeDailyTurn_ConcurrentSameQuest_OnlyOneSucceeds(t *testing.T) {
	store := newConcurrentDailyTurnStore()
	svc := NewDailyTurnService(store, &DailyTurnConfig{XP: 10, MaxTurnsPerDay: 3, Timezone: "UTC"})

	var wg sync.WaitGroup
	results := make([]*game.DailyTurn, 5)
	errs := make([]error, 5)

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = svc.ConsumeDailyTurn(context.Background(), "user-1", "2026-08-05", "morning-light")
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
		t.Errorf("expected exactly 1 successful consume, got %d", successCount)
	}
}
