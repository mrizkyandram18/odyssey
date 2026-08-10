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

// ChallengeBatchReader is an optional capability a QuestStore may implement
// to fetch challenges for many quests in a single request. QuestService uses
// it when listing a crew to avoid one sequential challenge read per quest.
type ChallengeBatchReader interface {
	ListChallengesByQuestIDs(ctx context.Context, questIDs []int64) ([]game.Challenge, error)
}

// CrewMember is the minimal explorer identity exposed in quest detail so the
// relay rotation UI can render names instead of raw UIDs. Read-only, no
// sensitive fields.
type CrewMember struct {
	UID          string `json:"uid"`
	ExplorerName string `json:"explorer_name"`
	Role         string `json:"role,omitempty"`
	Level        int    `json:"level,omitempty"`
}

// QuestWithChallenges is a quest returned together with its challenge list.
type QuestWithChallenges struct {
	game.Quest
	QuestType                 string           `json:"quest_type,omitempty"`
	Challenges                []game.Challenge `json:"challenges"`
	ActiveChallengeAssignedTo *string          `json:"active_challenge_assigned_to,omitempty"`
	// Members is a best-effort roster of the crew (UID -> name/role) used to
	// render relay leg ownership. Omitted when the user store is unavailable.
	Members []CrewMember `json:"members,omitempty"`
}

// QuestView is the quest summary used for list responses.
type QuestView struct {
	game.Quest
	QuestType                 string  `json:"quest_type,omitempty"`
	ChallengeCount            int     `json:"challenge_count"`
	CompletedCount            int     `json:"completed_count"`
	ActiveChallengeAssignedTo *string `json:"active_challenge_assigned_to,omitempty"`
}

// QuestService owns quest lifecycle and business logic.
// It depends only on the QuestStore interface — never on a concrete db adapter.
type QuestService struct {
	store   game.QuestStore
	gate    QuestGate
	content ContentGateway
	users   game.UserStore
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

// SetUserStore injects the UserStore for cooperative assignment.
func (s *QuestService) SetUserStore(users game.UserStore) {
	s.users = users
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

	challengesByQuest, err := s.challengesByQuestID(ctx, quests)
	if err != nil {
		return nil, err
	}

	views := make([]QuestView, 0, len(quests))
	for _, q := range quests {
		challenges := challengesByQuest[q.ID]
		done := 0
		for _, c := range challenges {
			if c.Status == string(ChallengeStatusDone) {
				done++
			}
		}
		views = append(views, QuestView{
			Quest:                     q,
			QuestType:                 string(TypeForSlug(q.TemplateSlug)),
			ChallengeCount:            len(challenges),
			CompletedCount:            done,
			ActiveChallengeAssignedTo: firstPendingAssignee(challenges),
		})
	}
	return views, nil
}

// firstPendingAssignee returns the assigned_to of the first PENDING challenge,
// or nil when every challenge is done or none is assigned.
func firstPendingAssignee(challenges []game.Challenge) *string {
	for _, c := range challenges {
		if c.Status == string(ChallengeStatusPending) && c.AssignedTo != nil {
			return c.AssignedTo
		}
	}
	return nil
}

// challengesByQuestID loads the challenges for all given quests, keyed by
// quest ID. When the store supports batch reads (ChallengeBatchReader), a
// single request fetches challenges for every quest; otherwise it falls back
// to one GetChallenges read per quest. Quests without challenges map to nil.
func (s *QuestService) challengesByQuestID(ctx context.Context, quests []game.Quest) (map[int64][]game.Challenge, error) {
	ids := make([]int64, 0, len(quests))
	for _, q := range quests {
		ids = append(ids, q.ID)
	}

	if batch, ok := s.store.(ChallengeBatchReader); ok {
		challenges, err := batch.ListChallengesByQuestIDs(ctx, ids)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch challenges: %w", err)
		}
		grouped := make(map[int64][]game.Challenge, len(ids))
		for _, c := range challenges {
			grouped[c.QuestID] = append(grouped[c.QuestID], c)
		}
		return grouped, nil
	}

	grouped := make(map[int64][]game.Challenge, len(ids))
	for _, q := range quests {
		challenges, err := s.store.GetChallenges(ctx, q.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch challenges for quest %d: %w", q.ID, err)
		}
		grouped[q.ID] = challenges
	}
	return grouped, nil
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
		Quest:                     *q,
		QuestType:                 string(TypeForSlug(q.TemplateSlug)),
		Challenges:                challenges,
		ActiveChallengeAssignedTo: firstPendingAssignee(challenges),
	}
	computed := string(s.ComputeStatus(challenges))
	if computed == string(QuestStatusPending) && (q.Status == string(QuestStatusActive) || q.StartedAt != nil) {
		result.Status = string(QuestStatusActive)
	} else {
		result.Status = computed
	}
	// Best-effort crew roster for name resolution in the relay rotation UI.
	// Never fails the request when the user store is missing or unreachable.
	if s.users != nil {
		if players, err := s.users.ListUsersByCrew(ctx, q.CrewID); err == nil {
			result.Members = toCrewMembers(players)
		}
	}
	return result, nil
}

// toCrewMembers maps player rows to the minimal CrewMember view.
func toCrewMembers(players []game.Player) []CrewMember {
	members := make([]CrewMember, 0, len(players))
	for _, p := range players {
		members = append(members, CrewMember{
			UID:          p.UID,
			ExplorerName: p.ExplorerName,
			Role:         p.Role,
			Level:        p.Level,
		})
	}
	return members
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
// Returns the recomputed quest status, whether this call transitioned the
// challenge from PENDING to DONE (progressed), and whether the quest *just*
// completed (i.e. it was not DONE before this call and is now DONE). When
// progressed is true, callers award challenge XP. When completed is true,
// callers trigger quest-completion rewards (XP, realm progress, events).
//
// Exactly-once guarantees:
//   - The challenge is finalized with an atomic CAS (PENDING -> DONE), so a
//     replay/retry of an already-completed challenge returns progressed=false
//     and never re-awards.
//   - The quest row is finalized with an atomic CAS against the current stored
//     status (with a bounded retry), so concurrent completions converge on
//     exactly one caller with completed=true.
func (s *QuestService) CompleteChallengeForQuest(ctx context.Context, questID, challengeID int64, uid string) (QuestStatus, bool, bool, error) {
	now := time.Now().UTC()

	challengePatch := map[string]any{
		"status":       string(ChallengeStatusDone),
		"completed_by": uid,
		"completed_at": &now,
	}
	matched, err := s.store.UpdateChallengeIfMatch(ctx, challengeID, string(ChallengeStatusPending), challengePatch)
	if err != nil {
		return "", false, false, err
	}
	progressed := matched

	challenges, err := s.store.GetChallenges(ctx, questID)
	if err != nil {
		return "", false, false, err
	}
	newStatus := computeStatus(challenges)

	if !progressed {
		// Replay of an already-completed challenge. Never re-award, but
		// reconcile a drifted stored quest status (e.g. a stuck quest whose
		// row was left ACTIVE after all challenges were completed).
		return s.reconcileQuestStatus(ctx, questID, newStatus, &now), false, false, nil
	}

	// We completed the challenge. Finalize the quest status with an atomic CAS
	// against the CURRENT stored status, retrying once if a concurrent
	// completion wins the race. Exactly one caller transitions non-DONE -> DONE.
	completed := false
	for attempt := 0; attempt < 2; attempt++ {
		stored, err := s.store.GetQuest(ctx, questID)
		if err != nil {
			return "", true, false, err
		}
		if stored == nil {
			return "", true, false, ErrQuestNotFound
		}
		if stored.Status == string(QuestStatusDone) {
			return newStatus, true, false, nil
		}
		questPatch := map[string]any{"status": string(newStatus)}
		if newStatus == QuestStatusDone {
			questPatch["completed_at"] = &now
		}
		if stored.StartedAt == nil {
			questPatch["started_at"] = &now
		}
		ok, err := s.store.UpdateQuestIfMatch(ctx, questID, stored.Status, questPatch)
		if err != nil {
			return "", true, false, err
		}
		if ok {
			completed = newStatus == QuestStatusDone
			break
		}
	}

	if !completed && newStatus == QuestStatusDone {
		// Two consecutive CAS losses while all challenges are DONE: verify the
		// stored state before reporting, instead of returning a stale result.
		if stored, err := s.store.GetQuest(ctx, questID); err == nil && stored != nil && stored.Status == string(QuestStatusDone) {
			return QuestStatusDone, true, false, nil
		}
	}

	// Assign next challenge owner if the quest is still active.
	if newStatus == QuestStatusActive {
		q, qErr := s.store.GetQuest(ctx, questID)
		if qErr == nil && q != nil {
			for _, c := range challenges {
				if c.Status == string(ChallengeStatusPending) {
					nextOwner := s.AssignNextChallengeOwner(ctx, q.CrewID, uid)
					if nextOwner != "" {
						_ = s.store.UpdateChallenge(ctx, c.ID, map[string]any{"assigned_to": nextOwner})
					}
					break
				}
			}
		}
	}

	return newStatus, true, completed, nil
}

// reconcileQuestStatus brings the stored quest status in line with the status
// computed from its challenges. Used when a challenge completion is a replay
// (already DONE) — the quest may still hold a stale stored status. The update
// is best-effort: a lost CAS is ignored because the concurrent writer owns the
// final state.
func (s *QuestService) reconcileQuestStatus(ctx context.Context, questID int64, computed QuestStatus, now *time.Time) QuestStatus {
	stored, err := s.store.GetQuest(ctx, questID)
	if err != nil || stored == nil {
		return computed
	}
	if stored.Status == string(computed) || stored.Status == string(QuestStatusDone) {
		return computed
	}
	patch := map[string]any{"status": string(computed)}
	if computed == QuestStatusDone {
		patch["completed_at"] = now
	}
	if stored.StartedAt == nil {
		patch["started_at"] = now
	}
	_, _ = s.store.UpdateQuestIfMatch(ctx, questID, stored.Status, patch)
	return computed
}

// AssignNextChallengeOwner determines the next user for a challenge round-robin.
func (s *QuestService) AssignNextChallengeOwner(ctx context.Context, crewID, completedBy string) string {
	if s.users == nil {
		return completedBy
	}
	players, err := s.users.ListUsersByCrew(ctx, crewID)
	if err != nil || len(players) == 0 {
		return completedBy
	}

	// Round robin: find index of completedBy, return next index.
	idx := -1
	for i, p := range players {
		if p.UID == completedBy {
			idx = i
			break
		}
	}

	if idx == -1 || idx == len(players)-1 {
		return players[0].UID
	}
	return players[idx+1].UID
}

// sortQuestsByCreatedAt sorts quests newest-first for display.
func sortQuestsByCreatedAt(quests []game.Quest) {
	for i := 1; i < len(quests); i++ {
		for j := i; j > 0 && quests[j].CreatedAt.After(quests[j-1].CreatedAt); j-- {
			quests[j], quests[j-1] = quests[j-1], quests[j]
		}
	}
}
