package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"odyssey/pkg/game"
)

// supabaseRealmProgressStore implements game.RealmProgressStore via Supabase.
type supabaseRealmProgressStore struct {
	client SupabaseClient
}

// NewRealmProgressStore constructs a game.RealmProgressStore backed by Supabase.
func NewRealmProgressStore(client SupabaseClient) game.RealmProgressStore {
	return &supabaseRealmProgressStore{client: client}
}

func (s *supabaseRealmProgressStore) GetRealmProgress(ctx context.Context, crewID, realm string) (*game.RealmProgress, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	v.Set("realm", "eq."+realm)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_realm_progress", params)
	if err != nil {
		return nil, fmt.Errorf("get realm progress: %w", err)
	}

	var progress []RealmProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, fmt.Errorf("parse realm progress: %w", err)
	}
	if len(progress) == 0 {
		return nil, game.ErrNotFound
	}

	return mapRealmProgress(progress[0]), nil
}

func (s *supabaseRealmProgressStore) CreateRealmProgress(ctx context.Context, rp *game.RealmProgress) (*game.RealmProgress, error) {
	payload := RealmProgress{
		CrewID:         rp.CrewID,
		Realm:          rp.Realm,
		Status:         rp.Status,
		StoryBranch:    rp.StoryBranch,
		Progress:       rp.Progress,
		LastUnlockedAt: rp.LastUnlockedAt,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_realm_progress", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create realm progress: %w", err)
	}

	var progress []RealmProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, fmt.Errorf("parse created realm progress: %w", err)
	}
	if len(progress) == 0 {
		return nil, fmt.Errorf("empty response creating realm progress")
	}
	return mapRealmProgress(progress[0]), nil
}

func (s *supabaseRealmProgressStore) UpdateRealmProgress(ctx context.Context, crewID, realm string, patch map[string]any) error {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	v.Set("realm", "eq."+realm)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_realm_progress", patch, params)
	if err != nil {
		return fmt.Errorf("update realm progress: %w", err)
	}
	return nil
}

func (s *supabaseRealmProgressStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, realm string, oldProgress int, patch map[string]any) (bool, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	v.Set("realm", "eq."+realm)
	v.Set("progress", "eq."+strconv.Itoa(oldProgress))
	params := v.Encode()
	raw, err := s.client.Mutate(ctx, "PATCH", "odyssey_realm_progress", patch, params+"&return=representation")
	if err != nil {
		return false, fmt.Errorf("update realm progress if match: %w", err)
	}

	var progress []RealmProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return false, fmt.Errorf("parse update realm progress response: %w", err)
	}
	return len(progress) > 0, nil
}

func (s *supabaseRealmProgressStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.RealmProgress, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_realm_progress", params)
	if err != nil {
		return nil, fmt.Errorf("list realm progress: %w", err)
	}

	var progress []RealmProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, fmt.Errorf("parse realm progress: %w", err)
	}

	result := make([]game.RealmProgress, 0, len(progress))
	for _, rp := range progress {
		result = append(result, *mapRealmProgress(rp))
	}
	return result, nil
}

func mapRealmProgress(rp RealmProgress) *game.RealmProgress {
	return &game.RealmProgress{
		CrewID:         rp.CrewID,
		Realm:          rp.Realm,
		Status:         rp.Status,
		StoryBranch:    rp.StoryBranch,
		Progress:       rp.Progress,
		LastUnlockedAt: rp.LastUnlockedAt,
		UpdatedAt:      rp.UpdatedAt,
	}
}

var _ game.RealmProgressStore = (*supabaseRealmProgressStore)(nil)
