package reward

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"odyssey/pkg/game"
)

type Service struct {
	repo *game.Repository
}

func NewService(repo *game.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GrantQuestReward gives coins and XP for completing a quest.
func (s *Service) GrantQuestReward(ctx context.Context, uid string, questID int64, xp int64) error {
	coins := int64(5) // default coins for quest complete

	err := s.grantReward(ctx, uid, "QUEST_COMPLETED", coins, "COINS", map[string]any{"quest_id": questID})
	if err != nil {
		return fmt.Errorf("grant quest reward ledger: %w", err)
	}

	return s.addCoinsToUser(ctx, uid, coins)
}

// GrantDailyReward gives coins for completing a daily turn.
func (s *Service) GrantDailyReward(ctx context.Context, uid string) error {
	coins := int64(1) // default coins for daily turn

	err := s.grantReward(ctx, uid, "DAILY_STREAK", coins, "COINS", nil)
	if err != nil {
		return fmt.Errorf("grant daily reward ledger: %w", err)
	}

	return s.addCoinsToUser(ctx, uid, coins)
}

// GetLedger returns the user's reward history.
func (s *Service) GetLedger(ctx context.Context, uid string) ([]game.RewardLedger, error) {
	return s.repo.RewardLedgers.ListByUser(ctx, uid)
}

func (s *Service) grantReward(ctx context.Context, uid, source string, amount int64, rewardType string, metadata map[string]any) error {
	var metaStr *string
	if metadata != nil {
		b, err := json.Marshal(metadata)
		if err == nil {
			str := string(b)
			metaStr = &str
		}
	}

	ledger := &game.RewardLedger{
		ID:         uuid.New().String(),
		UserID:     uid,
		Source:     source,
		Amount:     amount,
		RewardType: rewardType,
		Metadata:   metaStr,
		CreatedAt:  time.Now().UTC(),
	}

	return s.repo.RewardLedgers.CreateLedger(ctx, ledger)
}

func (s *Service) addCoinsToUser(ctx context.Context, uid string, coins int64) error {
	player, err := s.repo.Users.GetUser(ctx, uid)
	if err != nil {
		return fmt.Errorf("get user for reward: %w", err)
	}

	patch := map[string]any{
		"coins":   player.Coins + coins,
		"version": player.Version + 1,
	}

	ok, err := s.repo.Users.UpdateUserIfMatch(ctx, uid, player.Version, patch)
	if err != nil {
		return fmt.Errorf("update user coins: %w", err)
	}
	if !ok {
		// simple retry logic can be added here if needed, but erroring is fine for prototype
		return fmt.Errorf("concurrent modification when updating coins")
	}

	return nil
}
