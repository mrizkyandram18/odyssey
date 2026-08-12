package familystreak

import (
	"context"
	"sync"
	"time"

	"odyssey/pkg/game"
)

var tzCache sync.Map

func loadLocation(tz string) *time.Location {
	if cached, ok := tzCache.Load(tz); ok {
		return cached.(*time.Location)
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		loc = time.UTC
	}
	tzCache.Store(tz, loc)
	return loc
}

// Service computes the crew-level daily streak from existing daily activity
// data (odyssey_daily_activity, written on daily turn consumption). A crew
// streak day is any calendar day on which at least one current crew member
// recorded qualifying activity. Membership is resolved at read time, so joins
// and leaves are reflected without any historical recomputation.
type Service struct {
	users    game.UserStore
	activity game.ActivityStore
	tz       string
	now      func() time.Time
}

// NewService constructs a crew streak Service. tz is the IANA timezone used
// for calendar-day boundaries (the same timezone that daily activity dates are
// written in).
func NewService(users game.UserStore, activity game.ActivityStore, tz string) *Service {
	return &Service{users: users, activity: activity, tz: tz, now: time.Now}
}

// ComputeCrewStreak returns the number of consecutive calendar days ending at
// today (or yesterday if today has no activity) on which the crew recorded at
// least one qualifying activity.
func (s *Service) ComputeCrewStreak(ctx context.Context, crewID string) (int, error) {
	members, err := s.users.ListUsersByCrew(ctx, crewID)
	if err != nil {
		return 0, err
	}
	if len(members) == 0 {
		return 0, nil
	}

	uids := make([]string, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UID)
	}

	acts, err := s.activity.ListActivityDatesByUsers(ctx, uids)
	if err != nil {
		return 0, err
	}

	dates := make(map[string]struct{}, len(acts))
	for _, a := range acts {
		dates[a.ActivityDate] = struct{}{}
	}

	return CountConsecutiveDays(dates, s.now().In(loadLocation(s.tz))), nil
}

// CountConsecutiveDays counts the number of consecutive calendar days present
// in dates, ending at today. If today has no activity yet, the count starts at
// yesterday. This mirrors the personal-streak semantics of the activity store.
func CountConsecutiveDays(dates map[string]struct{}, now time.Time) int {
	if len(dates) == 0 {
		return 0
	}

	today := now.Format("2006-01-02")
	yesterday := now.AddDate(0, 0, -1).Format("2006-01-02")

	if _, ok := dates[today]; !ok {
		if _, ok := dates[yesterday]; !ok {
			return 0
		}
		now = now.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		d := now.Format("2006-01-02")
		if _, ok := dates[d]; !ok {
			break
		}
		streak++
		now = now.AddDate(0, 0, -1)
		if streak > 365 {
			break
		}
	}
	return streak
}
