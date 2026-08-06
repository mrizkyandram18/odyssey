package db

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
)

func TestRelicDefinitionStore_ListRelicDefinitions(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]RelicDefinition{
		{ID: 1, Slug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", Image: "🧭", CreatedAt: now, UpdatedAt: now},
	})
	store := NewRelicDefinitionStore(&mockSupabaseClient{data: data})
	defs, err := store.ListRelicDefinitions(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(defs) != 1 {
		t.Errorf("expected 1 definition, got %d", len(defs))
	}
}

func TestRelicDefinitionStore_GetRelicDefinition(t *testing.T) {
	now := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	data, _ := json.Marshal([]RelicDefinition{
		{ID: 1, Slug: "ancient-compass", Name: "Ancient Compass", Rarity: "COMMON", Image: "🧭", CreatedAt: now, UpdatedAt: now},
	})
	store := NewRelicDefinitionStore(&mockSupabaseClient{data: data})
	def, err := store.GetRelicDefinition(context.Background(), "ancient-compass")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if def.Slug != "ancient-compass" {
		t.Errorf("expected ancient-compass, got %s", def.Slug)
	}
}

func TestRelicDefinitionStore_GetRelicDefinition_NotFound(t *testing.T) {
	store := NewRelicDefinitionStore(&mockSupabaseClient{data: []byte("[]")})
	_, err := store.GetRelicDefinition(context.Background(), "missing")
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestRelicDefinitionStore_ImplementsInterface(t *testing.T) {
	var _ game.RelicDefinitionStore = (*supabaseRelicDefinitionStore)(nil)
}
