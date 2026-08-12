package events

import (
	"context"
	"sync"
	"testing"
)

type countingHandler struct {
	mu      sync.Mutex
	handled int
}

func (h *countingHandler) Handle(ctx context.Context, event Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handled++
	return nil
}

func TestDispatcher_ReplaySafe_HandlerIdempotent(t *testing.T) {
	d := NewDispatcher()
	h := &countingHandler{}
	d.Subscribe(EventTypeQuestCompleted, h)

	for i := 0; i < 5; i++ {
		d.Publish(context.Background(), QuestCompletedEvent{MissionID: 1, FamilyID: "c1"})
	}

	if h.handled != 5 {
		t.Errorf("expected handler called 5 times on replay, got %d", h.handled)
	}
}

func TestDispatcher_MultipleSubscribers_AllReceiveEvent(t *testing.T) {
	d := NewDispatcher()
	h1 := &countingHandler{}
	h2 := &countingHandler{}
	d.Subscribe(EventTypeQuestCompleted, h1)
	d.Subscribe(EventTypeQuestCompleted, h2)

	d.Publish(context.Background(), QuestCompletedEvent{MissionID: 1, FamilyID: "c1"})

	if h1.handled != 1 {
		t.Errorf("expected handler 1 called once, got %d", h1.handled)
	}
	if h2.handled != 1 {
		t.Errorf("expected handler 2 called once, got %d", h2.handled)
	}
}

func TestDispatcher_Close_StopsDelivery(t *testing.T) {
	d := NewDispatcher()
	h := &countingHandler{}
	d.Subscribe(EventTypeQuestCompleted, h)

	d.Close()
	d.Publish(context.Background(), QuestCompletedEvent{MissionID: 1, FamilyID: "c1"})

	if h.handled != 0 {
		t.Errorf("expected 0 deliveries after close, got %d", h.handled)
	}
}

func TestDispatcher_EventTypeMatching(t *testing.T) {
	d := NewDispatcher()
	qHandler := &countingHandler{}
	cHandler := &countingHandler{}
	d.Subscribe(EventTypeQuestCompleted, qHandler)
	d.Subscribe(EventTypeChapterCompleted, cHandler)

	d.Publish(context.Background(), QuestCompletedEvent{MissionID: 1, FamilyID: "c1"})
	d.Publish(context.Background(), ChapterCompletedEvent{FamilyID: "c1", Course: "ch-1"})

	if qHandler.handled != 1 {
		t.Errorf("expected quest handler called once, got %d", qHandler.handled)
	}
	if cHandler.handled != 1 {
		t.Errorf("expected course handler called once, got %d", cHandler.handled)
	}
}
