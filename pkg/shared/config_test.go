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
			name:       "Day 20 -> Closed (before default window 21-26)",
			startDay:   21,
			endDay:     26,
			date:       time.Date(2026, time.August, 20, 12, 0, 0, 0, loc),
			expectOpen: false,
			expectDay:  20,
		},
		{
			name:       "Day 21 -> Open (start of default window 21-26)",
			startDay:   21,
			endDay:     26,
			date:       time.Date(2026, time.August, 21, 0, 1, 0, 0, loc),
			expectOpen: true,
			expectDay:  21,
		},
		{
			name:       "Day 22 -> Open (inside default window 21-26)",
			startDay:   21,
			endDay:     26,
			date:       time.Date(2026, time.August, 22, 15, 30, 0, 0, loc),
			expectOpen: true,
			expectDay:  22,
		},
		{
			name:       "Day 25 -> Open (inside default window 21-26)",
			startDay:   21,
			endDay:     26,
			date:       time.Date(2026, time.August, 25, 23, 59, 0, 0, loc),
			expectOpen: true,
			expectDay:  25,
		},
		{
			name:       "Day 26 -> Open (end of default window 21-26)",
			startDay:   21,
			endDay:     26,
			date:       time.Date(2026, time.August, 26, 23, 59, 59, 0, loc),
			expectOpen: true,
			expectDay:  26,
		},
		{
			name:       "Day 27 -> Closed (after default window 21-26)",
			startDay:   21,
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
			if cfg.ConversionRate != 10 {
				t.Fatalf("expected ConversionRate=10, got %d", cfg.ConversionRate)
			}
			if cfg.Timezone != "Asia/Jakarta" {
				t.Fatalf("expected Timezone=Asia/Jakarta, got %s", cfg.Timezone)
			}
		})
	}
}
