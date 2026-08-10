package quest

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockQuestStore struct {
	quests             map[int64]*game.Quest
	questsByCrew       map[string][]game.Quest
	challenges         map[int64][]game.Challenge
	err                error
	getChallengeCalls  int
	listChallengeCalls int
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
	// Serve live quest rows so status patches applied via UpdateQuest are visible.
	if len(m.quests) > 0 {
		out := make([]game.Quest, 0, len(m.quests))
		for _, q := range m.quests {
			if q != nil && q.CrewID == crewID {
				out = append(out, *q)
			}
		}
		if len(out) > 0 {
			return out, nil
		}
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
	m.getChallengeCalls++
	return m.challenges[questID], nil
}

func (m *mockQuestStore) ListChallengesByQuestIDs(ctx context.Context, questIDs []int64) ([]game.Challenge, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.listChallengeCalls++
	var out []game.Challenge
	for _, id := range questIDs {
		out = append(out, m.challenges[id]...)
	}
	return out, nil
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

func (m *mockQuestStore) UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	for _, chs := range m.challenges {
		for i := range chs {
			if chs[i].ID == challengeID {
				if chs[i].Status != oldStatus {
					return false, nil
				}
				if err := m.UpdateChallenge(ctx, challengeID, patch); err != nil {
					return false, err
				}
				return true, nil
			}
		}
	}
	return false, game.ErrNotFound
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

func TestGetByCrewAndID_ExposesActiveChallengeAssignee(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusActive))
	store.challenges[1] = []game.Challenge{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	assigned := "user-2"
	store.challenges[1][1].AssignedTo = &assigned
	svc := NewQuestService(store)

	result, err := svc.GetByCrewAndID(context.Background(), 1, "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ActiveChallengeAssignedTo == nil {
		t.Fatal("expected active_challenge_assigned_to to be set")
	}
	if *result.ActiveChallengeAssignedTo != "user-2" {
		t.Errorf("expected assignee user-2, got %s", *result.ActiveChallengeAssignedTo)
	}
}

func TestGetByCrewAndID_AllDoneHasNoAssignee(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusDone))
	store.challenges[1] = []game.Challenge{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusDone)),
	}
	svc := NewQuestService(store)

	result, err := svc.GetByCrewAndID(context.Background(), 1, "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ActiveChallengeAssignedTo != nil {
		t.Errorf("expected nil active assignment, got %v", *result.ActiveChallengeAssignedTo)
	}
}

func TestCompleteChallengeForQuest_TransitionsToActive(t *testing.T) {
	store := newMockQuestStore()
	store.quests[1] = makeQuest(string(QuestStatusPending))
	c1 := makeChallenge(1, string(ChallengeStatusPending))
	c2 := makeChallenge(2, string(ChallengeStatusPending))
	store.challenges[1] = []game.Challenge{c1, c2}
	svc := NewQuestService(store)

	status, progressed, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 1, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !progressed {
		t.Error("expected challenge to be progressed")
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

	status, progressed, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 2, "user-2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !progressed {
		t.Error("expected challenge to be progressed")
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

	status, progressed, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 1, "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progressed {
		t.Error("expected challenge replay to report progressed=false (already done)")
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

func makeQuestWithID(id int64, status string) game.Quest {
	q := makeQuest(status)
	q.ID = id
	return *q
}

func makeChallengeForQuest(id, questID int64, status string) game.Challenge {
	c := makeChallenge(id, status)
	c.QuestID = questID
	return c
}

func TestList_BatchedChallengesAcrossQuests(t *testing.T) {
	store := newMockQuestStore()
	store.questsByCrew["crew-1"] = []game.Quest{
		makeQuestWithID(1, string(QuestStatusActive)),
		makeQuestWithID(2, string(QuestStatusActive)),
	}
	store.challenges[1] = []game.Challenge{
		makeChallengeForQuest(10, 1, string(ChallengeStatusDone)),
		makeChallengeForQuest(11, 1, string(ChallengeStatusPending)),
		makeChallengeForQuest(12, 1, string(ChallengeStatusPending)),
	}
	assigned := "user-2"
	store.challenges[1][1].AssignedTo = &assigned
	store.challenges[2] = []game.Challenge{
		makeChallengeForQuest(20, 2, string(ChallengeStatusDone)),
		makeChallengeForQuest(21, 2, string(ChallengeStatusDone)),
	}
	svc := NewQuestService(store)

	views, err := svc.List(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}

	byID := map[int64]QuestView{}
	for _, v := range views {
		byID[v.ID] = v
	}

	v1 := byID[1]
	if v1.ChallengeCount != 3 || v1.CompletedCount != 1 {
		t.Errorf("quest 1: expected 3 challenges / 1 done, got %d / %d", v1.ChallengeCount, v1.CompletedCount)
	}
	if v1.ActiveChallengeAssignedTo == nil {
		t.Error("quest 1: expected active challenge assigned_to to be set")
	}

	v2 := byID[2]
	if v2.ChallengeCount != 2 || v2.CompletedCount != 2 {
		t.Errorf("quest 2: expected 2 challenges / 2 done, got %d / %d", v2.ChallengeCount, v2.CompletedCount)
	}

	if store.listChallengeCalls != 1 {
		t.Errorf("expected 1 batched challenge read, got %d", store.listChallengeCalls)
	}
	if store.getChallengeCalls != 0 {
		t.Errorf("expected 0 sequential challenge reads, got %d", store.getChallengeCalls)
	}
}

func TestList_QuestsWithoutChallenges(t *testing.T) {
	store := newMockQuestStore()
	store.questsByCrew["crew-1"] = []game.Quest{
		makeQuestWithID(1, string(QuestStatusPending)),
		makeQuestWithID(2, string(QuestStatusDone)),
	}
	svc := NewQuestService(store)

	views, err := svc.List(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	for _, v := range views {
		if v.ChallengeCount != 0 {
			t.Errorf("quest %d: expected 0 challenges, got %d", v.ID, v.ChallengeCount)
		}
		if v.CompletedCount != 0 {
			t.Errorf("quest %d: expected 0 completed, got %d", v.ID, v.CompletedCount)
		}
		if v.ActiveChallengeAssignedTo != nil {
			t.Errorf("quest %d: expected nil active assignment, got %v", v.ID, v.ActiveChallengeAssignedTo)
		}
	}
}

// sequentialQuestStore wraps mockQuestStore but omits ListChallengesByQuestIDs
// so it only satisfies game.QuestStore, forcing the per-quest fallback path.
type sequentialQuestStore struct {
	inner *mockQuestStore
}

func (s *sequentialQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	return s.inner.GetQuest(ctx, questID)
}
func (s *sequentialQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	return s.inner.CreateQuest(ctx, q)
}
func (s *sequentialQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	return s.inner.ListQuestByCrew(ctx, crewID)
}
func (s *sequentialQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	return s.inner.UpdateQuest(ctx, questID, patch)
}
func (s *sequentialQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	return s.inner.UpdateQuestIfMatch(ctx, questID, oldStatus, patch)
}
func (s *sequentialQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	return s.inner.GetChallenges(ctx, questID)
}
func (s *sequentialQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	return s.inner.CreateChallenge(ctx, c)
}
func (s *sequentialQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	return s.inner.UpdateChallenge(ctx, challengeID, patch)
}
func (s *sequentialQuestStore) UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error) {
	return s.inner.UpdateChallengeIfMatch(ctx, challengeID, oldStatus, patch)
}

func TestList_FallbackSequentialWithoutBatchStore(t *testing.T) {
	inner := newMockQuestStore()
	inner.questsByCrew["crew-1"] = []game.Quest{
		makeQuestWithID(1, string(QuestStatusActive)),
		makeQuestWithID(2, string(QuestStatusActive)),
	}
	inner.challenges[1] = []game.Challenge{
		makeChallengeForQuest(10, 1, string(ChallengeStatusDone)),
		makeChallengeForQuest(11, 1, string(ChallengeStatusPending)),
	}
	store := &sequentialQuestStore{inner: inner}
	svc := NewQuestService(store)

	views, err := svc.List(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 views, got %d", len(views))
	}
	byID := map[int64]QuestView{}
	for _, v := range views {
		byID[v.ID] = v
	}
	v1 := byID[1]
	if v1.ChallengeCount != 2 || v1.CompletedCount != 1 {
		t.Errorf("quest 1: expected 2 challenges / 1 done for fallback quest, got %d / %d", v1.ChallengeCount, v1.CompletedCount)
	}
	if inner.listChallengeCalls != 0 {
		t.Errorf("expected 0 batched reads in fallback, got %d", inner.listChallengeCalls)
	}
	if inner.getChallengeCalls != 2 {
		t.Errorf("expected 2 sequential reads in fallback, got %d", inner.getChallengeCalls)
	}
}
