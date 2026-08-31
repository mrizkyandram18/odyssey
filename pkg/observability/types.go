package observability

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"
)

type contextKey int

const (
	requestIDKey contextKey = iota
	profileKeyT
)

const (
	HeaderRequestID      = "X-Request-ID"
	HeaderInternalToken  = "X-Internal-Token"
	DefaultSlowThreshold = 500 * time.Millisecond
	DefaultDBLatencyWarn = 100 * time.Millisecond
)

type responseWriterWrapper struct {
	http.ResponseWriter
	status        int
	headerWritten bool
}

var responseWriterPool = sync.Pool{
	New: func() any {
		return &responseWriterWrapper{status: http.StatusOK}
	},
}

func wrapResponseWriter(w http.ResponseWriter) *responseWriterWrapper {
	rw := responseWriterPool.Get().(*responseWriterWrapper)
	rw.ResponseWriter = w
	rw.status = http.StatusOK
	rw.headerWritten = false
	return rw
}

func (w *responseWriterWrapper) reset() {
	w.ResponseWriter = nil
	w.status = http.StatusOK
	w.headerWritten = false
}

func (w *responseWriterWrapper) WriteHeader(code int) {
	if w.headerWritten {
		return
	}
	w.headerWritten = true
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *responseWriterWrapper) Write(b []byte) (int, error) {
	if !w.headerWritten {
		w.status = http.StatusOK
		w.headerWritten = true
	}
	return w.ResponseWriter.Write(b)
}

func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey, id)
}

func RequestIDFromContext(ctx context.Context) string {
	id, _ := ctx.Value(requestIDKey).(string)
	return id
}

type LogFields struct {
	RequestID string
	UserID    string
	FamilyID  string
	AdminUID  string
	Endpoint  string
	Method    string
	Duration  time.Duration
	Status    int
	RemoteIP  string
	Error     string
}

type CheckResult struct {
	Status  string `json:"status"`
	Error   string `json:"error,omitempty"`
	Latency string `json:"latency,omitempty"`
}

type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Version   string                 `json:"version,omitempty"`
	Checks    map[string]CheckResult `json:"checks,omitempty"`
}

func remoteIP(r *http.Request) string {
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		if idx := commaIndex(ip); idx > 0 {
			return ip[:idx]
		}
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}
	return r.RemoteAddr
}

func commaIndex(s string) int {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			return i
		}
	}
	return -1
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func writeJSONError(w http.ResponseWriter, msg string, status int) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func nowRFC3339() string {
	return time.Now().UTC().Format(time.RFC3339)
}
