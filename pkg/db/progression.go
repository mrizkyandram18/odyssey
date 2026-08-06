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

func (s *supabaseProgressionStore) CreateRelic(ctx context.Context, r *game.Relic) (*game.Relic, error) {
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

func (s *supabaseProgressionStore) CreateChest(ctx context.Context, ch *game.Chest) (*game.Chest, error) {
	payload := Chest{
		UID:         ch.UID,
		Source:      ch.Source,
		ChestSlug:   ch.ChestSlug,
		Rarity:      ch.Rarity,
		Icon:        ch.Icon,
		Description: ch.Description,
		DropTable:   ch.DropTable,
		Opened:      ch.Opened,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_chests", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create chest: %w", err)
	}

	var chests []Chest
	if err := json.Unmarshal(raw, &chests); err != nil {
		return nil, fmt.Errorf("parse created chest: %w", err)
	}
	if len(chests) == 0 {
		return ch, nil
	}
	return mapChest(chests[0]), nil
}

func (s *supabaseProgressionStore) UpdateChest(ctx context.Context, chestID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(chestID, 10))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_chests", patch, params)
	if err != nil {
		return fmt.Errorf("update chest: %w", err)
	}
	return nil
}

func (s *supabaseProgressionStore) CreateAchievement(ctx context.Context, a *game.Achievement) (*game.Achievement, error) {
	payload := map[string]any{
		"uid":        a.UID,
		"crew_id":    a.CrewID,
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

func mapRelic(r Relic) *game.Relic {
	return &game.Relic{
		ID:          r.ID,
		UID:         r.UID,
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		Realm:       r.Realm,
		Rarity:      r.Rarity,
		Image:       r.Image,
		Lore:        r.Lore,
		AwardedAt:   r.AwardedAt,
		CreatedAt:   r.CreatedAt,
	}
}

func mapAchievement(a Achievement) *game.Achievement {
	return &game.Achievement{
		ID:              a.ID,
		UID:             a.UID,
		CrewID:          a.CrewID,
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
