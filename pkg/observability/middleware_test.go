package observability

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

type mockBuildInfo struct {
	gen int64
}

func (m *mockBuildInfo) GetContentGeneration() int64 {
	return m.gen
}

func TestObservability_NewObservability(t *testing.T) {
	obs := NewObservability()
	if obs.Logger == nil {
		t.Error("expected non-nil Logger")
	}
	if obs.Metrics == nil {
		t.Error("expected non-nil Metrics")
	}
	if obs.Profiler == nil {
		t.Error("expected non-nil Profiler")
	}
}

func TestObservability_Wrap(t *testing.T) {
	var buf testBuffer
	obs := &Observability{
		Logger:   NewLogger(&buf),
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rid := RequestIDFromContext(r.Context())
		if rid == "" {
			t.Error("expected request ID in context")
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok":true}`))
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Header().Get(HeaderRequestID) == "" {
		t.Fatal("expected X-Request-ID header")
	}

	obs.Logger.Flush()
	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["request_id"] == "" {
		t.Error("expected request_id in log")
	}
	if entry["endpoint"] != "/api/test" {
		t.Errorf("expected endpoint /api/test, got %v", entry["endpoint"])
	}
	if entry["method"] != "GET" {
		t.Errorf("expected method GET, got %v", entry["method"])
	}
	if entry["status"].(float64) != 200 {
		t.Errorf("expected status 200, got %v", entry["status"])
	}

	snap := obs.Metrics.Snapshot()
	if snap.RequestCount["GET /api/test"] != 1 {
		t.Errorf("expected 1 request, got %d", snap.RequestCount["GET /api/test"])
	}
}

func TestObservability_Wrap_PreservesExistingRequestID(t *testing.T) {
	obs := NewObservability()
	existingID := "existing-request-id"

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(HeaderRequestID, existingID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Header().Get(HeaderRequestID) != existingID {
		t.Errorf("expected %s, got %s", existingID, w.Header().Get(HeaderRequestID))
	}
}

func TestObservability_Wrap_LogsUserIdentity(t *testing.T) {
	claims := `{"uid":"user-abc","crew_id":"crew-xyz","role":"SEEKER"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	token := payload + ".sig"

	var buf testBuffer
	obs := &Observability{
		Logger:   NewLogger(&buf),
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/quests", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	obs.Logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["user_id"] != "user-abc" {
		t.Errorf("expected user_id user-abc, got %v", entry["user_id"])
	}
	if entry["crew_id"] != "crew-xyz" {
		t.Errorf("expected crew_id crew-xyz, got %v", entry["crew_id"])
	}
}

func TestObservability_Wrap_LogsAdminIdentity(t *testing.T) {
	claims := `{"uid":"admin-1","crew_id":"crew-1","role":"ADMIN"}`
	payload := base64.RawURLEncoding.EncodeToString([]byte(claims))
	token := payload + ".sig"

	var buf testBuffer
	obs := &Observability{
		Logger:   NewLogger(&buf),
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/admin/status", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	obs.Logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["admin_uid"] != "admin-1" {
		t.Errorf("expected admin_uid admin-1, got %v", entry["admin_uid"])
	}
}

func TestObservability_RecordLogin(t *testing.T) {
	obs := NewObservability()
	obs.RecordLogin(true)
	obs.RecordLogin(false)
	snap := obs.Metrics.Snapshot()
	if snap.LoginSuccess != 1 {
		t.Errorf("expected 1 login success, got %d", snap.LoginSuccess)
	}
	if snap.LoginFailure != 1 {
		t.Errorf("expected 1 login failure, got %d", snap.LoginFailure)
	}
}

func TestObservability_RecordBusinessEvent(t *testing.T) {
	obs := NewObservability()
	obs.RecordBusinessEvent("quest_completed")
	obs.RecordBusinessEvent("chest_opened")
	obs.RecordBusinessEvent("creative_submitted")
	snap := obs.Metrics.Snapshot()
	if snap.QuestCompleted != 1 {
		t.Errorf("expected 1 quest_completed, got %d", snap.QuestCompleted)
	}
	if snap.ChestOpened != 1 {
		t.Errorf("expected 1 chest_opened, got %d", snap.ChestOpened)
	}
	if snap.CreativeSubmitted != 1 {
		t.Errorf("expected 1 creative_submitted, got %d", snap.CreativeSubmitted)
	}
}

func TestObservability_RecordAdminOp(t *testing.T) {
	obs := NewObservability()
	obs.RecordAdminOp()
	snap := obs.Metrics.Snapshot()
	if snap.AdminOps != 1 {
		t.Errorf("expected 1 admin op, got %d", snap.AdminOps)
	}
}

func TestObservability_MergeCacheStats(t *testing.T) {
	obs := NewObservability()
	obs.MergeCacheStats(10, 5, 0)
	snap := obs.Metrics.Snapshot()
	if snap.CacheHits != 10 {
		t.Errorf("expected 10 cache hits, got %d", snap.CacheHits)
	}
	if snap.CacheMisses != 5 {
		t.Errorf("expected 5 cache misses, got %d", snap.CacheMisses)
	}
}

func TestObservability_Wrap_RecordsDuration(t *testing.T) {
	obs := NewObservability()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	snap := obs.Metrics.Snapshot()
	if len(snap.RequestLatencyMs) == 0 {
		t.Fatal("expected latency data")
	}
}

func TestContextWithRequestID(t *testing.T) {
	ctx := ContextWithRequestID(context.Background(), "my-request-id")
	if RequestIDFromContext(ctx) != "my-request-id" {
		t.Error("failed to retrieve request ID from context")
	}
}

func TestRequestID_Helper(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(HeaderRequestID, "helper-test-id")
	ctx := WithRequestID(context.Background(), "helper-test-id")
	req = req.WithContext(ctx)

	if RequestID(req) != "helper-test-id" {
		t.Errorf("expected helper-test-id, got %s", RequestID(req))
	}
}

func TestObservability_Wrap_RecoversPanic(t *testing.T) {
	var buf testBuffer
	obs := &Observability{
		Logger:   NewLogger(&buf),
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("deliberate test panic")
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 after panic, got %d", w.Code)
	}

	ct := w.Header().Get("Content-Type")
	if ct != "application/json" {
		t.Errorf("expected content-type application/json, got %s", ct)
	}

	body := w.Body.String()
	if body == "" {
		t.Fatal("expected non-empty error body")
	}

	obs.Logger.Flush()

	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["msg"] != "panic_recovered" {
		t.Errorf("expected panic_recovered log msg, got %v", entry["msg"])
	}
	if entry["error"] != "deliberate test panic" {
		t.Errorf("expected error field 'deliberate test panic', got %v", entry["error"])
	}
	if entry["request_id"] == nil || entry["request_id"] == "" {
		t.Error("expected request_id in panic log")
	}
	if entry["stack"] == nil || entry["stack"] == "" {
		t.Error("expected stack trace in panic log")
	}
}

func TestObservability_Wrap_RecoversPanic_NilLogger(t *testing.T) {
	obs := &Observability{
		Logger:   nil,
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("no logger panic")
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestObservability_Wrap_RecoversPanic_PreservesRequestID(t *testing.T) {
	var buf testBuffer
	obs := &Observability{
		Logger:   NewLogger(&buf),
		Metrics:  NewMetrics(),
		Profiler: NewProfiler(),
	}

	existingID := "preserved-req-id"
	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("request id test")
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	req.Header.Set(HeaderRequestID, existingID)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
	if w.Header().Get(HeaderRequestID) != existingID {
		t.Errorf("expected X-Request-ID %s, got %s", existingID, w.Header().Get(HeaderRequestID))
	}

	obs.Logger.Flush()
	var entry map[string]any
	if err := json.Unmarshal(buf.bytes(), &entry); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if entry["request_id"] != existingID {
		t.Errorf("expected request_id %s in log, got %v", existingID, entry["request_id"])
	}
}

func TestObservability_Wrap_PanicDoesNotCrash(t *testing.T) {
	obs := NewObservability()

	inner := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("crash test")
	})

	handler := obs.Wrap(inner)
	req := httptest.NewRequest(http.MethodGet, "/api/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
