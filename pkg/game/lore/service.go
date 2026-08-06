package lore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
)

var ErrLoreNotFound = errors.New("lore not found")

type LoreView struct {
	Slug       string `json:"slug"`
	Realm      string `json:"realm"`
	Chapter    string `json:"chapter"`
	Title      string `json:"title"`
	Content    string `json:"content"`
	Order      int    `json:"order"`
	Unlocked   bool   `json:"unlocked"`
	UnlockedAt string `json:"unlocked_at,omitempty"`
}

type LoreSummary struct {
	LockedCount   int       `json:"locked_count"`
	UnlockedCount int       `json:"unlocked_count"`
	Latest        *LoreView `json:"latest,omitempty"`
}

type LoreGateway interface {
	ListLore(ctx context.Context) ([]content.LoreDefinition, error)
	ListLoreByChapter(ctx context.Context, chapter string) ([]content.LoreDefinition, error)
	ListLoreByRealm(ctx context.Context, realm string) ([]content.LoreDefinition, error)
}

type LoreService struct {
	unlocks   game.LoreUnlockStore
	content   LoreGateway
	publisher events.Publisher
}

func NewLoreService(unlocks game.LoreUnlockStore, content LoreGateway) *LoreService {
	return &LoreService{
		unlocks:   unlocks,
		content:   content,
		publisher: events.NopPublisher{},
	}
}

func NewLoreServiceWithPublisher(unlocks game.LoreUnlockStore, content LoreGateway, publisher events.Publisher) *LoreService {
	if publisher == nil {
		publisher = events.NopPublisher{}
	}
	return &LoreService{
		unlocks:   unlocks,
		content:   content,
		publisher: publisher,
	}
}

func (s *LoreService) ListUnlocks(ctx context.Context, crewID string) ([]game.LoreUnlock, error) {
	return s.unlocks.ListLoreUnlocksByCrew(ctx, crewID)
}

func (s *LoreService) GetSummary(ctx context.Context, crewID string) (*LoreSummary, error) {
	all, err := s.content.ListLore(ctx)
	if err != nil {
		return nil, fmt.Errorf("list lore: %w", err)
	}
	unlocks, err := s.unlocks.ListLoreUnlocksByCrew(ctx, crewID)
	if err != nil {
		unlocks = nil
	}
	unlockMap := make(map[string]game.LoreUnlock, len(unlocks))
	for _, u := range unlocks {
		unlockMap[u.LoreSlug] = u
	}

	locked := 0
	unlocked := 0
	var latestSlug string
	var latestTime int64
	for _, ld := range all {
		if u, ok := unlockMap[ld.Slug]; ok {
			unlocked++
			if u.UnlockedAt.Unix() > latestTime {
				latestTime = u.UnlockedAt.Unix()
				latestSlug = ld.Slug
			}
		} else {
			locked++
		}
	}

	summary := &LoreSummary{
		LockedCount:   locked,
		UnlockedCount: unlocked,
	}
	if latestSlug != "" {
		summary.Latest = s.toView(unlockMap[latestSlug], latestSlug, all)
	}
	return summary, nil
}

func (s *LoreService) toView(u game.LoreUnlock, slug string, all []content.LoreDefinition) *LoreView {
	for _, ld := range all {
		if ld.Slug == slug {
			return &LoreView{
				Slug:       ld.Slug,
				Realm:      ld.Realm,
				Chapter:    ld.Chapter,
				Title:      ld.Title,
				Content:    ld.Content,
				Order:      ld.Order,
				Unlocked:   true,
				UnlockedAt: u.UnlockedAt.UTC().Format("2006-01-02T15:04:05Z"),
			}
		}
	}
	return nil
}

func (s *LoreService) defToView(ld content.LoreDefinition, u *game.LoreUnlock) LoreView {
	v := LoreView{
		Slug:     ld.Slug,
		Realm:    ld.Realm,
		Chapter:  ld.Chapter,
		Title:    ld.Title,
		Content:  ld.Content,
		Order:    ld.Order,
		Unlocked: false,
	}
	if u != nil {
		v.Unlocked = true
		v.UnlockedAt = u.UnlockedAt.UTC().Format("2006-01-02T15:04:05Z")
	}
	return v
}

func (s *LoreService) ListUnlocked(ctx context.Context, crewID string) ([]LoreView, error) {
	all, err := s.content.ListLore(ctx)
	if err != nil {
		return nil, err
	}
	unlocks, err := s.unlocks.ListLoreUnlocksByCrew(ctx, crewID)
	if err != nil {
		unlocks = nil
	}
	unlockMap := make(map[string]game.LoreUnlock, len(unlocks))
	for _, u := range unlocks {
		unlockMap[u.LoreSlug] = u
	}

	result := make([]LoreView, 0)
	for _, ld := range all {
		if u, ok := unlockMap[ld.Slug]; ok {
			result = append(result, s.defToView(ld, &u))
		}
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Order < result[j].Order
	})
	return result, nil
}

func (s *LoreService) UnlockForChapter(ctx context.Context, crewID, chapter string) ([]LoreView, error) {
	loreDefs, err := s.content.ListLoreByChapter(ctx, chapter)
	if err != nil {
		return nil, fmt.Errorf("list lore by chapter: %w", err)
	}
	now := time.Now().UTC()
	unlocked := make([]LoreView, 0)
	for _, ld := range loreDefs {
		if _, err := s.unlocks.GetLoreUnlock(ctx, crewID, ld.Slug); err == nil {
			continue
		}
		lu, err := s.unlocks.CreateLoreUnlock(ctx, &game.LoreUnlock{
			CrewID:     crewID,
			LoreSlug:   ld.Slug,
			Realm:      ld.Realm,
			Chapter:    chapter,
			UnlockedAt: now,
			CreatedAt:  now,
		})
		if err != nil {
			return nil, fmt.Errorf("unlock lore %s: %w", ld.Slug, err)
		}
		unlocked = append(unlocked, s.defToView(ld, lu))
	}
	return unlocked, nil
}

type ChapterCompletedHandler struct {
	svc *LoreService
}

func NewChapterCompletedHandler(svc *LoreService) *ChapterCompletedHandler {
	return &ChapterCompletedHandler{svc: svc}
}

func (h *ChapterCompletedHandler) Handle(ctx context.Context, event events.Event) error {
	e, ok := event.(events.ChapterCompletedEvent)
	if !ok {
		return nil
	}
	_, err := h.svc.UnlockForChapter(ctx, e.CrewID, e.Chapter)
	return err
}
