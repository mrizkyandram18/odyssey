package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game/balance"
)

type BalanceConfig struct {
	ID        int64           `json:"id"`
	Key       string          `json:"key"`
	Value     json.RawMessage `json:"value"`
	UpdatedBy string          `json:"updated_by"`
	CreatedAt string          `json:"created_at"`
	UpdatedAt string          `json:"updated_at"`
}

type supabaseBalanceStore struct {
	client SupabaseClient
}

func NewBalanceStore(client SupabaseClient) balance.Store {
	return &supabaseBalanceStore{client: client}
}

func (s *supabaseBalanceStore) GetOverride(ctx context.Context, key string) (*balance.Override, error) {
	v := make(url.Values)
	v.Set("key", "eq."+key)
	raw, err := s.client.Get(ctx, "odyssey_balance_configs", v.Encode())
	if err != nil {
		return nil, fmt.Errorf("get balance override %s: %w", key, err)
	}
	var rows []BalanceConfig
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse balance override: %w", err)
	}
	if len(rows) == 0 {
		return nil, balance.ErrConfigNotFound
	}
	return mapBalanceOverride(rows[0]), nil
}

func (s *supabaseBalanceStore) ListOverrides(ctx context.Context) ([]balance.Override, error) {
	raw, err := s.client.Get(ctx, "odyssey_balance_configs", "")
	if err != nil {
		return nil, fmt.Errorf("list balance overrides: %w", err)
	}
	var rows []BalanceConfig
	if err := json.Unmarshal(raw, &rows); err != nil {
		return nil, fmt.Errorf("parse balance overrides: %w", err)
	}
	result := make([]balance.Override, 0, len(rows))
	for i := range rows {
		result = append(result, *mapBalanceOverride(rows[i]))
	}
	return result, nil
}

func mapBalanceOverride(bc BalanceConfig) *balance.Override {
	var val int64
	if err := json.Unmarshal(bc.Value, &val); err != nil {
		val = 0
	}
	return &balance.Override{
		Key:       bc.Key,
		Value:     val,
		UpdatedBy: bc.UpdatedBy,
	}
}

var _ balance.Store = (*supabaseBalanceStore)(nil)
