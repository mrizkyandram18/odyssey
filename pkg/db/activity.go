package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"odyssey/pkg/game"
)

type activityStore struct {
	client SupabaseClient
}

func NewActivityStore(client SupabaseClient) game.ActivityStore {
	return &activityStore{client: client}
}

func (s *activityStore) RecordActivity(ctx context.Context, act *game.DailyActivity) (*game.DailyActivity, error) {
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_daily_activity", act, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("record activity: %w", err)
	}

	var resp []game.DailyActivity
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse activity response: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no activity returned after insert")
	}

	return &resp[0], nil
}

func (s *activityStore) GetStreak(ctx context.Context, uid string) (int, error) {
	v := url.Values{}
	v.Set("user_id", "eq."+uid)
	v.Set("select", "activity_date")
	v.Set("order", "activity_date.desc")
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_daily_activity", params)
	if err != nil {
		return 0, fmt.Errorf("get activity dates: %w", err)
	}

	var rows []struct {
		ActivityDate string `json:"activity_date"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return 0, fmt.Errorf("parse activity dates: %w", err)
	}

	if len(rows) == 0 {
		return 0, nil
	}

	uniqueDates := make([]string, 0, len(rows))
	seen := make(map[string]bool)
	for _, r := range rows {
		if !seen[r.ActivityDate] {
			seen[r.ActivityDate] = true
			uniqueDates = append(uniqueDates, r.ActivityDate)
		}
	}

	streak := 0
	today := time.Now().UTC().Format("2006-01-02")
	yesterday := time.Now().UTC().AddDate(0, 0, -1).Format("2006-01-02")

	expectedDate := today
	if len(uniqueDates) > 0 && uniqueDates[0] != today && uniqueDates[0] != yesterday {
		return 0, nil
	}
	if len(uniqueDates) > 0 && uniqueDates[0] == yesterday {
		expectedDate = yesterday
	}

	currentDate, _ := time.Parse("2006-01-02", expectedDate)

	for _, d := range uniqueDates {
		if d == expectedDate {
			streak++
			currentDate = currentDate.AddDate(0, 0, -1)
			expectedDate = currentDate.Format("2006-01-02")
		} else {
			break
		}
	}

	return streak, nil
}
