package crews

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

type mockCrewStore struct {
	crew *game.Crew
	err  error
}

func (m *mockCrewStore) GetCrew(ctx context.Context, crewID string) (*game.Crew, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.crew == nil {
		return nil, game.ErrNotFound
	}
	return m.crew, nil
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

func makeCrew() *game.Crew {
	return &game.Crew{
		ID:        "crew-1",
		Name:      "The Explorer Family",
		CreatedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func TestCrews_Handler_MethodNotAllowed(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/crews", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestCrews_Handler_Unauthorized(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()})
	req := httptest.NewRequest(http.MethodGet, "/api/crews", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCrews_Handler_NotConfigured(t *testing.T) {
	Setup(nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/crews", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestCrews_Handler_Success(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/crews", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp crewResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.ID != "crew-1" {
		t.Errorf("expected ID crew-1, got %s", resp.ID)
	}
	if resp.Name != "The Explorer Family" {
		t.Errorf("expected name 'The Explorer Family', got %s", resp.Name)
	}
}

func TestCrews_Handler_NotFound(t *testing.T) {
	Setup(&mockCrewStore{err: game.ErrNotFound})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/crews", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestCrews_Handler_StoreError(t *testing.T) {
	Setup(&mockCrewStore{err: errors.New("db failure")})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/crews", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
