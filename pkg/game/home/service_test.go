package home

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/crewstreak"
	"odyssey/pkg/game/dailyturn"
	"odyssey/pkg/game/quest"
	"odyssey/pkg/game/season"
)

type mockDailyTurnStore struct {
	turns []game.DailyTurn
	err   error
}

func (m *mockDailyTurnStore) CreateDailyTurn(ctx context.Context, dt *game.DailyTurn) (*game.DailyTurn, error) {
	return dt, m.err
}
func (m *mockDailyTurnStore) UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error {
	return m.err
}
func (m *mockDailyTurnStore) ListDailyTurns(ctx context.Context, uid string) ([]game.DailyTurn, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.turns, nil
}

type mockUserStore struct {
	player      *game.Player
	crewMembers []game.Player
	err         error
}

func (m *mockUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.player, nil
}
func (m *mockUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	return m.err
}
func (m *mockUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return m.err
}
func (m *mockUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	return false, m.err
}

type mockQuestStore struct {
	quests     []game.Quest
	challenges map[int64][]game.Challenge
	err        error
}

func (m *mockQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, q := range m.quests {
		if q.ID == questID {
			return &q, nil
		}
	}
	return nil, game.ErrNotFound
}
func (m *mockQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	return q, m.err
}
func (m *mockQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quests, nil
}
func (m *mockQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	return m.err
}
func (m *mockQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	return false, m.err
}
func (m *mockQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.challenges[questID], nil
}
func (m *mockQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	return c, m.err
}
func (m *mockQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	return m.err
}
func (m *mockQuestStore) UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error) {
	return false, m.err
}

type mockProgressionStore struct {
	relicCount int
	err        error
}

func (m *mockProgressionStore) CreateRelic(ctx context.Context, r *game.Relic) (*game.Relic, error) {
	return r, m.err
}
func (m *mockProgressionStore) CreateChest(ctx context.Context, ch *game.Chest) (*game.Chest, error) {
	return ch, m.err
}
func (m *mockProgressionStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	return m.err
}
func (m *mockProgressionStore) UpdateChestIfMatch(ctx context.Context, chestID int64, oldOpened bool, patch map[string]any) (bool, error) {
	return false, m.err
}
func (m *mockProgressionStore) CreateAchievement(ctx context.Context, a *game.Achievement) (*game.Achievement, error) {
	return a, m.err
}
func (m *mockProgressionStore) CountRelics(ctx context.Context, uid string) (int, error) {
	return m.relicCount, m.err
}

type mockRealmProgressStore struct {
	progress []game.RealmProgress
	err      error
}

func (m *mockRealmProgressStore) GetRealmProgress(ctx context.Context, crewID, realm string) (*game.RealmProgress, error) {
	return nil, m.err
}
func (m *mockRealmProgressStore) CreateRealmProgress(ctx context.Context, rp *game.RealmProgress) (*game.RealmProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return rp, nil
}
func (m *mockRealmProgressStore) UpdateRealmProgress(ctx context.Context, crewID, realm string, patch map[string]any) error {
	return m.err
}
func (m *mockRealmProgressStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, realm string, oldProgress int, patch map[string]any) (bool, error) {
	return false, m.err
}
func (m *mockRealmProgressStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.RealmProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.progress, nil
}

type mockCreativeSubmissionStore struct {
	subs []game.Submission
	err  error
}

func (m *mockCreativeSubmissionStore) CreateSubmission(ctx context.Context, s *game.Submission) (*game.Submission, error) {
	return s, m.err
}
func (m *mockCreativeSubmissionStore) ListByQuest(ctx context.Context, questID int64) ([]game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.subs, nil
}
func (m *mockCreativeSubmissionStore) ListByCrew(ctx context.Context, crewID string) ([]game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.subs, nil
}
func (m *mockCreativeSubmissionStore) ListByCrewAndKind(ctx context.Context, crewID, kind string) ([]game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.subs, nil
}
func (m *mockCreativeSubmissionStore) GetSubmission(ctx context.Context, submissionID int64) (*game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return nil, game.ErrNotFound
}
func (m *mockCreativeSubmissionStore) UpdateSubmission(ctx context.Context, submissionID int64, patch map[string]any) error {
	return m.err
}

func TestGetHome_Success(t *testing.T) {
	now := time.Now().UTC()
	completedAt := now
	player := &game.Player{UID: "u1", CrewID: "c1", ExplorerName: "Alice", Level: 3, XP: 100}
	qStore := &mockQuestStore{
		quests: []game.Quest{
			{ID: 1, CrewID: "c1", Title: "Quest A", Status: "PENDING", CreatedAt: now},
			{ID: 2, CrewID: "c1", Title: "Quest B", Status: "DONE", CompletedAt: &completedAt, CreatedAt: now},
		},
		challenges: map[int64][]game.Challenge{
			1: {
				{ID: 10, Status: "DONE"},
				{ID: 11, Status: "PENDING"},
			},
		},
	}
	progStore := &mockProgressionStore{relicCount: 5}
	realmStore := &mockRealmProgressStore{
		progress: []game.RealmProgress{
			{CrewID: "c1", Realm: "forest", Status: "ACTIVE", Progress: 2},
		},
	}

	qs := quest.NewQuestService(qStore)
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, progStore, realmStore, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Player.UID != "u1" {
		t.Errorf("expected player UID u1, got %s", resp.Player.UID)
	}
	if len(resp.Quests) != 2 {
		t.Fatalf("expected 2 quests, got %d", len(resp.Quests))
	}
	if resp.Quests[0].ChallengeCount != 2 {
		t.Errorf("expected 2 challenges, got %d", resp.Quests[0].ChallengeCount)
	}
	if resp.Quests[0].CompletedCount != 1 {
		t.Errorf("expected 1 completed, got %d", resp.Quests[0].CompletedCount)
	}
	if resp.DailyTurn.Today != dailyturn.TodayDate() {
		t.Errorf("expected today date, got %s", resp.DailyTurn.Today)
	}
	if resp.RelicCount != 5 {
		t.Errorf("expected 5 relics, got %d", resp.RelicCount)
	}
	if len(resp.RealmProgress) != 1 {
		t.Errorf("expected 1 realm progress, got %d", len(resp.RealmProgress))
	}
	if resp.DailyTurn.RemainingTurns != 1 {
		t.Errorf("expected remaining turns 1, got %d", resp.DailyTurn.RemainingTurns)
	}
	if len(resp.ActiveQuests) != 1 {
		t.Errorf("expected 1 active quest, got %d", len(resp.ActiveQuests))
	}
	if resp.ActiveQuests[0].Title != "Quest A" {
		t.Errorf("expected active quest Quest A, got %s", resp.ActiveQuests[0].Title)
	}
	if len(resp.CompletedQuestsToday) != 1 {
		t.Errorf("expected 1 completed today, got %d", len(resp.CompletedQuestsToday))
	}

	if resp.Sections.Player.UID != "u1" {
		t.Errorf("expected sections player UID u1, got %s", resp.Sections.Player.UID)
	}
	if resp.Sections.Player.XPToNext != 1400 {
		t.Errorf("expected xp_to_next 1400, got %d", resp.Sections.Player.XPToNext)
	}
	if len(resp.Sections.Quests.All) != 2 {
		t.Errorf("expected 2 quests in sections.all, got %d", len(resp.Sections.Quests.All))
	}
	if len(resp.Sections.Quests.Active) != 1 {
		t.Errorf("expected 1 active quest in sections.active, got %d", len(resp.Sections.Quests.Active))
	}
	if len(resp.Sections.Quests.Done) != 1 {
		t.Errorf("expected 1 done quest in sections.done, got %d", len(resp.Sections.Quests.Done))
	}
	if len(resp.Sections.Quests.DoneToday) != 1 {
		t.Errorf("expected 1 done-today in sections.done_today, got %d", len(resp.Sections.Quests.DoneToday))
	}
	if resp.Sections.DailyTurn.Today != resp.DailyTurn.Today {
		t.Errorf("sections daily_turn.today mismatch")
	}
	if resp.Sections.DailyTurn.StreakDays != resp.DailyTurn.StreakDays {
		t.Errorf("sections daily_turn.streak_days mismatch")
	}
	if len(resp.Sections.Realm.Progress) != 1 {
		t.Errorf("expected 1 realm progress in sections.realm.progress, got %d", len(resp.Sections.Realm.Progress))
	}
	if len(resp.AvailableChests) != 0 {
		t.Errorf("expected 0 available chests, got %d", len(resp.AvailableChests))
	}
}

type mockChestStore struct{}

func (m *mockChestStore) CreateChest(ctx context.Context, ch *game.Chest) (*game.Chest, error) {
	return ch, nil
}
func (m *mockChestStore) GetChest(ctx context.Context, chestID int64) (*game.Chest, error) {
	return nil, game.ErrNotFound
}
func (m *mockChestStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	return nil
}
func (m *mockChestStore) UpdateChestIfMatch(ctx context.Context, chestID int64, oldOpened bool, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockChestStore) ListChestsByUser(ctx context.Context, uid string) ([]game.Chest, error) {
	return nil, nil
}

func TestGetHome_SectionsBackwardCompat(t *testing.T) {
	player := &game.Player{UID: "u1", CrewID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Player.ExplorerName != "Alice" {
		t.Errorf("backward compat: expected player name Alice, got %s", resp.Player.ExplorerName)
	}
	if resp.Sections.Player.ExplorerName != "Alice" {
		t.Errorf("sections: expected player name Alice, got %s", resp.Sections.Player.ExplorerName)
	}
}

func TestGetHome_UserError(t *testing.T) {
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{
		err: errors.New("db error"),
	}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)
	_, err := svc.GetHome(context.Background(), "u1", "c1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetHome_QuestError(t *testing.T) {
	player := &game.Player{UID: "u1", CrewID: "c1"}
	qs := quest.NewQuestService(&mockQuestStore{err: errors.New("quest error")})
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)
	_, err := svc.GetHome(context.Background(), "u1", "c1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetHome_NilChapterLoreAchievementServices(t *testing.T) {
	player := &game.Player{UID: "u1", CrewID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.ChapterProgress != nil {
		t.Error("expected nil chapter progress when service not set")
	}
	if resp.LoreSummary != nil {
		t.Error("expected nil lore summary when service not set")
	}
	if resp.Achievements != nil {
		t.Error("expected nil achievements when service not set")
	}
	if resp.Sections.World.CurrentChapter != nil {
		t.Error("expected nil current chapter when service not set")
	}
	if resp.Sections.Lore.Summary != nil {
		t.Error("expected nil lore summary in section when service not set")
	}
}
func (m *mockUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	if m.crewMembers != nil {
		return m.crewMembers, nil
	}
	return nil, nil
}

type mockActivityStore struct {
	acts []game.DailyActivity
	err  error
}

func (m *mockActivityStore) RecordActivity(ctx context.Context, act *game.DailyActivity) (*game.DailyActivity, error) {
	return act, m.err
}

func (m *mockActivityStore) GetStreak(ctx context.Context, uid string) (int, error) {
	return 0, m.err
}

func (m *mockActivityStore) ListActivityDatesByUsers(ctx context.Context, uids []string) ([]game.DailyActivity, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.acts, nil
}

func TestGetHome_CrewStreak(t *testing.T) {
	player := &game.Player{UID: "u1", CrewID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)
	activity := &mockActivityStore{
		acts: []game.DailyActivity{
			{UserID: "u1", ActivityDate: dailyturn.TodayDate()},
			{UserID: "u2", ActivityDate: dailyturn.TodayDate()},
		},
	}
	userStore := &mockUserStore{player: player, crewMembers: []game.Player{{UID: "u1", CrewID: "c1"}, {UID: "u2", CrewID: "c1"}}}
	cs := crewstreak.NewService(userStore, activity, "UTC")
	svc.SetCrewStreakService(cs)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DailyTurn.CrewStreak != 1 {
		t.Errorf("expected crew streak 1, got %d", resp.DailyTurn.CrewStreak)
	}
	if resp.Sections.DailyTurn.CrewStreak != resp.DailyTurn.CrewStreak {
		t.Errorf("sections daily_turn.crew_streak mismatch")
	}
}

func TestGetHome_CrewStreakNilService(t *testing.T) {
	player := &game.Player{UID: "u1", CrewID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DailyTurn.CrewStreak != 0 {
		t.Errorf("expected crew streak 0 when service not set, got %d", resp.DailyTurn.CrewStreak)
	}
}

type mockSeasonService struct {
	summary *season.SeasonSummary
	err     error
}

func (m *mockSeasonService) IsActive(ctx context.Context, slug string) bool {
	return true
}

func (m *mockSeasonService) GetState(ctx context.Context, slug string) (season.SeasonState, error) {
	return season.SeasonStateActive, nil
}

func (m *mockSeasonService) GetCurrentSeason(ctx context.Context) (*season.SeasonSummary, error) {
	return m.summary, m.err
}

func (m *mockSeasonService) ListAll(ctx context.Context) ([]season.SeasonSummary, error) {
	return nil, m.err
}

type mockSeasonGateway struct {
	seasons []gamecontent.SeasonDefinition
	err     error
}

func (m *mockSeasonGateway) ListSeasons(ctx context.Context) ([]gamecontent.SeasonDefinition, error) {
	return m.seasons, m.err
}

func (m *mockSeasonGateway) GetSeason(ctx context.Context, slug string) (*gamecontent.SeasonDefinition, error) {
	if m.err != nil {
		return nil, m.err
	}
	for i := range m.seasons {
		if m.seasons[i].Slug == slug {
			return &m.seasons[i], nil
		}
	}
	return nil, nil
}

func TestGetHome_CurrentSeason(t *testing.T) {
	now := time.Now().UTC()
	player := &game.Player{UID: "u1", CrewID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	seasonSvc := season.NewSeasonService(&mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			{Slug: "season-spring-2026", Name: "Spring 2026", Realm: "whispering-woods", StartAt: now.Add(-24 * time.Hour), EndAt: now.Add(24 * time.Hour), Published: true},
		},
	}, nil)
	svc.SetSeasonService(seasonSvc)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CurrentSeason == nil {
		t.Fatal("expected current season to be set")
	}
	if resp.CurrentSeason.Definition.Slug != "season-spring-2026" {
		t.Errorf("expected season-spring-2026, got %s", resp.CurrentSeason.Definition.Slug)
	}
	if resp.SeasonProgress.SeasonSlug != "season-spring-2026" {
		t.Errorf("expected season_progress.season_slug season-spring-2026, got %s", resp.SeasonProgress.SeasonSlug)
	}
}

func TestGetHome_SeasonProgressQuestsCompleted(t *testing.T) {
	now := time.Now().UTC()
	player := &game.Player{UID: "u1", CrewID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}

	qStore := &mockQuestStore{
		quests: []game.Quest{
			{ID: 1, CrewID: "c1", TemplateSlug: "seasonal-q1", Title: "SQ1", Status: string(quest.QuestStatusDone), CompletedAt: &now, CreatedAt: now},
			{ID: 2, CrewID: "c1", TemplateSlug: "plain-q", Title: "Plain", Status: string(quest.QuestStatusDone), CompletedAt: &now, CreatedAt: now},
		},
		challenges: map[int64][]game.Challenge{},
	}
	content := &mockContentGatewayForHome{
		quests: []gamecontent.QuestDefinition{
			{Slug: "seasonal-q1", SeasonSlug: "season-spring-2026"},
			{Slug: "plain-q", SeasonSlug: ""},
		},
	}
	qs := quest.NewQuestServiceWithGate(qStore, nil, content)
	dts := dailyturn.NewDailyTurnService(&mockDailyTurnStore{}, &dailyturn.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	seasonSvc := season.NewSeasonService(&mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			{Slug: "season-spring-2026", Name: "Spring 2026", Realm: "whispering-woods", StartAt: now.Add(-24 * time.Hour), EndAt: now.Add(24 * time.Hour), Published: true},
		},
	}, nil)
	svc.SetSeasonService(seasonSvc)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SeasonProgress.QuestsCompleted != 1 {
		t.Errorf("expected 1 quests_completed for season, got %d", resp.SeasonProgress.QuestsCompleted)
	}
}

type mockContentGatewayForHome struct {
	quests []gamecontent.QuestDefinition
}

func (m *mockContentGatewayForHome) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.quests, nil
}
func (m *mockContentGatewayForHome) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	for _, q := range m.quests {
		if q.Slug == slug {
			return &q, nil
		}
	}
	return nil, nil
}
func (m *mockContentGatewayForHome) ListQuestsByRealm(ctx context.Context, realm string) ([]gamecontent.QuestDefinition, error) {
	return nil, nil
}
