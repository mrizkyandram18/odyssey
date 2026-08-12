package gifts

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	gamechest "odyssey/pkg/game/gift"
)

type mockChestService struct {
	gifts []gamechest.ChestView
	result *gamechest.OpenResult
	err    error
}

func (m *mockChestService) ListChests(ctx context.Context, uid string) ([]game.Gift, error) {
	if m.err != nil {
		return nil, m.err
	}
	gifts := make([]game.Gift, len(m.gifts))
	for i, c := range m.gifts {
		gifts[i] = game.Gift{
			ID: c.ID, UID: c.UID, Source: c.Source, GiftSlug: c.GiftSlug,
			Rarity: string(c.Rarity), Icon: c.Icon, Description: c.Description,
			Opened: c.Opened, OpenedAt: c.OpenedAt, CreatedAt: c.CreatedAt,
		}
	}
	return gifts, nil
}

func (m *mockChestService) GetChest(ctx context.Context, chestID int64, uid string) (*game.Gift, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, c := range m.gifts {
		if c.ID == chestID && c.UID == uid {
			return &game.Gift{
				ID: c.ID, UID: c.UID, Source: c.Source, GiftSlug: c.GiftSlug,
				Rarity: string(c.Rarity), Icon: c.Icon, Description: c.Description,
				Opened: c.Opened, OpenedAt: c.OpenedAt, CreatedAt: c.CreatedAt,
			}, nil
		}
	}
	return nil, gamechest.ErrChestNotFound
}

func (m *mockChestService) OpenChest(ctx context.Context, chestID int64, uid string) (*gamechest.OpenResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.result, nil
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

func TestChestsHandler_Unauthorized(t *testing.T) {
	Setup(&mockChestService{})
	req := httptest.NewRequest(http.MethodGet, "/api/gifts", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestChestsHandler_ListSuccess(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	Setup(&mockChestService{
		gifts: []gamechest.ChestView{
			{ID: 1, UID: "user-1", Source: "quest", GiftSlug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: false, CreatedAt: now},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/gifts", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var gifts []gamechest.ChestView
	if err := json.Unmarshal(w.Body.Bytes(), &gifts); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(gifts) != 1 {
		t.Errorf("expected 1 chest, got %d", len(gifts))
	}
}

func TestChestsHandler_GetSuccess(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	Setup(&mockChestService{
		gifts: []gamechest.ChestView{
			{ID: 1, UID: "user-1", Source: "quest", GiftSlug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: false, CreatedAt: now},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/gifts/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestChestsHandler_OpenSuccess(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	Setup(&mockChestService{
		gifts: []gamechest.ChestView{
			{ID: 1, UID: "user-1", Source: "quest", GiftSlug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: false, CreatedAt: now},
		},
		result: &gamechest.OpenResult{
			Gift: &gamechest.ChestView{
				ID: 1, UID: "user-1", Source: "quest", GiftSlug: "wooden-chest", Name: "Wooden Gift", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: true, OpenedAt: &now, CreatedAt: now,
			},
			Rewards:        []gamechest.RewardItem{{CollectionSlug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", IsNew: true}},
			NewCount:       1,
			DuplicateCount: 0,
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/gifts/1/open", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var result gamechest.OpenResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result.NewCount != 1 {
		t.Errorf("expected new_count 1, got %d", result.NewCount)
	}
}
