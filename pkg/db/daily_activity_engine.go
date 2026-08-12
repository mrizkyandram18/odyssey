package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game/dailyactivity"
)

type dailyActivityEngineStore struct {
	client SupabaseClient
}

func NewDailyActivityEngineStore(client SupabaseClient) dailyactivity.ActivityStore {
	return &dailyActivityEngineStore{client: client}
}

func (s *dailyActivityEngineStore) ListActiveActivities(ctx context.Context) ([]dailyactivity.ActivityQuestion, error) {
	v := url.Values{}
	v.Set("active", "eq.true")
	v.Set("order", "id.asc")
	
	raw, err := s.client.Get(ctx, "odyssey_daily_activities", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list active daily activities: %w", err)
	}

	var acts []dailyactivity.ActivityQuestion
	if err := json.Unmarshal(raw, &acts); err != nil {
		return nil, fmt.Errorf("parse daily activities: %w", err)
	}
	return acts, nil
}

func (s *dailyActivityEngineStore) HasCompletedToday(ctx context.Context, uid string, date string) (bool, error) {
	v := url.Values{}
	v.Set("user_id", "eq."+uid)
	v.Set("activity_date", "eq."+date)
	
	raw, err := s.client.Get(ctx, "odyssey_daily_activity_completions", v.Encode())
	if err != nil {
		return false, fmt.Errorf("check daily activity completion: %w", err)
	}

	var rows []map[string]interface{}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return false, fmt.Errorf("parse daily activity completions: %w", err)
	}
	
	return len(rows) > 0, nil
}

func (s *dailyActivityEngineStore) RecordCompletion(ctx context.Context, uid string, date string, activityID int64) error {
	payload := map[string]interface{}{
		"user_id":       uid,
		"activity_date": date,
		"activity_id":   activityID,
	}
	
	_, err := s.client.Mutate(ctx, "POST", "odyssey_daily_activity_completions", payload, "")
	if err != nil {
		return fmt.Errorf("record daily activity completion: %w", err)
	}
	return nil
}
