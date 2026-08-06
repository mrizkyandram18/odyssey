package quest

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockQuestStore struct {
	quests       map[int64]*game.Quest
	questsByCrew map[string][]game.Quest
	challenges   map[int64][]game.Challenge
	err          error
}

func newMockQuestStore() *mockQuestStore {
	return &mockQuestStore{
		quests:       make(map[int64]*game.Quest),
		questsByCrew: make(map[string][]game.Quest),
		challenges:   make(map[int64][]game.Challenge),
	}
}

func (m *mockQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	q, ok := m.quests[questID]
	if !ok {
		return nil, game.ErrNotFound
	}
	return q, nil
}

func (m *mockQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	q.ID = int64(len(m.quests) + 1)
	m.quests[q.ID] = q
	m.questsByCrew[q.CrewID] = append(m.questsByCrew[q.CrewID], *q)
	return q, nil
}

func (m *mockQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.questsByCrew[crewID], nil
}

func (m *mockQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	q, ok := m.quests[questID]
	if !ok {
		return game.ErrNotFound
	}
	if status, ok := patch["status"]; ok {
		q.Status = status.(string)
	}
	if startedAt, ok := patch["started_at"]; ok {
		if t, ok := startedAt.(*time.Time); ok {
			q.StartedAt = t
		}
	}
	if completedAt, ok := patch["completed_at"]; ok {
		if t, ok := completedAt.(*time.Time); ok {
			q.CompletedAt = t
		}
	}
	return nil
}

func (m *mockQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	q, ok := m.quests[questID]
	if !ok {
		return false, game.ErrNotFound
	}
	if q.Status != oldStatus {
		return false, nil
	}
	if err := m.UpdateQuest(ctx, questID, patch); err != nil {
		return false, err
	}
	return true, nil
}

func (m *mockQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.challenges[questID], nil
}

func (m *mockQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	if m.err != nil {
		return nil, m.err
	}
	c.ID = int64(len(m.challenges) + 1)
	m.challenges[c.QuestID] = append(m.challenges[c.QuestID], *c)
	return c, nil
}

func (m *mockQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	for _, chs := range m.challenges {
		for i := range chs {
			if chs[i].ID == challengeID {
				if status, ok := patch["status"]; ok {
					chs[i].Status = status.(string)
				}
				if completedBy, ok := patch["completed_by"]; ok {
					chs[i].CompletedBy = completedBy.(string)
				}
				if completedAt, ok := patch["completed_at"]; ok {
					if t, ok := completedAt.(*time.Time); ok {
						chs[i].CompletedAt = t
					}
				}
				return nil
			}
		}
	}
	return game.ErrNotFound
}

func makeQuest(status string) *game.Quest {
	return &game.Quest{
		ID:           1,
		CrewID:       "crew-1",
		TemplateSlug: "test-quest",
		Title:        "Test Quest",
		Status:       status,
		CreatedAt:    time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func makeChallenge(id int64, status string) game.Challenge {
	return game.Challenge{
		ID:          id,
		QuestID:     1,
		Slug:        "challenge-1",
		Description: "Do something",
		Status:      status,
		CreatedAt:   time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func TestComputeStatus_AllPending(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	ch := []game.Challenge{
		makeChallenge(1, string(ChallengeStatusPending)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	if got := svc.ComputeStatus(ch); got != QuestStatusPending {
		t.Errorf("expected %s, got %s", QuestStatusPending, got)
	}
}

func TestComputeStatus_SomeDone(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	ch := []game.Challenge{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	if got := svc.ComputeStatus(ch); got != QuestStatusActive {
		t.Errorf("expected %s, got %s", QuestStatusActive, got)
	}
}

func TestComputeStatus_AllDone(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	ch := []game.Challenge{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusDone)),
	}
	if got := svc.ComputeStatus(ch); got != QuestStatusDone {
		t.Errorf("expected %s, got %s", QuestStatusDone, got)
	}
}

func TestComputeStatus_Empty(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	if got := svc.ComputeStatus(nil); got != QuestStatusPending {
		t.Errorf("expected %s for empty, got %s", QuestStatusPending, got)
	}
}

func TestComputeStatus_NoChallenges(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	if got := svc.ComputeStatus([]game.Challenge{}); got != QuestStatusPending {
		t.Errorf("expected %s for empty slice, got %s", QuestStatusPending, got)
	}
}

func TestActiveQuests_FiltersDone(t *testing.T) {
	quests := []game.Quest{
		{Status: string(QuestStatusPending)},
		{Status: string(QuestStatusActive)},
		{Status: string(QuestStatusDone)},
	}
	active := ActiveQuests(quests)
	if len(active) != 2 {
		t.Fatalf("expected 2 active quests, got %d", len(active))
	}
}

func TestActiveQuests_Empty(t *testing.T) {
	active := ActiveQuests(nil)
	if len(active) != 0 {
		t.Fatalf("expected 0 active quests, got %d", len(active))
	}
}

func TestListByCrew_ReturnsQuests(t *testing.T) {
	store := newMockQuestStore()
	store.questsByCrew["crew-1"] = []game.Quest{
		{ID: 1, Title: "Quest A", CreatedAt: time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Quest B", CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)},
	}
	svc := NewQuestService(store)
	quests, err := svc.ListByCrew(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(quests) != 2 {
		t.Fatalf("expected 2 quests, got %d", len(quests))
	}
	if quests[0].Title != "Quest B" {
		t.Errorf("expected newest quest first, got %s", quests[0].Title)
	}
}

func TestListByCrew_Empty(t *testing.T) {
	store := newMockQuestStore()
	svc := NewQuestService(store)
	quests, err := svc.ListByCrew(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(quests) != 0 {
		t.Fatalf("expected 0 quests, got %d", len(quests))
	}
}

func TestGetByCrewAndID_Success(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusActive))
	store.challenges[1] = []game.Challenge{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	svc := NewQuestService(store)

	result, err := svc.GetByCrewAndID(context.Background(), 1, "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected quest ID 1, got %d", result.ID)
	}
	if len(result.Challenges) != 2 {
		t.Fatalf("expected 2 challenges, got %d", len(result.Challenges))
	}
}

func TestGetByCrewAndID_NotFound(t *testing.T) {
	store := newMockQuestStore()
	svc := NewQuestService(store)
	_, err := svc.GetByCrewAndID(context.Background(), 999, "crew-1")
	if !errors.Is(err, ErrQuestNotFound) {
		t.Fatalf("expected ErrQuestNotFound, got %v", err)
	}
}

func TestGetByCrewAndID_NotInCrew(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusActive))
	svc := NewQuestService(store)
	_, err := svc.GetByCrewAndID(context.Background(), 1, "other-crew")
	if !errors.Is(err, ErrQuestNotInCrew) {
		t.Fatalf("expected ErrQuestNotInCrew, got %v", err)
	}
}

func TestGetByCrewAndID_ComputesStatus(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusPending))
	store.challenges[1] = []game.Challenge{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusDone)),
	}
	svc := NewQuestService(store)
	result, err := svc.GetByCrewAndID(context.Background(), 1, "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != string(QuestStatusDone) {
		t.Errorf("expected status %s, got %s", QuestStatusDone, result.Status)
	}
}

func TestCompleteChallengeForQuest_TransitionsToActive(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusPending))
	c1 := makeChallenge(1, string(ChallengeStatusPending))
	c2 := makeChallenge(2, string(ChallengeStatusPending))
	store.challenges[1] = []game.Challenge{c1, c2}
	svc := NewQuestService(store)

	status, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 1, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != QuestStatusActive {
		t.Errorf("expected status %s, got %s", QuestStatusActive, status)
	}
	if completed {
		t.Error("expected quest not yet completed")
	}
	if store.quests[1].Status != string(QuestStatusActive) {
		t.Errorf("expected stored status %s, got %s", QuestStatusActive, store.quests[1].Status)
	}
}

func TestCompleteChallengeForQuest_TransitionsToDone(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusActive))
	c1 := makeChallenge(1, string(ChallengeStatusDone))
	c2 := makeChallenge(2, string(ChallengeStatusPending))
	store.challenges[1] = []game.Challenge{c1, c2}
	svc := NewQuestService(store)

	status, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 2, "user-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if status != QuestStatusDone {
		t.Errorf("expected status %s, got %s", QuestStatusDone, status)
	}
	if !completed {
		t.Error("expected quest to be newly completed")
	}
	if store.quests[1].Status != string(QuestStatusDone) {
		t.Errorf("expected stored status %s, got %s", QuestStatusDone, store.quests[1].Status)
	}
	if store.quests[1].CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestCompleteChallengeForQuest_AlreadyDoneNoReReward(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusDone))
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	store.quests[1].CompletedAt = &now
	c1 := makeChallenge(1, string(ChallengeStatusDone))
	c2 := makeChallenge(2, string(ChallengeStatusDone))
	store.challenges[1] = []game.Challenge{c1, c2}
	svc := NewQuestService(store)

	status, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 1, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != QuestStatusDone {
		t.Errorf("expected status %s, got %s", QuestStatusDone, status)
	}
	if completed {
		t.Error("expected completed=false when quest was already DONE")
	}
}

func TestStartQuest(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusPending))
	svc := NewQuestService(store)

	err := svc.StartQuest(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.quests[1].Status != string(QuestStatusActive) {
		t.Errorf("expected status %s, got %s", QuestStatusActive, store.quests[1].Status)
	}
	if store.quests[1].StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}
