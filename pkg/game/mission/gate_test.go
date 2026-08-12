package quest

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
)

type mockGateChapterStore struct {
	progress map[string]*game.CourseProgress
}

func (m *mockGateChapterStore) GetChapterProgress(ctx context.Context, crewID, course string) (*game.CourseProgress, error) {
	cp, ok := m.progress[crewID+"|"+course]
	if !ok {
		return nil, game.ErrNotFound
	}
	return cp, nil
}
func (m *mockGateChapterStore) CreateChapterProgress(ctx context.Context, cp *game.CourseProgress) (*game.CourseProgress, error) {
	return cp, nil
}
func (m *mockGateChapterStore) UpdateChapterProgress(ctx context.Context, crewID, course string, patch map[string]any) error {
	return nil
}
func (m *mockGateChapterStore) ListChapterProgressByCrew(ctx context.Context, crewID string) ([]game.CourseProgress, error) {
	return nil, nil
}

type mockGateRealmStore struct {
	progress map[string]*game.JourneyProgress
}

func (m *mockGateRealmStore) GetRealmProgress(ctx context.Context, crewID, journey string) (*game.JourneyProgress, error) {
	rp, ok := m.progress[crewID+"|"+journey]
	if !ok {
		return nil, game.ErrNotFound
	}
	return rp, nil
}
func (m *mockGateRealmStore) CreateRealmProgress(ctx context.Context, rp *game.JourneyProgress) (*game.JourneyProgress, error) {
	return rp, nil
}
func (m *mockGateRealmStore) UpdateRealmProgress(ctx context.Context, crewID, journey string, patch map[string]any) error {
	return nil
}
func (m *mockGateRealmStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, journey string, oldProgress int, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockGateRealmStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.JourneyProgress, error) {
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
	missions map[string][]game.Mission
}

func (m *mockGateQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Mission, error) {
	return nil, game.ErrNotFound
}
func (m *mockGateQuestStore) CreateQuest(ctx context.Context, q *game.Mission) (*game.Mission, error) {
	return q, nil
}
func (m *mockGateQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Mission, error) {
	return m.missions[crewID], nil
}
func (m *mockGateQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	return nil
}
func (m *mockGateQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockGateQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Exercise, error) {
	return nil, nil
}
func (m *mockGateQuestStore) CreateChallenge(ctx context.Context, c *game.Exercise) (*game.Exercise, error) {
	return c, nil
}
func (m *mockGateQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	return nil
}
func (m *mockGateQuestStore) UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error) {
	return true, nil
}

func TestQuestGate_IsChapterUnlocked_TrueWhenActive(t *testing.T) {
	chapters := &mockGateChapterStore{
		progress: map[string]*game.CourseProgress{
			"crew-1|ch-1": {FamilyID: "crew-1", Course: "ch-1", Status: "ACTIVE"},
		},
	}
	gate := NewQuestGate(chapters, nil, nil, nil, nil)
	if !gate.IsChapterUnlocked(context.Background(), "crew-1", "ch-1") {
		t.Error("expected course unlocked when ACTIVE")
	}
}

func TestQuestGate_IsChapterUnlocked_FalseWhenLocked(t *testing.T) {
	chapters := &mockGateChapterStore{
		progress: map[string]*game.CourseProgress{
			"crew-1|ch-1": {FamilyID: "crew-1", Course: "ch-1", Status: "LOCKED"},
		},
	}
	gate := NewQuestGate(chapters, nil, nil, nil, nil)
	if gate.IsChapterUnlocked(context.Background(), "crew-1", "ch-1") {
		t.Error("expected course locked when LOCKED")
	}
}

func TestQuestGate_IsChapterUnlocked_FalseWhenNoProgress(t *testing.T) {
	chapters := &mockGateChapterStore{progress: map[string]*game.CourseProgress{}}
	gate := NewQuestGate(chapters, nil, nil, nil, nil)
	if gate.IsChapterUnlocked(context.Background(), "crew-1", "ch-1") {
		t.Error("expected course not unlocked when no progress")
	}
}

func TestQuestGate_IsRealmActive_TrueWhenActive(t *testing.T) {
	realms := &mockGateRealmStore{
		progress: map[string]*game.JourneyProgress{
			"crew-1|r1": {FamilyID: "crew-1", Journey: "r1", Status: "ACTIVE"},
		},
	}
	gate := NewQuestGate(nil, realms, nil, nil, nil)
	if !gate.IsRealmActive(context.Background(), "crew-1", "r1") {
		t.Error("expected journey active when ACTIVE")
	}
}

func TestQuestGate_IsRealmActive_FalseWhenNoProgress(t *testing.T) {
	realms := &mockGateRealmStore{progress: map[string]*game.JourneyProgress{}}
	gate := NewQuestGate(nil, realms, nil, nil, nil)
	if gate.IsRealmActive(context.Background(), "crew-1", "r1") {
		t.Error("expected journey not active when no progress")
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
	missions := &mockGateQuestStore{
		missions: map[string][]game.Mission{
			"crew-1": {
				{FamilyID: "crew-1", TemplateSlug: "q1", Status: "DONE"},
				{FamilyID: "crew-1", TemplateSlug: "q2", Status: "ACTIVE"},
			},
		},
	}
	gate := NewQuestGate(nil, nil, nil, missions, nil)
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
		progress: map[string]*game.CourseProgress{
			"crew-1|ch-2": {FamilyID: "crew-1", Course: "ch-2", Status: "ACTIVE"},
		},
	}
	realms := &mockGateRealmStore{
		progress: map[string]*game.JourneyProgress{
			"crew-1|r1": {FamilyID: "crew-1", Journey: "r1", Status: "ACTIVE"},
		},
	}
	users := &mockGateUserStore{
		player: &game.Player{UID: "u1", Level: 3},
	}
	missions := &mockGateQuestStore{
		missions: map[string][]game.Mission{
			"crew-1": {
				{FamilyID: "crew-1", TemplateSlug: "prereq-q", Status: "DONE"},
			},
		},
	}
	seasonChecker := func(ctx context.Context, slug string) bool { return true }
	gate := NewQuestGate(chapters, realms, users, missions, seasonChecker)

	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:              "new-quest",
		Course:           "ch-2",
		Journey:             "r1",
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
func (m *mockContentGateway) ListQuestsByRealm(ctx context.Context, journey string) ([]gamecontent.QuestDefinition, error) {
	return nil, nil
}
func (m *mockGateUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}
