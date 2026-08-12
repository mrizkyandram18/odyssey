package concept

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
)

type mockLoreGateway struct {
	concept      []gamecontent.LoreDefinition
	byChapter map[string][]gamecontent.LoreDefinition
	byRealm   map[string][]gamecontent.LoreDefinition
	err       error
}

func (m *mockLoreGateway) ListLore(ctx context.Context) ([]gamecontent.LoreDefinition, error) {
	return m.concept, m.err
}
func (m *mockLoreGateway) ListLoreByChapter(ctx context.Context, course string) ([]gamecontent.LoreDefinition, error) {
	return m.byChapter[course], m.err
}
func (m *mockLoreGateway) ListLoreByRealm(ctx context.Context, journey string) ([]gamecontent.LoreDefinition, error) {
	return m.byRealm[journey], m.err
}

type mockLoreUnlockStore struct {
	unlocks map[string]game.LoreUnlock
	created []game.LoreUnlock
	getErr  error
	listErr error
}

func newMockLoreUnlockStore() *mockLoreUnlockStore {
	return &mockLoreUnlockStore{unlocks: make(map[string]game.LoreUnlock)}
}

func loreKey(crewID, slug string) string {
	return crewID + "|" + slug
}

func (m *mockLoreUnlockStore) GetLoreUnlock(ctx context.Context, crewID, loreSlug string) (*game.LoreUnlock, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	lu, ok := m.unlocks[loreKey(crewID, loreSlug)]
	if !ok {
		return nil, game.ErrNotFound
	}
	return &lu, nil
}
func (m *mockLoreUnlockStore) CreateLoreUnlock(ctx context.Context, lu *game.LoreUnlock) (*game.LoreUnlock, error) {
	m.created = append(m.created, *lu)
	m.unlocks[loreKey(lu.FamilyID, lu.ConceptSlug)] = *lu
	return lu, nil
}
func (m *mockLoreUnlockStore) UpdateLoreUnlock(ctx context.Context, crewID, loreSlug string, patch map[string]any) error {
	return nil
}
func (m *mockLoreUnlockStore) ListLoreUnlocksByCrew(ctx context.Context, crewID string) ([]game.LoreUnlock, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	result := make([]game.LoreUnlock, 0)
	for k, v := range m.unlocks {
		if len(k) > len(crewID) && k[:len(crewID)] == crewID {
			result = append(result, v)
		}
	}
	return result, nil
}

func makeLoreDef(slug, journey, course, title string, order int) gamecontent.LoreDefinition {
	return gamecontent.LoreDefinition{
		Slug:      slug,
		Journey:     journey,
		Course:   course,
		Title:     title,
		Content:   "content for " + title,
		Order:     order,
		CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
	}
}

func TestGetSummary_NoUnlocks(t *testing.T) {
	gw := &mockLoreGateway{
		concept: []gamecontent.LoreDefinition{
			makeLoreDef("l1", "r1", "ch1", "Concept 1", 1),
			makeLoreDef("l2", "r1", "ch1", "Concept 2", 2),
		},
		byChapter: map[string][]gamecontent.LoreDefinition{"ch1": {}},
		byRealm:   map[string][]gamecontent.LoreDefinition{"r1": {}},
	}
	store := newMockLoreUnlockStore()
	svc := NewLoreService(store, gw)

	summary, err := svc.GetSummary(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.LockedCount != 2 {
		t.Errorf("expected 2 locked, got %d", summary.LockedCount)
	}
	if summary.UnlockedCount != 0 {
		t.Errorf("expected 0 unlocked, got %d", summary.UnlockedCount)
	}
	if summary.Latest != nil {
		t.Error("expected no latest when all locked")
	}
}

func TestGetSummary_WithUnlocks(t *testing.T) {
	gw := &mockLoreGateway{
		concept: []gamecontent.LoreDefinition{
			makeLoreDef("l1", "r1", "ch1", "Concept 1", 1),
			makeLoreDef("l2", "r1", "ch1", "Concept 2", 2),
			makeLoreDef("l3", "r1", "ch1", "Concept 3", 3),
		},
	}
	store := newMockLoreUnlockStore()
	now := time.Now().UTC()
	store.unlocks[loreKey("crew-1", "l1")] = game.LoreUnlock{
		FamilyID: "crew-1", ConceptSlug: "l1", Journey: "r1", Course: "ch1",
		UnlockedAt: now, CreatedAt: now,
	}
	store.unlocks[loreKey("crew-1", "l2")] = game.LoreUnlock{
		FamilyID: "crew-1", ConceptSlug: "l2", Journey: "r1", Course: "ch1",
		UnlockedAt: now.Add(1 * time.Hour), CreatedAt: now,
	}
	svc := NewLoreService(store, gw)

	summary, err := svc.GetSummary(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.LockedCount != 1 {
		t.Errorf("expected 1 locked, got %d", summary.LockedCount)
	}
	if summary.UnlockedCount != 2 {
		t.Errorf("expected 2 unlocked, got %d", summary.UnlockedCount)
	}
	if summary.Latest == nil {
		t.Fatal("expected latest concept")
	}
	if summary.Latest.Slug != "l2" {
		t.Errorf("expected latest l2, got %s", summary.Latest.Slug)
	}
}

func TestListUnlocked_ReturnsUnlockedSortedByOrder(t *testing.T) {
	gw := &mockLoreGateway{
		concept: []gamecontent.LoreDefinition{
			makeLoreDef("l1", "r1", "ch1", "Concept 1", 3),
			makeLoreDef("l2", "r1", "ch1", "Concept 2", 1),
			makeLoreDef("l3", "r1", "ch1", "Concept 3", 2),
		},
	}
	store := newMockLoreUnlockStore()
	now := time.Now().UTC()
	for _, slug := range []string{"l1", "l2", "l3"} {
		store.unlocks[loreKey("crew-1", slug)] = game.LoreUnlock{
			FamilyID: "crew-1", ConceptSlug: slug, UnlockedAt: now,
		}
	}
	svc := NewLoreService(store, gw)

	result, err := svc.ListUnlocked(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 unlocked, got %d", len(result))
	}
	if result[0].Order != 1 {
		t.Errorf("expected order 1 first, got %d", result[0].Order)
	}
	if result[0].Slug != "l2" {
		t.Errorf("expected l2 first, got %s", result[0].Slug)
	}
}

func TestListUnlocks_ByCrew(t *testing.T) {
	store := newMockLoreUnlockStore()
	now := time.Now().UTC()
	store.unlocks[loreKey("crew-1", "l1")] = game.LoreUnlock{
		FamilyID: "crew-1", ConceptSlug: "l1", UnlockedAt: now,
	}
	store.unlocks[loreKey("crew-2", "l2")] = game.LoreUnlock{
		FamilyID: "crew-2", ConceptSlug: "l2", UnlockedAt: now,
	}
	svc := NewLoreService(store, &mockLoreGateway{})

	result, err := svc.ListUnlocks(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Errorf("expected 1 unlock for crew-1, got %d", len(result))
	}
}

func TestUnlockForChapter_CreatesMissingUnlocks(t *testing.T) {
	gw := &mockLoreGateway{
		byChapter: map[string][]gamecontent.LoreDefinition{
			"ch1": {
				makeLoreDef("l1", "r1", "ch1", "Concept 1", 1),
				makeLoreDef("l2", "r1", "ch1", "Concept 2", 2),
			},
		},
	}
	store := newMockLoreUnlockStore()
	store.unlocks[loreKey("crew-1", "l1")] = game.LoreUnlock{
		FamilyID: "crew-1", ConceptSlug: "l1", UnlockedAt: time.Now(),
	}
	svc := NewLoreService(store, gw)

	result, err := svc.UnlockForChapter(context.Background(), "crew-1", "ch1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 new unlock (l2), got %d", len(result))
	}
	if result[0].Slug != "l2" {
		t.Errorf("expected l2 unlocked, got %s", result[0].Slug)
	}
}

type errorLoreUnlockStore struct {
	*mockLoreUnlockStore
	createErr error
}

func (m *errorLoreUnlockStore) CreateLoreUnlock(ctx context.Context, lu *game.LoreUnlock) (*game.LoreUnlock, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return m.mockLoreUnlockStore.CreateLoreUnlock(ctx, lu)
}

func TestUnlockForChapter_PropagatesCreateError(t *testing.T) {
	gw := &mockLoreGateway{
		byChapter: map[string][]gamecontent.LoreDefinition{
			"ch1": {
				makeLoreDef("l1", "r1", "ch1", "Concept 1", 1),
			},
		},
	}
	store := &errorLoreUnlockStore{
		mockLoreUnlockStore: newMockLoreUnlockStore(),
		createErr:           errors.New("internal error"),
	}
	svc := NewLoreService(store, gw)

	_, err := svc.UnlockForChapter(context.Background(), "crew-1", "ch1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUnlockForChapter_AllAlreadyUnlocked(t *testing.T) {
	gw := &mockLoreGateway{
		byChapter: map[string][]gamecontent.LoreDefinition{
			"ch1": {
				makeLoreDef("l1", "r1", "ch1", "Concept 1", 1),
			},
		},
	}
	store := newMockLoreUnlockStore()
	store.unlocks[loreKey("crew-1", "l1")] = game.LoreUnlock{
		FamilyID: "crew-1", ConceptSlug: "l1", UnlockedAt: time.Now(),
	}
	svc := NewLoreService(store, gw)

	result, err := svc.UnlockForChapter(context.Background(), "crew-1", "ch1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 new unlocks, got %d", len(result))
	}
}

func TestChapterCompletedHandler_TriggersUnlock(t *testing.T) {
	gw := &mockLoreGateway{
		byChapter: map[string][]gamecontent.LoreDefinition{
			"ch1": {
				makeLoreDef("l1", "r1", "ch1", "Concept 1", 1),
			},
		},
	}
	store := newMockLoreUnlockStore()
	svc := NewLoreServiceWithPublisher(store, gw, events.NopPublisher{})
	handler := NewChapterCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.ChapterCompletedEvent{
		FamilyID:    "crew-1",
		Course:   "ch1",
		Journey:     "r1",
		PlayerUID: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	lu, err := store.GetLoreUnlock(context.Background(), "crew-1", "l1")
	if err != nil {
		t.Fatalf("expected concept unlock to be created: %v", err)
	}
	if lu.ConceptSlug != "l1" {
		t.Errorf("expected l1, got %s", lu.ConceptSlug)
	}
}

func TestChapterCompletedHandler_OtherEventIgnored(t *testing.T) {
	svc := NewLoreServiceWithPublisher(newMockLoreUnlockStore(), &mockLoreGateway{}, events.NopPublisher{})
	handler := NewChapterCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.QuestCompletedEvent{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
