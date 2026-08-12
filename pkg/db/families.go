package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

// supabaseCrewStore implements game.CrewStore via Supabase.
type supabaseCrewStore struct {
	client SupabaseClient
}

// NewCrewStore constructs a game.CrewStore backed by Supabase.
func NewCrewStore(client SupabaseClient) game.CrewStore {
	return &supabaseCrewStore{client: client}
}

func (s *supabaseCrewStore) GetCrew(ctx context.Context, crewID string) (*game.Family, error) {
	v := url.Values{}
	v.Set("id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_families", params)
	if err != nil {
		return nil, fmt.Errorf("get crew: %w", err)
	}

	var families []Family
	if err := json.Unmarshal(raw, &families); err != nil {
		return nil, fmt.Errorf("parse families: %w", err)
	}
	if len(families) == 0 {
		return nil, game.ErrNotFound
	}

	c := families[0]
	return &game.Family{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}, nil
}

func (s *supabaseCrewStore) CreateCrew(ctx context.Context, c *game.Family) error {
	payload := Family{
		ID:   c.ID,
		Name: c.Name,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_families", payload, "")
	if err != nil {
		return fmt.Errorf("create crew: %w", err)
	}
	return nil
}

func (s *supabaseCrewStore) UpdateCrew(ctx context.Context, crewID string, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+crewID)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_families", patch, params)
	if err != nil {
		return fmt.Errorf("update crew: %w", err)
	}
	return nil
}

var _ game.CrewStore = (*supabaseCrewStore)(nil)
