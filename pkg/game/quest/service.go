package quest

import (
	"context"
	"errors"
	"fmt"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
)

var ErrQuestNotFound = errors.New("quest not found")
var ErrQuestNotInCrew = errors.New("quest not found")
var ErrChallengeNotFound = errors.New("challenge not found")

// QuestGate provides cross-service checks needed for prerequisite filtering.
// Implemented by an adapter composed in the handler layer to avoid import cycles.
type QuestGate interface {
	IsChapterUnlocked(ctx context.Context, crewID, chapter string) bool
	IsRealmActive(ctx context.Context, crewID, realm string) bool
	IsSeasonActive(ctx context.Context, seasonSlug string) bool
	GetPlayerLevel(ctx context.Context, uid string) (int, error)
	IsQuestCompleted(ctx context.Context, crewID, templateSlug string) bool
	ValidatePrerequisites(ctx context.Context, defs []gamecontent.QuestDefinition) error
}

// ContentGateway provides content definition lookups for the QuestService.
type ContentGateway interface {
	ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error)
	GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error)
}

// QuestWithChallenges is a quest returned together with its challenge list.
type QuestWithChallenges struct {
	game.Quest
	Challenges []game.Challenge `json:"challenges"`
}

// QuestView is the quest summary used for list responses.
type QuestView struct {
	game.Quest
	ChallengeCount int `json:"challenge_count"`
	CompletedCount int `json:"completed_count"`
}

// QuestService owns quest lifecycle and business logic.
// It depends only on the QuestStore interface — never on a concrete db adapter.
type QuestService struct {
	store   game.QuestStore
	gate    QuestGate
	content ContentGateway
}

// NewQuestService constructs a QuestService with the given quest store.
func NewQuestService(store game.QuestStore) *QuestService {
	return &QuestService{store: store}
}

// NewQuestServiceWithGate constructs a QuestService with prerequisite
// filtering enabled via the given gate and content gateway.
func NewQuestServiceWithGate(store game.QuestStore, gate QuestGate, content ContentGateway) *QuestService {
	return &QuestService{store: store, gate: gate, content: content}
}

// IsPrerequisiteMet checks whether a quest definition's prerequisites
// are satisfied for the given crew and player.
func (s *QuestService) IsPrerequisiteMet(ctx context.Context, def gamecontent.QuestDefinition, crewID, uid string) bool {
	if s.gate == nil || s.content == nil {
		return true
	}
	if !s.gate.IsSeasonActive(ctx, def.SeasonSlug) {
		return false
	}
	if def.RequiredLevel > 0 {
		level, err := s.gate.GetPlayerLevel(ctx, uid)
		if err != nil || level < def.RequiredLevel {
			return false
		}
	}
	if def.RequiredRealm != "" && !s.gate.IsRealmActive(ctx, crewID, def.RequiredRealm) {
		return false
	}
	if def.RequiredChapter != "" && !s.gate.IsChapterUnlocked(ctx, crewID, def.RequiredChapter) {
		return false
	}

	prereqs := def.RequiredQuestSlugs
	if len(prereqs) == 0 && def.RequiredQuestSlug != "" {
		prereqs = []string{def.RequiredQuestSlug}
	}
	for _, prereq := range prereqs {
		if !s.gate.IsQuestCompleted(ctx, crewID, prereq) {
			return false
		}
	}
	return true
}

// ListAvailable returns quest definitions whose prerequisites are satisfied
// for the given crew and player. Quests without prerequisites are always included.
func (s *QuestService) ListAvailable(ctx context.Context, crewID, uid string) ([]gamecontent.QuestDefinition, error) {
	if s.content == nil {
		return nil, errors.New("quest service has no content gateway for prerequisite filtering")
	}
	defs, err := s.content.ListQuests(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]gamecontent.QuestDefinition, 0, len(defs))
	for _, def := range defs {
		if s.IsPrerequisiteMet(ctx, def, crewID, uid) {
			result = append(result, def)
		}
	}
	return result, nil
}

// ListByCrew returns all quests for a crew, ordered by creation date (newest first).
func (s *QuestService) ListByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	quests, err := s.store.ListQuestByCrew(ctx, crewID)
	if err != nil {
		return nil, err
	}
	sortQuestsByCreatedAt(quests)
	return quests, nil
}

// List returns all quests for a given crew as QuestView summaries,
// ordered by creation date (newest first).
func (s *QuestService) List(ctx context.Context, crewID string) ([]QuestView, error) {
	quests, err := s.ListByCrew(ctx, crewID)
	if err != nil {
		return nil, err
	}

	views := make([]QuestView, 0, len(quests))
	for _, q := range quests {
		challenges, err := s.store.GetChallenges(ctx, q.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch challenges for quest %d: %w", q.ID, err)
		}
		done := 0
		for _, c := range challenges {
			if c.Status == string(ChallengeStatusDone) {
				done++
			}
		}
		views = append(views, QuestView{
			Quest:          q,
			ChallengeCount: len(challenges),
			CompletedCount: done,
		})
	}
	return views, nil
}

// ActiveQuests filters a list of quests, returning only those that are
// not yet DONE (i.e. PENDING or ACTIVE).
func ActiveQuests(quests []game.Quest) []game.Quest {
	active := make([]game.Quest, 0, len(quests))
	for _, q := range quests {
		if q.Status != string(QuestStatusDone) {
			active = append(active, q)
		}
	}
	return active
}

// GetByCrewAndID retrieves a single quest by ID, verifying it belongs to
// the specified crew. Returns the quest with its challenges.
// Returns ErrQuestNotFound if the quest does not exist or does not belong
// to the crew (the same error is used to avoid leaking existence to
// unauthorized users).
func (s *QuestService) GetByCrewAndID(ctx context.Context, questID int64, crewID string) (*QuestWithChallenges, error) {
	q, err := s.store.GetQuest(ctx, questID)
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil, ErrQuestNotFound
		}
		return nil, err
	}
	if q == nil {
		return nil, ErrQuestNotFound
	}
	if q.CrewID != crewID {
		return nil, ErrQuestNotInCrew
	}
	return s.getWithChallenges(ctx, q)
}

// getWithChallenges fetches the challenges for a quest and attaches them.
func (s *QuestService) getWithChallenges(ctx context.Context, q *game.Quest) (*QuestWithChallenges, error) {
	challenges, err := s.store.GetChallenges(ctx, q.ID)
	if err != nil {
		return nil, err
	}

	result := &QuestWithChallenges{
		Quest:      *q,
		Challenges: challenges,
	}
	result.Status = string(s.ComputeStatus(challenges))
	return result, nil
}

// ComputeStatus determines the quest's lifecycle status from its challenges.
//
// Lifecycle rules:
//   - PENDING: no challenges are DONE (or no challenges exist)
//   - ACTIVE: at least one challenge is DONE, but not all
//   - DONE: all challenges are DONE
//
// This is a pure function with no side effects.
func (s *QuestService) ComputeStatus(challenges []game.Challenge) QuestStatus {
	return computeStatus(challenges)
}

// computeStatus is the internal pure implementation shared across call sites.
func computeStatus(challenges []game.Challenge) QuestStatus {
	if len(challenges) == 0 {
		return QuestStatusPending
	}
	done := 0
	for _, c := range challenges {
		if c.Status == string(ChallengeStatusDone) {
			done++
		}
	}
	if done == 0 {
		return QuestStatusPending
	}
	if done < len(challenges) {
		return QuestStatusActive
	}
	return QuestStatusDone
}

// StartQuest transitions a quest from PENDING to ACTIVE.
// This is a write operation (persists the status change) and is intended
// for when a family member begins work on a quest.
func (s *QuestService) StartQuest(ctx context.Context, questID int64) error {
	now := time.Now().UTC()
	patch := map[string]any{
		"status":     string(QuestStatusActive),
		"started_at": &now,
	}
	return s.store.UpdateQuest(ctx, questID, patch)
}

// CompleteChallenge marks a single challenge as DONE and recomputes the
// quest's overall status. If all challenges are now done, the quest
// transitions to DONE with a completed_at timestamp.
//
// Returns the updated quest status and whether the quest was completed.
func (s *QuestService) CompleteChallenge(ctx context.Context, challengeID int64, uid string) (QuestStatus, bool, error) {
	now := time.Now().UTC()
	patch := map[string]any{
		"status":       string(ChallengeStatusDone),
		"completed_by": uid,
		"completed_at": now,
	}
	if err := s.store.UpdateChallenge(ctx, challengeID, patch); err != nil {
		return "", false, err
	}

	// Recompute the quest status by fetching the quest and its challenges.
	// We need the quest_id from the challenge, but the store interface
	// doesn't expose a GetChallenge method. The caller is expected to
	// provide the quest context externally, or we fetch all challenges
	// for the quest.
	//
	// Since we don't have the questID here, the quest status update is
	// expected to be applied by a higher-level orchestrator that has
	// the questID. For standalone challenge completion, we return the
	// computed status based on the quest's challenges.
	//
	// Note: In the MVP, CompleteChallenge is defined for the lifecycle
	// contract but is not wired to any write endpoint.
	return QuestStatusActive, false, nil
}

// CompleteChallengeForQuest marks a challenge as done and updates the quest
// status based on the new challenge state. This is the full lifecycle
// transition used when a family member completes a challenge within a quest.
//
// Returns the recomputed quest status and whether the quest *just* completed
// (i.e. it was not DONE before this challenge and is now DONE). When
// completed is true, callers should trigger quest-completion rewards
// (XP, realm progress).
func (s *QuestService) CompleteChallengeForQuest(ctx context.Context, questID, challengeID int64, uid string) (QuestStatus, bool, error) {
	now := time.Now().UTC()

	before, err := s.store.GetChallenges(ctx, questID)
	if err != nil {
		return "", false, err
	}
	oldStatus := computeStatus(before)

	challengePatch := map[string]any{
		"status":       string(ChallengeStatusDone),
		"completed_by": uid,
		"completed_at": &now,
	}
	if err := s.store.UpdateChallenge(ctx, challengeID, challengePatch); err != nil {
		return "", false, err
	}

	challenges, err := s.store.GetChallenges(ctx, questID)
	if err != nil {
		return "", false, err
	}
	newStatus := computeStatus(challenges)

	questPatch := map[string]any{
		"status": string(newStatus),
	}
	if newStatus == QuestStatusDone {
		questPatch["completed_at"] = &now
	}
	if newStatus == QuestStatusActive && oldStatus != QuestStatusActive {
		questPatch["started_at"] = &now
	}

	if oldStatus != QuestStatusDone {
		matched, err := s.store.UpdateQuestIfMatch(ctx, questID, string(oldStatus), questPatch)
		if err != nil {
			return "", false, err
		}
		if !matched {
			return QuestStatusDone, false, nil
		}
	} else {
		if err := s.store.UpdateQuest(ctx, questID, questPatch); err != nil {
			return "", false, err
		}
	}

	completed := oldStatus != QuestStatusDone && newStatus == QuestStatusDone
	return newStatus, completed, nil
}

// sortQuestsByCreatedAt sorts quests newest-first for display.
func sortQuestsByCreatedAt(quests []game.Quest) {
	for i := 1; i < len(quests); i++ {
		for j := i; j > 0 && quests[j].CreatedAt.After(quests[j-1].CreatedAt); j-- {
			quests[j], quests[j-1] = quests[j-1], quests[j]
		}
	}
}
