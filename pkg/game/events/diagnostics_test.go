package events

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type testEvent struct {
	et string
}

func (e testEvent) EventType() string { return e.et }

type captureHandler struct {
	calls int
	err   error
	delay time.Duration
	mu    sync.Mutex
}

func (h *captureHandler) Handle(ctx context.Context, event Event) error {
	h.mu.Lock()
	h.calls++
	h.mu.Unlock()
	if h.delay > 0 {
		time.Sleep(h.delay)
	}
	return h.err
}

func (h *captureHandler) Calls() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.calls
}

func TestDispatcher_Diagnostics(t *testing.T) {
	d := NewDispatcher()
	ok1 := &captureHandler{}
	err1 := &captureHandler{err: errors.New("boom")}
	d.Subscribe("quest_completed", ok1)
	d.Subscribe("quest_completed", err1)
	d.Subscribe("chapter_completed", &captureHandler{})

	d.Publish(context.Background(), testEvent{et: "quest_completed"})
	d.Publish(context.Background(), testEvent{et: "quest_completed"})
	d.Publish(context.Background(), testEvent{et: "chapter_completed"})

	diag := d.Diagnostics()
	qc := diag["quest_completed"]
	if qc.Published != 2 {
		t.Errorf("quest_completed published = %d, want 2", qc.Published)
	}
	if qc.Handled != 4 {
		t.Errorf("quest_completed handled = %d, want 4 (2 events x 2 handlers)", qc.Handled)
	}
	if qc.Errors != 2 {
		t.Errorf("quest_completed errors = %d, want 2", qc.Errors)
	}
	if qc.HandlerCount != 2 {
		t.Errorf("quest_completed handler_count = %d, want 2", qc.HandlerCount)
	}
	if qc.AvgHandlerDurationMs < 0 {
		t.Error("expected non-negative avg duration")
	}
	cc := diag["chapter_completed"]
	if cc.Published != 1 {
		t.Errorf("chapter_completed published = %d, want 1", cc.Published)
	}
	if cc.Handled != 1 {
		t.Errorf("chapter_completed handled = %d, want 1", cc.Handled)
	}
	if cc.HandlerCount != 1 {
		t.Errorf("chapter_completed handler_count = %d, want 1", cc.HandlerCount)
	}
	if len(diag) != 2 {
		t.Errorf("expected 2 event types in diagnostics, got %d", len(diag))
	}
}

type recordingSink struct {
	mu         sync.Mutex
	published  map[string]int
	handled    map[string]int
	errors     map[string]int
	latencySum time.Duration
	calls      int
}

func (r *recordingSink) Init() {
	r.published = make(map[string]int)
	r.handled = make(map[string]int)
	r.errors = make(map[string]int)
}

func (r *recordingSink) RecordEventPublished(eventType string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.published[eventType]++
}

func (r *recordingSink) RecordEventHandler(eventType string, duration time.Duration, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handled[eventType]++
	r.calls++
	r.latencySum += duration
	if err != nil {
		r.errors[eventType]++
	}
}

func TestDispatcher_Recorder(t *testing.T) {
	d := NewDispatcher()
	sink := &recordingSink{}
	sink.Init()
	d.SetRecorder(sink)

	h := &captureHandler{}
	d.Subscribe("quest_completed", h)
	d.Publish(context.Background(), testEvent{et: "quest_completed"})

	sink.mu.Lock()
	pub := sink.published["quest_completed"]
	hnd := sink.handled["quest_completed"]
	sink.mu.Unlock()

	if pub != 1 {
		t.Errorf("expected 1 published to recorder, got %d", pub)
	}
	if hnd != 1 {
		t.Errorf("expected 1 handled to recorder, got %d", hnd)
	}
	if h.Calls() != 1 {
		t.Errorf("expected handler invoked once, got %d", h.Calls())
	}
}

func TestDispatcher_Concurrency(t *testing.T) {
	d := NewDispatcher()
	var sum sync.WaitGroup
	var counter int64
	var mu sync.Mutex
	d.Subscribe("quest_completed", HandlerFunc(func(ctx context.Context, event Event) error {
		mu.Lock()
		counter++
		mu.Unlock()
		return nil
	}))

	const goroutines = 50
	const perG = 40
	for i := 0; i < goroutines; i++ {
		sum.Add(perG)
		go func() {
			for j := 0; j < perG; j++ {
				d.Publish(context.Background(), testEvent{et: "quest_completed"})
				sum.Done()
			}
		}()
	}
	sum.Wait()

	diag := d.Diagnostics()
	got := diag["quest_completed"].Handled
	if got != int64(goroutines*perG) {
		t.Errorf("expected %d handled, got %d", goroutines*perG, got)
	}
	if diag["quest_completed"].Published != int64(goroutines*perG) {
		t.Errorf("expected %d published, got %d", goroutines*perG, diag["quest_completed"].Published)
	}
}

// HandlerFunc adapts a function to the Handler interface.
type HandlerFunc func(ctx context.Context, event Event) error

func (f HandlerFunc) Handle(ctx context.Context, event Event) error { return f(ctx, event) }
