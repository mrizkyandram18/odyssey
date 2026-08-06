package db

import (
	"context"
	"encoding/json"
	"fmt"

	"odyssey/pkg/game"
)

// supabaseCreativeStore implements game.CreativeStore via Supabase.
type supabaseCreativeStore struct {
	client SupabaseClient
}

// NewCreativeStore constructs a game.CreativeStore backed by Supabase.
func NewCreativeStore(client SupabaseClient) game.CreativeStore {
	return &supabaseCreativeStore{client: client}
}

func (s *supabaseCreativeStore) CreateCreativeItem(ctx context.Context, item *game.CreativeItem) (*game.CreativeItem, error) {
	payload := CreativeItem{
		CrewID:    item.CrewID,
		Realm:     item.Realm,
		AuthorUID: item.AuthorUID,
		Kind:      item.Kind,
		Payload:   item.Payload,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_creative_items", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create creative item: %w", err)
	}

	var items []CreativeItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse created creative item: %w", err)
	}
	if len(items) == 0 {
		return item, nil
	}
	return mapCreativeItem(items[0]), nil
}

func mapCreativeItem(i CreativeItem) *game.CreativeItem {
	return &game.CreativeItem{
		ID:        i.ID,
		CrewID:    i.CrewID,
		Realm:     i.Realm,
		AuthorUID: i.AuthorUID,
		Kind:      i.Kind,
		Payload:   i.Payload,
		CreatedAt: i.CreatedAt,
	}
}

var _ game.CreativeStore = (*supabaseCreativeStore)(nil)
