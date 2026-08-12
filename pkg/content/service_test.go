package content

import (
	"context"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
)

type mockDefinitionStore struct {
	realms   []gamecontent.RealmDefinition
	chapters []gamecontent.ChapterDefinition
	missions   []gamecontent.QuestDefinition
}

func (m *mockDefinitionStore) ListRealms(ctx context.Context) ([]gamecontent.RealmDefinition, error) {
	return m.realms, nil
}
func (m *mockDefinitionStore) GetRealm(ctx context.Context, slug string) (*gamecontent.RealmDefinition, error) {
	for _, r := range m.realms {
		if r.Slug == slug {
			return &r, nil
		}
	}
	return nil, nil
}
func (m *mockDefinitionStore) ListChapters(ctx context.Context, journey string) ([]gamecontent.ChapterDefinition, error) {
	if journey == "" {
		return m.chapters, nil
	}
	var result []gamecontent.ChapterDefinition
	for _, c := range m.chapters {
		if c.Journey == journey {
			result = append(result, c)
		}
	}
	return result, nil
}
func (m *mockDefinitionStore) GetChapter(ctx context.Context, slug string) (*gamecontent.ChapterDefinition, error) {
	for _, c := range m.chapters {
		if c.Slug == slug {
			return &c, nil
		}
	}
	return nil, nil
}
func (m *mockDefinitionStore) ListChaptersByRealm(ctx context.Context, journey string) ([]gamecontent.ChapterDefinition, error) {
	return m.ListChapters(ctx, journey)
}

type mockQuestStore struct {
	missions []gamecontent.QuestDefinition
}

func (m *mockQuestStore) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.missions, nil
}
func (m *mockQuestStore) ListQuestsByRealm(ctx context.Context, journey string) ([]gamecontent.QuestDefinition, error) {
	var result []gamecontent.QuestDefinition
	for _, q := range m.missions {
		if q.Journey == journey {
			result = append(result, q)
		}
	}
	return result, nil
}
func (m *mockQuestStore) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	for _, q := range m.missions {
		if q.Slug == slug {
			return &q, nil
		}
	}
	return nil, nil
}

type mockPromptStore struct{}

func (m *mockPromptStore) ListPrompts(ctx context.Context) ([]gamecontent.CreativePromptDefinition, error) {
	return nil, nil
}
func (m *mockPromptStore) ListPromptsByRealm(ctx context.Context, journey string) ([]gamecontent.CreativePromptDefinition, error) {
	return nil, nil
}
func (m *mockPromptStore) GetPrompt(ctx context.Context, slug string) (*gamecontent.CreativePromptDefinition, error) {
	return nil, nil
}

type mockAchievementStore struct{}

func (m *mockAchievementStore) ListAchievements(ctx context.Context) ([]gamecontent.AchievementDefinition, error) {
	return nil, nil
}
func (m *mockAchievementStore) GetAchievement(ctx context.Context, code string) (*gamecontent.AchievementDefinition, error) {
	return nil, nil
}

type mockSeasonStore struct{}

func (m *mockSeasonStore) ListSeasons(ctx context.Context) ([]gamecontent.SeasonDefinition, error) {
	return nil, nil
}
func (m *mockSeasonStore) GetSeason(ctx context.Context, slug string) (*gamecontent.SeasonDefinition, error) {
	return nil, nil
}

type mockLoreStore struct{}

func (m *mockLoreStore) ListLore(ctx context.Context) ([]gamecontent.LoreDefinition, error) {
	return nil, nil
}
func (m *mockLoreStore) ListLoreByRealm(ctx context.Context, journey string) ([]gamecontent.LoreDefinition, error) {
	return nil, nil
}
func (m *mockLoreStore) GetLore(ctx context.Context, slug string) (*gamecontent.LoreDefinition, error) {
	return nil, nil
}

type mockCheatStore struct{}

func (m *mockCheatStore) ListChestDefinitions(ctx context.Context) ([]game.ChestDefinition, error) {
	return nil, nil
}
func (m *mockCheatStore) GetChestDefinition(ctx context.Context, slug string) (*game.ChestDefinition, error) {
	return nil, nil
}
func (m *mockCheatStore) ListDropTableEntries(ctx context.Context, chestSlug string) ([]game.DropTableEntry, error) {
	return nil, nil
}

type mockRelicStore struct{}

func (m *mockRelicStore) ListRelicDefinitions(ctx context.Context) ([]game.RelicDefinition, error) {
	return nil, nil
}
func (m *mockRelicStore) GetRelicDefinition(ctx context.Context, slug string) (*game.RelicDefinition, error) {
	return nil, nil
}

func TestCache_HitAndMiss(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")

	v, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit for key1")
	}
	if v != "value1" {
		t.Errorf("expected value1, got %v", v)
	}

	_, ok = c.Get("nonexistent")
	if ok {
		t.Error("expected cache miss for nonexistent key")
	}

	stats := c.Stats()
	if stats.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", stats.Misses)
	}
}

func TestCache_Expiration(t *testing.T) {
	c := NewCache(50 * time.Millisecond)
	c.Set("key1", "value1")

	_, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected cache hit before expiration")
	}

	time.Sleep(60 * time.Millisecond)
	_, ok = c.Get("key1")
	if ok {
		t.Error("expected cache miss after expiration")
	}
}

func TestCache_Invalidate(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")

	if !c.Invalidate("key1") {
		t.Error("expected Invalidate to return true for existing key")
	}
	if c.Invalidate("key1") {
		t.Error("expected Invalidate to return false for non-existent key")
	}

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be evicted")
	}
	_, ok = c.Get("key2")
	if !ok {
		t.Error("expected key2 to still be cached")
	}
}

func TestCache_Clear(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "value1")
	c.Set("key2", "value2")

	beforeGen := c.Generation()
	c.Clear()
	afterGen := c.Generation()

	if afterGen <= beforeGen {
		t.Error("expected generation to increase after Clear")
	}

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be cleared")
	}
}

func TestCache_HitRatio(t *testing.T) {
	c := NewCache(5 * time.Minute)
	c.Set("key1", "value1")
	c.Get("key1")     // hit
	c.Get("missing1") // miss
	c.Get("missing2") // miss

	ratio := c.HitRatio()
	expected := 1.0 / 3.0
	if ratio != expected {
		t.Errorf("expected hit ratio %f, got %f", expected, ratio)
	}
}

func TestContentService_Reload(t *testing.T) {
	realmStore := &mockDefinitionStore{
		realms: []gamecontent.RealmDefinition{
			{ID: 1, Slug: "forest", Name: "Forest"},
		},
	}
	questStore := &mockQuestStore{
		missions: []gamecontent.QuestDefinition{
			{ID: 1, Slug: "q1", Journey: "forest", Title: "Mission 1"},
		},
	}

	svc := NewContentService(
		realmStore,
		&mockDefinitionStore{},
		questStore,
		&mockPromptStore{},
		&mockAchievementStore{},
		&mockSeasonStore{},
		&mockLoreStore{},
		&mockCheatStore{},
		&mockRelicStore{},
		ContentServiceConfig{CacheTTL: 5 * time.Minute},
	)

	ctx := context.Background()
	if err := svc.Reload(ctx); err != nil {
		t.Fatalf("reload failed: %v", err)
	}

	realms, err := svc.ListRealms(ctx)
	if err != nil {
		t.Fatalf("ListRealms failed: %v", err)
	}
	if len(realms) != 1 {
		t.Fatalf("expected 1 journey, got %d", len(realms))
	}

	missions, err := svc.ListQuests(ctx)
	if err != nil {
		t.Fatalf("ListQuests failed: %v", err)
	}
	if len(missions) != 1 {
		t.Fatalf("expected 1 quest, got %d", len(missions))
	}
}

func TestContentService_CacheStats(t *testing.T) {
	svc := NewContentService(
		&mockDefinitionStore{},
		&mockDefinitionStore{},
		&mockQuestStore{},
		&mockPromptStore{},
		&mockAchievementStore{},
		&mockSeasonStore{},
		&mockLoreStore{},
		&mockCheatStore{},
		&mockRelicStore{},
		ContentServiceConfig{CacheTTL: 5 * time.Minute},
	)

	stats := svc.CacheStats()
	if stats.Hits != 0 || stats.Misses != 0 {
		t.Errorf("expected zero stats, got hits=%d misses=%d", stats.Hits, stats.Misses)
	}
}

func TestContentService_Invalidate(t *testing.T) {
	svc := NewContentService(
		&mockDefinitionStore{},
		&mockDefinitionStore{},
		&mockQuestStore{},
		&mockPromptStore{},
		&mockAchievementStore{},
		&mockSeasonStore{},
		&mockLoreStore{},
		&mockCheatStore{},
		&mockRelicStore{},
		ContentServiceConfig{CacheTTL: 5 * time.Minute},
	)

	svc.Invalidate("")
	svc.Invalidate("somekey")

	gen := svc.CacheGeneration()
	svc.Invalidate("")
	if svc.CacheGeneration() <= gen {
		t.Error("expected generation to increase after Invalidate('')")
	}
}
