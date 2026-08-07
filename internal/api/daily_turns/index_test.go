package daily_turns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/game/dailyturn"
)

type mockDailyTurnHandler struct {
	turns []game.DailyTurn
	err   error
}

func (m *mockDailyTurnHandler) List(ctx context.Context, uid string) ([]game.DailyTurn, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.turns, nil
}

func (m *mockDailyTurnHandler) Consume(ctx context.Context, uid string, questSlug string) (*dailyturn.ConsumeDailyTurnResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &dailyturn.ConsumeDailyTurnResult{
		Turn: game.DailyTurn{
			ID:        1,
			UID:       uid,
			Date:      dailyturn.TodayDate(),
			QuestSlug: questSlug,
			Completed: true,
			CreatedAt: time.Now().UTC(),
		},
		XP:         dailyturn.DailyTurnXP,
		NewLevel:   2,
		LevelUp:    true,
		StreakDays: 3,
	}, nil
}

func makeUserToken(t *testing.T, issuer *auth.HMACSessionIssuer) string {
	t.Helper()
	token, _, err := issuer.IssueSession(auth.SessionKindUser, "user-1", &auth.SessionConfig{
		Role:   auth.RoleSeeker,
		CrewID: "crew-1",
	})
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func makeDailyTurn() game.DailyTurn {
	return game.DailyTurn{
		ID:        1,
		UID:       "user-1",
		Date:      "2026-08-03",
		QuestSlug: "morning-light",
		Completed: false,
		CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func TestDailyTurnsHandler_MethodNotAllowed(t *testing.T) {
	Setup(&mockDailyTurnHandler{})
	req := httptest.NewRequest(http.MethodPost, "/api/daily_turns", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestDailyTurnsHandler_Unauthorized(t *testing.T) {
	Setup(&mockDailyTurnHandler{turns: []game.DailyTurn{makeDailyTurn()}})
	req := httptest.NewRequest(http.MethodGet, "/api/daily_turns", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestDailyTurnsHandler_NotConfigured(t *testing.T) {
	Setup(nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/daily_turns", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestDailyTurnsHandler_List(t *testing.T) {
	Setup(&mockDailyTurnHandler{turns: []game.DailyTurn{makeDailyTurn()}})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/daily_turns", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var turns []game.DailyTurn
	if err := json.Unmarshal(w.Body.Bytes(), &turns); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(turns) != 1 {
		t.Fatalf("expected 1 turn, got %d", len(turns))
	}
	if turns[0].Date != "2026-08-03" {
		t.Errorf("expected date 2026-08-03, got %s", turns[0].Date)
	}
	if turns[0].QuestSlug != "morning-light" {
		t.Errorf("expected quest_slug morning-light, got %s", turns[0].QuestSlug)
	}
}

func TestDailyTurnsHandler_Consume_Success(t *testing.T) {
	Setup(&mockDailyTurnHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/daily_turns/consume", strings.NewReader(`{"quest_slug":"morning-light"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result dailyturn.ConsumeDailyTurnResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !result.Turn.Completed {
		t.Error("expected turn to be completed")
	}
	if result.XP != dailyturn.DailyTurnXP {
		t.Errorf("expected XP %d, got %d", dailyturn.DailyTurnXP, result.XP)
	}
	if result.StreakDays != 3 {
		t.Errorf("expected streak 3, got %d", result.StreakDays)
	}
}

func TestDailyTurnsHandler_Consume_NoTurnsRemaining(t *testing.T) {
	Setup(&mockDailyTurnHandler{err: dailyturn.ErrNoTurnsRemaining})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/daily_turns/consume", strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestDailyTurnsHandler_Consume_InvalidBody(t *testing.T) {
	Setup(&mockDailyTurnHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/daily_turns/consume", strings.NewReader("invalid"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDailyTurnsHandler_ListError(t *testing.T) {
	Setup(&mockDailyTurnHandler{err: ErrDailyTurnNotFound})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/daily_turns", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDailyTurnsHandler_ListError_Internal(t *testing.T) {
	Setup(&mockDailyTurnHandler{err: ErrInternal})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/daily_turns", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestDailyTurnsHandler_ImplementsInterface(t *testing.T) {
	var _ DailyTurnHandler = (*mockDailyTurnHandler)(nil)
}

var ErrDailyTurnNotFound = dailyturnNotFoundError("daily turn not found")
var ErrInternal = internalError("internal")

type dailyturnNotFoundError string

func (e dailyturnNotFoundError) Error() string { return string(e) }

type internalError string

func (e internalError) Error() string { return string(e) }
