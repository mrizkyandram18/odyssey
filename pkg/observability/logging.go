package observability

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var sensitiveKeys = map[string]bool{
	"password": true, "password_hash": true, "secret": true,
	"token": true, "session": true, "apikey": true, "authorization": true,
	"service_key": true, "credentials": true,
	"firebase_credentials":                true,
	"google_application_credentials_json": true,
}

const defaultLogBufferSize = 1024

type logEntry struct {
	level  string
	msg    string
	fields map[string]any
}

type Logger struct {
	ch       chan logEntry
	w        io.Writer
	fallback sync.Mutex
	wg       sync.WaitGroup
	done     chan struct{}
	once     sync.Once
}

func NewLogger(w io.Writer) *Logger {
	return NewLoggerWithBuffer(w, defaultLogBufferSize)
}

func NewLoggerWithBuffer(w io.Writer, bufferSize int) *Logger {
	if w == nil {
		w = os.Stdout
	}
	if bufferSize <= 0 {
		bufferSize = defaultLogBufferSize
	}
	l := &Logger{
		w:    w,
		ch:   make(chan logEntry, bufferSize),
		done: make(chan struct{}),
	}
	go l.process()
	return l
}

func DefaultLogger() *Logger {
	return NewLogger(os.Stdout)
}

func (l *Logger) process() {
	for entry := range l.ch {
		l.write(entry)
		l.wg.Done()
	}
	close(l.done)
}

func (l *Logger) write(entry logEntry) {
	data, err := l.marshal(entry.level, entry.msg, entry.fields)
	if err != nil {
		return
	}
	_, _ = l.w.Write(data)
	_, _ = l.w.Write([]byte("\n"))
}

func (l *Logger) marshal(level, msg string, fields map[string]any) ([]byte, error) {
	entry := map[string]any{
		"ts":  nowRFC3339(),
		"lvl": level,
		"msg": msg,
	}
	for k, v := range fields {
		entry[k] = v
	}
	return json.Marshal(entry)
}

func (l *Logger) log(level, msg string, fields map[string]any) {
	sanitized := sanitizeFields(fields)
	select {
	case l.ch <- logEntry{level: level, msg: msg, fields: sanitized}:
		l.wg.Add(1)
	default:
		l.fallbackLog(level, msg, sanitized)
	}
}

func (l *Logger) fallbackLog(level, msg string, fields map[string]any) {
	data, err := l.marshal(level, msg, fields)
	if err != nil {
		return
	}
	l.fallback.Lock()
	_, _ = l.w.Write(data)
	_, _ = l.w.Write([]byte("\n"))
	l.fallback.Unlock()
}

func (l *Logger) Flush() {
	if l == nil {
		return
	}
	l.wg.Wait()
}

func (l *Logger) Close() {
	if l == nil {
		return
	}
	l.once.Do(func() {
		close(l.ch)
		<-l.done
	})
}

func (l *Logger) Info(msg string, fields map[string]any) {
	l.log("INFO", msg, fields)
}

func (l *Logger) Error(msg string, fields map[string]any) {
	l.log("ERROR", msg, fields)
}

func (l *Logger) Warn(msg string, fields map[string]any) {
	l.log("WARN", msg, fields)
}

func (l *Logger) LogRequest(fields LogFields) {
	f := map[string]any{
		"request_id":  fields.RequestID,
		"user_id":     fields.UserID,
		"crew_id":     fields.CrewID,
		"admin_uid":   fields.AdminUID,
		"endpoint":    fields.Endpoint,
		"method":      fields.Method,
		"duration_ms": float64(fields.Duration.Nanoseconds()) / 1e6,
		"status":      fields.Status,
		"remote_ip":   fields.RemoteIP,
	}
	if fields.Error != "" {
		f["error"] = fields.Error
	}
	l.log("INFO", "request", f)
}

// ServiceCallFields describes a single business-logic call for structured
// service-level logging. Outcome is one of "ok", "conflict", "duplicate",
// "invalid", or "error".
type ServiceCallFields struct {
	RequestID         string
	Operation         string
	EntityID          string
	Duration          time.Duration
	Outcome           string
	Retried           bool
	Conflict          bool
	IdempotencySkip   bool
	ValidationFailure bool
	Error             string
}

// LogServiceCall emits a structured service-call log line correlatable to the
// originating request via RequestID.
func (l *Logger) LogServiceCall(fields ServiceCallFields) {
	if l == nil {
		return
	}
	if fields.Outcome == "" {
		fields.Outcome = "ok"
	}
	f := map[string]any{
		"request_id":         fields.RequestID,
		"op":                 fields.Operation,
		"duration_ms":        float64(fields.Duration.Nanoseconds()) / 1e6,
		"outcome":            fields.Outcome,
		"retried":            fields.Retried,
		"conflict":           fields.Conflict,
		"idempotency_skip":   fields.IdempotencySkip,
		"validation_failure": fields.ValidationFailure,
	}
	if fields.EntityID != "" {
		f["entity_id"] = fields.EntityID
	}
	if fields.Error != "" {
		f["error"] = fields.Error
	}
	l.log("INFO", "service_call", f)
}

func sanitizeFields(fields map[string]any) map[string]any {
	out := make(map[string]any, len(fields))
	for k, v := range fields {
		if sensitiveKeys[strings.ToLower(k)] {
			out[k] = "[REDACTED]"
		} else {
			out[k] = v
		}
	}
	return out
}

func extractIdentityFromToken(r *http.Request) (uid, crewID, adminUID string) {
	token := r.Header.Get("Authorization")
	if len(token) >= 7 && strings.EqualFold(token[:7], "Bearer ") {
		token = strings.TrimSpace(token[7:])
	} else {
		token = strings.TrimSpace(r.Header.Get("X-User-Session"))
	}
	if token == "" {
		if cookie, err := r.Cookie("odyssey_session"); err == nil {
			token = strings.TrimSpace(cookie.Value)
		}
	}
	if token == "" {
		return
	}
	parts := strings.SplitN(token, ".", 2)
	if len(parts) != 2 || parts[0] == "" {
		return
	}
	raw, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return
	}
	var claims struct {
		UID    string `json:"uid"`
		CrewID string `json:"crew_id"`
		Role   string `json:"role"`
	}
	if err := json.Unmarshal(raw, &claims); err != nil {
		return
	}
	uid = claims.UID
	crewID = claims.CrewID
	if claims.Role == "ADMIN" {
		adminUID = claims.UID
	}
	return
}
