package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockInnerClient struct {
	err error
}

func (m *mockInnerClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(`[]`), nil
}

func (m *mockInnerClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(`[]`), nil
}

func (m *mockInnerClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(`[]`), nil
}

func TestProfilingClient_Get(t *testing.T) {
	inner := &mockInnerClient{}
	pc := NewProfilingClient(inner, nil)
	_, err := pc.Get(context.Background(), "test_table", "limit=1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProfilingClient_Mutate(t *testing.T) {
	inner := &mockInnerClient{}
	pc := NewProfilingClient(inner, nil)
	_, err := pc.Mutate(context.Background(), "POST", "test_table", map[string]any{"x": 1}, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProfilingClient_MutateAtomic(t *testing.T) {
	inner := &mockInnerClient{}
	pc := NewProfilingClient(inner, nil)
	_, err := pc.MutateAtomic(context.Background(), "PATCH", "test_table", map[string]any{"x": 1}, "", "return=representation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestProfilingClient_RecordsLatency(t *testing.T) {
	inner := &mockInnerClient{}
	metrics := NewMetrics()
	pc := NewProfilingClient(inner, metrics)
	_, _ = pc.Get(context.Background(), "test_table", "limit=1")
	_, _ = pc.Mutate(context.Background(), "POST", "test_table", nil, "")
	snap := metrics.Snapshot()
	if snap.DBCallCount != 2 {
		t.Errorf("expected 2 DB calls recorded, got %d", snap.DBCallCount)
	}
}

func TestProfilingClient_RecordsProfilingContext(t *testing.T) {
	inner := &mockInnerClient{}
	pc := NewProfilingClient(inner, nil)
	profiler := NewProfiler()
	rp := profiler.NewRequestProfile("/api/test", "GET")
	ctx := withProfile(context.Background(), rp)
	_, _ = pc.Get(ctx, "test_table", "limit=1")
	_, _ = pc.Get(ctx, "test_table", "limit=1")
	if rp.GetDBQueryCount() != 2 {
		t.Errorf("expected 2 DB queries recorded, got %d", rp.GetDBQueryCount())
	}
}

func TestProfilingMiddleware_SlowHandler(t *testing.T) {
	profiler := NewProfiler()
	metrics := NewMetrics()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := ProfilingMiddleware(profiler, DefaultSlowThreshold, metrics, inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	snap := metrics.Snapshot()
	if snap.RequestCount["GET /api/test"] != 1 {
		t.Errorf("expected 1 request recorded, got %d", snap.RequestCount["GET /api/test"])
	}
}

func TestProfilingMiddleware_FastHandler(t *testing.T) {
	profiler := NewProfiler()
	metrics := NewMetrics()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := ProfilingMiddleware(profiler, DefaultSlowThreshold, metrics, inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	snap := profiler.Snapshot()
	if len(snap.SlowHandlers) != 0 {
		t.Errorf("expected no slow handlers, got %d", len(snap.SlowHandlers))
	}
}

func TestProfiler_Snapshot_Empty(t *testing.T) {
	p := NewProfiler()
	snap := p.Snapshot()
	if len(snap.SlowHandlers) != 0 {
		t.Error("expected no slow handlers")
	}
	if len(snap.Recommendations) != 0 {
		t.Error("expected no recommendations")
	}
}

func TestProfiler_Reset(t *testing.T) {
	p := NewProfiler()
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := ProfilingMiddleware(p, 500*time.Millisecond, nil, inner)
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/slow", nil)
		w := httptest.NewRecorder()
		handler(w, req)
	}

	snap := p.Snapshot()
	if len(snap.SlowHandlers) != 3 {
		t.Fatalf("expected 3 slow handlers, got %d", len(snap.SlowHandlers))
	}

	p.Reset()
	snap = p.Snapshot()
	if len(snap.SlowHandlers) != 0 {
		t.Errorf("expected 0 slow handlers after reset, got %d", len(snap.SlowHandlers))
	}
}

func TestProfiler_Prune(t *testing.T) {
	p := NewProfiler()
	oldTime := time.Now().Add(-10 * time.Minute)

	p.mu.Lock()
	p.slowPaths = []slowEntry{
		{Endpoint: "/old", Method: "GET", Timestamp: oldTime},
		{Endpoint: "/recent", Method: "GET", Timestamp: time.Now()},
	}
	p.recommendations = []recEntry{
		{message: "old rec", timestamp: oldTime},
		{message: "recent rec", timestamp: time.Now()},
	}
	p.mu.Unlock()

	p.Prune()

	snap := p.Snapshot()
	if len(snap.SlowHandlers) != 1 {
		t.Errorf("expected 1 slow handler after prune, got %d", len(snap.SlowHandlers))
	}
	if snap.SlowHandlers[0].Endpoint != "/recent" {
		t.Errorf("expected /recent, got %s", snap.SlowHandlers[0].Endpoint)
	}
	if len(snap.Recommendations) != 1 {
		t.Errorf("expected 1 recommendation after prune, got %d", len(snap.Recommendations))
	}
	if snap.Recommendations[0] != "recent rec" {
		t.Errorf("expected 'recent rec', got %s", snap.Recommendations[0])
	}
}

func TestProfiler_PruneKeepsRecentEntries(t *testing.T) {
	p := NewProfiler()

	p.mu.Lock()
	p.slowPaths = []slowEntry{
		{Endpoint: "/recent1", Method: "GET", Timestamp: time.Now().Add(-1 * time.Minute)},
		{Endpoint: "/recent2", Method: "GET", Timestamp: time.Now().Add(-30 * time.Second)},
	}
	p.recommendations = []recEntry{
		{message: "recent rec1", timestamp: time.Now().Add(-1 * time.Minute)},
		{message: "recent rec2", timestamp: time.Now().Add(-10 * time.Second)},
	}
	p.mu.Unlock()

	p.Prune()

	snap := p.Snapshot()
	if len(snap.SlowHandlers) != 2 {
		t.Errorf("expected 2 slow handlers (within 5min), got %d", len(snap.SlowHandlers))
	}
	if len(snap.Recommendations) != 2 {
		t.Errorf("expected 2 recommendations (within 5min), got %d", len(snap.Recommendations))
	}
}

func TestProfileHandler(t *testing.T) {
	p := NewProfiler()
	p.AddSlowHandler("GET", "/api/slow", 600*time.Millisecond)

	req := httptest.NewRequest(http.MethodGet, "/debug/profile", nil)
	w := httptest.NewRecorder()
	ProfileHandler(p)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestProfileHandler_MethodNotAllowed(t *testing.T) {
	p := NewProfiler()
	req := httptest.NewRequest(http.MethodPost, "/debug/profile", nil)
	w := httptest.NewRecorder()
	ProfileHandler(p)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestRecommendationsHandler(t *testing.T) {
	p := NewProfiler()
	p.AddSlowHandler("GET", "/api/slow", 600*time.Millisecond)
	p.AddRecommendation("N+1 detected: test")

	req := httptest.NewRequest(http.MethodGet, "/debug/profile/recommendations", nil)
	w := httptest.NewRecorder()
	RecommendationsHandler(p)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
