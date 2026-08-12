package observability

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMetrics_RecordRequest(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest("GET", "/api/missions", 200, 10*time.Millisecond)
	m.RecordRequest("POST", "/api/login", 401, 5*time.Millisecond)
	m.RecordRequest("GET", "/api/missions", 200, 15*time.Millisecond)

	snap := m.Snapshot()
	if snap.RequestCount["GET /api/missions"] != 2 {
		t.Errorf("expected 2 requests for GET /api/missions, got %d", snap.RequestCount["GET /api/missions"])
	}
	if snap.RequestCount["POST /api/login"] != 1 {
		t.Errorf("expected 1 request for POST /api/login, got %d", snap.RequestCount["POST /api/login"])
	}
}

func TestMetrics_RecordLogin(t *testing.T) {
	m := NewMetrics()
	m.RecordLogin(true)
	m.RecordLogin(true)
	m.RecordLogin(false)

	snap := m.Snapshot()
	if snap.LoginSuccess != 2 {
		t.Errorf("expected 2 login successes, got %d", snap.LoginSuccess)
	}
	if snap.LoginFailure != 1 {
		t.Errorf("expected 1 login failure, got %d", snap.LoginFailure)
	}
}

func TestMetrics_RecordCache(t *testing.T) {
	m := NewMetrics()
	m.RecordCacheHit()
	m.RecordCacheHit()
	m.RecordCacheMiss()

	snap := m.Snapshot()
	if snap.CacheHits != 2 {
		t.Errorf("expected 2 cache hits, got %d", snap.CacheHits)
	}
	if snap.CacheMisses != 1 {
		t.Errorf("expected 1 cache miss, got %d", snap.CacheMisses)
	}
}

func TestMetrics_RecordBusinessEvent(t *testing.T) {
	m := NewMetrics()
	m.RecordBusinessEvent("quest_completed")
	m.RecordBusinessEvent("quest_completed")
	m.RecordBusinessEvent("chest_opened")
	m.RecordBusinessEvent("creative_submitted")
	m.RecordBusinessEvent("unknown_event")

	snap := m.Snapshot()
	if snap.QuestCompleted != 2 {
		t.Errorf("expected 2 quest completions, got %d", snap.QuestCompleted)
	}
	if snap.ChestOpened != 1 {
		t.Errorf("expected 1 chest opened, got %d", snap.ChestOpened)
	}
	if snap.CreativeSubmitted != 1 {
		t.Errorf("expected 1 creative submitted, got %d", snap.CreativeSubmitted)
	}
}

func TestMetrics_RecordAdminOp(t *testing.T) {
	m := NewMetrics()
	m.RecordAdminOp()
	m.RecordAdminOp()
	m.RecordAdminOp()

	snap := m.Snapshot()
	if snap.AdminOps != 3 {
		t.Errorf("expected 3 admin ops, got %d", snap.AdminOps)
	}
}

func TestMetrics_RecordDBLatency(t *testing.T) {
	m := NewMetrics()
	m.RecordDBLatency(5 * time.Millisecond)
	m.RecordDBLatency(10 * time.Millisecond)
	m.RecordDBLatency(2 * time.Millisecond)

	snap := m.Snapshot()
	if snap.DBCallCount != 3 {
		t.Errorf("expected 3 DB calls, got %d", snap.DBCallCount)
	}
	avg := snap.DBAvgLatencyMs
	if avg < 2 || avg > 10 {
		t.Errorf("expected avg latency between 2-10ms, got %f", avg)
	}
}

func TestMetrics_RecordDBLatency_Slow(t *testing.T) {
	m := NewMetrics()
	m.RecordDBLatency(200 * time.Millisecond)
	m.RecordDBLatency(50 * time.Millisecond)

	snap := m.Snapshot()
	if snap.DBSlowCount != 1 {
		t.Errorf("expected 1 slow query, got %d", snap.DBSlowCount)
	}
}

func TestMetrics_MergeCacheStats(t *testing.T) {
	m := NewMetrics()
	m.RecordCacheHit()
	m.MergeCacheStats(5, 3, 0)

	snap := m.Snapshot()
	if snap.CacheHits != 6 {
		t.Errorf("expected 6 cache hits, got %d", snap.CacheHits)
	}
	if snap.CacheMisses != 3 {
		t.Errorf("expected 3 cache misses, got %d", snap.CacheMisses)
	}
}

func TestMetrics_SnapshotJSON(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest("GET", "/api/test", 200, 5*time.Millisecond)
	m.RecordLogin(true)

	data, err := m.SnapshotJSON()
	if err != nil {
		t.Fatalf("SnapshotJSON error: %v", err)
	}

	var snap MetricsSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if snap.LoginSuccess != 1 {
		t.Errorf("expected 1 login success, got %d", snap.LoginSuccess)
	}
	if snap.BootTime == "" {
		t.Error("expected non-empty boot_time")
	}
	if snap.UptimeSeconds < 0 {
		t.Errorf("expected non-negative uptime, got %d", snap.UptimeSeconds)
	}
}

func TestMetricsHandler(t *testing.T) {
	m := NewMetrics()
	m.RecordRequest("GET", "/api/test", 200, 5*time.Millisecond)
	m.RecordLogin(true)
	m.RecordLogin(false)

	handler := MetricsHandler(m)
	snap := handler
	_ = snap
}

func TestMetrics_Concurrency(t *testing.T) {
	m := NewMetrics()
	done := make(chan struct{})
	for i := 0; i < 100; i++ {
		go func() {
			m.RecordRequest("GET", "/api/test", 200, time.Millisecond)
			m.RecordLogin(true)
			m.RecordCacheHit()
			m.RecordDBLatency(time.Millisecond)
			done <- struct{}{}
		}()
	}
	for i := 0; i < 100; i++ {
		<-done
	}

	snap := m.Snapshot()
	if snap.RequestCount["GET /api/test"] != 100 {
		t.Errorf("expected 100 requests, got %d", snap.RequestCount["GET /api/test"])
	}
	if snap.LoginSuccess != 100 {
		t.Errorf("expected 100 login successes, got %d", snap.LoginSuccess)
	}
	if snap.CacheHits != 100 {
		t.Errorf("expected 100 cache hits, got %d", snap.CacheHits)
	}
	if snap.DBCallCount != 100 {
		t.Errorf("expected 100 DB calls, got %d", snap.DBCallCount)
	}
}
