package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestChestDefinitionStore_ListChestDefinitions(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]ChestDefinition{
		{ID: 1, Slug: "wooden-chest", Name: "Wooden Chest", Rarity: "COMMON", Icon: "📦", Description: "A chest", CreatedAt: now, UpdatedAt: now},
	})
	store := NewChestDefinitionStore(&mockSupabaseClient{data: data})
	defs, err := store.ListChestDefinitions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("expected 1 definition, got %d", len(defs))
	}
}

func TestChestDefinitionStore_GetChestDefinition(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]ChestDefinition{
		{ID: 1, Slug: "wooden-chest", Name: "Wooden Chest", Rarity: "COMMON", Icon: "📦", Description: "A chest", CreatedAt: now, UpdatedAt: now},
	})
	store := NewChestDefinitionStore(&mockSupabaseClient{data: data})
	def, err := store.GetChestDefinition(context.Background(), "wooden-chest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Slug != "wooden-chest" {
		t.Errorf("expected wooden-chest, got %s", def.Slug)
	}
}

func TestChestDefinitionStore_GetChestDefinition_NotFound(t *testing.T) {
	store := NewChestDefinitionStore(&mockSupabaseClient{data: []byte("[]")})
	_, err := store.GetChestDefinition(context.Background(), "missing")
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestChestDefinitionStore_ListDropTableEntries(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]DropTableEntry{
		{ID: 1, ChestSlug: "wooden-chest", Rarity: "COMMON", Weight: 0.7, CreatedAt: now},
		{ID: 2, ChestSlug: "wooden-chest", Rarity: "UNCOMMON", Weight: 0.3, CreatedAt: now},
	})
	store := NewChestDefinitionStore(&mockSupabaseClient{data: data})
	entries, err := store.ListDropTableEntries(context.Background(), "wooden-chest")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestChestDefinitionStore_ImplementsInterface(t *testing.T) {
	var _ game.ChestDefinitionStore = (*supabaseChestDefinitionStore)(nil)
}
