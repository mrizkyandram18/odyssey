package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestRelicStore_CreateRelic(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Relic{
		{ID: 1, UID: "user-1", Code: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", AwardedAt: now, CreatedAt: now},
	})
	store := NewRelicStore(&mockSupabaseClient{data: data})
	r := &game.Relic{
		UID:  "user-1",
		Code: "ancient-compass",
		Name: "Ancient Compass",
	}
	result, err := store.CreateRelic(context.Background(), r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestRelicStore_GetRelic(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Relic{
		{ID: 1, UID: "user-1", Code: "ancient-compass", AwardedAt: now, CreatedAt: now},
	})
	store := NewRelicStore(&mockSupabaseClient{data: data})
	r, err := store.GetRelic(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.ID != 1 {
		t.Errorf("expected ID 1, got %d", r.ID)
	}
}

func TestRelicStore_GetRelic_NotFound(t *testing.T) {
	store := NewRelicStore(&mockSupabaseClient{data: []byte("[]")})
	_, err := store.GetRelic(context.Background(), 99)
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRelicStore_ListRelics(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Relic{
		{ID: 1, UID: "user-1", Code: "r1", AwardedAt: now, CreatedAt: now},
		{ID: 2, UID: "user-1", Code: "r2", AwardedAt: now, CreatedAt: now},
	})
	store := NewRelicStore(&mockSupabaseClient{data: data})
	relics, err := store.ListRelics(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(relics) != 2 {
		t.Errorf("expected 2 relics, got %d", len(relics))
	}
}

func TestRelicStore_CountRelics(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Relic{
		{ID: 1, UID: "user-1", AwardedAt: now, CreatedAt: now},
		{ID: 2, UID: "user-1", AwardedAt: now, CreatedAt: now},
	})
	store := NewRelicStore(&mockSupabaseClient{data: data})
	count, err := store.CountRelics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 relics, got %d", count)
	}
}

func TestRelicStore_ImplementsInterface(t *testing.T) {
	var _ game.RelicStore = (*supabaseRelicStore)(nil)
}
