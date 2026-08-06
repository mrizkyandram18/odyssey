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

func (s *supabaseRelicStore) CreateRelic(ctx context.Context, r *game.Relic) (*game.Relic, error) {
	payload := Relic{
		UID:         r.UID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Realm:       r.Realm,
		Rarity:      r.Rarity,
		Image:       r.Image,
		Lore:        r.Lore,
		AwardedAt:   r.AwardedAt,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_relics", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create relic: %w", err)
	}

	var relics []Relic
	if err := json.Unmarshal(raw, &relics); err != nil {
		return nil, fmt.Errorf("parse created relic: %w", err)
	}
	if len(relics) == 0 {
		return r, nil
	}
	return mapRelic(relics[0]), nil
}

func (s *supabaseRelicStore) GetRelic(ctx context.Context, relicID int64) (*game.Relic, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(relicID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_relics", params)
	if err != nil {
		return nil, fmt.Errorf("get relic: %w", err)
	}

	var relics []Relic
	if err := json.Unmarshal(raw, &relics); err != nil {
		return nil, fmt.Errorf("parse relic: %w", err)
	}
	if len(relics) == 0 {
		return nil, game.ErrNotFound
	}
	return mapRelic(relics[0]), nil
}

func (s *supabaseRelicStore) ListRelics(ctx context.Context) ([]game.Relic, error) {
	raw, err := s.client.Get(ctx, "odyssey_relics", "")
	if err != nil {
		return nil, fmt.Errorf("list relics: %w", err)
	}

	var dbRelics []Relic
	if err := json.Unmarshal(raw, &dbRelics); err != nil {
		return nil, fmt.Errorf("parse relics: %w", err)
	}
	result := make([]game.Relic, 0, len(dbRelics))
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

	raw, err := s.client.Get(ctx, "odyssey_relics", params)
	if err != nil {
		return 0, fmt.Errorf("count relics: %w", err)
	}

	var relics []Relic
	if err := json.Unmarshal(raw, &relics); err != nil {
		return 0, fmt.Errorf("parse relics count: %w", err)
	}
	return len(relics), nil
}

var _ game.RelicStore = (*supabaseRelicStore)(nil)
