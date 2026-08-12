package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"odyssey/pkg/game"
)

// supabaseDailyTurnStore implements game.DailyTurnStore via Supabase.
type supabaseDailyTurnStore struct {
	client SupabaseClient
}

// NewDailyTurnStore constructs a game.DailyTurnStore backed by Supabase.
func NewDailyTurnStore(client SupabaseClient) game.DailyTurnStore {
	return &supabaseDailyTurnStore{client: client}
}

func (s *supabaseDailyTurnStore) CreateDailyTurn(ctx context.Context, dt *game.DailyMission) (*game.DailyMission, error) {
	// Build a create payload without id/created_at. PostgREST rejects explicit
	// id=0 on serial columns, which is what encoding the zero-value DTO did.
	payload := map[string]any{
		"uid":        dt.UID,
		"date":       dt.Date,
		"mission_slug": dt.MissionSlug,
		"completed":  dt.Completed,
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_daily_missions", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create daily turn: %w", err)
	}

	var turns []DailyMission
	if err := json.Unmarshal(raw, &turns); err != nil {
		return nil, fmt.Errorf("parse created daily turn: %w", err)
	}
	if len(turns) == 0 {
		return dt, nil
	}
	return mapDailyTurn(turns[0]), nil
}

func (s *supabaseDailyTurnStore) UpdateDailyTurn(ctx context.Context, turnID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+fmt.Sprintf("%d", turnID))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_daily_missions", patch, params)
	if err != nil {
		return fmt.Errorf("update daily turn: %w", err)
	}
	return nil
}

func (s *supabaseDailyTurnStore) ListDailyTurns(ctx context.Context, uid string) ([]game.DailyMission, error) {
	v := url.Values{}
	v.Set("uid", "eq."+uid)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_daily_missions", params)
	if err != nil {
		return nil, fmt.Errorf("list daily turns: %w", err)
	}

	var turns []DailyMission
	if err := json.Unmarshal(raw, &turns); err != nil {
		return nil, fmt.Errorf("parse daily turns: %w", err)
	}

	result := make([]game.DailyMission, 0, len(turns))
	for _, t := range turns {
		result = append(result, *mapDailyTurn(t))
	}
	return result, nil
}

func mapDailyTurn(t DailyMission) *game.DailyMission {
	return &game.DailyMission{
		ID:        t.ID,
		UID:       t.UID,
		Date:      t.Date,
		MissionSlug: t.MissionSlug,
		Completed: t.Completed,
		CreatedAt: t.CreatedAt,
	}
}

var _ game.DailyTurnStore = (*supabaseDailyTurnStore)(nil)
