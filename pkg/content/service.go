package content

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/content"
)

var ErrNotFound = errors.New("content not found")

// CacheEntry holds cached data with an expiry timestamp and generation.
type CacheEntry struct {
	Data       any
	ExpiresAt  time.Time
	Generation int64
}

// CacheStats tracks cache hit/miss counters for observability.
type CacheStats struct {
	Hits      int64
	Misses    int64
	Evictions int64
}

// Cache is a generation-aware, TTL-based in-memory cache with statistics.
type Cache struct {
	mu         sync.RWMutex
	entries    map[string]CacheEntry
	ttl        time.Duration
	generation int64
	stats      CacheStats
}

func NewCache(ttl time.Duration) *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
		ttl:     ttl,
	}
}

func (c *Cache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.entries[key]
	if !ok {
		atomic.AddInt64(&c.stats.Misses, 1)
		return nil, false
	}
	if time.Now().After(entry.ExpiresAt) {
		atomic.AddInt64(&c.stats.Evictions, 1)
		return nil, false
	}
	atomic.AddInt64(&c.stats.Hits, 1)
	return entry.Data, true
}

func (c *Cache) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = CacheEntry{
		Data:       value,
		ExpiresAt:  time.Now().Add(c.ttl),
		Generation: c.generation,
	}
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, existed := c.entries[key]
	delete(c.entries, key)
	if existed {
		atomic.AddInt64(&c.stats.Evictions, 1)
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := len(c.entries)
	c.entries = make(map[string]CacheEntry)
	c.generation++
	if count > 0 {
		atomic.AddInt64(&c.stats.Evictions, int64(count))
	}
}

func (c *Cache) Stats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return CacheStats{
		Hits:      atomic.LoadInt64(&c.stats.Hits),
		Misses:    atomic.LoadInt64(&c.stats.Misses),
		Evictions: atomic.LoadInt64(&c.stats.Evictions),
	}
}

func (c *Cache) HitRatio() float64 {
	stats := c.Stats()
	total := stats.Hits + stats.Misses
	if total == 0 {
		return 0
	}
	return float64(stats.Hits) / float64(total)
}

func (c *Cache) Generation() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.generation
}

func (c *Cache) Invalidate(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	_, ok := c.entries[key]
	if ok {
		delete(c.entries, key)
	}
	return ok
}

// ContentServiceConfig configures the ContentService cache behavior.
type ContentServiceConfig struct {
	CacheTTL          time.Duration
	BackgroundRefresh bool
	RefreshInterval   time.Duration
}

// ContentService is the central caching layer for all game content definitions.
// It loads only published definitions and provides admin operations for
// draft preview, draft save, publish, and soft-delete restore.
type ContentService struct {
	realmStore        content.RealmDefinitionStore
	chapterStore      content.ChapterDefinitionStore
	questStore        content.QuestDefinitionStore
	promptStore       content.CreativePromptStore
	achievementStore  content.AchievementDefinitionStore
	seasonStore       content.SeasonDefinitionStore
	loreStore         content.LoreDefinitionStore
	chestStore        game.ChestDefinitionStore
	relicStore        game.RelicDefinitionStore
	adminStore        content.AdminStore
	cache             *Cache
	mu                sync.RWMutex
	backgroundRefresh bool
	refreshInterval   time.Duration
	stopRefresh       chan struct{}
}

func NewContentService(
	realmStore content.RealmDefinitionStore,
	chapterStore content.ChapterDefinitionStore,
	questStore content.QuestDefinitionStore,
	promptStore content.CreativePromptStore,
	achievementStore content.AchievementDefinitionStore,
	seasonStore content.SeasonDefinitionStore,
	loreStore content.LoreDefinitionStore,
	chestStore game.ChestDefinitionStore,
	relicStore game.RelicDefinitionStore,
	cfg ContentServiceConfig,
) *ContentService {
	if cfg.CacheTTL <= 0 {
		cfg.CacheTTL = 5 * time.Minute
	}
	if cfg.RefreshInterval <= 0 {
		cfg.RefreshInterval = cfg.CacheTTL / 2
	}
	svc := &ContentService{
		realmStore:        realmStore,
		chapterStore:      chapterStore,
		questStore:        questStore,
		promptStore:       promptStore,
		achievementStore:  achievementStore,
		seasonStore:       seasonStore,
		loreStore:         loreStore,
		chestStore:        chestStore,
		relicStore:        relicStore,
		cache:             NewCache(cfg.CacheTTL),
		backgroundRefresh: cfg.BackgroundRefresh,
		refreshInterval:   cfg.RefreshInterval,
	}
	if cfg.BackgroundRefresh {
		svc.startBackgroundRefresh()
	}
	return svc
}

func (s *ContentService) SetAdminStore(store content.AdminStore) {
	s.adminStore = store
}

func (s *ContentService) startBackgroundRefresh() {
	s.stopRefresh = make(chan struct{})
	go s.backgroundRefreshLoop()
}

func (s *ContentService) backgroundRefreshLoop() {
	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.backgroundReload(context.Background())
		case <-s.stopRefresh:
			return
		}
	}
}

func (s *ContentService) backgroundReload(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	_ = s.reload(ctx)
}

func (s *ContentService) Stop() {
	if s.stopRefresh != nil {
		close(s.stopRefresh)
	}
}

// CacheStats returns current cache statistics.
func (s *ContentService) CacheStats() CacheStats {
	return s.cache.Stats()
}

func (s *ContentService) CacheHitRatio() float64 {
	return s.cache.HitRatio()
}

// CacheGeneration returns the current cache generation number.
func (s *ContentService) CacheGeneration() int64 {
	return s.cache.Generation()
}

// Status returns a summary of content counts and cache statistics.
func (s *ContentService) Status(ctx context.Context) (map[string]any, error) {
	status := map[string]any{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	if s.realmStore != nil {
		realms, err := s.realmStore.ListRealms(ctx)
		if err != nil {
			return nil, err
		}
		status["realms"] = len(realms)
	}
	if s.chapterStore != nil {
		chapters, err := s.chapterStore.ListChapters(ctx, "")
		if err != nil {
			return nil, err
		}
		status["chapters"] = len(chapters)
	}
	if s.questStore != nil {
		missions, err := s.questStore.ListQuests(ctx)
		if err != nil {
			return nil, err
		}
		status["missions"] = len(missions)
	}
	if s.promptStore != nil {
		prompts, err := s.promptStore.ListPrompts(ctx)
		if err != nil {
			return nil, err
		}
		status["prompts"] = len(prompts)
	}
	if s.achievementStore != nil {
		achievements, err := s.achievementStore.ListAchievements(ctx)
		if err != nil {
			return nil, err
		}
		status["achievements"] = len(achievements)
	}
	if s.seasonStore != nil {
		seasons, err := s.seasonStore.ListSeasons(ctx)
		if err != nil {
			return nil, err
		}
		status["seasons"] = len(seasons)
	}
	if s.loreStore != nil {
		concept, err := s.loreStore.ListLore(ctx)
		if err != nil {
			return nil, err
		}
		status["concept"] = len(concept)
	}
	if s.chestStore != nil {
		gifts, err := s.chestStore.ListChestDefinitions(ctx)
		if err != nil {
			return nil, err
		}
		status["gifts"] = len(gifts)
	}
	if s.relicStore != nil {
		collections, err := s.relicStore.ListRelicDefinitions(ctx)
		if err != nil {
			return nil, err
		}
		status["collections"] = len(collections)
	}
	status["cache_generation"] = s.cache.Generation()
	status["cache_stats"] = s.cache.Stats()
	status["cache_hit_ratio"] = s.cache.HitRatio()
	return status, nil
}

// Invalidate removes a specific cache entry. If the key is empty,
// clears the entire cache.
func (s *ContentService) Invalidate(key string) {
	if key == "" {
		s.cache.Clear()
		return
	}
	s.cache.Invalidate(key)
}

// InvalidateResource clears all cache entries for a given content type.
func (s *ContentService) InvalidateResource(resourceType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache.Clear()
}

func (s *ContentService) Reload(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.reload(ctx)
}

func (s *ContentService) reload(ctx context.Context) error {
	s.cache.Clear()

	if s.realmStore != nil {
		realms, err := s.realmStore.ListRealms(ctx)
		if err != nil {
			return err
		}
		for _, r := range realms {
			s.cache.setRealm(r)
		}
	}

	if s.chapterStore != nil {
		chapters, err := s.chapterStore.ListChapters(ctx, "")
		if err != nil {
			return err
		}
		for _, c := range chapters {
			s.cache.setChapter(c)
		}
	}

	if s.questStore != nil {
		missions, err := s.questStore.ListQuests(ctx)
		if err != nil {
			return err
		}
		for _, q := range missions {
			s.cache.setQuest(q)
		}
	}

	if s.promptStore != nil {
		prompts, err := s.promptStore.ListPrompts(ctx)
		if err != nil {
			return err
		}
		for _, p := range prompts {
			s.cache.setPrompt(p)
		}
	}

	if s.achievementStore != nil {
		achievements, err := s.achievementStore.ListAchievements(ctx)
		if err != nil {
			return err
		}
		for _, a := range achievements {
			s.cache.setAchievement(a)
		}
	}

	if s.seasonStore != nil {
		seasons, err := s.seasonStore.ListSeasons(ctx)
		if err != nil {
			return err
		}
		for _, sd := range seasons {
			s.cache.setSeason(sd)
		}
	}

	if s.loreStore != nil {
		concept, err := s.loreStore.ListLore(ctx)
		if err != nil {
			return err
		}
		for _, l := range concept {
			s.cache.setLore(l)
		}
	}

	if s.chestStore != nil {
		gifts, err := s.chestStore.ListChestDefinitions(ctx)
		if err != nil {
			return err
		}
		result := make([]content.ChestDefinition, 0, len(gifts))
		for _, d := range gifts {
			result = append(result, content.ChestDefinition{
				ID:          d.ID,
				Slug:        d.Slug,
				Name:        d.Name,
				Rarity:      d.Rarity,
				Icon:        d.Icon,
				Description: d.Description,
				SeasonSlug:  d.SeasonSlug,
				Published:   d.Published,
				Version:     d.Version,
				UpdatedBy:   d.UpdatedBy,
				DeletedAt:   d.DeletedAt,
				CreatedAt:   d.CreatedAt,
				UpdatedAt:   d.UpdatedAt,
			})
		}
		s.cache.setChests(result)
	}

	if s.relicStore != nil {
		collections, err := s.relicStore.ListRelicDefinitions(ctx)
		if err != nil {
			return err
		}
		result := make([]content.RelicDefinition, 0, len(collections))
		for _, d := range collections {
			result = append(result, content.RelicDefinition{
				ID:          d.ID,
				Slug:        d.Slug,
				Name:        d.Name,
				Description: d.Description,
				Journey:       d.Journey,
				Rarity:      d.Rarity,
				Image:       d.Image,
				Concept:        d.Concept,
				Published:   d.Published,
				Version:     d.Version,
				UpdatedBy:   d.UpdatedBy,
				DeletedAt:   d.DeletedAt,
				CreatedAt:   d.CreatedAt,
				UpdatedAt:   d.UpdatedAt,
			})
		}
		s.cache.setRelics(result)
	}

	return nil
}

// ReloadResource performs a partial reload of a single content type.
func (s *ContentService) ReloadResource(ctx context.Context, resourceType string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache.Invalidate("realms:all")
	switch resourceType {
	case "realms":
		s.cache.Invalidate("realms:all")
		if s.realmStore != nil {
			realms, err := s.realmStore.ListRealms(ctx)
			if err != nil {
				return err
			}
			for _, r := range realms {
				s.cache.setRealm(r)
			}
		}
	case "chapters":
		s.cache.Invalidate("chapters:all")
		if s.chapterStore != nil {
			chapters, err := s.chapterStore.ListChapters(ctx, "")
			if err != nil {
				return err
			}
			for _, c := range chapters {
				s.cache.setChapter(c)
			}
		}
	case "missions":
		s.cache.Invalidate("missions:all")
		if s.questStore != nil {
			missions, err := s.questStore.ListQuests(ctx)
			if err != nil {
				return err
			}
			for _, q := range missions {
				s.cache.setQuest(q)
			}
		}
	case "prompts":
		s.cache.Invalidate("prompts:all")
		if s.promptStore != nil {
			prompts, err := s.promptStore.ListPrompts(ctx)
			if err != nil {
				return err
			}
			for _, p := range prompts {
				s.cache.setPrompt(p)
			}
		}
	case "achievements":
		s.cache.Invalidate("achievements:all")
		if s.achievementStore != nil {
			achievements, err := s.achievementStore.ListAchievements(ctx)
			if err != nil {
				return err
			}
			for _, a := range achievements {
				s.cache.setAchievement(a)
			}
		}
	case "seasons":
		s.cache.Invalidate("seasons:all")
		if s.seasonStore != nil {
			seasons, err := s.seasonStore.ListSeasons(ctx)
			if err != nil {
				return err
			}
			for _, sd := range seasons {
				s.cache.setSeason(sd)
			}
		}
	case "concept":
		s.cache.Invalidate("concept:all")
		if s.loreStore != nil {
			concept, err := s.loreStore.ListLore(ctx)
			if err != nil {
				return err
			}
			for _, l := range concept {
				s.cache.setLore(l)
			}
		}
	case "gifts":
		s.cache.Invalidate("gifts:all")
		if s.chestStore != nil {
			gifts, err := s.chestStore.ListChestDefinitions(ctx)
			if err != nil {
				return err
			}
			result := make([]content.ChestDefinition, 0, len(gifts))
			for _, d := range gifts {
				result = append(result, content.ChestDefinition{
					ID:          d.ID,
					Slug:        d.Slug,
					Name:        d.Name,
					Rarity:      d.Rarity,
					Icon:        d.Icon,
					Description: d.Description,
					SeasonSlug:  d.SeasonSlug,
					Published:   d.Published,
					Version:     d.Version,
					UpdatedBy:   d.UpdatedBy,
					DeletedAt:   d.DeletedAt,
					CreatedAt:   d.CreatedAt,
					UpdatedAt:   d.UpdatedAt,
				})
			}
			s.cache.setChests(result)
		}
	case "collections":
		s.cache.Invalidate("collections:all")
		if s.relicStore != nil {
			collections, err := s.relicStore.ListRelicDefinitions(ctx)
			if err != nil {
				return err
			}
			result := make([]content.RelicDefinition, 0, len(collections))
			for _, d := range collections {
				result = append(result, content.RelicDefinition{
					ID:          d.ID,
					Slug:        d.Slug,
					Name:        d.Name,
					Description: d.Description,
					Journey:       d.Journey,
					Rarity:      d.Rarity,
					Image:       d.Image,
					Concept:        d.Concept,
					Published:   d.Published,
					Version:     d.Version,
					UpdatedBy:   d.UpdatedBy,
					DeletedAt:   d.DeletedAt,
					CreatedAt:   d.CreatedAt,
					UpdatedAt:   d.UpdatedAt,
				})
			}
			s.cache.setRelics(result)
		}
	}
	return nil
}

func (s *ContentService) GetRealm(ctx context.Context, slug string) (*content.RealmDefinition, error) {
	key := "journey:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.RealmDefinition); ok {
			return def, nil
		}
	}
	def, err := s.realmStore.GetRealm(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		s.cache.setRealm(*def)
	}
	return def, nil
}

func (s *ContentService) ListRealms(ctx context.Context) ([]content.RealmDefinition, error) {
	key := "realms:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.RealmDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.realmStore.ListRealms(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.setRealms(defs)
	return defs, nil
}

func (s *ContentService) GetChapter(ctx context.Context, slug string) (*content.ChapterDefinition, error) {
	key := "course:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.ChapterDefinition); ok {
			return def, nil
		}
	}
	def, err := s.chapterStore.GetChapter(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		s.cache.setChapter(*def)
	}
	return def, nil
}

func (s *ContentService) ListChaptersByRealm(ctx context.Context, journey string) ([]content.ChapterDefinition, error) {
	key := "chapters:journey:" + journey
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.ChapterDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.chapterStore.ListChapters(ctx, journey)
	if err != nil {
		return nil, err
	}
	s.cache.setChapters(defs)
	return defs, nil
}

func (s *ContentService) ListChapters(ctx context.Context) ([]content.ChapterDefinition, error) {
	key := "chapters:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.ChapterDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.chapterStore.ListChapters(ctx, "")
	if err != nil {
		return nil, err
	}
	s.cache.setChapters(defs)
	return defs, nil
}

func (s *ContentService) ActiveSeasons(ctx context.Context) ([]content.SeasonDefinition, error) {
	all, err := s.ListSeasons(ctx)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	active := make([]content.SeasonDefinition, 0)
	for _, sd := range all {
		if !sd.StartAt.After(now) && !sd.EndAt.Before(now) {
			active = append(active, sd)
		}
	}
	return active, nil
}

func (s *ContentService) activeSeasonSlugs(ctx context.Context) (map[string]bool, error) {
	seasons, err := s.ActiveSeasons(ctx)
	if err != nil {
		return nil, err
	}
	result := make(map[string]bool, len(seasons))
	for _, sd := range seasons {
		result[sd.Slug] = true
	}
	return result, nil
}

func (s *ContentService) IsSeasonActive(ctx context.Context, slug string) bool {
	if slug == "" {
		return true
	}
	seasons, err := s.ActiveSeasons(ctx)
	if err != nil || len(seasons) == 0 {
		return true
	}
	for _, sd := range seasons {
		if sd.Slug == slug {
			return true
		}
	}
	return false
}

func (s *ContentService) GetQuest(ctx context.Context, slug string) (*content.QuestDefinition, error) {
	key := "quest:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.QuestDefinition); ok {
			return def, nil
		}
	}
	def, err := s.questStore.GetQuest(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		s.cache.setQuest(*def)
	}
	return def, nil
}

func (s *ContentService) ListQuests(ctx context.Context) ([]content.QuestDefinition, error) {
	key := "missions:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.QuestDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.questStore.ListQuests(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.setQuests(defs)
	return defs, nil
}

func (s *ContentService) ListQuestsByRealm(ctx context.Context, journey string) ([]content.QuestDefinition, error) {
	key := "missions:journey:" + journey
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.QuestDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.questStore.ListQuestsByRealm(ctx, journey)
	if err != nil {
		return nil, err
	}
	s.cache.setQuests(defs)
	return defs, nil
}

func (s *ContentService) GetPrompt(ctx context.Context, slug string) (*content.CreativePromptDefinition, error) {
	key := "prompt:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.CreativePromptDefinition); ok {
			return def, nil
		}
	}
	def, err := s.promptStore.GetPrompt(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		s.cache.setPrompt(*def)
	}
	return def, nil
}

func (s *ContentService) ListPrompts(ctx context.Context) ([]content.CreativePromptDefinition, error) {
	key := "prompts:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.CreativePromptDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.promptStore.ListPrompts(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.setPrompts(defs)
	return defs, nil
}

func (s *ContentService) ListPromptsByRealm(ctx context.Context, journey string) ([]content.CreativePromptDefinition, error) {
	key := "prompts:journey:" + journey
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.CreativePromptDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.promptStore.ListPromptsByRealm(ctx, journey)
	if err != nil {
		return nil, err
	}
	s.cache.setPrompts(defs)
	return defs, nil
}

func (s *ContentService) GetAchievement(ctx context.Context, code string) (*content.AchievementDefinition, error) {
	key := "achievement:" + code
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.AchievementDefinition); ok {
			return def, nil
		}
	}
	def, err := s.achievementStore.GetAchievement(ctx, code)
	if err != nil {
		return nil, err
	}
	if def != nil {
		s.cache.setAchievement(*def)
	}
	return def, nil
}

func (s *ContentService) ListAchievements(ctx context.Context) ([]content.AchievementDefinition, error) {
	key := "achievements:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.AchievementDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.achievementStore.ListAchievements(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.setAchievements(defs)
	return defs, nil
}

func (s *ContentService) GetSeason(ctx context.Context, slug string) (*content.SeasonDefinition, error) {
	key := "season:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.SeasonDefinition); ok {
			return def, nil
		}
	}
	def, err := s.seasonStore.GetSeason(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		s.cache.setSeason(*def)
	}
	return def, nil
}

func (s *ContentService) ListSeasons(ctx context.Context) ([]content.SeasonDefinition, error) {
	key := "seasons:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.SeasonDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.seasonStore.ListSeasons(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.setSeasons(defs)
	return defs, nil
}

func (s *ContentService) GetLore(ctx context.Context, slug string) (*content.LoreDefinition, error) {
	key := "concept:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.LoreDefinition); ok {
			return def, nil
		}
	}
	def, err := s.loreStore.GetLore(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		s.cache.setLore(*def)
	}
	return def, nil
}

func (s *ContentService) ListLore(ctx context.Context) ([]content.LoreDefinition, error) {
	key := "concept:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.LoreDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.loreStore.ListLore(ctx)
	if err != nil {
		return nil, err
	}
	s.cache.setLoreEntries(defs)
	return defs, nil
}

func (s *ContentService) ListLoreByRealm(ctx context.Context, journey string) ([]content.LoreDefinition, error) {
	key := "concept:journey:" + journey
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.LoreDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.loreStore.ListLoreByRealm(ctx, journey)
	if err != nil {
		return nil, err
	}
	s.cache.setLoreEntries(defs)
	return defs, nil
}

func (s *ContentService) ListLoreByChapter(ctx context.Context, course string) ([]content.LoreDefinition, error) {
	key := "concept:course:" + course
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.LoreDefinition); ok {
			return defs, nil
		}
	}
	all, err := s.ListLore(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]content.LoreDefinition, 0)
	for _, ld := range all {
		if ld.Course == course {
			result = append(result, ld)
		}
	}
	s.cache.Set(key, result)
	return result, nil
}

func (s *ContentService) ListChests(ctx context.Context) ([]content.ChestDefinition, error) {
	key := "gifts:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.ChestDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.chestStore.ListChestDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]content.ChestDefinition, 0, len(defs))
	for _, d := range defs {
		result = append(result, content.ChestDefinition{
			ID:          d.ID,
			Slug:        d.Slug,
			Name:        d.Name,
			Rarity:      d.Rarity,
			Icon:        d.Icon,
			Description: d.Description,
			SeasonSlug:  d.SeasonSlug,
			Published:   d.Published,
			Version:     d.Version,
			UpdatedBy:   d.UpdatedBy,
			DeletedAt:   d.DeletedAt,
			CreatedAt:   d.CreatedAt,
			UpdatedAt:   d.UpdatedAt,
		})
	}
	s.cache.setChests(result)
	return result, nil
}

func (s *ContentService) GetChest(ctx context.Context, slug string) (*content.ChestDefinition, error) {
	key := "chest:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.ChestDefinition); ok {
			return def, nil
		}
	}
	def, err := s.chestStore.GetChestDefinition(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		cd := content.ChestDefinition{
			ID:          def.ID,
			Slug:        def.Slug,
			Name:        def.Name,
			Rarity:      def.Rarity,
			Icon:        def.Icon,
			Description: def.Description,
			SeasonSlug:  def.SeasonSlug,
			Published:   def.Published,
			Version:     def.Version,
			UpdatedBy:   def.UpdatedBy,
			DeletedAt:   def.DeletedAt,
			CreatedAt:   def.CreatedAt,
			UpdatedAt:   def.UpdatedAt,
		}
		s.cache.setChest(cd)
		return &cd, nil
	}
	return nil, nil
}

func (s *ContentService) ListRelics(ctx context.Context) ([]content.RelicDefinition, error) {
	key := "collections:all"
	if v, ok := s.cache.Get(key); ok {
		if defs, ok := v.([]content.RelicDefinition); ok {
			return defs, nil
		}
	}
	defs, err := s.relicStore.ListRelicDefinitions(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]content.RelicDefinition, 0, len(defs))
	for _, d := range defs {
		rd := content.RelicDefinition{
			ID:          d.ID,
			Slug:        d.Slug,
			Name:        d.Name,
			Description: d.Description,
			Journey:       d.Journey,
			Rarity:      d.Rarity,
			Image:       d.Image,
			Concept:        d.Concept,
			Published:   d.Published,
			Version:     d.Version,
			UpdatedBy:   d.UpdatedBy,
			DeletedAt:   d.DeletedAt,
			CreatedAt:   d.CreatedAt,
			UpdatedAt:   d.UpdatedAt,
		}
		result = append(result, rd)
	}
	s.cache.setRelics(result)
	return result, nil
}

func (s *ContentService) GetRelic(ctx context.Context, slug string) (*content.RelicDefinition, error) {
	key := "relic:" + slug
	if v, ok := s.cache.Get(key); ok {
		if def, ok := v.(*content.RelicDefinition); ok {
			return def, nil
		}
	}
	def, err := s.relicStore.GetRelicDefinition(ctx, slug)
	if err != nil {
		return nil, err
	}
	if def != nil {
		rd := content.RelicDefinition{
			ID:          def.ID,
			Slug:        def.Slug,
			Name:        def.Name,
			Description: def.Description,
			Journey:       def.Journey,
			Rarity:      def.Rarity,
			Image:       def.Image,
			Concept:        def.Concept,
			Published:   def.Published,
			Version:     def.Version,
			UpdatedBy:   def.UpdatedBy,
			DeletedAt:   def.DeletedAt,
			CreatedAt:   def.CreatedAt,
			UpdatedAt:   def.UpdatedAt,
		}
		s.cache.setRelic(rd)
		return &rd, nil
	}
	return nil, nil
}

// GetDraft retrieves a draft definition by slug, bypassing the published filter.
// Returns nil if no draft exists and the definition is not found.
func (s *ContentService) GetDraft(ctx context.Context, table, slug string) (map[string]any, error) {
	if s.adminStore == nil {
		return nil, fmt.Errorf("admin store not configured")
	}
	return s.adminStore.GetBySlug(ctx, table, slug)
}

// SaveDraft saves a draft of a definition by writing the provided patch to the
// row's draft columns, incrementing version and setting updated_by.
func (s *ContentService) SaveDraft(ctx context.Context, table, slug string, patch map[string]any, updatedBy string) error {
	if s.adminStore == nil {
		return fmt.Errorf("admin store not configured")
	}
	return s.adminStore.UpdateDraft(ctx, table, slug, patch, updatedBy)
}

// CreateDraft creates a new definition in draft state (published=false).
func (s *ContentService) CreateDraft(ctx context.Context, table string, data map[string]any) (map[string]any, error) {
	if s.adminStore == nil {
		return nil, fmt.Errorf("admin store not configured")
	}
	return s.adminStore.Create(ctx, table, data)
}

// Publish promotes a draft to published status.
func (s *ContentService) Publish(ctx context.Context, table, slug, updatedBy string) error {
	if s.adminStore == nil {
		return fmt.Errorf("admin store not configured")
	}
	if err := s.adminStore.Publish(ctx, table, slug, updatedBy); err != nil {
		return err
	}
	_ = s.ReloadResource(ctx, tableToResource(table))
	return nil
}

// SoftDelete marks a definition as deleted (hidden from players).
func (s *ContentService) SoftDelete(ctx context.Context, table, slug string) error {
	if s.adminStore == nil {
		return fmt.Errorf("admin store not configured")
	}
	if err := s.adminStore.SoftDelete(ctx, table, slug); err != nil {
		return err
	}
	_ = s.ReloadResource(ctx, tableToResource(table))
	return nil
}

// Restore clears the deleted_at flag, making a definition visible again.
func (s *ContentService) Restore(ctx context.Context, table, slug string) error {
	if s.adminStore == nil {
		return fmt.Errorf("admin store not configured")
	}
	if err := s.adminStore.Restore(ctx, table, slug); err != nil {
		return err
	}
	_ = s.ReloadResource(ctx, tableToResource(table))
	return nil
}

func tableToResource(table string) string {
	switch table {
	case content.TableRealms:
		return "realms"
	case content.TableChapters:
		return "chapters"
	case content.TableQuests:
		return "missions"
	case content.TablePrompts:
		return "prompts"
	case content.TableChests:
		return "gifts"
	case content.TableDropTables:
		return "droptables"
	case content.TableRelics:
		return "collections"
	case content.TableAchievements:
		return "achievements"
	case content.TableSeasons:
		return "seasons"
	case content.TableLore:
		return "concept"
	default:
		return table
	}
}

func (c *Cache) setChest(def content.ChestDefinition) {
	c.Set("chest:"+def.Slug, def)
}

func (c *Cache) setChests(defs []content.ChestDefinition) {
	c.Set("gifts:all", defs)
}

func (c *Cache) setRelic(def content.RelicDefinition) {
	c.Set("relic:"+def.Slug, def)
}

func (c *Cache) setRelics(defs []content.RelicDefinition) {
	c.Set("collections:all", defs)
}

func (c *Cache) setRealm(def content.RealmDefinition) {
	c.Set("journey:"+def.Slug, def)
}

func (c *Cache) setRealms(defs []content.RealmDefinition) {
	c.Set("realms:all", defs)
}

func (c *Cache) setChapter(def content.ChapterDefinition) {
	c.Set("course:"+def.Slug, def)
}

func (c *Cache) setChapters(defs []content.ChapterDefinition) {
	c.Set("chapters:all", defs)
}

func (c *Cache) setQuest(def content.QuestDefinition) {
	c.Set("quest:"+def.Slug, def)
}

func (c *Cache) setQuests(defs []content.QuestDefinition) {
	c.Set("missions:all", defs)
}

func (c *Cache) setPrompt(def content.CreativePromptDefinition) {
	c.Set("prompt:"+def.Slug, def)
}

func (c *Cache) setPrompts(defs []content.CreativePromptDefinition) {
	c.Set("prompts:all", defs)
}

func (c *Cache) setAchievement(def content.AchievementDefinition) {
	c.Set("achievement:"+def.Code, def)
}

func (c *Cache) setAchievements(defs []content.AchievementDefinition) {
	c.Set("achievements:all", defs)
}

func (c *Cache) setSeason(def content.SeasonDefinition) {
	c.Set("season:"+def.Slug, def)
}

func (c *Cache) setSeasons(defs []content.SeasonDefinition) {
	c.Set("seasons:all", defs)
}

func (c *Cache) setLore(def content.LoreDefinition) {
	c.Set("concept:"+def.Slug, def)
}

func (c *Cache) setLoreEntries(defs []content.LoreDefinition) {
	c.Set("concept:all", defs)
}
