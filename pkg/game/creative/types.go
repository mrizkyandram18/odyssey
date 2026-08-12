package creative

import (
	"time"

	"odyssey/pkg/game"
)

// SubmissionView is the API-safe view of a Submission.
// JSON tags are snake_case to match the frontend CreativeSubmission contract.
type SubmissionView struct {
	ID              int64                 `json:"id"`
	MissionID         int64                 `json:"mission_id"`
	ExerciseID     int64                 `json:"exercise_id"`
	FamilyID          string                `json:"family_id"`
	AuthorUID       string                `json:"author_uid"`
	Kind            game.SubmissionKind   `json:"kind"`
	Content         string                `json:"content"`
	Status          game.SubmissionStatus `json:"status"`
	ReviewedBy      string                `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time            `json:"reviewed_at,omitempty"`
	RejectionReason string                `json:"rejection_reason,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// SubmitRequest is the incoming payload for creating a submission.
type SubmitRequest struct {
	MissionID     int64               `json:"mission_id"`
	ExerciseID int64               `json:"exercise_id"`
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
