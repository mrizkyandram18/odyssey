package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestPlayerRelicStore_CreatePlayerRelic(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]PlayerRelic{
		{ID: 1, UID: "user-1", RelicSlug: "ancient-compass", RelicID: 1, OwnedCount: 1, IsNew: true, DiscoveredAt: now, CreatedAt: now, UpdatedAt: now},
	})
	store := NewPlayerRelicStore(&mockSupabaseClient{data: data})
	pr := &game.PlayerRelic{
		UID:       "user-1",
		RelicSlug: "ancient-compass",
		RelicID:   1,
	}
	result, err := store.CreatePlayerRelic(context.Background(), pr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", result.UID)
	}
	if result.OwnedCount != 1 {
		t.Errorf("expected owned_count 1, got %d", result.OwnedCount)
	}
}

func TestPlayerRelicStore_GetPlayerRelic(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]PlayerRelic{
		{ID: 1, UID: "user-1", RelicSlug: "ancient-compass", RelicID: 1, OwnedCount: 2, CreatedAt: now, UpdatedAt: now},
	})
	store := NewPlayerRelicStore(&mockSupabaseClient{data: data})
	pr, err := store.GetPlayerRelic(context.Background(), "user-1", "ancient-compass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pr.OwnedCount != 2 {
		t.Errorf("expected owned_count 2, got %d", pr.OwnedCount)
	}
}

func TestPlayerRelicStore_GetPlayerRelic_NotFound(t *testing.T) {
	store := NewPlayerRelicStore(&mockSupabaseClient{data: []byte("[]")})
	_, err := store.GetPlayerRelic(context.Background(), "user-1", "missing")
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestPlayerRelicStore_UpdatePlayerRelic(t *testing.T) {
	store := NewPlayerRelicStore(&mockSupabaseClient{data: []byte("[]")})
	err := store.UpdatePlayerRelic(context.Background(), "user-1", "ancient-compass", map[string]any{"owned_count": 5})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPlayerRelicStore_ListPlayerRelics(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]PlayerRelic{
		{ID: 1, UID: "user-1", RelicSlug: "r1", OwnedCount: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 2, UID: "user-1", RelicSlug: "r2", OwnedCount: 3, CreatedAt: now, UpdatedAt: now},
	})
	store := NewPlayerRelicStore(&mockSupabaseClient{data: data})
	relics, err := store.ListPlayerRelics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(relics) != 2 {
		t.Errorf("expected 2 player relics, got %d", len(relics))
	}
}

func TestPlayerRelicStore_CountUniqueRelics(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]PlayerRelic{
		{ID: 1, UID: "user-1", RelicSlug: "r1", OwnedCount: 1, CreatedAt: now, UpdatedAt: now},
		{ID: 2, UID: "user-1", RelicSlug: "r2", OwnedCount: 3, CreatedAt: now, UpdatedAt: now},
		{ID: 3, UID: "user-1", RelicSlug: "r1", OwnedCount: 2, CreatedAt: now, UpdatedAt: now},
	})
	store := NewPlayerRelicStore(&mockSupabaseClient{data: data})
	count, err := store.CountUniqueRelics(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 unique relics, got %d", count)
	}
}

func TestPlayerRelicStore_ImplementsInterface(t *testing.T) {
	var _ game.PlayerRelicStore = (*supabasePlayerRelicStore)(nil)
}
