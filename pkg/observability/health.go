package observability

import (
	"context"
	"net/http"
	"sync"
	"time"
)

const DefaultHealthCacheTTL = 5 * time.Second

type HealthCheckFunc func(ctx context.Context) error

type HealthCheck struct {
	Name string
	Fn   HealthCheckFunc
}

type HealthChecker struct {
	checks   []HealthCheck
	cacheTTL time.Duration
	mu       sync.RWMutex
	cached   *HealthStatus
	cachedAt time.Time
}

func NewHealthChecker(checks ...HealthCheck) *HealthChecker {
	return &HealthChecker{
		checks:   checks,
		cacheTTL: DefaultHealthCacheTTL,
	}
}

func (hc *HealthChecker) AddCheck(name string, fn HealthCheckFunc) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.checks = append(hc.checks, HealthCheck{Name: name, Fn: fn})
}

func (hc *HealthChecker) SetCacheTTL(ttl time.Duration) {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.cacheTTL = ttl
}

func (hc *HealthChecker) ResetCache() {
	hc.mu.Lock()
	defer hc.mu.Unlock()
	hc.cached = nil
}

func (hc *HealthChecker) Check(ctx context.Context) HealthStatus {
	hc.mu.RLock()
	if hc.cached != nil && time.Since(hc.cachedAt) < hc.cacheTTL {
		cached := *hc.cached
		hc.mu.RUnlock()
		return cached
	}
	hc.mu.RUnlock()

	hc.mu.Lock()
	defer hc.mu.Unlock()
	if hc.cached != nil && time.Since(hc.cachedAt) < hc.cacheTTL {
		return *hc.cached
	}

	results := make(map[string]CheckResult)
	allPass := true
	for _, c := range hc.checks {
		start := time.Now()
		err := c.Fn(ctx)
		latency := time.Since(start).String()
		if err != nil {
			results[c.Name] = CheckResult{Status: "fail", Error: err.Error(), Latency: latency}
			allPass = false
		} else {
			results[c.Name] = CheckResult{Status: "pass", Latency: latency}
		}
	}
	status := "ok"
	if !allPass {
		status = "degraded"
	}
	hs := HealthStatus{
		Status:    status,
		Timestamp: nowRFC3339(),
		Checks:    results,
	}
	hc.cached = &hs
	hc.cachedAt = time.Now()
	return hs
}

func (hc *HealthChecker) CheckLive(ctx context.Context) HealthStatus {
	return HealthStatus{
		Status:    "alive",
		Timestamp: nowRFC3339(),
	}
}

func (hc *HealthChecker) CheckReady(ctx context.Context) (HealthStatus, error) {
	hs := hc.Check(ctx)
	if hs.Status == "degraded" {
		return hs, errHealthDegraded
	}
	return hs, nil
}

var errHealthDegraded = newHealthError("one or more health checks failed")

type healthError struct{ msg string }

func (e *healthError) Error() string { return e.msg }

func newHealthError(msg string) error {
	return &healthError{msg: msg}
}

func HealthHandler(hc *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		status := hc.Check(r.Context())
		code := http.StatusOK
		if status.Status == "degraded" {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, status)
	}
}

func LiveHandler(hc *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    "alive",
			"timestamp": nowRFC3339(),
		})
	}
}

func ReadyHandler(hc *HealthChecker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		status, err := hc.CheckReady(r.Context())
		code := http.StatusOK
		if err != nil {
			code = http.StatusServiceUnavailable
		}
		writeJSON(w, code, status)
	}
}

func DBHealthCheck(client SupabaseClient, tableName string) HealthCheckFunc {
	return func(ctx context.Context) error {
		_, err := client.Get(ctx, tableName, "limit=1")
		if err != nil {
			return err
		}
		return nil
	}
}

func ConfigHealthCheck(isConfigured bool) HealthCheckFunc {
	return func(ctx context.Context) error {
		if !isConfigured {
			return errConfigNotLoaded
		}
		return nil
	}
}

var (
	errConfigNotLoaded = newHealthError("configuration not loaded")
)
