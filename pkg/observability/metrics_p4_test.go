package observability

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMetrics_RecordXP(t *testing.T) {
	m := NewMetrics()
	m.RecordXP(60)
	m.RecordXP(0)
	m.RecordXP(20)

	snap := m.Snapshot()
	if snap.XPAwarded != 80 {
		t.Errorf("expected xp_awarded 80, got %d", snap.XPAwarded)
	}
}

func TestMetrics_RecordRealmCompleted(t *testing.T) {
	m := NewMetrics()
	m.RecordRealmCompleted()
	m.RecordRealmCompleted()
	m.RecordRealmCompleted()

	if m.Snapshot().RealmCompleted != 3 {
		t.Errorf("expected 3 journey completions, got %d", m.Snapshot().RealmCompleted)
	}
}

func TestMetrics_RecordChestCreated(t *testing.T) {
	m := NewMetrics()
	m.RecordChestCreated()
	m.RecordChestCreated()

	if m.Snapshot().ChestsCreated != 2 {
		t.Errorf("expected 2 gifts created, got %d", m.Snapshot().ChestsCreated)
	}
}

func TestMetrics_RecordRewardsAndDuplicates(t *testing.T) {
	m := NewMetrics()
	m.RecordRewardsGenerated(5)
	m.RecordRewardsGenerated(0)
	m.RecordRewardsGenerated(3)
	m.RecordDuplicatePrevented(2)

	snap := m.Snapshot()
	if snap.RewardsGenerated != 8 {
		t.Errorf("expected 8 rewards generated, got %d", snap.RewardsGenerated)
	}
	if snap.DuplicatesPrevented != 2 {
		t.Errorf("expected 2 duplicates prevented, got %d", snap.DuplicatesPrevented)
	}
}

func TestMetrics_RecordLockAndReplay(t *testing.T) {
	m := NewMetrics()
	m.RecordLockConflict()
	m.RecordLockConflict()
	m.RecordReplayIgnored()

	snap := m.Snapshot()
	if snap.LockConflicts != 2 {
		t.Errorf("expected 2 lock conflicts, got %d", snap.LockConflicts)
	}
	if snap.ReplayIgnored != 1 {
		t.Errorf("expected 1 replay ignored, got %d", snap.ReplayIgnored)
	}
}

func TestMetrics_RecordValidationFailure(t *testing.T) {
	m := NewMetrics()
	m.RecordValidationFailure()
	m.RecordValidationFailure()

	if m.Snapshot().ValidationFailures != 2 {
		t.Errorf("expected 2 validation failures, got %d", m.Snapshot().ValidationFailures)
	}
}

func TestMetrics_RecordEventPipeline(t *testing.T) {
	m := NewMetrics()
	m.RecordEventPublished("quest_completed")
	m.RecordEventPublished("quest_completed")
	m.RecordEventPublished("course_completed")
	m.RecordEventHandler("quest_completed", 2*time.Millisecond, nil)
	m.RecordEventHandler("quest_completed", 4*time.Millisecond, nil)
	m.RecordEventHandler("course_completed", 1*time.Millisecond, nil)
	m.RecordEventHandler("course_completed", 0, errHandler)

	snap := m.Snapshot()
	if snap.EventsPublished != 3 {
		t.Errorf("expected 3 events published, got %d", snap.EventsPublished)
	}
	if snap.EventsHandled != 4 {
		t.Errorf("expected 4 events handled, got %d", snap.EventsHandled)
	}
	if snap.EventsHandlerErrors != 1 {
		t.Errorf("expected 1 handler error, got %d", snap.EventsHandlerErrors)
	}
	et := snap.EventTypes["quest_completed"]
	if et.Published != 2 {
		t.Errorf("expected 2 published for quest_completed, got %d", et.Published)
	}
	if et.Handled != 2 {
		t.Errorf("expected 2 handled for quest_completed, got %d", et.Handled)
	}
	if et.Errors != 0 {
		t.Errorf("expected 0 errors for quest_completed, got %d", et.Errors)
	}
	if snap.EventsHandlerAvgMs <= 0 || snap.EventsHandlerAvgMs > 5 {
		t.Errorf("expected avg latency in (0,5], got %f", snap.EventsHandlerAvgMs)
	}

	data, err := m.SnapshotJSON()
	if err != nil {
		t.Fatalf("SnapshotJSON error: %v", err)
	}
	var dec struct {
		EventTypes map[string]EventTypeStat `json:"event_types"`
	}
	if err := json.Unmarshal(data, &dec); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if dec.EventTypes["course_completed"].Errors != 1 {
		t.Errorf("expected 1 error in JSON for course_completed, got %d", dec.EventTypes["course_completed"].Errors)
	}
}

func TestMetrics_BootTime(t *testing.T) {
	m := NewMetrics()
	bt := m.BootTime()
	if bt.IsZero() {
		t.Error("expected non-zero boot time")
	}
	snap := m.Snapshot()
	if snap.BootTime == "" {
		t.Error("expected non-empty boot_time in snapshot")
	}
}

var errHandler = errSentinel("handler boom")

type errSentinel string

func (e errSentinel) Error() string { return string(e) }
