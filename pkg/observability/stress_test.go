package observability

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

// TestObservability_Wrap_Concurrent exercises the middleware under concurrent
// load to validate the goroutine-safety of metrics, profiler, and the response
// writer pool. Run with -race.
func TestObservability_Wrap_Concurrent(t *testing.T) {
	var buf testBuffer
	obs := &Observability{
		Logger:   NewLogger(&buf),
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})
	handler := obs.Wrap(inner)

	const workers = 50
	const perWorker = 40
	var wg sync.WaitGroup
	start := make(chan struct{})
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			<-start
			for j := 0; j < perWorker; j++ {
				req := httptest.NewRequest(http.MethodPost, "/api/tasks/1/submit", nil)
				req.Header.Set(HeaderRequestID, "stress-req")
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, req)
			}
		}()
	}
	close(start)
	wg.Wait()

	obs.Logger.Flush()
	snap := obs.Metrics.Snapshot()
	total := int64(workers * perWorker)
	if snap.RequestCount["POST /api/tasks/1/submit"] != total {
		t.Errorf("expected %d requests, got %d", total, snap.RequestCount["POST /api/tasks/1/submit"])
	}
	if len(snap.RequestLatencyMs) == 0 {
		t.Error("expected request latency entries")
	}
}
