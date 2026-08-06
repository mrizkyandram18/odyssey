package social

import (
	"context"
	"fmt"
	"odyssey/pkg/game"
)

// ReactionService coordinates reaction creation and listing.
type ReactionService struct {
	store game.ReactionStore
}

// NewReactionService returns a new ReactionService.
func NewReactionService(store game.ReactionStore) *ReactionService {
	return &ReactionService{store: store}
}

// AddReaction adds a new reaction.
func (s *ReactionService) AddReaction(ctx context.Context, creatorID string, targetUserID string, questID *string, emojiCode string) (*game.Reaction, error) {
	if creatorID == "" || targetUserID == "" || emojiCode == "" {
		return nil, fmt.Errorf("invalid reaction payload")
	}

	r := &game.Reaction{
		CreatorID:    creatorID,
		TargetUserID: targetUserID,
		QuestID:      questID,
		EmojiCode:    emojiCode,
	}

	return s.store.CreateReaction(ctx, r)
}

// ListReactionsForTarget lists all reactions meant for a given user.
func (s *ReactionService) ListReactionsForTarget(ctx context.Context, targetUserID string) ([]game.Reaction, error) {
	if targetUserID == "" {
		return nil, fmt.Errorf("target user ID required")
	}
	return s.store.GetReactionsForTarget(ctx, targetUserID)
}
