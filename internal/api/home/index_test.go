package home

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	gamehome "odyssey/pkg/game/home"
	"odyssey/pkg/game/mission"
)

type mockHomeService struct {
	resp *gamehome.HomeResponse
	err  error
}

func (m *mockHomeService) GetHome(ctx context.Context, uid string, crewID string) (*gamehome.HomeResponse, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.resp, nil
}

func makeUserToken(t *testing.T, issuer *auth.HMACSessionIssuer) string {
	t.Helper()
	token, _, err := issuer.IssueSession(auth.SessionKindUser, "user-1", &auth.SessionConfig{
		Role:   auth.RoleSeeker,
		FamilyID: "crew-1",
	})
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func makeHomeResponse() *gamehome.HomeResponse {
	return &gamehome.HomeResponse{
		Player: game.Player{
			UID:          "user-1",
			FamilyID:       "crew-1",
			ExplorerName: "Alice",
			Role:         "SEEKER",
			Level:        3,
			XP:           500,
			CreatedAt:    time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
			UpdatedAt:    time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		},
		Missions: []quest.QuestView{},
		DailyMission: gamehome.DailyTurnView{
			Today:          "2026-08-03",
			Completed:      false,
			Available:      true,
			StreakDays:     2,
			RemainingTurns: 1,
		},
		JourneyProgress:        []game.JourneyProgress{},
		RelicCount:           5,
		ActiveQuests:         []quest.QuestView{},
		CompletedQuestsToday: []quest.QuestView{},
	}
}

func TestHomeHandler_MethodNotAllowed(t *testing.T) {
	Setup(&mockHomeService{resp: makeHomeResponse()})
	req := httptest.NewRequest(http.MethodPost, "/api/home", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestHomeHandler_Unauthorized(t *testing.T) {
	Setup(&mockHomeService{resp: makeHomeResponse()})
	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestHomeHandler_NotConfigured(t *testing.T) {
	Setup(nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestHomeHandler_Success(t *testing.T) {
	Setup(&mockHomeService{resp: makeHomeResponse()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp gamehome.HomeResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Player.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", resp.Player.UID)
	}
	if resp.DailyMission.StreakDays != 2 {
		t.Errorf("expected streak 2, got %d", resp.DailyMission.StreakDays)
	}
	if resp.RelicCount != 5 {
		t.Errorf("expected relic count 5, got %d", resp.RelicCount)
	}
}

func TestHomeHandler_ServiceError(t *testing.T) {
	Setup(&mockHomeService{err: ErrHomeError})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/home", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

var ErrHomeError = homeError("home error")

type homeError string

func (e homeError) Error() string { return string(e) }
