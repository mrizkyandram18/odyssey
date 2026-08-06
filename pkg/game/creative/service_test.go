package creative

import (
	"context"
	"errors"
	"testing"
	"time"

	"odyssey/pkg/game"
	"odyssey/pkg/game/events"
)

type mockCreativeSubmissionStore struct {
	subs    map[int64]*game.Submission
	byQuest map[int64][]game.Submission
	byCrew  map[string][]game.Submission
	err     error
}

func newMockSubmissionStore() *mockCreativeSubmissionStore {
	return &mockCreativeSubmissionStore{
		subs:    make(map[int64]*game.Submission),
		byQuest: make(map[int64][]game.Submission),
		byCrew:  make(map[string][]game.Submission),
	}
}

func (m *mockCreativeSubmissionStore) CreateSubmission(ctx context.Context, s *game.Submission) (*game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	s.ID = int64(len(m.subs) + 1)
	s.CreatedAt = time.Now().UTC()
	s.UpdatedAt = s.CreatedAt
	m.subs[s.ID] = s
	m.byQuest[s.QuestID] = append(m.byQuest[s.QuestID], *s)
	m.byCrew[s.CrewID] = append(m.byCrew[s.CrewID], *s)
	return s, nil
}

func (m *mockCreativeSubmissionStore) ListByQuest(ctx context.Context, questID int64) ([]game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byQuest[questID], nil
}

func (m *mockCreativeSubmissionStore) ListByCrew(ctx context.Context, crewID string) ([]game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.byCrew[crewID], nil
}

func (m *mockCreativeSubmissionStore) GetSubmission(ctx context.Context, submissionID int64) (*game.Submission, error) {
	if m.err != nil {
		return nil, m.err
	}
	s, ok := m.subs[submissionID]
	if !ok {
		return nil, game.ErrNotFound
	}
	return s, nil
}

func (m *mockCreativeSubmissionStore) UpdateSubmission(ctx context.Context, submissionID int64, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	s, ok := m.subs[submissionID]
	if !ok {
		return game.ErrNotFound
	}
	if status, ok := patch["status"]; ok {
		s.Status = game.SubmissionStatus(status.(string))
	}
	if reviewedBy, ok := patch["reviewed_by"]; ok {
		s.ReviewedBy = reviewedBy.(string)
	}
	if reviewedAt, ok := patch["reviewed_at"]; ok {
		if t, ok := reviewedAt.(time.Time); ok {
			s.ReviewedAt = &t
		}
	}
	if rejectionReason, ok := patch["rejection_reason"]; ok {
		s.RejectionReason = rejectionReason.(string)
	}
	s.UpdatedAt = time.Now().UTC()
	return nil
}

type mockQuestStore struct {
	quests     map[int64]*game.Quest
	challenges map[int64][]game.Challenge
	err        error
}

func newMockQuestStore() *mockQuestStore {
	return &mockQuestStore{
		quests:     make(map[int64]*game.Quest),
		challenges: make(map[int64][]game.Challenge),
	}
}

func (m *mockQuestStore) GetQuest(ctx context.Context, questID int64) (*game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	q, ok := m.quests[questID]
	if !ok {
		return nil, game.ErrNotFound
	}
	return q, nil
}

func (m *mockQuestStore) CreateQuest(ctx context.Context, q *game.Quest) (*game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	q.ID = int64(len(m.quests) + 1)
	q.CreatedAt = time.Now().UTC()
	m.quests[q.ID] = q
	return q, nil
}

func (m *mockQuestStore) ListQuestByCrew(ctx context.Context, crewID string) ([]game.Quest, error) {
	if m.err != nil {
		return nil, m.err
	}
	var result []game.Quest
	for _, q := range m.quests {
		if q.CrewID == crewID {
			result = append(result, *q)
		}
	}
	return result, nil
}

func (m *mockQuestStore) UpdateQuest(ctx context.Context, questID int64, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	q, ok := m.quests[questID]
	if !ok {
		return game.ErrNotFound
	}
	if status, ok := patch["status"]; ok {
		q.Status = status.(string)
	}
	return nil
}
func (m *mockQuestStore) UpdateQuestIfMatch(ctx context.Context, questID int64, oldStatus string, patch map[string]any) (bool, error) {
	if m.err != nil {
		return false, m.err
	}
	q, ok := m.quests[questID]
	if !ok {
		return false, game.ErrNotFound
	}
	if q.Status != oldStatus {
		return false, nil
	}
	if err := m.UpdateQuest(ctx, questID, patch); err != nil {
		return false, err
	}
	return true, nil
}

func (m *mockQuestStore) GetChallenges(ctx context.Context, questID int64) ([]game.Challenge, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.challenges[questID], nil
}

func (m *mockQuestStore) CreateChallenge(ctx context.Context, c *game.Challenge) (*game.Challenge, error) {
	if m.err != nil {
		return nil, m.err
	}
	c.ID = int64(len(m.challenges[c.QuestID]) + 1)
	m.challenges[c.QuestID] = append(m.challenges[c.QuestID], *c)
	return c, nil
}

func (m *mockQuestStore) UpdateChallenge(ctx context.Context, challengeID int64, patch map[string]any) error {
	if m.err != nil {
		return m.err
	}
	for _, chs := range m.challenges {
		for i := range chs {
			if chs[i].ID == challengeID {
				if status, ok := patch["status"]; ok {
					chs[i].Status = status.(string)
				}
				return nil
			}
		}
	}
	return game.ErrNotFound
}

func makeQuest(id int64, status string, crewID string) *game.Quest {
	return &game.Quest{
		ID:        id,
		CrewID:    crewID,
		Title:     "Test Quest",
		Status:    status,
		CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func makeChallenge(id int64, questID int64, status string) game.Challenge {
	return game.Challenge{
		ID:        id,
		QuestID:   questID,
		Slug:      "challenge-1",
		Status:    status,
		CreatedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC),
	}
}

func TestSubmit_QuestNotFound(t *testing.T) {
	svc := NewCreativeService(newMockSubmissionStore(), newMockQuestStore())
	_, err := svc.Submit(context.Background(), &game.Submission{QuestID: 999})
	if !errors.Is(err, ErrQuestNotFound) {
		t.Fatalf("expected ErrQuestNotFound, got %v", err)
	}
}

func TestSubmit_QuestNotActive(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "PENDING", "c1"))
	svc := NewCreativeService(newMockSubmissionStore(), qs)
	_, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "hello"})
	if !errors.Is(err, ErrQuestNotActive) {
		t.Fatalf("expected ErrQuestNotActive, got %v", err)
	}
}

func TestSubmit_ChallengeDone(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "DONE")}
	svc := NewCreativeService(newMockSubmissionStore(), qs)
	_, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "hello"})
	if !errors.Is(err, ErrChallengeDone) {
		t.Fatalf("expected ErrChallengeDone, got %v", err)
	}
}

func TestSubmit_ChallengeNotFound(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	svc := NewCreativeService(newMockSubmissionStore(), qs)
	_, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 999, Kind: game.SubmissionStory, Content: "hello"})
	if !errors.Is(err, ErrChallengeNotFound) {
		t.Fatalf("expected ErrChallengeNotFound, got %v", err)
	}
}

func TestSubmit_InvalidKind(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	svc := NewCreativeService(newMockSubmissionStore(), qs)
	_, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: "INVALID", Content: "hello"})
	if !errors.Is(err, ErrInvalidKind) {
		t.Fatalf("expected ErrInvalidKind, got %v", err)
	}
}

func TestSubmit_ContentTooShort(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	svc := NewCreativeService(newMockSubmissionStore(), qs)
	_, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: ""})
	if !errors.Is(err, ErrContentTooShort) {
		t.Fatalf("expected ErrContentTooShort, got %v", err)
	}
}

func TestSubmit_Success(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	store := newMockSubmissionStore()
	svc := NewCreativeService(store, qs)
	sub, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "hello world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != game.SubmissionStatusPending {
		t.Errorf("expected status PENDING, got %s", sub.Status)
	}
	if sub.Content != "hello world" {
		t.Errorf("expected content 'hello world', got %s", sub.Content)
	}
	if sub.AuthorUID != "" {
		t.Errorf("expected empty author uid before auth injection, got %s", sub.AuthorUID)
	}
}

func TestListByQuest(t *testing.T) {
	store := newMockSubmissionStore()
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	svc := NewCreativeService(store, qs)
	svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "story 1"})
	svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionComic, Content: "comic 1"})

	subs, err := svc.ListByQuest(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(subs) != 2 {
		t.Fatalf("expected 2 submissions, got %d", len(subs))
	}
}

func TestGetSubmission(t *testing.T) {
	store := newMockSubmissionStore()
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	svc := NewCreativeService(store, qs)
	created, _ := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "story 1"})

	sub, err := svc.GetSubmission(context.Background(), created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Content != "story 1" {
		t.Errorf("expected content 'story 1', got %s", sub.Content)
	}
}

func TestGetSubmission_NotFound(t *testing.T) {
	svc := NewCreativeService(newMockSubmissionStore(), newMockQuestStore())
	_, err := svc.GetSubmission(context.Background(), 999)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestApprove(t *testing.T) {
	store := newMockSubmissionStore()
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	svc := NewCreativeService(store, qs)
	created, _ := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "story 1"})

	sub, err := svc.Approve(context.Background(), created.ID, "reviewer-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != game.SubmissionStatusApproved {
		t.Errorf("expected status APPROVED, got %s", sub.Status)
	}
	if sub.ReviewedBy != "reviewer-1" {
		t.Errorf("expected reviewer 'reviewer-1', got %s", sub.ReviewedBy)
	}
	if sub.ReviewedAt == nil {
		t.Error("expected reviewed_at to be set")
	}
}

func TestReject(t *testing.T) {
	store := newMockSubmissionStore()
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	svc := NewCreativeService(store, qs)
	created, _ := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "story 1"})

	sub, err := svc.Reject(context.Background(), created.ID, "reviewer-1", "not creative enough")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != game.SubmissionStatusRejected {
		t.Errorf("expected status REJECTED, got %s", sub.Status)
	}
	if sub.RejectionReason != "not creative enough" {
		t.Errorf("expected rejection reason 'not creative enough', got %s", sub.RejectionReason)
	}
}

type capturePublisher struct {
	events []events.Event
}

func (p *capturePublisher) Publish(ctx context.Context, event events.Event) {
	p.events = append(p.events, event)
}

func TestSubmit_PublishesCreativeSubmissionEvent(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	store := newMockSubmissionStore()
	pub := &capturePublisher{}
	svc := NewCreativeServiceWithPublisher(store, qs, pub)

	sub, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "hello world", AuthorUID: "user-1", CrewID: "crew-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sub.Status != game.SubmissionStatusPending {
		t.Errorf("expected status PENDING, got %s", sub.Status)
	}
	if len(pub.events) != 1 {
		t.Fatalf("expected 1 event published, got %d", len(pub.events))
	}
	ev, ok := pub.events[0].(events.CreativeSubmissionEvent)
	if !ok {
		t.Fatalf("expected CreativeSubmissionEvent, got %T", pub.events[0])
	}
	if ev.UID != "user-1" {
		t.Errorf("expected UID user-1, got %s", ev.UID)
	}
	if ev.CrewID != "crew-1" {
		t.Errorf("expected CrewID crew-1, got %s", ev.CrewID)
	}
	if ev.QuestID != 1 {
		t.Errorf("expected QuestID 1, got %d", ev.QuestID)
	}
	if ev.ChallengeID != 1 {
		t.Errorf("expected ChallengeID 1, got %d", ev.ChallengeID)
	}
	if ev.Kind != string(game.SubmissionStory) {
		t.Errorf("expected kind STORY, got %s", ev.Kind)
	}
}

func TestSubmit_NoPublisher_NoEvent(t *testing.T) {
	qs := newMockQuestStore()
	qs.CreateQuest(context.Background(), makeQuest(1, "ACTIVE", "c1"))
	qs.challenges[1] = []game.Challenge{makeChallenge(1, 1, "PENDING")}
	store := newMockSubmissionStore()
	svc := NewCreativeService(store, qs)

	_, err := svc.Submit(context.Background(), &game.Submission{QuestID: 1, ChallengeID: 1, Kind: game.SubmissionStory, Content: "hello world", AuthorUID: "user-1", CrewID: "crew-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
