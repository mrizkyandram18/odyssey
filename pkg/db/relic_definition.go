package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

// supabaseRelicDefinitionStore implements game.RelicDefinitionStore via Supabase.
type supabaseRelicDefinitionStore struct {
	client SupabaseClient
}

// NewRelicDefinitionStore constructs a game.RelicDefinitionStore backed by Supabase.
func NewRelicDefinitionStore(client SupabaseClient) game.RelicDefinitionStore {
	return &supabaseRelicDefinitionStore{client: client}
}

func (s *supabaseRelicDefinitionStore) ListRelicDefinitions(ctx context.Context) ([]game.RelicDefinition, error) {
	raw, err := s.client.Get(ctx, "odyssey_relic_definitions", "")
	if err != nil {
		return nil, fmt.Errorf("list relic definitions: %w", err)
	}

	var dbDefs []RelicDefinition
	if err := json.Unmarshal(raw, &dbDefs); err != nil {
		return nil, fmt.Errorf("parse relic definitions: %w", err)
	}
	result := make([]game.RelicDefinition, 0, len(dbDefs))
	for i := range dbDefs {
		result = append(result, *mapRelicDefinition(dbDefs[i]))
	}
	return result, nil
}

func (s *supabaseRelicDefinitionStore) GetRelicDefinition(ctx context.Context, slug string) (*game.RelicDefinition, error) {
	v := url.Values{}
	v.Set("slug", "eq."+slug)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_relic_definitions", params)
	if err != nil {
		return nil, fmt.Errorf("get relic definition: %w", err)
	}

	var defs []RelicDefinition
	if err := json.Unmarshal(raw, &defs); err != nil {
		return nil, fmt.Errorf("parse relic definition: %w", err)
	}
	if len(defs) == 0 {
		return nil, game.ErrNotFound
	}
	return mapRelicDefinition(defs[0]), nil
}

func mapRelicDefinition(d RelicDefinition) *game.RelicDefinition {
	return &game.RelicDefinition{
		ID:          d.ID,
		Slug:        d.Slug,
		Name:        d.Name,
		Description: d.Description,
		Realm:       d.Realm,
		Rarity:      d.Rarity,
		Image:       d.Image,
		Lore:        d.Lore,
		Published:   d.Published,
		Version:     d.Version,
		UpdatedBy:   d.UpdatedBy,
		DeletedAt:   d.DeletedAt,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

var _ game.RelicDefinitionStore = (*supabaseRelicDefinitionStore)(nil)
