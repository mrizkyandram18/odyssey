package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestProgressionStore_CountRelics(t *testing.T) {
	data, _ := json.Marshal([]Relic{
		{ID: 1, UID: "user-1", Code: "relic-1"},
		{ID: 2, UID: "user-1", Code: "relic-2"},
		{ID: 3, UID: "user-1", Code: "relic-3"},
	})
	store := NewProgressionStore(&mockSupabaseClient{data: data})
	count, err := store.CountRelics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Errorf("expected 3 relics, got %d", count)
	}
}

func TestProgressionStore_CountRelics_Empty(t *testing.T) {
	data, _ := json.Marshal([]Relic{})
	store := NewProgressionStore(&mockSupabaseClient{data: data})
	count, err := store.CountRelics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Errorf("expected 0 relics, got %d", count)
	}
}

func TestProgressionStore_CountRelics_Error(t *testing.T) {
	store := NewProgressionStore(&mockSupabaseClient{err: errors.New("network")})
	_, err := store.CountRelics(context.Background(), "user-1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestProgressionStore_CreateRelic(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Relic{
		{ID: 1, UID: "user-1", Code: "relic-1", AwardedAt: now, CreatedAt: now},
	})
	store := NewProgressionStore(&mockSupabaseClient{data: data})
	r := &game.Relic{
		UID:  "user-1",
		Code: "relic-1",
	}
	result, err := store.CreateRelic(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestProgressionStore_CreateChest(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Chest{
		{ID: 1, UID: "user-1", Source: "daily", Opened: false, CreatedAt: now},
	})
	store := NewProgressionStore(&mockSupabaseClient{data: data})
	ch := &game.Chest{
		UID:    "user-1",
		Source: "daily",
		Opened: false,
	}
	result, err := store.CreateChest(context.Background(), ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestProgressionStore_UpdateChest(t *testing.T) {
	store := NewProgressionStore(&mockSupabaseClient{data: []byte("[]")})
	err := store.UpdateChest(context.Background(), 1, map[string]any{"opened": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProgressionStore_CreateAchievement(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Achievement{
		{ID: 1, UID: "user-1", CrewID: "crew-1", Code: "first-step", Kind: "PERSONAL", AwardedAt: now, CreatedAt: now},
	})
	store := NewProgressionStore(&mockSupabaseClient{data: data})
	a := &game.Achievement{
		UID:    "user-1",
		CrewID: "crew-1",
		Code:   "first-step",
		Kind:   "PERSONAL",
	}
	result, err := store.CreateAchievement(context.Background(), a)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestProgressionStore_ImplementsInterface(t *testing.T) {
	var _ game.ProgressionStore = (*supabaseProgressionStore)(nil)
}
