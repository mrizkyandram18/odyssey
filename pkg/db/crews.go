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

func (s *supabaseCrewStore) GetCrew(ctx context.Context, crewID string) (*game.Crew, error) {
	v := url.Values{}
	v.Set("id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_crews", params)
	if err != nil {
		return nil, fmt.Errorf("get crew: %w", err)
	}

	var crews []Crew
	if err := json.Unmarshal(raw, &crews); err != nil {
		return nil, fmt.Errorf("parse crews: %w", err)
	}
	if len(crews) == 0 {
		return nil, game.ErrNotFound
	}

	c := crews[0]
	return &game.Crew{
		ID:        c.ID,
		Name:      c.Name,
		CreatedAt: c.CreatedAt,
		UpdatedAt: c.UpdatedAt,
	}, nil
}

func (s *supabaseCrewStore) CreateCrew(ctx context.Context, c *game.Crew) error {
	payload := Crew{
		ID:   c.ID,
		Name: c.Name,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_crews", payload, "")
	if err != nil {
		return fmt.Errorf("create crew: %w", err)
	}
	return nil
}

var _ game.CrewStore = (*supabaseCrewStore)(nil)
