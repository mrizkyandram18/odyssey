package db

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestQuestStore_GetQuest_Success(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	ts := now
	data, _ := json.Marshal([]QuestInstance{
		{
			ID:           1,
			FamilyID:       "crew-1",
			TemplateSlug: "morning-light",
			Title:        "Test Mission",
			Status:       "PENDING",
			CreatedAt:    now,
			StartedAt:    &ts,
		},
	})
	store := NewQuestStore(&mockSupabaseClient{data: data})
	q, err := store.GetQuest(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if q.ID != 1 {
		t.Errorf("expected ID 1, got %d", q.ID)
	}
	if q.TemplateSlug != "morning-light" {
		t.Errorf("expected slug morning-light, got %s", q.TemplateSlug)
	}
	if q.Status != "PENDING" {
		t.Errorf("expected status PENDING, got %s", q.Status)
	}
}

func TestQuestStore_GetQuest_NotFound(t *testing.T) {
	data, _ := json.Marshal([]QuestInstance{})
	store := NewQuestStore(&mockSupabaseClient{data: data})
	_, err := store.GetQuest(context.Background(), 999)
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestQuestStore_ListByCrew(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]QuestInstance{
		{ID: 1, FamilyID: "crew-1", Title: "Mission A", Status: "DONE", CreatedAt: now},
		{ID: 2, FamilyID: "crew-1", Title: "Mission B", Status: "ACTIVE", StartedAt: &now, CreatedAt: now},
	})
	store := NewQuestStore(&mockSupabaseClient{data: data})
	missions, err := store.ListQuestByCrew(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(missions) != 2 {
		t.Fatalf("expected 2 missions, got %d", len(missions))
	}
	if missions[0].Title != "Mission A" {
		t.Errorf("expected first quest 'Mission A', got %s", missions[0].Title)
	}
}

func TestQuestStore_GetChallenges(t *testing.T) {
	data, _ := json.Marshal([]Exercise{
		{ID: 10, MissionID: 1, Slug: "obs", Description: "Observe", Status: "DONE"},
		{ID: 11, MissionID: 1, Slug: "research", Description: "Research", Status: "PENDING"},
	})
	store := NewQuestStore(&mockSupabaseClient{data: data})
	exercises, err := store.GetChallenges(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exercises) != 2 {
		t.Fatalf("expected 2 exercises, got %d", len(exercises))
	}
	if exercises[0].Status != "DONE" {
		t.Errorf("expected first challenge DONE, got %s", exercises[0].Status)
	}
}

func TestQuestStore_ListChallengesByQuestIDs(t *testing.T) {
	data, _ := json.Marshal([]Exercise{
		{ID: 10, MissionID: 1, Slug: "obs", Description: "Observe", Status: "DONE"},
		{ID: 11, MissionID: 1, Slug: "research", Description: "Research", Status: "PENDING"},
		{ID: 12, MissionID: 3, Slug: "draw", Description: "Draw", Status: "DONE"},
	})
	client := &mockSupabaseClient{data: data}
	store := NewQuestStore(client).(*supabaseQuestStore)
	exercises, err := store.ListChallengesByQuestIDs(context.Background(), []int64{1, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exercises) != 3 {
		t.Fatalf("expected 3 exercises, got %d", len(exercises))
	}
	if len(client.getCalls) != 1 {
		t.Fatalf("expected 1 batched request, got %d", len(client.getCalls))
	}
	unescaped, err := url.QueryUnescape(client.getCalls[0])
	if err != nil {
		t.Fatalf("parse recorded params: %v", err)
	}
	if !strings.Contains(unescaped, "mission_id=in.(1,3)") {
		t.Errorf("expected in.(1,3) filter, got %q", client.getCalls[0])
	}
	if !strings.HasPrefix(client.getCalls[0], "odyssey_exercises?") {
		t.Errorf("expected odyssey_exercises table, got %q", client.getCalls[0])
	}
}

func TestQuestStore_ListChallengesByQuestIDs_Empty(t *testing.T) {
	client := &mockSupabaseClient{data: []byte("[]")}
	store := NewQuestStore(client).(*supabaseQuestStore)
	exercises, err := store.ListChallengesByQuestIDs(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(exercises) != 0 {
		t.Fatalf("expected 0 exercises, got %d", len(exercises))
	}
	if len(client.getCalls) != 0 {
		t.Fatalf("expected no request for empty ids, got %d", len(client.getCalls))
	}
}

func TestQuestStore_UpdateQuest(t *testing.T) {
	store := NewQuestStore(&mockSupabaseClient{data: []byte("[]")})
	err := store.UpdateQuest(context.Background(), 1, map[string]any{"status": "ACTIVE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestQuestStore_ImplementsInterface(t *testing.T) {
	var _ game.QuestStore = (*supabaseQuestStore)(nil)
}
