package dailyturn

import (
	"context"
	"errors"
	"fmt"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/observability"
)

// ConsumeDailyTurnResult is returned after a daily turn is consumed.
type ConsumeDailyTurnResult struct {
	Turn       game.DailyTurn `json:"turn"`
	XP         int64          `json:"xp"`
	NewLevel   int            `json:"new_level"`
	LevelUp    bool           `json:"level_up"`
	StreakDays int            `json:"streak_days"`
}

// DailyTurnAPIHandler orchestrates daily turn consumption together with
// progression side-effects (XP award, streak computation). It composes
// DailyTurnService and ProgressionService.
type DailyTurnAPIHandler struct {
	dts       *DailyTurnService
	prog      *progression.ProgressionService
	publisher events.Publisher
	balance   *balance.Service
	metrics   *observability.Metrics
	logger    *observability.Logger
	activity  game.ActivityStore
}

// SetMetrics attaches an optional metrics sink. Safe to call with nil.
func (h *DailyTurnAPIHandler) SetMetrics(m *observability.Metrics) {
	h.metrics = m
}

// SetLogger attaches an optional structured logger for service-call logs.
// Safe to call with nil.
func (h *DailyTurnAPIHandler) SetLogger(l *observability.Logger) {
	h.logger = l
}

// SetActivityStore attaches an ActivityStore to record daily turns as generic activity.
func (h *DailyTurnAPIHandler) SetActivityStore(s game.ActivityStore) {
	h.activity = s
}

// NewDailyTurnAPIHandler constructs a DailyTurnAPIHandler from its collaborators.
func NewDailyTurnAPIHandler(dts *DailyTurnService, prog *progression.ProgressionService) *DailyTurnAPIHandler {
	return &DailyTurnAPIHandler{dts: dts, prog: prog, publisher: events.NopPublisher{}}
}

func NewDailyTurnAPIHandlerWithPublisher(dts *DailyTurnService, prog *progression.ProgressionService, publisher events.Publisher, balanceSvc ...*balance.Service) *DailyTurnAPIHandler {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	s := &DailyTurnAPIHandler{dts: dts, prog: prog, publisher: publisher}
	if len(balanceSvc) > 0 && balanceSvc[0] != nil {
		s.balance = balanceSvc[0]
	}
	return s
}

// List delegates to DailyTurnService.
func (h *DailyTurnAPIHandler) List(ctx context.Context, uid string) ([]game.DailyTurn, error) {
	return h.dts.List(ctx, uid)
}

// Consume marks today's daily turn as completed and awards XP.
// It returns ErrNoTurnsRemaining if the user has already completed today's turn.
// XP awarded is controlled by DailyTurnService.Config().XP, with optional
// balance overrides.
func (h *DailyTurnAPIHandler) Consume(ctx context.Context, uid string, questSlug string) (result *ConsumeDailyTurnResult, err error) {
	start := time.Now()
	defer func() {
		if h.logger == nil {
			return
		}
		outcome := "ok"
		idempotencySkip := false
		errMsg := ""
		if err != nil {
			errMsg = err.Error()
			if errors.Is(err, ErrNoTurnsRemaining) {
				outcome = "duplicate"
				idempotencySkip = true
			} else {
				outcome = "error"
			}
		}
		h.logger.LogServiceCall(observability.ServiceCallFields{
			RequestID:       observability.RequestIDFromContext(ctx),
			Operation:       "dailyturn.consume",
			EntityID:        uid,
			Duration:        time.Since(start),
			Outcome:         outcome,
			IdempotencySkip: idempotencySkip,
			Error:           errMsg,
		})
	}()

	date := h.dts.TodayDate()
	turn, err := h.dts.ConsumeDailyTurn(ctx, uid, date, questSlug)
	if err != nil {
		return nil, err
	}

	xpAmount := int64(h.dts.cfg.XP)
	if h.balance != nil {
		xpAmount = h.balance.OverrideDailyTurnXP(xpAmount)
	}
	player, levelUp, err := h.prog.AwardXP(ctx, uid, xpAmount)
	if err != nil {
		_ = h.dts.UpdateDailyTurn(ctx, turn.ID, map[string]any{"completed": false})
		return nil, fmt.Errorf("award daily turn xp: %w", err)
	}

	streak := 0
	if h.activity != nil {
		_, _ = h.activity.RecordActivity(ctx, &game.DailyActivity{
			UserID:       uid,
			ActivityDate: date,
			ActivityType: "daily_turn",
		})
		s, err := h.activity.GetStreak(ctx, uid)
		if err == nil {
			streak = s
		}
	} else {
		s, err := h.dts.ComputeStreak(ctx, uid)
		if err == nil {
			streak = s
		}
	}

	h.publisher.Publish(ctx, events.DailyTurnCompletedEvent{
		UID:        uid,
		StreakDays: streak,
	})

	return &ConsumeDailyTurnResult{
		Turn:       *turn,
		XP:         xpAmount,
		NewLevel:   player.Level,
		LevelUp:    levelUp,
		StreakDays: streak,
	}, nil
}
