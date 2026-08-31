package shop

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

type mockSupabaseClient struct {
	getResp    []byte
	getErr     error
	mutateResp []byte
	mutateErr  error
	rpcResp    []byte
	rpcErr     error
	uploadResp string
	uploadErr  error
}

func (m *mockSupabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.getResp, nil
}

func (m *mockSupabaseClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	if m.mutateErr != nil {
		return nil, m.mutateErr
	}
	return m.mutateResp, nil
}

func (m *mockSupabaseClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return m.Mutate(ctx, method, table, payload, params)
}

func (m *mockSupabaseClient) RPC(ctx context.Context, fnName string, payload any) ([]byte, error) {
	if m.rpcErr != nil {
		return nil, m.rpcErr
	}
	return m.rpcResp, nil
}

func (m *mockSupabaseClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	if m.uploadErr != nil {
		return "", m.uploadErr
	}
	return m.uploadResp, nil
}

func TestShopRedeemSuccess(t *testing.T) {
	client := &mockSupabaseClient{
		rpcResp: []byte(`{"success":true,"claim_id":12,"new_balance":500}`),
	}
	api := NewAPI(client)

	body := `{"reward_id":1,"coins":1000,"target_type":"EWALLET","target_value":"GoPay - 0812345678"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-123", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminListClaims_NonGuideForbidden(t *testing.T) {
	client := &mockSupabaseClient{}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/claims", nil)
	claims := &auth.SessionClaims{UID: "user-123", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403 Forbidden for non-guide user, got %d", rec.Code)
	}
}

func TestAdminProcessClaim_Success(t *testing.T) {
	claimData, _ := json.Marshal([]map[string]any{
		{"user_uid": "member-1"},
	})
	client := &mockSupabaseClient{
		getResp: claimData,
		rpcResp: []byte(`{"success":true,"status":"APPROVED"}`),
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/claims/5/process", strings.NewReader(`{"status":"APPROVED","notes":"Transfer pulsa sukses"}`))
	req.Header.Set("Content-Type", "application/json")
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "GUIDE"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}
