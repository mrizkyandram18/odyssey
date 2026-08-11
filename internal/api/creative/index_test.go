package creative

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	"odyssey/pkg/game/creative"
)

type mockCreativeHandler struct {
	subs         []creative.SubmissionView
	sub          *creative.SubmissionView
	err          error
	notFound     bool
	submitCalled bool
	submitReq    *game.Submission
}

func (m *mockCreativeHandler) Submit(ctx context.Context, uid string, req *game.Submission) (*creative.SubmissionView, error) {
	m.submitCalled = true
	m.submitReq = req
	if m.err != nil {
		return nil, m.err
	}
	if m.notFound {
		return nil, creative.ErrNotFound
	}
	if m.sub != nil {
		return m.sub, nil
	}
	return &creative.SubmissionView{
		ID:      1,
		QuestID: req.QuestID,
		Kind:    req.Kind,
		Content: req.Content,
		Status:  game.SubmissionStatusPending,
	}, nil
}

func (m *mockCreativeHandler) ListByQuest(ctx context.Context, questID int64) ([]creative.SubmissionView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.subs, nil
}

func (m *mockCreativeHandler) ListByCrew(ctx context.Context, crewID string) ([]creative.SubmissionView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.subs, nil
}

func (m *mockCreativeHandler) ListByCrewAndKind(ctx context.Context, crewID, kind string) ([]creative.SubmissionView, error) {
	if m.err != nil {
		return nil, m.err
	}
	var filtered []creative.SubmissionView
	for _, sub := range m.subs {
		if string(sub.Kind) == kind {
			filtered = append(filtered, sub)
		}
	}
	return filtered, nil
}

func (m *mockCreativeHandler) GetSubmission(ctx context.Context, submissionID int64) (*creative.SubmissionView, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.notFound {
		return nil, creative.ErrNotFound
	}
	return m.sub, nil
}

func (m *mockCreativeHandler) Approve(ctx context.Context, submissionID int64, reviewerUID string) (*creative.SubmissionView, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.notFound {
		return nil, creative.ErrNotFound
	}
	return m.sub, nil
}

func (m *mockCreativeHandler) Reject(ctx context.Context, submissionID int64, reviewerUID string, reason string) (*creative.SubmissionView, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.notFound {
		return nil, creative.ErrNotFound
	}
	return m.sub, nil
}

func makeUserToken(t *testing.T, issuer *auth.HMACSessionIssuer, role auth.Role) string {
	t.Helper()
	token, _, err := issuer.IssueSession(auth.SessionKindUser, "user-1", &auth.SessionConfig{
		Role:   role,
		CrewID: "crew-1",
	})
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func makeReviewerToken(t *testing.T, issuer *auth.HMACSessionIssuer) string {
	t.Helper()
	return makeUserToken(t, issuer, auth.RoleGuide)
}

func TestHandler_Submit_Success(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "My creative story", Status: game.SubmissionStatusPending},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	body, _ := json.Marshal(map[string]any{
		"quest_id":     1,
		"challenge_id": 1,
		"kind":         "STORY",
		"content":      "My creative story",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/creative", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}

	var resp creative.SubmissionView
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Kind != game.SubmissionStory {
		t.Errorf("expected kind STORY, got %s", resp.Kind)
	}
	if resp.Content != "My creative story" {
		t.Errorf("expected content 'My creative story', got %s", resp.Content)
	}
}

func TestHandler_Submit_InvalidBody(t *testing.T) {
	Setup(&mockCreativeHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	req := httptest.NewRequest(http.MethodPost, "/api/creative", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestHandler_ListByQuest(t *testing.T) {
	Setup(&mockCreativeHandler{
		subs: []creative.SubmissionView{
			{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusPending},
			{ID: 2, QuestID: 1, Kind: game.SubmissionComic, Content: "c1", Status: game.SubmissionStatusPending},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeReviewerToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/creative?quest_id=1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var subs []creative.SubmissionView
	if err := json.Unmarshal(w.Body.Bytes(), &subs); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 submissions, got %d", len(subs))
	}
}

func TestHandler_ListByCrewAndKind(t *testing.T) {
	Setup(&mockCreativeHandler{
		subs: []creative.SubmissionView{
			{ID: 1, CrewID: "crew-1", Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusPending},
			{ID: 2, CrewID: "crew-1", Kind: game.SubmissionComic, Content: "c1", Status: game.SubmissionStatusPending},
			{ID: 3, CrewID: "crew-1", Kind: game.SubmissionPhoto, Content: "p1", Status: game.SubmissionStatusPending},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	req := httptest.NewRequest(http.MethodGet, "/api/creative?kind=COMIC", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var subs []creative.SubmissionView
	if err := json.Unmarshal(w.Body.Bytes(), &subs); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(subs) != 1 {
		t.Fatalf("expected 1 submission, got %d", len(subs))
	}
	if subs[0].Kind != game.SubmissionComic {
		t.Errorf("expected kind COMIC, got %s", subs[0].Kind)
	}
}

func TestHandler_ListByCrew_AllWhenKindMissing(t *testing.T) {
	Setup(&mockCreativeHandler{
		subs: []creative.SubmissionView{
			{ID: 1, CrewID: "crew-1", Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusPending},
			{ID: 2, CrewID: "crew-1", Kind: game.SubmissionComic, Content: "c1", Status: game.SubmissionStatusPending},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	req := httptest.NewRequest(http.MethodGet, "/api/creative", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}

	var subs []creative.SubmissionView
	if err := json.Unmarshal(w.Body.Bytes(), &subs); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 submissions, got %d", len(subs))
	}
}

func TestHandler_GetSubmission(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusPending},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeReviewerToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/creative/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHandler_GetSubmission_ForbiddenForSeeker(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusPending, AuthorUID: "other-user"},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	req := httptest.NewRequest(http.MethodGet, "/api/creative/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for seeker, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Approve(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusApproved},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeReviewerToken(t, issuer)

	req := httptest.NewRequest(http.MethodPatch, "/api/creative/1/approve", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHandler_Approve_ForbiddenForSeeker(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusApproved},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	req := httptest.NewRequest(http.MethodPatch, "/api/creative/1/approve", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for seeker, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Reject(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusRejected, RejectionReason: "too short"},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeReviewerToken(t, issuer)

	body, _ := json.Marshal(map[string]string{"reason": "too short"})
	req := httptest.NewRequest(http.MethodPatch, "/api/creative/1/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, w.Code, w.Body.String())
	}
}

func TestHandler_Reject_ForbiddenForSeeker(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionStory, Content: "s1", Status: game.SubmissionStatusRejected, RejectionReason: "too short"},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	body, _ := json.Marshal(map[string]string{"reason": "too short"})
	req := httptest.NewRequest(http.MethodPatch, "/api/creative/1/reject", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for seeker, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandler_Unauthorized(t *testing.T) {
	Setup(nil)

	req := httptest.NewRequest(http.MethodGet, "/api/creative", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	Setup(nil)

	req := httptest.NewRequest(http.MethodDelete, "/api/creative", nil)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
	}
}

func minVideoBytes() []byte {
	return []byte{0x00, 0x00, 0x00, 0x14, 'f', 't', 'y', 'p', 'i', 's', 'o', 'm', 0x00, 0x00, 0x00, 0x00, 'i', 's', 'o', 'm'}
}

func videoSubmitBody(t *testing.T, content string) []byte {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"quest_id":     1,
		"challenge_id": 1,
		"kind":         "VIDEO",
		"content":      content,
	})
	return body
}

func TestHandler_Submit_Video_Success(t *testing.T) {
	Setup(&mockCreativeHandler{
		sub: &creative.SubmissionView{ID: 1, QuestID: 1, Kind: game.SubmissionVideo, Content: "{}", Status: game.SubmissionStatusPending},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	enc := base64.StdEncoding.EncodeToString(minVideoBytes())
	content := `{"v":1,"video":"data:video/mp4;base64,` + enc + `","caption":"clip"}`
	req := httptest.NewRequest(http.MethodPost, "/api/creative", bytes.NewReader(videoSubmitBody(t, content)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d: %s", http.StatusCreated, w.Code, w.Body.String())
	}
	var resp creative.SubmissionView
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if resp.Kind != game.SubmissionVideo {
		t.Errorf("expected kind VIDEO, got %s", resp.Kind)
	}
}

func TestHandler_Submit_Video_Invalid(t *testing.T) {
	// The mock surfaces a validator-derived error; the handler layer must map
	// every ErrVideo* to HTTP 400 (never 500).
	Setup(&mockCreativeHandler{err: creative.ErrVideoBadMagic})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, auth.RoleSeeker)

	enc := base64.StdEncoding.EncodeToString([]byte("not a video"))
	content := "{\"v\":1,\"video\":\"data:video/mp4;base64," + enc + "\"}"
	req := httptest.NewRequest(http.MethodPost, "/api/creative", bytes.NewReader(videoSubmitBody(t, content)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d (400), got %d: %s", http.StatusBadRequest, w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "video") {
		t.Errorf("expected error body to mention video, got %s", w.Body.String())
	}
}
