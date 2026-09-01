package shop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
	"odyssey/pkg/shared"
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

type mockSupabaseClientDynamic struct {
	GetFn func(ctx context.Context, table string, params string) ([]byte, error)
	RPCFn func(ctx context.Context, fnName string, payload any) ([]byte, error)
}

func (m *mockSupabaseClientDynamic) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.GetFn != nil {
		return m.GetFn(ctx, table, params)
	}
	return nil, nil
}

func (m *mockSupabaseClientDynamic) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return nil, nil
}

func (m *mockSupabaseClientDynamic) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return nil, nil
}

func (m *mockSupabaseClientDynamic) RPC(ctx context.Context, fnName string, payload any) ([]byte, error) {
	if m.RPCFn != nil {
		return m.RPCFn(ctx, fnName, payload)
	}
	return nil, nil
}

func (m *mockSupabaseClientDynamic) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
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

func TestShopConfig_Get(t *testing.T) {
	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": "10"},
		{"key": "redemption_end_day", "value": "20"},
	})
	client := &mockSupabaseClient{
		getResp: configData,
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodGet, "/api/shop/config", nil)
	claims := &auth.SessionClaims{UID: "user-1", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var res map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}
	if res["redemption_start_day"] != float64(10) || res["redemption_end_day"] != float64(20) {
		t.Fatalf("expected start 10, end 20, got %v", res)
	}
	if res["conversion_rate"] != float64(100) {
		t.Fatalf("expected conversion_rate 100, got %v", res["conversion_rate"])
	}
	if res["max_payout_coins"] != float64(3200) {
		t.Fatalf("expected max_payout_coins 3200, got %v", res["max_payout_coins"])
	}
}

func TestShopRedeem_OpenWindow_Success(t *testing.T) {
	var capturedPayload shared.TelegramPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &capturedPayload)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"result":{"message_id":99,"chat":{"id":"test-chat"}}}`))
	}))
	defer ts.Close()

	shared.SetTelegramBaseURLForTest(ts.URL)
	defer shared.SetTelegramBaseURLForTest("")

	t.Setenv("TELEGRAM_BOT_TOKEN", "mock-token")
	t.Setenv("TELEGRAM_CHAT_ID", "mock-chat")

	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": "1"},
		{"key": "redemption_end_day", "value": "31"},
		{"key": "coin_conversion_rate", "value": "100"},
	})
	profileData, _ := json.Marshal([]map[string]any{
		{"explorer_name": "Explorer Budi"},
	})
	client := &mockSupabaseClient{
		getResp: configData,
		rpcResp: []byte(`{"success":true,"claim_id":12,"new_balance":500}`),
	}
	// Dynamic get response based on table
	clientGet := func(ctx context.Context, table string, params string) ([]byte, error) {
		if strings.Contains(table, "odyssey_user_profiles") {
			return profileData, nil
		}
		return configData, nil
	}
	api := NewAPI(&mockSupabaseClientDynamic{GetFn: clientGet, RPCFn: client.RPC})

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

	// Verify Telegram payload content
	if !strings.Contains(capturedPayload.Text, "#12") {
		t.Errorf("expected Telegram text to contain Claim ID #12, got: %s", capturedPayload.Text)
	}
	if !strings.Contains(capturedPayload.Text, "Explorer Budi") {
		t.Errorf("expected Telegram text to contain Explorer Name, got: %s", capturedPayload.Text)
	}
	if !strings.Contains(capturedPayload.Text, "1.000 Koin (Rp 100.000)") {
		t.Errorf("expected Telegram text to contain formatted coins & rupiah, got: %s", capturedPayload.Text)
	}
	if !strings.Contains(capturedPayload.Text, "CONFIRMED") {
		t.Errorf("expected Telegram text to contain confirmation info, got: %s", capturedPayload.Text)
	}
}

func TestShopRedeem_ClosedWindow_Rejected(t *testing.T) {
	// Configure window to day 10-12 (closed when day is outside 10-12)
	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": "10"},
		{"key": "redemption_end_day", "value": "12"},
	})
	client := &mockSupabaseClient{
		getResp: configData,
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

	// Outside [10, 12], must be rejected with 400 Bad Request
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request for closed window, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "Periode penukaran koin saat ini ditutup") {
		t.Fatalf("expected closed message, got %s", rec.Body.String())
	}
}

func TestShopRedeem_TelegramFailure_DoesNotFailClaim(t *testing.T) {
	// Telegram endpoint fails with 500 error
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	shared.SetTelegramBaseURLForTest(ts.URL)
	defer shared.SetTelegramBaseURLForTest("")

	t.Setenv("TELEGRAM_BOT_TOKEN", "mock-token")
	t.Setenv("TELEGRAM_CHAT_ID", "mock-chat")

	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": "1"},
		{"key": "redemption_end_day", "value": "31"},
		{"key": "coin_conversion_rate", "value": "100"},
	})
	client := &mockSupabaseClient{
		getResp: configData,
		rpcResp: []byte(`{"success":true,"claim_id":99,"new_balance":400}`),
	}
	api := NewAPI(client)

	body := `{"coins":500,"target_type":"BANK","target_value":"BCA - 1234567890 (a.n Ani)"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-456", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	// Claim creation must still succeed with 200 OK even if Telegram notification fails
	if rec.Code != http.StatusOK {
		t.Fatalf("expected claim to succeed with status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestShopRedeem_DuplicatePending_Rejected(t *testing.T) {
	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": "1"},
		{"key": "redemption_end_day", "value": "31"},
	})
	client := &mockSupabaseClient{
		getResp: configData,
		rpcErr:  fmt.Errorf("Anda masih memiliki klaim pending yang belum diproses"),
	}
	api := NewAPI(client)

	body := `{"coins":500,"target_type":"BANK","target_value":"BCA - 1234567890"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-456", Role: "SEEKER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 Bad Request on duplicate pending claim, got %d: %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "klaim pending") {
		t.Errorf("expected duplicate error message in response body, got %s", rec.Body.String())
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
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}
