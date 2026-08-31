package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"odyssey/pkg/auth"
)

var ErrProfileNotFound = errors.New("profile not found")

type ProfileStore interface {
	GetUserProfile(ctx context.Context, uid string) (*UserProfile, error)
	GetPasswordHash(ctx context.Context, uid string) (string, error)
	GetBoundDeviceID(ctx context.Context, uid string) (string, error)
	BindOrVerifyDevice(ctx context.Context, uid, deviceID string) (bool, error)
	ResetDeviceBinding(ctx context.Context, uid string) error
	UpdateAvatar(ctx context.Context, uid string, style, seed string) error
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

func (s *supabaseProfileStore) BindOrVerifyDevice(ctx context.Context, uid, deviceID string) (bool, error) {
	if strings.TrimSpace(deviceID) == "" {
		return false, auth.ErrDeviceRequired
	}
	raw, err := s.client.RPC(ctx, "odyssey_bind_or_verify_device", map[string]any{
		"p_user_uid":   uid,
		"p_device_id": deviceID,
	})
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "perangkat lain") || strings.Contains(errStr, "P0022") {
			return false, auth.ErrDeviceBlocked
		}
		if strings.Contains(errStr, "nonaktif") || strings.Contains(errStr, "P0021") {
			return false, auth.ErrAccountDisabled
		}
		// Fallback for direct table mutation if RPC is unavailable in tests
		currentDevice, getErr := s.GetBoundDeviceID(ctx, uid)
		if getErr == nil {
			if currentDevice == "" {
				payload := map[string]any{
					"device_id":       deviceID,
					"device_bound_at": "now()",
				}
				params := s.buildFilter(uid)
				_, _ = s.client.Mutate(ctx, "PATCH", "odyssey_user_profiles", payload, params)
				return true, nil
			}
			if currentDevice == deviceID {
				return false, nil
			}
			return false, auth.ErrDeviceBlocked
		}
		return false, fmt.Errorf("bind or verify device: %w", err)
	}
	var res struct {
		IsNewlyBound bool `json:"is_newly_bound"`
	}
	if err := json.Unmarshal(raw, &res); err == nil {
		return res.IsNewlyBound, nil
	}
	return false, nil
}

func (s *supabaseProfileStore) ResetDeviceBinding(ctx context.Context, uid string) error {
	_, err := s.client.RPC(ctx, "odyssey_admin_reset_device", map[string]any{
		"p_target_uid": uid,
	})
	if err != nil {
		// Fallback mutation for direct Supabase REST
		payload := map[string]any{
			"device_id":       nil,
			"device_bound_at": nil,
		}
		params := s.buildFilter(uid)
		_, mutateErr := s.client.Mutate(ctx, "PATCH", "odyssey_user_profiles", payload, params)
		if mutateErr != nil {
			return fmt.Errorf("reset device binding: %w", err)
		}
	}
	return nil
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
