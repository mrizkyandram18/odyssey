package admin

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/content"
	gamecontent "odyssey/pkg/game/content"
)

type mockAdminStore struct {
	data      map[string][]map[string]any
	getBySlug func(ctx context.Context, table, slug string) (map[string]any, error)
}

func newMockAdminStore() *mockAdminStore {
	return &mockAdminStore{
		data: make(map[string][]map[string]any),
	}
}

func (m *mockAdminStore) GetByID(ctx context.Context, table string, id int64) (map[string]any, error) {
	for _, row := range m.data[table] {
		if rowID, ok := row["id"].(int64); ok && rowID == id {
			return row, nil
		}
	}
	return nil, nil
}

func (m *mockAdminStore) GetBySlug(ctx context.Context, table, slug string) (map[string]any, error) {
	if m.getBySlug != nil {
		return m.getBySlug(ctx, table, slug)
	}
	for _, row := range m.data[table] {
		if rowSlug, ok := row["slug"].(string); ok && rowSlug == slug {
			return row, nil
		}
	}
	return nil, nil
}

func (m *mockAdminStore) ListAll(ctx context.Context, table string, includeDeleted bool) ([]map[string]any, error) {
	return m.data[table], nil
}

func (m *mockAdminStore) Create(ctx context.Context, table string, data map[string]any) (map[string]any, error) {
	data["id"] = int64(len(m.data[table]) + 1)
	m.data[table] = append(m.data[table], data)
	return data, nil
}

func (m *mockAdminStore) UpdateDraft(ctx context.Context, table, slug string, patch map[string]any, updatedBy string) error {
	rows := m.data[table]
	for i, row := range rows {
		if rowSlug, ok := row["slug"].(string); ok && rowSlug == slug {
			for k, v := range patch {
				rows[i][k] = v
			}
			rows[i]["version"] = rows[i]["version"].(int) + 1
			rows[i]["updated_by"] = updatedBy
			rows[i]["updated_at"] = time.Now().UTC().Format(time.RFC3339)
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockAdminStore) Publish(ctx context.Context, table, slug string, updatedBy string) error {
	rows := m.data[table]
	for i, row := range rows {
		if rowSlug, ok := row["slug"].(string); ok && rowSlug == slug {
			rows[i]["published"] = true
			rows[i]["published_at"] = time.Now().UTC().Format(time.RFC3339)
			if v, ok := rows[i]["version"].(int); ok {
				rows[i]["version"] = v + 1
			}
			rows[i]["updated_by"] = updatedBy
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockAdminStore) SoftDelete(ctx context.Context, table, slug string) error {
	rows := m.data[table]
	for i, row := range rows {
		if rowSlug, ok := row["slug"].(string); ok && rowSlug == slug {
			rows[i]["deleted_at"] = time.Now().UTC().Format(time.RFC3339)
			rows[i]["published"] = false
			return nil
		}
	}
	return errors.New("not found")
}

func (m *mockAdminStore) Restore(ctx context.Context, table, slug string) error {
	rows := m.data[table]
	for i, row := range rows {
		if rowSlug, ok := row["slug"].(string); ok && rowSlug == slug {
			rows[i]["deleted_at"] = nil
			return nil
		}
	}
	return errors.New("not found")
}

type mockContentServiceForAdmin struct {
	reloadCalled         bool
	reloadResourceCalled bool
	reloadedResource     string
	realms               []gamecontent.RealmDefinition
	chapters             []gamecontent.ChapterDefinition
	quests               []gamecontent.QuestDefinition
	counter              int64
}

func (m *mockContentServiceForAdmin) Reload(ctx context.Context) error {
	m.reloadCalled = true
	return nil
}

func (m *mockContentServiceForAdmin) ReloadResource(ctx context.Context, resourceType string) error {
	m.reloadResourceCalled = true
	m.reloadedResource = resourceType
	return nil
}

func (m *mockContentServiceForAdmin) Status(ctx context.Context) (map[string]any, error) {
	return map[string]any{
		"realms":   len(m.realms),
		"chapters": len(m.chapters),
		"quests":   len(m.quests),
	}, nil
}

func (m *mockContentServiceForAdmin) CacheStats() content.CacheStats {
	return content.CacheStats{Hits: 10, Misses: 5, Evictions: 1}
}

func (m *mockContentServiceForAdmin) CacheHitRatio() float64 { return 0.66 }
func (m *mockContentServiceForAdmin) CacheGeneration() int64 { return 42 }
func (m *mockContentServiceForAdmin) Invalidate(key string)  {}

func (m *mockContentServiceForAdmin) ListRealms(ctx context.Context) ([]gamecontent.RealmDefinition, error) {
	return m.realms, nil
}
func (m *mockContentServiceForAdmin) ListChapters(ctx context.Context) ([]gamecontent.ChapterDefinition, error) {
	return m.chapters, nil
}
func (m *mockContentServiceForAdmin) ListChaptersByRealm(ctx context.Context, realm string) ([]gamecontent.ChapterDefinition, error) {
	return m.chapters, nil
}
func (m *mockContentServiceForAdmin) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.quests, nil
}
func (m *mockContentServiceForAdmin) ListQuestsByRealm(ctx context.Context, realm string) ([]gamecontent.QuestDefinition, error) {
	return m.quests, nil
}
func (m *mockContentServiceForAdmin) ListPrompts(ctx context.Context) ([]gamecontent.CreativePromptDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListPromptsByRealm(ctx context.Context, realm string) ([]gamecontent.CreativePromptDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListAchievements(ctx context.Context) ([]gamecontent.AchievementDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListSeasons(ctx context.Context) ([]gamecontent.SeasonDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListLore(ctx context.Context) ([]gamecontent.LoreDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListLoreByRealm(ctx context.Context, realm string) ([]gamecontent.LoreDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListLoreByChapter(ctx context.Context, chapter string) ([]gamecontent.LoreDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListChests(ctx context.Context) ([]gamecontent.ChestDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) ListRelics(ctx context.Context) ([]gamecontent.RelicDefinition, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) GetDraft(ctx context.Context, table, slug string) (map[string]any, error) {
	return nil, nil
}
func (m *mockContentServiceForAdmin) SaveDraft(ctx context.Context, table, slug string, patch map[string]any, updatedBy string) error {
	return nil
}
func (m *mockContentServiceForAdmin) CreateDraft(ctx context.Context, table string, data map[string]any) (map[string]any, error) {
	return data, nil
}
func (m *mockContentServiceForAdmin) Publish(ctx context.Context, table, slug, updatedBy string) error {
	return nil
}
func (m *mockContentServiceForAdmin) SoftDelete(ctx context.Context, table, slug string) error {
	return nil
}
func (m *mockContentServiceForAdmin) Restore(ctx context.Context, table, slug string) error {
	return nil
}

func TestAdminService_Create(t *testing.T) {
	mockStore := newMockAdminStore()
	mockContent := &mockContentServiceForAdmin{}
	svc := NewAdminService(mockContent, mockStore, nil)

	data := map[string]any{"slug": "test-realm", "name": "Test Realm"}
	created, err := svc.Create(context.Background(), "realms", data)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if created["slug"] != "test-realm" {
		t.Errorf("expected slug test-realm, got %v", created["slug"])
	}
	if !mockContent.reloadResourceCalled {
		t.Error("expected ReloadResource to be called after Create")
	}
}

func TestAdminService_SaveDraft(t *testing.T) {
	mockStore := newMockAdminStore()
	mockContent := &mockContentServiceForAdmin{}
	svc := NewAdminService(mockContent, mockStore, nil)

	mockStore.data["odyssey_realm_definitions"] = []map[string]any{
		{"id": int64(1), "slug": "forest", "name": "Forest", "version": 1},
	}

	patch := map[string]any{"name": "Updated Forest", "description": "New desc"}
	err := svc.SaveDraft(context.Background(), "realms", "forest", patch)
	if err != nil {
		t.Fatalf("SaveDraft failed: %v", err)
	}
	if !mockContent.reloadResourceCalled {
		t.Error("expected ReloadResource to be called after SaveDraft")
	}
	if mockContent.reloadedResource != "realms" {
		t.Errorf("expected reloaded resource 'realms', got '%s'", mockContent.reloadedResource)
	}
}

func TestAdminService_Publish(t *testing.T) {
	mockStore := newMockAdminStore()
	mockContent := &mockContentServiceForAdmin{}
	svc := NewAdminService(mockContent, mockStore, nil)

	mockStore.data["odyssey_realm_definitions"] = []map[string]any{
		{"id": int64(1), "slug": "forest", "name": "Forest", "published": false, "version": 1},
	}

	err := svc.Publish(context.Background(), "realms", "forest")
	if err != nil {
		t.Fatalf("Publish failed: %v", err)
	}
	if !mockContent.reloadResourceCalled {
		t.Error("expected ReloadResource to be called after Publish")
	}
}

func TestAdminService_Delete(t *testing.T) {
	mockStore := newMockAdminStore()
	mockContent := &mockContentServiceForAdmin{}
	svc := NewAdminService(mockContent, mockStore, nil)

	mockStore.data["odyssey_quest_definitions"] = []map[string]any{
		{"id": int64(1), "slug": "quest1", "title": "Quest 1", "published": true, "version": 1},
	}

	err := svc.Delete(context.Background(), "quests", "quest1")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !mockContent.reloadResourceCalled {
		t.Error("expected ReloadResource to be called after Delete")
	}
}

func TestAdminService_Restore(t *testing.T) {
	mockStore := newMockAdminStore()
	mockContent := &mockContentServiceForAdmin{}
	svc := NewAdminService(mockContent, mockStore, nil)

	mockStore.data["odyssey_quest_definitions"] = []map[string]any{
		{"id": int64(1), "slug": "quest1", "title": "Quest 1", "published": false, "version": 1, "deleted_at": time.Now().Format(time.RFC3339)},
	}

	err := svc.Restore(context.Background(), "quests", "quest1")
	if err != nil {
		t.Fatalf("Restore failed: %v", err)
	}
	if !mockContent.reloadResourceCalled {
		t.Error("expected ReloadResource to be called after Restore")
	}
}

func TestAdminService_Status(t *testing.T) {
	mockContent := &mockContentServiceForAdmin{
		realms:   []gamecontent.RealmDefinition{{Slug: "forest"}},
		chapters: []gamecontent.ChapterDefinition{{Slug: "ch1"}},
		quests:   []gamecontent.QuestDefinition{{Slug: "q1"}},
	}
	svc := NewAdminService(mockContent, nil, nil)

	status, err := svc.Status(context.Background())
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}
	if status["realms"] != 1 {
		t.Errorf("expected 1 realm, got %v", status["realms"])
	}
	if status["chapters"] != 1 {
		t.Errorf("expected 1 chapter, got %v", status["chapters"])
	}
}

func TestAdminService_UnknownResource(t *testing.T) {
	mockContent := &mockContentServiceForAdmin{}
	svc := NewAdminService(mockContent, nil, nil)

	_, err := svc.Create(context.Background(), "unknown_type", map[string]any{})
	if err == nil {
		t.Error("expected error for unknown resource type")
	}
}

func TestGetTable(t *testing.T) {
	tests := []struct {
		resource  string
		wantTable string
		wantOk    bool
	}{
		{"realms", "odyssey_realm_definitions", true},
		{"chapters", "odyssey_chapter_definitions", true},
		{"quests", "odyssey_quest_definitions", true},
		{"prompts", "odyssey_creative_prompt_definitions", true},
		{"chests", "odyssey_chest_definitions", true},
		{"relics", "odyssey_relic_definitions", true},
		{"achievements", "odyssey_achievement_definitions", true},
		{"seasons", "odyssey_season_definitions", true},
		{"lore", "odyssey_lore_definitions", true},
		{"unknown", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.resource, func(t *testing.T) {
			table, ok := getTable(tt.resource)
			if ok != tt.wantOk {
				t.Errorf("getTable(%q) ok = %v, want %v", tt.resource, ok, tt.wantOk)
			}
			if table != tt.wantTable {
				t.Errorf("getTable(%q) table = %q, want %q", tt.resource, table, tt.wantTable)
			}
		})
	}
}
