package story_fragments

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/game/fragment"
)

type mockFragmentHandler struct {
	fragments []fragment.StoryFragmentView
	discRes   *fragment.DiscoverResult
	repRes    *fragment.ReplayResult
	err       error
	notFound  bool
}

func (m *mockFragmentHandler) ListPlayerFragments(ctx context.Context, uid string) ([]fragment.StoryFragmentView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.fragments, nil
}

func (m *mockFragmentHandler) DiscoverFragment(ctx context.Context, uid, crewID, slug string) (*fragment.DiscoverResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.notFound {
		return nil, fragment.ErrFragmentNotFound
	}
	return m.discRes, nil
}

func (m *mockFragmentHandler) ReplayRealm(ctx context.Context, uid, crewID, realm string) (*fragment.ReplayResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.repRes, nil
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

func TestStoryFragmentsHandler_Unauthorized(t *testing.T) {
	Setup(&mockFragmentHandler{})
	req := httptest.NewRequest(http.MethodGet, "/api/story_fragments", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestStoryFragmentsHandler_List(t *testing.T) {
	mockList := []fragment.StoryFragmentView{
		{Slug: "ancient-bark-whisper", Title: "Bisikan Pepohonan Tua", Discovered: true},
	}
	Setup(&mockFragmentHandler{fragments: mockList})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/story_fragments", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res []fragment.StoryFragmentView
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(res) != 1 || res[0].Slug != "ancient-bark-whisper" {
		t.Errorf("unexpected response: %+v", res)
	}
}

func TestStoryFragmentsHandler_DiscoverValid(t *testing.T) {
	mockRes := &fragment.DiscoverResult{
		Fragment:   fragment.StoryFragmentView{Slug: "ancient-bark-whisper", Discovered: true},
		Discovered: true,
		XPGranted:  20,
	}
	Setup(&mockFragmentHandler{discRes: mockRes})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	body := strings.NewReader(`{"slug":"ancient-bark-whisper"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/story_fragments/discover", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res fragment.DiscoverResult
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !res.Discovered || res.XPGranted != 20 {
		t.Errorf("unexpected discover res: %+v", res)
	}
}

func TestStoryFragmentsHandler_DiscoverNotFound(t *testing.T) {
	Setup(&mockFragmentHandler{notFound: true})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	body := strings.NewReader(`{"slug":"unseeded-fragment"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/story_fragments/discover", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for unseeded fragment slug, got %d", w.Code)
	}
}

func TestStoryFragmentsHandler_ReplayRealm(t *testing.T) {
	mockRep := &fragment.ReplayResult{
		Realm:         "whispering-woods",
		IsReplay:      true,
		BonusDialogue: "Selamat datang kembali!",
		UnlockedFragments: []fragment.StoryFragmentView{
			{Slug: "echo-of-the-first-explorer", Discovered: true},
		},
	}
	Setup(&mockFragmentHandler{repRes: mockRep})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	body := strings.NewReader(`{"realm":"whispering-woods"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/story_fragments/replay", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res fragment.ReplayResult
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !res.IsReplay || len(res.UnlockedFragments) != 1 {
		t.Errorf("unexpected replay res: %+v", res)
	}
}
