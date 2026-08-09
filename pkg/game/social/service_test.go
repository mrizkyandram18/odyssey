package social

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
)

type mockReactionStore struct {
	upserted *game.Reaction
	err      error
}

func (m *mockReactionStore) UpsertReaction(ctx context.Context, r *game.Reaction) (*game.Reaction, error) {
	if m.err != nil {
		return nil, m.err
	}
	m.upserted = r
	r.ID = 1
	r.CreatedAt = time.Now()
	return r, nil
}

func (m *mockReactionStore) ListReactionsForTarget(ctx context.Context, crewID, targetType string, targetID int64) ([]game.Reaction, error) {
	return nil, nil
}

type mockCreativeStore struct {
	sub *game.Submission
	err error
}

func (m *mockCreativeStore) GetSubmission(ctx context.Context, id int64) (*game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.sub, nil
}
func (m *mockCreativeStore) CreateSubmission(ctx context.Context, s *game.Submission) (*game.Submission, error) { return nil, nil }
func (m *mockCreativeStore) ListByQuest(ctx context.Context, questID int64) ([]game.Submission, error) { return nil, nil }
func (m *mockCreativeStore) ListByCrew(ctx context.Context, crewID string) ([]game.Submission, error) { return nil, nil }
func (m *mockCreativeStore) UpdateSubmission(ctx context.Context, submissionID int64, patch map[string]any) error { return nil }

type mockQuestStore struct {
	quest *game.Quest
	err   error
}

func (m *mockQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.quest, nil
}
func (m *mockQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) { return nil, nil }
func (m *mockQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) { return nil, nil }
func (m *mockQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error { return nil }
func (m *mockQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) { return false, nil }
func (m *mockQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) { return nil, nil }
func (m *mockQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) { return nil, nil }
func (m *mockQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error { return nil }

func TestReactionService_AddReaction_Valid(t *testing.T) {
	rs := &mockReactionStore{}
	cs := &mockCreativeStore{sub: &game.Submission{ID: 10, CrewID: "crew-A"}}
	qs := &mockQuestStore{}
	
	svc := NewReactionService(rs, cs, qs)
	
	r, err := svc.AddReaction(context.Background(), "crew-A", "user-1", "JOURNAL", 10, "STAR")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if r.TargetType != "JOURNAL" || r.ReactionType != "STAR" {
		t.Errorf("wrong reaction returned: %+v", r)
	}
}

func TestReactionService_AddReaction_CrossCrewSpoof(t *testing.T) {
	rs := &mockReactionStore{}
	// Target item belongs to crew-B
	cs := &mockCreativeStore{sub: &game.Submission{ID: 10, CrewID: "crew-B"}}
	qs := &mockQuestStore{}
	
	svc := NewReactionService(rs, cs, qs)
	
	// Actor in crew-A tries to react to crew-B's item
	_, err := svc.AddReaction(context.Background(), "crew-A", "user-1", "JOURNAL", 10, "STAR")
	if err == nil {
		t.Fatalf("Expected error for cross-crew spoof, got nil")
	}
	if err.Error() != "cross-crew reaction not allowed" {
		t.Errorf("unexpected error msg: %v", err)
	}
}

func TestReactionService_AddReaction_InvalidTargetType(t *testing.T) {
	rs := &mockReactionStore{}
	cs := &mockCreativeStore{}
	qs := &mockQuestStore{}
	
	svc := NewReactionService(rs, cs, qs)
	
	_, err := svc.AddReaction(context.Background(), "crew-A", "user-1", "HACKER", 10, "STAR")
	if err == nil || err.Error() != "invalid target type" {
		t.Fatalf("expected 'invalid target type' err, got %v", err)
	}
}

func TestReactionService_AddReaction_InvalidReactionType(t *testing.T) {
	rs := &mockReactionStore{}
	cs := &mockCreativeStore{sub: &game.Submission{ID: 10, CrewID: "crew-A"}}
	qs := &mockQuestStore{}
	
	svc := NewReactionService(rs, cs, qs)
	
	// Send arbitrary string
	_, err := svc.AddReaction(context.Background(), "crew-A", "user-1", "JOURNAL", 10, "POOP")
	if err == nil || err.Error() != "invalid reaction type" {
		t.Fatalf("expected 'invalid reaction type' err, got %v", err)
	}
}
