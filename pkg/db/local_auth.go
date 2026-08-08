package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/auth"
)

type supabaseLocalUserStore struct {
	client SupabaseClient
}

func NewLocalUserStore(client SupabaseClient) auth.LocalUserStore {
	return &supabaseLocalUserStore{client: client}
}

func (s *supabaseLocalUserStore) GetLocalUserByUsername(ctx context.Context, username string) (*auth.LocalUser, error) {
	v := url.Values{}
	v.Set("username", "eq."+username)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_local_users", params)
	if err != nil {
		return nil, fmt.Errorf("get local user: %w", err)
	}

	var users []struct {
		ID           string `json:"id"`
		Username     string `json:"username"`
		PasswordHash string `json:"password_hash"`
		ProfileUID   string `json:"profile_uid"`
	}

	if err := json.Unmarshal(raw, &users); err != nil {
		return nil, fmt.Errorf("parse local user: %w", err)
	}

	if len(users) == 0 {
		return nil, auth.ErrLocalUserNotFound
	}

	u := users[0]
	return &auth.LocalUser{
		ID:           u.ID,
		Username:     u.Username,
		PasswordHash: u.PasswordHash,
		ProfileUID:   u.ProfileUID,
	}, nil
}
