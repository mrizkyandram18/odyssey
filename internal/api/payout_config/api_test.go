package payout_config

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

type dynamicClient struct {
	getFn    func(table, params string) []byte
	mutateOK bool
}

func (m *dynamicClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.getFn != nil {
		return m.getFn(table, params), nil
	}
	return []byte("[]"), nil
}

func (m *dynamicClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *dynamicClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *dynamicClient) RPC(ctx context.Context, fnName string, payload any) ([]byte, error) {
	return []byte("{}"), nil
}

func (m *dynamicClient) UploadStorage(ctx context.Context, bucket, path, contentType string, data []byte) (string, error) {
	return "", nil
}

func upsertClient() *dynamicClient {
	sysRows, _ := json.Marshal([]map[string]any{
		{"key": "default_minimum_withdrawal_coins", "value": "500"},
		{"key": "redemption_start_day", "value": "21"},
		{"key": "redemption_end_day", "value": "26"},
	})
	profRows, _ := json.Marshal([]map[string]any{
		{"family_id": "fam-1"},
	})
	return &dynamicClient{
		getFn: func(table, params string) []byte {
			switch table {
			case "odyssey_system_config":
				return sysRows
			case "odyssey_user_profiles":
				return profRows
			default:
				return []byte("[]")
			}
		},
	}
}

func upsertReq(t *testing.T, api *API, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/admin/payout-config/user", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "admin-1", FamilyID: "fam-1", Role: "ADMIN"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	return rec
}

func TestHandleUpsert_ValidationPreserved(t *testing.T) {
	api := NewAPI(upsertClient())
	cases := []struct {
		name       string
		body       string
		wantCode   int
		wantSubstr string
	}{
		{
			name:       "invalid frequency rejected",
			body:       `{"user_uid":"u1","payout_frequency":"DAILY","minimum_withdrawal_coins":500}`,
			wantCode:   http.StatusBadRequest,
			wantSubstr: "payout_frequency tidak valid (THRESHOLD|WEEKLY|MONTHLY)",
		},
		{
			name:       "zero minimum rejected",
			body:       `{"user_uid":"u1","payout_frequency":"THRESHOLD","minimum_withdrawal_coins":0}`,
			wantCode:   http.StatusBadRequest,
			wantSubstr: "minimum_withdrawal_coins harus > 0",
		},
		{
			name:       "below system floor rejected",
			body:       `{"user_uid":"u1","payout_frequency":"THRESHOLD","minimum_withdrawal_coins":100}`,
			wantCode:   http.StatusBadRequest,
			wantSubstr: "di bawah system minimum (500)",
		},
		{
			name:       "above hard cap rejected",
			body:       `{"user_uid":"u1","payout_frequency":"THRESHOLD","minimum_withdrawal_coins":200000}`,
			wantCode:   http.StatusBadRequest,
			wantSubstr: "terlalu besar (max 100000)",
		},
		{
			name:       "weekly bad weekday rejected",
			body:       `{"user_uid":"u1","payout_frequency":"WEEKLY","minimum_withdrawal_coins":500,"payout_weekday":9}`,
			wantCode:   http.StatusBadRequest,
			wantSubstr: "payout_weekday harus 0..6 (0=Sunday)",
		},
		{
			name:       "monthly inverted window rejected",
			body:       `{"user_uid":"u1","payout_frequency":"MONTHLY","minimum_withdrawal_coins":500,"payout_month_start_day":27,"payout_month_end_day":26}`,
			wantCode:   http.StatusBadRequest,
			wantSubstr: "tidak boleh lebih besar dari end_day",
		},
		{
			name:     "threshold without schedule succeeds",
			body:     `{"user_uid":"u1","payout_frequency":"THRESHOLD","minimum_withdrawal_coins":500}`,
			wantCode: http.StatusOK,
		},
		{
			name:     "monthly with window succeeds",
			body:     `{"user_uid":"u1","payout_frequency":"MONTHLY","minimum_withdrawal_coins":500,"payout_month_start_day":21,"payout_month_end_day":26}`,
			wantCode: http.StatusOK,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := upsertReq(t, api, tc.body)
			if rec.Code != tc.wantCode {
				t.Fatalf("expected %d, got %d: %s", tc.wantCode, rec.Code, rec.Body.String())
			}
			if tc.wantSubstr != "" {
				// Compare against the decoded error string: Go escapes
				// '>' as \u003e in JSON responses.
				var decoded struct {
					Error string `json:"error"`
				}
				_ = json.Unmarshal(rec.Body.Bytes(), &decoded)
				if !strings.Contains(decoded.Error, tc.wantSubstr) {
					t.Errorf("expected error to contain %q, got %q", tc.wantSubstr, decoded.Error)
				}
			}
		})
	}
}
