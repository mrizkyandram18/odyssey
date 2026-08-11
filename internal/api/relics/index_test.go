package relics

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/auth"
	gamerelic "odyssey/pkg/game/relic"
)

type mockRelicService struct {
	relics    []gamerelic.RelicDefinition
	inventory []gamerelic.InventoryItem
	err       error
	giftRes   *gamerelic.GiftResult
	giftErr   error
}

func (m *mockRelicService) ListRelics(ctx context.Context) ([]gamerelic.RelicDefinition, error) {
	return m.relics, m.err
}

func (m *mockRelicService) GetRelic(ctx context.Context, slug string) (*gamerelic.RelicDefinition, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, r := range m.relics {
		if r.Slug == slug {
			return &r, nil
		}
	}
	return nil, gamerelic.ErrRelicNotFound
}

func (m *mockRelicService) ListInventory(ctx context.Context, uid string) ([]gamerelic.InventoryItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.inventory, nil
}

func (m *mockRelicService) GiftRelic(ctx context.Context, senderUID, recipientUID, relicSlug, crewID string) (*gamerelic.GiftResult, error) {
	if m.giftErr != nil {
		return nil, m.giftErr
	}
	return m.giftRes, nil
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

func TestRelicsHandler_Unauthorized(t *testing.T) {
	Setup(&mockRelicService{})
	req := httptest.NewRequest(http.MethodGet, "/api/relics", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRelicsHandler_InventorySuccess(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	Setup(&mockRelicService{
		inventory: []gamerelic.InventoryItem{
			{RelicSlug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", Image: "🧭", OwnedCount: 1, IsNew: true, DiscoveredAt: now, CreatedAt: now},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/relics/inventory", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var items []gamerelic.InventoryItem
	if err := json.Unmarshal(w.Body.Bytes(), &items); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(items) != 1 {
		t.Errorf("expected 1 item, got %d", len(items))
	}
}

func TestRelicsHandler_GetRelicSuccess(t *testing.T) {
	Setup(&mockRelicService{
		relics: []gamerelic.RelicDefinition{
			{ID: 1, Slug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", Image: "🧭"},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/relics/ancient-compass", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var def gamerelic.RelicDefinition
	if err := json.Unmarshal(w.Body.Bytes(), &def); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if def.Slug != "ancient-compass" {
		t.Errorf("expected ancient-compass, got %s", def.Slug)
	}
}

func TestRelicsHandler_GetRelic_NotFound(t *testing.T) {
	Setup(&mockRelicService{})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/relics/nonexistent", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestRelicsHandler_GiftSuccess(t *testing.T) {
	Setup(&mockRelicService{
		giftRes: &gamerelic.GiftResult{
			RelicSlug:     "ancient-compass",
			RelicName:     "Ancient Compass",
			RecipientUID:  "recipient-uid",
			RecipientName: "Aria",
			SenderCount:   1,
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	body := `{"recipient_uid":"recipient-uid","relic_slug":"ancient-compass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/relics/gift", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestRelicsHandler_GiftErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		code int
	}{
		{gamerelic.ErrSelfGift, http.StatusBadRequest},
		{gamerelic.ErrCrossCrewGift, http.StatusForbidden},
		{gamerelic.ErrRecipientNotFound, http.StatusNotFound},
		{gamerelic.ErrRelicNotOwned, http.StatusConflict},
		{gamerelic.ErrRelicNotFound, http.StatusNotFound},
		{context.DeadlineExceeded, http.StatusInternalServerError},
	}

	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	for _, tc := range cases {
		Setup(&mockRelicService{giftErr: tc.err})
		body := `{"recipient_uid":"recipient-uid","relic_slug":"ancient-compass"}`
		req := httptest.NewRequest(http.MethodPost, "/api/relics/gift", strings.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		mw.RequireAuth(Handler)(w, req)

		if w.Code != tc.code {
			t.Errorf("expected %d for error %v, got %d", tc.code, tc.err, w.Code)
		}
	}
}
