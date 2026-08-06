package dailyturn

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/progression"
)

type mockDailyTurnStore struct {
	turns []game.DailyTurn
	err   error
}

func (m *mockDailyTurnStore) CreateDailyTurn(ctx context.Context, dt *game.DailyTurn) (*game.DailyTurn, error) {
	if m.err != nil {
		return nil, m.err
	}
	dt.ID = int64(len(m.turns) + 1)
	m.turns = append(m.turns, *dt)
	return dt, nil
}

func (m *mockDailyTurnStore) UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *mockDailyTurnStore) ListDailyTurns(ctx context.Context, uid string) ([]game.DailyTurn, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.turns, nil
}

func makeTurn(date string, completed bool) game.DailyTurn {
	return game.DailyTurn{
		ID:        1,
		UID:       "user-1",
		Date:      date,
		QuestSlug: "morning-light",
		Completed: completed,
		CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func newService(turns []game.DailyTurn, cfg *DailyTurnConfig) *DailyTurnService {
	if cfg == nil {
		cfg = &DailyTurnConfig{}
	}
	cfg = &DailyTurnConfig{
		XP:             cfg.XP,
		MaxTurnsPerDay: cfg.MaxTurnsPerDay,
		Timezone:       cfg.Timezone,
		Now:            cfg.Now,
	}
	if cfg.Timezone == "" {
		cfg.Timezone = "UTC"
	}
	if cfg.Now == nil {
		cfg.Now = time.Now
	}
	return NewDailyTurnService(&mockDailyTurnStore{turns: turns}, &DailyTurnConfig{})
}

func TestList_ReturnsTurnsNewestFirst(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn("2026-08-01", true),
			makeTurn("2026-08-03", false),
			makeTurn("2026-08-02", true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	turns, err := svc.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(turns) != 3 {
		t.Fatalf("expected 3 turns, got %d", len(turns))
	}
	if turns[0].Date != "2026-08-03" {
		t.Errorf("expected newest first (2026-08-03), got %s", turns[0].Date)
	}
}

func TestList_Empty(t *testing.T) {
	store := &mockDailyTurnStore{}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	turns, err := svc.List(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(turns) != 0 {
		t.Fatalf("expected 0 turns, got %d", len(turns))
	}
}

func TestHasCompletedToday_True(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(TodayDate(), true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	done, err := svc.HasCompletedToday(context.Background(), "user-1", TodayDate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done {
		t.Error("expected completed today = true")
	}
}

func TestHasCompletedToday_False(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(TodayDate(), false),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	done, err := svc.HasCompletedToday(context.Background(), "user-1", TodayDate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if done {
		t.Error("expected completed today = false")
	}
}

func TestHasCompletedToday_NoTurnToday(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn("2000-01-01", true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	done, err := svc.HasCompletedToday(context.Background(), "user-1", TodayDate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if done {
		t.Error("expected completed today = false for no turn")
	}
}

func TestIsAvailableToday_NoRecord(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn("2000-01-01", true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	available, err := svc.IsAvailableToday(context.Background(), "user-1", TodayDate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected available = true when no turn record exists")
	}
}

func TestIsAvailableToday_Completed(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(TodayDate(), true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	available, err := svc.IsAvailableToday(context.Background(), "user-1", TodayDate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected available = false when turn is completed")
	}
}

func TestIsAvailableToday_NotCompleted(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(TodayDate(), false),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	available, err := svc.IsAvailableToday(context.Background(), "user-1", TodayDate())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !available {
		t.Error("expected available = true when turn exists but not completed")
	}
}

func TestIsAvailableToday_RespectsMaxTurnsPerDay(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			{ID: 1, UID: "user-1", Date: today, QuestSlug: "morning-light", Completed: true, CreatedAt: time.Now().UTC()},
			{ID: 2, UID: "user-1", Date: today, QuestSlug: "riddle-of-the-stones", Completed: true, CreatedAt: time.Now().UTC()},
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{MaxTurnsPerDay: 1})
	available, err := svc.IsAvailableToday(context.Background(), "user-1", today)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if available {
		t.Error("expected available = false when max daily turns exceeded")
	}
}

func TestComputeStreak_ThreeConsecutiveDays(t *testing.T) {
	today := TodayDate()
	yesterday := offsetDate(today, -1)
	dayBefore := offsetDate(today, -2)

	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(today, true),
			makeTurn(yesterday, true),
			makeTurn(dayBefore, true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	streak, err := svc.ComputeStreak(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if streak != 3 {
		t.Errorf("expected streak 3, got %d", streak)
	}
}

func TestComputeStreak_BreaksAtUncompleted(t *testing.T) {
	today := TodayDate()
	yesterday := offsetDate(today, -1)
	dayBefore := offsetDate(today, -2)

	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(today, true),
			makeTurn(yesterday, false),
			makeTurn(dayBefore, true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	streak, err := svc.ComputeStreak(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if streak != 1 {
		t.Errorf("expected streak 1, got %d", streak)
	}
}

func TestComputeStreak_TodayNotCompleted(t *testing.T) {
	yesterday := offsetDate(TodayDate(), -1)
	dayBefore := offsetDate(TodayDate(), -2)

	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(yesterday, true),
			makeTurn(dayBefore, true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	streak, err := svc.ComputeStreak(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if streak != 2 {
		t.Errorf("expected streak 2, got %d", streak)
	}
}

func TestComputeStreak_NoCompletedTurns(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	streak, err := svc.ComputeStreak(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if streak != 0 {
		t.Errorf("expected streak 0, got %d", streak)
	}
}

func TestComputeStreak_MissingDayBreaksStreak(t *testing.T) {
	today := TodayDate()
	dayBefore := offsetDate(today, -2)

	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			makeTurn(today, true),
			makeTurn(dayBefore, true),
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	streak, err := svc.ComputeStreak(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if streak != 1 {
		t.Errorf("expected streak 1, got %d", streak)
	}
}

func TestTodayDate_Format(t *testing.T) {
	d := TodayDate()
	if len(d) != 10 {
		t.Errorf("expected YYYY-MM-DD format, got %s (len %d)", d, len(d))
	}
	_, err := time.Parse("2006-01-02", d)
	if err != nil {
		t.Errorf("expected valid date format, got error: %v", err)
	}
}

func TestParseDate_Valid(t *testing.T) {
	t1, err := ParseDate("2026-08-03")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if t1.Year() != 2026 || t1.Month() != 8 || t1.Day() != 3 {
		t.Errorf("expected 2026-08-03, got %v", t1)
	}
}

func TestParseDate_WithWhitespace(t *testing.T) {
	_, err := ParseDate("  2026-08-03  ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestParseDate_Invalid(t *testing.T) {
	_, err := ParseDate("not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid date")
	}
}

func TestDailyTurnService_ImplementsStore(t *testing.T) {
	var _ game.DailyTurnStore = (*mockDailyTurnStore)(nil)
}

func offsetDate(base string, days int) string {
	t, err := time.Parse("2006-01-02", base)
	if err != nil {
		panic(err)
	}
	return t.AddDate(0, 0, days).Format("2006-01-02")
}

func TestList_Error(t *testing.T) {
	store := &mockDailyTurnStore{err: errors.New("db error")}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	_, err := svc.List(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConsumeDailyTurn_Success(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			{ID: 1, UID: "user-1", Date: today, QuestSlug: "morning-light", Completed: false, CreatedAt: time.Now().UTC()},
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	turn, err := svc.ConsumeDailyTurn(context.Background(), "user-1", today, "morning-light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !turn.Completed {
		t.Error("expected turn to be completed")
	}
	if turn.QuestSlug != "morning-light" {
		t.Errorf("expected quest_slug morning-light, got %s", turn.QuestSlug)
	}
}

func TestConsumeDailyTurn_NoRecord(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{turns: []game.DailyTurn{}}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	turn, err := svc.ConsumeDailyTurn(context.Background(), "user-1", today, "morning-light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !turn.Completed {
		t.Error("expected turn to be completed")
	}
}

func TestConsumeDailyTurn_AlreadyCompleted(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			{ID: 1, UID: "user-1", Date: today, QuestSlug: "morning-light", Completed: true, CreatedAt: time.Now().UTC()},
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	_, err := svc.ConsumeDailyTurn(context.Background(), "user-1", today, "morning-light")
	if err != ErrNoTurnsRemaining {
		t.Errorf("expected ErrNoTurnsRemaining, got %v", err)
	}
}

func TestConsumeDailyTurn_StoreError(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{err: errors.New("db error")}
	svc := NewDailyTurnService(store, &DailyTurnConfig{})
	_, err := svc.ConsumeDailyTurn(context.Background(), "user-1", today, "morning-light")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConsumeDailyTurn_ExceedsMaxTurns(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			{ID: 1, UID: "user-1", Date: today, QuestSlug: "morning-light", Completed: true, CreatedAt: time.Now().UTC()},
			{ID: 2, UID: "user-1", Date: today, QuestSlug: "riddle-of-the-stones", Completed: true, CreatedAt: time.Now().UTC()},
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{MaxTurnsPerDay: 1})
	_, err := svc.ConsumeDailyTurn(context.Background(), "user-1", today, "gather-herbs")
	if err != ErrNoTurnsRemaining {
		t.Errorf("expected ErrNoTurnsRemaining, got %v", err)
	}
}

func TestConsumeDailyTurn_WithinMaxTurns(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			{ID: 1, UID: "user-1", Date: today, QuestSlug: "morning-light", Completed: true, CreatedAt: time.Now().UTC()},
		},
	}
	svc := NewDailyTurnService(store, &DailyTurnConfig{MaxTurnsPerDay: 2})
	turn, err := svc.ConsumeDailyTurn(context.Background(), "user-1", today, "gather-herbs")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !turn.Completed {
		t.Error("expected turn to be completed")
	}
}

func TestTodayDate_RespectsTimezone(t *testing.T) {
	fixedTime := time.Date(2026, 8, 4, 1, 0, 0, 0, time.UTC)
	svc := NewDailyTurnService(&mockDailyTurnStore{}, &DailyTurnConfig{
		Timezone: "Asia/Jakarta",
		Now:      func() time.Time { return fixedTime },
	})
	// Jakarta is UTC+7, so 01:00 UTC = 08:00 Jakarta, date is still 2026-08-04
	if svc.TodayDate() != "2026-08-04" {
		t.Errorf("expected 2026-08-04 for Asia/Jakarta, got %s", svc.TodayDate())
	}
}

type capturePublisher struct {
	events []events.Event
}

func (p *capturePublisher) Publish(ctx context.Context, event events.Event) {
	p.events = append(p.events, event)
}

type mockUserStoreForDailyTurn struct {
	player *game.Player
}

func (m *mockUserStoreForDailyTurn) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	return m.player, nil
}

func (m *mockUserStoreForDailyTurn) CreateUser(ctx context.Context, p *game.Player) error {
	return nil
}

func (m *mockUserStoreForDailyTurn) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	if v, ok := patch["xp"].(int64); ok {
		m.player.XP = v
	}
	if v, ok := patch["level"].(int); ok {
		m.player.Level = v
	}
	if v, ok := patch["version"].(int); ok {
		m.player.Version = v
	}
	return nil
}
func (m *mockUserStoreForDailyTurn) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
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

func TestDailyTurnAPIHandler_PublishesDailyTurnCompletedEvent(t *testing.T) {
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			{ID: 1, UID: "user-1", Date: "2026-08-03", QuestSlug: "morning-light", Completed: false, CreatedAt: time.Now().UTC()},
		},
	}
	dts := NewDailyTurnService(store, &DailyTurnConfig{XP: 10, MaxTurnsPerDay: 3})

	userStore := &mockUserStoreForDailyTurn{player: &game.Player{UID: "user-1", CrewID: "crew-1", Level: 1, XP: 0}}
	prog := progression.NewProgressionService(userStore, nil)

	pub := &capturePublisher{}
	handler := NewDailyTurnAPIHandlerWithPublisher(dts, prog, pub)

	result, err := handler.Consume(context.Background(), "user-1", "morning-light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StreakDays != 1 {
		t.Errorf("expected streak 1, got %d", result.StreakDays)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.events))
	}
	ev, ok := pub.events[0].(events.DailyTurnCompletedEvent)
	if !ok {
		t.Fatalf("expected DailyTurnCompletedEvent, got %T", pub.events[0])
	}
	if ev.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", ev.UID)
	}
	if ev.StreakDays != 1 {
		t.Errorf("expected streak 1, got %d", ev.StreakDays)
	}
}

func TestDailyTurnAPIHandler_UsesBalanceOverrideForXP(t *testing.T) {
	today := TodayDate()
	store := &mockDailyTurnStore{
		turns: []game.DailyTurn{
			{ID: 1, UID: "user-1", Date: today, QuestSlug: "morning-light", Completed: false, CreatedAt: time.Now().UTC()},
		},
	}
	dts := NewDailyTurnService(store, &DailyTurnConfig{XP: 10, MaxTurnsPerDay: 3})

	userStore := &mockUserStoreForDailyTurn{player: &game.Player{UID: "user-1", CrewID: "crew-1", Level: 1, XP: 0}}
	prog := progression.NewProgressionService(userStore, nil)

	bal := balance.NewService(&mockBalanceStoreForDailyTurn{xpOverride: 50})
	if err := bal.Load(context.Background()); err != nil {
		t.Fatalf("balance load failed: %v", err)
	}
	handler := NewDailyTurnAPIHandlerWithPublisher(dts, prog, nil, bal)

	result, err := handler.Consume(context.Background(), "user-1", "morning-light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.XP != 50 {
		t.Errorf("expected XP 50 from balance override, got %d", result.XP)
	}
	if userStore.player.XP != 50 {
		t.Errorf("expected persisted XP 50, got %d", userStore.player.XP)
	}
}

type mockBalanceStoreForDailyTurn struct {
	xpOverride int64
}

func (m *mockBalanceStoreForDailyTurn) GetOverride(ctx context.Context, key string) (*balance.Override, error) {
	if key == "daily_turn_xp" {
		return &balance.Override{Key: key, Value: m.xpOverride}, nil
	}
	return nil, balance.ErrConfigNotFound
}

func (m *mockBalanceStoreForDailyTurn) ListOverrides(ctx context.Context) ([]balance.Override, error) {
	return []balance.Override{
		{Key: "daily_turn_xp", Value: m.xpOverride},
	}, nil
}
func (m *mockUserStoreForDailyTurn) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) { return nil, nil }
