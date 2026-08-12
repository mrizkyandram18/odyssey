package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestChestStore_CreateChest(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Gift{
		{ID: 1, UID: "user-1", Source: "quest", GiftSlug: "wooden-chest", Rarity: "COMMON", Icon: "📦", Description: "A chest", Opened: false, CreatedAt: now},
	})
	store := NewChestStore(&mockSupabaseClient{data: data})
	ch := &game.Gift{
		UID:       "user-1",
		Source:    "quest",
		GiftSlug: "wooden-chest",
		Rarity:    "COMMON",
		Icon:      "📦",
	}
	result, err := store.CreateChest(context.Background(), ch)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != 1 {
		t.Errorf("expected ID 1, got %d", result.ID)
	}
}

func TestChestStore_GetChest(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Gift{
		{ID: 1, UID: "user-1", Source: "quest", GiftSlug: "wooden-chest", Opened: false, CreatedAt: now},
	})
	store := NewChestStore(&mockSupabaseClient{data: data})
	ch, err := store.GetChest(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ch.ID != 1 {
		t.Errorf("expected ID 1, got %d", ch.ID)
	}
}

func TestChestStore_GetChest_NotFound(t *testing.T) {
	store := NewChestStore(&mockSupabaseClient{data: []byte("[]")})
	_, err := store.GetChest(context.Background(), 99)
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestChestStore_UpdateChest(t *testing.T) {
	store := NewChestStore(&mockSupabaseClient{data: []byte("[]")})
	err := store.UpdateChest(context.Background(), 1, map[string]any{"opened": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestChestStore_ListChestsByUser(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]Gift{
		{ID: 1, UID: "user-1", Source: "quest", Opened: false, CreatedAt: now},
	})
	store := NewChestStore(&mockSupabaseClient{data: data})
	gifts, err := store.ListChestsByUser(context.Background(), "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gifts) != 1 {
		t.Errorf("expected 1 chest, got %d", len(gifts))
	}
}

func TestChestStore_ImplementsInterface(t *testing.T) {
	var _ game.ChestStore = (*supabaseChestStore)(nil)
}
