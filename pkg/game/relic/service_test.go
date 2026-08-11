package relic_test

import (
	"context"
	"errors"
	"testing"

	"odyssey/pkg/game"
	"odyssey/pkg/game/relic"
)

type mockUserStore struct {
	game.UserStore
	users map[string]*game.Player
}

func (m *mockUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	u, ok := m.users[uid]
	if !ok {
		return nil, game.ErrNotFound
	}
	return u, nil
}

type mockLedgerStore struct {
	game.RewardLedgerStore
	ledgers []*game.RewardLedger
	failOn  string // if non-empty, fail when UserID == failOn
}

func (m *mockLedgerStore) CreateLedger(ctx context.Context, ledger *game.RewardLedger) error {
	if m.failOn != "" && ledger.UserID == m.failOn {
		return errors.New("ledger store error")
	}
	m.ledgers = append(m.ledgers, ledger)
	return nil
}

type mockPlayerRelicStore struct {
	game.PlayerRelicStore
	relics     map[string]*game.PlayerRelic
	failUpdate string // uid to fail update on
	failCreate bool
}

func (m *mockPlayerRelicStore) GetPlayerRelic(ctx context.Context, uid, relicSlug string) (*game.PlayerRelic, error) {
	r, ok := m.relics[uid+"_"+relicSlug]
	if !ok {
		return nil, game.ErrNotFound
	}
	return r, nil
}

func (m *mockPlayerRelicStore) CreatePlayerRelic(ctx context.Context, item *game.PlayerRelic) (*game.PlayerRelic, error) {
	if m.failCreate {
		return nil, errors.New("create relic failed")
	}
	m.relics[item.UID+"_"+item.RelicSlug] = item
	return item, nil
}

func (m *mockPlayerRelicStore) UpdatePlayerRelic(ctx context.Context, uid, relicSlug string, patch map[string]any) error {
	if m.failUpdate == uid {
		return errors.New("update player relic failed")
	}
	r, ok := m.relics[uid+"_"+relicSlug]
	if !ok {
		return game.ErrNotFound
	}
	if v, ok := patch["owned_count"].(int); ok {
		r.OwnedCount = v
	}
	return nil
}

// newSvc builds a RelicService with the given stores and the default (hardcoded) catalog.
func newSvc(prStore game.PlayerRelicStore, uStore game.UserStore, lStore game.RewardLedgerStore) *relic.RelicService {
	svc := relic.NewRelicService(nil, prStore, nil)
	svc.SetUserStore(uStore)
	svc.SetLedgerStore(lStore)
	return svc
}

// defaultStores returns standard happy-path stores reset for each sub-test.
func defaultStores() (*mockPlayerRelicStore, *mockUserStore, *mockLedgerStore) {
	prStore := &mockPlayerRelicStore{
		relics: map[string]*game.PlayerRelic{
			"sender1_ancient-compass": {UID: "sender1", RelicSlug: "ancient-compass", OwnedCount: 2},
			"sender1_crystal-shard":   {UID: "sender1", RelicSlug: "crystal-shard", OwnedCount: 0},
		},
	}
	uStore := &mockUserStore{
		users: map[string]*game.Player{
			"sender1": {UID: "sender1", CrewID: "crewA", ExplorerName: "Sender"},
			"recv1":   {UID: "recv1", CrewID: "crewA", ExplorerName: "Recv"},
			"recvB":   {UID: "recvB", CrewID: "crewB", ExplorerName: "Recv B"},
		},
	}
	lStore := &mockLedgerStore{}
	return prStore, uStore, lStore
}

func TestGiftRelic(t *testing.T) {
	ctx := context.Background()

	t.Run("HappyPath", func(t *testing.T) {
		prStore, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)

		res, err := svc.GiftRelic(ctx, "sender1", "recv1", "ancient-compass", "crewA")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.SenderCount != 1 {
			t.Errorf("expected SenderCount=1, got %d", res.SenderCount)
		}
		if prStore.relics["sender1_ancient-compass"].OwnedCount != 1 {
			t.Errorf("sender store count not decremented")
		}
		if prStore.relics["recv1_ancient-compass"].OwnedCount != 1 {
			t.Errorf("recipient store count not incremented")
		}
		// Two ledger entries: SENT + RECEIVED
		if len(lStore.ledgers) != 2 {
			t.Errorf("expected 2 ledger entries, got %d", len(lStore.ledgers))
		}
		// Economy invariants: Amount=0 on all ledger entries (no coins/XP)
		for _, l := range lStore.ledgers {
			if l.Amount != 0 {
				t.Errorf("ledger Amount must be 0 (no coins/XP), got %d", l.Amount)
			}
			if l.RewardType == "COINS" || l.RewardType == "XP" {
				t.Errorf("ledger RewardType must not be COINS or XP, got %s", l.RewardType)
			}
		}
	})

	t.Run("SelfGift", func(t *testing.T) {
		prStore, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)
		_, err := svc.GiftRelic(ctx, "sender1", "sender1", "ancient-compass", "crewA")
		if err != relic.ErrSelfGift {
			t.Errorf("expected ErrSelfGift, got %v", err)
		}
		// No store mutations on self-gift
		if prStore.relics["sender1_ancient-compass"].OwnedCount != 2 {
			t.Errorf("sender count must not change on self-gift error")
		}
		if len(lStore.ledgers) != 0 {
			t.Errorf("no ledger entries expected on self-gift error")
		}
	})

	t.Run("CrossCrew", func(t *testing.T) {
		prStore, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)
		_, err := svc.GiftRelic(ctx, "sender1", "recvB", "ancient-compass", "crewA")
		if err != relic.ErrCrossCrewGift {
			t.Errorf("expected ErrCrossCrewGift, got %v", err)
		}
		if prStore.relics["sender1_ancient-compass"].OwnedCount != 2 {
			t.Errorf("sender count must not change on cross-crew error")
		}
		if len(lStore.ledgers) != 0 {
			t.Errorf("no ledger entries expected on cross-crew error")
		}
	})

	t.Run("ZeroCount", func(t *testing.T) {
		prStore, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)
		_, err := svc.GiftRelic(ctx, "sender1", "recv1", "crystal-shard", "crewA")
		if err != relic.ErrRelicNotOwned {
			t.Errorf("expected ErrRelicNotOwned, got %v", err)
		}
		if len(lStore.ledgers) != 0 {
			t.Errorf("no ledger entries expected when relic not owned")
		}
	})

	t.Run("RelicNotFound", func(t *testing.T) {
		prStore, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)
		_, err := svc.GiftRelic(ctx, "sender1", "recv1", "unknown-slug", "crewA")
		if err != relic.ErrRelicNotFound {
			t.Errorf("expected ErrRelicNotFound, got %v", err)
		}
	})

	t.Run("RecipientNotFound", func(t *testing.T) {
		prStore, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)
		_, err := svc.GiftRelic(ctx, "sender1", "nobody", "ancient-compass", "crewA")
		if err != relic.ErrRecipientNotFound {
			t.Errorf("expected ErrRecipientNotFound, got %v", err)
		}
		// Sender count must not be decremented
		if prStore.relics["sender1_ancient-compass"].OwnedCount != 2 {
			t.Errorf("sender count must not change when recipient not found")
		}
		if len(lStore.ledgers) != 0 {
			t.Errorf("no ledger entries expected when recipient not found")
		}
	})

	t.Run("SenderDecrementFailure_NoGiftResult", func(t *testing.T) {
		// Simulate sender UpdatePlayerRelic failing.
		// Expected: error returned, recipient NOT credited, no ledger entries.
		prStore := &mockPlayerRelicStore{
			relics: map[string]*game.PlayerRelic{
				"sender1_ancient-compass": {UID: "sender1", RelicSlug: "ancient-compass", OwnedCount: 2},
			},
			failUpdate: "sender1",
		}
		_, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)

		_, err := svc.GiftRelic(ctx, "sender1", "recv1", "ancient-compass", "crewA")
		if err == nil {
			t.Fatal("expected error when sender decrement fails, got nil")
		}
		// Recipient must NOT have been credited
		if _, ok := prStore.relics["recv1_ancient-compass"]; ok {
			t.Errorf("recipient must not be credited when sender decrement fails")
		}
		// No ledger entries
		if len(lStore.ledgers) != 0 {
			t.Errorf("no ledger entries expected when sender decrement fails, got %d", len(lStore.ledgers))
		}
	})

	t.Run("RecipientCreditFailure_Error", func(t *testing.T) {
		// Simulate recipient CreatePlayerRelic failing (recipient has no existing row).
		// Expected: error returned, no successful gift response.
		prStore := &mockPlayerRelicStore{
			relics: map[string]*game.PlayerRelic{
				"sender1_ancient-compass": {UID: "sender1", RelicSlug: "ancient-compass", OwnedCount: 2},
			},
			failCreate: true,
		}
		_, uStore, lStore := defaultStores()
		svc := newSvc(prStore, uStore, lStore)

		_, err := svc.GiftRelic(ctx, "sender1", "recv1", "ancient-compass", "crewA")
		if err == nil {
			t.Fatal("expected error when recipient credit fails, got nil")
		}
		// No ledger entries
		if len(lStore.ledgers) != 0 {
			t.Errorf("no ledger entries expected when recipient credit fails, got %d", len(lStore.ledgers))
		}
	})

	t.Run("LedgerFailure_BestEffort_GiftSucceeds", func(t *testing.T) {
		// Simulate ledger store failing. Per architecture: ledger is best-effort (errors discarded with _).
		// Gift itself must STILL succeed (inventory was already mutated).
		prStore, uStore, _ := defaultStores()
		lStore := &mockLedgerStore{failOn: "sender1"} // fail the SENT entry
		svc := newSvc(prStore, uStore, lStore)

		res, err := svc.GiftRelic(ctx, "sender1", "recv1", "ancient-compass", "crewA")
		if err != nil {
			t.Fatalf("gift should succeed even if ledger fails (best-effort): %v", err)
		}
		if res.SenderCount != 1 {
			t.Errorf("expected SenderCount=1, got %d", res.SenderCount)
		}
	})
}

// TestGiftRelic_EconomyInvariants verifies no coin/XP/role/level mutations occur.
func TestGiftRelic_EconomyInvariants(t *testing.T) {
	ctx := context.Background()
	prStore, uStore, lStore := defaultStores()
	svc := newSvc(prStore, uStore, lStore)

	// Capture user state before
	senderBefore := *uStore.users["sender1"]
	recvBefore := *uStore.users["recv1"]

	_, err := svc.GiftRelic(ctx, "sender1", "recv1", "ancient-compass", "crewA")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Users must be unchanged (GiftRelic must not call UpdateUser)
	senderAfter := *uStore.users["sender1"]
	recvAfter := *uStore.users["recv1"]

	if senderAfter.Coins != senderBefore.Coins {
		t.Errorf("sender coins changed: %d → %d", senderBefore.Coins, senderAfter.Coins)
	}
	if senderAfter.XP != senderBefore.XP {
		t.Errorf("sender XP changed: %d → %d", senderBefore.XP, senderAfter.XP)
	}
	if senderAfter.Role != senderBefore.Role {
		t.Errorf("sender role changed: %s → %s", senderBefore.Role, senderAfter.Role)
	}
	if senderAfter.Level != senderBefore.Level {
		t.Errorf("sender level changed: %d → %d", senderBefore.Level, senderAfter.Level)
	}
	if recvAfter.Coins != recvBefore.Coins {
		t.Errorf("recipient coins changed: %d → %d", recvBefore.Coins, recvAfter.Coins)
	}
	if recvAfter.XP != recvBefore.XP {
		t.Errorf("recipient XP changed: %d → %d", recvBefore.XP, recvAfter.XP)
	}
	if recvAfter.Role != recvBefore.Role {
		t.Errorf("recipient role changed: %s → %s", recvBefore.Role, recvAfter.Role)
	}
	if recvAfter.Level != recvBefore.Level {
		t.Errorf("recipient level changed: %d → %d", recvBefore.Level, recvAfter.Level)
	}

	// All ledger entries must have Amount=0 and non-economy RewardType
	for _, l := range lStore.ledgers {
		if l.Amount != 0 {
			t.Errorf("ledger Amount must be 0, got %d for source=%s", l.Amount, l.Source)
		}
		if l.RewardType == "COINS" || l.RewardType == "XP" {
			t.Errorf("ledger RewardType must not be COINS or XP, got %s", l.RewardType)
		}
	}
}
