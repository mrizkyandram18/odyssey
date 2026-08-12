package season

import (
	"context"
	"testing"
	"time"

	gamecontent "odyssey/pkg/game/content"
)

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
	for _, s := range m.seasons {
		if s.Slug == slug {
			return &s, nil
		}
	}
	return nil, nil
}

func makeSeason(slug, name string, start, end time.Time) gamecontent.SeasonDefinition {
	return gamecontent.SeasonDefinition{
		Slug:      slug,
		Name:      name,
		StartAt:   start,
		EndAt:     end,
		Journey:     "journey-1",
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestIsActive_EmptySlugAlwaysActive(t *testing.T) {
	svc := NewSeasonService(&mockSeasonGateway{}, nil)
	if !svc.IsActive(context.Background(), "") {
		t.Error("expected empty slug to be active")
	}
}

func TestIsActive_UnknownSeasonAlwaysActive(t *testing.T) {
	svc := NewSeasonService(&mockSeasonGateway{}, nil)
	if !svc.IsActive(context.Background(), "nonexistent") {
		t.Error("expected unknown season to be active")
	}
}

func TestIsActive_CurrentSeason(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("summer", "Summer",
				now.Add(-24*time.Hour), now.Add(24*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)
	if !svc.IsActive(context.Background(), "summer") {
		t.Error("expected summer season to be active")
	}
}

func TestIsActive_SeasonNotStarted(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("winter", "Winter",
				now.Add(24*time.Hour), now.Add(48*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)
	if svc.IsActive(context.Background(), "winter") {
		t.Error("expected winter season to be inactive (not started)")
	}
}

func TestIsActive_SeasonEnded(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("autumn", "Autumn",
				now.Add(-48*time.Hour), now.Add(-24*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)
	if svc.IsActive(context.Background(), "autumn") {
		t.Error("expected autumn season to be inactive (ended)")
	}
}

func TestGetState_Upcoming(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("future", "Future",
				now.Add(24*time.Hour), now.Add(48*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)

	state, err := svc.GetState(context.Background(), "future")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != SeasonStateUpcoming {
		t.Errorf("expected UPCOMING, got %s", state)
	}
}

func TestGetState_Active(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("summer", "Summer",
				now.Add(-24*time.Hour), now.Add(24*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)

	state, err := svc.GetState(context.Background(), "summer")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != SeasonStateActive {
		t.Errorf("expected ACTIVE, got %s", state)
	}
}

func TestGetState_Expired(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("autumn", "Autumn",
				now.Add(-48*time.Hour), now.Add(-24*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)

	state, err := svc.GetState(context.Background(), "autumn")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != SeasonStateExpired {
		t.Errorf("expected EXPIRED, got %s", state)
	}
}

func TestGetState_EmptySlug(t *testing.T) {
	svc := NewSeasonService(&mockSeasonGateway{}, nil)
	state, err := svc.GetState(context.Background(), "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if state != SeasonStateActive {
		t.Errorf("expected ACTIVE, got %s", state)
	}
}

func TestGetCurrentSeason_FindsActive(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("past", "Past",
				now.Add(-48*time.Hour), now.Add(-24*time.Hour)),
			makeSeason("summer", "Summer",
				now.Add(-12*time.Hour), now.Add(12*time.Hour)),
			makeSeason("future", "Future",
				now.Add(24*time.Hour), now.Add(48*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)

	result, err := svc.GetCurrentSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected active season")
	}
	if result.Definition.Slug != "summer" {
		t.Errorf("expected summer, got %s", result.Definition.Slug)
	}
}

func TestGetCurrentSeason_NoneActive(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("past", "Past",
				now.Add(-48*time.Hour), now.Add(-24*time.Hour)),
			makeSeason("future", "Future",
				now.Add(24*time.Hour), now.Add(48*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)

	result, err := svc.GetCurrentSeason(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Error("expected nil when no active season")
	}
}

func TestListAll_SortedByStartDate(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	gw := &mockSeasonGateway{
		seasons: []gamecontent.SeasonDefinition{
			makeSeason("future", "Future",
				now.Add(48*time.Hour), now.Add(72*time.Hour)),
			makeSeason("past", "Past",
				now.Add(-48*time.Hour), now.Add(-24*time.Hour)),
			makeSeason("summer", "Summer",
				now.Add(-12*time.Hour), now.Add(12*time.Hour)),
		},
	}
	cfg := &SeasonServiceConfig{Now: func() time.Time { return now }}
	svc := NewSeasonService(gw, cfg)

	result, err := svc.ListAll(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 seasons, got %d", len(result))
	}
	if result[0].Definition.Slug != "past" {
		t.Errorf("expected past first (sorted by start), got %s", result[0].Definition.Slug)
	}
	if result[1].State != SeasonStateActive {
		t.Errorf("expected summer active, got %s", result[1].State)
	}
	if result[2].Definition.Slug != "future" {
		t.Errorf("expected future last, got %s", result[2].Definition.Slug)
	}
}
