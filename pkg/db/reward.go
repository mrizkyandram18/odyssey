package db

import (
	"context"
	"encoding/json"
	"fmt"

	"odyssey/pkg/game"
)

type RewardLedgerStore struct {
	client SupabaseClient
}

func NewRewardLedgerStore(client SupabaseClient) *RewardLedgerStore {
	return &RewardLedgerStore{client: client}
}

func (s *RewardLedgerStore) CreateLedger(ctx context.Context, ledger *game.RewardLedger) error {
	payload := map[string]any{
		"id":          ledger.ID,
		"user_id":     ledger.UserID,
		"source":      ledger.Source,
		"amount":      ledger.Amount,
		"reward_type": ledger.RewardType,
		"created_at":  ledger.CreatedAt,
	}
	if ledger.Metadata != nil {
		payload["metadata"] = ledger.Metadata
	}

	_, err := s.client.Mutate(ctx, "POST", "odyssey_reward_ledgers", payload, "")
	if err != nil {
		return fmt.Errorf("create reward ledger: %w", err)
	}
	return nil
}

func (s *RewardLedgerStore) ListByUser(ctx context.Context, userID string) ([]game.RewardLedger, error) {
	params := fmt.Sprintf("user_id=eq.%s&order=created_at.desc", userID)
	raw, err := s.client.Get(ctx, "odyssey_reward_ledgers", params)
	if err != nil {
		return nil, fmt.Errorf("list reward ledgers: %w", err)
	}

	var ledgers []RewardLedger
	if err := json.Unmarshal(raw, &ledgers); err != nil {
		return nil, fmt.Errorf("unmarshal reward ledgers: %w", err)
	}

	result := make([]game.RewardLedger, len(ledgers))
	for i, l := range ledgers {
		result[i] = game.RewardLedger{
			ID:         l.ID,
			UserID:     l.UserID,
			Source:     l.Source,
			Amount:     l.Amount,
			RewardType: l.RewardType,
			Metadata:   l.Metadata,
			CreatedAt:  l.CreatedAt,
		}
	}
	return result, nil
}
