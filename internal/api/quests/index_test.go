package quests

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
	"odyssey/pkg/game/quest"
)

type mockQuestHandler struct {
	quests   []quest.QuestView
	quest    *quest.QuestWithChallenges
	err      error
	notFound bool
	// mutation tracking
	startedQuestID     int64
	startedErr         error
	completedQuestID   int64
	completedChallenge int64
	completeResult     *quest.CompleteChallengeResult
	completeErr        error
	selectedBranch     string
	selectBranchErr    error
}

func (m *mockQuestHandler) List(ctx context.Context, crewID string) ([]quest.QuestView, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quests, nil
}

func (m *mockQuestHandler) GetByCrewAndID(ctx context.Context, questID int64, crewID string) (*quest.QuestWithChallenges, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.notFound {
		return nil, quest.ErrQuestNotFound
	}
	return m.quest, nil
}

func (m *mockQuestHandler) StartQuest(ctx context.Context, questID int64, crewID string) error {
	if m.startedErr != nil {
		return m.startedErr
	}
	if m.notFound {
		return quest.ErrQuestNotFound
	}
	m.startedQuestID = questID
	return nil
}

func (m *mockQuestHandler) CompleteChallenge(ctx context.Context, questID, challengeID int64, crewID, uid string) (*quest.CompleteChallengeResult, error) {
	if m.completeErr != nil {
		return nil, m.completeErr
	}
	if m.notFound {
		return nil, quest.ErrQuestNotFound
	}
	m.completedQuestID = questID
	m.completedChallenge = challengeID
	return m.completeResult, nil
}

func (m *mockQuestHandler) SelectBranch(ctx context.Context, questID int64, crewID, branchChoice string) (*quest.SelectBranchResult, error) {
	if m.selectBranchErr != nil {
		return nil, m.selectBranchErr
	}
	if m.notFound {
		return nil, quest.ErrQuestNotFound
	}
	m.selectedBranch = branchChoice
	return &quest.SelectBranchResult{
		Success:     true,
		StoryBranch: branchChoice,
		Realm:       "clockwork-city",
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

func makeQuestView() quest.QuestView {
	return quest.QuestView{
		Quest: game.Quest{
			ID:        1,
			CrewID:    "crew-1",
			Title:     "Test Quest",
			Status:    "PENDING",
			CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		},
		ChallengeCount: 3,
		CompletedCount: 1,
	}
}

func makeQuestWithChallenges() *quest.QuestWithChallenges {
	return &quest.QuestWithChallenges{
		Quest: game.Quest{
			ID:        1,
			CrewID:    "crew-1",
			Title:     "Test Quest",
			Status:    "ACTIVE",
			CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
		},
		Challenges: []game.Challenge{
			{ID: 10, QuestID: 1, Slug: "obs", Description: "Observe", Status: "DONE"},
			{ID: 11, QuestID: 1, Slug: "research", Description: "Research", Status: "PENDING"},
		},
	}
}

func TestQuestsHandler_MethodNotAllowed(t *testing.T) {
	Setup(&mockQuestHandler{})
	req := httptest.NewRequest(http.MethodPost, "/api/quests", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", w.Code)
	}
}

func TestQuestsHandler_Unauthorized(t *testing.T) {
	Setup(&mockQuestHandler{quests: []quest.QuestView{makeQuestView()}})

	req := httptest.NewRequest(http.MethodGet, "/api/quests", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestQuestsHandler_NotConfigured(t *testing.T) {
	Setup(nil)
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/quests", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", w.Code)
	}
}

func TestQuestsHandler_List(t *testing.T) {
	Setup(&mockQuestHandler{quests: []quest.QuestView{makeQuestView()}})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/quests", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var quests []quest.QuestView
	if err := json.Unmarshal(w.Body.Bytes(), &quests); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(quests) != 1 {
		t.Fatalf("expected 1 quest, got %d", len(quests))
	}
	if quests[0].Title != "Test Quest" {
		t.Errorf("expected title 'Test Quest', got %s", quests[0].Title)
	}
	if quests[0].ChallengeCount != 3 {
		t.Errorf("expected 3 challenges, got %d", quests[0].ChallengeCount)
	}
	if quests[0].CompletedCount != 1 {
		t.Errorf("expected 1 completed, got %d", quests[0].CompletedCount)
	}
}

func TestQuestsHandler_GetByID(t *testing.T) {
	Setup(&mockQuestHandler{quest: makeQuestWithChallenges()})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/quests/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var q quest.QuestWithChallenges
	if err := json.Unmarshal(w.Body.Bytes(), &q); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if q.ID != 1 {
		t.Errorf("expected ID 1, got %d", q.ID)
	}
	if len(q.Challenges) != 2 {
		t.Errorf("expected 2 challenges, got %d", len(q.Challenges))
	}
}

func TestQuestsHandler_GetByID_NotFound(t *testing.T) {
	Setup(&mockQuestHandler{notFound: true})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/quests/999", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestQuestsHandler_GetByID_InvalidID(t *testing.T) {
	Setup(&mockQuestHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/quests/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestQuestsHandler_ListError(t *testing.T) {
	Setup(&mockQuestHandler{err: quest.ErrQuestNotFound})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/quests", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestQuestsHandler_StartQuest(t *testing.T) {
	Setup(&mockQuestHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/quests/1/start", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	mh := handler.(*mockQuestHandler)
	if mh.startedQuestID != 1 {
		t.Errorf("expected StartQuest called with id 1, got %d", mh.startedQuestID)
	}
}

func TestQuestsHandler_StartQuest_NotFound(t *testing.T) {
	Setup(&mockQuestHandler{notFound: true})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/quests/999/start", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestQuestsHandler_StartQuest_InvalidID(t *testing.T) {
	Setup(&mockQuestHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/quests/abc/start", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestQuestsHandler_CompleteChallenge(t *testing.T) {
	result := &quest.CompleteChallengeResult{
		Quest:          makeQuestWithChallenges(),
		QuestCompleted: true,
		XP:             80,
		NewLevel:       2,
		LevelUp:        true,
	}
	Setup(&mockQuestHandler{completeResult: result})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/quests/1/challenges/11/complete", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var r quest.CompleteChallengeResult
	if err := json.Unmarshal(w.Body.Bytes(), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !r.QuestCompleted {
		t.Error("expected quest_completed true")
	}
	if r.XP != 80 {
		t.Errorf("expected XP 80, got %d", r.XP)
	}
	mh := handler.(*mockQuestHandler)
	if mh.completedQuestID != 1 || mh.completedChallenge != 11 {
		t.Errorf("expected complete called with quest=1 challenge=11, got quest=%d challenge=%d", mh.completedQuestID, mh.completedChallenge)
	}
}

func TestQuestsHandler_CompleteChallenge_NotFound(t *testing.T) {
	Setup(&mockQuestHandler{notFound: true})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/quests/1/challenges/11/complete", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestQuestsHandler_CompleteChallenge_InvalidIDs(t *testing.T) {
	Setup(&mockQuestHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/quests/x/challenges/y/complete", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid ids, got %d", w.Code)
	}
}

func TestQuestsHandler_ImplementsInterface(t *testing.T) {
	var _ QuestHandler = (*mockQuestHandler)(nil)
}

func TestQuestsHandler_SelectBranch_Valid(t *testing.T) {
	Setup(&mockQuestHandler{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	body := strings.NewReader(`{"branch":"path-of-copper"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/quests/1/branch", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var res quest.SelectBranchResult
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !res.Success || res.StoryBranch != "path-of-copper" {
		t.Errorf("unexpected res: %+v", res)
	}
}

func TestQuestsHandler_SelectBranch_InvalidChoice(t *testing.T) {
	Setup(&mockQuestHandler{selectBranchErr: quest.ErrInvalidBranchChoice})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	body := strings.NewReader(`{"branch":"invalid-branch"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/quests/1/branch", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid branch choice, got %d", w.Code)
	}
}

func TestQuestsHandler_SelectBranch_LinearQuestNoBranch(t *testing.T) {
	Setup(&mockQuestHandler{selectBranchErr: quest.ErrNoBranchOptions})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	body := strings.NewReader(`{"branch":"some-branch"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/quests/1/branch", body)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for quest with no branch options, got %d", w.Code)
	}
}

func TestQuestsHandler_SelectBranch_Unauthorized(t *testing.T) {
	Setup(&mockQuestHandler{})

	body := strings.NewReader(`{"branch":"path-of-copper"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/quests/1/branch", body)
	w := httptest.NewRecorder()
	Handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 unauthorized, got %d", w.Code)
	}
}
