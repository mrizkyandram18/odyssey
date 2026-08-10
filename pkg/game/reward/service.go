package reward

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"

	"odyssey/pkg/game"
)

// Slice 2.1 earn amounts — fixed; spending is out of scope.
const (
	CoinsPerQuestComplete int64 = 5
	CoinsPerDailyTurn     int64 = 1

	SourceQuestCompleted = "QUEST_COMPLETED"
	SourceDailyStreak    = "DAILY_STREAK"
	RewardTypeCoins      = "COINS"
)

type Service struct {
	repo *game.Repository
}

func NewService(repo *game.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// GrantQuestReward credits +5 coins for completing a quest (once per quest_id).
func (s *Service) GrantQuestReward(ctx context.Context, uid string, questID int64, xp int64) error {
	_ = xp // XP is awarded elsewhere; coins use fixed Slice 2.1 amount.
	coins := CoinsPerQuestComplete

	// Soft idempotency: do not double-credit the same quest completion.
	if granted, err := s.hasQuestCoinGrant(ctx, uid, questID); err != nil {
		return fmt.Errorf("check quest coin grant: %w", err)
	} else if granted {
		return nil
	}

	err := s.grantReward(ctx, uid, SourceQuestCompleted, coins, RewardTypeCoins, map[string]any{"quest_id": questID})
	if err != nil {
		return fmt.Errorf("grant quest reward ledger: %w", err)
	}

	return s.addCoinsToUser(ctx, uid, coins)
}

// GrantDailyReward credits +1 coin for completing a daily turn.
// Call sites must only invoke this once per successful daily consume.
func (s *Service) GrantDailyReward(ctx context.Context, uid string) error {
	coins := CoinsPerDailyTurn

	err := s.grantReward(ctx, uid, SourceDailyStreak, coins, RewardTypeCoins, nil)
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
		return fmt.Errorf("concurrent modification when updating coins")
	}

	return nil
}

func (s *Service) hasQuestCoinGrant(ctx context.Context, uid string, questID int64) (bool, error) {
	ledgers, err := s.repo.RewardLedgers.ListByUser(ctx, uid)
	if err != nil {
		return false, err
	}
	want := strconv.FormatInt(questID, 10)
	for _, l := range ledgers {
		if l.Source != SourceQuestCompleted || l.RewardType != RewardTypeCoins {
			continue
		}
		if l.Metadata == nil {
			continue
		}
		// Metadata may be a JSON object string: {"quest_id":102}
		if metadataHasQuestID(*l.Metadata, want) {
			return true, nil
		}
	}
	return false, nil
}

func metadataHasQuestID(meta, questID string) bool {
	var m map[string]any
	if err := json.Unmarshal([]byte(meta), &m); err == nil {
		switch v := m["quest_id"].(type) {
		case float64:
			return strconv.FormatInt(int64(v), 10) == questID
		case json.Number:
			return v.String() == questID
		case string:
			return v == questID
		}
	}
	// Fallback substring for loosely stored metadata.
	return strings.Contains(meta, `"quest_id":`+questID) ||
		strings.Contains(meta, `"quest_id": `+questID)
}
