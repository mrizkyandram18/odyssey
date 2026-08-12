package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

// supabaseChestDefinitionStore implements game.ChestDefinitionStore via Supabase.
type supabaseChestDefinitionStore struct {
	client SupabaseClient
}

// NewChestDefinitionStore constructs a game.ChestDefinitionStore backed by Supabase.
func NewChestDefinitionStore(client SupabaseClient) game.ChestDefinitionStore {
	return &supabaseChestDefinitionStore{client: client}
}

func (s *supabaseChestDefinitionStore) ListChestDefinitions(ctx context.Context) ([]game.ChestDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_chest_definitions", "")
	if err != nil {
		return nil, fmt.Errorf("list chest definitions: %w", err)
	}

	var dbDefs []ChestDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse chest definitions: %w", err)
	}
	result := make([]game.ChestDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapChestDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseChestDefinitionStore) GetChestDefinition(ctx context.Context, slug string) (*game.ChestDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_chest_definitions", params)
	if err != nil {
		return nil, fmt.Errorf("get chest definition: %w", err)
	}

	var defs []ChestDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse chest definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, game.ErrNotFound
	}
	return mapChestDefinition(defs[0]), nil
}

func (s *supabaseChestDefinitionStore) ListDropTableEntries(ctx context.Context, chestSlug string) ([]game.DropTableEntry, error) {
	v := url.Values{}
	v.Set("gift_slug", "eq."+chestSlug)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_drop_tables", params)
	if err != nil {
		return nil, fmt.Errorf("list drop table entries: %w", err)
	}

	var dbEntries []DropTableEntry
	if err := json.Unmarshal(raw, &dbEntries); err != nil {
		return nil, fmt.Errorf("parse drop table entries: %w", err)
	}
	result := make([]game.DropTableEntry, 0, len(dbEntries))
	for i := range dbEntries {
		result = append(result, *mapDropTableEntry(dbEntries[i]))
	}
	return result, nil
}

func mapChestDefinition(d ChestDefinition) *game.ChestDefinition {
	return &game.ChestDefinition{
		ID:          d.ID,
		Slug:        d.Slug,
		Name:        d.Name,
		Rarity:      d.Rarity,
		Icon:        d.Icon,
		Description: d.Description,
		SeasonSlug:  d.SeasonSlug,
		Published:   d.Published,
		Version:     d.Version,
		UpdatedBy:   d.UpdatedBy,
		DeletedAt:   d.DeletedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func mapDropTableEntry(e DropTableEntry) *game.DropTableEntry {
	return &game.DropTableEntry{
		ID:        e.ID,
		GiftSlug: e.GiftSlug,
		CollectionID:   e.CollectionID,
		Rarity:    e.Rarity,
		Weight:    e.Weight,
		CreatedAt: e.CreatedAt,
	}
}

var _ game.ChestDefinitionStore = (*supabaseChestDefinitionStore)(nil)
