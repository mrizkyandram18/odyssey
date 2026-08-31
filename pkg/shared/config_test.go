package shared

import (
	"testing"
	"time"
)

func TestResolveRedemptionConfig_Boundaries(t *testing.T) {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	tests := []struct {
		name       string
		startDay   int
		endDay     int
		date       time.Time
		expectOpen bool
		expectDay  int
	}{
		{
			name:       "Day 20 -> Closed (before default window 24-26)",
			startDay:   24,
			endDay:     26,
			date:       time.Date(2026, time.August, 20, 12, 0, 0, 0, loc),
			expectOpen: false,
			expectDay:  20,
		},
		{
			name:       "Day 24 -> Open (start of default window 24-26)",
			startDay:   24,
			endDay:     26,
			date:       time.Date(2026, time.August, 24, 0, 1, 0, 0, loc),
			expectOpen: true,
			expectDay:  24,
		},
		{
			name:       "Day 25 -> Open (inside default window 24-26)",
			startDay:   24,
			endDay:     26,
			date:       time.Date(2026, time.August, 25, 15, 30, 0, 0, loc),
			expectOpen: true,
			expectDay:  25,
		},
		{
			name:       "Day 26 -> Open (end of default window 24-26)",
			startDay:   24,
			endDay:     26,
			date:       time.Date(2026, time.August, 26, 23, 59, 59, 0, loc),
			expectOpen: true,
			expectDay:  26,
		},
		{
			name:       "Day 27 -> Closed (after default window 24-26)",
			startDay:   24,
			endDay:     26,
			date:       time.Date(2026, time.August, 27, 0, 0, 1, 0, loc),
			expectOpen: false,
			expectDay:  27,
		},
		{
			name:       "Custom Range 10-15: Day 12 -> Open",
			startDay:   10,
			endDay:     15,
			date:       time.Date(2026, time.September, 12, 10, 0, 0, 0, loc),
			expectOpen: true,
			expectDay:  12,
		},
		{
			name:       "Custom Range 10-15: Day 16 -> Closed",
			startDay:   10,
			endDay:     15,
			date:       time.Date(2026, time.September, 16, 10, 0, 0, 0, loc),
			expectOpen: false,
			expectDay:  16,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := ResolveRedemptionConfig(tc.startDay, tc.endDay, "Asia/Jakarta", tc.date)
			if cfg.IsOpen != tc.expectOpen {
				t.Fatalf("expected IsOpen=%v, got %v for day %d", tc.expectOpen, cfg.IsOpen, tc.expectDay)
			}
			if cfg.CurrentDay != tc.expectDay {
				t.Fatalf("expected CurrentDay=%d, got %d", tc.expectDay, cfg.CurrentDay)
			}
			if cfg.ConversionRate != 100 {
				t.Fatalf("expected ConversionRate=100, got %d", cfg.ConversionRate)
			}
			if cfg.MaxPayoutCoins != 3200 {
				t.Fatalf("expected MaxPayoutCoins=3200, got %d", cfg.MaxPayoutCoins)
			}
			if cfg.Timezone != "Asia/Jakarta" {
				t.Fatalf("expected Timezone=Asia/Jakarta, got %s", cfg.Timezone)
			}
		})
	}
}

func TestConfigurability_ScenariosWithoutCodeChange(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	sampleDate := time.Date(2026, time.September, 25, 10, 0, 0, 0, loc)

	// Scenario 1: Standard Rate 100, Target 320,000 -> Target Coins 3200
	cfg1 := ResolveRedemptionConfigFull(ResolveRedemptionConfigParams{
		ConversionRate:     100,
		PayoutTargetRupiah: 320000,
		MaxPayoutCoins:     3200,
		PayoutDay:          24,
		EarningPeriodDays:  30,
		Timezone:           "Asia/Jakarta",
		Now:                sampleDate,
	})
	if cfg1.PayoutTargetCoins != 3200 {
		t.Fatalf("Scenario 1 failed: expected 3200 coins, got %d", cfg1.PayoutTargetCoins)
	}

	// Scenario 2: Different Rate 50, Target 320,000 -> Target Coins 6400
	cfg2 := ResolveRedemptionConfigFull(ResolveRedemptionConfigParams{
		ConversionRate:     50,
		PayoutTargetRupiah: 320000,
		MaxPayoutCoins:     7000,
		PayoutDay:          25,
		EarningPeriodDays:  28,
		Timezone:           "Asia/Jakarta",
		Now:                sampleDate,
	})
	if cfg2.PayoutTargetCoins != 6400 {
		t.Fatalf("Scenario 2 failed: expected 6400 coins, got %d", cfg2.PayoutTargetCoins)
	}
	if !cfg2.IsPayoutDay {
		t.Fatalf("Scenario 2 failed: expected IsPayoutDay=true for day 25")
	}
	if cfg2.EarningPeriodDays != 28 {
		t.Fatalf("Scenario 2 failed: expected EarningPeriodDays=28, got %d", cfg2.EarningPeriodDays)
	}

	// Scenario 3: Validation failure when target coins > max payout coins
	cfg3 := ResolveRedemptionConfigFull(ResolveRedemptionConfigParams{
		ConversionRate:     100,
		PayoutTargetRupiah: 500000, // 5000 coins
		MaxPayoutCoins:     3200,   // max 3200
		PayoutDay:          24,
		EarningPeriodDays:  30,
		Timezone:           "Asia/Jakarta",
		Now:                sampleDate,
	})
	if err := ValidateEconomyConfig(cfg3); err == nil {
		t.Fatalf("Scenario 3 failed: expected validation error when target coins (5000) > max payout (3200)")
	}
}
