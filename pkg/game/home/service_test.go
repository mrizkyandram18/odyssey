package home

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/familystreak"
	"odyssey/pkg/game/dailymission"
	"odyssey/pkg/game/mission"
	"odyssey/pkg/game/season"
)

type mockDailyTurnStore struct {
	turns []game.DailyMission
	err   error
}

func (m *mockDailyTurnStore) CreateDailyTurn(ctx context.Context, dt *game.DailyMission) (*game.DailyMission, error) {
	return dt, m.err
}
func (m *mockDailyTurnStore) UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error {
	return m.err
}
func (m *mockDailyTurnStore) ListDailyTurns(ctx context.Context, uid string) ([]game.DailyMission, error) {
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
	missions     []game.Mission
	exercises map[int64][]game.Exercise
	err        error
}

func (m *mockQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Mission, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, q := range m.missions {
		if q.ID == questID {
			return &q, nil
		}
	}
	return nil, game.ErrNotFound
}
func (m *mockQuestStore) CreateQuest(ctx context.Context, q *game.Mission) (*game.Mission, error) {
	return q, m.err
}
func (m *mockQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Mission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.missions, nil
}
func (m *mockQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	return m.err
}
func (m *mockQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	return false, m.err
}
func (m *mockQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Exercise, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.exercises[questID], nil
}
func (m *mockQuestStore) CreateChallenge(ctx context.Context, c *game.Exercise) (*game.Exercise, error) {
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

func (m *mockProgressionStore) CreateRelic(ctx context.Context, r *game.Collection) (*game.Collection, error) {
	return r, m.err
}
func (m *mockProgressionStore) CreateChest(ctx context.Context, ch *game.Gift) (*game.Gift, error) {
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
	progress []game.JourneyProgress
	err      error
}

func (m *mockRealmProgressStore) GetRealmProgress(ctx context.Context, crewID, journey string) (*game.JourneyProgress, error) {
	return nil, m.err
}
func (m *mockRealmProgressStore) CreateRealmProgress(ctx context.Context, rp *game.JourneyProgress) (*game.JourneyProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return rp, nil
}
func (m *mockRealmProgressStore) UpdateRealmProgress(ctx context.Context, crewID, journey string, patch map[string]any) error {
	return m.err
}
func (m *mockRealmProgressStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, journey string, oldProgress int, patch map[string]any) (bool, error) {
	return false, m.err
}
func (m *mockRealmProgressStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.JourneyProgress, error) {
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
	player := &game.Player{UID: "u1", FamilyID: "c1", ExplorerName: "Alice", Level: 3, XP: 100}
	qStore := &mockQuestStore{
		missions: []game.Mission{
			{ID: 1, FamilyID: "c1", Title: "Mission A", Status: "PENDING", CreatedAt: now},
			{ID: 2, FamilyID: "c1", Title: "Mission B", Status: "DONE", CompletedAt: &completedAt, CreatedAt: now},
		},
		exercises: map[int64][]game.Exercise{
			1: {
				{ID: 10, Status: "DONE"},
				{ID: 11, Status: "PENDING"},
			},
		},
	}
	progStore := &mockProgressionStore{relicCount: 5}
	realmStore := &mockRealmProgressStore{
		progress: []game.JourneyProgress{
			{FamilyID: "c1", Journey: "forest", Status: "ACTIVE", Progress: 2},
		},
	}

	qs := quest.NewQuestService(qStore)
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, progStore, realmStore, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Player.UID != "u1" {
		t.Errorf("expected player UID u1, got %s", resp.Player.UID)
	}
	if len(resp.Missions) != 2 {
		t.Fatalf("expected 2 missions, got %d", len(resp.Missions))
	}
	if resp.Missions[0].ChallengeCount != 2 {
		t.Errorf("expected 2 exercises, got %d", resp.Missions[0].ChallengeCount)
	}
	if resp.Missions[0].CompletedCount != 1 {
		t.Errorf("expected 1 completed, got %d", resp.Missions[0].CompletedCount)
	}
	if resp.DailyMission.Today != dailymission.TodayDate() {
		t.Errorf("expected today date, got %s", resp.DailyMission.Today)
	}
	if resp.RelicCount != 5 {
		t.Errorf("expected 5 collections, got %d", resp.RelicCount)
	}
	if len(resp.JourneyProgress) != 1 {
		t.Errorf("expected 1 journey progress, got %d", len(resp.JourneyProgress))
	}
	if resp.DailyMission.RemainingTurns != 1 {
		t.Errorf("expected remaining turns 1, got %d", resp.DailyMission.RemainingTurns)
	}
	if len(resp.ActiveQuests) != 1 {
		t.Errorf("expected 1 active quest, got %d", len(resp.ActiveQuests))
	}
	if resp.ActiveQuests[0].Title != "Mission A" {
		t.Errorf("expected active quest Mission A, got %s", resp.ActiveQuests[0].Title)
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
	if len(resp.Sections.Missions.All) != 2 {
		t.Errorf("expected 2 missions in sections.all, got %d", len(resp.Sections.Missions.All))
	}
	if len(resp.Sections.Missions.Active) != 1 {
		t.Errorf("expected 1 active quest in sections.active, got %d", len(resp.Sections.Missions.Active))
	}
	if len(resp.Sections.Missions.Done) != 1 {
		t.Errorf("expected 1 done quest in sections.done, got %d", len(resp.Sections.Missions.Done))
	}
	if len(resp.Sections.Missions.DoneToday) != 1 {
		t.Errorf("expected 1 done-today in sections.done_today, got %d", len(resp.Sections.Missions.DoneToday))
	}
	if resp.Sections.DailyMission.Today != resp.DailyMission.Today {
		t.Errorf("sections daily_mission.today mismatch")
	}
	if resp.Sections.DailyMission.StreakDays != resp.DailyMission.StreakDays {
		t.Errorf("sections daily_mission.streak_days mismatch")
	}
	if len(resp.Sections.Journey.Progress) != 1 {
		t.Errorf("expected 1 journey progress in sections.journey.progress, got %d", len(resp.Sections.Journey.Progress))
	}
	if len(resp.AvailableChests) != 0 {
		t.Errorf("expected 0 available gifts, got %d", len(resp.AvailableChests))
	}
}

type mockChestStore struct{}

func (m *mockChestStore) CreateChest(ctx context.Context, ch *game.Gift) (*game.Gift, error) {
	return ch, nil
}
func (m *mockChestStore) GetChest(ctx context.Context, chestID int64) (*game.Gift, error) {
	return nil, game.ErrNotFound
}
func (m *mockChestStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	return nil
}
func (m *mockChestStore) UpdateChestIfMatch(ctx context.Context, chestID int64, oldOpened bool, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockChestStore) ListChestsByUser(ctx context.Context, uid string) ([]game.Gift, error) {
	return nil, nil
}

func TestGetHome_SectionsBackwardCompat(t *testing.T) {
	player := &game.Player{UID: "u1", FamilyID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
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
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{
		err: errors.New("db error"),
	}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)
	_, err := svc.GetHome(context.Background(), "u1", "c1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetHome_QuestError(t *testing.T) {
	player := &game.Player{UID: "u1", FamilyID: "c1"}
	qs := quest.NewQuestService(&mockQuestStore{err: errors.New("quest error")})
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)
	_, err := svc.GetHome(context.Background(), "u1", "c1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGetHome_NilChapterLoreAchievementServices(t *testing.T) {
	player := &game.Player{UID: "u1", FamilyID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.CourseProgress != nil {
		t.Error("expected nil course progress when service not set")
	}
	if resp.LoreSummary != nil {
		t.Error("expected nil concept summary when service not set")
	}
	if resp.Achievements != nil {
		t.Error("expected nil achievements when service not set")
	}
	if resp.Sections.World.CurrentChapter != nil {
		t.Error("expected nil current course when service not set")
	}
	if resp.Sections.Concept.Summary != nil {
		t.Error("expected nil concept summary in section when service not set")
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
	player := &game.Player{UID: "u1", FamilyID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)
	activity := &mockActivityStore{
		acts: []game.DailyActivity{
			{UserID: "u1", ActivityDate: dailymission.TodayDate()},
			{UserID: "u2", ActivityDate: dailymission.TodayDate()},
		},
	}
	userStore := &mockUserStore{player: player, crewMembers: []game.Player{{UID: "u1", FamilyID: "c1"}, {UID: "u2", FamilyID: "c1"}}}
	cs := familystreak.NewService(userStore, activity, "UTC")
	svc.SetCrewStreakService(cs)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DailyMission.CrewStreak != 1 {
		t.Errorf("expected crew streak 1, got %d", resp.DailyMission.CrewStreak)
	}
	if resp.Sections.DailyMission.CrewStreak != resp.DailyMission.CrewStreak {
		t.Errorf("sections daily_mission.crew_streak mismatch")
	}
}

func TestGetHome_CrewStreakNilService(t *testing.T) {
	player := &game.Player{UID: "u1", FamilyID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.DailyMission.CrewStreak != 0 {
		t.Errorf("expected crew streak 0 when service not set, got %d", resp.DailyMission.CrewStreak)
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
	player := &game.Player{UID: "u1", FamilyID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}
	qs := quest.NewQuestService(&mockQuestStore{})
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	seasonSvc := season.NewSeasonService(&mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			{Slug: "season-spring-2026", Name: "Spring 2026", Journey: "whispering-woods", StartAt: now.Add(-24 * time.Hour), EndAt: now.Add(24 * time.Hour), Published: true},
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
	player := &game.Player{UID: "u1", FamilyID: "c1", ExplorerName: "Alice", Level: 1, XP: 0}

	qStore := &mockQuestStore{
		missions: []game.Mission{
			{ID: 1, FamilyID: "c1", TemplateSlug: "seasonal-q1", Title: "SQ1", Status: string(quest.QuestStatusDone), CompletedAt: &now, CreatedAt: now},
			{ID: 2, FamilyID: "c1", TemplateSlug: "plain-q", Title: "Plain", Status: string(quest.QuestStatusDone), CompletedAt: &now, CreatedAt: now},
		},
		exercises: map[int64][]game.Exercise{},
	}
	content := &mockContentGatewayForHome{
		missions: []gamecontent.QuestDefinition{
			{Slug: "seasonal-q1", SeasonSlug: "season-spring-2026"},
			{Slug: "plain-q", SeasonSlug: ""},
		},
	}
	qs := quest.NewQuestServiceWithGate(qStore, nil, content)
	dts := dailymission.NewDailyTurnService(&mockDailyTurnStore{}, &dailymission.DailyTurnConfig{})
	svc := NewHomeService(qs, dts, &mockProgressionStore{}, &mockRealmProgressStore{}, &mockUserStore{player: player}, &mockCreativeSubmissionStore{}, &mockChestStore{}, nil)

	seasonSvc := season.NewSeasonService(&mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			{Slug: "season-spring-2026", Name: "Spring 2026", Journey: "whispering-woods", StartAt: now.Add(-24 * time.Hour), EndAt: now.Add(24 * time.Hour), Published: true},
		},
	}, nil)
	svc.SetSeasonService(seasonSvc)

	resp, err := svc.GetHome(context.Background(), "u1", "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.SeasonProgress.QuestsCompleted != 1 {
		t.Errorf("expected 1 missions_completed for season, got %d", resp.SeasonProgress.QuestsCompleted)
	}
}

type mockContentGatewayForHome struct {
	missions []gamecontent.QuestDefinition
}

func (m *mockContentGatewayForHome) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.missions, nil
}
func (m *mockContentGatewayForHome) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	for _, q := range m.missions {
		if q.Slug == slug {
			return &q, nil
		}
	}
	return nil, nil
}
func (m *mockContentGatewayForHome) ListQuestsByRealm(ctx context.Context, journey string) ([]gamecontent.QuestDefinition, error) {
	return nil, nil
}
