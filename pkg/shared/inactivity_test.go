package shared

import (
	"testing"
	"time"
)

func TestIsInactiveByCalendarDays(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	// Today = 2026-09-06 10:00 WIB
	today := time.Date(2026, 9, 6, 10, 0, 0, 0, loc)
	threshold := 5

	cases := []struct {
		name     string
		last     *time.Time
		expected bool
	}{
		{"completed today -> NOT blocked", ptrTime(time.Date(2026, 9, 6, 9, 0, 0, 0, loc)), false},
		{"completed yesterday -> NOT blocked", ptrTime(time.Date(2026, 9, 5, 23, 59, 0, 0, loc)), false},
		{"completed 4 days ago -> NOT blocked", ptrTime(time.Date(2026, 9, 2, 12, 0, 0, 0, loc)), false},
		{"completed 5 days ago -> BLOCKED", ptrTime(time.Date(2026, 9, 1, 10, 0, 0, 0, loc)), true},
		{"completed 6 days ago -> BLOCKED", ptrTime(time.Date(2026, 8, 31, 10, 0, 0, 0, loc)), true},
		{"never completed -> NOT blocked", nil, false},
		{"month boundary: Aug 31 to Sep 5 = 5 days -> BLOCKED on Sep 5", ptrTime(time.Date(2026, 8, 31, 10, 0, 0, 0, loc)), true},
	}
	// For month boundary, today Sep 5
	todaySep5 := time.Date(2026, 9, 5, 10, 0, 0, 0, loc)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			todayUse := today
			if tc.name == "month boundary: Aug 31 to Sep 5 = 5 days -> BLOCKED on Sep 5" {
				todayUse = todaySep5
				// last = Aug31, threshold 5 -> Sep5 diff =5 -> blocked
				got := IsInactiveByCalendarDays(tc.last, todayUse, threshold, loc)
				if got != tc.expected {
					t.Fatalf("expected %v got %v", tc.expected, got)
				}
				return
			}
			got := IsInactiveByCalendarDays(tc.last, todayUse, threshold, loc)
			if got != tc.expected {
				t.Fatalf("expected %v got %v for %s", tc.expected, got, tc.name)
			}
		})
	}
	// Timezone midnight boundary: completed Sep1 23:59 WIB, today Sep6 00:01 WIB => 5 days
	t.Run("timezone midnight", func(t *testing.T) {
		last := time.Date(2026, 9, 1, 23, 59, 0, 0, loc)
		todayMid := time.Date(2026, 9, 6, 0, 1, 0, 0, loc)
		if !IsInactiveByCalendarDays(&last, todayMid, 5, loc) {
			t.Fatalf("expected blocked at midnight boundary")
		}
	})
}

func TestParseAutoBlockThreshold(t *testing.T) {
	tests := []struct {
		in   string
		want int
	}{
		{"5", 5},
		{"", 5},
		{"  ", 5},
		{"invalid", 5},
		{"0", 0},
		{"-1", 0},
		{"365", 365},
		{"366", 5},
		{"1000", 5},
	}
	for _, tc := range tests {
		got := ParseAutoBlockThreshold(tc.in)
		if got != tc.want {
			t.Fatalf("ParseAutoBlockThreshold(%q)=%d want %d", tc.in, got, tc.want)
		}
	}
}

func ptrTime(t time.Time) *time.Time { return &t }

func TestAutoBlockConfigIntegration(t *testing.T) {
	cfg := ResolveRedemptionConfigFull(ResolveRedemptionConfigParams{
		Timezone:                "Asia/Jakarta",
		Now:                     time.Now(),
		AutoBlockInactivityDays: 5,
	})
	if cfg.AutoBlockInactivityDays != 5 {
		t.Fatalf("expected 5 got %d", cfg.AutoBlockInactivityDays)
	}
	cfg2 := ResolveRedemptionConfigFull(ResolveRedemptionConfigParams{
		AutoBlockInactivityDays: 0,
		Now:                     time.Now(),
	})
	if cfg2.AutoBlockInactivityDays != 0 {
		t.Fatalf("expected 0 (disabled) got %d", cfg2.AutoBlockInactivityDays)
	}
}
