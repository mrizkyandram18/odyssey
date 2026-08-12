package dailymission

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/observability"
)

func newMetricService(turns []game.DailyMission, now time.Time) *DailyTurnService {
	cfg := &DailyTurnConfig{
		XP:             10,
		MaxTurnsPerDay: 1,
		Timezone:       "UTC",
		Now:            func() time.Time { return now },
	}
	return NewDailyTurnService(&mockDailyTurnStore{turns: turns}, cfg)
}

func TestDailyTurn_ReplayIgnoredMetric(t *testing.T) {
	now := time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC)
	date := "2026-08-05"
	svc := newMetricService([]game.DailyMission{
		makeTurn("2026-08-05", true),
	}, now)
	m := observability.NewMetrics()
	svc.SetMetrics(m)

	// Same quest slug already completed today -> idempotency guard fires.
	_, err := svc.ConsumeDailyTurn(context.Background(), "user-1", date, "morning-light")
	if err != ErrNoTurnsRemaining {
		t.Fatalf("expected ErrNoTurnsRemaining, got %v", err)
	}
	if m.Snapshot().ReplayIgnored != 1 {
		t.Errorf("expected 1 replay ignored, got %d", m.Snapshot().ReplayIgnored)
	}
}

func TestDailyTurn_SetMetricsNilSafe(t *testing.T) {
	svc := newMetricService(nil, time.Date(2026, 8, 5, 12, 0, 0, 0, time.UTC))
	svc.SetMetrics(nil)
	_, err := svc.ConsumeDailyTurn(context.Background(), "user-1", "2026-08-05", "morning-light")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
