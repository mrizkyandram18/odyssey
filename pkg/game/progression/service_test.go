package progression

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/events"
)

type mockUserStore struct {
	player *game.Player
	err    error
}

func (m *mockUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.player == nil {
		return nil, game.ErrNotFound
	}
	return m.player, nil
}

func (m *mockUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	return m.err
}

func (m *mockUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
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

func (m *mockUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
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

func makePlayer(level int, xp int64) *game.Player {
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

func TestXPForLevel(t *testing.T) {
	cases := []struct {
		level int
		want  int64
	}{
		{1, 0},
		{2, 500},
		{3, 1000},
		{5, 2000},
	}
	for _, c := range cases {
		if got := XPForLevel(c.level); got != c.want {
			t.Errorf("XPForLevel(%d) = %d, want %d", c.level, got, c.want)
		}
	}
}

func TestXPForLevel_NegativeClampsToOne(t *testing.T) {
	if got := XPForLevel(-1); got != 0 {
		t.Errorf("XPForLevel(-1) = %d, want 0", got)
	}
}

func TestLevelFromXP(t *testing.T) {
	cases := []struct {
		xp   int64
		want int
	}{
		{0, 1},
		{50, 1},
		{99, 1},
		{100, 1},
		{499, 1},
		{500, 2},
		{1000, 3},
		{2400, 5},
	}
	for _, c := range cases {
		if got := LevelFromXP(c.xp); got != c.want {
			t.Errorf("LevelFromXP(%d) = %d, want %d", c.xp, got, c.want)
		}
	}
}

func TestLevelFromXP_Negative(t *testing.T) {
	if got := LevelFromXP(-10); got != 1 {
		t.Errorf("LevelFromXP(-10) = %d, want 1", got)
	}
}

func TestXPToNext(t *testing.T) {
	if got := XPToNext(0); got != 500 {
		t.Errorf("XPToNext(0) = %d, want 500", got)
	}
	if got := XPToNext(150); got != 350 {
		t.Errorf("XPToNext(150) = %d, want 350", got)
	}
}

func TestAwardXP_AddsXPNoLevelUp(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0)}
	svc := NewProgressionService(store, nil)
	player, levelUp, err := svc.AwardXP(context.Background(), "user-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.XP != 50 {
		t.Errorf("expected XP 50, got %d", player.XP)
	}
	if levelUp {
		t.Error("expected no level up at 50 XP")
	}
}

func TestAwardXP_TriggersLevelUp(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 480)}
	svc := NewProgressionService(store, nil)
	player, levelUp, err := svc.AwardXP(context.Background(), "user-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.XP != 530 {
		t.Errorf("expected XP 530, got %d", player.XP)
	}
	if player.Level != 2 {
		t.Errorf("expected level 2, got %d", player.Level)
	}
	if !levelUp {
		t.Error("expected level up")
	}
}

func TestAwardXP_LevelAlreadyHigh(t *testing.T) {
	store := &mockUserStore{player: makePlayer(3, 1250)}
	svc := NewProgressionService(store, nil)
	player, levelUp, err := svc.AwardXP(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.XP != 1260 {
		t.Errorf("expected XP 1260, got %d", player.XP)
	}
	if player.Level != 3 {
		t.Errorf("expected level 3, got %d", player.Level)
	}
	if levelUp {
		t.Error("expected no level up")
	}
}

func TestAwardXP_PersistsToStore(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 490)}
	svc := NewProgressionService(store, nil)
	if _, _, err := svc.AwardXP(context.Background(), "user-1", 30); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.player.XP != 520 {
		t.Errorf("expected persisted XP 520, got %d", store.player.XP)
	}
	if store.player.Level != 2 {
		t.Errorf("expected persisted level 2, got %d", store.player.Level)
	}
}

func TestAwardXP_UserNotFound(t *testing.T) {
	store := &mockUserStore{player: nil}
	svc := NewProgressionService(store, nil)
	_, _, err := svc.AwardXP(context.Background(), "user-1", 50)
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestAwardXP_StoreError(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0), err: errors.New("db error")}
	svc := NewProgressionService(store, nil)
	_, _, err := svc.AwardXP(context.Background(), "user-1", 50)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestAwardXP_IdempotentWithSameAmount(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 50)}
	svc := NewProgressionService(store, nil)

	p1, _, err := svc.AwardXP(context.Background(), "user-1", 30)
	if err != nil {
		t.Fatalf("first award: %v", err)
	}
	if p1.XP != 80 {
		t.Fatalf("expected XP 80 after first award, got %d", p1.XP)
	}

	p2, _, err := svc.AwardXP(context.Background(), "user-1", 30)
	if err != nil {
		t.Fatalf("second award: %v", err)
	}
	if p2.XP != 110 {
		t.Errorf("expected XP 110 after second award (cumulative), got %d", p2.XP)
	}
}

func TestAwardXP_RespectsCustomConfig(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 50)}
	cfg := &ProgressionConfig{
		XPPerLevel:        200,
		ChallengeXP:       10,
		CompletionBonusXP: 50,
	}
	svc := NewProgressionService(store, cfg)
	if svc.Config().XPPerLevel != 200 {
		t.Errorf("expected XPPerLevel 200, got %d", svc.Config().XPPerLevel)
	}
	if svc.Config().ChallengeXP != 10 {
		t.Errorf("expected ChallengeXP 10, got %d", svc.Config().ChallengeXP)
	}

	player, levelUp, err := svc.AwardXP(context.Background(), "user-1", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if player.XP != 60 {
		t.Errorf("expected XP 60, got %d", player.XP)
	}
	if levelUp {
		t.Error("expected no level up at 60 XP with XPPerLevel=200")
	}
}

func TestDefaultProgressionConfig(t *testing.T) {
	cfg := DefaultProgressionConfig()
	if cfg.XPPerLevel != 500 {
		t.Errorf("expected default XPPerLevel 500, got %d", cfg.XPPerLevel)
	}
	if cfg.ChallengeXP != 20 {
		t.Errorf("expected default ChallengeXP 20, got %d", cfg.ChallengeXP)
	}
	if cfg.CompletionBonusXP != 60 {
		t.Errorf("expected default CompletionBonusXP 60, got %d", cfg.CompletionBonusXP)
	}
}

func TestNewProgressionService_NilConfigUsesDefaults(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0)}
	svc := NewProgressionService(store, nil)
	if svc.cfg.XPPerLevel != 500 {
		t.Errorf("expected default XPPerLevel when nil config, got %d", svc.cfg.XPPerLevel)
	}
}

type capturePublisher struct {
	events []events.Event
}

func (p *capturePublisher) Publish(ctx context.Context, event events.Event) {
	p.events = append(p.events, event)
}

func TestAwardXP_PublishesLevelReachedEvent(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 480)}
	pub := &capturePublisher{}
	svc := NewProgressionServiceWithPublisher(store, nil, pub)

	_, levelUp, err := svc.AwardXP(context.Background(), "user-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !levelUp {
		t.Fatal("expected level up")
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.events))
	}
	ev, ok := pub.events[0].(events.LevelReachedEvent)
	if !ok {
		t.Fatalf("expected LevelReachedEvent, got %T", pub.events[0])
	}
	if ev.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", ev.UID)
	}
	if ev.OldLevel != 1 {
		t.Errorf("expected old level 1, got %d", ev.OldLevel)
	}
	if ev.NewLevel != 2 {
		t.Errorf("expected new level 2, got %d", ev.NewLevel)
	}
}

func TestAwardXP_NoLevelUp_NoEventPublished(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0)}
	pub := &capturePublisher{}
	svc := NewProgressionServiceWithPublisher(store, nil, pub)

	_, levelUp, err := svc.AwardXP(context.Background(), "user-1", 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if levelUp {
		t.Fatal("expected no level up")
	}
	if len(pub.events) != 0 {
		t.Errorf("expected 0 events published, got %d", len(pub.events))
	}
}

func TestAwardXP_RepeatedLevelUp_EventsForEachLevel(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0)}
	pub := &capturePublisher{}
	svc := NewProgressionServiceWithPublisher(store, nil, pub)

	_, _, _ = svc.AwardXP(context.Background(), "user-1", 600)
	_, _, _ = svc.AwardXP(context.Background(), "user-1", 600)

	if len(pub.events) != 2 {
		t.Errorf("expected 2 events published, got %d", len(pub.events))
	}
	ev1, ok1 := pub.events[0].(events.LevelReachedEvent)
	if !ok1 {
		t.Fatalf("expected LevelReachedEvent at index 0, got %T", pub.events[0])
	}
	if ev1.OldLevel != 1 || ev1.NewLevel != 2 {
		t.Errorf("expected level 1→2, got %d→%d", ev1.OldLevel, ev1.NewLevel)
	}
	ev2, ok2 := pub.events[1].(events.LevelReachedEvent)
	if !ok2 {
		t.Fatalf("expected LevelReachedEvent at index 1, got %T", pub.events[1])
	}
	if ev2.OldLevel != 2 || ev2.NewLevel != 3 {
		t.Errorf("expected level 2→3, got %d→%d", ev2.OldLevel, ev2.NewLevel)
	}
}

func TestProgressionService_Config_NoBalance_UsesDefaults(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0)}
	cfg := &ProgressionConfig{ChallengeXP: 15, CompletionBonusXP: 45}
	svc := NewProgressionServiceWithPublisher(store, cfg, nil)
	got := svc.Config()
	if got.ChallengeXP != 15 {
		t.Errorf("expected ChallengeXP 15, got %d", got.ChallengeXP)
	}
	if got.CompletionBonusXP != 45 {
		t.Errorf("expected CompletionBonusXP 45, got %d", got.CompletionBonusXP)
	}
}

func TestProgressionService_Config_WithBalance_OverridesValues(t *testing.T) {
	store := &mockUserStore{player: makePlayer(1, 0)}
	cfg := &ProgressionConfig{ChallengeXP: 15, CompletionBonusXP: 45}
	bal := balance.NewService(&mockBalanceStore{
		overrides: map[string]int64{
			string(balance.KeyChallengeXP):       30,
			string(balance.KeyCompletionBonusXP): 90,
			string(balance.KeyXPPerLevel):        200,
		},
	})
	if err := bal.Load(context.Background()); err != nil {
		t.Fatalf("balance load failed: %v", err)
	}
	svc := NewProgressionServiceWithPublisher(store, cfg, nil, bal)
	got := svc.Config()
	if got.ChallengeXP != 30 {
		t.Errorf("expected ChallengeXP 30 from balance, got %d", got.ChallengeXP)
	}
	if got.CompletionBonusXP != 90 {
		t.Errorf("expected CompletionBonusXP 90 from balance, got %d", got.CompletionBonusXP)
	}
	if got.XPPerLevel != 200 {
		t.Errorf("expected XPPerLevel 200 from balance, got %d", got.XPPerLevel)
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
func (m *mockUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}
