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

func TestCycleAwareInactivity(t *testing.T) {
	loc, _ := time.LoadLocation("Asia/Jakarta")
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	threshold := 5
	// Cycle 2026-09-01 to 2026-09-25 (period_end exclusive Sep25)
	periodStart := time.Date(2026, 9, 1, 0, 0, 0, 0, loc)
	periodEnd := time.Date(2026, 9, 25, 0, 0, 0, 0, loc)
	// Case A: completion in previous cycle Sep20, today Sep25 (new cycle start) -> NOT blocked (last outside cycle)
	t.Run("Case A completion prev cycle Sep20 at Sep25 => NOT blocked", func(t *testing.T) {
		last := time.Date(2026, 9, 20, 10, 0, 0, 0, loc)
		today := time.Date(2026, 9, 25, 10, 0, 0, 0, loc)
		// periodStart is Sep01 for today Sep25 still, but spec expects new cycle Sep25-Oct25
		// For this test we simulate new cycle start Sep25 as periodStart
		ps := time.Date(2026, 9, 25, 0, 0, 0, 0, loc)
		pe := time.Date(2026, 10, 25, 0, 0, 0, 0, loc)
		if IsInactiveCycleAware(&last, today, threshold, loc, ps, pe) {
			t.Fatalf("expected NOT blocked for previous cycle completion")
		}
	})
	// Case B: no completion in current cycle, joins Sep25 -> NOT blocked
	t.Run("Case B never completes in current cycle -> NOT blocked", func(t *testing.T) {
		today := time.Date(2026, 9, 30, 10, 0, 0, 0, loc)
		if IsInactiveCycleAware(nil, today, threshold, loc, periodStart, periodEnd) {
			t.Fatalf("expected NOT blocked for nil last")
		}
	})
	// Case C: completed Sep25, at Sep30 diff 5 => BLOCKED
	t.Run("Case C completed Sep25 at Sep30 => BLOCKED", func(t *testing.T) {
		last := time.Date(2026, 9, 25, 10, 0, 0, 0, loc)
		today := time.Date(2026, 9, 30, 10, 0, 0, 0, loc)
		ps := time.Date(2026, 9, 25, 0, 0, 0, 0, loc)
		pe := time.Date(2026, 10, 25, 0, 0, 0, 0, loc)
		if !IsInactiveCycleAware(&last, today, threshold, loc, ps, pe) {
			t.Fatalf("expected BLOCKED")
		}
	})
	// Case D: completed yesterday Sep29 at Sep30 => NOT blocked
	t.Run("Case D completed yesterday => NOT blocked", func(t *testing.T) {
		last := time.Date(2026, 9, 29, 10, 0, 0, 0, loc)
		today := time.Date(2026, 9, 30, 10, 0, 0, 0, loc)
		ps := time.Date(2026, 9, 25, 0, 0, 0, 0, loc)
		pe := time.Date(2026, 10, 25, 0, 0, 0, 0, loc)
		if IsInactiveCycleAware(&last, today, threshold, loc, ps, pe) {
			t.Fatalf("expected NOT blocked for yesterday")
		}
	})
	// Case E: exactly threshold Sep25 at Sep30 => BLOCKED
	t.Run("Case E exactly threshold Sep25 at Sep30 => BLOCKED", func(t *testing.T) {
		last := time.Date(2026, 9, 25, 10, 0, 0, 0, loc)
		today := time.Date(2026, 9, 30, 10, 0, 0, 0, loc)
		ps := time.Date(2026, 9, 25, 0, 0, 0, 0, loc)
		pe := time.Date(2026, 10, 25, 0, 0, 0, 0, loc)
		if !IsInactiveCycleAware(&last, today, threshold, loc, ps, pe) {
			t.Fatalf("expected BLOCKED at threshold")
		}
	})
	// Join date cases
	t.Run("Join Oct23 never completes => NOT blocked", func(t *testing.T) {
		today := time.Date(2026, 10, 28, 10, 0, 0, 0, loc)
		ps := time.Date(2026, 10, 1, 0, 0, 0, 0, loc)
		pe := time.Date(2026, 11, 1, 0, 0, 0, 0, loc)
		if IsInactiveCycleAware(nil, today, threshold, loc, ps, pe) {
			t.Fatalf("new joiner never completed should not be blocked")
		}
	})
	t.Run("last outside cycle Aug31 with Sep period => NOT blocked", func(t *testing.T) {
		last := time.Date(2026, 8, 31, 10, 0, 0, 0, loc)
		today := time.Date(2026, 9, 10, 10, 0, 0, 0, loc)
		if IsInactiveCycleAware(&last, today, threshold, loc, periodStart, periodEnd) {
			t.Fatalf("last outside cycle should not trigger block")
		}
	})
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
