package observability

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

type DBQueryRecord struct {
	Table     string
	Operation string
	Params    string
	Duration  time.Duration
	Timestamp time.Time
}

type RequestProfile struct {
	mu         sync.Mutex
	dbQueries  []DBQueryRecord
	startTime  time.Time
	endpoint   string
	method     string
	handlerDur time.Duration
}

func newRequestProfile(endpoint, method string) *RequestProfile {
	return &RequestProfile{
		startTime: time.Now(),
		endpoint:  endpoint,
		method:    method,
	}
}

func (rp *RequestProfile) RecordDBQuery(table, operation, params string, duration time.Duration) {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	rp.dbQueries = append(rp.dbQueries, DBQueryRecord{
		Table:     table,
		Operation: operation,
		Params:    params,
		Duration:  duration,
		Timestamp: time.Now().UTC(),
	})
}

func (rp *RequestProfile) GetDBQueryCount() int {
	rp.mu.Lock()
	defer rp.mu.Unlock()
	return len(rp.dbQueries)
}

func (rp *RequestProfile) Finalize(handlerDur time.Duration) {
	rp.handlerDur = handlerDur
}

func profileFromContext(ctx context.Context) *RequestProfile {
	rp, _ := ctx.Value(profileKeyT).(*RequestProfile)
	return rp
}

func withProfile(ctx context.Context, rp *RequestProfile) context.Context {
	return context.WithValue(ctx, profileKeyT, rp)
}

const DefaultProfilingRetention = 5 * time.Minute

type slowEntry struct {
	Endpoint  string
	Method    string
	Duration  string
	Timestamp time.Time
}

type recEntry struct {
	message   string
	severity  string
	timestamp time.Time
}

type Profiler struct {
	mu              sync.RWMutex
	slowPaths       []slowEntry
	recommendations []recEntry
	maxSlow         int
}

type SlowHandlerRecord struct {
	Endpoint  string `json:"endpoint"`
	Method    string `json:"method"`
	Duration  string `json:"duration"`
	Timestamp string `json:"timestamp"`
}

type Recommendation struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

func NewProfiler() *Profiler {
	return &Profiler{maxSlow: 100}
}

func (p *Profiler) NewRequestProfile(endpoint, method string) *RequestProfile {
	return newRequestProfile(endpoint, method)
}

func (p *Profiler) RecordRequest(rp *RequestProfile) {
	if rp == nil {
		return
	}
	dur := rp.handlerDur
	if dur > DefaultSlowThreshold {
		p.mu.Lock()
		p.pruneLocked(time.Now().Add(-DefaultProfilingRetention))
		if len(p.slowPaths) >= p.maxSlow {
			p.slowPaths = p.slowPaths[1:]
		}
		now := time.Now().UTC()
		p.slowPaths = append(p.slowPaths, slowEntry{
			Endpoint:  rp.endpoint,
			Method:    rp.method,
			Duration:  dur.String(),
			Timestamp: now,
		})
		p.mu.Unlock()
	}

	recs := p.analyzeProfile(rp)
	if len(recs) > 0 {
		p.mu.Lock()
		p.pruneRecommendationsLocked(time.Now().Add(-DefaultProfilingRetention))
		now := time.Now().UTC()
		for _, rec := range recs {
			p.recommendations = append(p.recommendations, recEntry{
				message:   rec,
				severity:  "warning",
				timestamp: now,
			})
		}
		if len(p.recommendations) > 100 {
			p.recommendations = p.recommendations[len(p.recommendations)-100:]
		}
		p.mu.Unlock()
	}
}

func (p *Profiler) pruneLocked(cutoff time.Time) {
	filtered := p.slowPaths[:0]
	for _, s := range p.slowPaths {
		if s.Timestamp.After(cutoff) {
			filtered = append(filtered, s)
		}
	}
	p.slowPaths = filtered
}

func (p *Profiler) pruneRecommendationsLocked(cutoff time.Time) {
	filtered := p.recommendations[:0]
	for _, r := range p.recommendations {
		if r.timestamp.After(cutoff) {
			filtered = append(filtered, r)
		}
	}
	p.recommendations = filtered
}

func (p *Profiler) Prune() {
	p.mu.Lock()
	defer p.mu.Unlock()
	cutoff := time.Now().Add(-DefaultProfilingRetention)
	p.pruneLocked(cutoff)
	p.pruneRecommendationsLocked(cutoff)
}

func (p *Profiler) analyzeProfile(rp *RequestProfile) []string {
	var recs []string
	rp.mu.Lock()
	queries := make([]DBQueryRecord, len(rp.dbQueries))
	copy(queries, rp.dbQueries)
	rp.mu.Unlock()

	if len(queries) == 0 {
		return recs
	}

	byTable := make(map[string][]DBQueryRecord)
	for _, q := range queries {
		byTable[q.Table] = append(byTable[q.Table], q)
	}

	for table, qs := range byTable {
		if len(qs) > 5 {
			recs = append(recs, fmt.Sprintf(
				"N+1 detected: %d queries on %s in a single request. Consider batch loading.",
				len(qs), table,
			))
		}
	}

	exactDups := make(map[string]int)
	for _, q := range queries {
		key := fmt.Sprintf("%s:%s:%s", q.Table, q.Operation, q.Params)
		exactDups[key]++
	}
	for key, count := range exactDups {
		if count > 1 {
			recs = append(recs, fmt.Sprintf(
				"Duplicated query detected: %s executed %d times. Consider caching or batching.",
				key, count,
			))
		}
	}

	var maxDB time.Duration
	var maxQ DBQueryRecord
	for _, q := range queries {
		if q.Duration > maxDB {
			maxDB = q.Duration
			maxQ = q
		}
	}
	if maxDB > DefaultDBLatencyWarn {
		recs = append(recs, fmt.Sprintf(
			"Slow DB query: %s on %s took %s",
			maxQ.Operation, maxQ.Table, maxDB,
		))
	}

	return recs
}

type ProfilerSnapshot struct {
	SlowHandlers    []SlowHandlerRecord `json:"slow_handlers"`
	Recommendations []string            `json:"recommendations"`
}

func (p *Profiler) Snapshot() ProfilerSnapshot {
	p.mu.RLock()
	defer p.mu.RUnlock()

	slow := make([]SlowHandlerRecord, len(p.slowPaths))
	for i, s := range p.slowPaths {
		slow[i] = SlowHandlerRecord{
			Endpoint:  s.Endpoint,
			Method:    s.Method,
			Duration:  s.Duration,
			Timestamp: s.Timestamp.Format(time.RFC3339),
		}
	}

	recs := make([]string, len(p.recommendations))
	for i, r := range p.recommendations {
		recs[i] = r.message
	}

	return ProfilerSnapshot{
		SlowHandlers:    slow,
		Recommendations: recs,
	}
}

func ProfilingMiddleware(p *Profiler, slowThreshold time.Duration, metrics *Metrics, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw := &responseWriterWrapper{ResponseWriter: w, status: http.StatusOK}
		rp := p.NewRequestProfile(r.URL.Path, r.Method)
		ctx := withProfile(r.Context(), rp)

		start := time.Now()
		next.ServeHTTP(rw, r.WithContext(ctx))
		duration := time.Since(start)

		rp.Finalize(duration)
		if metrics != nil {
			metrics.RecordRequest(r.Method, r.URL.Path, rw.status, duration)
		}
		p.RecordRequest(rp)
	}
}

func ProfileHandler(p *Profiler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(w, http.StatusOK, p.Snapshot())
	}
}

func RecommendationsHandler(p *Profiler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSONError(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		snap := p.Snapshot()
		recs := make([]Recommendation, 0, len(snap.Recommendations))
		for _, rec := range snap.Recommendations {
			recs = append(recs, Recommendation{
				Type:     "performance",
				Message:  rec,
				Severity: "warning",
			})
		}
		sort.Slice(recs, func(i, j int) bool {
			return recs[i].Message < recs[j].Message
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"recommendations": recs,
			"timestamp":       nowRFC3339(),
		})
	}
}

func (p *Profiler) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.slowPaths = nil
	p.recommendations = nil
}

func (p *Profiler) AddSlowHandler(method, endpoint string, dur time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	now := time.Now().UTC()
	p.slowPaths = append(p.slowPaths, slowEntry{
		Endpoint:  endpoint,
		Method:    method,
		Duration:  dur.String(),
		Timestamp: now,
	})
}

func (p *Profiler) AddRecommendation(rec string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.recommendations = append(p.recommendations, recEntry{
		message:   rec,
		severity:  "warning",
		timestamp: time.Now().UTC(),
	})
}
