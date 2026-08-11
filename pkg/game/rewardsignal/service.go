package rewardsignal

import (
	"context"
	"log/slog"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/events"
)

type Service struct {
	repo *game.Repository
}

func NewService(repo *game.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// EvaluateSignal processes an achievement and emits a signal.
// This is best-effort and isolates failures to avoid crashing the caller.
func (s *Service) EvaluateSignal(ctx context.Context, uid, achievementCode string) {
	signal := &game.RewardSignal{
		UID:             uid,
		AchievementCode: achievementCode,
		IssuedAt:        time.Now().UTC(),
		Consumed:        false,
	}

	err := s.repo.RewardSignals.SaveSignal(ctx, signal)
	if err != nil {
		// Log the failure with context but NO PII.
		// We do not return the error because we don't want to fail the achievement transaction.
		slog.Error("failed to emit reward signal",
			"achievement_code", achievementCode,
			"error", err.Error(),
		)
	}
}

// AchievementEarnedHandler listens for achievements and triggers signal evaluation.
type AchievementEarnedHandler struct {
	svc *Service
}

func NewAchievementEarnedHandler(svc *Service) *AchievementEarnedHandler {
	return &AchievementEarnedHandler{svc: svc}
}

func (h *AchievementEarnedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.AchievementEarnedEvent)
	if !ok {
		return nil
	}
	
	// Best-effort signal emission. We do not return errors to the dispatcher.
	h.svc.EvaluateSignal(ctx, e.UID, e.AchievementCode)
	
	return nil
}
