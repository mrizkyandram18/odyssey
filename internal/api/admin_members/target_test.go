package admin_members

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

func TestTarget_CreateMember_Target3200Preserved(t *testing.T) {
	// Verify default target 3200 behavior (existing user compatibility)
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			return []byte("[]"), nil
		},
		mutateAtomicFunc: func(ctx context.Context, method string, table string, payload any, params string, extraHeader string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				m := payload.(map[string]any)
				if m["monthly_coin_target"] != 3200 {
					t.Fatalf("expected default target 3200, got %v", m["monthly_coin_target"])
				}
				return json.Marshal([]map[string]any{m})
			}
			return []byte(`[{"id":1}]`), nil
		},
	}
	api := NewAPI(mockClient)
	body, _ := json.Marshal(map[string]any{"username": "user1", "password": "secret123", "explorer_name": "User 1"})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin_1", Role: "ADMIN", FamilyID: "fam_1"}))
	w := httptest.NewRecorder()
	api.HandleCreateMember(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 got %d %s", w.Code, w.Body.String())
	}
}

func TestTarget_CreateMember_InvalidZero(t *testing.T) {
	mockClient := &mockSupabaseClient{}
	api := NewAPI(mockClient)
	body, _ := json.Marshal(map[string]any{"username": "user2", "password": "secret123", "explorer_name": "X", "monthly_coin_target": 0})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin_1", Role: "ADMIN", FamilyID: "fam_1"}))
	w := httptest.NewRecorder()
	api.HandleCreateMember(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for 0 got %d", w.Code)
	}
}

func TestTarget_UpdateMember_TargetValidation(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			p := []db.UserProfile{{UID: "usr1", FamilyID: "fam_1", Role: "MEMBER"}}
			return json.Marshal(p)
		},
	}
	api := NewAPI(mockClient)
	body, _ := json.Marshal(map[string]any{"monthly_coin_target": 10001})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr1", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin_1", Role: "ADMIN", FamilyID: "fam_1"}))
	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr1")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for 10001 got %d", w.Code)
	}
}
