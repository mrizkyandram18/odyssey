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

func (s *reactionStore) UpsertReaction(ctx context.Context, r *game.Reaction) (*game.Reaction, error) {
	// Upsert via POST with on_conflict.
	// Omit id and created_at so PostgREST assigns identity + DB default timestamps.
	// Sending id=0 / zero time creates invalid rows (same class of bug as creative submit).
	payload := map[string]any{
		"crew_id":       r.CrewID,
		"target_type":   r.TargetType,
		"target_id":     r.TargetID,
		"actor_uid":     r.ActorUID,
		"reaction_type": r.ReactionType,
	}
	prefer := "return=representation,resolution=merge-duplicates"
	params := "on_conflict=crew_id,target_type,target_id,actor_uid"

	raw, err := s.client.MutateAtomic(ctx, "POST", "odyssey_reactions", payload, params, prefer)
	if err != nil {
		return nil, fmt.Errorf("upsert reaction: %w", err)
	}

	var resp []game.Reaction
	if err := json.Unmarshal(raw, &resp); err != nil {
		return nil, fmt.Errorf("parse reaction response: %w", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no reaction returned after upsert")
	}

	return &resp[0], nil
}

func (s *reactionStore) ListReactionsForTarget(ctx context.Context, crewID, targetType string, targetID int64) ([]game.Reaction, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	v.Set("target_type", "eq."+targetType)
	v.Set("target_id", fmt.Sprintf("eq.%d", targetID))
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
