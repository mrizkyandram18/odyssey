package quest

import (
	"context"
	"sync"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/world"
)

// --- Mocks for QuestAPIHandler collaborators ---
// (mockQuestStore is shared from service_test.go in this package.)

type mockUserStoreForHandler struct {
	player *game.Player
	err    error
}

func (m *mockUserStoreForHandler) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.player, nil
}
func (m *mockUserStoreForHandler) CreateUser(ctx context.Context, p *game.Player) error { return m.err }
func (m *mockUserStoreForHandler) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	p := m.player
	if v, ok := patch["xp"].(int64); ok {
		p.XP = v
	}
	if v, ok := patch["level"].(int); ok {
		p.Level = v
	}
	if v, ok := patch["version"].(int); ok {
		p.Version = v
	}
	return nil
}
func (m *mockUserStoreForHandler) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	if m.player == nil {
		return false, game.ErrNotFound
	}
	if m.player.Version != version {
		return false, nil
	}
	if err := m.UpdateUser(ctx, uid, patch); err != nil {
		return false, err
	}
	return true, nil
}

type mockRealmProgressStoreForHandler struct {
	progress map[string]*game.JourneyProgress
	updates  []string
	err      error
}

func newMockRealmStore() *mockRealmProgressStoreForHandler {
	return &mockRealmProgressStoreForHandler{
		progress: map[string]*game.JourneyProgress{
			"crew-1|whispering-woods": {
				FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 10,
			},
		},
		updates: []string{},
	}
}

func (m *mockRealmProgressStoreForHandler) GetRealmProgress(ctx context.Context, crewID, journey string) (*game.JourneyProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.progress[crewID+"|"+journey], nil
}
func (m *mockRealmProgressStoreForHandler) CreateRealmProgress(ctx context.Context, rp *game.JourneyProgress) (*game.JourneyProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.progress[rp.FamilyID+"|"+rp.Journey] = rp
	return rp, nil
}
func (m *mockRealmProgressStoreForHandler) UpdateRealmProgress(ctx context.Context, crewID, journey string, patch map[string]any) error {
	m.updates = append(m.updates, crewID+"|"+journey)
	if m.err != nil {
		return m.err
	}
	rp := m.progress[crewID+"|"+journey]
	if rp != nil {
		if v, ok := patch["progress"].(int); ok {
			rp.Progress = v
		}
		if v, ok := patch["status"].(string); ok {
			rp.Status = v
		}
		if v, ok := patch["story_branch"].(string); ok {
			rp.StoryBranch = v
		}
		if v, ok := patch["last_unlocked_at"]; ok && v != nil {
			rp.LastUnlockedAt, _ = patch["last_unlocked_at"].(time.Time)
		}
	}
	return nil
}
func (m *mockRealmProgressStoreForHandler) UpdateRealmProgressIfMatch(ctx context.Context, crewID, journey string, oldProgress int, patch map[string]any) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	rp := m.progress[crewID+"|"+journey]
	if rp == nil {
		return false, game.ErrNotFound
	}
	if rp.Progress != oldProgress {
		return false, nil
	}
	if err := m.UpdateRealmProgress(ctx, crewID, journey, patch); err != nil {
		return false, err
	}
	return true, nil
}
func (m *mockRealmProgressStoreForHandler) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.JourneyProgress, error) {
	return nil, m.err
}

func makePlayerForHandler(level int, xp int64) *game.Player {
	return &game.Player{
		UID:          "user-1",
		FamilyID:       "crew-1",
		ExplorerName: "Alice",
		Level:        level,
		XP:           xp,
		CreatedAt:    time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

var defaultRealmCfg = world.NewRealmCatalog(world.DefaultRealmDefinitions)
var defaultProgCfg = progression.DefaultProgressionConfig()

func setupOrchestrator(t *testing.T, qStore *mockQuestStore, player *game.Player) (*QuestAPIHandler, *mockRealmProgressStoreForHandler) {
	t.Helper()
	qs := NewQuestService(qStore)
	userStore := &mockUserStoreForHandler{player: player}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	journey := newMockRealmStore()
	return NewQuestAPIHandler(qs, prog, journey, defaultRealmCfg, &defaultProgCfg), journey
}

func makeStoredQuest(questID int64, crewID, slug, status string) *game.Mission {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	return &game.Mission{
		ID:           questID,
		FamilyID:       crewID,
		TemplateSlug: slug,
		Title:        "Morning Light",
		Status:       status,
		CreatedAt:    now,
	}
}

func makeStoredChallenges(questID int64, statuses ...string) []game.Exercise {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	chs := make([]game.Exercise, 0, len(statuses))
	for i, s := range statuses {
		chs = append(chs, game.Exercise{
			ID:          questID*10 + int64(i+1),
			MissionID:     questID,
			Slug:        "ch-" + string(rune('a'+i)),
			Description: "Exercise " + string(rune('a'+i)),
			Status:      s,
			CreatedAt:   now,
		})
	}
	return chs
}

func TestStartQuest_Succeeds(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.missions[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusPending))
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}

	h, _ := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	if err := h.StartQuest(context.Background(), 1, "crew-1", "user-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if qStore.missions[1].Status != string(QuestStatusActive) {
		t.Errorf("expected ACTIVE, got %s", qStore.missions[1].Status)
	}
	if qStore.missions[1].StartedAt == nil {
		t.Error("expected started_at set")
	}
}

func TestStartQuest_NotInCrew(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.missions[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusPending))
	h, _ := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	err := h.StartQuest(context.Background(), 1, "other-crew", "user-1")
	if err == nil {
		t.Fatal("expected error when starting a quest not in crew")
	}
}

func TestCompleteChallenge_NonLastAwardsChallengeXPOnly(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.missions[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusPending), string(ChallengeStatusPending))

	h, journey := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	result, err := h.CompleteChallenge(context.Background(), 1, 11, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.QuestCompleted {
		t.Error("expected quest not yet completed")
	}
	if result.XP != ChallengeXP {
		t.Errorf("expected XP %d, got %d", ChallengeXP, result.XP)
	}
	if result.Mission.Status != string(QuestStatusActive) {
		t.Errorf("expected quest ACTIVE, got %s", result.Mission.Status)
	}
	if result.Mission.CompletedAt != nil {
		t.Error("expected completed_at not set for non-completing action")
	}
	if len(journey.updates) != 0 {
		t.Error("expected no journey progress update for non-completing action")
	}
}

func TestCompleteChallenge_LastChallengeCompletesQuest(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	h, journey := setupOrchestrator(t, qStore, makePlayerForHandler(1, 490))
	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.QuestCompleted {
		t.Error("expected quest to be completed")
	}
	if result.XP != ChallengeXP+CompletionBonusXP {
		t.Errorf("expected XP %d, got %d", ChallengeXP+CompletionBonusXP, result.XP)
	}
	if result.NewLevel != 2 {
		t.Errorf("expected level 2, got %d", result.NewLevel)
	}
	if !result.LevelUp {
		t.Error("expected level up flag")
	}
	if result.Mission.Status != string(QuestStatusDone) {
		t.Errorf("expected quest DONE, got %s", result.Mission.Status)
	}
	if result.Mission.CompletedAt == nil {
		t.Error("expected completed_at set")
	}
	if len(journey.updates) != 1 {
		t.Errorf("expected 1 journey progress update, got %d", len(journey.updates))
	}
}

func TestCompleteChallenge_FullLoop(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.missions[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusPending))
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusPending), string(ChallengeStatusPending))

	h, journey := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))

	if err := h.StartQuest(context.Background(), 1, "crew-1", "user-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if qStore.missions[1].Status != string(QuestStatusActive) {
		t.Fatalf("expected ACTIVE after start, got %s", qStore.missions[1].Status)
	}

	r1, err := h.CompleteChallenge(context.Background(), 1, 11, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("complete 1: %v", err)
	}
	if r1.QuestCompleted {
		t.Fatal("quest should not be completed after 1/2 exercises")
	}

	r2, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("complete 2: %v", err)
	}
	if !r2.QuestCompleted {
		t.Fatal("quest should be completed after 2/2 exercises")
	}
	if qStore.missions[1].Status != string(QuestStatusDone) {
		t.Errorf("expected DONE after loop, got %s", qStore.missions[1].Status)
	}
	if len(journey.updates) != 1 {
		t.Errorf("expected 1 journey update, got %d", len(journey.updates))
	}
}

func TestCompleteChallenge_UnknownQuest(t *testing.T) {
	qStore := newMockQuestStore()
	h, _ := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	_, err := h.CompleteChallenge(context.Background(), 999, 1, "crew-1", "user-1", "")
	if err == nil {
		t.Fatal("expected error for unknown quest")
	}
}

func TestRealmForSlug(t *testing.T) {
	if got := RealmForSlug("morning-light"); got != "whispering-woods" {
		t.Errorf("expected whispering-woods, got %s", got)
	}
	if got := RealmForSlug("unknown-slug"); got != "" {
		t.Errorf("expected empty journey for unknown slug, got %s", got)
	}
}

func TestCompleteChallenge_UsesCustomProgressionConfig(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	customCfg := &progression.ProgressionConfig{
		ChallengeXP:       10,
		CompletionBonusXP: 40,
	}
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, customCfg)
	journey := newMockRealmStore()
	h := NewQuestAPIHandler(NewQuestService(qStore), prog, journey, defaultRealmCfg, customCfg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.XP != customCfg.ChallengeXP+customCfg.CompletionBonusXP {
		t.Errorf("expected XP %d, got %d", customCfg.ChallengeXP+customCfg.CompletionBonusXP, result.XP)
	}
}

func TestCompleteChallenge_PublishesEventWithContentGateway(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	pub := &capturePublisher{}
	cg := &mockContentGatewayForHandler{
		missions: []gamecontent.QuestDefinition{
			{Slug: "morning-light", Course: "ch-1", SeasonSlug: ""},
		},
	}
	qs := NewQuestServiceWithGate(qStore, nil, cg)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 90)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	h := NewQuestAPIHandler(qs, prog, newMockRealmStore(), defaultRealmCfg, &defaultProgCfg)
	h.SetPublisher(pub)
	h.SetContentGateway(cg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.QuestCompleted {
		t.Fatal("expected quest completed")
	}
	if len(pub.published) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.published))
	}
	ev, ok := pub.published[0].(events.QuestCompletedEvent)
	if !ok {
		t.Fatalf("expected QuestCompletedEvent, got %T", pub.published[0])
	}
	if ev.Course != "ch-1" {
		t.Errorf("expected course ch-1, got %s", ev.Course)
	}
	if ev.Journey != "whispering-woods" {
		t.Errorf("expected journey whispering-woods, got %s", ev.Journey)
	}
}

func TestCompleteChallenge_NoPublisher_NoError(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 90)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	h := NewQuestAPIHandler(NewQuestService(qStore), prog, newMockRealmStore(), defaultRealmCfg, &defaultProgCfg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.QuestCompleted {
		t.Error("expected quest completed")
	}
}

type capturePublisher struct {
	published []events.Event
}

func (p *capturePublisher) Publish(ctx context.Context, event events.Event) {
	p.published = append(p.published, event)
}

type mockContentGatewayForHandler struct {
	missions []gamecontent.QuestDefinition
}

func (m *mockContentGatewayForHandler) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.missions, nil
}
func (m *mockContentGatewayForHandler) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	for _, q := range m.missions {
		if q.Slug == slug {
			return &q, nil
		}
	}
	return nil, nil
}
func (m *mockContentGatewayForHandler) ListQuestsByRealm(ctx context.Context, journey string) ([]gamecontent.QuestDefinition, error) {
	return nil, nil
}

func TestCompleteChallenge_UsesQuestRewardXP(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	cg := &mockContentGatewayForHandler{
		missions: []gamecontent.QuestDefinition{
			{Slug: "morning-light", RewardXP: 80},
		},
	}
	qs := NewQuestServiceWithGate(qStore, nil, cg)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	journey := newMockRealmStore()
	h := NewQuestAPIHandler(qs, prog, journey, defaultRealmCfg, &defaultProgCfg)
	h.SetContentGateway(cg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ChallengeXP + 80
	if result.XP != expected {
		t.Errorf("expected XP %d, got %d", expected, result.XP)
	}
}

func TestCompleteChallenge_UsesQuestRewardXPWithBalanceMultiplier(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	cg := &mockContentGatewayForHandler{
		missions: []gamecontent.QuestDefinition{
			{Slug: "morning-light", RewardXP: 100},
		},
	}
	qs := NewQuestServiceWithGate(qStore, nil, cg)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	journey := newMockRealmStore()
	h := NewQuestAPIHandler(qs, prog, journey, defaultRealmCfg, &defaultProgCfg)
	h.SetContentGateway(cg)
	bal := balance.NewService(&mockBalanceStore{
		overrides: map[string]int64{
			string(balance.KeyQuestRewardXP): 150,
		},
	})
	if err := bal.Load(context.Background()); err != nil {
		t.Fatalf("balance load failed: %v", err)
	}
	h.SetBalance(bal)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := ChallengeXP + int64(100*1.5)
	if result.XP != expected {
		t.Errorf("expected XP %d, got %d", expected, result.XP)
	}
}

func TestAdvanceRealm_UsesBalanceOverrides(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.missions[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}

	qs := NewQuestService(qStore)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	journey := newMockRealmStore()
	realmCfg := world.NewRealmCatalog([]world.RealmDefinition{
		{Slug: "whispering-woods", Name: "Whispering Woods", Order: 1, MaxProgress: 100},
	})
	h := NewQuestAPIHandler(qs, prog, journey, realmCfg, &defaultProgCfg)
	h.SetBalance(balance.NewService(&mockBalanceStore{
		overrides: map[string]int64{
			string(balance.KeyRealmProgressPerQuest):    50,
			string(balance.KeyRealmCompletionThreshold): 100,
		},
	}))

	err := h.advanceRealm(context.Background(), "crew-1", "whispering-woods")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(journey.updates) != 1 {
		t.Errorf("expected 1 journey update, got %d", len(journey.updates))
	}
}

type mockBalanceStore struct {
	overrides map[string]int64
}

func (m *mockBalanceStore) GetOverride(ctx context.Context, key string) (*balance.Override, error) {
	v, ok := m.overrides[key]
	if !ok {
		return nil, balance.ErrConfigNotFound
	}
	return &balance.Override{Key: key, Value: v}, nil
}

func (m *mockBalanceStore) ListOverrides(ctx context.Context) ([]balance.Override, error) {
	result := make([]balance.Override, 0, len(m.overrides))
	for k, v := range m.overrides {
		result = append(result, balance.Override{Key: k, Value: v})
	}
	return result, nil
}

func (m *mockUserStoreForHandler) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}

type countingRewardGateway struct {
	mu     sync.Mutex
	grants int
}

func (g *countingRewardGateway) GrantQuestReward(ctx context.Context, uid string, questID int64, xp int64) error {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.grants++
	return nil
}

// --- Journey-unlock regression tests (issue: 022 LOCKED rows + COMPLETE journey) ---

// completes the final challenge of quest 1, completing it and finishing the
// whispering-woods journey.
func completeFinalChallengeForQuest1(t *testing.T, h *QuestAPIHandler) *CompleteChallengeResult {
	t.Helper()
	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return result
}

func setupCompleteQuestScenario(t *testing.T, realmProgress map[string]*game.JourneyProgress) (*QuestAPIHandler, *mockRealmProgressStoreForHandler) {
	t.Helper()
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.missions[1] = q
	qStore.questsByCrew["crew-1"] = []game.Mission{*qStore.missions[1]}
	qStore.exercises[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	journey := newMockRealmStore()
	journey.progress = realmProgress

	h := NewQuestAPIHandler(NewQuestService(qStore),
		progression.NewProgressionService(&mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}, &defaultProgCfg),
		journey, defaultRealmCfg, &defaultProgCfg)
	return h, journey
}

func TestCompleteChallenge_UnlocksExistingLockedRealm(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	realmProgress := map[string]*game.JourneyProgress{
		"crew-1|whispering-woods": {FamilyID: "crew-1", Journey: "whispering-woods", Status: "COMPLETE", Progress: 100, LastUnlockedAt: now},
		"crew-1|clockwork-city":   {FamilyID: "crew-1", Journey: "clockwork-city", Status: "LOCKED", Progress: 0},
	}
	h, journey := setupCompleteQuestScenario(t, realmProgress)

	result := completeFinalChallengeForQuest1(t, h)
	if !result.QuestCompleted {
		t.Fatal("expected quest to be completed")
	}

	cw := journey.progress["crew-1|clockwork-city"]
	if cw == nil {
		t.Fatal("expected clockwork-city journey row to exist")
	}
	if cw.Status != "ACTIVE" {
		t.Errorf("expected clockwork-city ACTIVE (locked row activated), got %s", cw.Status)
	}
	updates := map[string]bool{}
	for _, u := range journey.updates {
		updates[u] = true
	}
	if !updates["crew-1|clockwork-city"] {
		t.Error("expected an update to the existing LOCKED clockwork-city row, not an insert")
	}
}

func TestCompleteChallenge_CreatesMissingNextRealm(t *testing.T) {
	realmProgress := map[string]*game.JourneyProgress{
		"crew-1|whispering-woods": {FamilyID: "crew-1", Journey: "whispering-woods", Status: "COMPLETE", Progress: 100},
	}
	h, journey := setupCompleteQuestScenario(t, realmProgress)

	result := completeFinalChallengeForQuest1(t, h)
	if !result.QuestCompleted {
		t.Fatal("expected quest to be completed")
	}

	cw := journey.progress["crew-1|clockwork-city"]
	if cw == nil {
		t.Fatal("expected clockwork-city journey row to be created")
	}
	if cw.Status != "ACTIVE" {
		t.Errorf("expected clockwork-city ACTIVE, got %s", cw.Status)
	}
}

// TestCompleteChallenge_ConcurrentExactlyOnceReward verifies that two
// concurrent completions of the FINAL challenge — the classic double-tap race —
// award challenge XP, the quest-completion reward, and the QuestCompleted
// event exactly once. The loser is treated as a replay (no XP, no reward).
func TestCompleteChallenge_ConcurrentExactlyOnceReward(t *testing.T) {
	store := newConcurrentQuestStore()
	now := time.Now().UTC()
	q := &game.Mission{
		ID:           1,
		FamilyID:       "crew-1",
		TemplateSlug: "morning-light",
		Title:        "Morning Light",
		Status:       string(QuestStatusActive),
		StartedAt:    &now,
		CreatedAt:    now,
	}
	store.missions[1] = q
	store.exercises[1] = []game.Exercise{
		{ID: 11, MissionID: 1, Slug: "find-the-dew", Status: string(ChallengeStatusDone), CreatedAt: now},
		{ID: 12, MissionID: 1, Slug: "morning-fact", Status: string(ChallengeStatusPending), CreatedAt: now},
	}

	pub := &capturePublisher{}
	rewards := &countingRewardGateway{}
	prog := progression.NewProgressionService(
		&concurrentUserStore{player: &game.Player{UID: "user-1", FamilyID: "crew-1", Level: 1, XP: 0, Version: 1}},
		&defaultProgCfg,
	)
	journey := newConcurrentRealmStore()
	_, _ = journey.CreateRealmProgress(context.Background(), &game.JourneyProgress{
		FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 0,
	})
	h := NewQuestAPIHandler(NewQuestService(store), prog, journey, defaultRealmCfg, &defaultProgCfg)
	h.SetPublisher(pub)
	h.SetRewardService(rewards)

	var wg sync.WaitGroup
	results := make([]*CompleteChallengeResult, 2)
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("call %d error: %v", i, err)
		}
	}

	questCompletedEvents := 0
	for _, ev := range pub.published {
		if _, ok := ev.(events.QuestCompletedEvent); ok {
			questCompletedEvents++
		}
	}
	if questCompletedEvents != 1 {
		t.Errorf("expected exactly 1 QuestCompletedEvent, got %d", questCompletedEvents)
	}

	if rewards.grants != 1 {
		t.Errorf("expected exactly 1 reward grant, got %d", rewards.grants)
	}

	totalXP := results[0].XP + results[1].XP
	expected := ChallengeXP + CompletionBonusXP
	if totalXP != expected {
		t.Errorf("expected total XP %d (challenge+bounty), got %d", expected, totalXP)
	}

	quest, _ := store.GetQuest(context.Background(), 1)
	if quest.Status != string(QuestStatusDone) {
		t.Errorf("expected quest DONE, got %s", quest.Status)
	}
}

func TestSelectBranch_Valid_PersistsBranch(t *testing.T) {
	qsStore := newMockQuestStore()
	qsStore.missions[1] = &game.Mission{
		ID:           1,
		FamilyID:       "crew-1",
		TemplateSlug: "riddle-of-the-stones",
		Title:        "Riddle of the Stones",
		Status:       "ACTIVE",
	}

	realmStore := &mockRealmProgressStoreForHandler{
		progress: map[string]*game.JourneyProgress{
			"crew-1|whispering-woods": {FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 10},
		},
	}

	svc := NewQuestService(qsStore)
	handler := NewQuestAPIHandler(svc, nil, realmStore, nil, nil)

	ctx := context.Background()
	res, err := handler.SelectBranch(ctx, 1, "crew-1", "path-of-echoes")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !res.Success || res.StoryBranch != "path-of-echoes" || res.Journey != "whispering-woods" {
		t.Errorf("unexpected res: %+v", res)
	}

	if realmStore.progress["crew-1|whispering-woods"].StoryBranch != "path-of-echoes" {
		t.Errorf("expected story_branch 'path-of-echoes', got %s", realmStore.progress["crew-1|whispering-woods"].StoryBranch)
	}
}

func TestSelectBranch_InvalidChoice_Rejected(t *testing.T) {
	qsStore := newMockQuestStore()
	qsStore.missions[1] = &game.Mission{
		ID:           1,
		FamilyID:       "crew-1",
		TemplateSlug: "riddle-of-the-stones",
		Title:        "Riddle of the Stones",
		Status:       "ACTIVE",
	}

	realmStore := &mockRealmProgressStoreForHandler{
		progress: map[string]*game.JourneyProgress{
			"crew-1|whispering-woods": {FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 10},
		},
	}

	svc := NewQuestService(qsStore)
	handler := NewQuestAPIHandler(svc, nil, realmStore, nil, nil)

	ctx := context.Background()
	_, err := handler.SelectBranch(ctx, 1, "crew-1", "arbitrary-hacked-branch")
	if err == nil || err != ErrInvalidBranchChoice {
		t.Fatalf("expected ErrInvalidBranchChoice, got %v", err)
	}
}

func TestSelectBranch_LinearQuest_NoBranch(t *testing.T) {
	qsStore := newMockQuestStore()
	qsStore.missions[1] = &game.Mission{
		ID:           1,
		FamilyID:       "crew-1",
		TemplateSlug: "morning-light",
		Title:        "Morning Light",
		Status:       "ACTIVE",
	}

	realmStore := &mockRealmProgressStoreForHandler{
		progress: map[string]*game.JourneyProgress{
			"crew-1|whispering-woods": {FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 10},
		},
	}

	svc := NewQuestService(qsStore)
	handler := NewQuestAPIHandler(svc, nil, realmStore, nil, nil)

	ctx := context.Background()
	_, err := handler.SelectBranch(ctx, 1, "crew-1", "path-of-echoes")
	if err == nil || err != ErrNoBranchOptions {
		t.Fatalf("expected ErrNoBranchOptions, got %v", err)
	}
}

func TestListAvailable_MapsDefinitionsToViews(t *testing.T) {
	qStore := newMockQuestStore()
	cg := &mockContentGatewayForHandler{
		missions: []gamecontent.QuestDefinition{
			{Slug: "available-1", Title: "Avail 1", QuestType: "SOLO", ChallengeDefs: []gamecontent.ChallengeDef{{Slug: "c1"}}, CreatedAt: time.Now()},
			{Slug: "available-2", Title: "Avail 2", QuestType: "RELAY", ChallengeDefs: []gamecontent.ChallengeDef{{Slug: "c1"}, {Slug: "c2"}}, CreatedAt: time.Now()},
		},
	}
	qs := NewQuestServiceWithGate(qStore, &mockQuestGate{seasonActive: true, chapterUnlocked: true, realmActive: true, playerLevel: 10, questCompleted: true}, cg)
	h := NewQuestAPIHandler(qs, nil, newMockRealmStore(), defaultRealmCfg, &defaultProgCfg)

	views, err := h.ListAvailable(context.Background(), "crew-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 available missions, got %d", len(views))
	}
	if views[0].TemplateSlug != "available-1" {
		t.Errorf("expected first quest available-1, got %s", views[0].TemplateSlug)
	}
	if views[0].Status != string(QuestStatusPending) {
		t.Errorf("expected PENDING status, got %s", views[0].Status)
	}
	if views[0].ChallengeCount != 1 {
		t.Errorf("expected 1 challenge, got %d", views[0].ChallengeCount)
	}
	if views[1].ChallengeCount != 2 {
		t.Errorf("expected 2 exercises, got %d", views[1].ChallengeCount)
	}
}
