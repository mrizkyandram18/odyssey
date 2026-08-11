package seasons

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"odyssey/pkg/auth"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/season"
)

type mockSeasonHandler struct {
	current *season.SeasonSummary
	all     []season.SeasonSummary
	err     error
}

func (m *mockSeasonHandler) GetCurrentSeason(ctx context.Context) (*season.SeasonSummary, error) {
	return m.current, m.err
}

func (m *mockSeasonHandler) ListAll(ctx context.Context) ([]season.SeasonSummary, error) {
	return m.all, m.err
}

func makeSeasonDef(slug, name, realm string, start, end time.Time) gamecontent.SeasonDefinition {
	return gamecontent.SeasonDefinition{
		Slug:      slug,
		Name:      name,
		Realm:     realm,
		StartAt:   start,
		EndAt:     end,
		Published: true,
	}
}

func TestSeasons_ListAll(t *testing.T) {
	now := time.Date(2026, 8, 4, 12, 0, 0, 0, time.UTC)
	Setup(&mockSeasonHandler{
		all: []season.SeasonSummary{
			{Definition: makeSeasonDef("s1", "S1", "r1", now.Add(-24*time.Hour), now.Add(24*time.Hour)), State: season.SeasonStateActive},
			{Definition: makeSeasonDef("s2", "S2", "r2", now.Add(24*time.Hour), now.Add(48*time.Hour)), State: season.SeasonStateUpcoming},
		},
	})
	defer func() { handler = nil }()

	claims := &auth.SessionClaims{UID: "u1", CrewID: "c1", Kind: "user"}
	req := httptest.NewRequest(http.MethodGet, "/api/seasons", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	if !contains(body, "S1") || !contains(body, "S2") {
		t.Fatalf("expected both seasons in response, got: %s", body)
	}
}

func TestSeasons_MethodNotAllowed(t *testing.T) {
	Setup(&mockSeasonHandler{})
	defer func() { handler = nil }()

	req := httptest.NewRequest(http.MethodPost, "/api/seasons", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
