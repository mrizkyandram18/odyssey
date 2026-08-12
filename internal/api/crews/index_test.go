package crews

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (m *mockCrewStore) UpdateCrew(ctx context.Context, crewID string, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	if m.crew != nil {
		if v, ok := patch["banner_url"].(string); ok {
			m.crew.BannerURL = v
		}
		if v, ok := patch["theme"].(string); ok {
			m.crew.Theme = v
		}
	}
	return nil
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
		BannerURL: "",
		Theme:     "default",
		CreatedAt: time.Date(2026, 8, 1, 12, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func TestCrews_Handler_MethodNotAllowed(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()}, nil)
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
	Setup(&mockCrewStore{crew: makeCrew()}, nil)
	req := httptest.NewRequest(http.MethodGet, "/api/crews", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCrews_Handler_NotConfigured(t *testing.T) {
	Setup(nil, nil)
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
	Setup(&mockCrewStore{crew: makeCrew()}, nil)
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
	Setup(&mockCrewStore{err: game.ErrNotFound}, nil)
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
	Setup(&mockCrewStore{err: errors.New("db failure")}, nil)
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

func TestCrews_Handler_PatchSuccess(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()}, nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	body := map[string]string{"banner_url": "https://example.com/banner.png", "theme": "forest"}
	req := httptest.NewRequest(http.MethodPatch, "/api/crews", marshalJSON(t, body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp crewResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.BannerURL != "https://example.com/banner.png" {
		t.Errorf("expected banner_url 'https://example.com/banner.png', got '%s'", resp.BannerURL)
	}
	if resp.Theme != "forest" {
		t.Errorf("expected theme 'forest', got '%s'", resp.Theme)
	}
}

func TestCrews_Handler_PatchUnauthorized(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()}, nil)
	req := httptest.NewRequest(http.MethodPatch, "/api/crews", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCrews_Handler_PatchBadRequest(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()}, nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	req := httptest.NewRequest(http.MethodPatch, "/api/crews", strings.NewReader("not-json"))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCrews_Handler_PatchNoFields(t *testing.T) {
	Setup(&mockCrewStore{crew: makeCrew()}, nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeToken(t, issuer)

	body := map[string]string{}
	req := httptest.NewRequest(http.MethodPatch, "/api/crews", marshalJSON(t, body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func marshalJSON(t *testing.T, v any) *strings.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return strings.NewReader(string(b))
}
