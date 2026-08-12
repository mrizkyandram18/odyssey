package reward

import (
	"context"
	"fmt"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockUserStore struct {
	player *game.Player
	// updates records successful patches for assertions
	updates []map[string]any
	// failMatch forces UpdateUserIfMatch to return ok=false
	failMatch bool
}

func (m *mockUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	if m.player == nil || m.player.UID != uid {
		return nil, game.ErrNotFound
	}
	// return a copy so callers cannot mutate without update path
	p := *m.player
	return &p, nil
}
func (m *mockUserStore) CreateUser(ctx context.Context, p *game.Player) error { return nil }
func (m *mockUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return nil
}
func (m *mockUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	if m.failMatch {
		return false, nil
	}
	if m.player == nil || m.player.UID != uid || m.player.Version != version {
		return false, nil
	}
	if c, ok := patch["coins"].(int64); ok {
		m.player.Coins = c
	}
	if v, ok := patch["version"].(int); ok {
		m.player.Version = v
	}
	m.updates = append(m.updates, patch)
	return true, nil
}
func (m *mockUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}

type mockLedgerStore struct {
	rows []game.RewardLedger
}

func (m *mockLedgerStore) CreateLedger(ctx context.Context, ledger *game.RewardLedger) error {
	if ledger == nil {
		return fmt.Errorf("nil ledger")
	}
	row := *ledger
	m.rows = append(m.rows, row)
	return nil
}
func (m *mockLedgerStore) ListByUser(ctx context.Context, userID string) ([]game.RewardLedger, error) {
	out := make([]game.RewardLedger, 0)
	for _, r := range m.rows {
		if r.UserID == userID {
			out = append(out, r)
		}
	}
	return out, nil
}

func newTestService(player *game.Player, ledger *mockLedgerStore) *Service {
	users := &mockUserStore{player: player}
	if ledger == nil {
		ledger = &mockLedgerStore{}
	}
	repo := &game.Repository{
		Users:         users,
		RewardLedgers: ledger,
	}
	return NewService(repo)
}

func TestGrantQuestReward_AddsFiveCoinsAndLedger(t *testing.T) {
	player := &game.Player{UID: "u1", Coins: 10, Version: 1}
	svc := newTestService(player, nil)

	if err := svc.GrantQuestReward(context.Background(), "u1", 102, 80); err != nil {
		t.Fatalf("GrantQuestReward: %v", err)
	}

	if player.Coins != 15 {
		t.Fatalf("expected coins 15 (+5), got %d", player.Coins)
	}
	ledgers, _ := svc.GetLedger(context.Background(), "u1")
	if len(ledgers) != 1 {
		t.Fatalf("expected 1 ledger row, got %d", len(ledgers))
	}
	if ledgers[0].Amount != CoinsPerQuestComplete || ledgers[0].Source != SourceQuestCompleted {
		t.Fatalf("unexpected ledger: %+v", ledgers[0])
	}
	if ledgers[0].RewardType != RewardTypeCoins {
		t.Fatalf("expected COINS reward type, got %s", ledgers[0].RewardType)
	}
}

func TestGrantDailyReward_AddsOneCoinAndLedger(t *testing.T) {
	player := &game.Player{UID: "u1", Coins: 3, Version: 2}
	svc := newTestService(player, nil)

	if err := svc.GrantDailyReward(context.Background(), "u1"); err != nil {
		t.Fatalf("GrantDailyReward: %v", err)
	}
	if player.Coins != 4 {
		t.Fatalf("expected coins 4 (+1), got %d", player.Coins)
	}
	ledgers, _ := svc.GetLedger(context.Background(), "u1")
	if len(ledgers) != 1 || ledgers[0].Amount != CoinsPerDailyTurn || ledgers[0].Source != SourceDailyStreak {
		t.Fatalf("unexpected ledger: %+v", ledgers)
	}
}

func TestGrantQuestReward_NoDuplicateCreditForSameQuest(t *testing.T) {
	player := &game.Player{UID: "u1", Coins: 0, Version: 1}
	meta := `{"mission_id":102}`
	ledger := &mockLedgerStore{rows: []game.RewardLedger{
		{
			ID:         "existing",
			UserID:     "u1",
			Source:     SourceQuestCompleted,
			Amount:     CoinsPerQuestComplete,
			RewardType: RewardTypeCoins,
			Metadata:   &meta,
			CreatedAt:  time.Now().UTC(),
		},
	}}
	svc := newTestService(player, ledger)

	if err := svc.GrantQuestReward(context.Background(), "u1", 102, 80); err != nil {
		t.Fatalf("GrantQuestReward: %v", err)
	}
	if player.Coins != 0 {
		t.Fatalf("expected no coin change on duplicate, got %d", player.Coins)
	}
	// still only the pre-existing row
	if len(ledger.rows) != 1 {
		t.Fatalf("expected still 1 ledger row, got %d", len(ledger.rows))
	}
}

func TestGrantQuestReward_SecondCallAfterSuccessIsIdempotent(t *testing.T) {
	player := &game.Player{UID: "u1", Coins: 0, Version: 1}
	svc := newTestService(player, nil)

	if err := svc.GrantQuestReward(context.Background(), "u1", 55, 10); err != nil {
		t.Fatalf("first grant: %v", err)
	}
	if err := svc.GrantQuestReward(context.Background(), "u1", 55, 10); err != nil {
		t.Fatalf("second grant: %v", err)
	}
	if player.Coins != CoinsPerQuestComplete {
		t.Fatalf("expected coins %d after duplicate-safe grants, got %d", CoinsPerQuestComplete, player.Coins)
	}
	ledgers, _ := svc.GetLedger(context.Background(), "u1")
	if len(ledgers) != 1 {
		t.Fatalf("expected single ledger entry, got %d", len(ledgers))
	}
}

func TestAddCoins_ConcurrentVersionMismatch(t *testing.T) {
	player := &game.Player{UID: "u1", Coins: 5, Version: 3}
	users := &mockUserStore{player: player, failMatch: true}
	ledger := &mockLedgerStore{}
	svc := NewService(&game.Repository{Users: users, RewardLedgers: ledger})

	// Force ledger success then fail coin update
	err := svc.GrantDailyReward(context.Background(), "u1")
	if err == nil {
		t.Fatal("expected concurrent modification error")
	}
	if player.Coins != 5 {
		t.Fatalf("coins should be unchanged on failed match, got %d", player.Coins)
	}
}
