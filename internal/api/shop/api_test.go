package shop

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/push"
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
	// Deterministic closed window: a single day that is never today, derived
	// from the wall clock in the same timezone the redeem path evaluates
	// (shared.DefaultTimezone). A fixed window such as day 10-12 is open
	// whenever the test happens to run inside it, which made this test
	// date-dependent. closedDay is always a valid day (1-28) and by
	// construction differs from every possible today (1-31).
	loc, err := time.LoadLocation(shared.DefaultTimezone)
	if err != nil {
		loc = time.UTC
	}
	today := time.Now().In(loc).Day()
	closedDay := today%28 + 1
	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": strconv.Itoa(closedDay)},
		{"key": "redemption_end_day", "value": strconv.Itoa(closedDay)},
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

	// Today is always outside the single closedDay, so redeem must be
	// rejected with 400 Bad Request on every calendar day.
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

type fakeClaimNotifier struct {
	uid string
	n   push.Notification
}

func (f *fakeClaimNotifier) SendToUser(ctx context.Context, uid string, n push.Notification) error {
	f.uid = uid
	f.n = n
	return nil
}

func TestAdminProcessClaim_SendsPushNotification(t *testing.T) {
	claimData, _ := json.Marshal([]map[string]any{
		{"user_uid": "member-1", "coins_redeemed": 3500},
	})
	rpcStatus := "APPROVED"
	client := &mockSupabaseClient{
		getResp: claimData,
		rpcResp: []byte(`{"success":true}`),
		rpcErr:  nil,
	}
	notifier := &fakeClaimNotifier{}
	api := NewAPI(client)
	api.SetPushNotifier(notifier)

	mkReq := func() *http.Request {
		req := httptest.NewRequest(http.MethodPost, "/api/admin/claims/5/process", strings.NewReader(`{"status":"`+rpcStatus+`","notes":"catatan uji"}`))
		req.Header.Set("Content-Type", "application/json")
		guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
		return req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))
	}

	rpcStatus = "APPROVED"
	rec := httptest.NewRecorder()
	api.Handler(rec, mkReq())
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	if notifier.uid != "member-1" {
		t.Fatalf("expected push to member-1, got %q", notifier.uid)
	}
	if notifier.n.Title != "Pencairan Disetujui 🎉" || !strings.Contains(notifier.n.Body, "3.500 Koin") || !strings.Contains(notifier.n.URL, "/shop") {
		t.Fatalf("unexpected approved notification: %+v", notifier.n)
	}

	notifier.uid = ""
	notifier.n = push.Notification{}
	rpcStatus = "REJECTED"
	rec = httptest.NewRecorder()
	api.Handler(rec, mkReq())
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
	if notifier.n.Title != "Pencairan Ditolak" || !strings.Contains(notifier.n.Body, "dikembalikan ke saldo") || !strings.Contains(notifier.n.Body, "catatan uji") {
		t.Fatalf("unexpected rejected notification: %+v", notifier.n)
	}
}

func TestAdminProcessClaim_NoPushWhenNotifierUnset(t *testing.T) {
	claimData, _ := json.Marshal([]map[string]any{
		{"user_uid": "member-1", "coins_redeemed": 3500},
	})
	client := &mockSupabaseClient{
		getResp: claimData,
		rpcResp: []byte(`{"success":true}`),
	}
	api := NewAPI(client)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/claims/5/process", strings.NewReader(`{"status":"APPROVED"}`))
	req.Header.Set("Content-Type", "application/json")
	guideClaims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), guideClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminListClaims_TenantIsolationAndPagination(t *testing.T) {
	var requestedProfileParams string
	var requestedClaimParams string
	var requestedCatParams string

	clientGet := func(ctx context.Context, table string, params string) ([]byte, error) {
		if table == "odyssey_user_profiles" {
			requestedProfileParams = params
			profiles := []map[string]any{
				{"uid": "usr_claim_1", "explorer_name": "Claim User"},
			}
			return json.Marshal(profiles)
		}
		if table == "odyssey_claims" {
			requestedClaimParams = params
			rewardID := int64(1)
			claims := []map[string]any{
				{
					"id":             int64(77),
					"user_uid":       "usr_claim_1",
					"reward_id":      &rewardID,
					"coins_redeemed": 100,
					"target_type":    "EWALLET",
					"target_value":   "GoPay 08123",
					"status":         "PENDING",
					"created_at":     "2026-09-01T12:00:00Z",
				},
			}
			return json.Marshal(claims)
		}
		if table == "odyssey_reward_catalog" {
			requestedCatParams = params
			cats := []map[string]any{
				{"id": int64(1), "title": "Pulsa 10k"},
			}
			return json.Marshal(cats)
		}
		return []byte("[]"), nil
	}

	api := NewAPI(&mockSupabaseClientDynamic{GetFn: clientGet})

	req := httptest.NewRequest(http.MethodGet, "/api/admin/claims?status=PENDING&page=1&limit=25", nil)
	adminClaims := &auth.SessionClaims{
		UID:      "admin-1",
		FamilyID: "fam-charlie",
		Role:     "ADMIN",
	}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), adminClaims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", rec.Code, rec.Body.String())
	}

	if !strings.Contains(requestedProfileParams, "family_id=eq.fam-charlie") {
		t.Errorf("expected profile query scoped to fam-charlie, got %s", requestedProfileParams)
	}
	if !strings.Contains(requestedClaimParams, "user_uid=in.(usr_claim_1)") {
		t.Errorf("expected claims scoped to family member UIDs, got %s", requestedClaimParams)
	}
	if !strings.Contains(requestedClaimParams, "limit=25&offset=0") {
		t.Errorf("expected pagination limit=25 offset=0, got %s", requestedClaimParams)
	}
	if !strings.Contains(requestedCatParams, "id=in.(1)") {
		t.Errorf("expected catalog scoped to id=in.(1), got %s", requestedCatParams)
	}

	var res struct {
		Items []ClaimView `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response envelope: %v", err)
	}
	if len(res.Items) != 1 || res.Items[0].RewardTitle != "Pulsa 10k" {
		t.Errorf("expected enriched claim view, got %+v", res.Items)
	}
}

func TestShopRedeem_MultipleOf500Enforced(t *testing.T) {
	configData, _ := json.Marshal([]map[string]any{
		{"key": "redemption_start_day", "value": "1"},
		{"key": "redemption_end_day", "value": "31"},
	})
	client := &mockSupabaseClient{
		getResp: configData,
		rpcResp: []byte(`{"success":true,"claim_id":101,"new_balance":220}`),
	}
	api := NewAPI(client)

	// Non-multiple of 500 (e.g. 720) must be rejected with 400 Bad Request
	body720 := `{"coins":720,"target_type":"EWALLET","target_value":"GoPay - 0812345678"}`
	req1 := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", strings.NewReader(body720))
	req1.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "user-720", Role: "MEMBER"}
	req1 = req1.WithContext(auth.ContextWithClaims(req1.Context(), claims))

	rec1 := httptest.NewRecorder()
	api.Handler(rec1, req1)
	if rec1.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for 720 coins (non-multiple of 500), got %d: %s", rec1.Code, rec1.Body.String())
	}
	if !strings.Contains(rec1.Body.String(), "kelipatan 500") {
		t.Fatalf("expected multiple of 500 error message, got %s", rec1.Body.String())
	}

	// Multiple of 500 (e.g. 500) must pass
	body500 := `{"coins":500,"target_type":"EWALLET","target_value":"GoPay - 0812345678"}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", strings.NewReader(body500))
	req2.Header.Set("Content-Type", "application/json")
	req2 = req2.WithContext(auth.ContextWithClaims(req2.Context(), claims))

	rec2 := httptest.NewRecorder()
	api.Handler(rec2, req2)
	if rec2.Code != http.StatusOK {
		t.Fatalf("expected 200 for 500 coins, got %d: %s", rec2.Code, rec2.Body.String())
	}
}

func TestShopConfig_Authenticated_ComputesEarnedThisPeriod(t *testing.T) {
	clientGet := func(ctx context.Context, table string, params string) ([]byte, error) {
		if table == "odyssey_system_config" {
			configRows := []map[string]any{
				{"key": "default_monthly_coin_target", "value": "3320"},
				{"key": "timezone", "value": "Asia/Jakarta"},
				{"key": "default_payout_frequency", "value": "THRESHOLD"},
				{"key": "default_minimum_withdrawal_coins", "value": "500"},
			}
			return json.Marshal(configRows)
		}
		if table == "odyssey_user_payout_config" {
			return []byte("[]"), nil
		}
		if table == "odyssey_user_profiles" {
			profiles := []map[string]any{
				{"coins": 1200, "monthly_coin_target": 3320},
			}
			return json.Marshal(profiles)
		}
		if table == "odyssey_coin_transactions" {
			if !strings.Contains(params, "type=eq.TASK_REWARD") {
				t.Errorf("expected coin transaction query to filter by type=eq.TASK_REWARD, got %s", params)
			}
			txs := []map[string]any{
				{"amount": 100},
				{"amount": 250},
				{"amount": 50},
			}
			return json.Marshal(txs)
		}
		return []byte("[]"), nil
	}

	api := NewAPI(&mockSupabaseClientDynamic{GetFn: clientGet})

	req := httptest.NewRequest(http.MethodGet, "/api/shop/config", nil)
	claims := &auth.SessionClaims{UID: "user-test-period", Role: "MEMBER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))

	rec := httptest.NewRecorder()
	api.Handler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res struct {
		EarnedThisPeriod           int    `json:"earned_this_period"`
		EffectiveMonthlyTarget     int    `json:"effective_monthly_target"`
		EffectivePayoutFrequency   string `json:"effective_payout_frequency"`
		EffectiveMinimumWithdrawal int    `json:"effective_minimum_withdrawal"`
		EffectiveIsEligible        bool   `json:"effective_is_eligible"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if res.EarnedThisPeriod != 400 {
		t.Errorf("expected earned_this_period to be 400 (100+250+50), got %d", res.EarnedThisPeriod)
	}
	if res.EffectiveMonthlyTarget != 3320 {
		t.Errorf("expected effective_monthly_target to be 3320, got %d", res.EffectiveMonthlyTarget)
	}
	if res.EffectivePayoutFrequency != "THRESHOLD" {
		t.Errorf("expected effective_payout_frequency to be THRESHOLD, got %s", res.EffectivePayoutFrequency)
	}
	if res.EffectiveMinimumWithdrawal != 500 {
		t.Errorf("expected effective_minimum_withdrawal to be 500, got %d", res.EffectiveMinimumWithdrawal)
	}
	if !res.EffectiveIsEligible {
		t.Errorf("expected effective_is_eligible to be true for balance 1200 >= 500")
	}
}
