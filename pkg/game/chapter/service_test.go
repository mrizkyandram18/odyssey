package chapter

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	gamecontent "odyssey/pkg/game/content"
	"odyssey/pkg/game/events"
)

type mockChapterGateway struct {
	chapters []gamecontent.ChapterDefinition
	quests   []gamecontent.QuestDefinition
	err      error
	getErr   error
}

func (m *mockChapterGateway) ListChapters(ctx context.Context) ([]gamecontent.ChapterDefinition, error) {
	return m.chapters, m.err
}
func (m *mockChapterGateway) GetChapter(ctx context.Context, slug string) (*gamecontent.ChapterDefinition, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, c := range m.chapters {
		if c.Slug == slug {
			return &c, nil
		}
	}
	return nil, nil
}
func (m *mockChapterGateway) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.quests, m.err
}
func (m *mockChapterGateway) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	for _, q := range m.quests {
		if q.Slug == slug {
			return &q, nil
		}
	}
	return nil, nil
}

type mockChapterProgressStore struct {
	progress map[string]*game.ChapterProgress
	err      error
	updates  []string
}

func newMockChapterProgressStore() *mockChapterProgressStore {
	return &mockChapterProgressStore{
		progress: make(map[string]*game.ChapterProgress),
	}
}

func (m *mockChapterProgressStore) GetChapterProgress(ctx context.Context, crewID, chapter string) (*game.ChapterProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	cp, ok := m.progress[crewID+"|"+chapter]
	if !ok {
		return nil, game.ErrNotFound
	}
	return cp, nil
}
func (m *mockChapterProgressStore) CreateChapterProgress(ctx context.Context, cp *game.ChapterProgress) (*game.ChapterProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	key := cp.CrewID + "|" + cp.Chapter
	if _, exists := m.progress[key]; exists {
		return nil, errors.New("already exists")
	}
	m.progress[key] = cp
	return cp, nil
}
func (m *mockChapterProgressStore) UpdateChapterProgress(ctx context.Context, crewID, chapter string, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	key := crewID + "|" + chapter
	cp, ok := m.progress[key]
	if !ok {
		return game.ErrNotFound
	}
	if s, ok := patch["status"].(string); ok {
		cp.Status = s
	}
	if ca, ok := patch["completed_at"].(*time.Time); ok {
		cp.CompletedAt = ca
	}
	m.updates = append(m.updates, key)
	return nil
}
func (m *mockChapterProgressStore) ListChapterProgressByCrew(ctx context.Context, crewID string) ([]game.ChapterProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make([]game.ChapterProgress, 0)
	for key, cp := range m.progress {
		if len(key) > len(crewID) && key[:len(crewID)] == crewID {
			result = append(result, *cp)
		}
	}
	return result, nil
}

func makeChapterDef(slug, realm, title string, order int) gamecontent.ChapterDefinition {
	return gamecontent.ChapterDefinition{
		Slug:        slug,
		Realm:       realm,
		Title:       title,
		Description: "desc",
		Order:       order,
	}
}

func TestListByRealm_ReturnsSortedChapters(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-b", "realm-1", "B", 2),
			makeChapterDef("ch-a", "realm-1", "A", 1),
			makeChapterDef("ch-c", "realm-2", "C", 1),
		},
	}
	svc := NewChapterService(newMockChapterProgressStore(), nil, gw, events.NopPublisher{})

	chapters, err := svc.ListByRealm(context.Background(), "realm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chapters) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(chapters))
	}
	if chapters[0].Order != 1 || chapters[1].Order != 2 {
		t.Errorf("expected sorted by order 1,2; got %d,%d", chapters[0].Order, chapters[1].Order)
	}
}

func TestGet_ChapterNotFound(t *testing.T) {
	gw := &mockChapterGateway{chapters: []gamecontent.ChapterDefinition{}}
	svc := NewChapterService(newMockChapterProgressStore(), nil, gw, events.NopPublisher{})

	_, err := svc.Get(context.Background(), "crew-1", "unknown")
	if !errors.Is(err, ErrChapterNotFound) {
		t.Errorf("expected ErrChapterNotFound, got %v", err)
	}
}

func TestGet_ChapterLockedWhenNoProgress(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Chapter 1", 1),
		},
	}
	svc := NewChapterService(newMockChapterProgressStore(), nil, gw, events.NopPublisher{})

	result, err := svc.Get(context.Background(), "crew-1", "ch-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != ChapterStatusLocked {
		t.Errorf("expected LOCKED, got %s", result.Status)
	}
}

func TestGet_ChapterShowsActiveProgress(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Chapter 1", 1),
		},
	}
	store := newMockChapterProgressStore()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID:  "crew-1",
		Chapter: "ch-1",
		Realm:   "realm-1",
		Status:  ChapterStatusActive,
	}
	svc := NewChapterService(store, nil, gw, events.NopPublisher{})

	result, err := svc.Get(context.Background(), "crew-1", "ch-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != ChapterStatusActive {
		t.Errorf("expected ACTIVE, got %s", result.Status)
	}
}

func TestListProgress_LockedWhenNoProgress(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Chapter 1", 1),
			makeChapterDef("ch-2", "realm-1", "Chapter 2", 2),
		},
	}
	svc := NewChapterService(newMockChapterProgressStore(), nil, gw, events.NopPublisher{})

	result, err := svc.ListProgress(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 chapters, got %d", len(result))
	}
	if result[0].Status != ChapterStatusLocked || result[1].Status != ChapterStatusLocked {
		t.Errorf("expected both chapters LOCKED")
	}
}

func TestListProgress_ShowsExistingProgress(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Chapter 1", 1),
			makeChapterDef("ch-2", "realm-1", "Chapter 2", 2),
		},
	}
	store := newMockChapterProgressStore()
	completed := time.Now().UTC()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-1", Realm: "realm-1",
		Status: ChapterStatusComplete, CompletedAt: &completed,
	}
	svc := NewChapterService(store, nil, gw, events.NopPublisher{})

	result, err := svc.ListProgress(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result[0].Status != ChapterStatusComplete {
		t.Errorf("expected ch-1 COMPLETE, got %s", result[0].Status)
	}
	if result[1].Status != ChapterStatusLocked {
		t.Errorf("expected ch-2 LOCKED, got %s", result[1].Status)
	}
}

func TestGetProgressView_IdentifiesCurrentAndNext(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
			makeChapterDef("ch-2", "realm-1", "Ch 2", 2),
			makeChapterDef("ch-3", "realm-1", "Ch 3", 3),
		},
	}
	store := newMockChapterProgressStore()
	completed := time.Now().UTC()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-1", Realm: "realm-1",
		Status: ChapterStatusComplete, CompletedAt: &completed,
	}
	store.progress["crew-1|ch-2"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-2", Realm: "realm-1",
		Status: ChapterStatusActive,
	}
	svc := NewChapterService(store, nil, gw, events.NopPublisher{})

	view, err := svc.GetProgressView(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if view.CurrentChapter == nil || view.CurrentChapter.Definition.Slug != "ch-2" {
		t.Errorf("expected current chapter ch-2, got %v", view.CurrentChapter)
	}
	if view.NextChapter == nil || view.NextChapter.Definition.Slug != "ch-3" {
		t.Errorf("expected next chapter ch-3, got %v", view.NextChapter)
	}
	if len(view.CompletedChapters) != 1 {
		t.Errorf("expected 1 completed chapter, got %d", len(view.CompletedChapters))
	}
	if len(view.UnlockedChapters) != 2 {
		t.Errorf("expected 2 unlocked chapters, got %d", len(view.UnlockedChapters))
	}
}

func TestCheckCompletion_AllMandatoryDone(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
		quests: []gamecontent.QuestDefinition{
			{Slug: "q1", Chapter: "ch-1", IsMandatory: true},
			{Slug: "q2", Chapter: "ch-1", IsMandatory: true},
		},
	}
	qs := &mockQuestStore{}
	svc := NewChapterService(newMockChapterProgressStore(), qs, gw, events.NopPublisher{})

	qs.quests = []game.Quest{
		{CrewID: "c1", TemplateSlug: "q1", Status: "DONE"},
		{CrewID: "c1", TemplateSlug: "q2", Status: "DONE"},
	}

	done, err := svc.CheckCompletion(context.Background(), "c1", "ch-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done {
		t.Error("expected chapter complete, all mandatory quests done")
	}
}

func TestCheckCompletion_NotAllDone(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
		quests: []gamecontent.QuestDefinition{
			{Slug: "q1", Chapter: "ch-1", IsMandatory: true},
			{Slug: "q2", Chapter: "ch-1", IsMandatory: true},
		},
	}
	qs := &mockQuestStore{}
	svc := NewChapterService(newMockChapterProgressStore(), qs, gw, events.NopPublisher{})

	qs.quests = []game.Quest{
		{CrewID: "c1", TemplateSlug: "q1", Status: "DONE"},
		{CrewID: "c1", TemplateSlug: "q2", Status: "PENDING"},
	}

	done, err := svc.CheckCompletion(context.Background(), "c1", "ch-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if done {
		t.Error("expected chapter not complete, q2 not done")
	}
}

func TestCheckCompletion_NoMandatoryQuests(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
		quests: []gamecontent.QuestDefinition{
			{Slug: "q1", Chapter: "ch-1", IsMandatory: false},
		},
	}
	qs := &mockQuestStore{}
	svc := NewChapterService(newMockChapterProgressStore(), qs, gw, events.NopPublisher{})

	done, err := svc.CheckCompletion(context.Background(), "c1", "ch-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !done {
		t.Error("expected chapter complete when no mandatory quests")
	}
}

type mockQuestStore struct {
	quests          []game.Quest
	getErr          error
	listErr         error
	updatePatch     map[string]any
	updateCalledFor int64
}

func (m *mockQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return nil, game.ErrNotFound
}
func (m *mockQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	return q, nil
}
func (m *mockQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.quests, nil
}
func (m *mockQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	m.updatePatch = patch
	m.updateCalledFor = questID
	return nil
}
func (m *mockQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	m.updatePatch = patch
	m.updateCalledFor = questID
	return true, nil
}
func (m *mockQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	return nil, nil
}
func (m *mockQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	return c, nil
}
func (m *mockQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	return nil
}
func (m *mockQuestStore) UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error) {
	return true, nil
}

func TestMarkComplete_UpdatesStatusAndPublishes(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
	}
	store := newMockChapterProgressStore()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-1", Realm: "realm-1",
		Status: ChapterStatusActive,
	}
	eventsPublished := &[]events.Event{}
	capturingPublisher := &capturePublisher{events: eventsPublished}

	svc := NewChapterService(store, nil, gw, capturingPublisher)

	err := svc.MarkComplete(context.Background(), "crew-1", "ch-1", "user-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.progress["crew-1|ch-1"].Status != ChapterStatusComplete {
		t.Errorf("expected COMPLETE, got %s", store.progress["crew-1|ch-1"].Status)
	}
	if len(*eventsPublished) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(*eventsPublished))
	}
	ev, ok := (*eventsPublished)[0].(events.ChapterCompletedEvent)
	if !ok {
		t.Fatalf("expected ChapterCompletedEvent, got %T", (*eventsPublished)[0])
	}
	if ev.Chapter != "ch-1" || ev.CrewID != "crew-1" || ev.Realm != "realm-1" {
		t.Errorf("unexpected event fields: %+v", ev)
	}
}

func TestMarkComplete_ChapterNotFound(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
	}
	svc := NewChapterService(newMockChapterProgressStore(), nil, gw, events.NopPublisher{})

	err := svc.MarkComplete(context.Background(), "crew-1", "ch-1", "user-1")
	if !errors.Is(err, ErrChapterNotFound) {
		t.Errorf("expected ErrChapterNotFound, got %v", err)
	}
}

func TestHandleQuestCompleted_CompletesChapter(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
		quests: []gamecontent.QuestDefinition{
			{Slug: "q1", Chapter: "ch-1", IsMandatory: true},
		},
	}
	store := newMockChapterProgressStore()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-1", Realm: "realm-1",
		Status: ChapterStatusActive,
	}
	qs := &mockQuestStore{
		quests: []game.Quest{
			{CrewID: "crew-1", TemplateSlug: "q1", Status: "DONE"},
		},
	}
	eventsPublished := &[]events.Event{}
	capturingPublisher := &capturePublisher{events: eventsPublished}

	svc := NewChapterService(store, qs, gw, capturingPublisher)
	handler := NewQuestCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.QuestCompletedEvent{
		CrewID:       "crew-1",
		TemplateSlug: "q1",
		Chapter:      "ch-1",
		Realm:        "realm-1",
		PlayerUID:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.progress["crew-1|ch-1"].Status != ChapterStatusComplete {
		t.Errorf("expected COMPLETE, got %s", store.progress["crew-1|ch-1"].Status)
	}
	if len(*eventsPublished) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(*eventsPublished))
	}
}

func TestHandleQuestCompleted_LooksUpChapterFromQuestDef(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
		quests: []gamecontent.QuestDefinition{
			{Slug: "q1", Chapter: "ch-1", IsMandatory: true},
		},
	}
	store := newMockChapterProgressStore()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-1", Realm: "realm-1",
		Status: ChapterStatusActive,
	}
	qs := &mockQuestStore{
		quests: []game.Quest{
			{CrewID: "crew-1", TemplateSlug: "q1", Status: "DONE"},
		},
	}
	svc := NewChapterService(store, qs, gw, events.NopPublisher{})
	handler := NewQuestCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.QuestCompletedEvent{
		CrewID:       "crew-1",
		TemplateSlug: "q1",
		Chapter:      "",
		Realm:        "realm-1",
		PlayerUID:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if store.progress["crew-1|ch-1"].Status != ChapterStatusComplete {
		t.Errorf("expected COMPLETE, got %s", store.progress["crew-1|ch-1"].Status)
	}
}

func TestHandleQuestCompleted_AlreadyComplete_PreventsDuplicateEvent(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
		quests: []gamecontent.QuestDefinition{
			{Slug: "q1", Chapter: "ch-1", IsMandatory: true},
		},
	}
	store := newMockChapterProgressStore()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-1", Realm: "realm-1",
		Status: ChapterStatusComplete,
	}
	qs := &mockQuestStore{
		quests: []game.Quest{
			{CrewID: "crew-1", TemplateSlug: "q1", Status: "DONE"},
		},
	}
	eventsPublished := &[]events.Event{}
	capturingPublisher := &capturePublisher{events: eventsPublished}

	svc := NewChapterService(store, qs, gw, capturingPublisher)
	handler := NewQuestCompletedHandler(svc)

	err := handler.Handle(context.Background(), events.QuestCompletedEvent{
		CrewID:       "crew-1",
		TemplateSlug: "q1",
		Chapter:      "ch-1",
		Realm:        "realm-1",
		PlayerUID:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(*eventsPublished) != 0 {
		t.Errorf("expected 0 events published for already-complete chapter, got %d", len(*eventsPublished))
	}
}

func TestUnlockFirstChapter(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
			makeChapterDef("ch-2", "realm-1", "Ch 2", 2),
		},
	}
	store := newMockChapterProgressStore()
	svc := NewChapterService(store, nil, gw, events.NopPublisher{})

	err := svc.UnlockFirstChapter(context.Background(), "crew-1", "realm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp, err := store.GetChapterProgress(context.Background(), "crew-1", "ch-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cp.Status != ChapterStatusActive {
		t.Errorf("expected ACTIVE, got %s", cp.Status)
	}
	_, err = store.GetChapterProgress(context.Background(), "crew-1", "ch-2")
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ch-2 not created, got err %v", err)
	}
}

func TestUnlockFirstChapter_NoChapters(t *testing.T) {
	gw := &mockChapterGateway{chapters: []gamecontent.ChapterDefinition{}}
	svc := NewChapterService(newMockChapterProgressStore(), nil, gw, events.NopPublisher{})

	err := svc.UnlockFirstChapter(context.Background(), "crew-1", "realm-1")
	if err != nil {
		t.Fatalf("expected no error for empty realm, got %v", err)
	}
}

func TestGetCurrentChapter_NoProgress(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
			makeChapterDef("ch-2", "realm-1", "Ch 2", 2),
		},
	}
	svc := NewChapterService(newMockChapterProgressStore(), nil, gw, events.NopPublisher{})

	result, err := svc.GetCurrentChapter(context.Background(), "crew-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected first chapter as fallback")
	}
	if result.Definition.Slug != "ch-1" {
		t.Errorf("expected ch-1, got %s", result.Definition.Slug)
	}
}

func TestEnsureFirstChapterUnlocked_AlreadyHasProgress(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
		},
	}
	store := newMockChapterProgressStore()
	store.progress["crew-1|ch-1"] = &game.ChapterProgress{
		CrewID: "crew-1", Chapter: "ch-1", Realm: "realm-1",
		Status: ChapterStatusActive,
	}
	svc := NewChapterService(store, nil, gw, events.NopPublisher{})

	err := svc.EnsureFirstChapterUnlocked(context.Background(), "crew-1", "realm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEnsureFirstChapterUnlocked_NoProgress_UnlocksFirst(t *testing.T) {
	gw := &mockChapterGateway{
		chapters: []gamecontent.ChapterDefinition{
			makeChapterDef("ch-1", "realm-1", "Ch 1", 1),
			makeChapterDef("ch-2", "realm-1", "Ch 2", 2),
		},
	}
	store := newMockChapterProgressStore()
	svc := NewChapterService(store, nil, gw, events.NopPublisher{})

	err := svc.EnsureFirstChapterUnlocked(context.Background(), "crew-1", "realm-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cp, err := store.GetChapterProgress(context.Background(), "crew-1", "ch-1")
	if err != nil {
		t.Fatalf("expected ch-1 to be unlocked: %v", err)
	}
	if cp.Status != ChapterStatusActive {
		t.Errorf("expected ACTIVE, got %s", cp.Status)
	}
	_, err = store.GetChapterProgress(context.Background(), "crew-1", "ch-2")
	if !errors.Is(err, game.ErrNotFound) {
		t.Errorf("expected ch-2 to not be unlocked, got err %v", err)
	}
}

type capturePublisher struct {
	events *[]events.Event
}

func (p *capturePublisher) Publish(ctx context.Context, event events.Event) {
	*p.events = append(*p.events, event)
}
