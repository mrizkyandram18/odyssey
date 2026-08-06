package chapter

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
	"odyssey/pkg/observability"
)

var ErrChapterNotFound = errors.New("chapter not found")
var ErrChapterNotUnlocked = errors.New("chapter not unlocked")

const (
	ChapterStatusLocked   = "LOCKED"
	ChapterStatusActive   = "ACTIVE"
	ChapterStatusComplete = "COMPLETE"
)

type ChapterState string

const (
	StateLocked   ChapterState = ChapterStatusLocked
	StateActive   ChapterState = ChapterStatusActive
	StateComplete ChapterState = ChapterStatusComplete
)

type ChapterSummary struct {
	Definition  content.ChapterDefinition `json:"definition"`
	Status      string                    `json:"status"`
	CompletedAt *time.Time                `json:"completed_at,omitempty"`
}

type ChapterProgressView struct {
	CurrentChapter    *ChapterSummary  `json:"current_chapter,omitempty"`
	NextChapter       *ChapterSummary  `json:"next_chapter,omitempty"`
	CompletedChapters []ChapterSummary `json:"completed_chapters"`
	UnlockedChapters  []ChapterSummary `json:"unlocked_chapters"`
}

type ChapterServiceConfig struct{}

type ChapterGateway interface {
	ListChapters(ctx context.Context) ([]content.ChapterDefinition, error)
	GetChapter(ctx context.Context, slug string) (*content.ChapterDefinition, error)
	ListQuests(ctx context.Context) ([]content.QuestDefinition, error)
	GetQuest(ctx context.Context, slug string) (*content.QuestDefinition, error)
}

type ChapterService struct {
	progress  game.ChapterProgressStore
	quests    game.QuestStore
	content   ChapterGateway
	publisher events.Publisher
	metrics   *observability.Metrics
}

// SetMetrics attaches an optional metrics sink for replay telemetry.
// Safe to call with nil.
func (s *ChapterService) SetMetrics(m *observability.Metrics) {
	s.metrics = m
}

func NewChapterService(
	progress game.ChapterProgressStore,
	quests game.QuestStore,
	content ChapterGateway,
	publisher events.Publisher,
) *ChapterService {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	return &ChapterService{
		progress:  progress,
		quests:    quests,
		content:   content,
		publisher: publisher,
	}
}

func (s *ChapterService) ListByRealm(ctx context.Context, realm string) ([]content.ChapterDefinition, error) {
	chapters, err := s.content.ListChapters(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]content.ChapterDefinition, 0)
	for _, c := range chapters {
		if c.Realm == realm {
			result = append(result, c)
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})
	return result, nil
}

func (s *ChapterService) Get(ctx context.Context, crewID, chapterSlug string) (*ChapterSummary, error) {
	def, err := s.content.GetChapter(ctx, chapterSlug)
	if err != nil {
		return nil, err
	}
	if def == nil {
		return nil, ErrChapterNotFound
	}
	cp, err := s.progress.GetChapterProgress(ctx, crewID, chapterSlug)
	if err != nil || cp == nil {
		return &ChapterSummary{
			Definition: *def,
			Status:     ChapterStatusLocked,
		}, nil
	}
	return &ChapterSummary{
		Definition:  *def,
		Status:      cp.Status,
		CompletedAt: cp.CompletedAt,
	}, nil
}

func (s *ChapterService) ListProgress(ctx context.Context, crewID string) ([]ChapterSummary, error) {
	chapters, err := s.content.ListChapters(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(chapters, func(i, j int) bool {
		return chapters[i].Order < chapters[j].Order
	})

	progressRows, err := s.progress.ListChapterProgressByCrew(ctx, crewID)
	if err != nil {
		progressRows = nil
	}
	progressMap := make(map[string]game.ChapterProgress, len(progressRows))
	for _, p := range progressRows {
		progressMap[p.Chapter] = p
	}

	result := make([]ChapterSummary, 0, len(chapters))
	for _, def := range chapters {
		cp, ok := progressMap[def.Slug]
		if !ok {
			result = append(result, ChapterSummary{
				Definition: def,
				Status:     ChapterStatusLocked,
			})
			continue
		}
		result = append(result, ChapterSummary{
			Definition:  def,
			Status:      cp.Status,
			CompletedAt: cp.CompletedAt,
		})
	}
	return result, nil
}

func (s *ChapterService) GetProgressView(ctx context.Context, crewID string) (*ChapterProgressView, error) {
	summaries, err := s.ListProgress(ctx, crewID)
	if err != nil {
		return nil, err
	}

	completed := make([]ChapterSummary, 0)
	unlocked := make([]ChapterSummary, 0)
	var current *ChapterSummary
	var next *ChapterSummary

	for i := range summaries {
		switch summaries[i].Status {
		case ChapterStatusComplete:
			completed = append(completed, summaries[i])
		case ChapterStatusActive:
			current = &summaries[i]
		case ChapterStatusLocked:
			if next == nil {
				next = &summaries[i]
			}
		}
		if summaries[i].Status == ChapterStatusActive || summaries[i].Status == ChapterStatusComplete {
			unlocked = append(unlocked, summaries[i])
		}
	}

	if current == nil && len(unlocked) > 0 {
		current = &unlocked[len(unlocked)-1]
	}
	if current == nil && len(completed) > 0 {
		current = &completed[len(completed)-1]
	}

	return &ChapterProgressView{
		CurrentChapter:    current,
		NextChapter:       next,
		CompletedChapters: completed,
		UnlockedChapters:  unlocked,
	}, nil
}

func (s *ChapterService) GetCurrentChapter(ctx context.Context, crewID string) (*ChapterSummary, error) {
	view, err := s.GetProgressView(ctx, crewID)
	if err != nil {
		return nil, err
	}
	if view.CurrentChapter == nil {
		all, err := s.ListProgress(ctx, crewID)
		if err != nil {
			return nil, err
		}
		if len(all) > 0 {
			first := all[0]
			return &first, nil
		}
		return nil, nil
	}
	return view.CurrentChapter, nil
}

func (s *ChapterService) UnlockFirstChapter(ctx context.Context, crewID, realm string) error {
	chapters, err := s.ListByRealm(ctx, realm)
	if err != nil {
		return err
	}
	if len(chapters) == 0 {
		return nil
	}
	first := chapters[0]
	_, err = s.progress.CreateChapterProgress(ctx, &game.ChapterProgress{
		CrewID:  crewID,
		Chapter: first.Slug,
		Realm:   realm,
		Status:  ChapterStatusActive,
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("create chapter progress: %w", err)
	}
	return nil
}

func (s *ChapterService) EnsureFirstChapterUnlocked(ctx context.Context, crewID, realm string) error {
	existing, err := s.progress.ListChapterProgressByCrew(ctx, crewID)
	if err == nil && len(existing) > 0 {
		return nil
	}
	return s.UnlockFirstChapter(ctx, crewID, realm)
}

func (s *ChapterService) unlockNextChapter(ctx context.Context, crewID, currentRealm, currentChapter string) error {
	chapters, err := s.ListByRealm(ctx, currentRealm)
	if err != nil {
		return err
	}
	currentIdx := -1
	for i, c := range chapters {
		if c.Slug == currentChapter {
			currentIdx = i
			break
		}
	}
	if currentIdx < 0 || currentIdx+1 >= len(chapters) {
		return nil
	}
	next := chapters[currentIdx+1]

	_, err = s.progress.CreateChapterProgress(ctx, &game.ChapterProgress{
		CrewID:  crewID,
		Chapter: next.Slug,
		Realm:   next.Realm,
		Status:  ChapterStatusActive,
	})
	if err != nil {
		if errors.Is(err, game.ErrNotFound) {
			return nil
		}
		return fmt.Errorf("unlock next chapter: %w", err)
	}
	return nil
}

func (s *ChapterService) CheckCompletion(ctx context.Context, crewID, chapterSlug string) (bool, error) {
	def, err := s.content.GetChapter(ctx, chapterSlug)
	if err != nil {
		return false, err
	}
	if def == nil {
		return false, ErrChapterNotFound
	}

	quests, err := s.content.ListQuests(ctx)
	if err != nil {
		return false, err
	}
	mandatory := make(map[string]bool)
	for _, q := range quests {
		if q.Chapter == chapterSlug && q.IsMandatory {
			mandatory[q.Slug] = true
		}
	}
	if len(mandatory) == 0 {
		return true, nil
	}

	instances, err := s.quests.ListQuestByCrew(ctx, crewID)
	if err != nil {
		return false, err
	}
	completedSlugs := make(map[string]bool)
	for _, q := range instances {
		if q.Status == "DONE" {
			completedSlugs[q.TemplateSlug] = true
		}
	}

	for slug := range mandatory {
		if !completedSlugs[slug] {
			return false, nil
		}
	}
	return true, nil
}

func (s *ChapterService) MarkComplete(ctx context.Context, crewID, chapterSlug, playerUID string) error {
	cp, err := s.progress.GetChapterProgress(ctx, crewID, chapterSlug)
	if err != nil || cp == nil {
		return ErrChapterNotFound
	}
	now := time.Now().UTC()
	patch := map[string]any{
		"status":       ChapterStatusComplete,
		"completed_at": &now,
	}
	if err := s.progress.UpdateChapterProgress(ctx, crewID, chapterSlug, patch); err != nil {
		return fmt.Errorf("mark chapter complete: %w", err)
	}

	def, err := s.content.GetChapter(ctx, chapterSlug)
	if err == nil && def != nil {
		s.publisher.Publish(ctx, events.ChapterCompletedEvent{
			CrewID:     crewID,
			Chapter:    chapterSlug,
			Realm:      def.Realm,
			SeasonSlug: "",
			PlayerUID:  playerUID,
		})
	}
	return nil
}

func (s *ChapterService) handleQuestCompleted(ctx context.Context, event events.Event) error {
	e, ok := event.(events.QuestCompletedEvent)
	if !ok {
		return nil
	}
	chapter := e.Chapter
	if chapter == "" && e.TemplateSlug != "" {
		questDef, qErr := s.content.GetQuest(ctx, e.TemplateSlug)
		if qErr == nil && questDef != nil {
			chapter = questDef.Chapter
		}
	}
	if chapter == "" {
		return nil
	}
	complete, err := s.CheckCompletion(ctx, e.CrewID, chapter)
	if err != nil {
		return err
	}
	if !complete {
		return nil
	}

	cp, err := s.progress.GetChapterProgress(ctx, e.CrewID, chapter)
	if err != nil || cp == nil {
		return nil
	}
	if cp.Status == ChapterStatusComplete {
		if s.metrics != nil {
			s.metrics.RecordReplayIgnored()
		}
		return nil
	}

	if err := s.MarkComplete(ctx, e.CrewID, chapter, e.PlayerUID); err != nil {
		return fmt.Errorf("complete chapter: %w", err)
	}

	return s.unlockNextChapter(ctx, e.CrewID, e.Realm, chapter)
}

type QuestCompletedHandler struct {
	svc *ChapterService
}

func NewQuestCompletedHandler(svc *ChapterService) *QuestCompletedHandler {
	return &QuestCompletedHandler{svc: svc}
}

func (h *QuestCompletedHandler) Handle(ctx context.Context, event events.Event) error {
	return h.svc.handleQuestCompleted(ctx, event)
}
