package observability

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockSupabaseClient struct {
	err  error
	data []byte
}

func (m *mockSupabaseClient) Get(ctx context.Context, table string, params string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.data != nil {
		return m.data, nil
	}
	return []byte(`[]`), nil
}

func (m *mockSupabaseClient) Mutate(ctx context.Context, method, table string, payload any, params string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(`[]`), nil
}

func (m *mockSupabaseClient) MutateAtomic(ctx context.Context, method, table string, payload any, params string, prefer string) ([]byte, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []byte(`[]`), nil
}

func TestHealthChecker_AllPass(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "config", Fn: func(ctx context.Context) error { return nil }},
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return nil }},
	)

	status := hc.Check(context.Background())
	if status.Status != "ok" {
		t.Errorf("expected ok, got %s", status.Status)
	}
	if len(status.Checks) != 2 {
		t.Errorf("expected 2 checks, got %d", len(status.Checks))
	}
	if status.Checks["config"].Status != "pass" {
		t.Errorf("expected config to pass")
	}
	if status.Checks["database"].Status != "pass" {
		t.Errorf("expected database to pass")
	}
}

func TestHealthChecker_Degraded(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "config", Fn: func(ctx context.Context) error { return nil }},
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return errDBUnavailable }},
	)

	status := hc.Check(context.Background())
	if status.Status != "degraded" {
		t.Errorf("expected degraded, got %s", status.Status)
	}
	if status.Checks["database"].Status != "fail" {
		t.Errorf("expected database to fail")
	}
	if status.Checks["database"].Error == "" {
		t.Error("expected error message")
	}
}

func TestHealthChecker_CheckLive(t *testing.T) {
	hc := NewHealthChecker()
	status := hc.CheckLive(context.Background())
	if status.Status != "alive" {
		t.Errorf("expected alive, got %s", status.Status)
	}
}

func TestHealthChecker_CheckReady_Pass(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return nil }},
	)

	status, err := hc.CheckReady(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if status.Status != "ok" {
		t.Errorf("expected ok, got %s", status.Status)
	}
}

func TestHealthChecker_CheckReady_Fail(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return errDBUnavailable }},
	)

	status, err := hc.CheckReady(context.Background())
	if err == nil {
		t.Fatal("expected error for degraded readiness")
	}
	if status.Status != "degraded" {
		t.Errorf("expected degraded, got %s", status.Status)
	}
}

func TestDBHealthCheck_Pass(t *testing.T) {
	fn := DBHealthCheck(&mockSupabaseClient{}, "odyssey_user_profiles")
	if err := fn(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestDBHealthCheck_Fail(t *testing.T) {
	fn := DBHealthCheck(&mockSupabaseClient{err: errDBUnavailable}, "odyssey_user_profiles")
	if err := fn(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigHealthCheck_Pass(t *testing.T) {
	fn := ConfigHealthCheck(true)
	if err := fn(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestConfigHealthCheck_Fail(t *testing.T) {
	fn := ConfigHealthCheck(false)
	if err := fn(context.Background()); err == nil {
		t.Fatal("expected error for unconfigured")
	}
}

func TestCacheHealthCheck_Nil(t *testing.T) {
	fn := CacheHealthCheck(nil)
	if err := fn(context.Background()); err == nil {
		t.Fatal("expected error for nil cache")
	}
}

func TestContentHealthCheck_Pass(t *testing.T) {
	cs := &mockContentProvider{}
	fn := ContentHealthCheck(cs)
	if err := fn(context.Background()); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestContentHealthCheck_Fail(t *testing.T) {
	cs := &mockContentProvider{err: errContentUnavailable}
	fn := ContentHealthCheck(cs)
	if err := fn(context.Background()); err == nil {
		t.Fatal("expected error")
	}
}

func TestContentHealthCheck_Nil(t *testing.T) {
	fn := ContentHealthCheck(nil)
	if err := fn(context.Background()); err == nil {
		t.Fatal("expected error for nil content service")
	}
}

func TestHealthHandler(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return nil }},
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	HealthHandler(hc)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHealthHandler_Degraded(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return errDBUnavailable }},
	)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	HealthHandler(hc)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHealthHandler_MethodNotAllowed(t *testing.T) {
	hc := NewHealthChecker()
	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	w := httptest.NewRecorder()
	HealthHandler(hc)(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected 405, got %d", w.Code)
	}
}

func TestLiveHandler(t *testing.T) {
	hc := NewHealthChecker()
	req := httptest.NewRequest(http.MethodGet, "/live", nil)
	w := httptest.NewRecorder()
	LiveHandler(hc)(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReadyHandler_Ready(t *testing.T) {
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

func TestReadyHandler_NotReady(t *testing.T) {
	hc := NewHealthChecker(
		HealthCheck{Name: "database", Fn: func(ctx context.Context) error { return errDBUnavailable }},
	)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	ReadyHandler(hc)(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHealthChecker_AddCheck(t *testing.T) {
	hc := NewHealthChecker()
	hc.AddCheck("new_check", func(ctx context.Context) error { return nil })

	status := hc.Check(context.Background())
	if status.Checks["new_check"].Status != "pass" {
		t.Error("expected new_check to pass")
	}
}

func TestHealthChecker_Caching(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker(
		HealthCheck{Name: "db", Fn: func(ctx context.Context) error {
			callCount++
			return nil
		}},
	)
	hc.SetCacheTTL(5 * time.Second)

	for i := 0; i < 10; i++ {
		hc.Check(context.Background())
	}

	if callCount != 1 {
		t.Errorf("expected check function called once (cached), got %d calls", callCount)
	}
}

func TestHealthChecker_NoCache(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker(
		HealthCheck{Name: "db", Fn: func(ctx context.Context) error {
			callCount++
			return nil
		}},
	)
	hc.SetCacheTTL(0)

	for i := 0; i < 5; i++ {
		hc.Check(context.Background())
	}

	if callCount != 5 {
		t.Errorf("expected check function called 5 times (no cache), got %d calls", callCount)
	}
}

func TestHealthChecker_CacheExpiring(t *testing.T) {
	callCount := 0
	hc := NewHealthChecker(
		HealthCheck{Name: "db", Fn: func(ctx context.Context) error {
			callCount++
			return nil
		}},
	)
	hc.SetCacheTTL(50 * time.Millisecond)

	hc.Check(context.Background())
	if callCount != 1 {
		t.Fatalf("expected 1 call initially, got %d", callCount)
	}

	hc.Check(context.Background())
	if callCount != 1 {
		t.Fatalf("expected still 1 call (cached), got %d", callCount)
	}

	time.Sleep(60 * time.Millisecond)
	hc.Check(context.Background())
	if callCount != 2 {
		t.Fatalf("expected 2 calls after cache expiry, got %d", callCount)
	}
}

var errDBUnavailable = newHealthError("db unavailable")
