package admin_members

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/auth"
	"odyssey/pkg/shared"
)

func TestApplyCapBonusCeiling(t *testing.T) {
	cfg := `{"10":300,"15":500,"20":1000}`
	cases := []struct {
		name        string
		base, level int
		bonus       string
		ceiling     int
		want        int
	}{
		{"unlimited stays unlimited", 0, 20, cfg, 10000, 0},
		{"unconfigured stays unconfigured", -1, 20, cfg, 10000, -1},
		{"no bonus below threshold", 3000, 8, cfg, 10000, 3000},
		{"level bonus applied", 3000, 20, cfg, 10000, 4000},
		{"highest threshold wins", 3000, 15, cfg, 10000, 3500},
		{"ceiling clamps", 9500, 20, cfg, 10000, 10000},
		{"malformed bonus ignored", 3000, 20, "invalid-json", 10000, 3000},
		{"no ceiling configured", 3000, 20, cfg, -1, 4000},
	}
	for _, c := range cases {
		if got := applyCapBonusCeiling(c.base, c.level, c.bonus, c.ceiling); got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, got, c.want)
		}
	}
}

func TestHandleListMembers_FullMonthCycleAndEffectiveCap(t *testing.T) {
	cap := 3000
	mock := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			switch table {
			case "odyssey_user_profiles":
				switch {
				case strings.Contains(params, "select=uid,is_active"):
					return json.Marshal([]map[string]any{{"uid": "u1", "is_active": true}})
				case strings.Contains(params, "select=uid,password_hash"):
					return json.Marshal([]map[string]any{{"uid": "u1", "password_hash": "h"}})
				default:
					return json.Marshal([]map[string]any{{
						"uid": "u1", "username": "selvi", "family_id": "fam_1",
						"explorer_name": "Selvi", "role": "MEMBER", "is_active": true,
						"level": 20, "xp": 1000, "coins": 341,
						"monthly_coin_target": 3320, "monthly_earning_cap": cap,
						"created_at": "2026-09-01T00:00:00Z",
					}})
				}
			case "odyssey_system_config":
				switch {
				case strings.Contains(params, "key=eq.timezone"):
					return json.Marshal([]map[string]string{{"value": "Asia/Jakarta"}})
				case strings.Contains(params, "key=eq.level_cap_bonus"):
					return json.Marshal([]map[string]string{{"value": `{"10":300,"15":500,"20":1000}`}})
				case strings.Contains(params, "key=eq.max_monthly_earning_cap_ceiling"):
					return json.Marshal([]map[string]string{{"value": "10000"}})
				case strings.Contains(params, "default_monthly_coin_target"):
					return json.Marshal([]map[string]string{{"value": "0"}})
				case strings.Contains(params, "default_monthly_earning_cap"):
					return json.Marshal([]map[string]string{{"value": "3320"}})
				}
				return []byte("[]"), nil
			case "odyssey_coin_transactions", "odyssey_task_submissions", "odyssey_user_payout_config":
				return []byte("[]"), nil
			}
			return []byte("[]"), nil
		},
	}
	api := NewAPI(mock)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/members", nil)
	req = req.WithContext(auth.ContextWithClaims(req.Context(), &auth.SessionClaims{
		UID: "admin_1", Role: "ADMIN", FamilyID: "fam_1",
	}))
	rec := httptest.NewRecorder()
	api.HandleListMembers(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var resp shared.PaginatedResponse[MemberView]
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 member, got %d", len(resp.Items))
	}
	m := resp.Items[0]

	// Full calendar month bounds in Asia/Jakarta.
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, loc).Day()
	wantStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).Format("2006-01-02")
	wantEnd := time.Date(now.Year(), now.Month(), lastDay, 0, 0, 0, 0, loc).Format("2006-01-02")
	if m.CurrentCycleStart != wantStart || m.CurrentCycleEnd != wantEnd {
		t.Errorf("cycle = %s..%s, want %s..%s", m.CurrentCycleStart, m.CurrentCycleEnd, wantStart, wantEnd)
	}
	// Effective cap = base 3000 + level-20 bonus 1000, ceiling 10000.
	if m.EffectiveEarningCap == nil || *m.EffectiveEarningCap != 4000 {
		got := -9999
		if m.EffectiveEarningCap != nil {
			got = *m.EffectiveEarningCap
		}
		t.Errorf("effective_earning_cap = %d, want 4000", got)
	}
	// Raw stored cap preserved.
	if m.MonthlyEarningCap == nil || *m.MonthlyEarningCap != 3000 {
		t.Errorf("monthly_earning_cap not preserved")
	}
	if m.EarningLocked {
		t.Errorf("member should not be locked with 0 earned vs cap 4000")
	}
}
