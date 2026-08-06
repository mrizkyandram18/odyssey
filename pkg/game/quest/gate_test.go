package quest

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
)

type mockGateChapterStore struct {
	progress map[string]*game.ChapterProgress
}

func (m *mockGateChapterStore) GetChapterProgress(ctx context.Context, crewID, chapter string) (*game.ChapterProgress, error) {
	cp, ok := m.progress[crewID+"|"+chapter]
	if !ok {
		return nil, game.ErrNotFound
	}
	return cp, nil
}
func (m *mockGateChapterStore) CreateChapterProgress(ctx context.Context, cp *game.ChapterProgress) (*game.ChapterProgress, error) {
	return cp, nil
}
func (m *mockGateChapterStore) UpdateChapterProgress(ctx context.Context, crewID, chapter string, patch map[string]any) error {
	return nil
}
func (m *mockGateChapterStore) ListChapterProgressByCrew(ctx context.Context, crewID string) ([]game.ChapterProgress, error) {
	return nil, nil
}

type mockGateRealmStore struct {
	progress map[string]*game.RealmProgress
}

func (m *mockGateRealmStore) GetRealmProgress(ctx context.Context, crewID, realm string) (*game.RealmProgress, error) {
	rp, ok := m.progress[crewID+"|"+realm]
	if !ok {
		return nil, game.ErrNotFound
	}
	return rp, nil
}
func (m *mockGateRealmStore) CreateRealmProgress(ctx context.Context, rp *game.RealmProgress) (*game.RealmProgress, error) {
	return rp, nil
}
func (m *mockGateRealmStore) UpdateRealmProgress(ctx context.Context, crewID, realm string, patch map[string]any) error {
	return nil
}
func (m *mockGateRealmStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, realm string, oldProgress int, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockGateRealmStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.RealmProgress, error) {
	return nil, nil
}

type mockGateUserStore struct {
	player *game.Player
	err    error
}

func (m *mockGateUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	return m.player, m.err
}
func (m *mockGateUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	return m.err
}
func (m *mockGateUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return m.err
}
func (m *mockGateUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	return false, m.err
}

type mockGateQuestStore struct {
	quests map[string][]game.Quest
}

func (m *mockGateQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	return nil, game.ErrNotFound
}
func (m *mockGateQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	return q, nil
}
func (m *mockGateQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	return m.quests[crewID], nil
}
func (m *mockGateQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	return nil
}
func (m *mockGateQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockGateQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	return nil, nil
}
func (m *mockGateQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	return c, nil
}
func (m *mockGateQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	return nil
}

func TestQuestGate_IsChapterUnlocked_TrueWhenActive(t *testing.T) {
	chapters := &mockGateChapterStore{
		progress: map[string]*game.ChapterProgress{
			"crew-1|ch-1": {CrewID: "crew-1", Chapter: "ch-1", Status: "ACTIVE"},
		},
	}
	gate := NewQuestGate(chapters, nil, nil, nil, nil)
	if !gate.IsChapterUnlocked(context.Background(), "crew-1", "ch-1") {
		t.Error("expected chapter unlocked when ACTIVE")
	}
}

func TestQuestGate_IsChapterUnlocked_FalseWhenLocked(t *testing.T) {
	chapters := &mockGateChapterStore{
		progress: map[string]*game.ChapterProgress{
			"crew-1|ch-1": {CrewID: "crew-1", Chapter: "ch-1", Status: "LOCKED"},
		},
	}
	gate := NewQuestGate(chapters, nil, nil, nil, nil)
	if gate.IsChapterUnlocked(context.Background(), "crew-1", "ch-1") {
		t.Error("expected chapter locked when LOCKED")
	}
}

func TestQuestGate_IsChapterUnlocked_FalseWhenNoProgress(t *testing.T) {
	chapters := &mockGateChapterStore{progress: map[string]*game.ChapterProgress{}}
	gate := NewQuestGate(chapters, nil, nil, nil, nil)
	if gate.IsChapterUnlocked(context.Background(), "crew-1", "ch-1") {
		t.Error("expected chapter not unlocked when no progress")
	}
}

func TestQuestGate_IsRealmActive_TrueWhenActive(t *testing.T) {
	realms := &mockGateRealmStore{
		progress: map[string]*game.RealmProgress{
			"crew-1|r1": {CrewID: "crew-1", Realm: "r1", Status: "ACTIVE"},
		},
	}
	gate := NewQuestGate(nil, realms, nil, nil, nil)
	if !gate.IsRealmActive(context.Background(), "crew-1", "r1") {
		t.Error("expected realm active when ACTIVE")
	}
}

func TestQuestGate_IsRealmActive_FalseWhenNoProgress(t *testing.T) {
	realms := &mockGateRealmStore{progress: map[string]*game.RealmProgress{}}
	gate := NewQuestGate(nil, realms, nil, nil, nil)
	if gate.IsRealmActive(context.Background(), "crew-1", "r1") {
		t.Error("expected realm not active when no progress")
	}
}

func TestQuestGate_IsSeasonActive_NilChecker(t *testing.T) {
	gate := NewQuestGate(nil, nil, nil, nil, nil)
	if !gate.IsSeasonActive(context.Background(), "summer") {
		t.Error("expected season active when checker is nil")
	}
}

func TestQuestGate_IsSeasonActive_Checker(t *testing.T) {
	checker := func(ctx context.Context, slug string) bool {
		return slug == "summer"
	}
	gate := NewQuestGate(nil, nil, nil, nil, checker)
	if !gate.IsSeasonActive(context.Background(), "summer") {
		t.Error("expected summer active")
	}
	if gate.IsSeasonActive(context.Background(), "winter") {
		t.Error("expected winter inactive")
	}
}

func TestQuestGate_GetPlayerLevel(t *testing.T) {
	now := time.Date(2026, 8, 4, 0, 0, 0, 0, time.UTC)
	users := &mockGateUserStore{
		player: &game.Player{UID: "u1", Level: 5, UpdatedAt: now},
	}
	gate := NewQuestGate(nil, nil, users, nil, nil)
	level, err := gate.GetPlayerLevel(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level != 5 {
		t.Errorf("expected level 5, got %d", level)
	}
}

func TestQuestGate_IsQuestCompleted_True(t *testing.T) {
	quests := &mockGateQuestStore{
		quests: map[string][]game.Quest{
			"crew-1": {
				{CrewID: "crew-1", TemplateSlug: "q1", Status: "DONE"},
				{CrewID: "crew-1", TemplateSlug: "q2", Status: "ACTIVE"},
			},
		},
	}
	gate := NewQuestGate(nil, nil, nil, quests, nil)
	if !gate.IsQuestCompleted(context.Background(), "crew-1", "q1") {
		t.Error("expected q1 completed")
	}
	if gate.IsQuestCompleted(context.Background(), "crew-1", "q2") {
		t.Error("expected q2 not completed")
	}
	if gate.IsQuestCompleted(context.Background(), "crew-1", "unknown") {
		t.Error("expected unknown quest not completed")
	}
}

func TestQuestGate_IsPrerequisiteMet_AllConditionsMet(t *testing.T) {
	chapters := &mockGateChapterStore{
		progress: map[string]*game.ChapterProgress{
			"crew-1|ch-2": {CrewID: "crew-1", Chapter: "ch-2", Status: "ACTIVE"},
		},
	}
	realms := &mockGateRealmStore{
		progress: map[string]*game.RealmProgress{
			"crew-1|r1": {CrewID: "crew-1", Realm: "r1", Status: "ACTIVE"},
		},
	}
	users := &mockGateUserStore{
		player: &game.Player{UID: "u1", Level: 3},
	}
	quests := &mockGateQuestStore{
		quests: map[string][]game.Quest{
			"crew-1": {
				{CrewID: "crew-1", TemplateSlug: "prereq-q", Status: "DONE"},
			},
		},
	}
	seasonChecker := func(ctx context.Context, slug string) bool { return true }
	gate := NewQuestGate(chapters, realms, users, quests, seasonChecker)

	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:              "new-quest",
		Chapter:           "ch-2",
		Realm:             "r1",
		RequiredLevel:     3,
		RequiredQuestSlug: "prereq-q",
		RequiredChapter:   "ch-2",
		RequiredRealm:     "r1",
		SeasonSlug:        "summer",
	}

	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected prerequisite met")
	}
}

func TestQuestGate_ValidatePrerequisites_NoCycle(t *testing.T) {
	gate := NewQuestGate(nil, nil, nil, nil, nil)
	defs := []gamecontent.QuestDefinition{
		{Slug: "q1", RequiredQuestSlug: "q2"},
		{Slug: "q2", RequiredQuestSlug: "q3"},
		{Slug: "q3"},
	}
	if err := gate.ValidatePrerequisites(context.Background(), defs); err != nil {
		t.Fatalf("expected no cycle, got %v", err)
	}
}

func TestQuestGate_ValidatePrerequisites_CycleDetected(t *testing.T) {
	gate := NewQuestGate(nil, nil, nil, nil, nil)
	defs := []gamecontent.QuestDefinition{
		{Slug: "qA", RequiredQuestSlugs: []string{"qB"}},
		{Slug: "qB", RequiredQuestSlugs: []string{"qA"}},
	}
	if err := gate.ValidatePrerequisites(context.Background(), defs); err == nil {
		t.Error("expected cycle error")
	}
}

func TestQuestGate_ValidatePrerequisites_MultiPrereqNoCycle(t *testing.T) {
	gate := NewQuestGate(nil, nil, nil, nil, nil)
	defs := []gamecontent.QuestDefinition{
		{Slug: "q1", RequiredQuestSlugs: []string{"q2", "q3"}},
		{Slug: "q2"},
		{Slug: "q3"},
	}
	if err := gate.ValidatePrerequisites(context.Background(), defs); err != nil {
		t.Fatalf("expected no cycle, got %v", err)
	}
}

func TestQuestGate_ValidatePrerequisites_BackwardCompatSingleSlug(t *testing.T) {
	gate := NewQuestGate(nil, nil, nil, nil, nil)
	defs := []gamecontent.QuestDefinition{
		{Slug: "q1", RequiredQuestSlug: "q2"},
		{Slug: "q2"},
	}
	if err := gate.ValidatePrerequisites(context.Background(), defs); err != nil {
		t.Fatalf("expected no cycle, got %v", err)
	}
}

type mockContentGateway struct{}

func (m *mockContentGateway) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return nil, nil
}
func (m *mockContentGateway) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	return nil, nil
}
func (m *mockContentGateway) ListQuestsByRealm(ctx context.Context, realm string) ([]gamecontent.QuestDefinition, error) {
	return nil, nil
}
func (m *mockGateUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) { return nil, nil }
