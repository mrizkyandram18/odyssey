package db

import (
	"context"
	"fmt"

	"odyssey/pkg/game"
)

type RewardSignalStore struct {
	client SupabaseClient
}

func NewRewardSignalStore(client SupabaseClient) *RewardSignalStore {
	return &RewardSignalStore{client: client}
}

// SaveSignal inserts a new reward signal into the database.
// It is designed to be idempotent; if the signal already exists (uid, achievement_code),
// it will ignore the conflict to avoid failing the achievement pipeline.
func (s *RewardSignalStore) SaveSignal(ctx context.Context, signal *game.RewardSignal) error {
	payload := map[string]any{
		"uid":              signal.UID,
		"achievement_code": signal.AchievementCode,
		"issued_at":        signal.IssuedAt,
		"consumed":         signal.Consumed,
	}

	// We use "Prefer: resolution=ignore-duplicates" header in PostgREST to silently ignore
	// inserts that conflict on the primary key (uid, achievement_code).
	_, err := s.client.Mutate(ctx, "POST", "odyssey_reward_signals", payload, "resolution=ignore-duplicates")
	if err != nil {
		return fmt.Errorf("create reward signal: %w", err)
	}
	return nil
}
