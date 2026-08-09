package reactions

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
)

type mockReactionService struct {
	addedReaction *game.Reaction
	addErr        error
	listReactions []game.Reaction
	listErr       error
}

func (m *mockReactionService) AddReaction(ctx context.Context, crewID, actorUID string, targetType string, targetID int64, reactionType string) (*game.Reaction, error) {
	if m.addErr != nil {
		return nil, m.addErr
	}
	r := &game.Reaction{
		CrewID:       crewID,
		TargetType:   targetType,
		TargetID:     targetID,
		ActorUID:     actorUID,
		ReactionType: reactionType,
	}
	m.addedReaction = r
	return r, nil
}

func (m *mockReactionService) ListReactionsForTarget(ctx context.Context, crewID, targetType string, targetID int64) ([]game.Reaction, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.listReactions, nil
}

func makeUserToken(t *testing.T, issuer *auth.HMACSessionIssuer, uid, crewID string) string {
	t.Helper()
	token, _, err := issuer.IssueSession(auth.SessionKindUser, uid, &auth.SessionConfig{
		Role:   auth.RoleSeeker,
		CrewID: crewID,
	})
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func TestReactionsHandler_SpoofRejection(t *testing.T) {
	// Evidence-first test: Ensure that actor_uid sent by client is IGNORED,
	// and the backend uses the claims.UID.
	svc := &mockReactionService{}
	Setup(svc)

	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	
	// Real token is for "user-123" in "crew-A"
	token := makeUserToken(t, issuer, "user-123", "crew-A")

	// Client maliciously tries to react as "user-999" (a different family member)
	payload := map[string]any{
		"target_type":   "JOURNAL",
		"target_id":     10,
		"reaction_type": "HEART",
		"actor_uid":     "user-999", // Spoof attempt!
	}
	
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest(http.MethodPost, "/api/reactions", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	if svc.addedReaction == nil {
		t.Fatalf("expected AddReaction to be called")
	}

	// Verify that the backend safely ignored "user-999" and used "user-123"
	if svc.addedReaction.ActorUID != "user-123" {
		t.Errorf("SPOOF FAILURE: expected actor_uid to be securely extracted as 'user-123', got '%s'", svc.addedReaction.ActorUID)
	}
	
	// Verify crew isolation is maintained
	if svc.addedReaction.CrewID != "crew-A" {
		t.Errorf("expected crew_id to be 'crew-A', got '%s'", svc.addedReaction.CrewID)
	}
}

func TestReactionsHandler_List(t *testing.T) {
	svc := &mockReactionService{
		listReactions: []game.Reaction{
			{ID: 1, ReactionType: "CLAP", ActorUID: "user-123"},
		},
	}
	Setup(svc)

	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, "user-456", "crew-A")

	req := httptest.NewRequest(http.MethodGet, "/api/reactions?target_type=QUEST&target_id=5", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", w.Code)
	}

	var res map[string][]game.Reaction
	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("unmarshal err: %v", err)
	}

	reactions := res["reactions"]
	if len(reactions) != 1 || reactions[0].ReactionType != "CLAP" {
		t.Errorf("unexpected response: %v", reactions)
	}
}

func TestReactionsHandler_ListMissingParams(t *testing.T) {
	Setup(&mockReactionService{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer, "user-456", "crew-A")

	// Missing target_id
	req := httptest.NewRequest(http.MethodGet, "/api/reactions?target_type=QUEST", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 Bad Request, got %d", w.Code)
	}
}
