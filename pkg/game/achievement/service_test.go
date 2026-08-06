package achievement

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
)

type mockAchievementGateway struct {
	defs   []gamecontent.AchievementDefinition
	err    error
	getErr error
}

func (m *mockAchievementGateway) ListAchievements(ctx context.Context) ([]gamecontent.AchievementDefinition, error) {
	return m.defs, m.err
}
func (m *mockAchievementGateway) GetAchievement(ctx context.Context, code string) (*gamecontent.AchievementDefinition, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, d := range m.defs {
		if d.Code == code {
			return &d, nil
		}
	}
	return nil, nil
}

type mockAchievementStore struct {
	earned  map[string]*game.Achievement
	err     error
	created []*game.Achievement
}

func newMockAchievementStore() *mockAchievementStore {
	return &mockAchievementStore{earned: make(map[string]*game.Achievement)}
}

func (m *mockAchievementStore) CreateAchievement(ctx context.Context, a *game.Achievement) (*game.Achievement, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := a.UID + "|" + a.Code
	m.earned[key] = a
	m.created = append(m.created, a)
	return a, nil
}
func (m *mockAchievementStore) GetAchievementByCode(ctx context.Context, uid, code string) (*game.Achievement, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := uid + "|" + code
	a, ok := m.earned[key]
	if !ok {
		return nil, game.ErrNotFound
	}
	return a, nil
}
func (m *mockAchievementStore) ListAchievementsByPlayer(ctx context.Context, uid string) ([]game.Achievement, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]game.Achievement, 0)
	for key, a := range m.earned {
		if len(key) > len(uid) && key[:len(uid)] == uid {
			result = append(result, *a)
		}
	}
	return result, nil
}
func (m *mockAchievementStore) CountAchievementsByKind(ctx context.Context, uid string, kind string) (int, error) {
	return 0, m.err
}

type mockProgressReader struct {
	completedQuests     int
	completedRealms     int
	completedChapters   int
	collectedRelics     int
	dailyStreak         int
	creativeSubmissions int
	playerLevel         int
	err                 error
}

func (m *mockProgressReader) CountCompletedQuests(ctx context.Context, crewID string) (int, error) {
	return m.completedQuests, m.err
}
func (m *mockProgressReader) CountCompletedRealms(ctx context.Context, crewID string) (int, error) {
	return m.completedRealms, m.err
}
func (m *mockProgressReader) CountCompletedChapters(ctx context.Context, crewID string) (int, error) {
	return m.completedChapters, m.err
}
func (m *mockProgressReader) CountCollectedRelics(ctx context.Context, uid string) (int, error) {
	return m.collectedRelics, m.err
}
func (m *mockProgressReader) CountDailyStreak(ctx context.Context, uid string) (int, error) {
	return m.dailyStreak, m.err
}
func (m *mockProgressReader) CountCreativeSubmissions(ctx context.Context, crewID string) (int, error) {
	return m.creativeSubmissions, m.err
}
func (m *mockProgressReader) GetPlayerLevel(ctx context.Context, uid string) (int, error) {
	return m.playerLevel, m.err
}

func makeAchievementDef(code, kind, trigger string, threshold int) gamecontent.AchievementDefinition {
	return gamecontent.AchievementDefinition{
		Code:        code,
		Title:       "Test " + code,
		Description: "desc",
		Kind:        kind,
		Trigger:     trigger,
		Threshold:   threshold,
		RewardXP:    0,
		CreatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestListByPlayer_ShowsProgressAndUnlocked(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_QUESTS_10", "PERSONAL", "QUEST_COMPLETED", 10),
			makeAchievementDef("ACH_LEVEL_5", "PERSONAL", "LEVEL_REACHED", 5),
		},
	}
	store := newMockAchievementStore()
	now := time.Now().UTC()
	store.earned["u1|ACH_QUESTS_10"] = &game.Achievement{
		UID: "u1", CrewID: "c1", Code: "ACH_QUESTS_10", Kind: "PERSONAL",
		Trigger: "QUEST_COMPLETED", AwardedAt: now,
	}
	rdr := &mockProgressReader{
		completedQuests: 7,
		playerLevel:     3,
	}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})

	result, err := svc.ListByPlayer(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 achievements, got %d", len(result))
	}

	a1 := result[0]
	if a1.Code != "ACH_QUESTS_10" {
		t.Errorf("expected ACH_QUESTS_10, got %s", a1.Code)
	}
	if a1.Progress != 7 {
		t.Errorf("expected progress 7, got %d", a1.Progress)
	}
	if !a1.Unlocked {
		t.Error("expected unlocked")
	}

	a2 := result[1]
	if a2.Code != "ACH_LEVEL_5" {
		t.Errorf("expected ACH_LEVEL_5, got %s", a2.Code)
	}
	if a2.Progress != 3 {
		t.Errorf("expected progress 3, got %d", a2.Progress)
	}
	if a2.Unlocked {
		t.Error("expected not unlocked")
	}
}

func TestEvaluate_AwardsAchievementWhenThresholdMet(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_QUESTS_5", "PERSONAL", "QUEST_COMPLETED", 5),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{completedQuests: 5}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})

	err := svc.evaluate(context.Background(), TriggerQuestCompleted, "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 achievement created, got %d", len(store.created))
	}
	if store.created[0].Code != "ACH_QUESTS_5" {
		t.Errorf("expected ACH_QUESTS_5, got %s", store.created[0].Code)
	}
}

func TestEvaluate_DoesNotAwardWhenAlreadyEarned(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_QUESTS_5", "PERSONAL", "QUEST_COMPLETED", 5),
		},
	}
	store := newMockAchievementStore()
	store.earned["u1|ACH_QUESTS_5"] = &game.Achievement{
		UID: "u1", Code: "ACH_QUESTS_5",
	}
	rdr := &mockProgressReader{completedQuests: 10}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})

	err := svc.evaluate(context.Background(), TriggerQuestCompleted, "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 0 {
		t.Errorf("expected 0 new achievements, got %d", len(store.created))
	}
}

func TestEvaluate_DoesNotAwardWhenBelowThreshold(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_QUESTS_5", "PERSONAL", "QUEST_COMPLETED", 5),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{completedQuests: 3}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})

	err := svc.evaluate(context.Background(), TriggerQuestCompleted, "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 0 {
		t.Errorf("expected 0 achievements below threshold, got %d", len(store.created))
	}
}

func TestEvaluate_FiltersByTrigger(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_QUESTS_5", "PERSONAL", "QUEST_COMPLETED", 5),
			makeAchievementDef("ACH_LEVEL_5", "PERSONAL", "LEVEL_REACHED", 5),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{
		completedQuests: 10,
		playerLevel:     10,
	}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})

	err := svc.evaluate(context.Background(), TriggerQuestCompleted, "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Fatalf("expected 1 achievement for QUEST_COMPLETED trigger, got %d", len(store.created))
	}
	if store.created[0].Code != "ACH_QUESTS_5" {
		t.Errorf("expected ACH_QUESTS_5, got %s", store.created[0].Code)
	}
}

func TestQuestCompletedHandler_TriggersEvaluation(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_QUESTS_1", "PERSONAL", "QUEST_COMPLETED", 1),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{completedQuests: 1}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})
	handler := NewQuestCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.QuestCompletedEvent{
		CrewID:    "c1",
		PlayerUID: "u1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Errorf("expected 1 achievement created, got %d", len(store.created))
	}
}

func TestChapterCompletedHandler_TriggersEvaluation(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_CHAPTER_1", "GROUP", "CHAPTER_COMPLETED", 1),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{completedChapters: 1}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})
	handler := NewChapterCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.ChapterCompletedEvent{
		CrewID:    "c1",
		PlayerUID: "u1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Errorf("expected 1 achievement created, got %d", len(store.created))
	}
}

func TestRelicCollectedHandler_TriggersEvaluation(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_RELIC_1", "PERSONAL", "RELIC_COLLECTED", 1),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{collectedRelics: 1}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})
	handler := NewRelicCollectedHandler(svc)

	err := handler.Handle(context.Background(), events.RelicCollectedEvent{
		UID:    "u1",
		CrewID: "c1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Errorf("expected 1 achievement created, got %d", len(store.created))
	}
}

func TestLevelReachedHandler_TriggersEvaluation(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_LEVEL_3", "PERSONAL", "LEVEL_REACHED", 3),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{playerLevel: 3}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})
	handler := NewLevelReachedHandler(svc)

	err := handler.Handle(context.Background(), events.LevelReachedEvent{
		UID:      "u1",
		CrewID:   "c1",
		OldLevel: 2,
		NewLevel: 3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Errorf("expected 1 achievement created, got %d", len(store.created))
	}
}

func TestDailyTurnCompletedHandler_TriggersEvaluation(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_STREAK_7", "PERSONAL", "DAILY_STREAK", 7),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{dailyStreak: 7}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})
	handler := NewDailyTurnCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.DailyTurnCompletedEvent{
		UID: "u1", StreakDays: 7,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Errorf("expected 1 achievement created, got %d", len(store.created))
	}
}

func TestCreativeSubmissionHandler_TriggersEvaluation(t *testing.T) {
	gw := &mockAchievementGateway{
		defs: []gamecontent.AchievementDefinition{
			makeAchievementDef("ACH_SUBMIT_3", "GROUP", "CREATIVE_SUBMISSION", 3),
		},
	}
	store := newMockAchievementStore()
	rdr := &mockProgressReader{creativeSubmissions: 3}
	svc := NewAchievementService(gw, store, rdr, events.NopPublisher{})
	handler := NewCreativeSubmissionHandler(svc)

	err := handler.Handle(context.Background(), events.CreativeSubmissionEvent{
		UID: "u1", CrewID: "c1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(store.created) != 1 {
		t.Errorf("expected 1 achievement created, got %d", len(store.created))
	}
}

func TestHandler_OtherEventIgnored(t *testing.T) {
	svc := NewAchievementService(&mockAchievementGateway{}, newMockAchievementStore(), &mockProgressReader{}, events.NopPublisher{})
	handler := NewQuestCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.ChapterCompletedEvent{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(handler.svc.store.(*mockAchievementStore).created) != 0 {
		t.Error("expected no achievements for wrong event type")
	}
}
