package db

import (
	"context"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

// supabaseConfigStore implements game.ConfigStore via Supabase.
type supabaseConfigStore struct {
	client SupabaseClient
}

// NewConfigStore constructs a game.ConfigStore backed by Supabase.
func NewConfigStore(client SupabaseClient) game.ConfigStore {
	return &supabaseConfigStore{client: client}
}

func (s *supabaseConfigStore) GetSystemConfig(ctx context.Context, key string) (string, error) {
	v := url.Values{}
	v.Set("key", "eq."+key)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_system_config", params)
	if err != nil {
		return "", fmt.Errorf("get system config: %w", err)
	}

	return string(raw), nil
}

var _ game.ConfigStore = (*supabaseConfigStore)(nil)
