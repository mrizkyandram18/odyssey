package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegration_MetricsEndpoint(t *testing.T) {
	obs := NewObservability()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		obs.RecordLogin(true)
		w.WriteHeader(http.StatusOK)
	})

	handler := obs.Wrap(http.HandlerFunc(inner))
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get(HeaderRequestID) == "" {
		t.Fatal("expected X-Request-ID header")
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	mw := httptest.NewRecorder()
	MetricsHandler(obs.Metrics)(mw, metricsReq)

	if mw.Code != http.StatusOK {
		t.Fatalf("expected 200 for metrics, got %d", mw.Code)
	}

	var snap MetricsSnapshot
	if err := json.Unmarshal(mw.Body.Bytes(), &snap); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if snap.RequestCount["GET /api/test"] != 1 {
		t.Errorf("expected 1 request, got %d", snap.RequestCount["GET /api/test"])
	}
	if snap.LoginSuccess != 1 {
		t.Errorf("expected 1 login success, got %d", snap.LoginSuccess)
	}
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "config", Fn: func(ctx context.Context) error { return nil }},
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return nil }},
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	HealthHandler(hc)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var hs HealthStatus
	if err := json.Unmarshal(w.Body.Bytes(), &hs); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if hs.Status != "ok" {
		t.Errorf("expected ok, got %s", hs.Status)
	}
	if hs.Checks["config"].Status != "pass" {
		t.Error("expected config check to pass")
	}
}

func TestIntegration_VersionEndpoint(t *testing.T) {
	GitCommit = "abcdef"
	BuildDate = "2024-01-15T10:30:00Z"
	Version = "1.0.0"
	SchemaVersion = "8"
	SetRuntimeSchemaVersion("")
	defer func() {
		GitCommit = "unknown"
		BuildDate = "unknown"
		Version = "dev"
		SchemaVersion = "0"
		SetRuntimeSchemaVersion("")
	}()

	bi := &StaticBuildInfo{ContentGen: 5}
	req := httptest.NewRequest(http.MethodGet, "/version", nil)
	w := httptest.NewRecorder()
	VersionHandler(bi)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var vi VersionInfo
	if err := json.Unmarshal(w.Body.Bytes(), &vi); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if vi.GitCommit != "abcdef" {
		t.Errorf("expected git_commit abcdef, got %s", vi.GitCommit)
	}
	if vi.SemanticVersion != "1.0.0" {
		t.Errorf("expected version 1.0.0, got %s", vi.SemanticVersion)
	}
	if vi.SchemaVersion != "8" {
		t.Errorf("expected schema_version 8, got %s", vi.SchemaVersion)
	}
	if vi.ContentGeneration != 5 {
		t.Errorf("expected content_generation 5, got %d", vi.ContentGeneration)
	}
}

func TestIntegration_LiveEndpoint(t *testing.T) {
	hc := NewHealthChecker()
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()
	LiveHandler(hc)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestIntegration_ReadyEndpoint(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return nil }},
	)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	ReadyHandler(hc)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestIntegration_ProfilingEndpoint(t *testing.T) {
	p := NewProfiler()
	obs := &Observability{
		Logger:   DefaultLogger(),
		Metrics:  NewMetrics(),
		Profiler: p,
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate a slow handler
		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := obs.Wrap(http.HandlerFunc(inner))
	req := httptest.NewRequest(http.MethodGet, "/api/slow", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// Check profiling data was recorded
	profileReq := httptest.NewRequest(http.MethodGet, "/debug/profile", nil)
	pw := httptest.NewRecorder()
	ProfileHandler(p)(pw, profileReq)

	if pw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", pw.Code)
	}

	var snap ProfilerSnapshot
	if err := json.Unmarshal(pw.Body.Bytes(), &snap); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if len(snap.SlowHandlers) != 0 {
		t.Errorf("expected no slow handlers (10ms < 500ms threshold), got %d", len(snap.SlowHandlers))
	}
}

func TestIntegration_WrapWithSecuredEndpoints(t *testing.T) {
	GitCommit = "inttest"
	Version = "1.0.0"
	SchemaVersion = "8"
	SetRuntimeSchemaVersion("")
	defer func() {
		GitCommit = "unknown"
		Version = "dev"
		SchemaVersion = "0"
		SetRuntimeSchemaVersion("")
	}()

	const testToken = "super-secret-metrics-token"

	obs := NewObservability()
	hc := NewHealthChecker(
		HealthCheck{Name: "config", Fn: ConfigHealthCheck(true)},
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", InternalTokenMiddleware(testToken, MetricsHandler(obs.Metrics)))
	mux.HandleFunc("/version", VersionHandler(&StaticBuildInfo{ContentGen: 1}))
	mux.HandleFunc("/health", HealthHandler(hc))
	mux.HandleFunc("/live", LiveHandler(hc))
	mux.HandleFunc("/ready", ReadyHandler(hc))
	mux.HandleFunc("/debug/profile", InternalTokenMiddleware(testToken, ProfileHandler(obs.Profiler)))
	mux.HandleFunc("/debug/profile/recommendations", InternalTokenMiddleware(testToken, RecommendationsHandler(obs.Profiler)))

	handler := obs.Wrap(mux)

	for _, route := range []string{"/metrics", "/version", "/health", "/live", "/ready", "/debug/profile", "/debug/profile/recommendations"} {
		req := httptest.NewRequest(http.MethodGet, route, nil)
		req.Header.Set(HeaderInternalToken, testToken)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("route %s: expected 200, got %d", route, w.Code)
		}
		if w.Header().Get(HeaderRequestID) == "" {
			t.Errorf("route %s: expected X-Request-ID header", route)
		}
	}

	snap := obs.Metrics.Snapshot()
	if len(snap.RequestCount) != 7 {
		t.Errorf("expected 7 request routes, got %d", len(snap.RequestCount))
	}
}

func TestIntegration_MetricsEndpointWithoutToken(t *testing.T) {
	const testToken = "secret-token"
	obs := NewObservability()

	handler := InternalTokenMiddleware(testToken, MetricsHandler(obs.Metrics))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 without token, got %d", w.Code)
	}
}

func TestIntegration_MetricsEndpointWithWrongToken(t *testing.T) {
	const testToken = "secret-token"
	obs := NewObservability()

	handler := InternalTokenMiddleware(testToken, MetricsHandler(obs.Metrics))
	req := httptest.NewRequest(http.MethodGet, "/metrics?token=wrong", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 with wrong token, got %d", w.Code)
	}
}

func TestIntegration_MetricsEndpointWithCorrectToken(t *testing.T) {
	const testToken = "secret-token"
	obs := NewObservability()
	obs.Metrics.RecordRequest("GET", "/api/test", 200, time.Millisecond)

	handler := InternalTokenMiddleware(testToken, MetricsHandler(obs.Metrics))
	req := httptest.NewRequest(http.MethodGet, "/metrics?token=secret-token", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200 with correct token, got %d", w.Code)
	}
}

func TestIntegration_MetricsEndpointNoTokenConfigured(t *testing.T) {
	obs := NewObservability()

	handler := InternalTokenMiddleware("", MetricsHandler(obs.Metrics))
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503 with no token configured, got %d", w.Code)
	}
}
