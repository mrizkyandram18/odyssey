package realm_progress

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
)

type mockRealmProgressStore struct {
	progress []game.RealmProgress
	err      error
}

func (m *mockRealmProgressStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.RealmProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.progress, nil
}

func makeToken(t *testing.T, issuer *auth.HMACSessionIssuer) string {
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

func makeProgressList() []game.RealmProgress {
	return []game.RealmProgress{
		{
			CrewID:         "crew-1",
			Realm:          "whispering-woods",
			Status:         "ACTIVE",
			StoryBranch:    "main-path",
			Progress:       45,
			LastUnlockedAt: time.Date(2026, 8, 1, 10, 0, 0, 0, time.UTC),
			UpdatedAt:      time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		},
		{
			CrewID:         "crew-1",
			Realm:          "clockwork-city",
			Status:         "LOCKED",
			StoryBranch:    "",
			Progress:       0,
			LastUnlockedAt: time.Time{},
			UpdatedAt:      time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
		},
	}
}

func TestRealmProgress_Handler_MethodNotAllowed(t *testing.T) {
	Setup(&mockRealmProgressStore{progress: makeProgressList()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/realm_progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestRealmProgress_Handler_Unauthorized(t *testing.T) {
	Setup(&mockRealmProgressStore{progress: makeProgressList()})
	req := httptest.NewRequest(http.MethodGet, "/api/realm_progress", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRealmProgress_Handler_NotConfigured(t *testing.T) {
	Setup(nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/realm_progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestRealmProgress_Handler_Success(t *testing.T) {
	Setup(&mockRealmProgressStore{progress: makeProgressList()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/realm_progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []game.RealmProgress
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(resp))
	}
	if resp[0].Realm != "whispering-woods" {
		t.Errorf("expected first realm whispering-woods, got %s", resp[0].Realm)
	}
	if resp[0].Progress != 45 {
		t.Errorf("expected progress 45, got %d", resp[0].Progress)
	}
}

func TestRealmProgress_Handler_EmptyResult(t *testing.T) {
	Setup(&mockRealmProgressStore{progress: nil})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/realm_progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp []game.RealmProgress
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(resp) != 0 {
		t.Errorf("expected 0 entries, got %d", len(resp))
	}
}

func TestRealmProgress_Handler_StoreError(t *testing.T) {
	Setup(&mockRealmProgressStore{err: errors.New("db failure")})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/realm_progress", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
