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
	progress map[string]*game.RealmProgress
	updates  []string
	err      error
}

func newMockRealmStore() *mockRealmProgressStoreForHandler {
	return &mockRealmProgressStoreForHandler{
		progress: map[string]*game.RealmProgress{
			"crew-1|whispering-woods": {
				CrewID: "crew-1", Realm: "whispering-woods", Status: "ACTIVE", Progress: 10,
			},
		},
		updates: []string{},
	}
}

func (m *mockRealmProgressStoreForHandler) GetRealmProgress(ctx context.Context, crewID, realm string) (*game.RealmProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.progress[crewID+"|"+realm], nil
}
func (m *mockRealmProgressStoreForHandler) CreateRealmProgress(ctx context.Context, rp *game.RealmProgress) (*game.RealmProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.progress[rp.CrewID+"|"+rp.Realm] = rp
	return rp, nil
}
func (m *mockRealmProgressStoreForHandler) UpdateRealmProgress(ctx context.Context, crewID, realm string, patch map[string]any) error {
	m.updates = append(m.updates, crewID+"|"+realm)
	if m.err != nil {
		return m.err
	}
	rp := m.progress[crewID+"|"+realm]
	if rp != nil {
		if v, ok := patch["progress"].(int); ok {
			rp.Progress = v
		}
		if v, ok := patch["status"].(string); ok {
			rp.Status = v
		}
		if v, ok := patch["last_unlocked_at"]; ok && v != nil {
			rp.LastUnlockedAt, _ = patch["last_unlocked_at"].(time.Time)
		}
	}
	return nil
}
func (m *mockRealmProgressStoreForHandler) UpdateRealmProgressIfMatch(ctx context.Context, crewID, realm string, oldProgress int, patch map[string]any) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	rp := m.progress[crewID+"|"+realm]
	if rp == nil {
		return false, game.ErrNotFound
	}
	if rp.Progress != oldProgress {
		return false, nil
	}
	if err := m.UpdateRealmProgress(ctx, crewID, realm, patch); err != nil {
		return false, err
	}
	return true, nil
}
func (m *mockRealmProgressStoreForHandler) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.RealmProgress, error) {
	return nil, m.err
}

func makePlayerForHandler(level int, xp int64) *game.Player {
	return &game.Player{
		UID:          "user-1",
		CrewID:       "crew-1",
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
	realm := newMockRealmStore()
	return NewQuestAPIHandler(qs, prog, realm, defaultRealmCfg, &defaultProgCfg), realm
}

func makeStoredQuest(questID int64, crewID, slug, status string) *game.Quest {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	return &game.Quest{
		ID:           questID,
		CrewID:       crewID,
		TemplateSlug: slug,
		Title:        "Morning Light",
		Status:       status,
		CreatedAt:    now,
	}
}

func makeStoredChallenges(questID int64, statuses ...string) []game.Challenge {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	chs := make([]game.Challenge, 0, len(statuses))
	for i, s := range statuses {
		chs = append(chs, game.Challenge{
			ID:          questID*10 + int64(i+1),
			QuestID:     questID,
			Slug:        "ch-" + string(rune('a'+i)),
			Description: "Challenge " + string(rune('a'+i)),
			Status:      s,
			CreatedAt:   now,
		})
	}
	return chs
}

func TestStartQuest_Succeeds(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.quests[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusPending))
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}

	h, _ := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	if err := h.StartQuest(context.Background(), 1, "crew-1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if qStore.quests[1].Status != string(QuestStatusActive) {
		t.Errorf("expected ACTIVE, got %s", qStore.quests[1].Status)
	}
	if qStore.quests[1].StartedAt == nil {
		t.Error("expected started_at set")
	}
}

func TestStartQuest_NotInCrew(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.quests[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusPending))
	h, _ := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	err := h.StartQuest(context.Background(), 1, "other-crew")
	if err == nil {
		t.Fatal("expected error when starting a quest not in crew")
	}
}

func TestCompleteChallenge_NonLastAwardsChallengeXPOnly(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.quests[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusPending), string(ChallengeStatusPending))

	h, realm := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	result, err := h.CompleteChallenge(context.Background(), 1, 11, "crew-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.QuestCompleted {
		t.Error("expected quest not yet completed")
	}
	if result.XP != ChallengeXP {
		t.Errorf("expected XP %d, got %d", ChallengeXP, result.XP)
	}
	if result.Quest.Status != string(QuestStatusActive) {
		t.Errorf("expected quest ACTIVE, got %s", result.Quest.Status)
	}
	if result.Quest.CompletedAt != nil {
		t.Error("expected completed_at not set for non-completing action")
	}
	if len(realm.updates) != 0 {
		t.Error("expected no realm progress update for non-completing action")
	}
}

func TestCompleteChallenge_LastChallengeCompletesQuest(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.quests[1] = q
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	h, realm := setupOrchestrator(t, qStore, makePlayerForHandler(1, 90))
	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
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
	if result.Quest.Status != string(QuestStatusDone) {
		t.Errorf("expected quest DONE, got %s", result.Quest.Status)
	}
	if result.Quest.CompletedAt == nil {
		t.Error("expected completed_at set")
	}
	if len(realm.updates) != 1 {
		t.Errorf("expected 1 realm progress update, got %d", len(realm.updates))
	}
}

func TestCompleteChallenge_FullLoop(t *testing.T) {
	qStore := newMockQuestStore()
	qStore.quests[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusPending))
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusPending), string(ChallengeStatusPending))

	h, realm := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))

	if err := h.StartQuest(context.Background(), 1, "crew-1"); err != nil {
		t.Fatalf("start: %v", err)
	}
	if qStore.quests[1].Status != string(QuestStatusActive) {
		t.Fatalf("expected ACTIVE after start, got %s", qStore.quests[1].Status)
	}

	r1, err := h.CompleteChallenge(context.Background(), 1, 11, "crew-1", "user-1")
	if err != nil {
		t.Fatalf("complete 1: %v", err)
	}
	if r1.QuestCompleted {
		t.Fatal("quest should not be completed after 1/2 challenges")
	}

	r2, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
	if err != nil {
		t.Fatalf("complete 2: %v", err)
	}
	if !r2.QuestCompleted {
		t.Fatal("quest should be completed after 2/2 challenges")
	}
	if qStore.quests[1].Status != string(QuestStatusDone) {
		t.Errorf("expected DONE after loop, got %s", qStore.quests[1].Status)
	}
	if len(realm.updates) != 1 {
		t.Errorf("expected 1 realm update, got %d", len(realm.updates))
	}
}

func TestCompleteChallenge_UnknownQuest(t *testing.T) {
	qStore := newMockQuestStore()
	h, _ := setupOrchestrator(t, qStore, makePlayerForHandler(1, 0))
	_, err := h.CompleteChallenge(context.Background(), 999, 1, "crew-1", "user-1")
	if err == nil {
		t.Fatal("expected error for unknown quest")
	}
}

func TestRealmForSlug(t *testing.T) {
	if got := RealmForSlug("morning-light"); got != "whispering-woods" {
		t.Errorf("expected whispering-woods, got %s", got)
	}
	if got := RealmForSlug("unknown-slug"); got != "" {
		t.Errorf("expected empty realm for unknown slug, got %s", got)
	}
}

func TestCompleteChallenge_UsesCustomProgressionConfig(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.quests[1] = q
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	customCfg := &progression.ProgressionConfig{
		ChallengeXP:       10,
		CompletionBonusXP: 40,
	}
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, customCfg)
	realm := newMockRealmStore()
	h := NewQuestAPIHandler(NewQuestService(qStore), prog, realm, defaultRealmCfg, customCfg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
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
	qStore.quests[1] = q
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	pub := &capturePublisher{}
	cg := &mockContentGatewayForHandler{
		quests: []gamecontent.QuestDefinition{
			{Slug: "morning-light", Chapter: "ch-1", SeasonSlug: ""},
		},
	}
	qs := NewQuestServiceWithGate(qStore, nil, cg)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 90)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	h := NewQuestAPIHandler(qs, prog, newMockRealmStore(), defaultRealmCfg, &defaultProgCfg)
	h.SetPublisher(pub)
	h.SetContentGateway(cg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
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
	if ev.Chapter != "ch-1" {
		t.Errorf("expected chapter ch-1, got %s", ev.Chapter)
	}
	if ev.Realm != "whispering-woods" {
		t.Errorf("expected realm whispering-woods, got %s", ev.Realm)
	}
}

func TestCompleteChallenge_NoPublisher_NoError(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.quests[1] = q
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 90)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	h := NewQuestAPIHandler(NewQuestService(qStore), prog, newMockRealmStore(), defaultRealmCfg, &defaultProgCfg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
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
	quests []gamecontent.QuestDefinition
}

func (m *mockContentGatewayForHandler) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.quests, nil
}
func (m *mockContentGatewayForHandler) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	for _, q := range m.quests {
		if q.Slug == slug {
			return &q, nil
		}
	}
	return nil, nil
}
func (m *mockContentGatewayForHandler) ListQuestsByRealm(ctx context.Context, realm string) ([]gamecontent.QuestDefinition, error) {
	return nil, nil
}

func TestCompleteChallenge_UsesQuestRewardXP(t *testing.T) {
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.quests[1] = q
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	cg := &mockContentGatewayForHandler{
		quests: []gamecontent.QuestDefinition{
			{Slug: "morning-light", RewardXP: 80},
		},
	}
	qs := NewQuestServiceWithGate(qStore, nil, cg)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	realm := newMockRealmStore()
	h := NewQuestAPIHandler(qs, prog, realm, defaultRealmCfg, &defaultProgCfg)
	h.SetContentGateway(cg)

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
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
	qStore.quests[1] = q
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	cg := &mockContentGatewayForHandler{
		quests: []gamecontent.QuestDefinition{
			{Slug: "morning-light", RewardXP: 100},
		},
	}
	qs := NewQuestServiceWithGate(qStore, nil, cg)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	realm := newMockRealmStore()
	h := NewQuestAPIHandler(qs, prog, realm, defaultRealmCfg, &defaultProgCfg)
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

	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
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
	qStore.quests[1] = makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}

	qs := NewQuestService(qStore)
	userStore := &mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}
	prog := progression.NewProgressionService(userStore, &defaultProgCfg)
	realm := newMockRealmStore()
	realmCfg := world.NewRealmCatalog([]world.RealmDefinition{
		{Slug: "whispering-woods", Name: "Whispering Woods", Order: 1, MaxProgress: 100},
	})
	h := NewQuestAPIHandler(qs, prog, realm, realmCfg, &defaultProgCfg)
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
	if len(realm.updates) != 1 {
		t.Errorf("expected 1 realm update, got %d", len(realm.updates))
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

// --- Realm-unlock regression tests (issue: 022 LOCKED rows + COMPLETE realm) ---

// completes the final challenge of quest 1, completing it and finishing the
// whispering-woods realm.
func completeFinalChallengeForQuest1(t *testing.T, h *QuestAPIHandler) *CompleteChallengeResult {
	t.Helper()
	result, err := h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return result
}

func setupCompleteQuestScenario(t *testing.T, realmProgress map[string]*game.RealmProgress) (*QuestAPIHandler, *mockRealmProgressStoreForHandler) {
	t.Helper()
	qStore := newMockQuestStore()
	q := makeStoredQuest(1, "crew-1", "morning-light", string(QuestStatusActive))
	q.StartedAt = &time.Time{}
	qStore.quests[1] = q
	qStore.questsByCrew["crew-1"] = []game.Quest{*qStore.quests[1]}
	qStore.challenges[1] = makeStoredChallenges(1, string(ChallengeStatusDone), string(ChallengeStatusPending))

	realm := newMockRealmStore()
	realm.progress = realmProgress

	h := NewQuestAPIHandler(NewQuestService(qStore),
		progression.NewProgressionService(&mockUserStoreForHandler{player: makePlayerForHandler(1, 0)}, &defaultProgCfg),
		realm, defaultRealmCfg, &defaultProgCfg)
	return h, realm
}

func TestCompleteChallenge_UnlocksExistingLockedRealm(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	realmProgress := map[string]*game.RealmProgress{
		"crew-1|whispering-woods": {CrewID: "crew-1", Realm: "whispering-woods", Status: "COMPLETE", Progress: 100, LastUnlockedAt: now},
		"crew-1|clockwork-city":   {CrewID: "crew-1", Realm: "clockwork-city", Status: "LOCKED", Progress: 0},
	}
	h, realm := setupCompleteQuestScenario(t, realmProgress)

	result := completeFinalChallengeForQuest1(t, h)
	if !result.QuestCompleted {
		t.Fatal("expected quest to be completed")
	}

	cw := realm.progress["crew-1|clockwork-city"]
	if cw == nil {
		t.Fatal("expected clockwork-city realm row to exist")
	}
	if cw.Status != "ACTIVE" {
		t.Errorf("expected clockwork-city ACTIVE (locked row activated), got %s", cw.Status)
	}
	updates := map[string]bool{}
	for _, u := range realm.updates {
		updates[u] = true
	}
	if !updates["crew-1|clockwork-city"] {
		t.Error("expected an update to the existing LOCKED clockwork-city row, not an insert")
	}
}

func TestCompleteChallenge_CreatesMissingNextRealm(t *testing.T) {
	realmProgress := map[string]*game.RealmProgress{
		"crew-1|whispering-woods": {CrewID: "crew-1", Realm: "whispering-woods", Status: "COMPLETE", Progress: 100},
	}
	h, realm := setupCompleteQuestScenario(t, realmProgress)

	result := completeFinalChallengeForQuest1(t, h)
	if !result.QuestCompleted {
		t.Fatal("expected quest to be completed")
	}

	cw := realm.progress["crew-1|clockwork-city"]
	if cw == nil {
		t.Fatal("expected clockwork-city realm row to be created")
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
	q := &game.Quest{
		ID:           1,
		CrewID:       "crew-1",
		TemplateSlug: "morning-light",
		Title:        "Morning Light",
		Status:       string(QuestStatusActive),
		StartedAt:    &now,
		CreatedAt:    now,
	}
	store.quests[1] = q
	store.challenges[1] = []game.Challenge{
		{ID: 11, QuestID: 1, Slug: "find-the-dew", Status: string(ChallengeStatusDone), CreatedAt: now},
		{ID: 12, QuestID: 1, Slug: "morning-fact", Status: string(ChallengeStatusPending), CreatedAt: now},
	}

	pub := &capturePublisher{}
	rewards := &countingRewardGateway{}
	prog := progression.NewProgressionService(
		&concurrentUserStore{player: &game.Player{UID: "user-1", CrewID: "crew-1", Level: 1, XP: 0, Version: 1}},
		&defaultProgCfg,
	)
	realm := newConcurrentRealmStore()
	_, _ = realm.CreateRealmProgress(context.Background(), &game.RealmProgress{
		CrewID: "crew-1", Realm: "whispering-woods", Status: "ACTIVE", Progress: 0,
	})
	h := NewQuestAPIHandler(NewQuestService(store), prog, realm, defaultRealmCfg, &defaultProgCfg)
	h.SetPublisher(pub)
	h.SetRewardService(rewards)

	var wg sync.WaitGroup
	results := make([]*CompleteChallengeResult, 2)
	errs := make([]error, 2)
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1")
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
