package events

import (
	"context"
	"sync"
	"testing"
)

type recordingHandler struct {
	mu      sync.Mutex
	events  []Event
	delayed bool
}

func (h *recordingHandler) Handle(ctx context.Context, event Event) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.delayed {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
	}
	h.events = append(h.events, event)
	return nil
}

func (h *recordingHandler) Events() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]Event, len(h.events))
	copy(out, h.events)
	return out
}

func TestDispatcher_PublishToSubscribedHandler(t *testing.T) {
	d := NewDispatcher()
	h := &recordingHandler{}
	d.Subscribe(EventTypeQuestCompleted, h)

	d.Publish(context.Background(), QuestCompletedEvent{MissionID: 1, FamilyID: "c1"})

	got := h.Events()
	if len(got) != 1 {
		t.Fatalf("expected 1 event, got %d", len(got))
	}
	ev, ok := got[0].(QuestCompletedEvent)
	if !ok {
		t.Fatalf("expected QuestCompletedEvent, got %T", got[0])
	}
	if ev.MissionID != 1 {
		t.Errorf("expected quest ID 1, got %d", ev.MissionID)
	}
}

func TestDispatcher_PublishToMultipleHandlers(t *testing.T) {
	d := NewDispatcher()
	h1 := &recordingHandler{}
	h2 := &recordingHandler{}
	d.Subscribe(EventTypeQuestCompleted, h1)
	d.Subscribe(EventTypeQuestCompleted, h2)

	d.Publish(context.Background(), QuestCompletedEvent{FamilyID: "c1"})

	if len(h1.Events()) != 1 {
		t.Errorf("expected handler 1 to receive 1 event, got %d", len(h1.Events()))
	}
	if len(h2.Events()) != 1 {
		t.Errorf("expected handler 2 to receive 1 event, got %d", len(h2.Events()))
	}
}

func TestDispatcher_NoHandlersForEventType(t *testing.T) {
	d := NewDispatcher()
	h := &recordingHandler{}
	d.Subscribe(EventTypeQuestCompleted, h)

	d.Publish(context.Background(), ChapterCompletedEvent{FamilyID: "c1"})

	if len(h.Events()) != 0 {
		t.Errorf("expected 0 events for unregistered type, got %d", len(h.Events()))
	}
}

func TestDispatcher_EventTypes(t *testing.T) {
	tests := []struct {
		name  string
		event Event
		want  string
	}{
		{"QuestCompleted", QuestCompletedEvent{}, EventTypeQuestCompleted},
		{"ChapterCompleted", ChapterCompletedEvent{}, EventTypeChapterCompleted},
		{"RelicCollected", RelicCollectedEvent{}, EventTypeRelicCollected},
		{"DailyTurnCompleted", DailyTurnCompletedEvent{}, EventTypeDailyTurnCompleted},
		{"LevelReached", LevelReachedEvent{}, EventTypeLevelReached},
		{"CreativeSubmission", CreativeSubmissionEvent{}, EventTypeCreativeSubmission},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.event.EventType(); got != tt.want {
				t.Errorf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestNopPublisher_Publish(t *testing.T) {
	p := NopPublisher{}
	p.Publish(context.Background(), QuestCompletedEvent{})
}

func TestNopHandler_Handle(t *testing.T) {
	h := NopHandler{}
	if err := h.Handle(context.Background(), QuestCompletedEvent{}); err != nil {
		t.Errorf("expected nil error, got %v", err)
	}
}

func TestDispatcher_Close(t *testing.T) {
	d := NewDispatcher()
	h := &recordingHandler{}
	d.Subscribe(EventTypeQuestCompleted, h)
	d.Close()

	d.Publish(context.Background(), QuestCompletedEvent{})

	if len(h.Events()) != 0 {
		t.Errorf("expected 0 events after close, got %d", len(h.Events()))
	}
}
