package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"odyssey/pkg/game"
)

// supabaseProgressionStore implements game.ProgressionStore via Supabase.
type supabaseProgressionStore struct {
	client SupabaseClient
}

// NewProgressionStore constructs a game.ProgressionStore backed by Supabase.
func NewProgressionStore(client SupabaseClient) game.ProgressionStore {
	return &supabaseProgressionStore{client: client}
}

// NewAchievementStore constructs a game.AchievementStore backed by Supabase.
func NewAchievementStore(client SupabaseClient) game.AchievementStore {
	return &supabaseProgressionStore{client: client}
}

func (s *supabaseProgressionStore) CreateRelic(ctx context.Context, r *game.Collection) (*game.Collection, error) {
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

func (s *supabaseProgressionStore) CreateChest(ctx context.Context, ch *game.Gift) (*game.Gift, error) {
	payload := Gift{
		UID:         ch.UID,
		Source:      ch.Source,
		GiftSlug:   ch.GiftSlug,
		Rarity:      ch.Rarity,
		Icon:        ch.Icon,
		Description: ch.Description,
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

func (s *supabaseProgressionStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(chestID, 10))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_gifts", patch, params)
	if err != nil {
		return fmt.Errorf("update chest: %w", err)
	}
	return nil
}

func (s *supabaseProgressionStore) CreateAchievement(ctx context.Context, a *game.Achievement) (*game.Achievement, error) {
	payload := map[string]any{
		"uid":        a.UID,
		"family_id":    a.FamilyID,
		"code":       a.Code,
		"kind":       a.Kind,
		"awarded_at": a.AwardedAt,
	}
	if a.Trigger != "" {
		payload["trigger"] = a.Trigger
	}
	if a.CompletionCount > 0 {
		payload["completion_count"] = a.CompletionCount
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_achievements", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create achievement: %w", err)
	}

	var achievements []Achievement
	if err := json.Unmarshal(raw, &achievements); err != nil {
		return nil, fmt.Errorf("parse created achievement: %w", err)
	}
	if len(achievements) == 0 {
		return a, nil
	}
	return mapAchievement(achievements[0]), nil
}

func (s *supabaseProgressionStore) CountRelics(ctx context.Context, uid string) (int, error) {
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

func mapRelic(r Collection) *game.Collection {
	return &game.Collection{
		ID:          r.ID,
		UID:         r.UID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Journey:       r.Journey,
		Rarity:      r.Rarity,
		Image:       r.Image,
		Concept:        r.Concept,
		AwardedAt:   r.AwardedAt,
		CreatedAt:   r.CreatedAt,
	}
}

func mapAchievement(a Achievement) *game.Achievement {
	return &game.Achievement{
		ID:              a.ID,
		UID:             a.UID,
		FamilyID:          a.FamilyID,
		Code:            a.Code,
		Kind:            a.Kind,
		Trigger:         a.Trigger,
		CompletionCount: a.CompletionCount,
		AwardedAt:       a.AwardedAt,
		CreatedAt:       a.CreatedAt,
	}
}

func (s *supabaseProgressionStore) GetAchievementByCode(ctx context.Context, uid, code string) (*game.Achievement, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("code", "eq."+code)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_achievements", params)
	if err != nil {
		return nil, fmt.Errorf("get achievement: %w", err)
	}

	var rows []Achievement
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse achievement: %w", err)
	}
	if len(rows) == 0 {
		return nil, game.ErrNotFound
	}
	return mapAchievement(rows[0]), nil
}

func (s *supabaseProgressionStore) ListAchievementsByPlayer(ctx context.Context, uid string) ([]game.Achievement, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_achievements", params)
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}

	var rows []Achievement
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse achievements: %w", err)
	}

	result := make([]game.Achievement, 0, len(rows))
	for i := range rows {
		result = append(result, *mapAchievement(rows[i]))
	}
	return result, nil
}

func (s *supabaseProgressionStore) CountAchievementsByKind(ctx context.Context, uid string, kind string) (int, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("trigger", "eq."+kind)
	v.Set("select", "id")
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_achievements", params)
	if err != nil {
		return 0, fmt.Errorf("count achievements by kind: %w", err)
	}

	var rows []Achievement
	if err := json.Unmarshal(raw, &rows); err != nil {
		return 0, fmt.Errorf("parse achievement count: %w", err)
	}
	return len(rows), nil
}

var _ game.ProgressionStore = (*supabaseProgressionStore)(nil)
var _ game.AchievementStore = (*supabaseProgressionStore)(nil)
