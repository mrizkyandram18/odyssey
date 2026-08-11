package events

import (
	"context"
	"sync"
	"time"
)

type Event interface {
	EventType() string
}

const (
	EventTypeQuestCompleted     = "quest_completed"
	EventTypeChapterCompleted   = "chapter_completed"
	EventTypeRelicCollected     = "relic_collected"
	EventTypeDailyTurnCompleted = "daily_turn_completed"
	EventTypeLevelReached       = "level_reached"
	EventTypeCreativeSubmission = "creative_submission"
	EventTypeRelayHandoff       = "relay_handoff"
)

type QuestCompletedEvent struct {
	QuestID      int64
	CrewID       string
	TemplateSlug string
	Realm        string
	Chapter      string
	SeasonSlug   string
	PlayerUID    string
}

func (e QuestCompletedEvent) EventType() string { return EventTypeQuestCompleted }

type ChapterCompletedEvent struct {
	CrewID     string
	Chapter    string
	Realm      string
	SeasonSlug string
	PlayerUID  string
}

func (e ChapterCompletedEvent) EventType() string { return EventTypeChapterCompleted }

type RelicCollectedEvent struct {
	UID       string
	CrewID    string
	RelicSlug string
	Realm     string
}

func (e RelicCollectedEvent) EventType() string { return EventTypeRelicCollected }

type DailyTurnCompletedEvent struct {
	UID        string
	StreakDays int
}

func (e DailyTurnCompletedEvent) EventType() string { return EventTypeDailyTurnCompleted }

// RelayHandoffEvent is published when a relay quest leg is assigned to the next
// explorer. It is emitted only after the assignment has been persisted.
type RelayHandoffEvent struct {
	FromUID    string // Explorer who completed the previous leg
	ToUID      string // Explorer now assigned the next leg
	QuestID    int64
	QuestTitle string
	CrewID     string
}

func (e RelayHandoffEvent) EventType() string { return EventTypeRelayHandoff }

type LevelReachedEvent struct {
	UID      string
	CrewID   string
	OldLevel int
	NewLevel int
}

func (e LevelReachedEvent) EventType() string { return EventTypeLevelReached }

type CreativeSubmissionEvent struct {
	UID         string
	CrewID      string
	QuestID     int64
	ChallengeID int64
	Kind        string
}

func (e CreativeSubmissionEvent) EventType() string { return EventTypeCreativeSubmission }

type Publisher interface {
	Publish(ctx context.Context, event Event)
}

type Handler interface {
	Handle(ctx context.Context, event Event) error
}

// Recorder receives event-pipeline telemetry as events are dispatched.
// It is optional and never blocks dispatch. The observability Metrics type
// satisfies this interface structurally.
type Recorder interface {
	RecordEventPublished(eventType string)
	RecordEventHandler(eventType string, duration time.Duration, err error)
}

// EventDiagnostics aggregates pipeline counters for a single event type.
type EventDiagnostics struct {
	Published            int64   `json:"published"`
	Handled              int64   `json:"handled"`
	Errors               int64   `json:"errors"`
	HandlerCount         int     `json:"handler_count"`
	AvgHandlerDurationMs float64 `json:"avg_handler_duration_ms"`
}

type Dispatcher struct {
	mu        sync.RWMutex
	handlers  map[string][]Handler
	recorder  Recorder
	published map[string]int64
	handled   map[string]int64
	errors    map[string]int64
	latency   map[string]time.Duration
	calls     map[string]int64
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		handlers:  make(map[string][]Handler),
		published: make(map[string]int64),
		handled:   make(map[string]int64),
		errors:    make(map[string]int64),
		latency:   make(map[string]time.Duration),
		calls:     make(map[string]int64),
	}
}

// SetRecorder attaches an optional telemetry recorder for event dispatch.
func (d *Dispatcher) SetRecorder(r Recorder) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.recorder = r
}

func (d *Dispatcher) Subscribe(eventType string, h Handler) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers[eventType] = append(d.handlers[eventType], h)
}

func (d *Dispatcher) Publish(ctx context.Context, event Event) {
	etype := event.EventType()
	d.mu.RLock()
	hs := d.handlers[etype]
	recorder := d.recorder
	d.mu.RUnlock()

	if recorder != nil {
		recorder.RecordEventPublished(etype)
	}
	d.mu.Lock()
	d.published[etype]++
	d.mu.Unlock()

	for _, h := range hs {
		start := time.Now()
		err := h.Handle(ctx, event)
		duration := time.Since(start)

		if recorder != nil {
			recorder.RecordEventHandler(etype, duration, err)
		}
		d.mu.Lock()
		d.handled[etype]++
		d.latency[etype] += duration
		d.calls[etype]++
		if err != nil {
			d.errors[etype]++
		}
		d.mu.Unlock()
	}
}

// Diagnostics returns per-event-type pipeline counters. HandlerCount reflects
// the number of subscribed handlers; Published/Handled/Errors and average
// handler latency reflect dispatch activity.
func (d *Dispatcher) Diagnostics() map[string]EventDiagnostics {
	d.mu.RLock()
	defer d.mu.RUnlock()

	out := make(map[string]EventDiagnostics, len(d.handlers))
	for etype := range d.handlers {
		st := EventDiagnostics{
			Published:    d.published[etype],
			Handled:      d.handled[etype],
			Errors:       d.errors[etype],
			HandlerCount: len(d.handlers[etype]),
		}
		if d.calls[etype] > 0 {
			st.AvgHandlerDurationMs = float64(d.latency[etype].Nanoseconds()) / float64(d.calls[etype]) / 1e6
		}
		out[etype] = st
	}
	return out
}

func (d *Dispatcher) Close() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.handlers = make(map[string][]Handler)
	d.published = make(map[string]int64)
	d.handled = make(map[string]int64)
	d.errors = make(map[string]int64)
	d.latency = make(map[string]time.Duration)
	d.calls = make(map[string]int64)
	d.recorder = nil
}

type NopPublisher struct{}

func (NopPublisher) Publish(ctx context.Context, event Event) {}

type NopHandler struct{}

func (NopHandler) Handle(ctx context.Context, event Event) error { return nil }
