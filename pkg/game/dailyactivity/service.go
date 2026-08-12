package dailyactivity

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/progression"
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

type Service struct {
	store    ActivityStore
	activity game.ActivityStore // For streak tracking
	prog     *progression.ProgressionService
	timezone string
}

func NewService(store ActivityStore, activity game.ActivityStore, prog *progression.ProgressionService, timezone string) *Service {
	if timezone == "" {
		timezone = "Asia/Jakarta" // Default timezone
	}
	return &Service{
		store:    store,
		activity: activity,
		prog:     prog,
		timezone: timezone,
	}
}

func (s *Service) TodayDate() string {
	return time.Now().In(loadLocation(s.timezone)).Format("2006-01-02")
}

func (s *Service) getDailyActivityFromPool(ctx context.Context, date string) (*ActivityQuestion, error) {
	activities, err := s.store.ListActiveActivities(ctx)
	if err != nil {
		return nil, err
	}
	if len(activities) == 0 {
		return nil, ErrActivityNotFound
	}
	
	loc := loadLocation(s.timezone)
	parsedDate, err := time.ParseInLocation("2006-01-02", date, loc)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}
	
	dayOfYear := parsedDate.YearDay()
	index := (dayOfYear - 1) % len(activities)
	if index < 0 {
		index = 0
	}
	return &activities[index], nil
}

func (s *Service) GetToday(ctx context.Context, uid string) (*ActivityView, error) {
	date := s.TodayDate()
	act, err := s.getDailyActivityFromPool(ctx, date)
	if err != nil {
		return nil, err
	}

	completed, err := s.store.HasCompletedToday(ctx, uid, date)
	if err != nil {
		return nil, err
	}

	return &ActivityView{
		ID:        act.ID,
		Title:     act.Title,
		Type:      act.Type,
		Question:  act.Question,
		Options:   act.Options,
		Completed: completed,
		XPReward:  act.XPReward,
	}, nil
}

func (s *Service) CompleteActivity(ctx context.Context, uid string, activityID int64, answer string) (*ActivityResult, error) {
	date := s.TodayDate()
	
	completed, err := s.store.HasCompletedToday(ctx, uid, date)
	if err != nil {
		return nil, err
	}
	if completed {
		return nil, ErrAlreadyCompleted
	}
	
	act, err := s.getDailyActivityFromPool(ctx, date)
	if err != nil {
		return nil, err
	}
	
	if act.ID != activityID {
		return nil, fmt.Errorf("activity ID mismatch: expected %d, got %d", act.ID, activityID)
	}
	
	isCorrect := strings.EqualFold(strings.TrimSpace(answer), strings.TrimSpace(act.CorrectAnswer))
	if !isCorrect {
		return &ActivityResult{
			Correct:     false,
			Completed:   false,
			Explanation: act.Explanation,
		}, nil
	}
	
	err = s.store.RecordCompletion(ctx, uid, date, activityID)
	if err != nil {
		return nil, err
	}
	
	_, _, err = s.prog.AwardXP(ctx, uid, act.XPReward)
	if err != nil {
		return nil, fmt.Errorf("award xp: %w", err)
	}
	
	if s.activity != nil {
		_, _ = s.activity.RecordActivity(ctx, &game.DailyActivity{
			UserID:       uid,
			ActivityDate: date,
			ActivityType: "daily_activity",
		})
	}
	
	return &ActivityResult{
		Correct:     true,
		Completed:   true,
		Explanation: act.Explanation,
		XPAwarded:   act.XPReward,
	}, nil
}
