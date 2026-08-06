package shared

import (
	"fmt"
	"net/http"
	"time"
)

type SecurityEventType string

const (
	SecEventFailedLogin       SecurityEventType = "failed_login"
	SecEventForbiddenAccess   SecurityEventType = "forbidden_access"
	SecEventInvalidToken      SecurityEventType = "invalid_token"
	SecEventRateLimit         SecurityEventType = "rate_limit"
	SecEventValidationFailure SecurityEventType = "validation_failure"
	SecEventCSRFFailure       SecurityEventType = "csrf_failure"
	SecEventIDORAttempt       SecurityEventType = "idor_attempt"
)

type SecurityEvent struct {
	Type       SecurityEventType
	RemoteAddr string
	UserAgent  string
	Path       string
	Method     string
	Detail     string
	Timestamp  time.Time
}

var securityEvents chan SecurityEvent

func InitSecurityLogger(bufferSize int) {
	if bufferSize <= 0 {
		bufferSize = 100
	}
	securityEvents = make(chan SecurityEvent, bufferSize)
	go processSecurityEvents()
}

func LogSecurityEvent(eventType SecurityEventType, r *http.Request, detail string) {
	if securityEvents == nil {
		return
	}
	ua := r.UserAgent()
	if ua == "" {
		ua = "unknown"
	}
	securityEvents <- SecurityEvent{
		Type:       eventType,
		RemoteAddr: r.RemoteAddr,
		UserAgent:  ua,
		Path:       r.URL.Path,
		Method:     r.Method,
		Detail:     detail,
		Timestamp:  time.Now().UTC(),
	}
}

func processSecurityEvents() {
	for event := range securityEvents {
		fmt.Printf("[SECURITY] %s | %s | %s %s | %s | %s\n",
			event.Timestamp.Format(time.RFC3339),
			event.Type,
			event.Method,
			event.Path,
			event.RemoteAddr,
			event.UserAgent,
		)
		if event.Detail != "" {
			fmt.Printf("  detail: %s\n", event.Detail)
		}
	}
}
