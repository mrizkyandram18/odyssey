package dailyactivity

import (
	"context"
	"errors"
	"testing"

	"odyssey/pkg/game"
	"odyssey/pkg/game/progression"
)

type mockActivityStore struct {
	activities []ActivityQuestion
	completions map[string]bool
}

func (m *mockActivityStore) ListActiveActivities(ctx context.Context) ([]ActivityQuestion, error) {
	return m.activities, nil
}

func (m *mockActivityStore) HasCompletedToday(ctx context.Context, uid string, date string) (bool, error) {
	return m.completions[uid+date], nil
}

func (m *mockActivityStore) RecordCompletion(ctx context.Context, uid string, date string, activityID int64) error {
	m.completions[uid+date] = true
	return nil
}

type mockProgressionStore struct {
	game.ProgressionStore
	game.UserStore
	xp int64
}

func (m *mockProgressionStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	return &game.Player{UID: uid, XP: m.xp}, nil
}

func (m *mockProgressionStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	if xp, ok := patch["xp"]; ok {
		switch v := xp.(type) {
		case int64: m.xp = v
		case float64: m.xp = int64(v)
		case int: m.xp = int64(v)
		}
	}
	return nil
}

func (m *mockProgressionStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	m.UpdateUser(ctx, uid, patch)
	return true, nil
}

func (m *mockProgressionStore) GetPlayer(ctx context.Context, uid string) (*game.Player, error) {
	return &game.Player{UID: uid, XP: m.xp, Level: 1}, nil
}

func (m *mockProgressionStore) UpdatePlayer(ctx context.Context, uid string, patch map[string]any) error {
	if xp, ok := patch["xp"]; ok {
		m.xp = xp.(int64)
	}
	return nil
}

type mockGenericActivityStore struct {
	game.ActivityStore
	recorded int
}

func (m *mockGenericActivityStore) RecordActivity(ctx context.Context, act *game.DailyActivity) (*game.DailyActivity, error) {
	m.recorded++
	return act, nil
}

func setupTestService() (*Service, *mockActivityStore, *mockProgressionStore, *mockGenericActivityStore) {
	store := &mockActivityStore{
		activities: []ActivityQuestion{
			{ID: 1, Slug: "a1", CorrectAnswer: "A", XPReward: 10, Explanation: "E1"},
			{ID: 2, Slug: "a2", CorrectAnswer: "B", XPReward: 10, Explanation: "E2"},
			{ID: 3, Slug: "a3", CorrectAnswer: "C", XPReward: 10, Explanation: "E3"},
		},
		completions: make(map[string]bool),
	}
	progStore := &mockProgressionStore{}
	progSvc := progression.NewProgressionService(progStore, nil)
	genericActStore := &mockGenericActivityStore{}
	
	svc := NewService(store, genericActStore, progSvc, "UTC")
	return svc, store, progStore, genericActStore
}

func TestSelection_DeterministicAndWraps(t *testing.T) {
	svc, _, _, _ := setupTestService()
	ctx := context.Background()
	
	// YearDay for 2026-01-01 is 1 -> pool[0]
	// YearDay for 2026-01-02 is 2 -> pool[1]
	// YearDay for 2026-01-03 is 3 -> pool[2]
	// YearDay for 2026-01-04 is 4 -> pool[0] (wraps)
	
	act1, _ := svc.getDailyActivityFromPool(ctx, "2026-01-01")
	if act1.ID != 1 { t.Errorf("expected 1, got %d", act1.ID) }
	
	act2, _ := svc.getDailyActivityFromPool(ctx, "2026-01-02")
	if act2.ID != 2 { t.Errorf("expected 2, got %d", act2.ID) }
	
	act4, _ := svc.getDailyActivityFromPool(ctx, "2026-01-04")
	if act4.ID != 1 { t.Errorf("expected 1, got %d", act4.ID) }
}

func TestComplete_CorrectAnswer(t *testing.T) {
	svc, _, progStore, genStore := setupTestService()
	ctx := context.Background()
	uid := "user1"
	
	// Complete with correct answer
	date := svc.TodayDate()
	act, _ := svc.getDailyActivityFromPool(ctx, date)
	
	res, err := svc.CompleteActivity(ctx, uid, act.ID, act.CorrectAnswer)
	if err != nil { t.Fatal(err) }
	
	if !res.Correct || !res.Completed {
		t.Errorf("expected correct and completed")
	}
	
	if progStore.xp != 10 {
		t.Errorf("expected 10 XP, got %d", progStore.xp)
	}
	
	if genStore.recorded != 1 {
		t.Errorf("expected activity recorded for streak")
	}
	
	// duplicate
	_, err = svc.CompleteActivity(ctx, uid, act.ID, act.CorrectAnswer)
	if !errors.Is(err, ErrAlreadyCompleted) {
		t.Errorf("expected already completed, got %v", err)
	}
}

func TestComplete_IncorrectAnswer(t *testing.T) {
	svc, _, progStore, _ := setupTestService()
	ctx := context.Background()
	uid := "user2"
	
	date := svc.TodayDate()
	act, _ := svc.getDailyActivityFromPool(ctx, date)
	
	res, err := svc.CompleteActivity(ctx, uid, act.ID, "WRONG")
	if err != nil { t.Fatal(err) }
	
	if res.Correct || res.Completed {
		t.Errorf("expected incorrect and not completed")
	}
	
	if progStore.xp != 0 {
		t.Errorf("expected 0 XP, got %d", progStore.xp)
	}
	
	// retry
	res, err = svc.CompleteActivity(ctx, uid, act.ID, act.CorrectAnswer)
	if err != nil { t.Fatal(err) }
	if !res.Correct || !res.Completed {
		t.Errorf("retry should succeed")
	}
}
