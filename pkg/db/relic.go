package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"odyssey/pkg/game"
)

// supabaseRelicStore implements game.RelicStore via Supabase.
type supabaseRelicStore struct {
	client SupabaseClient
}

// NewRelicStore constructs a game.RelicStore backed by Supabase.
func NewRelicStore(client SupabaseClient) game.RelicStore {
	return &supabaseRelicStore{client: client}
}

func (s *supabaseRelicStore) CreateRelic(ctx context.Context, r *game.Collection) (*game.Collection, error) {
	payload := Collection{
		UID:         r.UID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Journey:       r.Journey,
		Rarity:      r.Rarity,
		Image:       r.Image,
		Concept:        r.Concept,
		AwardedAt:   r.AwardedAt,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_collections", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create relic: %w", err)
	}

	var collections []Collection
	if err := json.Unmarshal(raw, &collections); err != nil {
		return nil, fmt.Errorf("parse created relic: %w", err)
	}
	if len(collections) == 0 {
		return r, nil
	}
	return mapRelic(collections[0]), nil
}

func (s *supabaseRelicStore) GetRelic(ctx context.Context, relicID int64) (*game.Collection, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(relicID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_collections", params)
	if err != nil {
		return nil, fmt.Errorf("get relic: %w", err)
	}

	var collections []Collection
	if err := json.Unmarshal(raw, &collections); err != nil {
		return nil, fmt.Errorf("parse relic: %w", err)
	}
	if len(collections) == 0 {
		return nil, game.ErrNotFound
	}
	return mapRelic(collections[0]), nil
}

func (s *supabaseRelicStore) ListRelics(ctx context.Context) ([]game.Collection, error) {
	raw, err := s.client.Get(ctx, "odyssey_collections", "")
	if err != nil {
		return nil, fmt.Errorf("list collections: %w", err)
	}

	var dbRelics []Collection
	if err := json.Unmarshal(raw, &dbRelics); err != nil {
		return nil, fmt.Errorf("parse collections: %w", err)
	}
	result := make([]game.Collection, 0, len(dbRelics))
	for i := range dbRelics {
		result = append(result, *mapRelic(dbRelics[i]))
	}
	return result, nil
}

func (s *supabaseRelicStore) CountRelics(ctx context.Context, uid string) (int, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("select", "id")
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_collections", params)
	if err != nil {
		return 0, fmt.Errorf("count collections: %w", err)
	}

	var collections []Collection
	if err := json.Unmarshal(raw, &collections); err != nil {
		return 0, fmt.Errorf("parse collections count: %w", err)
	}
	return len(collections), nil
}

var _ game.RelicStore = (*supabaseRelicStore)(nil)
