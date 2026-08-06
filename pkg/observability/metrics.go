package observability

import (
	"crypto/subtle"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type Metrics struct {
	mu                    sync.RWMutex
	requestCount          map[string]int64
	requestLatency        map[string][]latencyEntry
	loginSuccess          int64
	loginFailure          int64
	cacheHits             int64
	cacheMisses           int64
	questCompleted        int64
	chestOpened           int64
	creativeSubmitted     int64
	adminOps              int64
	xpAwarded             int64
	realmCompleted        int64
	chestsCreated         int64
	rewardsGenerated      int64
	duplicatesPrevented   int64
	lockConflicts         int64
	replayIgnored         int64
	validationFailures    int64
	eventsPublished       int64
	eventsHandled         int64
	eventsHandlerErrors   int64
	eventHandlerLatency   time.Duration
	eventHandlerCallCount int64
	eventTypeStats        map[string]EventTypeStat
	dbLatencyTotal        time.Duration
	dbCallCount           int64
	dbSlowCount           int64
	bootTime              time.Time
}

// EventTypeStat aggregates event-pipeline counters for a single event type.
type EventTypeStat struct {
	Published int64 `json:"published"`
	Handled   int64 `json:"handled"`
	Errors    int64 `json:"errors"`
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
		eventTypeStats: make(map[string]EventTypeStat),
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

func (m *Metrics) RecordCacheHit() {
	m.mu.Lock()
	m.cacheHits++
	m.mu.Unlock()
}

func (m *Metrics) RecordCacheMiss() {
	m.mu.Lock()
	m.cacheMisses++
	m.mu.Unlock()
}

func (m *Metrics) RecordBusinessEvent(event string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	switch event {
	case "quest_completed":
		m.questCompleted++
	case "chest_opened":
		m.chestOpened++
	case "creative_submitted":
		m.creativeSubmitted++
	}
}

func (m *Metrics) RecordAdminOp() {
	m.mu.Lock()
	m.adminOps++
	m.mu.Unlock()
}

// RecordXP adds the amount of XP awarded by the gameplay layer.
func (m *Metrics) RecordXP(amount int64) {
	if amount <= 0 {
		return
	}
	m.mu.Lock()
	m.xpAwarded += amount
	m.mu.Unlock()
}

// RecordRealmCompleted counts a realm reaching its completion threshold.
func (m *Metrics) RecordRealmCompleted() {
	m.mu.Lock()
	m.realmCompleted++
	m.mu.Unlock()
}

// RecordChestCreated counts a chest instance created for a player.
func (m *Metrics) RecordChestCreated() {
	m.mu.Lock()
	m.chestsCreated++
	m.mu.Unlock()
}

// RecordRewardsGenerated counts the total rewards produced by chest openings.
func (m *Metrics) RecordRewardsGenerated(n int) {
	if n <= 0 {
		return
	}
	m.mu.Lock()
	m.rewardsGenerated += int64(n)
	m.mu.Unlock()
}

// RecordDuplicatePrevented counts duplicate rewards merged instead of duplicated.
func (m *Metrics) RecordDuplicatePrevented(n int) {
	if n <= 0 {
		return
	}
	m.mu.Lock()
	m.duplicatesPrevented += int64(n)
	m.mu.Unlock()
}

// RecordLockConflict counts an optimistic-lock CAS miss (version mismatch).
func (m *Metrics) RecordLockConflict() {
	m.mu.Lock()
	m.lockConflicts++
	m.mu.Unlock()
}

// RecordReplayIgnored counts an idempotency guard short-circuiting a replay.
func (m *Metrics) RecordReplayIgnored() {
	m.mu.Lock()
	m.replayIgnored++
	m.mu.Unlock()
}

// RecordValidationFailure counts an invalid content validation result.
func (m *Metrics) RecordValidationFailure() {
	m.mu.Lock()
	m.validationFailures++
	m.mu.Unlock()
}

// RecordEventPublished counts an event dispatched into the pipeline.
func (m *Metrics) RecordEventPublished(eventType string) {
	if eventType == "" {
		return
	}
	m.mu.Lock()
	m.eventsPublished++
	st := m.eventTypeStats[eventType]
	st.Published++
	m.eventTypeStats[eventType] = st
	m.mu.Unlock()
}

// RecordEventHandler counts a handler invocation, its duration, and whether it
// returned an error. It satisfies the events.Recorder contract used by the
// Dispatcher.
func (m *Metrics) RecordEventHandler(eventType string, duration time.Duration, err error) {
	if eventType == "" {
		return
	}
	m.mu.Lock()
	m.eventsHandled++
	m.eventHandlerLatency += duration
	m.eventHandlerCallCount++
	st := m.eventTypeStats[eventType]
	st.Handled++
	if err != nil {
		m.eventsHandlerErrors++
		st.Errors++
	}
	m.eventTypeStats[eventType] = st
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

func (m *Metrics) MergeCacheStats(hits, misses, evictions int64) {
	m.mu.Lock()
	m.cacheHits += hits
	m.cacheMisses += misses
	m.mu.Unlock()
}

type MetricsSnapshot struct {
	BootTime            string                   `json:"boot_time"`
	UptimeSeconds       int64                    `json:"uptime_seconds"`
	RequestCount        map[string]int64         `json:"request_count"`
	RequestLatencyMs    map[string]float64       `json:"request_latency_ms"`
	LoginSuccess        int64                    `json:"login_success"`
	LoginFailure        int64                    `json:"login_failure"`
	CacheHits           int64                    `json:"cache_hits"`
	CacheMisses         int64                    `json:"cache_misses"`
	QuestCompleted      int64                    `json:"quest_completed"`
	ChestOpened         int64                    `json:"chest_opened"`
	CreativeSubmitted   int64                    `json:"creative_submitted"`
	AdminOps            int64                    `json:"admin_operations"`
	XPAwarded           int64                    `json:"xp_awarded"`
	RealmCompleted      int64                    `json:"realm_completed"`
	ChestsCreated       int64                    `json:"chests_created"`
	RewardsGenerated    int64                    `json:"rewards_generated"`
	DuplicatesPrevented int64                    `json:"duplicates_prevented"`
	LockConflicts       int64                    `json:"lock_conflicts"`
	ReplayIgnored       int64                    `json:"replay_ignored"`
	ValidationFailures  int64                    `json:"validation_failures"`
	EventsPublished     int64                    `json:"events_published"`
	EventsHandled       int64                    `json:"events_handled"`
	EventsHandlerErrors int64                    `json:"events_handler_errors"`
	EventsHandlerAvgMs  float64                  `json:"events_handler_avg_latency_ms"`
	EventTypes          map[string]EventTypeStat `json:"event_types"`
	DBCallCount         int64                    `json:"db_calls"`
	DBAvgLatencyMs      float64                  `json:"db_avg_latency_ms"`
	DBSlowCount         int64                    `json:"db_slow_queries"`
	Timestamp           string                   `json:"timestamp"`
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

	var eventAvg float64
	if m.eventHandlerCallCount > 0 {
		eventAvg = float64(m.eventHandlerLatency.Nanoseconds()) / float64(m.eventHandlerCallCount) / 1e6
	}

	eventTypes := make(map[string]EventTypeStat, len(m.eventTypeStats))
	for k, v := range m.eventTypeStats {
		eventTypes[k] = v
	}

	return MetricsSnapshot{
		BootTime:            m.bootTime.Format(time.RFC3339),
		UptimeSeconds:       int64(time.Since(m.bootTime).Seconds()),
		RequestCount:        reqCounts,
		RequestLatencyMs:    avgLatency,
		LoginSuccess:        m.loginSuccess,
		LoginFailure:        m.loginFailure,
		CacheHits:           m.cacheHits,
		CacheMisses:         m.cacheMisses,
		QuestCompleted:      m.questCompleted,
		ChestOpened:         m.chestOpened,
		CreativeSubmitted:   m.creativeSubmitted,
		AdminOps:            m.adminOps,
		XPAwarded:           m.xpAwarded,
		RealmCompleted:      m.realmCompleted,
		ChestsCreated:       m.chestsCreated,
		RewardsGenerated:    m.rewardsGenerated,
		DuplicatesPrevented: m.duplicatesPrevented,
		LockConflicts:       m.lockConflicts,
		ReplayIgnored:       m.replayIgnored,
		ValidationFailures:  m.validationFailures,
		EventsPublished:     m.eventsPublished,
		EventsHandled:       m.eventsHandled,
		EventsHandlerErrors: m.eventsHandlerErrors,
		EventsHandlerAvgMs:  eventAvg,
		EventTypes:          eventTypes,
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

// BootTime returns when the Metrics instance (and implicitly the server)
// started. Useful for uptime reporting on public status endpoints.
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
