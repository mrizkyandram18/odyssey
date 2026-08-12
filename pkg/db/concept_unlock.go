package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

type supabaseLoreUnlockStore struct {
	client SupabaseClient
}

func NewLoreUnlockStore(client SupabaseClient) game.LoreUnlockStore {
	return &supabaseLoreUnlockStore{client: client}
}

func (s *supabaseLoreUnlockStore) GetLoreUnlock(ctx context.Context, crewID, loreSlug string) (*game.LoreUnlock, error) {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	v.Set("concept_slug", "eq."+loreSlug)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_concept_unlocks", params)
	if err != nil {
		return nil, fmt.Errorf("get concept unlock: %w", err)
	}

	var rows []LoreUnlock
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse concept unlock: %w", err)
	}
	if len(rows) == 0 {
		return nil, game.ErrNotFound
	}
	return mapLoreUnlock(rows[0]), nil
}

func (s *supabaseLoreUnlockStore) CreateLoreUnlock(ctx context.Context, lu *game.LoreUnlock) (*game.LoreUnlock, error) {
	payload := LoreUnlock{
		FamilyID:     lu.FamilyID,
		ConceptSlug:   lu.ConceptSlug,
		Journey:      lu.Journey,
		Course:    lu.Course,
		UnlockedAt: lu.UnlockedAt,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_concept_unlocks", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create concept unlock: %w", err)
	}

	var rows []LoreUnlock
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse created concept unlock: %w", err)
	}
	if len(rows) == 0 {
		return lu, nil
	}
	return mapLoreUnlock(rows[0]), nil
}

func (s *supabaseLoreUnlockStore) UpdateLoreUnlock(ctx context.Context, crewID, loreSlug string, patch map[string]any) error {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	v.Set("concept_slug", "eq."+loreSlug)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_concept_unlocks", patch, params)
	if err != nil {
		return fmt.Errorf("update concept unlock: %w", err)
	}
	return nil
}

func (s *supabaseLoreUnlockStore) ListLoreUnlocksByCrew(ctx context.Context, crewID string) ([]game.LoreUnlock, error) {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_concept_unlocks", params)
	if err != nil {
		return nil, fmt.Errorf("list concept unlocks: %w", err)
	}

	var rows []LoreUnlock
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse concept unlocks: %w", err)
	}

	result := make([]game.LoreUnlock, 0, len(rows))
	for i := range rows {
		result = append(result, *mapLoreUnlock(rows[i]))
	}
	return result, nil
}

func mapLoreUnlock(lu LoreUnlock) *game.LoreUnlock {
	return &game.LoreUnlock{
		FamilyID:     lu.FamilyID,
		ConceptSlug:   lu.ConceptSlug,
		Journey:      lu.Journey,
		Course:    lu.Course,
		UnlockedAt: lu.UnlockedAt,
		CreatedAt:  lu.CreatedAt,
	}
}

var _ game.LoreUnlockStore = (*supabaseLoreUnlockStore)(nil)
