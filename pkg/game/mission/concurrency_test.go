package quest

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/events"
	"odyssey/pkg/game/progression"
	"odyssey/pkg/game/world"
)

type concurrentQuestStore struct {
	mu         sync.Mutex
	missions     map[int64]*game.Mission
	exercises map[int64][]game.Exercise
}

func newConcurrentQuestStore() *concurrentQuestStore {
	return &concurrentQuestStore{
		missions:     make(map[int64]*game.Mission),
		exercises: make(map[int64][]game.Exercise),
	}
}

func (m *concurrentQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Mission, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, ok := m.missions[questID]
	if !ok {
		return nil, game.ErrNotFound
	}
	p := *q
	return &p, nil
}

func (m *concurrentQuestStore) CreateQuest(ctx context.Context, q *game.Mission) (*game.Mission, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	q.ID = int64(len(m.missions) + 1)
	m.missions[q.ID] = q
	return q, nil
}

func (m *concurrentQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Mission, error) {
	return nil, nil
}

func (m *concurrentQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, ok := m.missions[questID]
	if !ok {
		return game.ErrNotFound
	}
	if status, ok := patch["status"].(string); ok {
		q.Status = status
	}
	if ca, ok := patch["completed_at"].(*time.Time); ok {
		q.CompletedAt = ca
	}
	return nil
}

func (m *concurrentQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	q, ok := m.missions[questID]
	if !ok {
		return false, game.ErrNotFound
	}
	if q.Status != oldStatus {
		return false, nil
	}
	if status, ok := patch["status"].(string); ok {
		q.Status = status
	}
	if ca, ok := patch["completed_at"].(*time.Time); ok {
		q.CompletedAt = ca
	}
	return true, nil
}

func (m *concurrentQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Exercise, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	chs := m.exercises[questID]
	result := make([]game.Exercise, len(chs))
	copy(result, chs)
	return result, nil
}

func (m *concurrentQuestStore) CreateChallenge(ctx context.Context, c *game.Exercise) (*game.Exercise, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c.ID = int64(len(m.exercises[c.MissionID]) + 1)
	m.exercises[c.MissionID] = append(m.exercises[c.MissionID], *c)
	return c, nil
}

func (m *concurrentQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, chs := range m.exercises {
		for i := range chs {
			if chs[i].ID == challengeID {
				if status, ok := patch["status"].(string); ok {
					chs[i].Status = status
				}
				if completedBy, ok := patch["completed_by"].(string); ok {
					chs[i].CompletedBy = completedBy
				}
				if ca, ok := patch["completed_at"].(*time.Time); ok {
					chs[i].CompletedAt = ca
				}
				if owner, ok := patch["assigned_to"].(string); ok {
					chs[i].AssignedTo = &owner
				}
				return nil
			}
		}
	}
	return game.ErrNotFound
}

func (m *concurrentQuestStore) UpdateChallengeIfMatch(ctx context.Context, challengeID int64, oldStatus string, patch map[string]any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, chs := range m.exercises {
		for i := range chs {
			if chs[i].ID == challengeID {
				if chs[i].Status != oldStatus {
					return false, nil
				}
				if status, ok := patch["status"].(string); ok {
					chs[i].Status = status
				}
				if completedBy, ok := patch["completed_by"].(string); ok {
					chs[i].CompletedBy = completedBy
				}
				if ca, ok := patch["completed_at"].(*time.Time); ok {
					chs[i].CompletedAt = ca
				}
				return true, nil
			}
		}
	}
	return false, game.ErrNotFound
}

type concurrentUserStore struct {
	mu     sync.Mutex
	player *game.Player
}

func (m *concurrentUserStore) GetUser(ctx context.Context, uid string) (*game.Player, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	p := *m.player
	return &p, nil
}

func (m *concurrentUserStore) CreateUser(ctx context.Context, p *game.Player) error {
	return nil
}

func (m *concurrentUserStore) UpdateUser(ctx context.Context, uid string, patch map[string]any) error {
	return errors.New("use UpdateUserIfMatch")
}

func (m *concurrentUserStore) UpdateUserIfMatch(ctx context.Context, uid string, version int, patch map[string]any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.player.Version != version {
		return false, nil
	}
	if v, ok := patch["xp"].(int64); ok {
		m.player.XP = v
	}
	if v, ok := patch["level"].(int); ok {
		m.player.Level = v
	}
	if v, ok := patch["version"].(int); ok {
		m.player.Version = v
	}
	return true, nil
}

type concurrentRealmStore struct {
	mu       sync.Mutex
	progress map[string]*game.JourneyProgress
}

func newConcurrentRealmStore() *concurrentRealmStore {
	return &concurrentRealmStore{
		progress: make(map[string]*game.JourneyProgress),
	}
}

func (m *concurrentRealmStore) GetRealmProgress(ctx context.Context, crewID, journey string) (*game.JourneyProgress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rp, ok := m.progress[crewID+"|"+journey]
	if !ok {
		return nil, game.ErrNotFound
	}
	p := *rp
	return &p, nil
}

func (m *concurrentRealmStore) CreateRealmProgress(ctx context.Context, rp *game.JourneyProgress) (*game.JourneyProgress, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.progress[rp.FamilyID+"|"+rp.Journey] = rp
	return rp, nil
}

func (m *concurrentRealmStore) UpdateRealmProgress(ctx context.Context, crewID, journey string, patch map[string]any) error {
	return nil
}

func (m *concurrentRealmStore) UpdateRealmProgressIfMatch(ctx context.Context, crewID, journey string, oldProgress int, patch map[string]any) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rp, ok := m.progress[crewID+"|"+journey]
	if !ok {
		return false, game.ErrNotFound
	}
	if rp.Progress != oldProgress {
		return false, nil
	}
	if p, ok := patch["progress"].(int); ok {
		rp.Progress = p
	}
	if s, ok := patch["status"].(string); ok {
		rp.Status = s
	}
	return true, nil
}

func (m *concurrentRealmStore) ListRealmProgressByCrew(ctx context.Context, crewID string) ([]game.JourneyProgress, error) {
	return nil, nil
}

func TestCompleteChallenge_ConcurrentLastChallenge(t *testing.T) {
	store := newConcurrentQuestStore()
	now := time.Now().UTC()
	q := &game.Mission{
		ID: 1, FamilyID: "crew-1", TemplateSlug: "morning-light",
		Title: "Morning Light", Status: string(QuestStatusActive), CreatedAt: now,
	}
	store.missions[1] = q
	store.exercises[1] = []game.Exercise{
		{ID: 11, MissionID: 1, Slug: "ch-a", Status: string(ChallengeStatusDone), CreatedAt: now},
		{ID: 12, MissionID: 1, Slug: "ch-b", Status: string(ChallengeStatusPending), CreatedAt: now},
	}

	userStore := &concurrentUserStore{player: &game.Player{
		UID: "user-1", FamilyID: "crew-1", Level: 1, XP: 0, Version: 1,
	}}
	prog := progression.NewProgressionService(userStore, nil)
	journey := newConcurrentRealmStore()
	_, _ = journey.CreateRealmProgress(context.Background(), &game.JourneyProgress{
		FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 0,
	})

	qs := NewQuestService(store)
	h := NewQuestAPIHandler(qs, prog, journey, world.NewRealmCatalog(world.DefaultRealmDefinitions), nil)

	var wg sync.WaitGroup
	results := make([]*CompleteChallengeResult, 2)
	errs := make([]error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results[idx], errs[idx] = h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		}
	}

	if successCount != 2 {
		t.Errorf("expected both calls to succeed, got %d successes", successCount)
	}

	quest, _ := store.GetQuest(context.Background(), 1)
	if quest.Status != string(QuestStatusDone) {
		t.Errorf("expected quest DONE, got %s", quest.Status)
	}
}

func TestCompleteChallenge_EventPublishedOnce(t *testing.T) {
	store := newConcurrentQuestStore()
	now := time.Now().UTC()
	q := &game.Mission{
		ID: 1, FamilyID: "crew-1", TemplateSlug: "morning-light",
		Title: "Morning Light", Status: string(QuestStatusActive), CreatedAt: now,
	}
	store.missions[1] = q
	store.exercises[1] = []game.Exercise{
		{ID: 11, MissionID: 1, Slug: "ch-a", Status: string(ChallengeStatusDone), CreatedAt: now},
		{ID: 12, MissionID: 1, Slug: "ch-b", Status: string(ChallengeStatusPending), CreatedAt: now},
	}

	userStore := &concurrentUserStore{player: &game.Player{
		UID: "user-1", FamilyID: "crew-1", Level: 1, XP: 0, Version: 1,
	}}
	prog := progression.NewProgressionService(userStore, nil)
	journey := newConcurrentRealmStore()
	_, _ = journey.CreateRealmProgress(context.Background(), &game.JourneyProgress{
		FamilyID: "crew-1", Journey: "whispering-woods", Status: "ACTIVE", Progress: 0,
	})

	pub := &capturePublisher{}
	qs := NewQuestService(store)
	h := NewQuestAPIHandler(qs, prog, journey, world.NewRealmCatalog(world.DefaultRealmDefinitions), nil)
	h.SetPublisher(pub)

	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.CompleteChallenge(context.Background(), 1, 12, "crew-1", "user-1", "")
		}()
	}
	wg.Wait()

	questCompletedEvents := 0
	for _, ev := range pub.published {
		if _, ok := ev.(events.QuestCompletedEvent); ok {
			questCompletedEvents++
		}
	}
	if questCompletedEvents != 1 {
		t.Errorf("expected 1 QuestCompletedEvent, got %d", questCompletedEvents)
	}
}
func (m *concurrentUserStore) ListUsersByCrew(ctx context.Context, crewID string) ([]game.Player, error) {
	return nil, nil
}
