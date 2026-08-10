package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileStore interface {
	GetUserProfile(ctx context.Context, uid string) (*UserProfile, error)
	GetPasswordHash(ctx context.Context, uid string) (string, error)
	GetBoundDeviceID(ctx context.Context, uid string) (string, error)
	UpdateAvatar(ctx context.Context, uid string, style, seed string) error
	SetAvatarFrame(ctx context.Context, uid, frame string) error
}

type supabaseProfileStore struct {
	client SupabaseClient
}

func NewProfileStore(client SupabaseClient) ProfileStore {
	return &supabaseProfileStore{client: client}
}

func (s *supabaseProfileStore) buildFilter(uid string) string {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	return v.Encode()
}

func (s *supabaseProfileStore) GetUserProfile(ctx context.Context, uid string) (*UserProfile, error) {
	params := s.buildFilter(uid)
	raw, err := s.client.Get(ctx, "odyssey_user_profiles", params)
	if err != nil {
		return nil, fmt.Errorf("get user profile: %w", err)
	}
	var profiles []UserProfile
	if err := json.Unmarshal(raw, &profiles); err != nil {
		return nil, fmt.Errorf("parse user profile: %w", err)
	}
	if len(profiles) == 0 {
		return nil, ErrProfileNotFound
	}
	return &profiles[0], nil
}

func (s *supabaseProfileStore) GetPasswordHash(ctx context.Context, uid string) (string, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("select", "password_hash")
	params := v.Encode()
	raw, err := s.client.Get(ctx, "odyssey_user_profiles", params)
	if err != nil {
		return "", fmt.Errorf("get password hash: %w", err)
	}
	var rows []struct {
		PasswordHash string `json:"password_hash"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return "", fmt.Errorf("parse password hash: %w", err)
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0].PasswordHash, nil
}

func (s *supabaseProfileStore) GetBoundDeviceID(ctx context.Context, uid string) (string, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	v.Set("select", "device_id")
	params := v.Encode()
	raw, err := s.client.Get(ctx, "odyssey_user_profiles", params)
	if err != nil {
		return "", fmt.Errorf("get bound device: %w", err)
	}
	var rows []struct {
		DeviceID string `json:"device_id"`
	}
	if err := json.Unmarshal(raw, &rows); err != nil {
		return "", fmt.Errorf("parse bound device: %w", err)
	}
	if len(rows) == 0 {
		return "", nil
	}
	return rows[0].DeviceID, nil
}

func (s *supabaseProfileStore) UpdateAvatar(ctx context.Context, uid string, style, seed string) error {
	payload := map[string]string{
		"avatar_style": style,
		"avatar_seed":  seed,
	}
	params := s.buildFilter(uid)
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_user_profiles", payload, params)
	if err != nil {
		return fmt.Errorf("update avatar: %w", err)
	}
	return nil
}
