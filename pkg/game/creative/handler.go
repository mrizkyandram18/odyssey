package creative

import (
	"context"
	"fmt"

	"odyssey/pkg/game"
)

// CreativeHandler is the interface the API handler depends on.
type CreativeHandler interface {
	Submit(ctx context.Context, uid string, req *game.Submission) (*SubmissionView, error)
	ListByQuest(ctx context.Context, questID int64) ([]SubmissionView, error)
	ListByCrew(ctx context.Context, crewID string) ([]SubmissionView, error)
	ListByCrewAndKind(ctx context.Context, crewID, kind string) ([]SubmissionView, error)
	GetSubmission(ctx context.Context, submissionID int64) (*SubmissionView, error)
	Approve(ctx context.Context, submissionID int64, reviewerUID string) (*SubmissionView, error)
	Reject(ctx context.Context, submissionID int64, reviewerUID string, reason string) (*SubmissionView, error)
}

// CreativeAPIHandler wraps CreativeService to satisfy the API handler interface.
type CreativeAPIHandler struct {
	svc *CreativeService
}

// NewCreativeAPIHandler constructs a CreativeAPIHandler.
func NewCreativeAPIHandler(svc *CreativeService) *CreativeAPIHandler {
	return &CreativeAPIHandler{svc: svc}
}

// Submit delegates to CreativeService.Submit.
func (h *CreativeAPIHandler) Submit(ctx context.Context, uid string, req *game.Submission) (*SubmissionView, error) {
	req.AuthorUID = uid
	sub, err := h.svc.Submit(ctx, req)
	if err != nil {
		return nil, err
	}
	return ToView(sub), nil
}

// ListByQuest delegates to CreativeService.ListByQuest.
func (h *CreativeAPIHandler) ListByQuest(ctx context.Context, questID int64) ([]SubmissionView, error) {
	subs, err := h.svc.ListByQuest(ctx, questID)
	if err != nil {
		return nil, fmt.Errorf("list submissions: %w", err)
	}
	return ToViews(subs), nil
}

// ListByCrew lists all submissions for a crew by iterating their active quests.
func (h *CreativeAPIHandler) ListByCrew(ctx context.Context, crewID string) ([]SubmissionView, error) {
	allSubs, err := h.svc.submissions.ListByCrew(ctx, crewID)
	if err != nil {
		return nil, fmt.Errorf("list crew submissions: %w", err)
	}
	return ToViews(allSubs), nil
}

// ListByCrewAndKind lists all submissions for a crew filtered by kind.
func (h *CreativeAPIHandler) ListByCrewAndKind(ctx context.Context, crewID, kind string) ([]SubmissionView, error) {
	subs, err := h.svc.ListByCrewAndKind(ctx, crewID, kind)
	if err != nil {
		return nil, fmt.Errorf("list crew submissions by kind: %w", err)
	}
	return ToViews(subs), nil
}

// GetSubmission delegates to CreativeService.GetSubmission.
func (h *CreativeAPIHandler) GetSubmission(ctx context.Context, submissionID int64) (*SubmissionView, error) {
	sub, err := h.svc.GetSubmission(ctx, submissionID)
	if err != nil {
		return nil, err
	}
	return ToView(sub), nil
}

// Approve delegates to CreativeService.Approve.
func (h *CreativeAPIHandler) Approve(ctx context.Context, submissionID int64, reviewerUID string) (*SubmissionView, error) {
	sub, err := h.svc.Approve(ctx, submissionID, reviewerUID)
	if err != nil {
		return nil, err
	}
	return ToView(sub), nil
}

// Reject delegates to CreativeService.Reject.
func (h *CreativeAPIHandler) Reject(ctx context.Context, submissionID int64, reviewerUID string, reason string) (*SubmissionView, error) {
	sub, err := h.svc.Reject(ctx, submissionID, reviewerUID, reason)
	if err != nil {
		return nil, err
	}
	return ToView(sub), nil
}

// ToViews converts []game.Submission to []SubmissionView.
func ToViews(subs []game.Submission) []SubmissionView {
	views := make([]SubmissionView, len(subs))
	for i, sub := range subs {
		views[i] = *ToView(&sub)
	}
	return views
}

// ToView converts a single game.Submission to SubmissionView.
func ToView(sub *game.Submission) *SubmissionView {
	if sub == nil {
		return nil
	}
	return &SubmissionView{
		ID:              sub.ID,
		QuestID:         sub.QuestID,
		ChallengeID:     sub.ChallengeID,
		CrewID:          sub.CrewID,
		AuthorUID:       sub.AuthorUID,
		Kind:            sub.Kind,
		Content:         sub.Content,
		Status:          sub.Status,
		ReviewedBy:      sub.ReviewedBy,
		ReviewedAt:      sub.ReviewedAt,
		RejectionReason: sub.RejectionReason,
		CreatedAt:       sub.CreatedAt,
		UpdatedAt:       sub.UpdatedAt,
	}
}
