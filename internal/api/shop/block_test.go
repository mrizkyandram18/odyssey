package shop

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"odyssey/pkg/auth"
)

func TestShopRedeem_BlockedUserRejected(t *testing.T) {
	client := &mockSupabaseClientDynamic{
		GetFn: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_user_profiles" {
				return []byte(`[{"is_active":false}]`), nil
			}
			return []byte(`[]`), nil
		},
	}
	api := NewAPI(client)
	body := `{"coins":100,"target_type":"EWALLET","target_value":"GoPay - 08123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/shop/redeem", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	claims := &auth.SessionClaims{UID: "blocked_user", Role: "MEMBER"}
	req = req.WithContext(auth.ContextWithClaims(req.Context(), claims))
	rec := httptest.NewRecorder()
	api.Handler(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for blocked user redeem got %d %s", rec.Code, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "diblokir") {
		t.Fatalf("expected block message got %s", rec.Body.String())
	}
}
