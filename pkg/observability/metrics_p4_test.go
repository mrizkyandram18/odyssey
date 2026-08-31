package observability

import (
	"testing"
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

func TestMetrics_RecordTaskSubmitted(t *testing.T) {
	m := NewMetrics()
	m.RecordBusinessEvent("task_submitted")
	m.RecordBusinessEvent("task_submitted")
	m.RecordBusinessEvent("task_submitted")

	if m.Snapshot().TaskSubmitted != 3 {
		t.Errorf("expected 3 task submissions, got %d", m.Snapshot().TaskSubmitted)
	}
}

func TestMetrics_RecordRewardRedeemed(t *testing.T) {
	m := NewMetrics()
	m.RecordBusinessEvent("reward_redeemed")
	m.RecordBusinessEvent("reward_redeemed")

	if m.Snapshot().RewardRedeemed != 2 {
		t.Errorf("expected 2 rewards redeemed, got %d", m.Snapshot().RewardRedeemed)
	}
}

func TestMetrics_RecordDuplicates(t *testing.T) {
	m := NewMetrics()
	m.RecordDuplicatePrevented(2)

	snap := m.Snapshot()
	if snap.DuplicatesPrevented != 2 {
		t.Errorf("expected 2 duplicates prevented, got %d", snap.DuplicatesPrevented)
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
