package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

type reactionStore struct {
	client SupabaseClient
}

func NewReactionStore(client SupabaseClient) game.ReactionStore {
	return &reactionStore{client: client}
}

func (s *reactionStore) CreateReaction(ctx context.Context, r *game.Reaction) (*game.Reaction, error) {
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_reactions", r, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("insert reaction: %w", err)
	}

	var resp []game.Reaction
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse reaction response: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no reaction returned after insert")
	}

	return &resp[0], nil
}

func (s *reactionStore) GetReactionsForTarget(ctx context.Context, targetUserID string) ([]game.Reaction, error) {
	v := url.Values{}
	v.Set("target_user_id", "eq."+targetUserID)
	v.Set("order", "created_at.desc")
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_reactions", params)
	if err != nil {
		return nil, fmt.Errorf("list reactions: %w", err)
	}

	var resp []game.Reaction
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse reactions: %w", err)
	}

	return resp, nil
}
