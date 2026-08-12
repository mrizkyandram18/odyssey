package chest

import (
	"context"
	"testing"

	"odyssey/pkg/game"
	"odyssey/pkg/observability"
)

func TestChestService_MetricsRecorded(t *testing.T) {
	store := newConcurrentChestStore()
	relicStore := &concurrentRelicStore{}
	playerRelicStore := newConcurrentPlayerRelicStore()
	userStore := &concurrentUserStore{player: &game.Player{UID: "uid1", FamilyID: "crew1", Level: 1, XP: 0, Version: 1}}
	engine := NewRewardEngine(nil, nil)
	svc := NewChestService(store, playerRelicStore, relicStore, userStore, engine)
	m := observability.NewMetrics()
	svc.SetMetrics(m)

	// Create a chest for the quest source.
	ch, err := svc.CreateChest(context.Background(), "uid1", "wooden-chest", "QUEST", "")
	if err != nil {
		t.Fatalf("create chest: %v", err)
	}
	if m.Snapshot().ChestsCreated != 1 {
		t.Errorf("expected 1 chest created, got %d", m.Snapshot().ChestsCreated)
	}

	// Opening grants 1 reward (common chest, 1 reward).
	result, err := svc.OpenChest(context.Background(), ch.ID, "uid1")
	if err != nil {
		t.Fatalf("open chest: %v", err)
	}
	if len(result.Rewards) == 0 {
		t.Fatal("expected at least 1 reward")
	}
	if m.Snapshot().RewardsGenerated < 1 {
		t.Errorf("expected rewards generated >= 1, got %d", m.Snapshot().RewardsGenerated)
	}

	// Duplicate open returns "already opened" and records replay-ignored.
	_, err = svc.OpenChest(context.Background(), ch.ID, "uid1")
	if err == nil {
		t.Fatal("expected error on duplicate open")
	}
	if m.Snapshot().ReplayIgnored != 1 {
		t.Errorf("expected 1 replay ignored, got %d", m.Snapshot().ReplayIgnored)
	}
}

func TestChestService_SetMetricsNilSafe(t *testing.T) {
	store := newConcurrentChestStore()
	svc := NewChestService(store, newConcurrentPlayerRelicStore(), &concurrentRelicStore{}, &concurrentUserStore{}, NewRewardEngine(nil, nil))
	svc.SetMetrics(nil)
	_, err := svc.CreateChest(context.Background(), "user-1", "wooden-chest", "TEST", "")
	if err != nil {
		t.Fatalf("create chest: %v", err)
	}
}
