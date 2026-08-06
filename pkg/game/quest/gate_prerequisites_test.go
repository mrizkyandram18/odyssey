package quest

import (
	"context"
	"testing"

	gamecontent "odyssey/pkg/game/content"
)

func TestIsPrerequisiteMet_NilGateReturnsTrue(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	def := gamecontent.QuestDefinition{
		Slug:          "test",
		RequiredLevel: 5,
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true when gate is nil")
	}
}

func TestIsPrerequisiteMet_SeasonNotActive(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive: false,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:       "test",
		SeasonSlug: "winter",
	}
	if svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected false when season not active")
	}
}

func TestIsPrerequisiteMet_SeasonActive(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive: true,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:       "test",
		SeasonSlug: "summer",
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true when season active")
	}
}

func TestIsPrerequisiteMet_LevelTooLow(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive: true,
		playerLevel:  2,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:          "test",
		RequiredLevel: 5,
	}
	if svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected false when player level too low")
	}
}

func TestIsPrerequisiteMet_LevelSufficient(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive: true,
		playerLevel:  5,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:          "test",
		RequiredLevel: 5,
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true when player level sufficient")
	}
}

func TestIsPrerequisiteMet_RealmNotActive(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive: true,
		realmActive:  false,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:          "test",
		RequiredRealm: "r1",
	}
	if svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected false when realm not active")
	}
}

func TestIsPrerequisiteMet_RealmActive(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive: true,
		realmActive:  true,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:          "test",
		RequiredRealm: "r1",
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true when realm active")
	}
}

func TestIsPrerequisiteMet_ChapterNotUnlocked(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:    true,
		chapterUnlocked: false,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:            "test",
		RequiredChapter: "ch-2",
	}
	if svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected false when chapter not unlocked")
	}
}

func TestIsPrerequisiteMet_ChapterUnlocked(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:    true,
		chapterUnlocked: true,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:            "test",
		RequiredChapter: "ch-2",
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true when chapter unlocked")
	}
}

func TestIsPrerequisiteMet_QuestNotCompleted(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:   true,
		questCompleted: false,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:              "test",
		RequiredQuestSlug: "prereq",
	}
	if svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected false when prerequisite quest not completed")
	}
}

func TestIsPrerequisiteMet_NoPrerequisites(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:    true,
		chapterUnlocked: true,
		realmActive:     true,
		playerLevel:     10,
		questCompleted:  true,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug: "test",
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true for quest with no prerequisites")
	}
}

func TestIsPrerequisiteMet_MultiQuestPrerequisites(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:    true,
		chapterUnlocked: true,
		realmActive:     true,
		playerLevel:     10,
		questCompleted:  false,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:               "test",
		RequiredQuestSlugs: []string{"qA", "qB"},
	}
	if svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected false when multi prerequisites not met")
	}
}

func TestIsPrerequisiteMet_MultiQuestPrerequisitesMet(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:    true,
		chapterUnlocked: true,
		realmActive:     true,
		playerLevel:     10,
		questCompleted:  true,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:               "test",
		RequiredQuestSlugs: []string{"qA", "qB"},
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true when all multi prerequisites met")
	}
}

func TestIsPrerequisiteMet_SingleSlugTakesPrecedenceWhenMultiEmpty(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:    true,
		chapterUnlocked: true,
		realmActive:     true,
		playerLevel:     10,
		questCompleted:  true,
	}
	svc := NewQuestServiceWithGate(nil, gate, &mockContentGateway{})
	def := gamecontent.QuestDefinition{
		Slug:               "test",
		RequiredQuestSlug:  "prereq",
		RequiredQuestSlugs: []string{},
	}
	if !svc.IsPrerequisiteMet(context.Background(), def, "crew-1", "u1") {
		t.Error("expected true when single prerequisite met and multi is empty")
	}
}

func TestListAvailable_FiltersByPrerequisites(t *testing.T) {
	gate := &mockQuestGate{
		seasonActive:    true,
		chapterUnlocked: true,
		realmActive:     true,
		playerLevel:     10,
		questCompleted:  true,
	}
	store := &mockQuestStore{}
	content := &mockContentGatewayWithQuests{
		quests: []gamecontent.QuestDefinition{
			{Slug: "q-available", RequiredLevel: 5},
			{Slug: "q-locked-level", RequiredLevel: 15},
		},
	}
	svc := NewQuestServiceWithGate(store, gate, content)

	result, err := svc.ListAvailable(context.Background(), "crew-1", "u1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 available quest, got %d", len(result))
	}
	if result[0].Slug != "q-available" {
		t.Errorf("expected q-available, got %s", result[0].Slug)
	}
}

func TestListAvailable_NoContentGateway(t *testing.T) {
	svc := NewQuestService(newMockQuestStore())
	_, err := svc.ListAvailable(context.Background(), "crew-1", "u1")
	if err == nil {
		t.Fatal("expected error when no content gateway")
	}
}

type mockQuestGate struct {
	seasonActive    bool
	chapterUnlocked bool
	realmActive     bool
	playerLevel     int
	questCompleted  bool
	getLevelErr     error
}

func (m *mockQuestGate) IsChapterUnlocked(ctx context.Context, crewID, chapter string) bool {
	return m.chapterUnlocked
}
func (m *mockQuestGate) IsRealmActive(ctx context.Context, crewID, realm string) bool {
	return m.realmActive
}
func (m *mockQuestGate) IsSeasonActive(ctx context.Context, seasonSlug string) bool {
	return m.seasonActive
}
func (m *mockQuestGate) GetPlayerLevel(ctx context.Context, uid string) (int, error) {
	return m.playerLevel, m.getLevelErr
}
func (m *mockQuestGate) IsQuestCompleted(ctx context.Context, crewID, templateSlug string) bool {
	return m.questCompleted
}

func (m *mockQuestGate) ValidatePrerequisites(ctx context.Context, defs []gamecontent.QuestDefinition) error {
	return nil
}

type mockContentGatewayWithQuests struct {
	quests []gamecontent.QuestDefinition
}

func (m *mockContentGatewayWithQuests) ListQuests(ctx context.Context) ([]gamecontent.QuestDefinition, error) {
	return m.quests, nil
}
func (m *mockContentGatewayWithQuests) GetQuest(ctx context.Context, slug string) (*gamecontent.QuestDefinition, error) {
	for _, q := range m.quests {
		if q.Slug == slug {
			return &q, nil
		}
	}
	return nil, nil
}
func (m *mockContentGatewayWithQuests) ListQuestsByRealm(ctx context.Context, realm string) ([]gamecontent.QuestDefinition, error) {
	result := make([]gamecontent.QuestDefinition, 0)
	for _, q := range m.quests {
		if q.Realm == realm {
			result = append(result, q)
		}
	}
	return result, nil
}
