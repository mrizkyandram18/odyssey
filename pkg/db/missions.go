package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

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

func (s *supabaseQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Mission, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(questID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_missions", params)
	if err != nil {
		return nil, fmt.Errorf("get quest: %w", err)
	}

	var missions []QuestInstance
	if err := json.Unmarshal(raw, &missions); err != nil {
		return nil, fmt.Errorf("parse missions: %w", err)
	}
	if len(missions) == 0 {
		return nil, game.ErrNotFound
	}

	q := missions[0]
	return mapQuest(q), nil
}

func (s *supabaseQuestStore) CreateQuest(ctx context.Context, q *game.Mission) (*game.Mission, error) {
	payload := QuestInstance{
		FamilyID:       q.FamilyID,
		TemplateSlug: q.TemplateSlug,
		Title:        q.Title,
		Status:       q.Status,
		StartedAt:    q.StartedAt,
		StartedBy:    q.StartedBy,
		CompletedAt:  q.CompletedAt,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_missions", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create quest: %w", err)
	}
	return q, nil
}

func (s *supabaseQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Mission, error) {
	v := url.Values{}
	v.Set("family_id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_missions", params)
	if err != nil {
		return nil, fmt.Errorf("list missions: %w", err)
	}

	var missions []QuestInstance
	if err := json.Unmarshal(raw, &missions); err != nil {
		return nil, fmt.Errorf("parse missions: %w", err)
	}

	result := make([]game.Mission, 0, len(missions))
	for _, q := range missions {
		result = append(result, *mapQuest(q))
	}
	return result, nil
}

func (s *supabaseQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(questID, 10))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_missions", patch, params)
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
	raw, err := s.client.Mutate(ctx, "PATCH", "odyssey_missions", patch, params+"&return=representation")
	if err != nil {
		return false, fmt.Errorf("update quest if match: %w", err)
	}

	var missions []QuestInstance
	if err := json.Unmarshal(raw, &missions); err != nil {
		return false, fmt.Errorf("parse update quest response: %w", err)
	}
	return len(missions) > 0, nil
}

func (s *supabaseQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Exercise, error) {
	v := url.Values{}
	v.Set("mission_id", "eq."+strconv.FormatInt(questID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_exercises", params)
	if err != nil {
		return nil, fmt.Errorf("get exercises: %w", err)
	}

	var exercises []Exercise
	if err := json.Unmarshal(raw, &exercises); err != nil {
		return nil, fmt.Errorf("parse exercises: %w", err)
	}

	result := make([]game.Exercise, 0, len(exercises))
	for _, c := range exercises {
		result = append(result, *mapChallenge(c))
	}
	return result, nil
}

// ListChallengesByQuestIDs fetches exercises for many missions in a single
// Supabase request using a PostgREST in.(...) filter. Mission IDs that have no
// exercises simply contribute no entries.
func (s *supabaseQuestStore) ListChallengesByQuestIDs(ctx context.Context, questIDs []int64) ([]game.Exercise, error) {
	if len(questIDs) == 0 {
		return nil, nil
	}
	v := url.Values{}
	v.Set("mission_id", "in.("+joinInt64(questIDs)+")")
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_exercises", params)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}

	var exercises []Exercise
	if err := json.Unmarshal(raw, &exercises); err != nil {
		return nil, fmt.Errorf("parse exercises: %w", err)
	}

	result := make([]game.Exercise, 0, len(exercises))
	for _, c := range exercises {
		result = append(result, *mapChallenge(c))
	}
	return result, nil
}

func (s *supabaseQuestStore) CreateChallenge(ctx context.Context, c *game.Exercise) (*game.Exercise, error) {
	payload := Exercise{
		MissionID:     c.MissionID,
		Slug:        c.Slug,
		Description: c.Description,
		Status:      c.Status,
		AssignedTo:  c.AssignedTo,
		CompletedBy: c.CompletedBy,
		CompletedAt: c.CompletedAt,
	}
	_, err := s.client.Mutate(ctx, "POST", "odyssey_exercises", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create challenge: %w", err)
	}
	return c, nil
}

func (s *supabaseQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(challengeID, 10))
	params := v.Encode()
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_exercises", patch, params)
	if err != nil {
		return fmt.Errorf("update challenge: %w", err)
	}
	return nil
}

func (s *supabaseQuestStore) UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(challengeID, 10))
	v.Set("status", "eq."+oldStatus)
	params := v.Encode()
	raw, err := s.client.Mutate(ctx, "PATCH", "odyssey_exercises", patch, params+"&return=representation")
	if err != nil {
		return false, fmt.Errorf("update challenge if match: %w", err)
	}

	var exercises []Exercise
	if err := json.Unmarshal(raw, &exercises); err != nil {
		return false, fmt.Errorf("parse update challenge response: %w", err)
	}
	return len(exercises) > 0, nil
}

func mapQuest(q QuestInstance) *game.Mission {
	return &game.Mission{
		ID:           q.ID,
		FamilyID:       q.FamilyID,
		TemplateSlug: q.TemplateSlug,
		Title:        q.Title,
		Status:       q.Status,
		StartedAt:    q.StartedAt,
		StartedBy:    q.StartedBy,
		CompletedAt:  q.CompletedAt,
		CreatedAt:    q.CreatedAt,
	}
}

func mapChallenge(c Exercise) *game.Exercise {
	return &game.Exercise{
		ID:          c.ID,
		MissionID:     c.MissionID,
		Slug:        c.Slug,
		Description: c.Description,
		Status:      c.Status,
		AssignedTo:  c.AssignedTo,
		CompletedBy: c.CompletedBy,
		CompletedAt: c.CompletedAt,
		CreatedAt:   c.CreatedAt,
	}
}

func joinInt64(ids []int64) string {
	var sb strings.Builder
	for i, id := range ids {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(strconv.FormatInt(id, 10))
	}
	return sb.String()
}

var _ game.QuestStore = (*supabaseQuestStore)(nil)
