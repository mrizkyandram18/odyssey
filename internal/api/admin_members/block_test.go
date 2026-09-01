package admin_members

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

func TestHandleBlockMember_Success(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: true}}
				return json.Marshal(profiles)
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/block", bytes.NewReader([]byte(`{"reason":"spam"}`)))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d body %s", w.Code, w.Body.String())
	}
	var res map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if res["success"] != true {
		t.Fatalf("expected success true %v", res)
	}
}

func TestHandleBlockMember_AlreadyBlocked(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", Role: "MEMBER", IsActive: false}}
			return json.Marshal(profiles)
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/block", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "already_blocked") {
		t.Fatalf("expected already_blocked in %s", w.Body.String())
	}
}

func TestHandleBlockMember_AdminCannotBeBlocked(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			profiles := []db.UserProfile{{UID: "admin_target", FamilyID: "fam_1", Role: "ADMIN", IsActive: true}}
			return json.Marshal(profiles)
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/admin_target/block", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for admin block got %d body %s", w.Code, w.Body.String())
	}
}

func TestHandleBlockMember_Unauthorized(t *testing.T) {
	mockClient := &mockSupabaseClient{}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/block", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), memberClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", w.Code)
	}
}

func TestHandleUnblockMember_Success(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", Role: "MEMBER", IsActive: false}}
			return json.Marshal(profiles)
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/unblock", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
}

func TestHandleUnblockMember_AlreadyActive(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", Role: "MEMBER", IsActive: true}}
			return json.Marshal(profiles)
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/usr_target/unblock", bytes.NewReader([]byte(`{}`)))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "already_active") {
		t.Fatalf("expected already_active %s", w.Body.String())
	}
}

func TestHandleAutoBlock_Success(t *testing.T) {
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			return []byte("[]"), nil
		},
	}
	// Mock RPC success via mutating?
	// Use custom client that overrides RPC
	type rpcClient struct {
		mockSupabaseClient
	}
	// Instead just test via Handler that fallback works if RPC not available
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/members/auto-block", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
}

func TestNoDeleteUserEndpoint(t *testing.T) {
	mockClient := &mockSupabaseClient{}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodDelete, "/api/admin/members/usr_target", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for DELETE (no delete endpoint) got %d", w.Code)
	}
}

func TestBlockedUserHistoryIntact(t *testing.T) {
	// Verify list members includes blocked users still visible
	mockClient := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				profiles := []map[string]any{
					{"uid": "usr_blocked", "family_id": "fam_1", "explorer_name": "Blocked User", "role": "MEMBER", "is_active": false, "blocked_at": "2026-09-01T00:00:00Z", "block_reason": "spam"},
					{"uid": "usr_active", "family_id": "fam_1", "explorer_name": "Active User", "role": "MEMBER", "is_active": true},
				}
				return json.Marshal(profiles)
			}
			if table == "odyssey_local_users" {
				return json.Marshal([]map[string]any{
					{"username": "blocked", "profile_uid": "usr_blocked"},
					{"username": "active", "profile_uid": "usr_active"},
				})
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(mockClient)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/members", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims("fam_1")))
	w := httptest.NewRecorder()
	api.Handler(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d %s", w.Code, w.Body.String())
	}
	var res struct {
		Items []MemberView `json:"items"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	if len(res.Items) != 2 {
		t.Fatalf("expected 2 members including blocked, got %d", len(res.Items))
	}
	foundBlocked := false
	for _, m := range res.Items {
		if m.UID == "usr_blocked" && !m.IsActive {
			foundBlocked = true
			if m.BlockReason == nil || *m.BlockReason != "spam" {
				t.Fatalf("expected block_reason spam got %v", m.BlockReason)
			}
		}
	}
	if !foundBlocked {
		t.Fatalf("blocked user not found in list")
	}
}
