package admin

import (
	"context"
	"testing"

	"odyssey/pkg/db"
)

type mockSupabase struct {
	db.SupabaseClient
}

func (m *mockSupabase) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if table == "odyssey_user_profiles" {
		return []byte(`[{"updated_at":"2030-01-01T00:00:00Z"}]`), nil
	}
	if table == "odyssey_missions" {
		return []byte(`[{"mission_slug":"test-quest","status":"DONE"}]`), nil
	}
	if table == "odyssey_daily_activity_completions" {
		return []byte(`[{"id":1,"activity_id":100}]`), nil
	}
	if table == "odyssey_quest_definitions" {
		return []byte(`[{"slug":"test-quest","title":"Test Mission","published":true}]`), nil
	}
	if table == "odyssey_daily_activities" {
		return []byte(`[{"id":100,"slug":"test-activity","title":"Test Activity","active":true}]`), nil
	}
	return nil, nil
}

func (m *mockSupabase) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	return []byte(`{}`), nil
}

func TestGetStats(t *testing.T) {
	svc := NewAdminService(&mockSupabase{})
	stats, err := svc.GetStats(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stats.TotalUsers != 1 {
		t.Errorf("expected 1 total user, got %d", stats.TotalUsers)
	}
	if stats.QuestCompletions != 1 {
		t.Errorf("expected 1 quest completion, got %d", stats.QuestCompletions)
	}
	// ActiveUsers7d will be 0 since 2030 is far in the future/past depending on time.Now()
}

func TestGetQuests(t *testing.T) {
	svc := NewAdminService(&mockSupabase{})
	missions, err := svc.GetQuests(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(missions) != 1 {
		t.Fatalf("expected 1 quest, got %d", len(missions))
	}
	if missions[0].Slug != "test-quest" || missions[0].CompletionCount != 1 {
		t.Errorf("unexpected quest data: %+v", missions[0])
	}
}

func TestGetDailyActivities(t *testing.T) {
	svc := NewAdminService(&mockSupabase{})
	acts, err := svc.GetDailyActivities(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(acts) != 1 {
		t.Fatalf("expected 1 activity, got %d", len(acts))
	}
	if acts[0].ID != 100 || acts[0].CompletionCount != 1 {
		t.Errorf("unexpected activity data: %+v", acts[0])
	}
}

func TestToggles(t *testing.T) {
	svc := NewAdminService(&mockSupabase{})
	ctx := context.Background()
	if err := svc.ToggleQuestPublished(ctx, "test-quest"); err != nil {
		t.Fatal(err)
	}
	if err := svc.ToggleActivityActive(ctx, 100); err != nil {
		t.Fatal(err)
	}
}
