package admin_members

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/db"
)

type rpcMockClient struct {
	mockSupabaseClient
	rpcFunc func(ctx context.Context, fn string, payload any) ([]byte, error)
}

func (m *rpcMockClient) RPC(ctx context.Context, fn string, payload any) ([]byte, error) {
	if m.rpcFunc != nil {
		return m.rpcFunc(ctx, fn, payload)
	}
	return []byte("{}"), nil
}

func TestHandleUpdateMember_ResetDevice_Success(t *testing.T) {
	var rpcCalled bool
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				if table == "odyssey_user_profiles" {
					profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: true}}
					return json.Marshal(profiles)
				}
				return []byte("[]"), nil
			},
			mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
				return []byte(`{"status":"updated"}`), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			if fn == "odyssey_admin_reset_device" {
				rpcCalled = true
				return []byte(`{"status":"reset"}`), nil
			}
			return []byte("{}"), nil
		},
	}
	api := NewAPI(mock)
	body, _ := json.Marshal(map[string]any{"reset_device": true})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin_1", Role: "ADMIN", FamilyID: "fam_1"}))
	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body %s", w.Code, w.Body.String())
	}
	if !rpcCalled {
		t.Fatal("expected RPC odyssey_admin_reset_device to be called")
	}
}

func TestHandleUpdateMember_ResetDevice_RPCFailure(t *testing.T) {
	mock := &rpcMockClient{
		mockSupabaseClient: mockSupabaseClient{
			getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
				profiles := []db.UserProfile{{UID: "usr_target", FamilyID: "fam_1", ExplorerName: "Target", Role: "MEMBER", IsActive: true}}
				return json.Marshal(profiles)
			},
			mutateFunc: func(ctx context.Context, method string, table string, payload any, params string) ([]byte, error) {
				return []byte(`{"status":"updated"}`), nil
			},
		},
		rpcFunc: func(ctx context.Context, fn string, payload any) ([]byte, error) {
			if fn == "odyssey_admin_reset_device" {
				return nil, errors.New("rpc unavailable: function not found")
			}
			return []byte("{}"), nil
		},
	}
	api := NewAPI(mock)
	body, _ := json.Marshal(map[string]any{"reset_device": true})
	req := httptest.NewRequest(http.MethodPatch, "/api/admin/members/usr_target", bytes.NewReader(body))
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{UID: "admin_1", Role: "ADMIN", FamilyID: "fam_1"}))
	w := httptest.NewRecorder()
	api.HandleUpdateMember(w, req, "usr_target")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when RPC fails, got %d body %s", w.Code, w.Body.String())
	}
	if !bytes.Contains(w.Body.Bytes(), []byte("gagal reset device")) {
		t.Fatalf("expected error message to contain 'gagal reset device', got %s", w.Body.String())
	}
}
