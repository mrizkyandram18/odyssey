package quest

import (
	"context"
	"errors"
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
	rewards   RewardGateway
	content   ContentGateway
	balance   *balance.Service
	metrics   *observability.Metrics
	logger    *observability.Logger
}

// RewardGateway provides reward logic
type RewardGateway interface {
	GrantQuestReward(ctx context.Context, uid string, questID int64, xp int64) error
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

// SetRewardService attaches a reward service.
func (h *QuestAPIHandler) SetRewardService(r RewardGateway) {
	h.rewards = r
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

// ListAvailable returns season- and prerequisite-filtered quest definitions
// mapped to QuestView so the frontend can render them as pending opportunities.
func (h *QuestAPIHandler) ListAvailable(ctx context.Context, crewID, uid string) ([]QuestView, error) {
	defs, err := h.qs.ListAvailable(ctx, crewID, uid)
	if err != nil {
		return nil, err
	}
	result := make([]QuestView, 0, len(defs))
	for _, def := range defs {
		result = append(result, QuestView{
			Quest: game.Quest{
				ID:           def.ID,
				CrewID:       crewID,
				TemplateSlug: def.Slug,
				Title:        def.Title,
				Status:       string(QuestStatusPending),
				CreatedAt:    def.CreatedAt,
			},
			QuestType:                 def.QuestType,
			ChallengeCount:            len(def.ChallengeDefs),
			CompletedCount:            0,
			ActiveChallengeAssignedTo: nil,
		})
	}
	return result, nil
}

// GetByCrewAndID delegates to QuestService.
func (h *QuestAPIHandler) GetByCrewAndID(ctx context.Context, questID int64, crewID string) (*QuestWithChallenges, error) {
	return h.qs.GetByCrewAndID(ctx, questID, crewID)
}

// StartQuest transitions a quest from PENDING to ACTIVE for a crew. It first
// verifies the quest belongs to the crew (returning ErrQuestNotFound for
// unknown or unauthorized quests) before mutating state.
func (h *QuestAPIHandler) StartQuest(ctx context.Context, questID int64, crewID, uid string) error {
	qwc, err := h.qs.GetByCrewAndID(ctx, questID, crewID)
	if err != nil {
		return err
	}

	if qwc.StartedAt == nil && qwc.Status == string(QuestStatusPending) {
		maxNew := 1
		if h.balance != nil {
			maxNew = h.balance.OverrideMaxNewQuestsPerDay(maxNew)
		}

		if maxNew > 0 {
			quests, err := h.qs.ListByCrew(ctx, crewID)
			if err != nil {
				return fmt.Errorf("list quests: %w", err)
			}

			today := time.Now().UTC().Truncate(24 * time.Hour)
			startedToday := 0
			for _, q := range quests {
				if q.StartedBy != nil && *q.StartedBy == uid && q.StartedAt != nil && q.StartedAt.UTC().Truncate(24*time.Hour).Equal(today) {
					startedToday++
				}
			}

			if startedToday >= maxNew {
				return errors.New("daily new quest limit reached")
			}
		}
	}

	return h.qs.StartQuest(ctx, questID, uid)
}

// CompleteChallenge marks a challenge done, recomputes the quest status, and
// — when the quest becomes DONE — awards XP (including a completion bonus) and
// advances the crew's shared Realm Progress. The completing Explorer is
// always the one identified by uid.
func (h *QuestAPIHandler) CompleteChallenge(ctx context.Context, questID, challengeID int64, crewID, uid string, answer string) (result *CompleteChallengeResult, err error) {
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

	status, progressed, questCompleted, err := h.qs.CompleteChallengeForQuest(ctx, questID, challengeID, uid, answer)
	if err != nil {
		return nil, fmt.Errorf("complete challenge: %w", err)
	}
	_ = status

	result = &CompleteChallengeResult{
		QuestCompleted: questCompleted,
		NextAction:     "",
	}

	if !progressed && !questCompleted {
		// Replay of an already-completed challenge (double-tap or client retry):
		// no XP, no rewards — return the fresh quest state.
		// However, if questCompleted is true, it means this replay fixed a stuck quest
		// (or won the CAS against the actual completer), so we must proceed to award quest XP.
		updated, err := h.qs.GetByCrewAndID(ctx, questID, crewID)
		if err != nil {
			return nil, fmt.Errorf("reload quest: %w", err)
		}
		result.Quest = updated
		return result, nil
	}

	challengeXP := ChallengeXP
	completionBonusXP := CompletionBonusXP
	if h.progCfg != nil {
		challengeXP = h.progCfg.ChallengeXP
		completionBonusXP = h.progCfg.CompletionBonusXP
	}
	progCfg := h.prog.Config()
	challengeXP = progCfg.ChallengeXP
	completionBonusXP = progCfg.CompletionBonusXP

	result.XP = 0
	if progressed {
		player, levelUp, err := h.prog.AwardXP(ctx, uid, challengeXP)
		if err != nil {
			return nil, fmt.Errorf("award challenge xp: %w", err)
		}
		result.XP = challengeXP
		result.NewLevel = player.Level
		result.LevelUp = levelUp
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

		if h.rewards != nil {
			_ = h.rewards.GrantQuestReward(ctx, uid, questID, questRewardXP)
		}

		h.publishQuestCompleted(ctx, qwc, uid)
	}

	updated, err := h.qs.GetByCrewAndID(ctx, questID, crewID)
	if err != nil {
		return nil, fmt.Errorf("reload quest: %w", err)
	}
	result.Quest = updated

	// Emit RelayHandoffEvent when a relay leg was handed off to the next explorer.
	// Conditions: the challenge progressed (not a replay), the quest is still
	// active (not yet done), and there is a pending challenge with a new assignee
	// that differs from the completer (uid). This is best-effort and never blocks
	// the response.
	if progressed && !questCompleted && h.publisher != nil {
		h.publishRelayHandoff(ctx, updated, uid)
	}

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
			if err := h.ensureNextRealmUnlocked(ctx, crewID, nextRealm); err != nil {
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

// ensureNextRealmUnlocked activates the realm that follows a completed realm.
//
// The target realm may already exist as LOCKED (migrations seed locked realm
// rows so the realm map shows them before they are earned). In that case the
// existing row is activated via update; otherwise a new row is inserted. If a
// concurrent completion wins the insert race, the update branch converges, so
// the unlock is exactly-once without introducing a unique-violation error like
// a naive INSERT would (see 022 realm-seed interaction).
func (h *QuestAPIHandler) ensureNextRealmUnlocked(ctx context.Context, crewID, realm string) error {
	unlock := map[string]any{
		"status":           "ACTIVE",
		"progress":         0,
		"last_unlocked_at": time.Now().UTC(),
	}
	rp, err := h.realm.GetRealmProgress(ctx, crewID, realm)
	if err == nil && rp != nil {
		return h.realm.UpdateRealmProgress(ctx, crewID, realm, unlock)
	}
	if err != nil && !errors.Is(err, game.ErrNotFound) {
		return err
	}
	_, err = h.realm.CreateRealmProgress(ctx, &game.RealmProgress{
		CrewID:         crewID,
		Realm:          realm,
		Status:         "ACTIVE",
		Progress:       0,
		LastUnlockedAt: time.Now().UTC(),
	})
	if err != nil {
		// Race: another caller inserted the row between our Get and Create.
		// Converge by activating the now-existing row.
		if rp, getErr := h.realm.GetRealmProgress(ctx, crewID, realm); getErr == nil && rp != nil {
			return h.realm.UpdateRealmProgress(ctx, crewID, realm, unlock)
		}
		return err
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

// publishRelayHandoff emits a RelayHandoffEvent when the quest is still active
// and its first pending challenge has an assignee other than the completer.
// Best-effort: called only when publisher is set and challenge progressed.
func (h *QuestAPIHandler) publishRelayHandoff(ctx context.Context, qwc *QuestWithChallenges, completedBy string) {
	if h.publisher == nil || qwc == nil {
		return
	}
	for _, c := range qwc.Challenges {
		if c.Status == string(ChallengeStatusPending) && c.AssignedTo != nil && *c.AssignedTo != "" && *c.AssignedTo != completedBy {
			h.publisher.Publish(ctx, events.RelayHandoffEvent{
				FromUID:    completedBy,
				ToUID:      *c.AssignedTo,
				QuestID:    qwc.ID,
				QuestTitle: qwc.Title,
				CrewID:     qwc.CrewID,
			})
			return
		}
	}
}

// SelectBranchResult is returned by SelectBranch when a family member picks a narrative branch.
type SelectBranchResult struct {
	Success     bool   `json:"success"`
	StoryBranch string `json:"story_branch"`
	Realm       string `json:"realm"`
}

// SelectBranch records the family's narrative branch choice for a quest.
func (h *QuestAPIHandler) SelectBranch(ctx context.Context, questID int64, crewID, branchChoice string) (*SelectBranchResult, error) {
	qwc, err := h.qs.GetByCrewAndID(ctx, questID, crewID)
	if err != nil {
		return nil, err
	}
	tpl, ok := LookupTemplate(qwc.TemplateSlug)
	if !ok || len(tpl.BranchOptions) == 0 {
		return nil, ErrNoBranchOptions
	}

	var valid bool
	for _, opt := range tpl.BranchOptions {
		if opt.Slug == branchChoice {
			valid = true
			break
		}
	}
	if !valid {
		return nil, ErrInvalidBranchChoice
	}

	realm := tpl.Realm
	if realm == "" {
		realm = MVPRealm
	}

	rp, err := h.realm.GetRealmProgress(ctx, crewID, realm)
	if err != nil {
		return nil, fmt.Errorf("get realm progress: %w", err)
	}
	if rp == nil {
		return nil, fmt.Errorf("realm progress not found for %s", realm)
	}

	if err := h.realm.UpdateRealmProgress(ctx, crewID, realm, map[string]any{
		"story_branch": branchChoice,
	}); err != nil {
		return nil, fmt.Errorf("update story_branch: %w", err)
	}

	return &SelectBranchResult{
		Success:     true,
		StoryBranch: branchChoice,
		Realm:       realm,
	}, nil
}
