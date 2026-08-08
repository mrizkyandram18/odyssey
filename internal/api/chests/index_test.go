package chests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/game"
	gamechest "odyssey/pkg/game/chest"
)

type mockChestService struct {
	chests []gamechest.ChestView
	result *gamechest.OpenResult
	err    error
}

func (m *mockChestService) ListChests(ctx context.Context, uid string) ([]game.Chest, error) {
	if m.err != nil {
		return nil, m.err
	}
	chests := make([]game.Chest, len(m.chests))
	for i, c := range m.chests {
		chests[i] = game.Chest{
			ID: c.ID, UID: c.UID, Source: c.Source, ChestSlug: c.ChestSlug,
			Rarity: string(c.Rarity), Icon: c.Icon, Description: c.Description,
			Opened: c.Opened, OpenedAt: c.OpenedAt, CreatedAt: c.CreatedAt,
		}
	}
	return chests, nil
}

func (m *mockChestService) GetChest(ctx context.Context, chestID int64, uid string) (*game.Chest, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, c := range m.chests {
		if c.ID == chestID && c.UID == uid {
			return &game.Chest{
				ID: c.ID, UID: c.UID, Source: c.Source, ChestSlug: c.ChestSlug,
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
		CrewID: "crew-1",
	})
	if err != nil {
		t.Fatalf("IssueSession: %v", err)
	}
	return token
}

func TestChestsHandler_Unauthorized(t *testing.T) {
	Setup(&mockChestService{})
	req := httptest.NewRequest(http.MethodGet, "/api/chests", nil)
	w := httptest.NewRecorder()
	Handler(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestChestsHandler_ListSuccess(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	Setup(&mockChestService{
		chests: []gamechest.ChestView{
			{ID: 1, UID: "user-1", Source: "quest", ChestSlug: "wooden-chest", Name: "Wooden Chest", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: false, CreatedAt: now},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/chests", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mw.RequireAuth(Handler)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var chests []gamechest.ChestView
	if err := json.Unmarshal(w.Body.Bytes(), &chests); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(chests) != 1 {
		t.Errorf("expected 1 chest, got %d", len(chests))
	}
}

func TestChestsHandler_GetSuccess(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	Setup(&mockChestService{
		chests: []gamechest.ChestView{
			{ID: 1, UID: "user-1", Source: "quest", ChestSlug: "wooden-chest", Name: "Wooden Chest", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: false, CreatedAt: now},
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodGet, "/api/chests/1", nil)
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
		chests: []gamechest.ChestView{
			{ID: 1, UID: "user-1", Source: "quest", ChestSlug: "wooden-chest", Name: "Wooden Chest", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: false, CreatedAt: now},
		},
		result: &gamechest.OpenResult{
			Chest: &gamechest.ChestView{
				ID: 1, UID: "user-1", Source: "quest", ChestSlug: "wooden-chest", Name: "Wooden Chest", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: true, OpenedAt: &now, CreatedAt: now,
			},
			Rewards:        []gamechest.RewardItem{{RelicSlug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", IsNew: true}},
			NewCount:       1,
			DuplicateCount: 0,
		},
	})
	issuer := auth.NewSessionIssuer("test-secret")
	mw := auth.NewMiddleware(issuer)
	token := makeUserToken(t, issuer)

	req := httptest.NewRequest(http.MethodPost, "/api/chests/1/open", nil)
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
