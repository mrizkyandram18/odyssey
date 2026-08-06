package quest

import (
	"context"
	"fmt"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/balance"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/world"
	"odyssey/pkg/observability"
)

// CompleteChallengeResult is returned by the orchestration layer when a
// family member completes a challenge. It carries the fresh quest state plus
// the progression rewards awarded as a consequence of the action.
type CompleteChallengeResult struct {
	Quest          *QuestWithChallenges `json:"quest"`
	QuestCompleted bool                 `json:"quest_completed"`
	NextAction     string               `json:"next_action,omitempty"`
	XP             int64                `json:"xp"`
	NewLevel       int                  `json:"new_level"`
	LevelUp        bool                 `json:"level_up"`
}

// QuestAPIHandler orchestrates quest lifecycle mutations together with the
// progression side-effects (XP award, level-up, realm progress). It composes
// the focused QuestService and ProgressionService rather than owning that
// logic directly, keeping each concern isolated.
//
// A concrete instance satisfies the api/quests.QuestHandler interface via Go's
// structural typing — no import of the api package is needed here.
type QuestAPIHandler struct {
	qs        *QuestService
	prog      *progression.ProgressionService
	realm     game.RealmProgressStore
	realmCfg  *world.RealmCatalog
	progCfg   *progression.ProgressionConfig
	publisher events.Publisher
	content   ContentGateway
	balance   *balance.Service
	metrics   *observability.Metrics
	logger    *observability.Logger
}

// NewQuestAPIHandler constructs a QuestAPIHandler from its collaborators.
// Pass nil for realmCfg or progCfg to use built-in defaults.
func NewQuestAPIHandler(qs *QuestService, prog *progression.ProgressionService, realm game.RealmProgressStore, realmCfg *world.RealmCatalog, progCfg *progression.ProgressionConfig) *QuestAPIHandler {
	return &QuestAPIHandler{qs: qs, prog: prog, realm: realm, realmCfg: realmCfg, progCfg: progCfg}
}

// SetPublisher attaches an event publisher to the handler.
// When set, a QuestCompleted event is published on quest completion.
func (h *QuestAPIHandler) SetPublisher(p events.Publisher) {
	h.publisher = p
}

// SetContentGateway attaches a content gateway for quest definition lookups.
// When set, the published QuestCompleted event will include the chapter and season.
func (h *QuestAPIHandler) SetContentGateway(cg ContentGateway) {
	h.content = cg
}

// SetBalance attaches a balance service for runtime parameter overrides.
func (h *QuestAPIHandler) SetBalance(b *balance.Service) {
	h.balance = b
}

// SetMetrics attaches an optional metrics sink for quest/realm telemetry.
// Safe to call with nil.
func (h *QuestAPIHandler) SetMetrics(m *observability.Metrics) {
	h.metrics = m
}

// SetLogger attaches an optional structured logger for service-call logs.
// Safe to call with nil.
func (h *QuestAPIHandler) SetLogger(l *observability.Logger) {
	h.logger = l
}

// List delegates to QuestService.
func (h *QuestAPIHandler) List(ctx context.Context, crewID string) ([]QuestView, error) {
	return h.qs.List(ctx, crewID)
}

// GetByCrewAndID delegates to QuestService.
func (h *QuestAPIHandler) GetByCrewAndID(ctx context.Context, questID int64, crewID string) (*QuestWithChallenges, error) {
	return h.qs.GetByCrewAndID(ctx, questID, crewID)
}

// StartQuest transitions a quest from PENDING to ACTIVE for a crew. It first
// verifies the quest belongs to the crew (returning ErrQuestNotFound for
// unknown or unauthorized quests) before mutating state.
func (h *QuestAPIHandler) StartQuest(ctx context.Context, questID int64, crewID string) error {
	if _, err := h.qs.GetByCrewAndID(ctx, questID, crewID); err != nil {
		return err
	}
	return h.qs.StartQuest(ctx, questID)
}

// CompleteChallenge marks a challenge done, recomputes the quest status, and
// — when the quest becomes DONE — awards XP (including a completion bonus) and
// advances the crew's shared Realm Progress. The completing Explorer is
// always the one identified by uid.
func (h *QuestAPIHandler) CompleteChallenge(ctx context.Context, questID, challengeID int64, crewID, uid string) (result *CompleteChallengeResult, err error) {
	start := time.Now()
	defer func() {
		if h.logger == nil {
			return
		}
		outcome := "ok"
		errMsg := ""
		if err != nil {
			outcome = "error"
			errMsg = err.Error()
		}
		h.logger.LogServiceCall(observability.ServiceCallFields{
			RequestID: observability.RequestIDFromContext(ctx),
			Operation: "quest.complete_challenge",
			EntityID:  fmt.Sprintf("%d/%d", questID, challengeID),
			Duration:  time.Since(start),
			Outcome:   outcome,
			Error:     errMsg,
		})
	}()

	// Verify the quest belongs to the crew before mutating anything.
	qwc, err := h.qs.GetByCrewAndID(ctx, questID, crewID)
	if err != nil {
		return nil, err
	}

	status, questCompleted, err := h.qs.CompleteChallengeForQuest(ctx, questID, challengeID, uid)
	if err != nil {
		return nil, fmt.Errorf("complete challenge: %w", err)
	}
	_ = status

	challengeXP := ChallengeXP
	completionBonusXP := CompletionBonusXP
	if h.progCfg != nil {
		challengeXP = h.progCfg.ChallengeXP
		completionBonusXP = h.progCfg.CompletionBonusXP
	}
	progCfg := h.prog.Config()
	challengeXP = progCfg.ChallengeXP
	completionBonusXP = progCfg.CompletionBonusXP

	player, levelUp, err := h.prog.AwardXP(ctx, uid, challengeXP)
	if err != nil {
		return nil, fmt.Errorf("award challenge xp: %w", err)
	}

	result = &CompleteChallengeResult{
		QuestCompleted: questCompleted,
		NextAction:     "",
		XP:             challengeXP,
		NewLevel:       player.Level,
		LevelUp:        levelUp,
	}

	if questCompleted {
		result.NextAction = "CREATE_MEMORY"
		qd := (*gamecontent.QuestDefinition)(nil)
		if h.content != nil {
			qd, _ = h.content.GetQuest(ctx, qwc.TemplateSlug)
		}
		questRewardXP := completionBonusXP
		if qd != nil && qd.RewardXP > 0 {
			multiplier := 1.0
			if h.balance != nil {
				multiplier = h.balance.OverrideQuestRewardXP(1.0)
			}
			questRewardXP = int64(float64(qd.RewardXP) * multiplier)
		}

		bonusPlayer, bonusLevelUp, err := h.prog.AwardXP(ctx, uid, questRewardXP)
		if err != nil {
			return nil, fmt.Errorf("award completion xp: %w", err)
		}
		result.XP += questRewardXP
		result.NewLevel = bonusPlayer.Level
		result.LevelUp = result.LevelUp || bonusLevelUp

		realm := RealmForSlug(qwc.TemplateSlug)
		if realm != "" {
			if err := h.advanceRealm(ctx, crewID, realm); err != nil {
				return nil, fmt.Errorf("advance realm: %w", err)
			}
		}

		h.publishQuestCompleted(ctx, qwc, uid)
	}

	updated, err := h.qs.GetByCrewAndID(ctx, questID, crewID)
	if err != nil {
		return nil, fmt.Errorf("reload quest: %w", err)
	}
	result.Quest = updated

	return result, nil
}

// advanceRealm increments the crew's realm progress bar for a completed quest.
// When progress reaches the completion threshold, the realm is marked COMPLETE
// and the next realm in the sequence is unlocked. Uses configurable realm
// progression values.
func (h *QuestAPIHandler) advanceRealm(ctx context.Context, crewID, realm string) error {
	rp, err := h.realm.GetRealmProgress(ctx, crewID, realm)
	if err != nil {
		return fmt.Errorf("get realm progress: %w", err)
	}
	if rp == nil {
		return nil
	}

	progressPerQuest := RealmProgressPerQuest
	completionThreshold := RealmCompletionThreshold
	if h.realmCfg != nil {
		if def, ok := h.realmCfg.Get(realm); ok {
			if def.MaxProgress > 0 {
				completionThreshold = def.MaxProgress
			}
		}
	}
	if h.balance != nil {
		progressPerQuest = h.balance.OverrideRealmProgressPerQuest(progressPerQuest)
		completionThreshold = h.balance.OverrideRealmCompletionThreshold(completionThreshold)
	}

	next := rp.Progress + progressPerQuest
	if next > completionThreshold {
		next = completionThreshold
	}

	if next >= completionThreshold {
		matched, err := h.realm.UpdateRealmProgressIfMatch(ctx, crewID, realm, rp.Progress, map[string]any{
			"progress": completionThreshold,
			"status":   "COMPLETE",
		})
		if err != nil {
			return fmt.Errorf("complete realm: %w", err)
		}
		if !matched {
			if h.metrics != nil {
				h.metrics.RecordLockConflict()
			}
			return nil
		}

		if h.metrics != nil {
			h.metrics.RecordRealmCompleted()
		}

		nextRealm := NextRealm(realm)
		if h.realmCfg != nil {
			nextRealm = h.realmCfg.NextRealm(realm)
		}
		if nextRealm != "" {
			_, err = h.realm.CreateRealmProgress(ctx, &game.RealmProgress{
				CrewID:   crewID,
				Realm:    nextRealm,
				Status:   "ACTIVE",
				Progress: 0,
			})
			if err != nil {
				return fmt.Errorf("unlock next realm: %w", err)
			}
		}
	} else {
		matched, err := h.realm.UpdateRealmProgressIfMatch(ctx, crewID, realm, rp.Progress, map[string]any{"progress": next})
		if err != nil {
			return fmt.Errorf("update realm progress: %w", err)
		}
		if !matched && h.metrics != nil {
			h.metrics.RecordLockConflict()
		}
	}
	return nil
}

// publishQuestCompleted emits a QuestCompleted event if a publisher is configured.
// The event carries the quest's template slug, realm, chapter, and the completing
// player's UID. Listeners (chapter, achievement, lore services) react independently.
func (h *QuestAPIHandler) publishQuestCompleted(ctx context.Context, qwc *QuestWithChallenges, uid string) {
	if h.publisher == nil {
		return
	}
	chapter := ""
	seasonSlug := ""
	if h.content != nil {
		if qd, err := h.content.GetQuest(ctx, qwc.TemplateSlug); err == nil && qd != nil {
			chapter = qd.Chapter
			seasonSlug = qd.SeasonSlug
		}
	}
	h.publisher.Publish(ctx, events.QuestCompletedEvent{
		QuestID:      qwc.ID,
		CrewID:       qwc.CrewID,
		TemplateSlug: qwc.TemplateSlug,
		Realm:        RealmForSlug(qwc.TemplateSlug),
		Chapter:      chapter,
		SeasonSlug:   seasonSlug,
		PlayerUID:    uid,
	})
}
