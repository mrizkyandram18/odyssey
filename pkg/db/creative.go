package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

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
	// Omit id so identity is assigned by PostgREST.
	payload := map[string]any{
		"family_id":    item.FamilyID,
		"journey":      item.Journey,
		"author_uid": item.AuthorUID,
		"kind":       item.Kind,
		"payload":    item.Payload,
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

func (s *supabaseCreativeStore) GetCreativeItem(ctx context.Context, id int64) (*game.CreativeItem, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(id, 10))
	v.Set("limit", "1")
	raw, err := s.client.Get(ctx, "odyssey_creative_items", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get creative item: %w", err)
	}
	var items []CreativeItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse creative item: %w", err)
	}
	if len(items) == 0 {
		return nil, game.ErrNotFound
	}
	return mapCreativeItem(items[0]), nil
}

func (s *supabaseCreativeStore) ListCreativeItemsByCrew(ctx context.Context, crewID, kind string) ([]game.CreativeItem, error) {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	if kind != "" {
		v.Set("kind", "eq."+kind)
	}
	v.Set("order", "created_at.desc")
	raw, err := s.client.Get(ctx, "odyssey_creative_items", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("list creative items: %w", err)
	}
	var items []CreativeItem
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse creative items: %w", err)
	}
	out := make([]game.CreativeItem, 0, len(items))
	for _, it := range items {
		out = append(out, *mapCreativeItem(it))
	}
	return out, nil
}

func mapCreativeItem(i CreativeItem) *game.CreativeItem {
	return &game.CreativeItem{
		ID:        i.ID,
		FamilyID:    i.FamilyID,
		Journey:     i.Journey,
		AuthorUID: i.AuthorUID,
		Kind:      i.Kind,
		Payload:   i.Payload,
		CreatedAt: i.CreatedAt,
	}
}

var _ game.CreativeStore = (*supabaseCreativeStore)(nil)
