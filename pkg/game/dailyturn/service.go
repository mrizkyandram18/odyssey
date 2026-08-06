package dailyturn

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/observability"
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

// DailyTurnService owns daily turn business logic: availability checks,
// streak computation, and listing. It depends only on the DailyTurnStore
// interface — never on a concrete db adapter.
type DailyTurnService struct {
	store   game.DailyTurnStore
	cfg     DailyTurnConfig
	metrics *observability.Metrics
}

// SetMetrics attaches an optional metrics sink for idempotency telemetry.
// Safe to call with nil.
func (s *DailyTurnService) SetMetrics(m *observability.Metrics) {
	s.metrics = m
}

// NewDailyTurnService constructs a DailyTurnService with the given store
// and an optional config. Pass nil to use DefaultDailyTurnConfig.
func NewDailyTurnService(store game.DailyTurnStore, cfg *DailyTurnConfig) *DailyTurnService {
	effective := DefaultDailyTurnConfig()
	if cfg != nil {
		if cfg.XP != 0 {
			effective.XP = cfg.XP
		}
		if cfg.MaxTurnsPerDay > 0 {
			effective.MaxTurnsPerDay = cfg.MaxTurnsPerDay
		}
		if cfg.Timezone != "" {
			effective.Timezone = cfg.Timezone
		}
		if cfg.Now != nil {
			effective.Now = cfg.Now
		}
	}
	return &DailyTurnService{store: store, cfg: effective}
}

// Config returns the effective daily turn configuration.
func (s *DailyTurnService) Config() DailyTurnConfig {
	return s.cfg
}

// List returns all daily turns for a user, ordered by date (newest first).
// The caller may filter the results further as needed.
func (s *DailyTurnService) List(ctx context.Context, uid string) ([]game.DailyTurn, error) {
	turns, err := s.store.ListDailyTurns(ctx, uid)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, ErrDailyTurnNotFound
		}
		return nil, err
	}
	sort.Slice(turns, func(i, j int) bool {
		return turns[i].Date > turns[j].Date
	})
	return turns, nil
}

// HasCompletedToday checks whether the user has a completed daily turn
// for the given date string (YYYY-MM-DD, service timezone).
func (s *DailyTurnService) HasCompletedToday(ctx context.Context, uid string, date string) (bool, error) {
	turns, err := s.store.ListDailyTurns(ctx, uid)
	if err != nil {
		return false, err
	}
	for _, t := range turns {
		if t.Date == date && t.Completed {
			return true, nil
		}
	}
	return false, nil
}

// IsAvailableToday reports whether the user has an uncompleted daily turn
// available for the given date string (YYYY-MM-DD), subject to the configured
// daily limit.
func (s *DailyTurnService) IsAvailableToday(ctx context.Context, uid string, date string) (bool, error) {
	turns, err := s.store.ListDailyTurns(ctx, uid)
	if err != nil {
		return false, err
	}
	completedToday := 0
	for _, t := range turns {
		if t.Date == date && t.Completed {
			completedToday++
		}
	}
	return completedToday < s.cfg.MaxTurnsPerDay, nil
}

// ComputeStreak calculates the number of consecutive days (ending at the
// most recent day with a completed turn) that the user has completed
// their daily turn.
//
// If today is completed, the streak starts from today. If today is not
// completed, the streak starts from yesterday. The streak breaks when a
// day has no completed turn.
func (s *DailyTurnService) ComputeStreak(ctx context.Context, uid string) (int, error) {
	turns, err := s.store.ListDailyTurns(ctx, uid)
	if err != nil {
		return 0, err
	}

	completed := make(map[string]bool, len(turns))
	for _, t := range turns {
		completed[t.Date] = t.Completed
	}

	loc := loadLocation(s.cfg.Timezone)
	now := s.cfg.Now().In(loc)
	dateStr := now.Format("2006-01-02")
	checkDate := now

	if !completed[dateStr] {
		checkDate = now.AddDate(0, 0, -1)
	}

	streak := 0
	for {
		d := checkDate.Format("2006-01-02")
		if !completed[d] {
			break
		}
		streak++
		checkDate = checkDate.AddDate(0, 0, -1)
		if streak > 365 {
			break
		}
	}

	return streak, nil
}

// ConsumeDailyTurn creates a completed daily turn for the user on the given date.
// It rejects the call if the user has already consumed MaxTurnsPerDay today or
// if the same quest slug was already completed today (idempotency guard).
// The caller is responsible for awarding XP as a side-effect.
func (s *DailyTurnService) ConsumeDailyTurn(ctx context.Context, uid string, date string, questSlug string) (*game.DailyTurn, error) {
	completedCount := 0
	alreadyDone := false
	turns, err := s.store.ListDailyTurns(ctx, uid)
	if err != nil && !errors.Is(err, game.ErrNotFound) {
		return nil, err
	}
	for _, t := range turns {
		if t.Date == date && t.Completed {
			completedCount++
			if t.QuestSlug == questSlug {
				alreadyDone = true
			}
		}
	}
	if alreadyDone {
		if s.metrics != nil {
			s.metrics.RecordReplayIgnored()
		}
		return nil, ErrNoTurnsRemaining
	}
	if completedCount >= s.cfg.MaxTurnsPerDay {
		return nil, ErrNoTurnsRemaining
	}

	dt := &game.DailyTurn{
		UID:       uid,
		Date:      date,
		QuestSlug: questSlug,
		Completed: true,
	}
	created, err := s.store.CreateDailyTurn(ctx, dt)
	if err != nil {
		return nil, fmt.Errorf("create daily turn: %w", err)
	}
	return created, nil
}

// UpdateDailyTurn updates a daily turn record.
func (s *DailyTurnService) UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error {
	return s.store.UpdateDailyTurn(ctx, turnID, patch)
}

// TodayDate returns the current date in the configured timezone formatted as YYYY-MM-DD.
func (s *DailyTurnService) TodayDate() string {
	return s.cfg.Now().In(loadLocation(s.cfg.Timezone)).Format("2006-01-02")
}

// TodayDateUTC is a backwards-compatible alias returning the UTC date.
// Deprecated: use DailyTurnService.TodayDate() which respects the configured timezone.
func TodayDate() string {
	return time.Now().UTC().Format("2006-01-02")
}

// FormatDate is a convenience helper for tests that need a date string
// offset from a base time.
func FormatDate(t time.Time) string {
	return t.UTC().Format("2006-01-02")
}

// ParseDate parses a YYYY-MM-DD date string.
func ParseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", strings.TrimSpace(s))
}
