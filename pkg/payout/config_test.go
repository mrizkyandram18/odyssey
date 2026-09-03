package payout

import (
	"context"
	"encoding/json"
	"testing"
	"time"
)

type mockClient struct {
	getFn func(table, params string) ([]byte, error)
}

func (m *mockClient) Get(_ context.Context, table string, params string) ([]byte, error) {
	if m.getFn != nil {
		return m.getFn(table, params)
	}
	return []byte("[]"), nil
}
func (m *mockClient) Mutate(_ context.Context, _, _ string, _ any, _ string) ([]byte, error) {
	return []byte("[]"), nil
}
func (m *mockClient) MutateAtomic(_ context.Context, _, _ string, _ any, _ string, _ string) ([]byte, error) {
	return []byte("[]"), nil
}
func (m *mockClient) RPC(_ context.Context, _ string, _ any) ([]byte, error) { return []byte("{}"), nil }
func (m *mockClient) UploadStorage(_ context.Context, _, _, _ string, _ []byte) (string, error) {
	return "", nil
}

func TestGetEffectivePayoutConfig_UserOverrides(t *testing.T) {
	sysCfg := []map[string]any{
		{"key": "default_payout_frequency", "value": "MONTHLY"},
		{"key": "default_minimum_withdrawal_coins", "value": "500"},
		{"key": "default_payout_weekday", "value": "1"},
		{"key": "redemption_start_day", "value": "24"},
		{"key": "redemption_end_day", "value": "26"},
	}
	sysData, _ := json.Marshal(sysCfg)
	userCfg := []map[string]any{
		{"user_uid": "u1", "payout_frequency": "THRESHOLD", "minimum_withdrawal_coins": 5000, "payout_weekday": 3, "payout_month_start_day": 10, "payout_month_end_day": 15, "enabled": true},
	}
	userData, _ := json.Marshal(userCfg)
	client := &mockClient{getFn: func(table, params string) ([]byte, error) {
		if table == "odyssey_system_config" {
			return sysData, nil
		}
		if table == "odyssey_user_payout_config" {
			return userData, nil
		}
		return []byte("[]"), nil
	}}
	cfg, err := GetEffectivePayoutConfig(context.Background(), client, "u1", time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Frequency != FrequencyThreshold {
		t.Fatalf("expected THRESHOLD, got %s", cfg.Frequency)
	}
	if cfg.MinimumWithdrawalCoins != 5000 {
		t.Fatalf("expected 5000, got %d", cfg.MinimumWithdrawalCoins)
	}
	if cfg.Source != "user" {
		t.Fatalf("expected user source")
	}
}

func TestIsEligible_Threshold(t *testing.T) {
	cfg := EffectivePayoutConfig{Frequency: FrequencyThreshold, MinimumWithdrawalCoins: 500}
	if ok, _ := IsEligible(cfg, 499, time.Now(), "Asia/Jakarta"); ok {
		t.Fatalf("499 should not be eligible")
	}
	if ok, _ := IsEligible(cfg, 500, time.Now(), "Asia/Jakarta"); !ok {
		t.Fatalf("500 should be eligible")
	}
	if ok, _ := IsEligible(cfg, 1000, time.Now(), "Asia/Jakarta"); !ok {
		t.Fatalf("1000 should be eligible")
	}
}

func TestIsEligible_Threshold_Various(t *testing.T) {
	for _, tc := range []struct {
		threshold int
		balance   int
		expect    bool
	}{
		{500, 499, false},
		{500, 500, true},
		{500, 501, true},
		{1000, 999, false},
		{1000, 1000, true},
		{5000, 4999, false},
		{5000, 5000, true},
		{5000, 9999, true}, // balance >= min is eligible; payout semantics single transaction
		{5000, 10000, true},
	} {
		cfg := EffectivePayoutConfig{Frequency: FrequencyThreshold, MinimumWithdrawalCoins: tc.threshold}
		ok, _ := IsEligible(cfg, tc.balance, time.Now(), "Asia/Jakarta")
		if ok != tc.expect {
			t.Fatalf("threshold %d balance %d expect %v got %v", tc.threshold, tc.balance, tc.expect, ok)
		}
	}
}

func TestIsEligible_Weekly(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	monday := time.Date(2026, 9, 7, 10, 0, 0, 0, loc) // Monday
	tuesday := time.Date(2026, 9, 8, 10, 0, 0, 0, loc)
	cfg := EffectivePayoutConfig{Frequency: FrequencyWeekly, MinimumWithdrawalCoins: 500, PayoutWeekday: 1}
	if ok, _ := IsEligible(cfg, 500, monday, "Asia/Jakarta"); !ok {
		t.Fatalf("Monday should be eligible")
	}
	if ok, _ := IsEligible(cfg, 500, tuesday, "Asia/Jakarta"); ok {
		t.Fatalf("Tuesday should not be eligible for Monday config")
	}
	if ok, _ := IsEligible(cfg, 499, monday, "Asia/Jakarta"); ok {
		t.Fatalf("499 below threshold should not be eligible even on correct day")
	}
}

func TestIsEligible_Monthly(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	cfg := EffectivePayoutConfig{Frequency: FrequencyMonthly, MinimumWithdrawalCoins: 5000, PayoutMonthStartDay: 21, PayoutMonthEndDay: 26}
	d20 := time.Date(2026, 9, 20, 10, 0, 0, 0, loc)
	d21 := time.Date(2026, 9, 21, 10, 0, 0, 0, loc)
	d26 := time.Date(2026, 9, 26, 10, 0, 0, 0, loc)
	d27 := time.Date(2026, 9, 27, 10, 0, 0, 0, loc)
	if ok, _ := IsEligible(cfg, 5000, d20, "Asia/Jakarta"); ok {
		t.Fatalf("20 should not be eligible")
	}
	if ok, _ := IsEligible(cfg, 5000, d21, "Asia/Jakarta"); !ok {
		t.Fatalf("21 should be eligible")
	}
	if ok, _ := IsEligible(cfg, 5000, d26, "Asia/Jakarta"); !ok {
		t.Fatalf("26 should be eligible")
	}
	if ok, _ := IsEligible(cfg, 5000, d27, "Asia/Jakarta"); ok {
		t.Fatalf("27 should not be eligible")
	}
	if ok, _ := IsEligible(cfg, 4999, d21, "Asia/Jakarta"); ok {
		t.Fatalf("4999 below threshold")
	}
}

func TestUserIsolation(t *testing.T) {
	sysCfg := []map[string]any{
		{"key": "default_payout_frequency", "value": "THRESHOLD"},
		{"key": "default_minimum_withdrawal_coins", "value": "500"},
	}
	sysData, _ := json.Marshal(sysCfg)
	clientA := &mockClient{getFn: func(table, params string) ([]byte, error) {
		if table == "odyssey_system_config" {
			return sysData, nil
		}
		if table == "odyssey_user_payout_config" {
			data, _ := json.Marshal([]map[string]any{{"user_uid": "userA", "payout_frequency": "THRESHOLD", "minimum_withdrawal_coins": 500}})
			return data, nil
		}
		return []byte("[]"), nil
	}}
	clientB := &mockClient{getFn: func(table, params string) ([]byte, error) {
		if table == "odyssey_system_config" {
			return sysData, nil
		}
		if table == "odyssey_user_payout_config" {
			data, _ := json.Marshal([]map[string]any{{"user_uid": "userB", "payout_frequency": "MONTHLY", "minimum_withdrawal_coins": 5000, "payout_month_start_day": 21, "payout_month_end_day": 26}})
			return data, nil
		}
		return []byte("[]"), nil
	}}
	cfgA, _ := GetEffectivePayoutConfig(context.Background(), clientA, "userA", time.Now())
	cfgB, _ := GetEffectivePayoutConfig(context.Background(), clientB, "userB", time.Now())
	if cfgA.Frequency == cfgB.Frequency || cfgA.MinimumWithdrawalCoins == cfgB.MinimumWithdrawalCoins {
		t.Fatalf("configs should be isolated: A %+v B %+v", cfgA, cfgB)
	}
}
