package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

var ErrFamilyNotFound = errors.New("family not found")

type FamilyStore interface {
	GetFamily(ctx context.Context, familyID string) (*Family, error)
	CreateFamily(ctx context.Context, f *Family) error
	UpdateFamily(ctx context.Context, familyID string, patch map[string]any) error
}

type supabaseFamilyStore struct {
	client SupabaseClient
}

func NewFamilyStore(client SupabaseClient) FamilyStore {
	return &supabaseFamilyStore{client: client}
}

func (s *supabaseFamilyStore) GetFamily(ctx context.Context, familyID string) (*Family, error) {
	v := url.Values{}
	v.Set("id", "eq."+familyID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_families", params)
	if err != nil {
		return nil, fmt.Errorf("get family: %w", err)
	}

	var families []Family
	if err := json.Unmarshal(raw, &families); err != nil {
		return nil, fmt.Errorf("parse families: %w", err)
	}
	if len(families) == 0 {
		return nil, ErrFamilyNotFound
	}

	f := families[0]
	return &f, nil
}

func (s *supabaseFamilyStore) CreateFamily(ctx context.Context, f *Family) error {
	payload := Family{
		ID:   f.ID,
		Name: f.Name,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_families", payload, "")
	if err != nil {
		return fmt.Errorf("create family: %w", err)
	}
	return nil
}

func (s *supabaseFamilyStore) UpdateFamily(ctx context.Context, familyID string, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+familyID)
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_families", patch, params)
	if err != nil {
		return fmt.Errorf("update family: %w", err)
	}
	return nil
}

var _ FamilyStore = (*supabaseFamilyStore)(nil)
