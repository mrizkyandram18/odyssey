package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"

	"odyssey/pkg/game"
)

// supabaseQuestStore implements game.QuestStore via Supabase.
type supabaseQuestStore struct {
	client SupabaseClient
}

// NewQuestStore constructs a game.QuestStore backed by Supabase.
func NewQuestStore(client SupabaseClient) game.QuestStore {
	return &supabaseQuestStore{client: client}
}

func (s *supabaseQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(questID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_quests", params)
	if err != nil {
		return nil, fmt.Errorf("get quest: %w", err)
	}

	var quests []QuestInstance
	if err := json.Unmarshal(raw, &quests); err != nil {
		return nil, fmt.Errorf("parse quests: %w", err)
	}
	if len(quests) == 0 {
		return nil, game.ErrNotFound
	}

	q := quests[0]
	return mapQuest(q), nil
}

func (s *supabaseQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	payload := QuestInstance{
		CrewID:       q.CrewID,
		TemplateSlug: q.TemplateSlug,
		Title:        q.Title,
		Status:       q.Status,
		StartedAt:    q.StartedAt,
		CompletedAt:  q.CompletedAt,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_quests", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create quest: %w", err)
	}
	return q, nil
}

func (s *supabaseQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_quests", params)
	if err != nil {
		return nil, fmt.Errorf("list quests: %w", err)
	}

	var quests []QuestInstance
	if err := json.Unmarshal(raw, &quests); err != nil {
		return nil, fmt.Errorf("parse quests: %w", err)
	}

	result := make([]game.Quest, 0, len(quests))
	for _, q := range quests {
		result = append(result, *mapQuest(q))
	}
	return result, nil
}

func (s *supabaseQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(questID, 10))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_quests", patch, params)
	if err != nil {
		return fmt.Errorf("update quest: %w", err)
	}
	return nil
}

func (s *supabaseQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(questID, 10))
	v.Set("status", "eq."+oldStatus)
	params := v.Encode()
	raw, err := s.client.Mutate(ctx, "PATCH", "odyssey_quests", patch, params+"&return=representation")
	if err != nil {
		return false, fmt.Errorf("update quest if match: %w", err)
	}

	var quests []QuestInstance
	if err := json.Unmarshal(raw, &quests); err != nil {
		return false, fmt.Errorf("parse update quest response: %w", err)
	}
	return len(quests) > 0, nil
}

func (s *supabaseQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	v := url.Values{}
	v.Set("quest_id", "eq."+strconv.FormatInt(questID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_challenges", params)
	if err != nil {
		return nil, fmt.Errorf("get challenges: %w", err)
	}

	var challenges []Challenge
	if err := json.Unmarshal(raw, &challenges); err != nil {
		return nil, fmt.Errorf("parse challenges: %w", err)
	}

	result := make([]game.Challenge, 0, len(challenges))
	for _, c := range challenges {
		result = append(result, *mapChallenge(c))
	}
	return result, nil
}

func (s *supabaseQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	payload := Challenge{
		QuestID:     c.QuestID,
		Slug:        c.Slug,
		Description: c.Description,
		Status:      c.Status,
		AssignedTo:  c.AssignedTo,
		CompletedBy: c.CompletedBy,
		CompletedAt: c.CompletedAt,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_challenges", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create challenge: %w", err)
	}
	return c, nil
}

func (s *supabaseQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(challengeID, 10))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_challenges", patch, params)
	if err != nil {
		return fmt.Errorf("update challenge: %w", err)
	}
	return nil
}

func mapQuest(q QuestInstance) *game.Quest {
	return &game.Quest{
		ID:           q.ID,
		CrewID:       q.CrewID,
		TemplateSlug: q.TemplateSlug,
		Title:        q.Title,
		Status:       q.Status,
		StartedAt:    q.StartedAt,
		CompletedAt:  q.CompletedAt,
		CreatedAt:    q.CreatedAt,
	}
}

func mapChallenge(c Challenge) *game.Challenge {
	return &game.Challenge{
		ID:          c.ID,
		QuestID:     c.QuestID,
		Slug:        c.Slug,
		Description: c.Description,
		Status:      c.Status,
		AssignedTo:  c.AssignedTo,
		CompletedBy: c.CompletedBy,
		CompletedAt: c.CompletedAt,
		CreatedAt:   c.CreatedAt,
	}
}

var _ game.QuestStore = (*supabaseQuestStore)(nil)
