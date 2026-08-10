package achievement

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/sync/errgroup"

	"odyssey/pkg/game"
	"odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
)

var (
	ErrAchievementNotFound = errors.New("achievement not found")
	ErrAlreadyUnlocked     = errors.New("achievement already unlocked")
)

type AchievementTrigger string

const (
	TriggerQuestCompleted     AchievementTrigger = "QUEST_COMPLETED"
	TriggerRealmCompleted     AchievementTrigger = "REALM_COMPLETED"
	TriggerChapterCompleted   AchievementTrigger = "CHAPTER_COMPLETED"
	TriggerRelicCollected     AchievementTrigger = "RELIC_COLLECTED"
	TriggerDailyStreak        AchievementTrigger = "DAILY_STREAK"
	TriggerCreativeSubmission AchievementTrigger = "CREATIVE_SUBMISSION"
	TriggerLevelReached       AchievementTrigger = "LEVEL_REACHED"
)

type AchievementView struct {
	ID        int64     `json:"id"`
	Code      string    `json:"code"`
	Title     string    `json:"title"`
	Kind      string    `json:"kind"`
	Trigger   string    `json:"trigger"`
	Threshold int       `json:"threshold"`
	Progress  int       `json:"progress"`
	Unlocked  bool      `json:"unlocked"`
	AwardedAt time.Time `json:"awarded_at,omitempty"`
}

type ProgressReader interface {
	CountCompletedQuests(ctx context.Context, crewID string) (int, error)
	CountCompletedRealms(ctx context.Context, crewID string) (int, error)
	CountCompletedChapters(ctx context.Context, crewID string) (int, error)
	CountCollectedRelics(ctx context.Context, uid string) (int, error)
	CountDailyStreak(ctx context.Context, uid string) (int, error)
	CountCreativeSubmissions(ctx context.Context, crewID string) (int, error)
	GetPlayerLevel(ctx context.Context, uid string) (int, error)
}

type AchievementGateway interface {
	ListAchievements(ctx context.Context) ([]content.AchievementDefinition, error)
	GetAchievement(ctx context.Context, code string) (*content.AchievementDefinition, error)
}

type AchievementService struct {
	defs      AchievementGateway
	store     game.AchievementStore
	progress  ProgressReader
	publisher events.Publisher
}

func NewAchievementService(
	defs AchievementGateway,
	store game.AchievementStore,
	progress ProgressReader,
	publisher events.Publisher,
) *AchievementService {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	return &AchievementService{
		defs:      defs,
		store:     store,
		progress:  progress,
		publisher: publisher,
	}
}

func (s *AchievementService) ListByPlayer(ctx context.Context, uid string) ([]AchievementView, error) {
	earned, err := s.store.ListAchievementsByPlayer(ctx, uid)
	if err != nil {
		earned = nil
	}
	earnedMap := make(map[string]game.Achievement, len(earned))
	for _, a := range earned {
		earnedMap[a.Code] = a
	}

	defs, err := s.defs.ListAchievements(ctx)
	if err != nil {
		return nil, fmt.Errorf("list achievement defs: %w", err)
	}

	// Prefetch distinct progress sources in parallel (one read per trigger key),
	// then assemble views. Soft-error → 0 is preserved; memoization remains.
	const crewID = ""
	progressCache := s.prefetchProgress(ctx, defs, uid, crewID)

	result := make([]AchievementView, 0, len(defs))
	for _, def := range defs {
		key := progressCacheKey(def.Trigger, uid, crewID)
		progress := progressCache[key]
		a, awarded := earnedMap[def.Code]
		awardedAt := time.Time{}
		if awarded {
			awardedAt = a.AwardedAt
		}
		result = append(result, AchievementView{
			ID:        def.ID,
			Code:      def.Code,
			Title:     def.Title,
			Kind:      def.Kind,
			Trigger:   def.Trigger,
			Threshold: def.Threshold,
			Progress:  progress,
			Unlocked:  awarded,
			AwardedAt: awardedAt,
		})
	}
	return result, nil
}

func progressCacheKey(trigger, uid, crewID string) string {
	return trigger + "|" + uid + "|" + crewID
}

// prefetchProgress loads progress for each distinct trigger source in parallel.
// Failed counts soft-error to 0 and are still cached so duplicate defs share
// one underlying read. Wait completes before any cache map is written.
func (s *AchievementService) prefetchProgress(
	ctx context.Context,
	defs []content.AchievementDefinition,
	uid, crewID string,
) map[string]int {
	type progressJob struct {
		key string
		def content.AchievementDefinition
	}

	// One job per distinct cache key (trigger + scope).
	seen := make(map[string]struct{}, len(defs))
	jobs := make([]progressJob, 0, len(defs))
	for _, def := range defs {
		key := progressCacheKey(def.Trigger, uid, crewID)
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		jobs = append(jobs, progressJob{key: key, def: def})
	}

	// Separate locals per job — never write the shared map from goroutines.
	values := make([]int, len(jobs))
	g, gCtx := errgroup.WithContext(ctx)
	for i := range jobs {
		i, job := i, jobs[i]
		g.Go(func() error {
			p, err := s.countProgress(gCtx, job.def, uid, crewID)
			if err != nil {
				p = 0 // soft-error → zero (same as pre-PERF-2)
			}
			values[i] = p
			return nil
		})
	}
	_ = g.Wait() // progress soft-errors never fail the group

	cache := make(map[string]int, len(jobs))
	for i, job := range jobs {
		cache[job.key] = values[i]
	}
	return cache
}

func (s *AchievementService) countProgress(ctx context.Context, def content.AchievementDefinition, uid, crewID string) (int, error) {
	switch AchievementTrigger(def.Trigger) {
	case TriggerQuestCompleted:
		return s.progress.CountCompletedQuests(ctx, crewID)
	case TriggerRealmCompleted:
		return s.progress.CountCompletedRealms(ctx, crewID)
	case TriggerChapterCompleted:
		return s.progress.CountCompletedChapters(ctx, crewID)
	case TriggerRelicCollected:
		return s.progress.CountCollectedRelics(ctx, uid)
	case TriggerDailyStreak:
		return s.progress.CountDailyStreak(ctx, uid)
	case TriggerCreativeSubmission:
		return s.progress.CountCreativeSubmissions(ctx, crewID)
	case TriggerLevelReached:
		level, err := s.progress.GetPlayerLevel(ctx, uid)
		if err != nil {
			return 0, err
		}
		return level, nil
	default:
		return 0, nil
	}
}

func (s *AchievementService) evaluate(ctx context.Context, trigger AchievementTrigger, uid, crewID string) error {
	defs, err := s.defs.ListAchievements(ctx)
	if err != nil {
		return fmt.Errorf("list achievement defs: %w", err)
	}

	for _, def := range defs {
		if AchievementTrigger(def.Trigger) != trigger {
			continue
		}
		existing, err := s.store.GetAchievementByCode(ctx, uid, def.Code)
		if err == nil && existing != nil {
			continue
		}

		progress, err := s.countProgress(ctx, def, uid, crewID)
		if err != nil {
			return fmt.Errorf("count progress for %s: %w", def.Code, err)
		}

		if progress >= def.Threshold {
			now := time.Now().UTC()
			a := &game.Achievement{
				UID:       uid,
				CrewID:    crewID,
				Code:      def.Code,
				Kind:      def.Kind,
				Trigger:   def.Trigger,
				AwardedAt: now,
			}
			_, err := s.store.CreateAchievement(ctx, a)
			if err != nil {
				return fmt.Errorf("create achievement %s: %w", def.Code, err)
			}
		}
	}
	return nil
}

type QuestCompletedHandler struct {
	svc *AchievementService
}

func NewQuestCompletedHandler(svc *AchievementService) *QuestCompletedHandler {
	return &QuestCompletedHandler{svc: svc}
}

func (h *QuestCompletedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.QuestCompletedEvent)
	if !ok {
		return nil
	}
	return h.svc.evaluate(ctx, TriggerQuestCompleted, e.PlayerUID, e.CrewID)
}

type ChapterCompletedHandler struct {
	svc *AchievementService
}

func NewChapterCompletedHandler(svc *AchievementService) *ChapterCompletedHandler {
	return &ChapterCompletedHandler{svc: svc}
}

func (h *ChapterCompletedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.ChapterCompletedEvent)
	if !ok {
		return nil
	}
	return h.svc.evaluate(ctx, TriggerChapterCompleted, e.PlayerUID, e.CrewID)
}

type RelicCollectedHandler struct {
	svc *AchievementService
}

func NewRelicCollectedHandler(svc *AchievementService) *RelicCollectedHandler {
	return &RelicCollectedHandler{svc: svc}
}

func (h *RelicCollectedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.RelicCollectedEvent)
	if !ok {
		return nil
	}
	return h.svc.evaluate(ctx, TriggerRelicCollected, e.UID, e.CrewID)
}

type DailyTurnCompletedHandler struct {
	svc *AchievementService
}

func NewDailyTurnCompletedHandler(svc *AchievementService) *DailyTurnCompletedHandler {
	return &DailyTurnCompletedHandler{svc: svc}
}

func (h *DailyTurnCompletedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.DailyTurnCompletedEvent)
	if !ok {
		return nil
	}
	return h.svc.evaluate(ctx, TriggerDailyStreak, e.UID, "")
}

type LevelReachedHandler struct {
	svc *AchievementService
}

func NewLevelReachedHandler(svc *AchievementService) *LevelReachedHandler {
	return &LevelReachedHandler{svc: svc}
}

func (h *LevelReachedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.LevelReachedEvent)
	if !ok {
		return nil
	}
	return h.svc.evaluate(ctx, TriggerLevelReached, e.UID, e.CrewID)
}

type CreativeSubmissionHandler struct {
	svc *AchievementService
}

func NewCreativeSubmissionHandler(svc *AchievementService) *CreativeSubmissionHandler {
	return &CreativeSubmissionHandler{svc: svc}
}

func (h *CreativeSubmissionHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.CreativeSubmissionEvent)
	if !ok {
		return nil
	}
	return h.svc.evaluate(ctx, TriggerCreativeSubmission, e.UID, e.CrewID)
}
