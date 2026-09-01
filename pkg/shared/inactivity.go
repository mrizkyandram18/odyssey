package shared

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// IsInactiveByCalendarDays returns true if (today - lastCompletionDate) >= threshold.
// All dates are in the configured timezone calendar days.
// lastCompletion may be nil (never completed) -> not inactive (do not auto-block new users).
// alreadyBlocked -> not eligible (caller should check).
// threshold <=0 => disabled (never inactive).
func IsInactiveByCalendarDays(lastCompletion *time.Time, today time.Time, threshold int, loc *time.Location) bool {
	if threshold <= 0 {
		return false
	}
	if lastCompletion == nil {
		return false
	}
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	todayDate := today.In(loc).Format("2006-01-02")
	lastDate := lastCompletion.In(loc).Format("2006-01-02")
	// Parse dates to compute delta in calendar days
	t1, _ := time.ParseInLocation("2006-01-02", lastDate, loc)
	t2, _ := time.ParseInLocation("2006-01-02", todayDate, loc)
	days := int(t2.Sub(t1) / (24 * time.Hour))
	return days >= threshold
}

// ParseAutoBlockThreshold parses auto_block_inactivity_days value with safe defaults.
// Returns 5 on missing/invalid, 0 if explicitly disabled (0 or negative), capped at 365.
func ParseAutoBlockThreshold(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return DefaultAutoBlockInactivityDays
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return DefaultAutoBlockInactivityDays
	}
	if n <= 0 {
		return 0 // disabled
	}
	if n > 365 {
		return DefaultAutoBlockInactivityDays
	}
	return n
}

// FormatInactivityBoundary documents the exact semantics for tests/docs.
// For threshold=5, last completion Sep 1 => blocked on Sep 6 (5 calendar days later).
func FormatInactivityBoundary(threshold int) string {
	return fmt.Sprintf("inactive when (today_date - last_success_date) >= %d calendar days in Asia/Jakarta; never-completed => not blocked; already blocked => skipped", threshold)
}

// IsInactiveCycleAware is cycle-aware variant: lastCompletion must be within [periodStart, periodEnd)
// Otherwise (outside cycle or nil) => NOT inactive (counter resets per cycle).
func IsInactiveCycleAware(lastCompletion *time.Time, today time.Time, threshold int, loc *time.Location, periodStart, periodEnd time.Time) bool {
	if threshold <= 0 {
		return false
	}
	if lastCompletion == nil {
		return false
	}
	if loc == nil {
		loc = time.FixedZone("WIB", 7*3600)
	}
	lastDateStr := lastCompletion.In(loc).Format("2006-01-02")
	lastDate, _ := time.ParseInLocation("2006-01-02", lastDateStr, loc)
	psStr := periodStart.In(loc).Format("2006-01-02")
	peStr := periodEnd.In(loc).Format("2006-01-02")
	ps, _ := time.ParseInLocation("2006-01-02", psStr, loc)
	pe, _ := time.ParseInLocation("2006-01-02", peStr, loc)
	if lastDate.Before(ps) || !lastDate.Before(pe) {
		return false
	}
	todayStr := today.In(loc).Format("2006-01-02")
	todayDate, _ := time.ParseInLocation("2006-01-02", todayStr, loc)
	days := int(todayDate.Sub(lastDate) / (24 * time.Hour))
	return days >= threshold
}
