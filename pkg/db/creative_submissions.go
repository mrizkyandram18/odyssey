package db

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"odyssey/pkg/game"
)

// supabaseCreativeSubmissionStore implements game.CreativeSubmissionStore via Supabase.
type supabaseCreativeSubmissionStore struct {
	client SupabaseClient
}

// NewCreativeSubmissionStore constructs a game.CreativeSubmissionStore backed by Supabase.
func NewCreativeSubmissionStore(client SupabaseClient) game.CreativeSubmissionStore {
	return &supabaseCreativeSubmissionStore{client: client}
}

func (s *supabaseCreativeSubmissionStore) CreateSubmission(ctx context.Context, sub *game.Submission) (*game.Submission, error) {
	// Omit id so PostgREST assigns the identity column (explicit id=0 creates
	// invalid zero-id rows and can break subsequent serial inserts).
	payload := map[string]any{
		"quest_id":     sub.QuestID,
		"challenge_id": sub.ChallengeID,
		"crew_id":      sub.CrewID,
		"author_uid":   sub.AuthorUID,
		"kind":         string(sub.Kind),
		"content":      sub.Content,
		"status":       string(sub.Status),
	}
	raw, err := s.client.Mutate(ctx, "POST", "odyssey_creative_submissions", payload, "return=representation")
	if err != nil {
		return nil, fmt.Errorf("create creative submission: %w", err)
	}

	var subs []CreativeSubmission
	if err := json.Unmarshal(raw, &subs); err != nil {
		return nil, fmt.Errorf("parse created creative submission: %w", err)
	}
	if len(subs) == 0 {
		return sub, nil
	}
	return mapCreativeSubmission(subs[0]), nil
}

func (s *supabaseCreativeSubmissionStore) ListByQuest(ctx context.Context, questID int64) ([]game.Submission, error) {
	v := url.Values{}
	v.Set("quest_id", "eq."+strconv.FormatInt(questID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_creative_submissions", params)
	if err != nil {
		return nil, fmt.Errorf("list creative submissions: %w", err)
	}

	var subs []CreativeSubmission
	if err := json.Unmarshal(raw, &subs); err != nil {
		return nil, fmt.Errorf("parse creative submissions: %w", err)
	}

	result := make([]game.Submission, 0, len(subs))
	for _, sub := range subs {
		result = append(result, *mapCreativeSubmission(sub))
	}
	return result, nil
}

func (s *supabaseCreativeSubmissionStore) ListByCrew(ctx context.Context, crewID string) ([]game.Submission, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_creative_submissions", params)
	if err != nil {
		return nil, fmt.Errorf("list crew creative submissions: %w", err)
	}

	var subs []CreativeSubmission
	if err := json.Unmarshal(raw, &subs); err != nil {
		return nil, fmt.Errorf("parse crew creative submissions: %w", err)
	}

	result := make([]game.Submission, 0, len(subs))
	for _, sub := range subs {
		result = append(result, *mapCreativeSubmission(sub))
	}
	return result, nil
}

func (s *supabaseCreativeSubmissionStore) ListByCrewAndKind(ctx context.Context, crewID, kind string) ([]game.Submission, error) {
	v := url.Values{}
	v.Set("crew_id", "eq."+crewID)
	v.Set("kind", "eq."+kind)
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_creative_submissions", params)
	if err != nil {
		return nil, fmt.Errorf("list crew creative submissions by kind: %w", err)
	}

	var subs []CreativeSubmission
	if err := json.Unmarshal(raw, &subs); err != nil {
		return nil, fmt.Errorf("parse crew creative submissions by kind: %w", err)
	}

	result := make([]game.Submission, 0, len(subs))
	for _, sub := range subs {
		result = append(result, *mapCreativeSubmission(sub))
	}
	return result, nil
}

func (s *supabaseCreativeSubmissionStore) GetSubmission(ctx context.Context, submissionID int64) (*game.Submission, error) {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(submissionID, 10))
	params := v.Encode()

	raw, err := s.client.Get(ctx, "odyssey_creative_submissions", params)
	if err != nil {
		return nil, fmt.Errorf("get creative submission: %w", err)
	}

	var subs []CreativeSubmission
	if err := json.Unmarshal(raw, &subs); err != nil {
		return nil, fmt.Errorf("parse creative submission: %w", err)
	}
	if len(subs) == 0 {
		return nil, game.ErrNotFound
	}
	return mapCreativeSubmission(subs[0]), nil
}

func (s *supabaseCreativeSubmissionStore) UpdateSubmission(ctx context.Context, submissionID int64, patch map[string]any) error {
	v := url.Values{}
	v.Set("id", "eq."+strconv.FormatInt(submissionID, 10))
	params := v.Encode()
	if _, ok := patch["updated_at"]; !ok {
		patch["updated_at"] = time.Now().UTC()
	}
	_, err := s.client.Mutate(ctx, "PATCH", "odyssey_creative_submissions", patch, params)
	if err != nil {
		return fmt.Errorf("update creative submission: %w", err)
	}
	return nil
}

func mapCreativeSubmission(sub CreativeSubmission) *game.Submission {
	var reviewedAt *time.Time
	if sub.ReviewedAt != nil {
		reviewedAt = sub.ReviewedAt
	}
	return &game.Submission{
		ID:              sub.ID,
		QuestID:         sub.QuestID,
		ChallengeID:     sub.ChallengeID,
		CrewID:          sub.CrewID,
		AuthorUID:       sub.AuthorUID,
		Kind:            game.SubmissionKind(sub.Kind),
		Content:         sub.Content,
		Status:          game.SubmissionStatus(sub.Status),
		ReviewedBy:      sub.ReviewedBy,
		ReviewedAt:      reviewedAt,
		RejectionReason: sub.RejectionReason,
		CreatedAt:       sub.CreatedAt,
		UpdatedAt:       sub.UpdatedAt,
	}
}

var _ game.CreativeSubmissionStore = (*supabaseCreativeSubmissionStore)(nil)
