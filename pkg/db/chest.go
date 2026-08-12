package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"odyssey/pkg/game"
)

// supabaseChestStore implements game.ChestStore via Supabase.
type supabaseChestStore struct {
	client SupabaseClient
}

// NewChestStore constructs a game.ChestStore backed by Supabase.
func NewChestStore(client SupabaseClient) game.ChestStore {
	return &supabaseChestStore{client: client}
}

func (s *supabaseChestStore) CreateChest(ctx context.Context, ch *game.Gift) (*game.Gift, error) {
	payload := Gift{
		UID:         ch.UID,
		Source:      ch.Source,
		GiftSlug:   ch.GiftSlug,
		Rarity:      ch.Rarity,
		Icon:        ch.Icon,
		Description: ch.Description,
		RewardRelic: ch.RewardRelic,
		DropTable:   ch.DropTable,
		Opened:      ch.Opened,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_gifts", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create chest: %w", err)
	}

	var gifts []Gift
	if err := json.Unmarshal(raw, &gifts); err != nil {
		return nil, fmt.Errorf("parse created chest: %w", err)
	}
	if len(gifts) == 0 {
		return ch, nil
	}
	return mapChest(gifts[0]), nil
}

func (s *supabaseChestStore) GetChest(ctx context.Context, chestID int64) (*game.Gift, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(chestID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_gifts", params)
	if err != nil {
		return nil, fmt.Errorf("get chest: %w", err)
	}

	var gifts []Gift
	if err := json.Unmarshal(raw, &gifts); err != nil {
		return nil, fmt.Errorf("parse chest: %w", err)
	}
	if len(gifts) == 0 {
		return nil, game.ErrNotFound
	}
	return mapChest(gifts[0]), nil
}

func (s *supabaseChestStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(chestID, 10))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_gifts", patch, params)
	if err != nil {
		return fmt.Errorf("update chest: %w", err)
	}
	return nil
}

func (s *supabaseChestStore) UpdateChestIfMatch(ctx context.Context, chestID int64, oldOpened bool, patch map[string]any) (bool, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(chestID, 10))
	if oldOpened {
		v.Set("opened", "eq.true")
	} else {
		v.Set("opened", "eq.false")
	}
	params := v.Encode()
	raw, err := s.client.Mutate(ctx, "PATCH", "odyssey_gifts", patch, params+"&return=representation")
	if err != nil {
		return false, fmt.Errorf("update chest if match: %w", err)
	}

	var gifts []Gift
	if err := json.Unmarshal(raw, &gifts); err != nil {
		return false, fmt.Errorf("parse update chest response: %w", err)
	}
	return len(gifts) > 0, nil
}

func (s *supabaseChestStore) ListChestsByUser(ctx context.Context, uid string) ([]game.Gift, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_gifts", params)
	if err != nil {
		return nil, fmt.Errorf("list gifts: %w", err)
	}

	var gifts []Gift
	if err := json.Unmarshal(raw, &gifts); err != nil {
		return nil, fmt.Errorf("parse gifts: %w", err)
	}
	result := make([]game.Gift, 0, len(gifts))
	for i := range gifts {
		result = append(result, *mapChest(gifts[i]))
	}
	return result, nil
}

func mapChest(ch Gift) *game.Gift {
	return &game.Gift{
		ID:          ch.ID,
		UID:         ch.UID,
		Source:      ch.Source,
		GiftSlug:   ch.GiftSlug,
		Rarity:      ch.Rarity,
		Icon:        ch.Icon,
		Description: ch.Description,
		RewardRelic: ch.RewardRelic,
		DropTable:   ch.DropTable,
		Opened:      ch.Opened,
		OpenedAt:    ch.OpenedAt,
		CreatedAt:   ch.CreatedAt,
	}
}

var _ game.ChestStore = (*supabaseChestStore)(nil)
