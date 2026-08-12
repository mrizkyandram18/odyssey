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

func (s *supabaseRealmProgressStore) GetRealmProgress(ctx context.Context, crewID, journey string) (*game.JourneyProgress, error) {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	v.Set("journey", "eq."+journey)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_journey_progress", params)
	if err != nil {
		return nil, fmt.Errorf("get journey progress: %w", err)
	}

	var progress []JourneyProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, fmt.Errorf("parse journey progress: %w", err)
	}
	if len(progress) == 0 {
		return nil, game.ErrNotFound
	}

	return mapRealmProgress(progress[0]), nil
}

func (s *supabaseRealmProgressStore) CreateRealmProgress(ctx context.Context, rp *game.JourneyProgress) (*game.JourneyProgress, error) {
	payload := JourneyProgress{
		FamilyID:         rp.FamilyID,
		Journey:          rp.Journey,
		Status:         rp.Status,
		StoryBranch:    rp.StoryBranch,
		Progress:       rp.Progress,
		LastUnlockedAt: rp.LastUnlockedAt,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_journey_progress", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create journey progress: %w", err)
	}

	var progress []JourneyProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, fmt.Errorf("parse created journey progress: %w", err)
	}
	if len(progress) == 0 {
		return nil, fmt.Errorf("empty response creating journey progress")
	}
	return mapRealmProgress(progress[0]), nil
}

func (s *supabaseRealmProgressStore) UpdateRealmProgress(ctx context.Context, crewID, journey string, patch map[string]any) error {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	v.Set("journey", "eq."+journey)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_journey_progress", patch, params)
	if err != nil {
		return fmt.Errorf("update journey progress: %w", err)
	}
	return nil
}

func (s *supabaseRealmProgressStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, journey string, oldProgress int, patch map[string]any) (bool, error) {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	v.Set("journey", "eq."+journey)
	v.Set("progress", "eq."+strconv.Itoa(oldProgress))
	params := v.Encode()
	raw, err := s.client.Mutate(ctx, "PATCH", "odyssey_journey_progress", patch, params+"&return=representation")
	if err != nil {
		return false, fmt.Errorf("update journey progress if match: %w", err)
	}

	var progress []JourneyProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return false, fmt.Errorf("parse update journey progress response: %w", err)
	}
	return len(progress) > 0, nil
}

func (s *supabaseRealmProgressStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.JourneyProgress, error) {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_journey_progress", params)
	if err != nil {
		return nil, fmt.Errorf("list journey progress: %w", err)
	}

	var progress []JourneyProgress
	if err := json.Unmarshal(raw, &progress); err != nil {
		return nil, fmt.Errorf("parse journey progress: %w", err)
	}

	result := make([]game.JourneyProgress, 0, len(progress))
	for _, rp := range progress {
		result = append(result, *mapRealmProgress(rp))
	}
	return result, nil
}

func mapRealmProgress(rp JourneyProgress) *game.JourneyProgress {
	return &game.JourneyProgress{
		FamilyID:         rp.FamilyID,
		Journey:          rp.Journey,
		Status:         rp.Status,
		StoryBranch:    rp.StoryBranch,
		Progress:       rp.Progress,
		LastUnlockedAt: rp.LastUnlockedAt,
		UpdatedAt:      rp.UpdatedAt,
	}
}

var _ game.RealmProgressStore = (*supabaseRealmProgressStore)(nil)
