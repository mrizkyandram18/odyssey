package observability

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

type Observability struct {
	Logger   *Logger
	Metrics  *Metrics
	Profiler *Profiler
	Health   *HealthChecker
}

func NewObservability() *Observability {
	return &Observability{
		Logger:   DefaultLogger(),
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}
}

func (o *Observability) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(HeaderRequestID)
		if requestID == "" {
			requestID = generateRequestID()
		}
		ctx := WithRequestID(r.Context(), requestID)
		w.Header().Set(HeaderRequestID, requestID)

		rw := wrapResponseWriter(w)

		var rp *RequestProfile
		if o.Profiler != nil {
			rp = o.Profiler.NewRequestProfile(r.URL.Path, r.Method)
			ctx = withProfile(ctx, rp)
		}

		start := time.Now()

		defer func() {
			if rec := recover(); rec != nil {
				stack := debug.Stack()
				if o.Logger != nil {
					o.Logger.Error("panic_recovered", map[string]any{
						"request_id": requestID,
						"endpoint":   r.URL.Path,
						"method":     r.Method,
						"error":      fmt.Sprintf("%v", rec),
						"stack":      string(stack),
					})
				}
				if !rw.headerWritten {
					writeJSONError(rw, "internal server error", http.StatusInternalServerError)
				}
			}
			rw.reset()
			responseWriterPool.Put(rw)
		}()

		next.ServeHTTP(rw, r.WithContext(ctx))
		duration := time.Since(start)

		if rp != nil {
			rp.Finalize(duration)
		}
		if o.Metrics != nil {
			o.Metrics.RecordRequest(r.Method, r.URL.Path, rw.status, duration)
			recordBusinessEvents(o.Metrics, r, rw.status)
		}
		if o.Profiler != nil {
			o.Profiler.RecordRequest(rp)
		}

		uid, familyID, adminUID := extractIdentityFromToken(r)
		if o.Logger != nil {
			o.Logger.LogRequest(LogFields{
				RequestID: requestID,
				UserID:    uid,
				FamilyID:  familyID,
				AdminUID:  adminUID,
				Endpoint:  r.URL.Path,
				Method:    r.Method,
				Duration:  duration,
				Status:    rw.status,
				RemoteIP:  remoteIP(r),
			})
		}
	})
}

func recordBusinessEvents(m *Metrics, r *http.Request, status int) {
	path := r.URL.Path
	method := r.Method
	is2xx := status >= 200 && status < 300

	switch {
	case path == "/api/login" || path == "/api/login/":
		if method == http.MethodPost {
			m.RecordLogin(is2xx)
		}
	case strings.HasPrefix(path, "/api/admin"):
		if is2xx {
			m.RecordAdminOp()
		}
	case strings.HasPrefix(path, "/api/tasks") && strings.HasSuffix(path, "/submit"):
		if method == http.MethodPost && is2xx {
			m.RecordBusinessEvent("task_submitted")
		}
	case path == "/api/shop/redeem" || path == "/api/shop/redeem/":
		if method == http.MethodPost && is2xx {
			m.RecordBusinessEvent("reward_redeemed")
		}
	}
}

func (o *Observability) RecordLogin(success bool) {
	if o.Metrics != nil {
		o.Metrics.RecordLogin(success)
	}
}

func (o *Observability) RecordBusinessEvent(event string) {
	if o.Metrics != nil {
		o.Metrics.RecordBusinessEvent(event)
	}
}

func (o *Observability) RecordAdminOp() {
	if o.Metrics != nil {
		o.Metrics.RecordAdminOp()
	}
}

func (o *Observability) RecordXP(amount int64) {
	if o.Metrics != nil {
		o.Metrics.RecordXP(amount)
	}
}

func (o *Observability) RecordDuplicatePrevented(n int) {
	if o.Metrics != nil {
		o.Metrics.RecordDuplicatePrevented(n)
	}
}

func (o *Observability) RecordValidationFailure() {
	if o.Metrics != nil {
		o.Metrics.RecordValidationFailure()
	}
}

func ContextWithRequestID(ctx context.Context, id string) context.Context {
	return WithRequestID(ctx, id)
}

func RequestID(r *http.Request) string {
	return RequestIDFromContext(r.Context())
}
