package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"odyssey/pkg/game"
)

// supabaseUserStore implements game.UserStore via Supabase.
type supabaseUserStore struct {
	client SupabaseClient
}

// NewUserStore constructs a game.UserStore backed by Supabase.
func NewUserStore(client SupabaseClient) game.UserStore {
	return &supabaseUserStore{client: client}
}

func (s *supabaseUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_user_profiles", params)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	var profiles []UserProfile
	if err := json.Unmarshal(raw, &profiles); err != nil {
		return nil, fmt.Errorf("parse user profiles: %w", err)
	}
	if len(profiles) == 0 {
		return nil, game.ErrNotFound
	}

	p := profiles[0]
	return &game.Player{
		UID:          p.UID,
		CrewID:       p.CrewID,
		ExplorerName: p.ExplorerName,
		Role:         p.Role,
		Level:        p.Level,
		XP:           p.XP,
		Version:      p.Version,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}, nil
}

func (s *supabaseUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	payload := UserProfile{
		UID:          p.UID,
		CrewID:       p.CrewID,
		ExplorerName: p.ExplorerName,
		Role:         p.Role,
		Level:        p.Level,
		XP:           p.XP,
		Version:      1,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_user_profiles", payload, "")
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (s *supabaseUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_user_profiles", patch, params)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

func (s *supabaseUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("version", "eq."+strconv.Itoa(version))
	params := v.Encode()
	raw, err := s.client.Mutate(ctx, "PATCH", "odyssey_user_profiles", patch, params+"&return=representation")
	if err != nil {
		return false, fmt.Errorf("update user if match: %w", err)
	}

	var profiles []UserProfile
	if err := json.Unmarshal(raw, &profiles); err != nil {
		return false, fmt.Errorf("parse update user response: %w", err)
	}
	return len(profiles) > 0, nil
}

var _ game.UserStore = (*supabaseUserStore)(nil)
