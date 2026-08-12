package familystreak

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockUserStore struct {
	game.UserStore
	members []game.Player
	err     error
}

func (m *mockUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.members, nil
}

type mockActivityStore struct {
	acts []game.DailyActivity
	err  error
	got  []string
}

func (m *mockActivityStore) RecordActivity(ctx context.Context, act *game.DailyActivity) (*game.DailyActivity, error) {
	return act, nil
}

func (m *mockActivityStore) GetStreak(ctx context.Context, uid string) (int, error) {
	return 0, nil
}

func (m *mockActivityStore) ListActivityDatesByUsers(ctx context.Context, uids []string) ([]game.DailyActivity, error) {
	m.got = append([]string{}, uids...)
	if m.err != nil {
		return nil, m.err
	}
	return m.acts, nil
}

func loc(tz string) *time.Location {
	l, err := time.LoadLocation(tz)
	if err != nil {
		return time.UTC
	}
	return l
}

func fixed(tz string, y int, mo time.Month, d, h, mi int) time.Time {
	return time.Date(y, mo, d, h, mi, 0, 0, loc(tz))
}

func dates(t *testing.T, svc *Service, crewID string) (int, error) {
	return svc.ComputeCrewStreak(context.Background(), crewID)
}

func TestComputeCrewStreak_ConsecutiveDays(t *testing.T) {
	acts := []game.DailyActivity{{UserID: "u1", ActivityDate: "2026-08-09"}, {UserID: "u1", ActivityDate: "2026-08-10"}, {UserID: "u1", ActivityDate: "2026-08-11"}}
	svc := NewService(&mockUserStore{members: []game.Player{{UID: "u1"}}}, &mockActivityStore{acts: acts}, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 12, 0) }

	got, err := dates(t, svc, "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 3 {
		t.Errorf("expected crew streak 3, got %d", got)
	}
}

func TestComputeCrewStreak_UnionOfMembers(t *testing.T) {
	// u1 active today, u2 active yesterday -> union gives 2 consecutive days.
	acts := []game.DailyActivity{
		{UserID: "u1", ActivityDate: "2026-08-11"},
		{UserID: "u2", ActivityDate: "2026-08-10"},
	}
	svc := NewService(&mockUserStore{members: []game.Player{{UID: "u1"}, {UID: "u2"}}}, &mockActivityStore{acts: acts}, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 18, 0) }

	got, err := dates(t, svc, "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("expected crew streak 2 from member union, got %d", got)
	}
}

func TestComputeCrewStreak_BrokenStreak(t *testing.T) {
	// Activity on 08-09 but not 08-10 -> stale run does not count today.
	acts := []game.DailyActivity{{UserID: "u1", ActivityDate: "2026-08-09"}}
	svc := NewService(&mockUserStore{members: []game.Player{{UID: "u1"}}}, &mockActivityStore{acts: acts}, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 12, 0) }

	got, err := dates(t, svc, "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("expected crew streak 0 on broken streak, got %d", got)
	}
}

func TestComputeCrewStreak_LastActivityYesterday(t *testing.T) {
	// Today missing but yesterday active -> streak counts from yesterday.
	acts := []game.DailyActivity{{UserID: "u1", ActivityDate: "2026-08-10"}, {UserID: "u1", ActivityDate: "2026-08-09"}}
	svc := NewService(&mockUserStore{members: []game.Player{{UID: "u1"}}}, &mockActivityStore{acts: acts}, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 8, 0) }

	got, err := dates(t, svc, "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("expected crew streak 2 (from yesterday), got %d", got)
	}
}

func TestComputeCrewStreak_NoActivity(t *testing.T) {
	svc := NewService(&mockUserStore{members: []game.Player{{UID: "u1"}}}, &mockActivityStore{}, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 12, 0) }

	got, err := dates(t, svc, "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("expected crew streak 0, got %d", got)
	}
}

func TestComputeCrewStreak_EmptyCrew(t *testing.T) {
	// Empty crew must never query activity data.
	activity := &mockActivityStore{}
	svc := NewService(&mockUserStore{members: nil}, activity, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 12, 0) }

	got, err := dates(t, svc, "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 0 {
		t.Errorf("expected crew streak 0 on empty crew, got %d", got)
	}
	if activity.got != nil {
		t.Errorf("activity store must not be queried for an empty crew, got %v", activity.got)
	}
}

func TestComputeCrewStreak_TimeZoneBoundary(t *testing.T) {
	// 2026-08-11 02:00 WIB == 2026-08-10 19:00 UTC. Under UTC the newest date
	// would be 08-11 which is not today/yesterday (UTC), so personal-UTC
	// semantics would miscount. Family streak must follow the write timezone.
	acts := []game.DailyActivity{{UserID: "u1", ActivityDate: "2026-08-10"}, {UserID: "u1", ActivityDate: "2026-08-11"}}
	svc := NewService(&mockUserStore{members: []game.Player{{UID: "u1"}}}, &mockActivityStore{acts: acts}, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 2, 0) }

	got, err := dates(t, svc, "c1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2 {
		t.Errorf("expected crew streak 2 across WIB day boundary, got %d", got)
	}
}

func TestComputeCrewStreak_CrewIsolation(t *testing.T) {
	// Queries must only ever ask for this crew's member uids.
	members := []game.Player{{UID: "u1"}, {UID: "u2"}}
	activity := &mockActivityStore{acts: []game.DailyActivity{{UserID: "u1", ActivityDate: "2026-08-11"}}}
	svc := NewService(&mockUserStore{members: members}, activity, "Asia/Jakarta")
	svc.now = func() time.Time { return fixed("Asia/Jakarta", 2026, 8, 11, 12, 0) }

	if _, err := dates(t, svc, "c1"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(activity.got) != 2 || activity.got[0] != "u1" || activity.got[1] != "u2" {
		t.Errorf("expected activity query restricted to crew uids [u1 u2], got %v", activity.got)
	}
}

func TestComputeCrewStreak_MembershipError(t *testing.T) {
	svc := NewService(&mockUserStore{err: errors.New("db down")}, &mockActivityStore{}, "Asia/Jakarta")
	_, err := dates(t, svc, "c1")
	if err == nil {
		t.Fatal("expected error from membership lookup")
	}
}

func TestComputeCrewStreak_ActivityError(t *testing.T) {
	svc := NewService(&mockUserStore{members: []game.Player{{UID: "u1"}}}, &mockActivityStore{err: errors.New("db down")}, "Asia/Jakarta")
	_, err := dates(t, svc, "c1")
	if err == nil {
		t.Fatal("expected error from activity lookup")
	}
}

func TestCountConsecutiveDays_Empty(t *testing.T) {
	if got := CountConsecutiveDays(nil, time.Now()); got != 0 {
		t.Errorf("expected 0 for empty dates, got %d", got)
	}
}

func TestCountConsecutiveDays_YesterdayOnly(t *testing.T) {
	dates := map[string]struct{}{"2026-08-10": {}}
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	if got := CountConsecutiveDays(dates, now); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}

func TestCountConsecutiveDays_TodayOnly(t *testing.T) {
	dates := map[string]struct{}{"2026-08-11": {}}
	now := time.Date(2026, 8, 11, 12, 0, 0, 0, time.UTC)
	if got := CountConsecutiveDays(dates, now); got != 1 {
		t.Errorf("expected 1, got %d", got)
	}
}
