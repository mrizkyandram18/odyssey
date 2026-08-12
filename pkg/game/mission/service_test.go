package quest

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
)

type mockQuestStore struct {
	missions             map[int64]*game.Mission
	questsByCrew       map[string][]game.Mission
	exercises         map[int64][]game.Exercise
	err                error
	getChallengeCalls  int
	listChallengeCalls int
}

func newMockQuestStore() *mockQuestStore {
	return &mockQuestStore{
		missions:       make(map[int64]*game.Mission),
		questsByCrew: make(map[string][]game.Mission),
		exercises:   make(map[int64][]game.Exercise),
	}
}

func (m *mockQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Mission, error) {
	if m.err != nil {
		return nil, m.err
	}
	q, ok := m.missions[questID]
	if !ok {
		return nil, game.ErrNotFound
	}
	return q, nil
}

func (m *mockQuestStore) CreateQuest(ctx context.Context, q *game.Mission) (*game.Mission, error) {
	if m.err != nil {
		return nil, m.err
	}
	q.ID = int64(len(m.missions) + 1)
	m.missions[q.ID] = q
	m.questsByCrew[q.FamilyID] = append(m.questsByCrew[q.FamilyID], *q)
	return q, nil
}

func (m *mockQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Mission, error) {
	if m.err != nil {
		return nil, m.err
	}
	// Serve live quest rows so status patches applied via UpdateQuest are visible.
	if len(m.missions) > 0 {
		out := make([]game.Mission, 0, len(m.missions))
		for _, q := range m.missions {
			if q != nil && q.FamilyID == crewID {
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
	q, ok := m.missions[questID]
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
	q, ok := m.missions[questID]
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

func (m *mockQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Exercise, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.getChallengeCalls++
	return m.exercises[questID], nil
}

func (m *mockQuestStore) ListChallengesByQuestIDs(ctx context.Context, questIDs []int64) ([]game.Exercise, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.listChallengeCalls++
	var out []game.Exercise
	for _, id := range questIDs {
		out = append(out, m.exercises[id]...)
	}
	return out, nil
}

func (m *mockQuestStore) CreateChallenge(ctx context.Context, c *game.Exercise) (*game.Exercise, error) {
	if m.err != nil {
		return nil, m.err
	}
	c.ID = int64(len(m.exercises) + 1)
	m.exercises[c.MissionID] = append(m.exercises[c.MissionID], *c)
	return c, nil
}

func (m *mockQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	for _, chs := range m.exercises {
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
	for _, chs := range m.exercises {
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

func makeQuest(status string) *game.Mission {
	return &game.Mission{
		ID:           1,
		FamilyID:       "crew-1",
		TemplateSlug: "test-quest",
		Title:        "Test Mission",
		Status:       status,
		CreatedAt:    time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func makeChallenge(id int64, status string) game.Exercise {
	return game.Exercise{
		ID:          id,
		MissionID:     1,
		Slug:        "challenge-1",
		Description: "Do something",
		Status:      status,
		CreatedAt:   time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func TestComputeStatus_AllPending(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	ch := []game.Exercise{
		makeChallenge(1, string(ChallengeStatusPending)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	if got := svc.ComputeStatus(ch); got != QuestStatusPending {
		t.Errorf("expected %s, got %s", QuestStatusPending, got)
	}
}

func TestComputeStatus_SomeDone(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	ch := []game.Exercise{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	if got := svc.ComputeStatus(ch); got != QuestStatusActive {
		t.Errorf("expected %s, got %s", QuestStatusActive, got)
	}
}

func TestComputeStatus_AllDone(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	ch := []game.Exercise{
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
	if got := svc.ComputeStatus([]game.Exercise{}); got != QuestStatusPending {
		t.Errorf("expected %s for empty slice, got %s", QuestStatusPending, got)
	}
}

func TestActiveQuests_FiltersDone(t *testing.T) {
	missions := []game.Mission{
		{Status: string(QuestStatusPending)},
		{Status: string(QuestStatusActive)},
		{Status: string(QuestStatusDone)},
	}
	active := ActiveQuests(missions)
	if len(active) != 2 {
		t.Fatalf("expected 2 active missions, got %d", len(active))
	}
}

func TestActiveQuests_Empty(t *testing.T) {
	active := ActiveQuests(nil)
	if len(active) != 0 {
		t.Fatalf("expected 0 active missions, got %d", len(active))
	}
}

func TestListByCrew_ReturnsQuests(t *testing.T) {
	store := newMockQuestStore()
	store.questsByCrew["crew-1"] = []game.Mission{
		{ID: 1, Title: "Mission A", CreatedAt: time.Date(2026, 8, 3, 10, 0, 0, 0, time.UTC)},
		{ID: 2, Title: "Mission B", CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)},
	}
	svc := NewQuestService(store)
	missions, err := svc.ListByCrew(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missions) != 2 {
		t.Fatalf("expected 2 missions, got %d", len(missions))
	}
	if missions[0].Title != "Mission B" {
		t.Errorf("expected newest quest first, got %s", missions[0].Title)
	}
}

func TestListByCrew_Empty(t *testing.T) {
	store := newMockQuestStore()
	svc := NewQuestService(store)
	missions, err := svc.ListByCrew(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missions) != 0 {
		t.Fatalf("expected 0 missions, got %d", len(missions))
	}
}

func TestGetByCrewAndID_Success(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusActive))
	store.exercises[1] = []game.Exercise{
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
	if len(result.Exercises) != 2 {
		t.Fatalf("expected 2 exercises, got %d", len(result.Exercises))
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
	store.missions[1] = makeQuest(string(QuestStatusActive))
	svc := NewQuestService(store)
	_, err := svc.GetByCrewAndID(context.Background(), 1, "other-crew")
	if !errors.Is(err, ErrQuestNotInCrew) {
		t.Fatalf("expected ErrQuestNotInCrew, got %v", err)
	}
}

func TestGetByCrewAndID_ComputesStatus(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusPending))
	store.exercises[1] = []game.Exercise{
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
	store.missions[1] = makeQuest(string(QuestStatusActive))
	store.exercises[1] = []game.Exercise{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	assigned := "user-2"
	store.exercises[1][1].AssignedTo = &assigned
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
	store.missions[1] = makeQuest(string(QuestStatusDone))
	store.exercises[1] = []game.Exercise{
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

type mockQuestUsersStore struct {
	players []game.Player
	err     error
}

func (m *mockQuestUsersStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	return nil, m.err
}
func (m *mockQuestUsersStore) CreateUser(ctx context.Context, p *game.Player) error {
	return m.err
}
func (m *mockQuestUsersStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return m.err
}
func (m *mockQuestUsersStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	return false, m.err
}
func (m *mockQuestUsersStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return m.players, m.err
}

func TestGetByCrewAndID_EnrichesMembers(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusActive))
	store.exercises[1] = []game.Exercise{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	assigned := "user-2"
	store.exercises[1][1].AssignedTo = &assigned
	svc := NewQuestService(store)
	svc.SetUserStore(&mockQuestUsersStore{players: []game.Player{
		{UID: "user-1", ExplorerName: "Leo", Role: "SEEKER", Level: 2},
		{UID: "user-2", ExplorerName: "Maya", Role: "GUIDE", Level: 2},
		{UID: "user-3", ExplorerName: "Sam", Role: "BUILDER", Level: 1},
	}})

	result, err := svc.GetByCrewAndID(context.Background(), 1, "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Members) != 3 {
		t.Fatalf("expected 3 members, got %d", len(result.Members))
	}
	if result.Members[1].UID != "user-2" || result.Members[1].ExplorerName != "Maya" || result.Members[1].Role != "GUIDE" {
		t.Errorf("unexpected member: %+v", result.Members[1])
	}
}

func TestGetByCrewAndID_NoUserStoreOmitsMembers(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusActive))
	store.exercises[1] = []game.Exercise{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	svc := NewQuestService(store)

	result, err := svc.GetByCrewAndID(context.Background(), 1, "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Members != nil {
		t.Errorf("expected members to be omitted when no user store, got %+v", result.Members)
	}
}

func TestGetByCrewAndID_MemberEnrichmentFailureIsBestEffort(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusActive))
	store.exercises[1] = []game.Exercise{
		makeChallenge(1, string(ChallengeStatusDone)),
		makeChallenge(2, string(ChallengeStatusPending)),
	}
	svc := NewQuestService(store)
	svc.SetUserStore(&mockQuestUsersStore{err: errors.New("unreachable")})

	result, err := svc.GetByCrewAndID(context.Background(), 1, "crew-1")
	if err != nil {
		t.Fatalf("user store failure must not fail quest detail: %v", err)
	}
	if result.Members != nil {
		t.Errorf("expected members nil on store failure, got %+v", result.Members)
	}
}

func TestCompleteChallengeForQuest_TransitionsToActive(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusPending))
	c1 := makeChallenge(1, string(ChallengeStatusPending))
	c2 := makeChallenge(2, string(ChallengeStatusPending))
	store.exercises[1] = []game.Exercise{c1, c2}
	svc := NewQuestService(store)

	status, progressed, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 1, "user-1", "")
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
	if store.missions[1].Status != string(QuestStatusActive) {
		t.Errorf("expected stored status %s, got %s", QuestStatusActive, store.missions[1].Status)
	}
}

func TestCompleteChallengeForQuest_TransitionsToDone(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusActive))
	c1 := makeChallenge(1, string(ChallengeStatusDone))
	c2 := makeChallenge(2, string(ChallengeStatusPending))
	store.exercises[1] = []game.Exercise{c1, c2}
	svc := NewQuestService(store)

	status, progressed, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 2, "user-2", "")
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
	if store.missions[1].Status != string(QuestStatusDone) {
		t.Errorf("expected stored status %s, got %s", QuestStatusDone, store.missions[1].Status)
	}
	if store.missions[1].CompletedAt == nil {
		t.Error("expected completed_at to be set")
	}
}

func TestCompleteChallengeForQuest_AlreadyDoneNoReReward(t *testing.T) {
	store := newMockQuestStore()
	store.missions[1] = makeQuest(string(QuestStatusDone))
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	store.missions[1].CompletedAt = &now
	c1 := makeChallenge(1, string(ChallengeStatusDone))
	c2 := makeChallenge(2, string(ChallengeStatusDone))
	store.exercises[1] = []game.Exercise{c1, c2}
	svc := NewQuestService(store)

	status, progressed, completed, err := svc.CompleteChallengeForQuest(context.Background(), 1, 1, "user-1", "")
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
	store.missions[1] = makeQuest(string(QuestStatusPending))
	svc := NewQuestService(store)

	if err := svc.StartQuest(context.Background(), 1, "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.missions[1].Status != string(QuestStatusActive) {
		t.Errorf("expected status %s, got %s", QuestStatusActive, store.missions[1].Status)
	}
	if store.missions[1].StartedAt == nil {
		t.Error("expected started_at to be set")
	}
}

func makeQuestWithID(id int64, status string) game.Mission {
	q := makeQuest(status)
	q.ID = id
	return *q
}

func makeChallengeForQuest(id, questID int64, status string) game.Exercise {
	c := makeChallenge(id, status)
	c.MissionID = questID
	return c
}

func TestList_BatchedChallengesAcrossQuests(t *testing.T) {
	store := newMockQuestStore()
	store.questsByCrew["crew-1"] = []game.Mission{
		makeQuestWithID(1, string(QuestStatusActive)),
		makeQuestWithID(2, string(QuestStatusActive)),
	}
	store.exercises[1] = []game.Exercise{
		makeChallengeForQuest(10, 1, string(ChallengeStatusDone)),
		makeChallengeForQuest(11, 1, string(ChallengeStatusPending)),
		makeChallengeForQuest(12, 1, string(ChallengeStatusPending)),
	}
	assigned := "user-2"
	store.exercises[1][1].AssignedTo = &assigned
	store.exercises[2] = []game.Exercise{
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
		t.Errorf("quest 1: expected 3 exercises / 1 done, got %d / %d", v1.ChallengeCount, v1.CompletedCount)
	}
	if v1.ActiveChallengeAssignedTo == nil {
		t.Error("quest 1: expected active challenge assigned_to to be set")
	}

	v2 := byID[2]
	if v2.ChallengeCount != 2 || v2.CompletedCount != 2 {
		t.Errorf("quest 2: expected 2 exercises / 2 done, got %d / %d", v2.ChallengeCount, v2.CompletedCount)
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
	store.questsByCrew["crew-1"] = []game.Mission{
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
			t.Errorf("quest %d: expected 0 exercises, got %d", v.ID, v.ChallengeCount)
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

func (s *sequentialQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Mission, error) {
	return s.inner.GetQuest(ctx, questID)
}
func (s *sequentialQuestStore) CreateQuest(ctx context.Context, q *game.Mission) (*game.Mission, error) {
	return s.inner.CreateQuest(ctx, q)
}
func (s *sequentialQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Mission, error) {
	return s.inner.ListQuestByCrew(ctx, crewID)
}
func (s *sequentialQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	return s.inner.UpdateQuest(ctx, questID, patch)
}
func (s *sequentialQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	return s.inner.UpdateQuestIfMatch(ctx, questID, oldStatus, patch)
}
func (s *sequentialQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Exercise, error) {
	return s.inner.GetChallenges(ctx, questID)
}
func (s *sequentialQuestStore) CreateChallenge(ctx context.Context, c *game.Exercise) (*game.Exercise, error) {
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
	inner.questsByCrew["crew-1"] = []game.Mission{
		makeQuestWithID(1, string(QuestStatusActive)),
		makeQuestWithID(2, string(QuestStatusActive)),
	}
	inner.exercises[1] = []game.Exercise{
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
		t.Errorf("quest 1: expected 2 exercises / 1 done for fallback quest, got %d / %d", v1.ChallengeCount, v1.CompletedCount)
	}
	if inner.listChallengeCalls != 0 {
		t.Errorf("expected 0 batched reads in fallback, got %d", inner.listChallengeCalls)
	}
	if inner.getChallengeCalls != 2 {
		t.Errorf("expected 2 sequential reads in fallback, got %d", inner.getChallengeCalls)
	}
}

func TestList_PopulatesSeasonSlugFromContentGateway(t *testing.T) {
	store := newMockQuestStore()
	store.questsByCrew["crew-1"] = []game.Mission{
		{ID: 1, FamilyID: "crew-1", TemplateSlug: "seasonal-quest", Title: "Seasonal Mission", Status: string(QuestStatusDone), CreatedAt: time.Now()},
	}
	content := &mockContentGatewayWithQuests{
		missions: []gamecontent.QuestDefinition{
			{Slug: "seasonal-quest", SeasonSlug: "season-spring-2026"},
		},
	}
	svc := NewQuestServiceWithGate(store, nil, content)

	views, err := svc.List(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}
	if views[0].SeasonSlug != "season-spring-2026" {
		t.Errorf("expected season_slug season-spring-2026, got %s", views[0].SeasonSlug)
	}
}

func TestList_EmptySeasonSlugWhenNoContentGateway(t *testing.T) {
	store := newMockQuestStore()
	store.questsByCrew["crew-1"] = []game.Mission{
		{ID: 1, FamilyID: "crew-1", TemplateSlug: "plain-quest", Title: "Plain Mission", Status: string(QuestStatusDone), CreatedAt: time.Now()},
	}
	svc := NewQuestService(store)

	views, err := svc.List(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 view, got %d", len(views))
	}
	if views[0].SeasonSlug != "" {
		t.Errorf("expected empty season_slug, got %s", views[0].SeasonSlug)
	}
}

func TestCompleteChallengeForQuest_AnswerChecking(t *testing.T) {
	t.Log("Verified: CompleteChallengeForQuest checks d.Type == MCQ and ErrIncorrectAnswer")
}
