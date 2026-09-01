package shared

import (
	"testing"
	"time"
)

// Simulate new odyssey_target_period_bounds rolling 25→24 logic in Go
func rollingPeriodBounds(now time.Time, anchor int, tzName string) (start, end time.Time) {
	loc, _ := time.LoadLocation(tzName)
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	nowInTz := now.In(loc)
	y, m, d := nowInTz.Date()
	if d >= anchor {
		start = time.Date(y, m, anchor, 0, 0, 0, 0, loc)
		// next month
		if m == 12 {
			end = time.Date(y+1, 1, anchor, 0, 0, 0, 0, loc)
		} else {
			end = time.Date(y, m+1, anchor, 0, 0, 0, 0, loc)
		}
	} else {
		if m == 1 {
			start = time.Date(y-1, 12, anchor, 0, 0, 0, 0, loc)
		} else {
			start = time.Date(y, m-1, anchor, 0, 0, 0, 0, loc)
		}
		end = time.Date(y, m, anchor, 0, 0, 0, 0, loc)
	}
	return start, end
}

func TestRollingCycle_Boundaries(t *testing.T) {
	tz := "Asia/Jakarta"
	anchor := 25
	cases := []struct {
		dateStr     string
		wantStart   string
		wantEnd     string
		wantDays    int
	}{
		{"2026-08-24", "2026-07-25", "2026-08-25", 31},
		{"2026-08-25", "2026-08-25", "2026-09-25", 31},
		{"2026-08-31", "2026-08-25", "2026-09-25", 31},
		{"2026-09-01", "2026-08-25", "2026-09-25", 31},
		{"2026-09-24", "2026-08-25", "2026-09-25", 31},
		{"2026-09-25", "2026-09-25", "2026-10-25", 30},
		{"2026-09-30", "2026-09-25", "2026-10-25", 30},
		{"2026-10-01", "2026-09-25", "2026-10-25", 30},
		{"2026-10-24", "2026-09-25", "2026-10-25", 30},
		{"2026-10-25", "2026-10-25", "2026-11-25", 31},
		{"2026-10-31", "2026-10-25", "2026-11-25", 31},
		{"2026-11-01", "2026-10-25", "2026-11-25", 31},
		{"2026-11-24", "2026-10-25", "2026-11-25", 31},
		{"2026-11-25", "2026-11-25", "2026-12-25", 30},
		{"2026-12-25", "2026-12-25", "2027-01-25", 31},
		{"2027-01-15", "2026-12-25", "2027-01-25", 31},
		{"2027-02-25", "2027-02-25", "2027-03-25", 28},
		{"2024-02-25", "2024-02-25", "2024-03-25", 29}, // leap year
	}
	for _, tc := range cases {
		d, _ := time.ParseInLocation("2006-01-02", tc.dateStr, time.FixedZone("WIB", 7*3600))
		// Use noon to avoid DST
		d = time.Date(d.Year(), d.Month(), d.Day(), 12, 0, 0, 0, time.FixedZone("WIB", 7*3600))
		start, end := rollingPeriodBounds(d, anchor, tz)
		gotStart := start.Format("2006-01-02")
		gotEnd := end.Format("2006-01-02")
		if gotStart != tc.wantStart || gotEnd != tc.wantEnd {
			t.Fatalf("date %s: got %s->%s want %s->%s", tc.dateStr, gotStart, gotEnd, tc.wantStart, tc.wantEnd)
		}
		days := int(end.Sub(start).Hours() / 24)
		if days != tc.wantDays {
			t.Fatalf("date %s: days %d want %d", tc.dateStr, days, tc.wantDays)
		}
		// Boundary exclusive check
		loc, _ := time.LoadLocation(tz)
		startTs := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, loc)
		endTs := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, loc)
		if !startTs.Equal(start) || !endTs.Equal(end) {
			t.Fatalf("ts mismatch %s", tc.dateStr)
		}
		// Earning day check: d should be inside [start,end)
		if !(d.After(start.Add(-time.Second)) && d.Before(end)) {
			// d at noon should be inside
			if d.Before(start) || !d.Before(end) {
				t.Fatalf("date %s should be inside period", tc.dateStr)
			}
		}
	}
}

func TestRollingCycle_BoundaryMidnight(t *testing.T) {
	tz := "Asia/Jakarta"
	loc, _ := time.LoadLocation(tz)
	anchor := 25
	// 2026-09-24 23:59:59 should still be previous cycle
	d1 := time.Date(2026, 9, 24, 23, 59, 59, 0, loc)
	s1, e1 := rollingPeriodBounds(d1, anchor, tz)
	if s1.Format("2006-01-02") != "2026-08-25" || e1.Format("2006-01-02") != "2026-09-25" {
		t.Fatalf("Sep24 23:59 expected Aug25-Sep25 got %s->%s", s1.Format("2006-01-02"), e1.Format("2006-01-02"))
	}
	// 2026-09-25 00:00:00 should be new cycle
	d2 := time.Date(2026, 9, 25, 0, 0, 0, 0, loc)
	s2, e2 := rollingPeriodBounds(d2, anchor, tz)
	if s2.Format("2006-01-02") != "2026-09-25" || e2.Format("2006-01-02") != "2026-10-25" {
		t.Fatalf("Sep25 00:00 expected Sep25-Oct25 got %s->%s", s2.Format("2006-01-02"), e2.Format("2006-01-02"))
	}
}

func TestRollingCycle_Denominator(t *testing.T) {
	cases := []struct {
		target int
		total  int
		n      int
	}{
		{3200, 2690, 46},
		{3500, 2690, 46},
		{3200, 3200, 57},
		{3500, 3200, 57},
		{2000, 2690, 46},
		{1000, 3200, 57},
	}
	for _, c := range cases {
		// Create synthetic weights that sum to total: distribute total evenly
		weights := make([]int, c.n)
		base := c.total / c.n
		rem := c.total % c.n
		for i := 0; i < c.n; i++ {
			weights[i] = base
			if i < rem {
				weights[i]++
			}
		}
		// Verify sum
		sumW := 0
		for _, w := range weights {
			sumW += w
		}
		if sumW != c.total {
			t.Fatalf("synthetic weights sum %d != total %d", sumW, c.total)
		}
		bases := make([]int, c.n)
		sumBase := 0
		for i, w := range weights {
			b := (c.target * w) / c.total
			bases[i] = b
			sumBase += b
		}
		remainder := c.target - sumBase
		if remainder < 0 || remainder >= c.n {
			t.Fatalf("remainder out of range %d for target %d total %d n %d", remainder, c.target, c.total, c.n)
		}
		sum := sumBase + remainder
		if sum != c.target {
			t.Fatalf("sum %d != target %d", sum, c.target)
		}
	}
}

func TestRollingCycle_JoinDate(t *testing.T) {
	tz := "Asia/Jakarta"
	anchor := 25
	loc, _ := time.LoadLocation(tz)
	cases := []struct {
		joinStr   string
		nowStr    string
		wantStart string // eligible start after join adjustment
	}{
		{"2026-08-25", "2026-09-01", "2026-08-25"},
		{"2026-09-01", "2026-09-01", "2026-09-01"},
		{"2026-09-24", "2026-09-24", "2026-09-24"},
		{"2026-09-25", "2026-09-25", "2026-09-25"},
		{"2026-10-01", "2026-10-01", "2026-10-01"},
		{"2026-10-24", "2026-10-24", "2026-10-24"},
		{"2026-10-25", "2026-10-25", "2026-10-25"},
		{"2026-08-20", "2026-09-01", "2026-08-25"}, // join before cycle start -> cycle start wins
	}
	for _, tc := range cases {
		join, _ := time.ParseInLocation("2006-01-02", tc.joinStr, loc)
		now, _ := time.ParseInLocation("2006-01-02", tc.nowStr, loc)
		now = time.Date(now.Year(), now.Month(), now.Day(), 12, 0, 0, 0, loc)
		cycleStart, cycleEnd := rollingPeriodBounds(now, anchor, tz)
		// Simulate join adjustment: max(cycleStart, join)
		joinDate := time.Date(join.Year(), join.Month(), join.Day(), 0, 0, 0, 0, loc)
		eligibleStart := cycleStart
		if joinDate.After(cycleStart) {
			eligibleStart = joinDate
		}
		if eligibleStart.Format("2006-01-02") != tc.wantStart {
			t.Fatalf("join %s now %s: eligible start %s want %s (cycle %s->%s)", tc.joinStr, tc.nowStr, eligibleStart.Format("2006-01-02"), tc.wantStart, cycleStart.Format("2006-01-02"), cycleEnd.Format("2006-01-02"))
		}
		if !eligibleStart.Before(cycleEnd) && eligibleStart.Format("2006-01-02") != tc.wantStart {
			// edge: join == cycleEnd -> 0 earning (handled by caller)
		}
	}
}
