package admin

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"odyssey/pkg/db"
)

type AdminStats struct {
	TotalUsers                    int `json:"total_users"`
	ActiveUsers7d                 int `json:"active_users_7d"`
	ActiveUsers30d                int `json:"active_users_30d"`
	QuestCompletions              int `json:"quest_completions"`
	DailyActivityCompletionsToday int `json:"daily_activity_completions_today"`
}

type QuestStat struct {
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	Published       bool   `json:"published"`
	CompletionCount int    `json:"completion_count"`
}

type ActivityStat struct {
	ID              int64  `json:"id"`
	Slug            string `json:"slug"`
	Title           string `json:"title"`
	Active          bool   `json:"active"`
	CompletionCount int    `json:"completion_count"`
}

type AdminService struct {
	client db.SupabaseClient
}

func NewAdminService(c db.SupabaseClient) *AdminService {
	return &AdminService{client: c}
}

func (s *AdminService) GetStats(ctx context.Context) (*AdminStats, error) {
	stats := &AdminStats{}

	// Users
	type User struct {
		UpdatedAt string `json:"updated_at"`
	}
	var users []User
	b, err := s.client.Get(ctx, "odyssey_user_profiles", "select=updated_at")
	if err == nil && len(b) > 0 {
		_ = json.Unmarshal(b, &users)
		stats.TotalUsers = len(users)
		now := time.Now()
		for _, u := range users {
			t, _ := time.Parse(time.RFC3339, u.UpdatedAt)
			if now.Sub(t).Hours() <= 7*24 {
				stats.ActiveUsers7d++
			}
			if now.Sub(t).Hours() <= 30*24 {
				stats.ActiveUsers30d++
			}
		}
	}

	// Mission completions
	type Mission struct {
		Status string `json:"status"`
	}
	var missions []Mission
	qb, err := s.client.Get(ctx, "odyssey_missions", "select=status&status=eq.DONE")
	if err == nil && len(qb) > 0 {
		_ = json.Unmarshal(qb, &missions)
		stats.QuestCompletions = len(missions)
	}

	// Daily activity completions today
	today := time.Now().Format("2006-01-02")
	type Completion struct {
		ID int64 `json:"id"`
	}
	var completions []Completion
	cb, err := s.client.Get(ctx, "odyssey_daily_activity_completions", fmt.Sprintf("select=id&activity_date=eq.%s", today))
	if err == nil && len(cb) > 0 {
		_ = json.Unmarshal(cb, &completions)
		stats.DailyActivityCompletionsToday = len(completions)
	}

	return stats, nil
}

func (s *AdminService) GetQuests(ctx context.Context) ([]QuestStat, error) {
	// Need definitions and missions status
	type QDef struct {
		Slug      string `json:"slug"`
		Title     string `json:"title"`
		Published bool   `json:"published"`
	}
	var defs []QDef
	b, err := s.client.Get(ctx, "odyssey_quest_definitions", "select=slug,title,published")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &defs); err != nil {
		return nil, err
	}

	type QRun struct {
		MissionSlug string `json:"mission_slug"`
		Status    string `json:"status"`
	}
	var runs []QRun
	rb, _ := s.client.Get(ctx, "odyssey_missions", "select=mission_slug,status&status=eq.DONE")
	if len(rb) > 0 {
		_ = json.Unmarshal(rb, &runs)
	}

	counts := make(map[string]int)
	for _, r := range runs {
		counts[r.MissionSlug]++
	}

	var res []QuestStat
	for _, d := range defs {
		res = append(res, QuestStat{
			Slug:            d.Slug,
			Title:           d.Title,
			Published:       d.Published,
			CompletionCount: counts[d.Slug],
		})
	}
	return res, nil
}

func (s *AdminService) GetDailyActivities(ctx context.Context) ([]ActivityStat, error) {
	type ADef struct {
		ID     int64  `json:"id"`
		Slug   string `json:"slug"`
		Title  string `json:"title"`
		Active bool   `json:"active"`
	}
	var defs []ADef
	b, err := s.client.Get(ctx, "odyssey_daily_activities", "select=id,slug,title,active")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &defs); err != nil {
		return nil, err
	}

	type CRun struct {
		ActivityID int64 `json:"activity_id"`
	}
	var runs []CRun
	rb, _ := s.client.Get(ctx, "odyssey_daily_activity_completions", "select=activity_id")
	if len(rb) > 0 {
		_ = json.Unmarshal(rb, &runs)
	}

	counts := make(map[int64]int)
	for _, r := range runs {
		counts[r.ActivityID]++
	}

	var res []ActivityStat
	for _, d := range defs {
		res = append(res, ActivityStat{
			ID:              d.ID,
			Slug:            d.Slug,
			Title:           d.Title,
			Active:          d.Active,
			CompletionCount: counts[d.ID],
		})
	}
	return res, nil
}

func (s *AdminService) ToggleQuestPublished(ctx context.Context, slug string) error {
	type QDef struct {
		Published bool `json:"published"`
	}
	var defs []QDef
	b, err := s.client.Get(ctx, "odyssey_quest_definitions", fmt.Sprintf("select=published&slug=eq.%s", slug))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &defs); err != nil || len(defs) == 0 {
		return fmt.Errorf("quest not found")
	}

	patch := map[string]interface{}{
		"published": !defs[0].Published,
	}
	_, err = s.client.Mutate(ctx, "PATCH", "odyssey_quest_definitions", patch, fmt.Sprintf("slug=eq.%s", slug))
	return err
}

func (s *AdminService) ToggleActivityActive(ctx context.Context, id int64) error {
	type ADef struct {
		Active bool `json:"active"`
	}
	var defs []ADef
	b, err := s.client.Get(ctx, "odyssey_daily_activities", fmt.Sprintf("select=active&id=eq.%d", id))
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &defs); err != nil || len(defs) == 0 {
		return fmt.Errorf("activity not found")
	}

	patch := map[string]interface{}{
		"active": !defs[0].Active,
	}
	_, err = s.client.Mutate(ctx, "PATCH", "odyssey_daily_activities", patch, fmt.Sprintf("id=eq.%d", id))
	return err
}
