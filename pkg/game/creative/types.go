package creative

import (
	"time"

	"odyssey/pkg/game"
)

// SubmissionView is the API-safe view of a Submission.
type SubmissionView struct {
	ID              int64
	QuestID         int64
	ChallengeID     int64
	CrewID          string
	AuthorUID       string
	Kind            game.SubmissionKind
	Content         string
	Status          game.SubmissionStatus
	ReviewedBy      string
	ReviewedAt      *time.Time
	RejectionReason string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// SubmitRequest is the incoming payload for creating a submission.
type SubmitRequest struct {
	QuestID     int64               `json:"quest_id"`
	ChallengeID int64               `json:"challenge_id"`
	Kind        game.SubmissionKind `json:"kind"`
	Content     string              `json:"content"`
}

// ListByQuestResult holds the result of a ListByQuest call.
type ListByQuestResult struct {
	Submissions []SubmissionView
}

// HomeCreativeSummary holds creative data for the home screen.
type HomeCreativeSummary struct {
	PendingReviewCount int             `json:"pending_review_count"`
	LastSubmission     *SubmissionView `json:"last_submission,omitempty"`
}
