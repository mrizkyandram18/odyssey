package rewardsignal

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/events"
)

type mockSignalStore struct {
	signals []*game.RewardSignal
	err     error
}

func (m *mockSignalStore) SaveSignal(ctx context.Context, signal *game.RewardSignal) error {
	if m.err != nil {
		return m.err
	}
	m.signals = append(m.signals, signal)
	return nil
}

func TestAchievementEarnedHandler(t *testing.T) {
	t.Run("successfully saves signal", func(t *testing.T) {
		store := &mockSignalStore{}
		repo := &game.Repository{
			RewardSignals: store,
		}
		svc := NewService(repo)
		handler := NewAchievementEarnedHandler(svc)

		event := events.AchievementEarnedEvent{
			UID:             "usr_123",
			FamilyID:          "crew_456",
			AchievementCode: "LEVEL_10",
		}

		err := handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(store.signals) != 1 {
			t.Fatalf("expected 1 signal, got %d", len(store.signals))
		}

		sig := store.signals[0]
		if sig.UID != "usr_123" {
			t.Errorf("expected UID usr_123, got %s", sig.UID)
		}
		if sig.AchievementCode != "LEVEL_10" {
			t.Errorf("expected AchievementCode LEVEL_10, got %s", sig.AchievementCode)
		}
		if sig.Consumed {
			t.Errorf("expected Consumed false, got true")
		}
		if time.Since(sig.IssuedAt) > time.Second {
			t.Errorf("IssuedAt is too old")
		}
	})

	t.Run("failure isolation - errors are swallowed", func(t *testing.T) {
		store := &mockSignalStore{
			err: errors.New("simulated db error"),
		}
		repo := &game.Repository{
			RewardSignals: store,
		}
		svc := NewService(repo)
		handler := NewAchievementEarnedHandler(svc)

		event := events.AchievementEarnedEvent{
			UID:             "usr_123",
			FamilyID:          "crew_456",
			AchievementCode: "LEVEL_10",
		}

		// The handler should swallow the error and return nil
		err := handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})
	
	t.Run("ignores wrong event type", func(t *testing.T) {
		store := &mockSignalStore{}
		repo := &game.Repository{
			RewardSignals: store,
		}
		svc := NewService(repo)
		handler := NewAchievementEarnedHandler(svc)

		// Pass a different event
		event := events.LevelReachedEvent{}

		err := handler.Handle(context.Background(), event)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(store.signals) != 0 {
			t.Fatalf("expected 0 signals, got %d", len(store.signals))
		}
	})
}
