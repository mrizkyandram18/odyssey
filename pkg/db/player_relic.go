package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"time"

	"odyssey/pkg/game"
)

// supabasePlayerRelicStore implements game.PlayerRelicStore via Supabase.
type supabasePlayerRelicStore struct {
	client SupabaseClient
}

// NewPlayerRelicStore constructs a game.PlayerRelicStore backed by Supabase.
func NewPlayerRelicStore(client SupabaseClient) game.PlayerRelicStore {
	return &supabasePlayerRelicStore{client: client}
}

func (s *supabasePlayerRelicStore) GetPlayerRelic(ctx context.Context, uid, relicSlug string) (*game.PlayerRelic, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("relic_slug", "eq."+relicSlug)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_player_relics", params)
	if err != nil {
		return nil, fmt.Errorf("get player relic: %w", err)
	}

	var items []PlayerRelic
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse player relic: %w", err)
	}
	if len(items) == 0 {
		return nil, game.ErrNotFound
	}
	return mapPlayerRelic(items[0]), nil
}

func (s *supabasePlayerRelicStore) CreatePlayerRelic(ctx context.Context, pr *game.PlayerRelic) (*game.PlayerRelic, error) {
	now := time.Now().UTC()
	payload := PlayerRelic{
		UID:          pr.UID,
		RelicSlug:    pr.RelicSlug,
		RelicID:      pr.RelicID,
		OwnedCount:   pr.OwnedCount,
		IsNew:        pr.IsNew,
		DiscoveredAt: pr.DiscoveredAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_player_relics", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create player relic: %w", err)
	}

	var items []PlayerRelic
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse created player relic: %w", err)
	}
	if len(items) == 0 {
		return pr, nil
	}
	return mapPlayerRelic(items[0]), nil
}

func (s *supabasePlayerRelicStore) UpdatePlayerRelic(ctx context.Context, uid, relicSlug string, patch map[string]any) error {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("relic_slug", "eq."+relicSlug)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_player_relics", patch, params)
	if err != nil {
		return fmt.Errorf("update player relic: %w", err)
	}
	return nil
}

func (s *supabasePlayerRelicStore) ListPlayerRelics(ctx context.Context, uid string) ([]game.PlayerRelic, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_player_relics", params)
	if err != nil {
		return nil, fmt.Errorf("list player relics: %w", err)
	}

	var items []PlayerRelic
	if err := json.Unmarshal(raw, &items); err != nil {
		return nil, fmt.Errorf("parse player relics: %w", err)
	}
	result := make([]game.PlayerRelic, 0, len(items))
	for i := range items {
		result = append(result, *mapPlayerRelic(items[i]))
	}
	return result, nil
}

func (s *supabasePlayerRelicStore) CountUniqueRelics(ctx context.Context, uid string) (int, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("select", "relic_slug")
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_player_relics", params)
	if err != nil {
		return 0, fmt.Errorf("count unique relics: %w", err)
	}

	var items []PlayerRelic
	if err := json.Unmarshal(raw, &items); err != nil {
		return 0, fmt.Errorf("parse unique relics count: %w", err)
	}
	seen := make(map[string]bool)
	count := 0
	for _, item := range items {
		if !seen[item.RelicSlug] {
			seen[item.RelicSlug] = true
			count++
		}
	}
	return count, nil
}

func mapPlayerRelic(pr PlayerRelic) *game.PlayerRelic {
	return &game.PlayerRelic{
		UID:          pr.UID,
		RelicSlug:    pr.RelicSlug,
		RelicID:      pr.RelicID,
		OwnedCount:   pr.OwnedCount,
		IsNew:        pr.IsNew,
		DiscoveredAt: pr.DiscoveredAt,
		CreatedAt:    pr.CreatedAt,
		UpdatedAt:    pr.UpdatedAt,
	}
}

var _ game.PlayerRelicStore = (*supabasePlayerRelicStore)(nil)
