package admin_members

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestEarnedThisPeriod_CalendarMonthBoundary(t *testing.T) {
	// Verify that earned_this_period uses calendar month 1 → last day, not 1-24
	// We capture the SQL params sent to odyssey_coin_transactions and check period bounds
	var capturedParams string
	mock := &mockSupabaseClient{
		getFunc: func(ctx context.Context, table string, params string) ([]byte, error) {
			if table == "odyssey_system_config" {
				// Return minimal config: timezone Asia/Jakarta, no target_earning_start/end (should be ignored for cap)
				if strings.Contains(params, "timezone") {
					return json.Marshal([]map[string]string{{"value": "Asia/Jakarta"}})
				}
				return []byte("[]"), nil
			}
			if table == "odyssey_coin_transactions" {
				capturedParams = params
				return json.Marshal([]map[string]any{
					{"user_uid": "u1", "amount": 100},
				})
			}
			if table == "odyssey_user_profiles" {
				return json.Marshal([]map[string]any{
					{"uid": "u1", "family_id": "fam1", "explorer_name": "A", "role": "MEMBER", "is_active": true},
				})
			}
			return []byte("[]"), nil
		},
	}
	// Use the public helper getEarnedThisPeriod (now calendar month) directly
	res := getEarnedThisPeriod(context.Background(), mock, []string{"u1"})
	if res["u1"] != 100 {
		t.Fatalf("expected 100, got %v", res)
	}
	if capturedParams == "" {
		t.Fatal("no coin_transactions query captured")
	}
	// Parse period bounds from params: should contain gte. and lt. with calendar month
	// Compute expected bounds using same logic as implementation: 1 → lastDay
	loc, _ := time.LoadLocation("Asia/Jakarta")
	now := time.Now().In(loc)
	lastDay := time.Date(now.Year(), now.Month()+1, 0, 0, 0, 0, 0, loc).Day()
	expStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	expEnd := time.Date(now.Year(), now.Month(), lastDay+1, 0, 0, 0, 0, loc).UTC().Format(time.RFC3339)
	if !strings.Contains(capturedParams, expStart) {
		t.Fatalf("expected periodStart %s in params %s", expStart, capturedParams)
	}
	if !strings.Contains(capturedParams, expEnd) {
		t.Fatalf("expected periodEnd %s in params %s", expEnd, capturedParams)
	}
	// Ensure old 1-24 window is NOT used when lastDay !=24 (most months)
	if lastDay != 24 {
		oldEnd := time.Date(now.Year(), now.Month(), 24+1, 0, 0, 0, 0, loc).UTC().Format(time.RFC3339)
		if strings.Contains(capturedParams, oldEnd) && strings.Contains(capturedParams, "target_earning") {
			t.Fatalf("should not use old 1-24 target_earning window")
		}
	}
}

func TestEarningCap_FebruaryLastDay(t *testing.T) {
	// Verify last day calculation for Feb non-leap
	loc, _ := time.LoadLocation("Asia/Jakarta")
	feb := time.Date(2026, 2, 15, 12, 0, 0, 0, loc) // 2026 not leap
	lastDay := time.Date(feb.Year(), feb.Month()+1, 0, 0, 0, 0, 0, loc).Day()
	if lastDay != 28 {
		t.Fatalf("expected Feb 2026 lastDay 28, got %d", lastDay)
	}
	febLeap := time.Date(2024, 2, 15, 12, 0, 0, 0, loc)
	if ld := time.Date(febLeap.Year(), febLeap.Month()+1, 0, 0, 0, 0, 0, loc).Day(); ld != 29 {
		t.Fatalf("expected Feb 2024 lastDay 29, got %d", ld)
	}
}
