package progression

import (
	"context"
	"testing"

	"odyssey/pkg/game"
	"odyssey/pkg/observability"
)

// conflictUserStore always reports a version mismatch, exercising the
// optimistic-lock conflict path in AwardXP.
type conflictUserStore struct{}

func (conflictUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	return makePlayer(1, 0), nil
}
func (conflictUserStore) CreateUser(ctx context.Context, p *game.Player) error { return nil }
func (conflictUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return nil
}
func (conflictUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	return false, nil // always mismatch -> force re-read / conflict path
}

func TestProgression_RecordXP(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0)}
	m := observability.NewMetrics()
	svc := NewProgressionService(store, nil)
	svc.SetMetrics(m)

	store.player.Version = 1
	if _, _, err := svc.AwardXP(context.Background(), "user-1", 60); err != nil {
		t.Fatalf("AwardXP error: %v", err)
	}
	if m.Snapshot().XPAwarded != 60 {
		t.Errorf("expected XP awarded 60, got %d", m.Snapshot().XPAwarded)
	}
}

func TestProgression_RecordLockConflict(t *testing.T) {
	m := observability.NewMetrics()
	svc := NewProgressionService(conflictUserStore{}, nil)
	svc.SetMetrics(m)

	// Force a version mismatch via the conflict-only store. No XP should be
	// recorded and a lock conflict should be counted.
	if _, _, err := svc.AwardXP(context.Background(), "user-1", 10); err != nil {
		t.Fatalf("AwardXP error: %v", err)
	}
	snap := m.Snapshot()
	if snap.XPAwarded != 0 {
		t.Errorf("expected 0 XP awarded on conflict, got %d", snap.XPAwarded)
	}
	if snap.LockConflicts != 1 {
		t.Errorf("expected 1 lock conflict, got %d", snap.LockConflicts)
	}
}

func TestProgression_SetMetricsNilSafe(t *testing.T) {
	svc := NewProgressionService(&mockUserStore{player: makePlayer(1, 0)}, nil)
	svc.SetMetrics(nil)
	store := svc.users.(*mockUserStore)
	store.player.Version = 1
	if _, _, err := svc.AwardXP(context.Background(), "user-1", 10); err != nil {
		t.Fatalf("AwardXP error: %v", err)
	}
}
func (m conflictUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}
