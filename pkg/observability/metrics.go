package observability

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	mu                  sync.RWMutex
	requestCount        map[string]int64
	requestLatency      map[string][]latencyEntry
	loginSuccess        int64
	loginFailure        int64
	taskSubmitted       int64
	rewardRedeemed      int64
	adminOps            int64
	xpAwarded           int64
	duplicatesPrevented int64
	validationFailures  int64
	dbLatencyTotal      time.Duration
	dbCallCount         int64
	dbSlowCount         int64
	bootTime            time.Time
}

type latencyEntry struct {
	Status    int
	Duration  time.Duration
	Timestamp time.Time
}

func NewMetrics() *Metrics {
	return &Metrics{
		requestCount:   make(map[string]int64),
		requestLatency: make(map[string][]latencyEntry),
		bootTime:       time.Now().UTC(),
	}
}

func (m *Metrics) RecordRequest(method, endpoint string, status int, duration time.Duration) {
	key := method + " " + endpoint
	m.mu.Lock()
	m.requestCount[key]++
	if len(m.requestLatency[key]) >= 1000 {
		m.requestLatency[key] = m.requestLatency[key][:1000]
	}
	m.requestLatency[key] = append(m.requestLatency[key], latencyEntry{
		Status:    status,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
	})
	m.mu.Unlock()
}

func (m *Metrics) RecordLogin(success bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if success {
		m.loginSuccess++
	} else {
		m.loginFailure++
	}
}

func (m *Metrics) RecordBusinessEvent(event string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch event {
	case "task_submitted":
		m.taskSubmitted++
	case "reward_redeemed":
		m.rewardRedeemed++
	}
}

func (m *Metrics) RecordAdminOp() {
	m.mu.Lock()
	m.adminOps++
	m.mu.Unlock()
}

func (m *Metrics) RecordXP(amount int64) {
	if amount <= 0 {
		return
	}
	m.mu.Lock()
	m.xpAwarded += amount
	m.mu.Unlock()
}

func (m *Metrics) RecordDuplicatePrevented(n int) {
	if n <= 0 {
		return
	}
	m.mu.Lock()
	m.duplicatesPrevented += int64(n)
	m.mu.Unlock()
}

func (m *Metrics) RecordValidationFailure() {
	m.mu.Lock()
	m.validationFailures++
	m.mu.Unlock()
}

func (m *Metrics) RecordDBLatency(d time.Duration) {
	m.mu.Lock()
	m.dbLatencyTotal += d
	m.dbCallCount++
	if d > DefaultDBLatencyWarn {
		m.dbSlowCount++
	}
	m.mu.Unlock()
}

type MetricsSnapshot struct {
	BootTime            string             `json:"boot_time"`
	UptimeSeconds       int64              `json:"uptime_seconds"`
	RequestCount        map[string]int64   `json:"request_count"`
	RequestLatencyMs    map[string]float64 `json:"request_latency_ms"`
	LoginSuccess        int64              `json:"login_success"`
	LoginFailure        int64              `json:"login_failure"`
	TaskSubmitted       int64              `json:"task_submitted"`
	RewardRedeemed      int64              `json:"reward_redeemed"`
	AdminOps            int64              `json:"admin_operations"`
	XPAwarded           int64              `json:"xp_awarded"`
	DuplicatesPrevented int64              `json:"duplicates_prevented"`
	ValidationFailures  int64              `json:"validation_failures"`
	DBCallCount         int64              `json:"db_calls"`
	DBAvgLatencyMs      float64            `json:"db_avg_latency_ms"`
	DBSlowCount         int64              `json:"db_slow_queries"`
	Timestamp           string             `json:"timestamp"`
}

func (m *Metrics) Snapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	reqCounts := make(map[string]int64, len(m.requestCount))
	for k, v := range m.requestCount {
		reqCounts[k] = v
	}

	avgLatency := make(map[string]float64, len(m.requestLatency))
	for k, entries := range m.requestLatency {
		if len(entries) == 0 {
			continue
		}
		var total time.Duration
		for _, e := range entries {
			total += e.Duration
		}
		avgLatency[k] = float64(total.Nanoseconds()) / float64(len(entries)) / 1e6
	}

	var dbAvg time.Duration
	if m.dbCallCount > 0 {
		dbAvg = m.dbLatencyTotal / time.Duration(m.dbCallCount)
	}

	return MetricsSnapshot{
		BootTime:            m.bootTime.Format(time.RFC3339),
		UptimeSeconds:       int64(time.Since(m.bootTime).Seconds()),
		RequestCount:        reqCounts,
		RequestLatencyMs:    avgLatency,
		LoginSuccess:        m.loginSuccess,
		LoginFailure:        m.loginFailure,
		TaskSubmitted:       m.taskSubmitted,
		RewardRedeemed:      m.rewardRedeemed,
		AdminOps:            m.adminOps,
		XPAwarded:           m.xpAwarded,
		DuplicatesPrevented: m.duplicatesPrevented,
		ValidationFailures:  m.validationFailures,
		DBCallCount:         m.dbCallCount,
		DBAvgLatencyMs:      float64(dbAvg.Nanoseconds()) / 1e6,
		DBSlowCount:         m.dbSlowCount,
		Timestamp:           nowRFC3339(),
	}
}

func MetricsHandler(metrics *Metrics) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, metrics.Snapshot())
	}
}

func (m *Metrics) BootTime() time.Time {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.bootTime
}

func InternalTokenMiddleware(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if token == "" {
			writeJSONError(w, "metrics endpoint disabled", http.StatusServiceUnavailable)
			return
		}
		provided := r.Header.Get(HeaderInternalToken)
		if provided == "" {
			provided = r.URL.Query().Get("token")
		}
		if subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeJSONError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func (m *Metrics) SnapshotJSON() ([]byte, error) {
	return json.MarshalIndent(m.Snapshot(), "", "  ")
}
