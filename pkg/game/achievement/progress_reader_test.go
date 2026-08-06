package achievement

import (
	"context"
	"errors"
	"testing"

	"odyssey/pkg/game"
)

type mockPRQuestStore struct {
	quests []game.Quest
	err    error
}

func (m *mockPRQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	return nil, game.ErrNotFound
}
func (m *mockPRQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	return q, nil
}
func (m *mockPRQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quests, nil
}
func (m *mockPRQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	return nil
}
func (m *mockPRQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockPRQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	return nil, nil
}
func (m *mockPRQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	return c, nil
}
func (m *mockPRQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	return nil
}

type mockPRRealmStore struct {
	progress []game.RealmProgress
	err      error
}

func (m *mockPRRealmStore) GetRealmProgress(ctx context.Context, crewID, realm string) (*game.RealmProgress, error) {
	return nil, game.ErrNotFound
}
func (m *mockPRRealmStore) CreateRealmProgress(ctx context.Context, rp *game.RealmProgress) (*game.RealmProgress, error) {
	return rp, nil
}
func (m *mockPRRealmStore) UpdateRealmProgress(ctx context.Context, crewID, realm string, patch map[string]any) error {
	return nil
}
func (m *mockPRRealmStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, realm string, oldProgress int, patch map[string]any) (bool, error) {
	return true, nil
}
func (m *mockPRRealmStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.RealmProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.progress, nil
}

type mockPRUserStore struct {
	player *game.Player
	err    error
}

func (m *mockPRUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	return m.player, m.err
}
func (m *mockPRUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	return m.err
}
func (m *mockPRUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return m.err
}
func (m *mockPRUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	return false, m.err
}

type mockPRRelicStore struct {
	relocs []game.PlayerRelic
	count  int
	err    error
}

func (m *mockPRRelicStore) GetPlayerRelic(ctx context.Context, uid, relicSlug string) (*game.PlayerRelic, error) {
	return nil, game.ErrNotFound
}
func (m *mockPRRelicStore) CreatePlayerRelic(ctx context.Context, pr *game.PlayerRelic) (*game.PlayerRelic, error) {
	return pr, nil
}
func (m *mockPRRelicStore) UpdatePlayerRelic(ctx context.Context, uid, relicSlug string, patch map[string]any) error {
	return nil
}
func (m *mockPRRelicStore) ListPlayerRelics(ctx context.Context, uid string) ([]game.PlayerRelic, error) {
	return m.relocs, m.err
}
func (m *mockPRRelicStore) CountUniqueRelics(ctx context.Context, uid string) (int, error) {
	return m.count, m.err
}

type mockPRDailyStore struct {
	turns []game.DailyTurn
	err   error
}

func (m *mockPRDailyStore) CreateDailyTurn(ctx context.Context, dt *game.DailyTurn) (*game.DailyTurn, error) {
	return dt, nil
}
func (m *mockPRDailyStore) UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error {
	return nil
}
func (m *mockPRDailyStore) ListDailyTurns(ctx context.Context, uid string) ([]game.DailyTurn, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.turns, nil
}

type mockPRCreativeStore struct {
	subs []game.Submission
	err  error
}

func (m *mockPRCreativeStore) CreateSubmission(ctx context.Context, s *game.Submission) (*game.Submission, error) {
	return s, nil
}
func (m *mockPRCreativeStore) ListByQuest(ctx context.Context, questID int64) ([]game.Submission, error) {
	return nil, nil
}
func (m *mockPRCreativeStore) ListByCrew(ctx context.Context, crewID string) ([]game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.subs, nil
}
func (m *mockPRCreativeStore) GetSubmission(ctx context.Context, submissionID int64) (*game.Submission, error) {
	return nil, game.ErrNotFound
}
func (m *mockPRCreativeStore) UpdateSubmission(ctx context.Context, submissionID int64, patch map[string]any) error {
	return nil
}

func TestProgressReader_CountCompletedQuests(t *testing.T) {
	qs := &mockPRQuestStore{
		quests: []game.Quest{
			{CrewID: "c1", Status: "DONE"},
			{CrewID: "c1", Status: "ACTIVE"},
			{CrewID: "c1", Status: "DONE"},
		},
	}
	r := NewProgressReader(qs, nil, nil, nil, nil, nil, nil)
	count, err := r.CountCompletedQuests(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 completed quests, got %d", count)
	}
}

func TestProgressReader_CountCompletedQuests_Error(t *testing.T) {
	qs := &mockPRQuestStore{err: errors.New("db error")}
	r := NewProgressReader(qs, nil, nil, nil, nil, nil, nil)
	_, err := r.CountCompletedQuests(context.Background(), "c1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProgressReader_CountCompletedRealms(t *testing.T) {
	rs := &mockPRRealmStore{
		progress: []game.RealmProgress{
			{CrewID: "c1", Realm: "r1", Status: "COMPLETE"},
			{CrewID: "c1", Realm: "r2", Status: "ACTIVE"},
			{CrewID: "c1", Realm: "r3", Status: "COMPLETE"},
		},
	}
	r := NewProgressReader(nil, rs, nil, nil, nil, nil, nil)
	count, err := r.CountCompletedRealms(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 completed realms, got %d", count)
	}
}

func TestProgressReader_CountCollectedRelics(t *testing.T) {
	rs := &mockPRRelicStore{count: 7}
	r := NewProgressReader(nil, nil, nil, rs, nil, nil, nil)
	count, err := r.CountCollectedRelics(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 7 {
		t.Errorf("expected 7 relics, got %d", count)
	}
}

func TestProgressReader_CountDailyStreak(t *testing.T) {
	ds := &mockPRDailyStore{
		turns: []game.DailyTurn{
			{Date: "2026-08-03", Completed: true},
			{Date: "2026-08-02", Completed: true},
			{Date: "2026-08-01", Completed: false},
		},
	}
	r := NewProgressReader(nil, nil, nil, nil, ds, nil, nil)
	count, err := r.CountDailyStreak(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected streak 2, got %d", count)
	}
}

func TestProgressReader_CountCreativeSubmissions(t *testing.T) {
	cs := &mockPRCreativeStore{
		subs: []game.Submission{
			{Status: "APPROVED"},
			{Status: "PENDING"},
			{Status: "APPROVED"},
		},
	}
	r := NewProgressReader(nil, nil, nil, nil, nil, cs, nil)
	count, err := r.CountCreativeSubmissions(context.Background(), "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 approved submissions, got %d", count)
	}
}

func TestProgressReader_GetPlayerLevel(t *testing.T) {
	us := &mockPRUserStore{
		player: &game.Player{UID: "u1", Level: 5},
	}
	r := NewProgressReader(nil, nil, us, nil, nil, nil, nil)
	level, err := r.GetPlayerLevel(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level != 5 {
		t.Errorf("expected level 5, got %d", level)
	}
}

func TestProgressReader_GetPlayerLevel_NilUser(t *testing.T) {
	us := &mockPRUserStore{player: nil}
	r := NewProgressReader(nil, nil, us, nil, nil, nil, nil)
	level, err := r.GetPlayerLevel(context.Background(), "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if level != 0 {
		t.Errorf("expected level 0 for nil user, got %d", level)
	}
}
