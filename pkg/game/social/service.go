package social

import (
	"context"
	"fmt"

	"odyssey/pkg/game"
)

// ReactionService coordinates reaction creation and listing.
type ReactionService struct {
	store     game.ReactionStore
	creatives game.CreativeSubmissionStore
	items     game.CreativeStore
	missions    game.QuestStore
}

// NewReactionService returns a new ReactionService.
func NewReactionService(store game.ReactionStore, creatives game.CreativeSubmissionStore, missions game.QuestStore) *ReactionService {
	return &ReactionService{store: store, creatives: creatives, missions: missions}
}

// NewReactionServiceWithItems includes creative_items for TEXT_BOARD reactions.
func NewReactionServiceWithItems(store game.ReactionStore, creatives game.CreativeSubmissionStore, items game.CreativeStore, missions game.QuestStore) *ReactionService {
	return &ReactionService{store: store, creatives: creatives, items: items, missions: missions}
}

// AddReaction adds or updates a reaction idempotently.
func (s *ReactionService) AddReaction(ctx context.Context, crewID, actorUID string, targetType string, targetID int64, reactionType string) (*game.Reaction, error) {
	if crewID == "" || actorUID == "" || targetType == "" || targetID == 0 || reactionType == "" {
		return nil, fmt.Errorf("invalid reaction payload")
	}

	// 1. Validate Target and verify it belongs to the same crew
	switch targetType {
	case game.ReactionTargetJournal:
		// JOURNAL maps to quest-bound creative submissions.
		sub, err := s.creatives.GetSubmission(ctx, targetID)
		if err != nil {
			return nil, fmt.Errorf("target not found")
		}
		if sub.FamilyID != crewID {
			return nil, fmt.Errorf("cross-crew reaction not allowed")
		}
	case game.ReactionTargetQuest:
		quest, err := s.missions.GetQuest(ctx, targetID)
		if err != nil {
			return nil, fmt.Errorf("target not found")
		}
		if quest.FamilyID != crewID {
			return nil, fmt.Errorf("cross-crew reaction not allowed")
		}
	case game.ReactionTargetTextBoard:
		if s.items == nil {
			return nil, fmt.Errorf("target not found")
		}
		item, err := s.items.GetCreativeItem(ctx, targetID)
		if err != nil {
			return nil, fmt.Errorf("target not found")
		}
		if item.FamilyID != crewID || item.Kind != game.KindSharedText {
			return nil, fmt.Errorf("cross-crew reaction not allowed")
		}
	default:
		return nil, fmt.Errorf("invalid target type")
	}

	// 2. Validate Reaction Type
	switch reactionType {
	case "HEART", "CLAP", "STAR":
		// valid
	default:
		return nil, fmt.Errorf("invalid reaction type")
	}

	r := &game.Reaction{
		FamilyID:       crewID,
		TargetType:   targetType,
		TargetID:     targetID,
		ActorUID:     actorUID,
		ReactionType: reactionType,
	}

	return s.store.UpsertReaction(ctx, r)
}

// ListReactionsForTarget lists all reactions meant for a given target.
func (s *ReactionService) ListReactionsForTarget(ctx context.Context, crewID, targetType string, targetID int64) ([]game.Reaction, error) {
	if crewID == "" || targetType == "" || targetID == 0 {
		return nil, fmt.Errorf("invalid target payload")
	}
	return s.store.ListReactionsForTarget(ctx, crewID, targetType, targetID)
}
