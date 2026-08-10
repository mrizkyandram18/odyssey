package creative

import (
	"context"
	"errors"
	"fmt"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/events"
)

var (
	ErrQuestNotFound     = errors.New("quest not found")
	ErrQuestNotActive    = errors.New("quest is not active")
	ErrChallengeNotFound = errors.New("challenge not found")
	ErrChallengeDone     = errors.New("challenge already completed")
	ErrInvalidKind       = errors.New("invalid submission kind")
	ErrContentTooShort   = errors.New("content is too short")
	ErrNotFound          = errors.New("not found")
)

// CreativeService handles creative quest submissions.
type CreativeService struct {
	submissions game.CreativeSubmissionStore
	quests      game.QuestStore
	publisher   events.Publisher
}

// NewCreativeService constructs a CreativeService.
func NewCreativeService(submissions game.CreativeSubmissionStore, quests game.QuestStore) *CreativeService {
	return &CreativeService{submissions: submissions, quests: quests, publisher: events.NopPublisher{}}
}

// NewCreativeServiceWithPublisher constructs a CreativeService with an event publisher.
func NewCreativeServiceWithPublisher(submissions game.CreativeSubmissionStore, quests game.QuestStore, publisher events.Publisher) *CreativeService {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	return &CreativeService{submissions: submissions, quests: quests, publisher: publisher}
}

// Submit creates a new creative submission.
// Validation:
//   - quest must be ACTIVE, or DONE (post-quest CREATE_MEMORY form)
//   - while ACTIVE, challenge must not already be DONE
//   - when quest is DONE, challenge may already be DONE (memory capture)
//   - submission kind must be valid
//   - content must be non-empty (minimal content)
func (s *CreativeService) Submit(ctx context.Context, sub *game.Submission) (*game.Submission, error) {
	q, err := s.quests.GetQuest(ctx, sub.QuestID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, ErrQuestNotFound
		}
		return nil, fmt.Errorf("get quest: %w", err)
	}
	// ACTIVE = in-progress creative challenge; DONE = post-completion memory.
	if q.Status != "ACTIVE" && q.Status != "DONE" {
		return nil, ErrQuestNotActive
	}

	challenges, err := s.quests.GetChallenges(ctx, sub.QuestID)
	if err != nil {
		return nil, fmt.Errorf("get challenges: %w", err)
	}
	found := false
	for _, c := range challenges {
		if c.ID == sub.ChallengeID {
			found = true
			// Only block already-done challenges while the quest is still ACTIVE.
			// After quest completion the UI opens CREATE_MEMORY against the last
			// challenge, which is already DONE.
			if q.Status == "ACTIVE" && c.Status == "DONE" {
				return nil, ErrChallengeDone
			}
			break
		}
	}
	if !found {
		return nil, ErrChallengeNotFound
	}

	if !isValidKind(sub.Kind) {
		return nil, ErrInvalidKind
	}
	if len(sub.Content) < 1 {
		return nil, ErrContentTooShort
	}

	if sub.Kind == game.SubmissionDrawing {
		if err := ValidateSVG(sub.Content); err != nil {
			return nil, err
		}
	}

	sub.Status = game.SubmissionStatusPending
	sub.CreatedAt = time.Now().UTC()
	sub.UpdatedAt = sub.CreatedAt

	created, err := s.submissions.CreateSubmission(ctx, sub)
	if err != nil {
		return nil, fmt.Errorf("create submission: %w", err)
	}

	s.publisher.Publish(ctx, events.CreativeSubmissionEvent{
		UID:         sub.AuthorUID,
		CrewID:      sub.CrewID,
		QuestID:     sub.QuestID,
		ChallengeID: sub.ChallengeID,
		Kind:        string(sub.Kind),
	})

	return created, nil
}

// ListByQuest returns all submissions for a given quest.
func (s *CreativeService) ListByQuest(ctx context.Context, questID int64) ([]game.Submission, error) {
	subs, err := s.submissions.ListByQuest(ctx, questID)
	if err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}
	return subs, nil
}

// ListByCrew returns all submissions for a given crew.
func (s *CreativeService) ListByCrew(ctx context.Context, crewID string) ([]game.Submission, error) {
	subs, err := s.submissions.ListByCrew(ctx, crewID)
	if err != nil {
		return nil, fmt.Errorf("list submissions by crew: %w", err)
	}
	return subs, nil
}

// GetSubmission returns a single submission by ID.
func (s *CreativeService) GetSubmission(ctx context.Context, submissionID int64) (*game.Submission, error) {
	sub, err := s.submissions.GetSubmission(ctx, submissionID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get submission: %w", err)
	}
	return sub, nil
}

// Approve marks a submission as APPROVED.
func (s *CreativeService) Approve(ctx context.Context, submissionID int64, reviewerUID string) (*game.Submission, error) {
	now := time.Now().UTC()
	patch := map[string]any{
		"status":      string(game.SubmissionStatusApproved),
		"reviewed_by": reviewerUID,
		"reviewed_at": now,
		"updated_at":  now,
	}
	if err := s.submissions.UpdateSubmission(ctx, submissionID, patch); err != nil {
		return nil, fmt.Errorf("approve submission: %w", err)
	}
	return s.GetSubmission(ctx, submissionID)
}

// Reject marks a submission as REJECTED with an optional reason.
func (s *CreativeService) Reject(ctx context.Context, submissionID int64, reviewerUID string, reason string) (*game.Submission, error) {
	now := time.Now().UTC()
	patch := map[string]any{
		"status":           string(game.SubmissionStatusRejected),
		"reviewed_by":      reviewerUID,
		"reviewed_at":      now,
		"rejection_reason": reason,
		"updated_at":       now,
	}
	if err := s.submissions.UpdateSubmission(ctx, submissionID, patch); err != nil {
		return nil, fmt.Errorf("reject submission: %w", err)
	}
	return s.GetSubmission(ctx, submissionID)
}

// isValidKind reports whether kind is a recognized submission kind.
func isValidKind(kind game.SubmissionKind) bool {
	switch kind {
	case game.SubmissionStory, game.SubmissionComic, game.SubmissionPhoto, game.SubmissionVideo, game.SubmissionDrawing:
		return true
	}
	return false
}
